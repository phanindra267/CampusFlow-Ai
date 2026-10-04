package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/campuscare/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CommunityRepository serves the collaboration layer: discussion threads,
// member notifications and opportunity applications.
type CommunityRepository struct {
	dbBase
}

func NewCommunityRepository(db *pgxpool.Pool) *CommunityRepository {
	return &CommunityRepository{dbBase: dbBase{db: db}}
}

const discussionColumns = `d.id, d.club_id, d.author_id, u.display_name, d.title, d.body,
	d.category, d.status, d.is_pinned, d.reply_count, d.last_activity_at,
	d.created_at, d.updated_at`

const discussionJoin = `FROM discussions d
	JOIN users u ON u.id = d.author_id
	LEFT JOIN clubs c ON c.id = d.club_id`

// ListDiscussions returns recent threads. When clubID is non-nil only that
// group's threads are returned, otherwise only campus-wide threads are.
func (r *CommunityRepository) ListDiscussions(ctx context.Context, clubID *string, limit, offset int) ([]domain.Discussion, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + discussionColumns + ` ` + discussionJoin + `
	WHERE ($1::uuid IS NULL AND d.club_id IS NULL) OR d.club_id = $1
	ORDER BY d.is_pinned DESC, d.last_activity_at DESC
	LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, clubID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list discussions: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Discussion, 0, limit)
	for rows.Next() {
		d, err := scanDiscussion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, rows.Err()
}

// GetDiscussion loads a single thread along with its replies.
func (r *CommunityRepository) GetDiscussion(ctx context.Context, id string) (*domain.Discussion, []domain.DiscussionReply, error) {
	if err := r.ready(); err != nil {
		return nil, nil, err
	}
	const query = `SELECT ` + discussionColumns + ` ` + discussionJoin + ` WHERE d.id = $1`

	d, err := scanDiscussion(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, fmt.Errorf("get discussion: %w", err)
	}

	replies, err := r.ListReplies(ctx, id, 500)
	if err != nil {
		return nil, nil, err
	}
	return d, replies, nil
}

func (r *CommunityRepository) ListReplies(ctx context.Context, discussionID string, limit int) ([]domain.DiscussionReply, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT r.id, r.discussion_id, r.author_id, u.display_name, r.parent_id,
		r.body, r.created_at, r.updated_at
	FROM discussion_replies r
	JOIN users u ON u.id = r.author_id
	WHERE r.discussion_id = $1
	ORDER BY r.created_at ASC
	LIMIT $2`

	rows, err := r.db.Query(ctx, query, discussionID, limit)
	if err != nil {
		return nil, fmt.Errorf("list replies: %w", err)
	}
	defer rows.Close()

	out := make([]domain.DiscussionReply, 0, limit)
	for rows.Next() {
		var reply domain.DiscussionReply
		if err := rows.Scan(&reply.ID, &reply.DiscussionID, &reply.AuthorID, &reply.AuthorName,
			&reply.ParentID, &reply.Body, &reply.CreatedAt, &reply.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan reply: %w", err)
		}
		out = append(out, reply)
	}
	return out, rows.Err()
}

// CreateDiscussion opens a thread and notifies the group's members.
func (r *CommunityRepository) CreateDiscussion(ctx context.Context, d *domain.Discussion) error {
	if err := r.ready(); err != nil {
		return err
	}
	const query = `INSERT INTO discussions (club_id, author_id, title, body, category)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, status, is_pinned, reply_count, last_activity_at, created_at, updated_at`

	if err := r.db.QueryRow(ctx, query, d.ClubID, d.AuthorID, d.Title, d.Body, d.Category).
		Scan(&d.ID, &d.Status, &d.IsPinned, &d.ReplyCount, &d.LastActivityAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
		if isForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		return fmt.Errorf("create discussion: %w", err)
	}

	if d.ClubID != nil {
		return r.notifyClubMembers(ctx, *d.ClubID, domain.NotificationTypeDiscussion,
			"New discussion in your group", d.Title, "/community?discussion="+d.ID, d.AuthorID)
	}
	return nil
}

