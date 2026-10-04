package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/campuscare/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AnnouncementRepository serves the campus notice board. An announcement is how a
// member learns about a closure, a deadline or a campus-wide change without
// subscribing to anything, so visibility is scheduled rather than permanent: a
// notice can be written now, published at 08:00 and expire after a week.
type AnnouncementRepository struct {
	dbBase
	notifications *NotificationRepository
}

func NewAnnouncementRepository(db *pgxpool.Pool) *AnnouncementRepository {
	return &AnnouncementRepository{
		dbBase:        dbBase{db: db},
		notifications: NewNotificationRepository(db),
	}
}

const announcementColumns = `a.id, a.title, a.body, a.category, a.audience,
	a.is_pinned, a.status, a.publish_at, a.expires_at, a.published_at, a.author_id,
	COALESCE(u.display_name, ''), a.created_at, a.updated_at`

const announcementJoins = `FROM announcements a JOIN users u ON u.id = a.author_id`

func scanAnnouncement(row pgx.Row) (*domain.Announcement, error) {
	var a domain.Announcement
	if err := row.Scan(&a.ID, &a.Title, &a.Body, &a.Category, &a.Audience, &a.IsPinned,
		&a.Status, &a.PublishAt, &a.ExpiresAt, &a.PublishedAt, &a.AuthorID,
		&a.AuthorName, &a.CreatedAt, &a.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan announcement: %w", err)
	}
	return &a, nil
}

