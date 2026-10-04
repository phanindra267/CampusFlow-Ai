package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/campuscare/api/internal/domain"
	"github.com/jackc/pgx/v5"
)

// profileColumns extends the base user columns with the optional, self-reported
// profile context. These fields describe who a member is and what they want to
// do; there are no marks, credits or attendance columns here.
const profileColumns = `id, email, display_name, password_hash, role, status,
	created_at, updated_at, COALESCE(identifier, ''), COALESCE(program, ''),
	COALESCE(branch, ''), year_of_study, COALESCE(headline, ''), COALESCE(bio, ''),
	COALESCE(avatar_url, ''), career_interests, open_to_opportunities`

func scanProfile(row pgx.Row) (*domain.User, error) {
	var user domain.User
	if err := row.Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.Identifier,
		&user.Program, &user.Branch, &user.YearOfStudy, &user.Headline, &user.Bio,
		&user.AvatarURL, &user.CareerInterests, &user.OpenToOpportunities); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan profile: %w", err)
	}
	return &user, nil
}

// GetProfile loads a member with their profile fields resolved.
func (r *UserRepository) GetProfile(ctx context.Context, id string) (*domain.User, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	user, err := scanProfile(r.db.QueryRow(ctx,
		`SELECT `+profileColumns+` FROM users WHERE id = $1`, id))
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return user, nil
}

// UpdateProfile patches the profile. display_name is required, so it is always
// written; every other field is coalesced, which means a field sent as an empty
// string clears it and a field left out is untouched.
func (r *UserRepository) UpdateProfile(ctx context.Context, id string, req *domain.UpdateProfileRequest) (*domain.User, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}

	const query = `UPDATE users SET
			display_name = $2,
			identifier = CASE WHEN $3 THEN NULLIF($4, '') ELSE identifier END,
			program = CASE WHEN $5 THEN NULLIF($6, '') ELSE program END,
			branch = CASE WHEN $7 THEN NULLIF($8, '') ELSE branch END,
			year_of_study = CASE WHEN $9 THEN $10 ELSE year_of_study END,
			headline = CASE WHEN $11 THEN NULLIF($12, '') ELSE headline END,
			bio = CASE WHEN $13 THEN NULLIF($14, '') ELSE bio END,
			avatar_url = CASE WHEN $15 THEN NULLIF($16, '') ELSE avatar_url END,
			career_interests = COALESCE($17, career_interests),
			open_to_opportunities = COALESCE($18, open_to_opportunities),
			updated_at = now()
		WHERE id = $1
		RETURNING id`

	if _, err := r.db.Exec(ctx, query, id, req.DisplayName,
		req.Identifier != nil, req.Identifier, req.Program != nil, req.Program,
		req.Branch != nil, req.Branch, req.YearOfStudy != nil, req.YearOfStudy,
		req.Headline != nil, req.Headline, req.Bio != nil, req.Bio,
		req.AvatarURL != nil, req.AvatarURL, req.CareerInterests,
		req.OpenToOpportunities); err != nil {
		if isCheckViolation(err) {
			return nil, fmt.Errorf("year of study must be between 1 and 10")
		}
		return nil, fmt.Errorf("update profile: %w", err)
	}

	return r.GetProfile(ctx, id)
}

// LoadProfile assembles the whole profile page: the member's own fields plus the
// clubs, saved items, registrations, applications, tickets and activity they have
// accumulated. It is one call because the page needs all of it, and a member
// editing their profile should not have to wait on eight round trips.
func (r *UserRepository) LoadProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}

	user, err := r.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile := &domain.Profile{User: *user}

	if profile.ResearchInterests, err = r.memberResearchInterests(ctx, userID); err != nil {
		return nil, err
	}
	if profile.Clubs, err = r.profileClubs(ctx, userID); err != nil {
		return nil, err
	}
	if profile.Following, err = r.profileFollowing(ctx, userID); err != nil {
		return nil, err
	}
	if profile.SavedEvents, err = r.profileSavedEvents(ctx, userID); err != nil {
		return nil, err
	}
	if profile.SavedOpportunities, err = r.profileSavedOpportunities(ctx, userID); err != nil {
		return nil, err
	}
	if profile.Applications, err = r.profileApplications(ctx, userID); err != nil {
		return nil, err
	}
	if profile.Registrations, err = r.profileRegistrations(ctx, userID); err != nil {
		return nil, err
	}
	if profile.ServiceRequests, err = r.profileServiceRequests(ctx, userID); err != nil {
		return nil, err
	}

	return profile, nil
}

