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

// ClubRepository owns the club lifecycle: creating and editing a club, following
// versus joining, what the club has been doing, and the numbers a club office
// reviews between meetings.
type ClubRepository struct {
	dbBase
}

func NewClubRepository(db *pgxpool.Pool) *ClubRepository {
	return &ClubRepository{dbBase: dbBase{db: db}}
}

// clubDetailColumns resolves member and follower counts and the caller's
// membership and follow state.
const clubDetailColumns = `c.id, c.name, c.slug, COALESCE(c.description, ''),
	COALESCE(c.category, ''), COALESCE(c.verification_status, 'PENDING'),
	COALESCE(c.status, 'ACTIVE'), COALESCE(c.logo_url, ''), COALESCE(c.contact_email, ''),
	c.member_count, c.follower_count, c.created_by, c.created_at, c.updated_at,
	EXISTS (SELECT 1 FROM club_memberships m WHERE m.club_id = c.id AND m.user_id = $1 AND m.status = 'ACTIVE') AS is_member,
	EXISTS (SELECT 1 FROM club_follows f WHERE f.club_id = c.id AND f.user_id = $1) AS is_following`

func scanClub(row pgx.Row) (*domain.Club, error) {
	var club domain.Club
	if err := row.Scan(&club.ID, &club.Name, &club.Slug, &club.Description,
		&club.Category, &club.VerificationStatus, &club.Status, &club.LogoURL,
		&club.ContactEmail, &club.MemberCount, &club.FollowerCount, &club.CreatedBy,
		&club.CreatedAt, &club.UpdatedAt, &club.IsMember, &club.IsFollowing); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan club: %w", err)
	}
	return &club, nil
}

// CreateClub registers a club on behalf of a member or organiser. The slug is
// derived from the name and de-duplicated with a numeric suffix, so two clubs
// with similar names do not fight over a URL.
func (r *ClubRepository) CreateClub(ctx context.Context, req *domain.CreateClubRequest, createdBy string) (*domain.Club, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}

	slug, err := r.uniqueSlug(ctx, slugify(req.Name))
	if err != nil {
		return nil, err
	}

	const query = `INSERT INTO clubs (name, slug, description, category, created_by,
			verification_status, status, logo_url, contact_email)
		VALUES ($1, $2, $3, $4, $5, 'PENDING', 'ACTIVE', NULLIF($6, ''), NULLIF($7, ''))
		RETURNING id`

	var id string
	if err := r.db.QueryRow(ctx, query, req.Name, slug, req.Description, req.Category,
		createdBy, req.LogoURL, req.ContactEmail).Scan(&id); err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		if isUniqueViolation(err) {
			return nil, ErrDuplicateKey
		}
		return nil, fmt.Errorf("create club: %w", err)
	}

	return r.GetClub(ctx, id, createdBy)
}

// UpdateClub patches a club. Empty strings leave a column unchanged so a partial
// payload cannot blank a description.
func (r *ClubRepository) UpdateClub(ctx context.Context, id string, req *domain.CreateClubRequest) (*domain.Club, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `UPDATE clubs SET
			name = COALESCE(NULLIF($2, ''), name),
			description = COALESCE(NULLIF($3, ''), description),
			category = COALESCE(NULLIF($4, ''), category),
			logo_url = COALESCE(NULLIF($5, ''), logo_url),
			contact_email = COALESCE(NULLIF($6, ''), contact_email),
			updated_at = now()
		WHERE id = $1
		RETURNING id`

	var updated string
	if err := r.db.QueryRow(ctx, query, id, req.Name, req.Description, req.Category,
		req.LogoURL, req.ContactEmail).Scan(&updated); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update club: %w", err)
	}
	return r.GetClub(ctx, id, "")
}

