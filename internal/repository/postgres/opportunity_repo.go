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

// OpportunityRepository owns listings and the applications against them: the
// poster's write path, the member's shortlist, and the applicant list a poster
// needs in order to respond.
type OpportunityRepository struct {
	dbBase
}

func NewOpportunityRepository(db *pgxpool.Pool) *OpportunityRepository {
	return &OpportunityRepository{dbBase: dbBase{db: db}}
}

// opportunityDetailColumns resolves the member's saved state. $1 is the member.
// The three dates are optional in the database, and the domain models them as
// plain times, so a missing date is read back as the zero time: "not set" is
// exactly what a zero time means everywhere else in this package.
const opportunityDetailColumns = `o.id, o.title, o.description, o.type, o.category,
	o.organizer_id, o.club_id,
	COALESCE(o.start_date, '0001-01-01 00:00:00+00'::timestamptz),
	COALESCE(o.end_date, '0001-01-01 00:00:00+00'::timestamptz),
	COALESCE(o.registration_deadline, '0001-01-01 00:00:00+00'::timestamptz),
	COALESCE(o.location, ''), COALESCE(o.delivery_mode, ''), COALESCE(o.capacity, 0),
	o.status, COALESCE(o.eligibility, ''), COALESCE(o.source, 'INTERNAL'),
	COALESCE(o.apply_url, ''), o.application_count, o.created_at, o.updated_at,
	EXISTS (SELECT 1 FROM saved_opportunities s
		WHERE s.opportunity_id = o.id AND s.user_id = $1) AS is_saved`

func scanOpportunity(row pgx.Row) (*domain.Opportunity, error) {
	var o domain.Opportunity
	if err := row.Scan(&o.ID, &o.Title, &o.Description, &o.Type, &o.Category,
		&o.OrganizerID, &o.ClubID, &o.StartDate, &o.EndDate, &o.RegistrationDeadline,
		&o.Location, &o.DeliveryMode, &o.Capacity, &o.Status, &o.Eligibility,
		&o.Source, &o.ApplyURL, &o.ApplicationCount, &o.CreatedAt, &o.UpdatedAt,
		&o.IsSaved); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan opportunity: %w", err)
	}
	return &o, nil
}

// nullableTime turns an omitted time into a NULL so a date that was never set
// stays unset instead of landing in the database as year one.
func nullableTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// CreateOpportunity posts a listing. source defaults to INTERNAL because
// everything posted from inside CampusCare AI is campus-run unless stated
// otherwise, and capacity 0 means "no limit".
func (r *OpportunityRepository) CreateOpportunity(ctx context.Context, req *domain.CreateOpportunityRequest, organizerID string) (*domain.Opportunity, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	if !req.StartDate.IsZero() && !req.EndDate.IsZero() && !req.EndDate.After(req.StartDate) {
		return nil, fmt.Errorf("an opportunity must end after it starts")
	}

	source := req.Source
	if source == "" {
		source = domain.OpportunitySourceInternal
	}

	const query = `INSERT INTO opportunities (title, description, type, category,
			organizer_id, club_id, start_date, end_date, registration_deadline,
			location, delivery_mode, capacity, status, eligibility, source, apply_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7::timestamptz, $8::timestamptz,
			$9::timestamptz, NULLIF($10, ''), NULLIF($11, ''), $12, 'OPEN',
			NULLIF($13, ''), $14, NULLIF($15, ''))
		RETURNING id`

	var id string
	if err := r.db.QueryRow(ctx, query, req.Title, req.Description, req.Type,
		req.Category, organizerID, req.ClubID, nullableTime(req.StartDate),
		nullableTime(req.EndDate), nullableTime(req.RegistrationDeadline),
		req.Location, req.DeliveryMode, req.Capacity,
		req.Eligibility, source, req.ApplyURL).Scan(&id); err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		if isCheckViolation(err) {
			return nil, fmt.Errorf("unknown opportunity type %q", req.Type)
		}
		return nil, fmt.Errorf("create opportunity: %w", err)
	}

	return r.GetOpportunity(ctx, id, organizerID)
}