// AddReply posts a response. The reply counter and the thread's activity
// timestamp are updated in the same transaction so the listing order and the
// reply count can never disagree.
func (r *CommunityRepository) AddReply(ctx context.Context, reply *domain.DiscussionReply) error {
	if err := r.ready(); err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin add reply: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const insert = `INSERT INTO discussion_replies (discussion_id, author_id, parent_id, body)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	if err := tx.QueryRow(ctx, insert, reply.DiscussionID, reply.AuthorID, reply.ParentID, reply.Body).
		Scan(&reply.ID, &reply.CreatedAt, &reply.UpdatedAt); err != nil {
		if isForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		return fmt.Errorf("insert reply: %w", err)
	}

	// Bump the counter and activity time only for threads that still accept
	// replies, so a reply to a locked thread cannot silently revive it.
	const bump = `UPDATE discussions
		SET reply_count = reply_count + 1,
		    last_activity_at = now(),
		    updated_at = now()
		WHERE id = $1 AND status = 'OPEN'`

	var open bool
	if err := tx.QueryRow(ctx,
		`SELECT status = 'OPEN' FROM discussions WHERE id = $1`,
		reply.DiscussionID).Scan(&open); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("lock check: %w", err)
	}
	if !open {
		return fmt.Errorf("discussion is not open for replies")
	}
	if _, err := tx.Exec(ctx, bump, reply.DiscussionID); err != nil {
		return fmt.Errorf("bump discussion: %w", err)
	}

	return tx.Commit(ctx)
}

// notifyClubMembers sends a notification to every member of a group, skipping
// the actor so people are not told about their own action.
func (r *CommunityRepository) notifyClubMembers(ctx context.Context, clubID string, kind domain.NotificationType, title, body, link, actorID string) error {
	const query = `INSERT INTO notifications (user_id, type, title, body, link)
		SELECT cm.user_id, $2, $3, $4, $5
		FROM club_memberships cm
		WHERE cm.club_id = $1 AND cm.user_id <> $6 AND cm.status = 'ACTIVE'`

	if _, err := r.db.Exec(ctx, query, clubID, kind, title, body, link, actorID); err != nil {
		return fmt.Errorf("notify club members: %w", err)
	}
	return nil
}

// ListNotifications returns a member's notifications, newest first.
func (r *CommunityRepository) ListNotifications(ctx context.Context, userID string, unreadOnly bool, limit, offset int) ([]domain.Notification, int, error) {
	if err := r.ready(); err != nil {
		return nil, 0, err
	}
	const countQuery = `SELECT count(*) FROM notifications WHERE user_id = $1 AND ($2::boolean = false OR read_at IS NULL)`

	var unread int
	if err := r.db.QueryRow(ctx, countQuery, userID, unreadOnly).Scan(&unread); err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}

	const query = `SELECT id, user_id, type, title, body, link, read_at, created_at
		FROM notifications
		WHERE user_id = $1 AND ($2::boolean = false OR read_at IS NULL)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(ctx, query, userID, unreadOnly, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Notification, 0, limit)
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Link, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan notification: %w", err)
		}
		out = append(out, n)
	}
	return out, unread, rows.Err()
}

func (r *CommunityRepository) MarkNotificationRead(ctx context.Context, userID, id string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`UPDATE notifications SET read_at = now() WHERE id = $1 AND user_id = $2 AND read_at IS NULL`,
		id, userID)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Either it does not exist, belongs to someone else, or was already
		// read. All three are indistinguishable and all are a no-op for the
		// caller, so this is reported as not found.
		return ErrNotFound
	}
	return nil
}

func (r *CommunityRepository) MarkAllNotificationsRead(ctx context.Context, userID string) (int64, error) {
	if err := r.ready(); err != nil {
		return 0, err
	}
	tag, err := r.db.Exec(ctx,
		`UPDATE notifications SET read_at = now() WHERE user_id = $1 AND read_at IS NULL`, userID)
	if err != nil {
		return 0, fmt.Errorf("mark all read: %w", err)
	}
	return tag.RowsAffected(), nil
}

// ApplyToOpportunity records an application, treating a repeat submission as a
// conflict rather than an error so a double-clicked button is harmless.
func (r *CommunityRepository) ApplyToOpportunity(ctx context.Context, app *domain.OpportunityApplication) error {
	if err := r.ready(); err != nil {
		return err
	}
	const query = `INSERT INTO opportunity_applications (opportunity_id, user_id, cover_note, resume_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, status, applied_at, updated_at`

	err := r.db.QueryRow(ctx, query, app.OpportunityID, app.UserID, app.CoverNote, app.ResumeURL).
		Scan(&app.ID, &app.Status, &app.AppliedAt, &app.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateKey
		}
		if isForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		return fmt.Errorf("apply to opportunity: %w", err)
	}

	// Keep the listing's counter in step with the applications table.
	if _, err := r.db.Exec(ctx,
		`UPDATE opportunities SET application_count = (
			SELECT count(*) FROM opportunity_applications WHERE opportunity_id = $1
		) WHERE id = $1`, app.OpportunityID); err != nil {
		return fmt.Errorf("update application count: %w", err)
	}
	return nil
}

