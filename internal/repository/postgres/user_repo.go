package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/campuscare/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrDuplicateKey is returned when an insert violates a unique constraint, so
// callers can distinguish "already exists" from a genuine database failure.
var ErrDuplicateKey = errors.New("duplicate key")

// ErrNotFound is returned when a lookup matched no row, letting callers
// distinguish absence from a real database error.
var ErrNotFound = errors.New("not found")

// ErrInvalidReference is returned when a referenced row exists but does not
// belong to the parent resource named in the same request.
var ErrInvalidReference = errors.New("invalid reference")

// ErrSlotUnavailable is returned when a booking overlaps an existing confirmed
// booking for the same resource. The overlap rule is enforced by an exclusion
// constraint, so this also covers two requests racing each other.
var ErrSlotUnavailable = errors.New("slot unavailable")

// ErrEventFull is returned when an event has no places left.
var ErrEventFull = errors.New("event is full")

// ErrCheckInRejected is returned when a check-in token is unknown, expired, or
// has already been used by that member for the session.
var ErrCheckInRejected = errors.New("check-in rejected")

// ErrNoDatabase is returned when a repository was built without a pool, which
// happens only in tests that exercise routing and authorisation without a
// database. Repositories report it instead of dereferencing a nil pool so a
// misconfigured deployment surfaces as a 503 rather than a recovered panic.
var ErrNoDatabase = errors.New("database not configured")

// dbBase is embedded by repositories that talk directly to PostgreSQL and
// provides the nil-pool guard.
type dbBase struct {
	db *pgxpool.Pool
}

// ready reports whether the repository can issue queries.
func (b dbBase) ready() error {
	if b.db == nil {
		return ErrNoDatabase
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

// isExclusionViolation matches 23P01, raised when a row conflicts with an
// exclusion constraint such as the resource booking overlap rule.
func isExclusionViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23P01"
}

// isCheckViolation matches 23514, raised by a CHECK constraint such as the
// 1-5 rating range on feedback.
func isCheckViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23514"
}

type UserRepository struct {
	dbBase
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{dbBase: dbBase{db: db}}
}

const userColumns = `id, email, display_name, password_hash, role, status, created_at, updated_at`

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if err := r.ready(); err != nil {
		return err
	}

	// A member with no chosen display name falls back to the local part of the
	// email address so community surfaces never show a blank author.
	displayName := strings.TrimSpace(user.DisplayName)
	if displayName == "" {
		displayName = strings.SplitN(user.Email, "@", 2)[0]
	}
	user.DisplayName = displayName

	query := `INSERT INTO users (id, email, display_name, password_hash, role, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`

	err := r.db.QueryRow(ctx, query, user.ID, user.Email, displayName,
		user.PasswordHash, user.Role, user.Status).Scan(&user.CreatedAt, &user.UpdatedAt)
	if isUniqueViolation(err) {
		return ErrDuplicateKey
	}
	return err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.getOne(ctx,
		`SELECT `+userColumns+` FROM users WHERE email = $1`,
		strings.ToLower(strings.TrimSpace(email)),
	)
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return r.getOne(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = $1`,
		strings.TrimSpace(id),
	)
}

func (r *UserRepository) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}

	const query = `SELECT id, email, display_name, role, status, created_at, updated_at
		FROM users
		ORDER BY created_at DESC, email ASC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0, limit)
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role,
			&user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// UpdateDisplayName changes how the member is shown across the community.
func (r *UserRepository) UpdateDisplayName(ctx context.Context, id, displayName string) error {
	if err := r.ready(); err != nil {
		return err
	}

	tag, err := r.db.Exec(ctx,
		`UPDATE users SET display_name = $2, updated_at = now() WHERE id = $1`, id, displayName)
	if err != nil {
		return fmt.Errorf("update display name: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) getOne(ctx context.Context, query, arg string) (*domain.User, error) {
	if arg == "" {
		return nil, ErrNotFound
	}
	if err := r.ready(); err != nil {
		return nil, err
	}

	var u domain.User
	if err := r.db.QueryRow(ctx, query, arg).Scan(
		&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.Role, &u.Status,
		&u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