// UpdateOpportunity patches a listing. Zero times are treated as "leave alone",
// which keeps the partial payload usable from a form that only sends some
// fields.
func (r *OpportunityRepository) UpdateOpportunity(ctx context.Context, id string, req *domain.CreateOpportunityRequest) (*domain.Opportunity, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}

	const query = `UPDATE opportunities SET
			title = COALESCE(NULLIF($2, ''), title),
			description = COALESCE(NULLIF($3, ''), description),
			type = COALESCE(NULLIF($4, ''), type),
			category = COALESCE(NULLIF($5, ''), category),
			club_id = COALESCE($6, club_id),
			location = COALESCE(NULLIF($7, ''), location),
			delivery_mode = COALESCE(NULLIF($8, ''), delivery_mode),
			capacity = CASE WHEN $9 >= 0 THEN $9 ELSE capacity END,
			eligibility = COALESCE(NULLIF($10, ''), eligibility),
			source = COALESCE(NULLIF($11, ''), source),
			apply_url = COALESCE(NULLIF($12, ''), apply_url),
			start_date = CASE WHEN $13 THEN start_date ELSE $14 END,
			end_date = CASE WHEN $15 THEN end_date ELSE $16 END,
			registration_deadline = CASE WHEN $17 THEN registration_deadline ELSE $18 END,
			updated_at = now()
		WHERE id = $1
		RETURNING id`

	var updated string
	if err := r.db.QueryRow(ctx, query, id, req.Title, req.Description, req.Type,
		req.Category, req.ClubID, req.Location, req.DeliveryMode, req.Capacity,
		req.Eligibility, req.Source, req.ApplyURL, !req.StartDate.IsZero(),
		req.StartDate, !req.EndDate.IsZero(), req.EndDate,
		!req.RegistrationDeadline.IsZero(), req.RegistrationDeadline).Scan(&updated); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		if isCheckViolation(err) {
			return nil, fmt.Errorf("unknown opportunity type %q", req.Type)
		}
		return nil, fmt.Errorf("update opportunity: %w", err)
	}

	return r.GetOpportunity(ctx, id, "")
}