func (r *CommunityRepository) ListMyApplications(ctx context.Context, userID string, limit, offset int) ([]domain.OpportunityApplication, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT a.id, a.opportunity_id, a.user_id, o.title, a.status,
		a.cover_note, a.resume_url, a.applied_at, a.updated_at
	FROM opportunity_applications a
	JOIN opportunities o ON o.id = a.opportunity_id
	WHERE a.user_id = $1
		AND a.status <> 'WITHDRAWN'
	ORDER BY a.applied_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}
	defer rows.Close()

	out := make([]domain.OpportunityApplication, 0, limit)
	for rows.Next() {
		var a domain.OpportunityApplication
		if err := rows.Scan(&a.ID, &a.OpportunityID, &a.UserID, &a.Title, &a.Status,
			&a.CoverNote, &a.ResumeURL, &a.AppliedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan application: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *CommunityRepository) WithdrawApplication(ctx context.Context, userID, opportunityID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`UPDATE opportunity_applications SET status = 'WITHDRAWN', updated_at = now()
		WHERE user_id = $1 AND opportunity_id = $2 AND status NOT IN ('WITHDRAWN', 'REJECTED')`,
		userID, opportunityID)
	if err != nil {
		return fmt.Errorf("withdraw application: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	if _, err := r.db.Exec(ctx,
		`UPDATE opportunities SET application_count = (
			SELECT count(*) FROM opportunity_applications
			WHERE opportunity_id = $1 AND status <> 'WITHDRAWN'
		) WHERE id = $1`, opportunityID); err != nil {
		return fmt.Errorf("update application count: %w", err)
	}
	return nil
}

// ListResources browses campus facilities and services.
func (r *CommunityRepository) ListResources(ctx context.Context, resourceType, search string, limit, offset int) ([]domain.CampusResource, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT id, name, resource_type, COALESCE(description, ''),
		COALESCE(location, ''), COALESCE(capacity, 0), is_bookable, requires_auth,
		status, created_at, updated_at
	FROM campus_resources
	WHERE ($1 = '' OR resource_type = $1)
		AND ($2 = '' OR name ILIKE '%' || $2 || '%' OR COALESCE(description, '') ILIKE '%' || $2 || '%')
	ORDER BY name ASC
	LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(ctx, query, resourceType, search, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list resources: %w", err)
	}
	defer rows.Close()

	out := make([]domain.CampusResource, 0, limit)
	for rows.Next() {
		var res domain.CampusResource
		if err := rows.Scan(&res.ID, &res.Name, &res.ResourceType, &res.Description, &res.Location,
			&res.Capacity, &res.IsBookable, &res.RequiresAuth, &res.Status,
			&res.CreatedAt, &res.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan resource: %w", err)
		}
		out = append(out, res)
	}
	return out, rows.Err()
}

// BookResource reserves a slot. The overlap rule lives in the database as an
// exclusion constraint, so a conflicting booking surfaces as a constraint
// violation rather than a lost race between two concurrent requests.
func (r *CommunityRepository) BookResource(ctx context.Context, b *domain.ResourceBooking) error {
	if err := r.ready(); err != nil {
		return err
	}
	const query = `INSERT INTO resource_bookings (resource_id, user_id, start_time, end_time, note)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, status, created_at`

	err := r.db.QueryRow(ctx, query, b.ResourceID, b.UserID, b.StartTime, b.EndTime, b.Note).
		Scan(&b.ID, &b.Status, &b.CreatedAt)
	if err != nil {
		if isForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		if isExclusionViolation(err) {
			return ErrSlotUnavailable
		}
		return fmt.Errorf("book resource: %w", err)
	}
	return nil
}

func (r *CommunityRepository) CancelBooking(ctx context.Context, userID, bookingID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`UPDATE resource_bookings SET status = 'CANCELLED'
		WHERE id = $1 AND user_id = $2 AND status = 'CONFIRMED'`, bookingID, userID)
	if err != nil {
		return fmt.Errorf("cancel booking: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CommunityRepository) ListMyBookings(ctx context.Context, userID string, upcomingOnly bool) ([]domain.ResourceBooking, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT b.id, b.resource_id, r.name, b.user_id, u.display_name,
		b.start_time, b.end_time, COALESCE(b.note, ''), b.status, b.created_at
	FROM resource_bookings b
	JOIN campus_resources r ON r.id = b.resource_id
	JOIN users u ON u.id = b.user_id
	WHERE b.user_id = $1
		AND ($2::boolean = false OR (b.status = 'CONFIRMED' AND b.end_time > now()))
	ORDER BY b.start_time DESC`

	rows, err := r.db.Query(ctx, query, userID, upcomingOnly)
	if err != nil {
		return nil, fmt.Errorf("list bookings: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ResourceBooking, 0, 16)
	for rows.Next() {
		var b domain.ResourceBooking
		if err := rows.Scan(&b.ID, &b.ResourceID, &b.ResourceNm, &b.UserID, &b.UserName,
			&b.StartTime, &b.EndTime, &b.Note, &b.Status, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// Search performs one keyword query across the community entities that members
// actually browse: events, groups, opportunities, services and discussions.
func (r *CommunityRepository) Search(ctx context.Context, query string, entityType string, limit int) (*domain.SearchResults, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	type source struct {
		kind       string
		sql        string
		eventAtCol string
	}

	sources := []source{
		{"EVENT", `SELECT e.id, e.title, COALESCE(e.description, ''), e.category,
			'/events/' || e.id, e.start_time, 1.0 FROM events e
			WHERE e.status = 'PUBLISHED'
				AND (e.title ILIKE $1 OR COALESCE(e.description, '') ILIKE $1
					OR COALESCE(e.venue, '') ILIKE $1)`, "6"},
		{"CLUB", `SELECT c.id, c.name, COALESCE(c.description, ''), COALESCE(c.category, ''),
			'/clubs/' || c.id, NULL::timestamptz, 1.0 FROM clubs c
			WHERE c.status = 'ACTIVE'
				AND (c.name ILIKE $1 OR COALESCE(c.description, '') ILIKE $1)`, ""},
		{"OPPORTUNITY", `SELECT o.id, o.title, o.description, o.type,
			'/opportunities/' || o.id, o.start_date, 1.0 FROM opportunities o
			WHERE o.status = 'OPEN'
				AND (o.title ILIKE $1 OR o.description ILIKE $1)`, "6"},
		{"RESOURCE", `SELECT r.id, r.name, COALESCE(r.description, ''), r.resource_type,
			'/services/' || r.id, NULL::timestamptz, 1.0 FROM campus_resources r
			WHERE r.status = 'AVAILABLE'
				AND (r.name ILIKE $1 OR COALESCE(r.description, '') ILIKE $1
					OR COALESCE(r.location, '') ILIKE $1)`, ""},
		{"DISCUSSION", `SELECT d.id, d.title, d.body, d.category,
			'/community?discussion=' || d.id, NULL::timestamptz, 1.0 FROM discussions d
			WHERE d.status = 'OPEN'
				AND (d.title ILIKE $1 OR d.body ILIKE $1)`, ""},
	}

	pattern := "%" + trimmed + "%"
	results := make([]domain.SearchResult, 0, limit)
	byType := map[string]int{}

	for _, src := range sources {
		if entityType != "" && entityType != src.kind {
			continue
		}

		cols := "id, title, summary, category, url"
		if src.eventAtCol != "" {
			cols += ", event_at"
		} else {
			cols += ", NULL::timestamptz"
		}
		cols += ", score"

		rows, err := r.db.Query(ctx, src.sql+` LIMIT $2`, pattern, limit)
		if err != nil {
			return nil, fmt.Errorf("search %s: %w", src.kind, err)
		}

		for rows.Next() {
			var res domain.SearchResult
			res.EntityType = src.kind
			if err := rows.Scan(&res.EntityID, &res.Title, &res.Summary, &res.Category,
				&res.URL, &res.EventAt, &res.Score); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan search %s: %w", src.kind, err)
			}
			results = append(results, res)
			byType[src.kind]++
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, fmt.Errorf("search %s: %w", src.kind, err)
		}
	}

	// Earlier starts sort first, so results stay predictable.
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].Score > results[j-1].Score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}
	if len(results) > limit {
		results = results[:limit]
	}

	return &domain.SearchResults{
		Query:   trimmed,
		Total:   len(results),
		Results: results,
		ByType:  byType,
	}, nil
}

func scanDiscussion(row pgx.Row) (*domain.Discussion, error) {
	var d domain.Discussion
	if err := row.Scan(&d.ID, &d.ClubID, &d.AuthorID, &d.AuthorName, &d.Title, &d.Body,
		&d.Category, &d.Status, &d.IsPinned, &d.ReplyCount, &d.LastActivityAt,
		&d.CreatedAt, &d.UpdatedAt); err != nil {
		return nil, err
	}
	return &d, nil
}