func (r *UserRepository) memberResearchInterests(ctx context.Context, userID string) ([]string, error) {
	const query = `SELECT interest FROM user_research_interests
		WHERE user_id = $1 ORDER BY interest`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("profile research interests: %w", err)
	}
	defer rows.Close()

	out := make([]string, 0, 8)
	for rows.Next() {
		var interest string
		if err := rows.Scan(&interest); err != nil {
			return nil, fmt.Errorf("scan research interest: %w", err)
		}
		out = append(out, interest)
	}
	return out, rows.Err()
}

func (r *UserRepository) profileClubs(ctx context.Context, userID string) ([]domain.ProfileClub, error) {
	const query = `SELECT c.id, c.name, c.slug, COALESCE(c.category, ''),
			m.role, c.follower_count, c.member_count
		FROM clubs c
		JOIN club_memberships m ON m.club_id = c.id
			AND m.user_id = $1 AND m.status = 'ACTIVE'
		WHERE COALESCE(c.status, 'ACTIVE') = 'ACTIVE'
		ORDER BY c.name`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("profile clubs: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ProfileClub, 0, 8)
	for rows.Next() {
		var club domain.ProfileClub
		if err := rows.Scan(&club.ID, &club.Name, &club.Slug, &club.Category,
			&club.Role, &club.FollowerCount, &club.MemberCount); err != nil {
			return nil, fmt.Errorf("scan profile club: %w", err)
		}
		out = append(out, club)
	}
	return out, rows.Err()
}

func (r *UserRepository) profileFollowing(ctx context.Context, userID string) ([]domain.ProfileClub, error) {
	const query = `SELECT c.id, c.name, c.slug, COALESCE(c.category, ''),
			'', c.follower_count, c.member_count
		FROM clubs c
		JOIN club_follows f ON f.club_id = c.id AND f.user_id = $1
		WHERE COALESCE(c.status, 'ACTIVE') = 'ACTIVE'
		ORDER BY c.name`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("profile following: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ProfileClub, 0, 8)
	for rows.Next() {
		var club domain.ProfileClub
		if err := rows.Scan(&club.ID, &club.Name, &club.Slug, &club.Category,
			&club.Role, &club.FollowerCount, &club.MemberCount); err != nil {
			return nil, fmt.Errorf("scan followed club: %w", err)
		}
		out = append(out, club)
	}
	return out, rows.Err()
}

func (r *UserRepository) profileSavedEvents(ctx context.Context, userID string) ([]domain.ProfileSavedEvent, error) {
	const query = `SELECT e.id, e.title, COALESCE(e.category, ''), e.start_time,
			COALESCE(e.venue, '')
		FROM events e
		JOIN saved_events s ON s.event_id = e.id AND s.user_id = $1
		WHERE e.status IN ('PUBLISHED', 'COMPLETED')
		ORDER BY e.start_time DESC LIMIT 20`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("profile saved events: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ProfileSavedEvent, 0, 8)
	for rows.Next() {
		var event domain.ProfileSavedEvent
		if err := rows.Scan(&event.ID, &event.Title, &event.Category, &event.StartTime,
			&event.Venue); err != nil {
			return nil, fmt.Errorf("scan saved event: %w", err)
		}
		out = append(out, event)
	}
	return out, rows.Err()
}

func (r *UserRepository) profileSavedOpportunities(ctx context.Context, userID string) ([]domain.ProfileSavedOpportunity, error) {
	const query = `SELECT o.id, o.title, o.type, COALESCE(o.category, ''),
			o.registration_deadline
		FROM opportunities o
		JOIN saved_opportunities s ON s.opportunity_id = o.id AND s.user_id = $1
		WHERE o.status = 'OPEN'
		ORDER BY o.registration_deadline ASC NULLS LAST LIMIT 20`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("profile saved opportunities: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ProfileSavedOpportunity, 0, 8)
	for rows.Next() {
		var opportunity domain.ProfileSavedOpportunity
		if err := rows.Scan(&opportunity.ID, &opportunity.Title, &opportunity.Type,
			&opportunity.Category, &opportunity.Deadline); err != nil {
			return nil, fmt.Errorf("scan saved opportunity: %w", err)
		}
		out = append(out, opportunity)
	}
	return out, rows.Err()
}

func (r *UserRepository) profileApplications(ctx context.Context, userID string) ([]domain.OpportunityApplication, error) {
	const query = `SELECT a.id, a.opportunity_id, a.user_id, a.status,
			COALESCE(a.cover_note, ''), a.resume_url, a.applied_at, a.updated_at,
			COALESCE(o.title, '')
		FROM opportunity_applications a
		LEFT JOIN opportunities o ON o.id = a.opportunity_id
		WHERE a.user_id = $1
		ORDER BY a.applied_at DESC LIMIT 20`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("profile applications: %w", err)
	}
	defer rows.Close()

	out := make([]domain.OpportunityApplication, 0, 8)
	for rows.Next() {
		var application domain.OpportunityApplication
		if err := rows.Scan(&application.ID, &application.OpportunityID,
			&application.UserID, &application.Status, &application.CoverNote,
			&application.ResumeURL, &application.AppliedAt, &application.UpdatedAt,
			&application.Title); err != nil {
			return nil, fmt.Errorf("scan application: %w", err)
		}
		out = append(out, application)
	}
	return out, rows.Err()
}

func (r *UserRepository) profileRegistrations(ctx context.Context, userID string) ([]domain.ProfileRegistration, error) {
	const query = `SELECT r.id, r.event_id, COALESCE(e.title, ''),
			COALESCE(e.category, ''), e.start_time, COALESCE(e.venue, ''), r.status
		FROM event_registrations r
		JOIN events e ON e.id = r.event_id
		WHERE r.member_id = $1 AND r.status <> 'CANCELLED'
		ORDER BY e.start_time DESC LIMIT 20`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("profile registrations: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ProfileRegistration, 0, 8)
	for rows.Next() {
		var registration domain.ProfileRegistration
		if err := rows.Scan(&registration.ID, &registration.EventID,
			&registration.Title, &registration.Category, &registration.StartTime,
			&registration.Venue, &registration.Status); err != nil {
			return nil, fmt.Errorf("scan registration: %w", err)
		}
		out = append(out, registration)
	}
	return out, rows.Err()
}

func (r *UserRepository) profileServiceRequests(ctx context.Context, userID string) ([]domain.ServiceRequest, error) {
	const query = `SELECT sr.id, sr.reference, sr.resource_id, COALESCE(cr.name, ''),
			sr.user_id, COALESCE(u.display_name, ''), sr.kind, sr.subject,
			sr.description, sr.status, sr.priority, sr.assigned_to,
			COALESCE(assignee.display_name, ''), COALESCE(sr.resolution, ''),
			sr.resolved_at, sr.feedback_rating, COALESCE(sr.feedback_comment, ''),
			sr.created_at, sr.updated_at
		FROM service_requests sr
		JOIN users u ON u.id = sr.user_id
		LEFT JOIN campus_resources cr ON cr.id = sr.resource_id
		LEFT JOIN users assignee ON assignee.id = sr.assigned_to
		WHERE sr.user_id = $1
		ORDER BY sr.created_at DESC LIMIT 10`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("profile service requests: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ServiceRequest, 0, 8)
	for rows.Next() {
		var request domain.ServiceRequest
		if err := rows.Scan(&request.ID, &request.Reference, &request.ServiceID,
			&request.ServiceName, &request.UserID, &request.RequesterName,
			&request.Kind, &request.Subject, &request.Description, &request.Status,
			&request.Priority, &request.AssignedTo, &request.AssignedName,
			&request.Resolution, &request.ResolvedAt, &request.Rating,
			&request.RatingComment, &request.CreatedAt, &request.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan service request: %w", err)
		}
		out = append(out, request)
	}
	return out, rows.Err()
}
