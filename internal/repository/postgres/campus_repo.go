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

// CampusRepository serves the group, listing and engagement surfaces: clubs and
// their memberships, opportunities, event registration, waitlists and event
// check-in.
type CampusRepository struct {
	dbBase
}

func NewCampusRepository(db *pgxpool.Pool) *CampusRepository {
	return &CampusRepository{dbBase: dbBase{db: db}}
}

// ---------------------------------------------------------------- clubs

const clubColumns = `c.id, c.name, c.slug, COALESCE(c.description, ''),
	COALESCE(c.category, ''), COALESCE(c.verification_status, 'PENDING'),
	COALESCE(c.status, 'ACTIVE'), c.created_by, c.created_at, c.updated_at`

// ListClubs browses active clubs, optionally filtered by category or keyword.
func (r *CampusRepository) ListClubs(ctx context.Context, category, search string, limit, offset int) ([]domain.Club, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + clubColumns + `
	FROM clubs c
	WHERE COALESCE(c.status, 'ACTIVE') = 'ACTIVE'
		AND ($1 = '' OR COALESCE(c.category, '') = $1)
		AND ($2 = '' OR c.name ILIKE '%' || $2 || '%' OR COALESCE(c.description, '') ILIKE '%' || $2 || '%')
	ORDER BY c.name ASC
	LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(ctx, query, category, search, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list clubs: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Club, 0, limit)
	for rows.Next() {
		var c domain.Club
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.Category,
			&c.VerificationStatus, &c.Status, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan club: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetClub loads one club with its member count.
func (r *CampusRepository) GetClub(ctx context.Context, id string) (*domain.Club, int, error) {
	if err := r.ready(); err != nil {
		return nil, 0, err
	}
	const query = `SELECT ` + clubColumns + `, c.member_count FROM clubs c WHERE c.id = $1`

	var c domain.Club
	var memberCount int
	err := r.db.QueryRow(ctx, query, id).Scan(&c.ID, &c.Name, &c.Slug, &c.Description,
		&c.Category, &c.VerificationStatus, &c.Status, &c.CreatedBy, &c.CreatedAt,
		&c.UpdatedAt, &memberCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, ErrNotFound
		}
		return nil, 0, fmt.Errorf("get club: %w", err)
	}
	return &c, memberCount, nil
}

func (r *CampusRepository) ListClubMembers(ctx context.Context, clubID string) ([]domain.ClubMembership, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT m.id, m.club_id, m.user_id, m.role, m.status, m.created_at
		FROM club_memberships m
		WHERE m.club_id = $1
		ORDER BY CASE m.role WHEN 'OWNER' THEN 0 WHEN 'ADMIN' THEN 1
			WHEN 'OFFICER' THEN 2 ELSE 3 END, m.created_at ASC`

	rows, err := r.db.Query(ctx, query, clubID)
	if err != nil {
		return nil, fmt.Errorf("list club members: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ClubMembership, 0, 32)
	for rows.Next() {
		var m domain.ClubMembership
		if err := rows.Scan(&m.ID, &m.ClubID, &m.UserID, &m.Role, &m.Status, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan club member: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// JoinClub adds the member as ACTIVE. Re-joining a group someone already belongs
// to reactivates the membership instead of failing, which is what a member who
// left and came back expects.
func (r *CampusRepository) JoinClub(ctx context.Context, clubID, userID string) (*domain.ClubMembership, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin join club: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const upsert = `INSERT INTO club_memberships (club_id, user_id, role, status)
		VALUES ($1, $2, 'MEMBER', 'ACTIVE')
		ON CONFLICT (club_id, user_id) DO UPDATE SET status = 'ACTIVE'
		RETURNING id, club_id, user_id, role, status, created_at`

	var m domain.ClubMembership
	if err := tx.QueryRow(ctx, upsert, clubID, userID).
		Scan(&m.ID, &m.ClubID, &m.UserID, &m.Role, &m.Status, &m.CreatedAt); err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		return nil, fmt.Errorf("join club: %w", err)
	}

	const count = `UPDATE clubs SET member_count = (
		SELECT count(*) FROM club_memberships WHERE club_id = $1 AND status = 'ACTIVE'
	) WHERE id = $1`
	if _, err := tx.Exec(ctx, count, clubID); err != nil {
		return nil, fmt.Errorf("update member count: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit join club: %w", err)
	}
	return &m, nil
}

func (r *CampusRepository) LeaveClub(ctx context.Context, clubID, userID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin leave club: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`DELETE FROM club_memberships WHERE club_id = $1 AND user_id = $2`, clubID, userID)
	if err != nil {
		return fmt.Errorf("leave club: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	const count = `UPDATE clubs SET member_count = (
		SELECT count(*) FROM club_memberships WHERE club_id = $1 AND status = 'ACTIVE'
	) WHERE id = $1`
	if _, err := tx.Exec(ctx, count, clubID); err != nil {
		return fmt.Errorf("update member count: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *CampusRepository) IsClubMember(ctx context.Context, clubID, userID string) (bool, error) {
	if err := r.ready(); err != nil {
		return false, err
	}
	const query = `SELECT EXISTS (
		SELECT 1 FROM club_memberships
		WHERE club_id = $1 AND user_id = $2 AND status = 'ACTIVE'
	)`

	var member bool
	if err := r.db.QueryRow(ctx, query, clubID, userID).Scan(&member); err != nil {
		return false, fmt.Errorf("is club member: %w", err)
	}
	return member, nil
}

// --------------------------------------------------------- opportunities

const opportunityColumns = `o.id, o.title, o.description, o.type, o.category,
	o.organizer_id, o.club_id,
	COALESCE(o.start_date, '0001-01-01 00:00:00+00'::timestamptz),
	COALESCE(o.end_date, '0001-01-01 00:00:00+00'::timestamptz),
	COALESCE(o.registration_deadline, '0001-01-01 00:00:00+00'::timestamptz),
	COALESCE(o.location, ''), COALESCE(o.delivery_mode, ''), COALESCE(o.capacity, 0),
	o.status, o.created_at, o.updated_at`

// ListOpportunities browses open listings. savedOnly restricts the result to
// the ones a member bookmarked.
func (r *CampusRepository) ListOpportunities(ctx context.Context, category, search string, savedOnly bool, userID string, limit, offset int) ([]domain.Opportunity, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	savedFilter := ""
	args := []any{category, search}
	if savedOnly {
		savedFilter = `AND EXISTS (
			SELECT 1 FROM saved_opportunities so
			WHERE so.opportunity_id = o.id AND so.user_id = $3
		)`
		args = append(args, userID)
	}

	// LIMIT/OFFSET are always the final two parameters so the saved filter can
	// reference $3 without changing the placeholder numbering.
	limitPos := len(args) + 1
	offsetPos := len(args) + 2

	query := fmt.Sprintf(`SELECT %s
		FROM opportunities o
		WHERE o.status = 'OPEN'
			AND ($1 = '' OR COALESCE(o.category, '') = $1)
			AND ($2 = '' OR o.title ILIKE '%%' || $2 || '%%' OR o.description ILIKE '%%' || $2 || '%%')
			%s
		ORDER BY o.start_date ASC NULLS LAST, o.created_at DESC
		LIMIT $%d OFFSET $%d`, opportunityColumns, savedFilter, limitPos, offsetPos)

	rows, err := r.db.Query(ctx, query, append(args, limit, offset)...)
	if err != nil {
		return nil, fmt.Errorf("list opportunities: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Opportunity, 0, limit)
	for rows.Next() {
		var o domain.Opportunity
		if err := rows.Scan(&o.ID, &o.Title, &o.Description, &o.Type, &o.Category,
			&o.OrganizerID, &o.ClubID, &o.StartDate, &o.EndDate, &o.RegistrationDeadline,
			&o.Location, &o.DeliveryMode, &o.Capacity, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan opportunity: %w", err)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *CampusRepository) GetOpportunity(ctx context.Context, id string) (*domain.Opportunity, int, error) {
	if err := r.ready(); err != nil {
		return nil, 0, err
	}
	const query = `SELECT ` + opportunityColumns + `, o.application_count
		FROM opportunities o WHERE o.id = $1`

	var o domain.Opportunity
	var applications int
	err := r.db.QueryRow(ctx, query, id).Scan(&o.ID, &o.Title, &o.Description, &o.Type,
		&o.Category, &o.OrganizerID, &o.ClubID, &o.StartDate, &o.EndDate,
		&o.RegistrationDeadline, &o.Location, &o.DeliveryMode, &o.Capacity, &o.Status,
		&o.CreatedAt, &o.UpdatedAt, &applications)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, ErrNotFound
		}
		return nil, 0, fmt.Errorf("get opportunity: %w", err)
	}
	return &o, applications, nil
}

func (r *CampusRepository) SaveOpportunity(ctx context.Context, userID, opportunityID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	const query = `INSERT INTO saved_opportunities (user_id, opportunity_id)
		VALUES ($1, $2) ON CONFLICT DO NOTHING`

	if _, err := r.db.Exec(ctx, query, userID, opportunityID); err != nil {
		if isForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		return fmt.Errorf("save opportunity: %w", err)
	}
	return nil
}

func (r *CampusRepository) UnsaveOpportunity(ctx context.Context, userID, opportunityID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`DELETE FROM saved_opportunities WHERE user_id = $1 AND opportunity_id = $2`, userID, opportunityID)
	if err != nil {
		return fmt.Errorf("unsave opportunity: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------------------------------------------------------------- events

const eventColumns = `e.id, e.title, e.description, e.category, e.organizer_id,
	COALESCE(e.venue, ''), e.is_online, e.start_time, e.end_time,
	e.registration_deadline, e.capacity, e.status, e.created_at, e.updated_at`

// ListEvents returns published events, with the member's registration status
// resolved in the same query so the UI can show "Registered" without a second
// round trip per event.
func (r *CampusRepository) ListEvents(ctx context.Context, category string, search string, userID string, limit, offset int) ([]domain.Event, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + eventColumns + `,
		(SELECT r.status FROM event_registrations r
			WHERE r.event_id = e.id AND r.member_id = $3 LIMIT 1) AS registration_status
		FROM events e
		WHERE e.status IN ('PUBLISHED', 'COMPLETED')
			AND ($1 = '' OR COALESCE(e.category, '') = $1)
			AND ($2 = '' OR e.title ILIKE '%' || $2 || '%' OR COALESCE(e.description, '') ILIKE '%' || $2 || '%')
		ORDER BY (e.status = 'PUBLISHED') DESC, e.start_time ASC
		LIMIT $4 OFFSET $5`

	rows, err := r.db.Query(ctx, query, category, search, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Event, 0, limit)
	for rows.Next() {
		var e domain.Event
		var regStatus *string
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.Category, &e.OrganizerID,
			&e.Venue, &e.IsOnline, &e.StartTime, &e.EndTime, &e.RegistrationDeadline,
			&e.Capacity, &e.Status, &e.CreatedAt, &e.UpdatedAt, &regStatus); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		if regStatus != nil {
			e.RegistrationStatus = *regStatus
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *CampusRepository) GetEvent(ctx context.Context, id, userID string) (*domain.Event, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + eventColumns + `,
		(SELECT r.status FROM event_registrations r
			WHERE r.event_id = e.id AND r.member_id = $2 LIMIT 1) AS registration_status
		FROM events e WHERE e.id = $1`

	var e domain.Event
	var regStatus *string
	err := r.db.QueryRow(ctx, query, id, userID).Scan(&e.ID, &e.Title, &e.Description,
		&e.Category, &e.OrganizerID, &e.Venue, &e.IsOnline, &e.StartTime, &e.EndTime,
		&e.RegistrationDeadline, &e.Capacity, &e.Status, &e.CreatedAt, &e.UpdatedAt, &regStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get event: %w", err)
	}
	if regStatus != nil {
		e.RegistrationStatus = *regStatus
	}
	return &e, nil
}

// RegisterForEvent books a place, refusing to overbook.
//
// Correctness under concurrency rests on two things:
//
//  1. `SELECT ... FOR UPDATE` on the events row. Every registration for an event
//     serialises on that single row lock, so two simultaneous requests cannot
//     both pass the capacity check for the last place.
//  2. The capacity count runs in a *separate statement* from the lock. Under
//     READ COMMITTED each statement takes a fresh snapshot, so the waiter sees
//     the winner's committed insert. Folding the count into the locking
//     statement would read the pre-lock snapshot and reintroduce the race the
//     lock exists to prevent.
//
// A capacity of 0 means unlimited, which is why the guard tests capacity > 0.
func (r *CampusRepository) RegisterForEvent(ctx context.Context, eventID, userID string) (*domain.EventRegistration, bool, error) {
	if err := r.ready(); err != nil {
		return nil, false, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("begin register: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Statement 1: take the per-event lock and read the immutable booking rules.
	const lock = `SELECT capacity, status FROM events WHERE id = $1 FOR UPDATE`

	var capacity int
	var status string
	if err := tx.QueryRow(ctx, lock, eventID).Scan(&capacity, &status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, ErrNotFound
		}
		return nil, false, fmt.Errorf("lock event: %w", err)
	}
	if status != string(domain.EventStatusPublished) {
		return nil, false, ErrEventNotOpen
	}

	// Statement 2: count committed registrations under a fresh snapshot.
	if capacity > 0 {
		var taken int
		const count = `SELECT count(*) FROM event_registrations
			WHERE event_id = $1 AND status = 'REGISTERED'`
		if err := tx.QueryRow(ctx, count, eventID).Scan(&taken); err != nil {
			return nil, false, fmt.Errorf("count registrations: %w", err)
		}
		if taken >= capacity {
			return nil, false, ErrEventFull
		}
	}

	const insert = `INSERT INTO event_registrations (event_id, member_id, status)
		VALUES ($1, $2, 'REGISTERED')
		RETURNING id, event_id, member_id, status, created_at`

	var reg domain.EventRegistration
	err = tx.QueryRow(ctx, insert, eventID, userID).
		Scan(&reg.ID, &reg.EventID, &reg.MemberID, &reg.Status, &reg.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, false, ErrDuplicateKey
		}
		if isForeignKeyViolation(err) {
			return nil, false, ErrInvalidReference
		}
		return nil, false, fmt.Errorf("insert registration: %w", err)
	}

	// Recording engagement here keeps the activity feed honest without a second
	// write path that could be forgotten.
	if _, err := tx.Exec(ctx,
		`INSERT INTO engagement_activities (member_id, activity_type, activity_id, event_type)
		VALUES ($1, 'EVENT', $2, 'REGISTERED_EVENT')`, userID, eventID); err != nil {
		return nil, false, fmt.Errorf("record engagement: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, fmt.Errorf("commit register: %w", err)
	}
	return &reg, true, nil
}

// CancelRegistration gives up a place and hands it straight to whoever has waited
// longest for it.
//
// It returns the promoted member's registration, or (nil, nil) when the waitlist
// was empty. The handler uses that to tell the person who got the seat.
//
// Four things have to happen atomically here, which is why this is a transaction
// rather than the two bare statements it replaces:
//
//  1. Take the per-event row lock, the same one RegisterForEvent and
//     PromoteWaitlist take. Without it a cancellation could interleave with a
//     registration claiming the freed place and either double-allocate it or
//     promote someone into a seat that is already gone.
//  2. Delete the registration.
//  3. Delete the engagement row that RegisterForEvent wrote. Leaving it behind
//     keeps a cancelled event in the member's activity feed as though they were
//     still going, which is the drift this method used to cause.
//  4. Promote the next waiting member, writing their engagement row in the same
//     commit so a promoted person never appears in the feed without a
//     registration behind them.
func (r *CampusRepository) CancelRegistration(ctx context.Context, eventID, userID string) (*domain.EventRegistration, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin cancel: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const lock = `SELECT id FROM events WHERE id = $1 FOR UPDATE`

	var lockedEventID string
	if err := tx.QueryRow(ctx, lock, eventID).Scan(&lockedEventID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lock event: %w", err)
	}

	// ATTENDED is deliberately excluded: somebody who actually turned up at the
	// event is a fact about the past, not a booking that can be withdrawn.
	tag, err := tx.Exec(ctx,
		`DELETE FROM event_registrations
		WHERE event_id = $1 AND member_id = $2 AND status <> 'ATTENDED'`, eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("cancel registration: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}

	if _, err := tx.Exec(ctx,
		`DELETE FROM engagement_activities
		WHERE member_id = $1 AND activity_type = 'EVENT'
			AND event_type = 'REGISTERED_EVENT'
			AND activity_id = $2`, userID, eventID); err != nil {
		return nil, fmt.Errorf("clear cancelled registration engagement: %w", err)
	}

	promoted, err := promoteNextFromWaitlist(ctx, tx, eventID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit cancel: %w", err)
	}
	return promoted, nil
}

// JoinWaitlist places a member on a full event's waitlist and reports their
// position. Re-joining refreshes an existing WAITING entry rather than
// duplicating it.
func (r *CampusRepository) JoinWaitlist(ctx context.Context, entry *domain.WaitlistEntry) (int, error) {
	if err := r.ready(); err != nil {
		return 0, err
	}
	const query = `INSERT INTO waitlist_entries (event_id, member_id, status)
		VALUES ($1, $2, 'WAITING')
		ON CONFLICT (event_id, member_id) DO UPDATE SET status = 'WAITING'
		RETURNING id, event_id, member_id, status, joined_at`

	if err := r.db.QueryRow(ctx, query, entry.EventID, entry.MemberID).
		Scan(&entry.ID, &entry.EventID, &entry.MemberID, &entry.Status, &entry.JoinedAt); err != nil {
		if isForeignKeyViolation(err) {
			return 0, ErrInvalidReference
		}
		return 0, fmt.Errorf("join waitlist: %w", err)
	}

	return r.waitlistPosition(ctx, entry.EventID, entry.MemberID)
}

func (r *CampusRepository) waitlistPosition(ctx context.Context, eventID, userID string) (int, error) {
	const query = `SELECT count(*) FROM waitlist_entries w
		WHERE w.event_id = $1 AND w.status = 'WAITING'
			AND w.joined_at <= (SELECT joined_at FROM waitlist_entries
				WHERE event_id = $1 AND member_id = $2)`

	var position int
	if err := r.db.QueryRow(ctx, query, eventID, userID).Scan(&position); err != nil {
		return 0, fmt.Errorf("waitlist position: %w", err)
	}
	return position, nil
}

func (r *CampusRepository) ListWaitlist(ctx context.Context, eventID string) ([]domain.WaitlistEntry, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT id, event_id, member_id, status, joined_at, promoted_at, expires_at
		FROM waitlist_entries WHERE event_id = $1 ORDER BY joined_at ASC`

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("list waitlist: %w", err)
	}
	defer rows.Close()

	out := make([]domain.WaitlistEntry, 0, 32)
	for rows.Next() {
		var e domain.WaitlistEntry
		if err := rows.Scan(&e.ID, &e.EventID, &e.MemberID, &e.Status, &e.JoinedAt,
			&e.PromotedAt, &e.ExpiresAt); err != nil {
			return nil, fmt.Errorf("scan waitlist: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ------------------------------------------------------------- check-in

// OpenCheckIn starts a check-in window for an event session and returns the
// token attendees scan. The token is generated here rather than accepted from
// the client so it cannot be chosen or replayed across sessions.
func (r *CampusRepository) OpenCheckIn(ctx context.Context, s *domain.CheckInSession) error {
	if err := r.ready(); err != nil {
		return err
	}
	token, err := generateCheckInToken()
	if err != nil {
		return fmt.Errorf("generate check-in token: %w", err)
	}

	const query = `INSERT INTO check_in_sessions (event_session_id, organizer_id, qr_token, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, event_session_id, organizer_id, qr_token, expires_at, created_at`

	if err := r.db.QueryRow(ctx, query, s.EventSessionID, s.OrganizerID, token, s.ExpiresAt).
		Scan(&s.ID, &s.EventSessionID, &s.OrganizerID, &s.QRToken, &s.ExpiresAt, &s.CreatedAt); err != nil {
		if isForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		if isUniqueViolation(err) {
			return fmt.Errorf("check-in window already open for this session")
		}
		return fmt.Errorf("open check-in: %w", err)
	}
	return nil
}

// SessionOwnedBy reports whether organizerID organises the event that
// eventSessionID belongs to.
//
// Organiser tooling is gated by role alone at the route, which is not enough:
// without this, any member holding the ORGANIZER role could open a check-in
// window on, or read the attendee list of, an event that belongs to a different
// organiser. The check is one indexed join on the session's primary key.
func (r *CampusRepository) SessionOwnedBy(ctx context.Context, eventSessionID, organizerID string) (bool, error) {
	if err := r.ready(); err != nil {
		return false, err
	}
	const query = `SELECT EXISTS (
			SELECT 1
			FROM event_sessions es
			JOIN events e ON e.id = es.event_id
			WHERE es.id = $1 AND e.organizer_id = $2
		)`

	var owned bool
	if err := r.db.QueryRow(ctx, query, eventSessionID, organizerID).Scan(&owned); err != nil {
		return false, fmt.Errorf("check session ownership: %w", err)
	}
	return owned, nil
}

// CheckInByToken records attendance against an open window.
//
// Three conditions must hold, and all three are enforced here rather than
// trusted from the caller:
//
//  1. The QR token names a check-in window that has not expired.
//  2. The member holds a live registration for the event that window belongs
//     to. The poster at the door is a location, not a capability: without this
//     join any authenticated member could photograph it and mark themselves
//     present at an event they never registered for, inflating the attendance
//     figures organiser analytics report.
//  3. Attendance is recorded at most once per member per session, which the
//     unique key on (event_session_id, member_id) enforces.
//
// When the insert produces no row the reason is probed, so the caller can tell
// "that code is not valid" apart from "you are not on the list".
func (r *CampusRepository) CheckInByToken(ctx context.Context, token, userID, method string) (*domain.AttendanceRecord, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin check in: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// The attendance row and the engagement row are written in one transaction.
	// They used to be two separate statements, which meant an engagement failure
	// could leave a member marked present with no record of it in their activity
	// feed, and the error returned claimed the whole check-in had failed when it
	// had in fact succeeded.
	const query = `WITH window AS (
			SELECT s.event_session_id, es.event_id
			FROM check_in_sessions s
			JOIN event_sessions es ON es.id = s.event_session_id
			WHERE s.qr_token = $1 AND s.expires_at > now()
		), eligible AS (
			SELECT w.event_session_id
			FROM window w
			WHERE EXISTS (
				SELECT 1 FROM event_registrations er
				WHERE er.event_id = w.event_id
					AND er.member_id = $2
					AND er.status IN ('REGISTERED', 'ATTENDED')
			)
		)
		INSERT INTO attendance_records (event_session_id, member_id, check_in_method, status)
		SELECT e.event_session_id, $2, $3, 'PRESENT' FROM eligible e
		ON CONFLICT (event_session_id, member_id) DO NOTHING
		RETURNING id, event_session_id, member_id, check_in_method, checked_in_at, status`

	var rec domain.AttendanceRecord
	err = tx.QueryRow(ctx, query, token, userID, method).
		Scan(&rec.ID, &rec.EventSessionID, &rec.MemberID, &rec.CheckInMethod, &rec.CheckedInAt, &rec.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, r.explainCheckInRejection(ctx, token, userID)
		}
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		return nil, fmt.Errorf("check in: %w", err)
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO engagement_activities (member_id, activity_type, activity_id, event_type)
		VALUES ($1, 'EVENT', $2, 'ATTENDED_EVENT')`, userID, rec.EventSessionID); err != nil {
		return nil, fmt.Errorf("record engagement: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit check in: %w", err)
	}
	return &rec, nil
}

// explainCheckInRejection distinguishes the reasons an insert above returned no
// row. ErrNotRegistered is the security-relevant one; the rest all mean the same
// thing to the member standing at the door, who is holding an invalid code,
// scanning too early or scanning twice.
func (r *CampusRepository) explainCheckInRejection(ctx context.Context, token, userID string) error {
	const probe = `SELECT EXISTS (
			SELECT 1
			FROM check_in_sessions s
			JOIN event_sessions es ON es.id = s.event_session_id
			JOIN event_registrations er
				ON er.event_id = es.event_id AND er.member_id = $2
			WHERE s.qr_token = $1
				AND s.expires_at > now()
				AND er.status IN ('REGISTERED', 'ATTENDED')
		)`

	var registered bool
	if err := r.db.QueryRow(ctx, probe, token, userID).Scan(&registered); err != nil {
		// The probe is diagnostic only. If it fails there is still nothing to
		// record, so the caller gets the generic rejection rather than an error
		// that implies the check-in might have succeeded.
		return ErrCheckInRejected
	}
	if !registered {
		return ErrNotRegistered
	}
	return ErrCheckInRejected
}

func (r *CampusRepository) ListEventAttendance(ctx context.Context, sessionID string) ([]domain.AttendanceRecord, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT id, event_session_id, member_id, check_in_method, checked_in_at, status
		FROM attendance_records WHERE event_session_id = $1 ORDER BY checked_in_at ASC`

	rows, err := r.db.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list attendance: %w", err)
	}
	defer rows.Close()

	out := make([]domain.AttendanceRecord, 0, 64)
	for rows.Next() {
		var rec domain.AttendanceRecord
		if err := rows.Scan(&rec.ID, &rec.EventSessionID, &rec.MemberID, &rec.CheckInMethod,
			&rec.CheckedInAt, &rec.Status); err != nil {
			return nil, fmt.Errorf("scan attendance: %w", err)
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// ------------------------------------------------------------- feedback

// SubmitFeedback records a rating. The unique constraint on
// (activity_type, activity_id, member_id) means one rating per member, and a
// repeat submission updates the existing one rather than failing.
func (r *CampusRepository) SubmitFeedback(ctx context.Context, f *domain.Feedback) error {
	if err := r.ready(); err != nil {
		return err
	}
	const query = `INSERT INTO feedback (activity_type, activity_id, member_id, rating, comment)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (activity_type, activity_id, member_id)
		DO UPDATE SET rating = EXCLUDED.rating, comment = EXCLUDED.comment
		RETURNING id, activity_type, activity_id, member_id, rating, comment, status, created_at`

	if err := r.db.QueryRow(ctx, query, f.ActivityType, f.ActivityID, f.MemberID, f.Rating, f.Comment).
		Scan(&f.ID, &f.ActivityType, &f.ActivityID, &f.MemberID, &f.Rating, &f.Comment, &f.Status, &f.CreatedAt); err != nil {
		if isForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		if isCheckViolation(err) {
			return fmt.Errorf("rating must be between 1 and 5")
		}
		return fmt.Errorf("submit feedback: %w", err)
	}
	return nil
}

// AverageFeedback reports the mean rating and total count for an activity.
func (r *CampusRepository) AverageFeedback(ctx context.Context, activityType, activityID string) (float64, int, error) {
	if err := r.ready(); err != nil {
		return 0, 0, err
	}
	const query = `SELECT COALESCE(avg(rating), 0), count(*)
		FROM feedback
		WHERE activity_type = $1 AND activity_id = $2 AND status = 'VISIBLE'`

	var average float64
	var total int
	if err := r.db.QueryRow(ctx, query, activityType, activityID).Scan(&average, &total); err != nil {
		return 0, 0, fmt.Errorf("average feedback: %w", err)
	}
	return average, total, nil
}

// ------------------------------------------------------------- activity

// RecentActivity returns the member's own participation history, newest first.
func (r *CampusRepository) RecentActivity(ctx context.Context, userID string, limit int) ([]domain.EngagementActivity, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT id, member_id, activity_type, activity_id, event_type, created_at
		FROM engagement_activities WHERE member_id = $1
		ORDER BY created_at DESC LIMIT $2`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("recent activity: %w", err)
	}
	defer rows.Close()

	out := make([]domain.EngagementActivity, 0, limit)
	for rows.Next() {
		var a domain.EngagementActivity
		if err := rows.Scan(&a.ID, &a.MemberID, &a.ActivityType, &a.ActivityID,
			&a.EventType, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan activity: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// checkInWindowTTL is how long a generated check-in token stays valid.
const checkInWindowTTL = 15 * time.Minute

// defaultCheckInExpiry returns the expiry for a window opened now.
func defaultCheckInExpiry() time.Time {
	return time.Now().Add(checkInWindowTTL)
}