// CloseOpportunity stops new applications without deleting the listing or the
// applications already submitted.
func (r *OpportunityRepository) CloseOpportunity(ctx context.Context, id string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`UPDATE opportunities SET status = 'CLOSED', updated_at = now()
		WHERE id = $1 AND status <> 'CLOSED'`, id)
	if err != nil {
		return fmt.Errorf("close opportunity: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *OpportunityRepository) GetOpportunity(ctx context.Context, id, userID string) (*domain.Opportunity, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + opportunityDetailColumns + ` FROM opportunities o WHERE o.id = $2`

	opportunity, err := scanOpportunity(r.db.QueryRow(ctx, query, userID, id))
	if err != nil {
		return nil, fmt.Errorf("get opportunity: %w", err)
	}
	return opportunity, nil
}

// ListOpportunities browses open listings with the type filter and the member's
// saved filter applied in one statement.
func (r *OpportunityRepository) ListOpportunities(ctx context.Context, userID, category, kind, search string, savedOnly bool, limit, offset int) ([]domain.Opportunity, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + opportunityDetailColumns + `
		FROM opportunities o
		WHERE o.status = 'OPEN'
			AND ($2 = '' OR COALESCE(o.category, '') = $2)
			AND ($3 = '' OR o.type = $3)
			AND ($4 = '' OR o.title ILIKE '%' || $4 || '%'
				OR COALESCE(o.description, '') ILIKE '%' || $4 || '%'
				OR COALESCE(o.eligibility, '') ILIKE '%' || $4 || '%')
			AND (NOT $5 OR EXISTS (
				SELECT 1 FROM saved_opportunities s
				WHERE s.opportunity_id = o.id AND s.user_id = $1
			))
		ORDER BY (o.registration_deadline IS NULL), o.registration_deadline ASC,
			o.created_at DESC
		LIMIT $6 OFFSET $7`

	rows, err := r.db.Query(ctx, query, userID, category, kind, search, savedOnly, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list opportunities: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Opportunity, 0, limit)
	for rows.Next() {
		opportunity, err := scanOpportunity(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *opportunity)
	}
	return out, rows.Err()
}

// ListOpportunityCategories counts the listings per category so the browse
// filters reflect what is actually posted.
func (r *OpportunityRepository) ListOpportunityCategories(ctx context.Context) ([]domain.CategoryCount, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT COALESCE(category, 'UNCATEGORISED'), count(*)
		FROM opportunities WHERE status = 'OPEN'
		GROUP BY 1 ORDER BY 2 DESC, 1 ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("opportunity categories: %w", err)
	}
	defer rows.Close()

	out := make([]domain.CategoryCount, 0, 16)
	for rows.Next() {
		var entry domain.CategoryCount
		if err := rows.Scan(&entry.Label, &entry.Count); err != nil {
			return nil, fmt.Errorf("scan opportunity category: %w", err)
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

// ListOpportunityTypes counts the open listings per type.
func (r *OpportunityRepository) ListOpportunityTypes(ctx context.Context) ([]domain.CategoryCount, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT type, count(*) FROM opportunities WHERE status = 'OPEN'
		GROUP BY 1 ORDER BY 2 DESC, 1 ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("opportunity types: %w", err)
	}
	defer rows.Close()

	out := make([]domain.CategoryCount, 0, 16)
	for rows.Next() {
		var entry domain.CategoryCount
		if err := rows.Scan(&entry.Label, &entry.Count); err != nil {
			return nil, fmt.Errorf("scan opportunity type: %w", err)
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

// SaveOpportunity bookmarks a listing for the member.
func (r *OpportunityRepository) SaveOpportunity(ctx context.Context, userID, opportunityID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	if _, err := r.db.Exec(ctx,
		`INSERT INTO saved_opportunities (user_id, opportunity_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`, userID, opportunityID); err != nil {
		if isForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		return fmt.Errorf("save opportunity: %w", err)
	}
	return nil
}

func (r *OpportunityRepository) UnsaveOpportunity(ctx context.Context, userID, opportunityID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`DELETE FROM saved_opportunities WHERE user_id = $1 AND opportunity_id = $2`,
		userID, opportunityID)
	if err != nil {
		return fmt.Errorf("unsave opportunity: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// --------------------------------------------------------- applications

const applicationColumns = `a.id, a.opportunity_id, a.user_id, a.status,
	COALESCE(a.cover_note, ''), a.resume_url, a.applied_at, a.updated_at`

// ApplyToOpportunity records an application and keeps the denormalised count on
// the listing in step, in one transaction so the badge cannot drift.
func (r *OpportunityRepository) ApplyToOpportunity(ctx context.Context, opportunityID, userID, coverNote string, resumeURL *string) (*domain.OpportunityApplication, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin apply: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const open = `SELECT status FROM opportunities WHERE id = $1`
	var status string
	if err := tx.QueryRow(ctx, open, opportunityID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read opportunity status: %w", err)
	}
	if status != "OPEN" {
		return nil, fmt.Errorf("this opportunity is no longer accepting applications")
	}

	const insert = `INSERT INTO opportunity_applications (opportunity_id, user_id,
			status, cover_note, resume_url)
		VALUES ($1, $2, 'SUBMITTED', NULLIF($3, ''), $4)
		RETURNING id`
	var id string
	if err := tx.QueryRow(ctx, insert, opportunityID, userID, coverNote,
		resumeURL).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateKey
		}
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		return nil, fmt.Errorf("insert application: %w", err)
	}

	if _, err := tx.Exec(ctx,
		`UPDATE opportunities SET application_count = (
			SELECT count(*) FROM opportunity_applications
			WHERE opportunity_id = $1 AND status <> 'WITHDRAWN'
		) WHERE id = $1`, opportunityID); err != nil {
		return nil, fmt.Errorf("update application count: %w", err)
	}

	const read = `SELECT ` + applicationColumns + `
		FROM opportunity_applications a WHERE a.id = $1`
	var application domain.OpportunityApplication
	if err := tx.QueryRow(ctx, read, id).Scan(&application.ID, &application.OpportunityID,
		&application.UserID, &application.Status, &application.CoverNote,
		&application.ResumeURL, &application.AppliedAt, &application.UpdatedAt); err != nil {
		return nil, fmt.Errorf("read application: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit apply: %w", err)
	}
	return &application, nil
}

// WithdrawApplication lets a member pull back. The row is kept with status
// WITHDRAWN rather than deleted so the poster's history stays truthful.
func (r *OpportunityRepository) WithdrawApplication(ctx context.Context, userID, opportunityID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`UPDATE opportunity_applications SET status = 'WITHDRAWN', updated_at = now()
		WHERE opportunity_id = $1 AND user_id = $2 AND status <> 'WITHDRAWN'`,
		opportunityID, userID)
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

// ListApplicationsForOpportunity is the poster's applicant list, enriched with
// the applicant's program and year so a shortlisting decision does not need
// twelve extra profile requests.
func (r *OpportunityRepository) ListApplicationsForOpportunity(ctx context.Context, opportunityID, status string, limit, offset int) ([]domain.ApplicationDetail, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT a.id, a.opportunity_id, a.user_id,
			COALESCE(u.display_name, ''), COALESCE(u.email, ''),
			COALESCE(u.program, ''), u.year_of_study, a.status,
			COALESCE(a.cover_note, ''), COALESCE(a.resume_url, ''),
			a.applied_at, a.updated_at
		FROM opportunity_applications a
		JOIN users u ON u.id = a.user_id
		WHERE a.opportunity_id = $1
			AND ($2 = '' OR a.status = $2)
			AND a.status <> 'WITHDRAWN'
		ORDER BY a.applied_at ASC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(ctx, query, opportunityID, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ApplicationDetail, 0, limit)
	for rows.Next() {
		var d domain.ApplicationDetail
		if err := rows.Scan(&d.ID, &d.OpportunityID, &d.UserID, &d.DisplayName, &d.Email,
			&d.Program, &d.YearOfStudy, &d.Status, &d.CoverNote, &d.ResumeURL,
			&d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan application: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// UpdateApplicationStatus moves an application forward and tells the applicant.
// The notification is written by the caller so a delivery failure does not undo
// a status change that already succeeded.
func (r *OpportunityRepository) UpdateApplicationStatus(ctx context.Context, opportunityID, applicationID, status string) (*domain.ApplicationDetail, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	switch domain.ApplicationStatus(status) {
	case domain.ApplicationStatusSubmitted, domain.ApplicationStatusUnderReview,
		domain.ApplicationStatusShortlisted, domain.ApplicationStatusAccepted,
		domain.ApplicationStatusRejected, domain.ApplicationStatusWithdrawn:
	default:
		return nil, fmt.Errorf("unknown application status %q", status)
	}

	const query = `UPDATE opportunity_applications SET status = $3, updated_at = now()
		WHERE id = $1 AND opportunity_id = $2
		RETURNING id, opportunity_id, user_id`
	var id, oppID, userID string
	if err := r.db.QueryRow(ctx, query, applicationID, opportunityID, status).
		Scan(&id, &oppID, &userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		if isCheckViolation(err) {
			return nil, fmt.Errorf("unknown application status %q", status)
		}
		return nil, fmt.Errorf("update application status: %w", err)
	}

	const read = `SELECT a.id, a.opportunity_id, a.user_id,
			COALESCE(u.display_name, ''), COALESCE(u.email, ''),
			COALESCE(u.program, ''), u.year_of_study, a.status,
			COALESCE(a.cover_note, ''), COALESCE(a.resume_url, ''),
			a.applied_at, a.updated_at
		FROM opportunity_applications a
		JOIN users u ON u.id = a.user_id
		WHERE a.id = $1`
	var detail domain.ApplicationDetail
	if err := r.db.QueryRow(ctx, read, id).Scan(&detail.ID, &detail.OpportunityID,
		&detail.UserID, &detail.DisplayName, &detail.Email, &detail.Program,
		&detail.YearOfStudy, &detail.Status, &detail.CoverNote, &detail.ResumeURL,
		&detail.CreatedAt, &detail.UpdatedAt); err != nil {
		return nil, fmt.Errorf("read application: %w", err)
	}
	return &detail, nil
}

// OpportunityAnalytics is the demand picture behind a listing.
type OpportunityAnalytics struct {
	OpportunityID  string  `json:"opportunity_id"`
	Title          string  `json:"title"`
	Capacity       int     `json:"capacity"`
	Applications   int     `json:"applications"`
	Saved          int     `json:"saved"`
	Withdrawals    int     `json:"withdrawals"`
	AcceptanceRate float64 `json:"acceptance_rate"`
	AvgOpenHours   float64 `json:"avg_days_open"`
}

// OpportunityAnalytics answers "is this listing working" for the poster: how many
// people applied, how many shortlisted, and how long it has been open.
func (r *OpportunityRepository) OpportunityAnalytics(ctx context.Context, opportunityID string) (*OpportunityAnalytics, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT o.id, o.title, COALESCE(o.capacity, 0),
		(SELECT count(*) FROM opportunity_applications a
			WHERE a.opportunity_id = o.id AND a.status NOT IN ('WITHDRAWN', 'REJECTED')),
		(SELECT count(*) FROM opportunity_applications a
			WHERE a.opportunity_id = o.id AND a.status = 'WITHDRAWN'),
		(SELECT count(*) FROM saved_opportunities s WHERE s.opportunity_id = o.id),
		(SELECT count(*) FROM opportunity_applications a
			WHERE a.opportunity_id = o.id AND a.status IN ('SHORTLISTED', 'ACCEPTED')),
		COALESCE(EXTRACT(EPOCH FROM (now() - o.created_at)) / 3600, 0)
		FROM opportunities o WHERE o.id = $1`

	stats := &OpportunityAnalytics{}
	var forwarded, hoursOpen float64
	if err := r.db.QueryRow(ctx, query, opportunityID).Scan(&stats.OpportunityID,
		&stats.Title, &stats.Capacity, &stats.Applications, &stats.Withdrawals,
		&stats.Saved, &forwarded, &hoursOpen); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("opportunity analytics: %w", err)
	}

	if stats.Applications > 0 {
		stats.AcceptanceRate = round2(float64(forwarded) / float64(stats.Applications))
	}
	stats.AvgOpenHours = round2(hoursOpen)
	return stats, nil
}