// ArchiveClub retires a club without deleting its history: its events and past
// contributions stay readable, but it no longer appears in listings.
func (r *ClubRepository) ArchiveClub(ctx context.Context, id string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`UPDATE clubs SET status = 'ARCHIVED', updated_at = now()
		WHERE id = $1 AND COALESCE(status, 'ACTIVE') <> 'ARCHIVED'`, id)
	if err != nil {
		return fmt.Errorf("archive club: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ClubRepository) GetClub(ctx context.Context, id, userID string) (*domain.Club, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + clubDetailColumns + ` FROM clubs c WHERE c.id = $2`

	club, err := scanClub(r.db.QueryRow(ctx, query, userID, id))
	if err != nil {
		return nil, fmt.Errorf("get club: %w", err)
	}
	return club, nil
}

// ListClubs browses active clubs with follower counts resolved.
func (r *ClubRepository) ListClubs(ctx context.Context, userID, category, search string, limit, offset int) ([]domain.Club, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + clubDetailColumns + `
		FROM clubs c
		WHERE COALESCE(c.status, 'ACTIVE') = 'ACTIVE'
			AND ($2 = '' OR COALESCE(c.category, '') = $2)
			AND ($3 = '' OR c.name ILIKE '%' || $3 || '%'
				OR COALESCE(c.description, '') ILIKE '%' || $3 || '%')
		ORDER BY c.follower_count DESC, c.name ASC
		LIMIT $4 OFFSET $5`

	rows, err := r.db.Query(ctx, query, userID, category, search, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list clubs: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Club, 0, limit)
	for rows.Next() {
		club, err := scanClub(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *club)
	}
	return out, rows.Err()
}

// ListMemberClubs is the member's own club list, resolved in one query instead
// of asking per club whether they belong.
func (r *ClubRepository) ListMemberClubs(ctx context.Context, userID string, limit, offset int) ([]domain.Club, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + clubDetailColumns + `
		FROM clubs c
		JOIN club_memberships m ON m.club_id = c.id AND m.user_id = $2 AND m.status = 'ACTIVE'
		WHERE COALESCE(c.status, 'ACTIVE') = 'ACTIVE'
		ORDER BY c.name ASC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(ctx, query, userID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list member clubs: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Club, 0, limit)
	for rows.Next() {
		club, err := scanClub(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *club)
	}
	return out, rows.Err()
}

// ListFollowedClubs is the member's follow list, which is not the same as
// membership: following means "tell me what this club is doing".
func (r *ClubRepository) ListFollowedClubs(ctx context.Context, userID string, limit, offset int) ([]domain.Club, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + clubDetailColumns + `
		FROM clubs c
		JOIN club_follows f ON f.club_id = c.id AND f.user_id = $2
		WHERE COALESCE(c.status, 'ACTIVE') = 'ACTIVE'
		ORDER BY f.created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(ctx, query, userID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list followed clubs: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Club, 0, limit)
	for rows.Next() {
		club, err := scanClub(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *club)
	}
	return out, rows.Err()
}

// FollowClub records interest without joining. The follower counter is
// recomputed in the same transaction so it cannot drift from the rows.
func (r *ClubRepository) FollowClub(ctx context.Context, clubID, userID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin follow club: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		`INSERT INTO club_follows (user_id, club_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`, userID, clubID); err != nil {
		if isForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		return fmt.Errorf("follow club: %w", err)
	}

	if err := refreshFollowerCount(ctx, tx, clubID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *ClubRepository) UnfollowClub(ctx context.Context, clubID, userID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin unfollow club: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`DELETE FROM club_follows WHERE club_id = $1 AND user_id = $2`, clubID, userID)
	if err != nil {
		return fmt.Errorf("unfollow club: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	if err := refreshFollowerCount(ctx, tx, clubID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// refreshFollowerCount recomputes the denormalised counter from the rows.
func refreshFollowerCount(ctx context.Context, tx pgx.Tx, clubID string) error {
	if _, err := tx.Exec(ctx,
		`UPDATE clubs SET follower_count = (
			SELECT count(*) FROM club_follows WHERE club_id = $1
		) WHERE id = $1`, clubID); err != nil {
		return fmt.Errorf("refresh follower count: %w", err)
	}
	return nil
}

// ListClubCategories returns the categories in use so the browse filter matches
// what actually exists.
func (r *ClubRepository) ListClubCategories(ctx context.Context) ([]domain.CategoryCount, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT COALESCE(category, 'UNCATEGORISED'), count(*)
		FROM clubs WHERE COALESCE(status, 'ACTIVE') = 'ACTIVE'
		GROUP BY 1 ORDER BY 2 DESC, 1 ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("club categories: %w", err)
	}
	defer rows.Close()

	out := make([]domain.CategoryCount, 0, 16)
	for rows.Next() {
		var entry domain.CategoryCount
		if err := rows.Scan(&entry.Label, &entry.Count); err != nil {
			return nil, fmt.Errorf("scan club category: %w", err)
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

// ----------------------------------------------------------- club activity

const clubActivityColumns = `a.id, a.club_id, a.title, COALESCE(a.description, ''),
	COALESCE(a.activity_type, ''), COALESCE(a.location, ''), a.starts_at, a.ends_at,
	a.created_by, a.created_at`

// AddClubActivity records something the club did, keeping the club page current
// without a separate feed.
func (r *ClubRepository) AddClubActivity(ctx context.Context, clubID string, req *domain.CreateClubActivityRequest, createdBy string) (*domain.ClubActivity, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	if req.StartsAt != nil && req.EndsAt != nil && !req.EndsAt.After(*req.StartsAt) {
		return nil, fmt.Errorf("activity must end after it starts")
	}

	const query = `INSERT INTO club_activities (club_id, title, description,
			activity_type, location, starts_at, ends_at, created_by)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), $6, $7, $8)
		RETURNING id`

	var id string
	if err := r.db.QueryRow(ctx, query, clubID, req.Title, req.Description, req.Type,
		req.Location, req.StartsAt, req.EndsAt, createdBy).Scan(&id); err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		return nil, fmt.Errorf("add club activity: %w", err)
	}

	activity, err := r.getClubActivity(ctx, id)
	if err != nil {
		return nil, err
	}
	// Followers are notified by the caller rather than here: the fan-out is a
	// separate concern with its own failure mode, and doing it after the
	// activity is committed keeps a notification problem from losing the post.
	return activity, nil
}

func (r *ClubRepository) getClubActivity(ctx context.Context, id string) (*domain.ClubActivity, error) {
	const query = `SELECT ` + clubActivityColumns + ` FROM club_activities a WHERE a.id = $1`

	var a domain.ClubActivity
	if err := r.db.QueryRow(ctx, query, id).Scan(&a.ID, &a.ClubID, &a.Title,
		&a.Description, &a.Type, &a.Location, &a.StartsAt, &a.EndsAt, &a.CreatedBy,
		&a.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get club activity: %w", err)
	}
	return &a, nil
}

// ListClubActivities returns a club's history, soonest upcoming first so a member
// reading the page sees what is next before what already happened.
func (r *ClubRepository) ListClubActivities(ctx context.Context, clubID string, limit, offset int) ([]domain.ClubActivity, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + clubActivityColumns + `
		FROM club_activities a WHERE a.club_id = $1
		ORDER BY (a.starts_at IS NULL), a.starts_at DESC, a.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, clubID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list club activities: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ClubActivity, 0, limit)
	for rows.Next() {
		var a domain.ClubActivity
		if err := rows.Scan(&a.ID, &a.ClubID, &a.Title, &a.Description, &a.Type,
			&a.Location, &a.StartsAt, &a.EndsAt, &a.CreatedBy, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan club activity: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *ClubRepository) DeleteClubActivity(ctx context.Context, clubID, activityID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`DELETE FROM club_activities WHERE id = $1 AND club_id = $2`, activityID, clubID)
	if err != nil {
		return fmt.Errorf("delete club activity: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ------------------------------------------------------------ club projects

const clubProjectColumns = `p.id, p.club_id, p.title, COALESCE(p.description, ''),
	p.status, p.created_by, p.created_at`

// AddClubProject records work the club is doing.
func (r *ClubRepository) AddClubProject(ctx context.Context, clubID string, req *domain.CreateClubProjectRequest, createdBy string) (*domain.ClubProject, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = "ACTIVE"
	}

	const query = `INSERT INTO club_projects (club_id, title, description, status, created_by)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5)
		RETURNING id`

	var id string
	if err := r.db.QueryRow(ctx, query, clubID, req.Title, req.Description, status,
		createdBy).Scan(&id); err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		if isCheckViolation(err) {
			return nil, fmt.Errorf("unknown project status %q", status)
		}
		return nil, fmt.Errorf("add club project: %w", err)
	}

	const read = `SELECT ` + clubProjectColumns + ` FROM club_projects p WHERE p.id = $1`
	var project domain.ClubProject
	if err := r.db.QueryRow(ctx, read, id).Scan(&project.ID, &project.ClubID,
		&project.Title, &project.Description, &project.Status, &project.CreatedBy,
		&project.CreatedAt); err != nil {
		return nil, fmt.Errorf("get club project: %w", err)
	}
	return &project, nil
}

func (r *ClubRepository) ListClubProjects(ctx context.Context, clubID string, limit, offset int) ([]domain.ClubProject, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + clubProjectColumns + `
		FROM club_projects p WHERE p.club_id = $1
		ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, clubID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list club projects: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ClubProject, 0, limit)
	for rows.Next() {
		var p domain.ClubProject
		if err := rows.Scan(&p.ID, &p.ClubID, &p.Title, &p.Description, &p.Status,
			&p.CreatedBy, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan club project: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// UpdateClubProjectStatus moves a project along its lifecycle.
func (r *ClubRepository) UpdateClubProjectStatus(ctx context.Context, clubID, projectID, status string) (*domain.ClubProject, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	switch status {
	case "PLANNED", "ACTIVE", "COMPLETED", "PAUSED":
	default:
		return nil, fmt.Errorf("unknown project status %q", status)
	}

	const query = `UPDATE club_projects SET status = $3, updated_at = now()
		WHERE id = $1 AND club_id = $2 RETURNING id`

	var updated string
	if err := r.db.QueryRow(ctx, query, projectID, clubID, status).Scan(&updated); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update club project: %w", err)
	}

	const read = `SELECT ` + clubProjectColumns + ` FROM club_projects p WHERE p.id = $1`
	var project domain.ClubProject
	if err := r.db.QueryRow(ctx, read, projectID).Scan(&project.ID, &project.ClubID,
		&project.Title, &project.Description, &project.Status, &project.CreatedBy,
		&project.CreatedAt); err != nil {
		return nil, fmt.Errorf("get club project: %w", err)
	}
	return &project, nil
}

// ------------------------------------------------------------ club analytics

// ClubAnalytics summarises one club for its officers. MostActiveMonths answers
// "when do we actually get engagement", which is what a club plans around.
func (r *ClubRepository) ClubAnalytics(ctx context.Context, clubID string) (*domain.ClubAnalytics, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}

	const query = `SELECT c.id, c.member_count, c.follower_count,
		(SELECT count(*) FROM engagement_activities ea
			WHERE ea.member_id IN (SELECT user_id FROM club_memberships WHERE club_id = c.id)
				AND ea.created_at > now() - interval '30 days'),
		(SELECT count(*) FROM club_activities WHERE club_id = c.id),
		(SELECT count(*) FROM events WHERE club_id = c.id),
		(SELECT count(*) FROM event_registrations r
			JOIN events e ON e.id = r.event_id WHERE e.club_id = c.id),
		(SELECT count(*) FROM opportunities WHERE club_id = c.id)
		FROM clubs c WHERE c.id = $1`

	stats := &domain.ClubAnalytics{}
	if err := r.db.QueryRow(ctx, query, clubID).Scan(&stats.ClubID, &stats.MemberCount,
		&stats.FollowerCount, &stats.ActiveMembers30d, &stats.Activities,
		&stats.EventsHosted, &stats.EventRegistrations,
		&stats.OpportunitiesPosted); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("club analytics: %w", err)
	}

	const rating = `SELECT COALESCE(avg(f.rating), 0), count(*)
		FROM feedback f
		WHERE f.status = 'VISIBLE'
			AND ((f.activity_type = 'EVENT' AND f.activity_id IN
					(SELECT id FROM events WHERE club_id = $1))
				OR (f.activity_type = 'OPPORTUNITY' AND f.activity_id IN
					(SELECT id FROM opportunities WHERE club_id = $1)))`

	if err := r.db.QueryRow(ctx, rating, clubID).Scan(&stats.AverageRating,
		&stats.RatingsCount); err != nil {
		return nil, fmt.Errorf("club rating aggregate: %w", err)
	}
	stats.AverageRating = round2(stats.AverageRating)

	const categoryQuery = `SELECT COALESCE(e.category, 'UNCATEGORISED'), count(*)
		FROM events e WHERE e.club_id = $1
		GROUP BY 1 ORDER BY 2 DESC, 1 ASC LIMIT 5`
	categoryRows, err := r.db.Query(ctx, categoryQuery, clubID)
	if err != nil {
		return nil, fmt.Errorf("club category breakdown: %w", err)
	}
	defer categoryRows.Close()
	for categoryRows.Next() {
		var entry domain.CategoryCount
		if err := categoryRows.Scan(&entry.Label, &entry.Count); err != nil {
			return nil, fmt.Errorf("scan club category: %w", err)
		}
		stats.TopCategories = append(stats.TopCategories, entry)
	}
	if err := categoryRows.Err(); err != nil {
		return nil, fmt.Errorf("club category breakdown: %w", err)
	}

	const months = `SELECT to_char(date_trunc('month', ea.created_at), 'YYYY-MM'), count(*)
		FROM engagement_activities ea
		WHERE ea.member_id IN (SELECT user_id FROM club_memberships WHERE club_id = $1)
			AND ea.created_at > now() - interval '6 months'
		GROUP BY 1 ORDER BY 1 ASC`
	monthRows, err := r.db.Query(ctx, months, clubID)
	if err != nil {
		return nil, fmt.Errorf("club monthly activity: %w", err)
	}
	defer monthRows.Close()
	for monthRows.Next() {
		var entry domain.CategoryCount
		if err := monthRows.Scan(&entry.Label, &entry.Count); err != nil {
			return nil, fmt.Errorf("scan club month: %w", err)
		}
		stats.MostActiveMonths = append(stats.MostActiveMonths, entry)
	}
	if err := monthRows.Err(); err != nil {
		return nil, fmt.Errorf("club monthly activity: %w", err)
	}

	return stats, nil
}

// slugify turns a club name into a URL fragment. Non-alphanumeric characters
// collapse to single hyphens.
func slugify(name string) string {
	var b strings.Builder
	lastHyphen := true
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		default:
			if !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "club"
	}
	if len(slug) > 60 {
		slug = strings.Trim(slug[:60], "-")
	}
	return slug
}

// uniqueSlug appends a numeric suffix until the slug is free.
func (r *ClubRepository) uniqueSlug(ctx context.Context, base string) (string, error) {
	candidate := base
	for attempt := 2; attempt <= 50; attempt++ {
		var taken bool
		if err := r.db.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM clubs WHERE slug = $1)`, candidate).Scan(&taken); err != nil {
			return "", fmt.Errorf("check slug availability: %w", err)
		}
		if !taken {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, attempt)
	}
	return "", fmt.Errorf("could not derive a unique slug for this club name")
}