// ListAnnouncements returns live notices: published, past their publish time and
// not yet expired, pinned first and then newest first.
func (r *AnnouncementRepository) ListAnnouncements(ctx context.Context, category, audience string, limit, offset int) ([]domain.Announcement, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + announcementColumns + ` ` + announcementJoins + `
		WHERE a.status = 'PUBLISHED'
			AND (a.publish_at IS NULL OR a.publish_at <= now())
			AND (a.expires_at IS NULL OR a.expires_at > now())
			AND ($1 = '' OR a.category = $1)
			AND ($2 = '' OR a.audience = $2 OR a.audience = 'ALL')
		ORDER BY a.is_pinned DESC, a.published_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(ctx, query, category, audience, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list announcements: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Announcement, 0, limit)
	for rows.Next() {
		announcement, err := scanAnnouncement(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *announcement)
	}
	return out, rows.Err()
}

func (r *AnnouncementRepository) GetAnnouncement(ctx context.Context, id string) (*domain.Announcement, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + announcementColumns + ` ` + announcementJoins + `
		WHERE a.id = $1`

	announcement, err := scanAnnouncement(r.db.QueryRow(ctx, query, id))
	if err != nil {
		return nil, fmt.Errorf("get announcement: %w", err)
	}
	return announcement, nil
}

// CreateAnnouncement writes a notice. published_at is set when the notice is
// live, and left null for a draft or a future-scheduled notice so the record
// shows when it actually went out.
func (r *AnnouncementRepository) CreateAnnouncement(ctx context.Context, req *domain.CreateAnnouncementRequest, authorID string) (*domain.Announcement, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}

	status := req.Status
	if status == "" {
		status = "PUBLISHED"
	}
	if req.ExpiresAt != nil && req.PublishAt != nil && !req.ExpiresAt.After(*req.PublishAt) {
		return nil, fmt.Errorf("an announcement must expire after it appears")
	}

	const query = `INSERT INTO announcements (title, body, category, audience, is_pinned,
			status, publish_at, expires_at, author_id, published_at)
		VALUES ($1, $2, $3, COALESCE(NULLIF($4, ''), 'ALL'), $5, $6::varchar,
			$7::timestamptz, $8::timestamptz, $9,
			CASE WHEN $6::varchar = 'PUBLISHED'
				AND ($7::timestamptz IS NULL OR $7::timestamptz <= now())
				THEN now() ELSE NULL END)
		RETURNING id`

	var id string
	if err := r.db.QueryRow(ctx, query, req.Title, req.Body, req.Category, req.Audience,
		req.IsPinned, status, req.PublishAt, req.ExpiresAt, authorID).Scan(&id); err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		if isCheckViolation(err) {
			return nil, fmt.Errorf("unknown announcement category or audience")
		}
		return nil, fmt.Errorf("create announcement: %w", err)
	}

	return r.GetAnnouncement(ctx, id)
}

// UpdateAnnouncement patches a notice. A notice that becomes live records
// published_at at that moment, and one that is pulled back clears it.
func (r *AnnouncementRepository) UpdateAnnouncement(ctx context.Context, id string, req *domain.UpdateAnnouncementRequest) (*domain.Announcement, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}

	const query = `UPDATE announcements SET
			title = COALESCE($2, title),
			body = COALESCE($3, body),
			category = COALESCE($4, category),
			audience = COALESCE($5, audience),
			is_pinned = COALESCE($6, is_pinned),
			status = COALESCE($7::varchar, status),
			publish_at = CASE WHEN $8::boolean THEN $9::timestamptz ELSE publish_at END,
			expires_at = CASE WHEN $10::boolean THEN $11::timestamptz ELSE expires_at END,
			published_at = CASE
				WHEN $7::varchar = 'PUBLISHED'
					AND published_at IS NULL
					AND (publish_at IS NULL OR publish_at <= now())
					THEN now()
				WHEN $7::varchar IS NOT NULL AND $7::varchar <> 'PUBLISHED' THEN NULL
				ELSE published_at END,
			updated_at = now()
		WHERE id = $1
		RETURNING id`

	var updated string
	if err := r.db.QueryRow(ctx, query, id, req.Title, req.Body, req.Category,
		req.Audience, req.IsPinned, req.Status, req.PublishAt != nil, req.PublishAt,
		req.ExpiresAt != nil, req.ExpiresAt).Scan(&updated); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		if isCheckViolation(err) {
			return nil, fmt.Errorf("unknown announcement category, audience or status")
		}
		return nil, fmt.Errorf("update announcement: %w", err)
	}

	return r.GetAnnouncement(ctx, id)
}

// PublishAnnouncement pushes a live notice to the members it targets and returns
// how many were reached. The audience vocabulary here mirrors the notification
// preferences: a member who switched announcements off is skipped.
func (r *AnnouncementRepository) PublishAnnouncement(ctx context.Context, id string) (int, error) {
	announcement, err := r.GetAnnouncement(ctx, id)
	if err != nil {
		return 0, err
	}
	if announcement.Status != "PUBLISHED" {
		return 0, fmt.Errorf("only a published announcement can be sent out")
	}

	return r.notifications.NotifyUsers(ctx, r.audienceMembers(ctx, announcement.Audience),
		domain.NotificationTypeAnnouncement, announcement.Title, announcement.Body,
		"/announcements/"+announcement.ID)
}

// audienceMembers resolves who a notice is for. An empty or ALL audience means
// every active member.
func (r *AnnouncementRepository) audienceMembers(ctx context.Context, audience string) []string {
	if audience == "" || audience == "ALL" || audience == "MEMBERS" {
		const query = `SELECT id::text FROM users WHERE status = 'ACTIVE'`
		rows, err := r.db.Query(ctx, query)
		if err != nil {
			return nil
		}
		defer rows.Close()

		out := make([]string, 0, 64)
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return out
			}
			out = append(out, id)
		}
		return out
	}

	// The remaining audiences are expressed as roles, which is what the platform
	// actually stores. STUDENTS and CLUBS map onto the accounts that carry those
	// roles rather than a separate membership table.
	rolesByAudience := map[string][]string{
		"STUDENTS":   {"MEMBER"},
		"ORGANISERS": {"ORGANIZER"},
		"ADMINS":     {"ADMIN", "SUPER_ADMIN"},
		"FACULTY":    {"ORGANIZER"},
		"STAFF":      {"ORGANIZER"},
		"CLUBS":      {"ORGANIZER", "MEMBER"},
	}

	roles, ok := rolesByAudience[audience]
	if !ok {
		return nil
	}

	const query = `SELECT id::text FROM users WHERE status = 'ACTIVE' AND role = ANY($1)`
	rows, err := r.db.Query(ctx, query, roles)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]string, 0, len(roles))
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return out
		}
		out = append(out, id)
	}
	return out
}

// ArchiveAnnouncement takes a notice down without deleting it.
func (r *AnnouncementRepository) ArchiveAnnouncement(ctx context.Context, id string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`UPDATE announcements SET status = 'ARCHIVED', published_at = NULL, updated_at = now()
		WHERE id = $1 AND status <> 'ARCHIVED'`, id)
	if err != nil {
		return fmt.Errorf("archive announcement: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ExpireAnnouncements closes notices whose window has passed. It is a maintenance
// step: filtering already hides them, so this only keeps the status honest for
// anyone reading the table directly.
func (r *AnnouncementRepository) ExpireAnnouncements(ctx context.Context, now time.Time) (int, error) {
	if err := r.ready(); err != nil {
		return 0, err
	}
	tag, err := r.db.Exec(ctx,
		`UPDATE announcements SET status = 'ARCHIVED', updated_at = now()
		WHERE status = 'PUBLISHED' AND expires_at IS NOT NULL AND expires_at <= $1`, now)
	if err != nil {
		return 0, fmt.Errorf("expire announcements: %w", err)
	}
	return int(tag.RowsAffected()), nil
}
