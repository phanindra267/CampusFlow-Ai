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

// PeopleRepository serves the campus people directory: the faculty, researchers,
// supervisors, mentors and coordinators a member can find when they need
// guidance or want to work with someone on research.
//
// Directory entries are managed by administrators rather than self-registered,
// because the directory names real staff. An entry may be linked to a
// CampusCare account, but does not have to be.
type PeopleRepository struct {
	dbBase
}

func NewPeopleRepository(db *pgxpool.Pool) *PeopleRepository {
	return &PeopleRepository{dbBase: dbBase{db: db}}
}

const personColumns = `p.id, p.user_id, p.display_name, p.person_role,
	COALESCE(p.school, ''), COALESCE(p.designation, ''), COALESCE(p.bio, ''),
	COALESCE(p.email, ''), COALESCE(p.office_location, ''), COALESCE(p.profile_url, ''),
	p.research_interests, p.accepting_students, p.status, p.created_at, p.updated_at`

func scanPerson(row pgx.Row) (*domain.Person, error) {
	var person domain.Person
	if err := row.Scan(&person.ID, &person.UserID, &person.DisplayName, &person.Role,
		&person.School, &person.Designation, &person.Bio, &person.Email,
		&person.OfficeLocation, &person.ProfileURL, &person.ResearchInterests,
		&person.AcceptingStudents, &person.Status, &person.CreatedAt,
		&person.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan person: %w", err)
	}
	return &person, nil
}

// ListPeople searches the directory. Search is a full-text match over the name
// and bio, with an optional interest filter, because the reason someone looks up
// a supervisor is usually a topic rather than a name they already know.
func (r *PeopleRepository) ListPeople(ctx context.Context, search, role, school, interest string, acceptingOnly bool, limit, offset int) ([]domain.Person, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + personColumns + `
		FROM people p
		WHERE p.status = 'ACTIVE'
			AND ($1 = '' OR p.display_name ILIKE '%' || $1 || '%'
				OR COALESCE(p.bio, '') ILIKE '%' || $1 || '%'
				OR COALESCE(p.designation, '') ILIKE '%' || $1 || '%')
			AND ($2 = '' OR p.person_role = $2)
			AND ($3 = '' OR p.school = $3)
			AND ($4 = '' OR $4 = ANY(p.research_interests))
			AND (NOT $5 OR p.accepting_students)
		ORDER BY p.accepting_students DESC, p.display_name ASC
		LIMIT $6 OFFSET $7`

	rows, err := r.db.Query(ctx, query, search, role, school, interest, acceptingOnly, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list people: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Person, 0, limit)
	for rows.Next() {
		person, err := scanPerson(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *person)
	}
	return out, rows.Err()
}

func (r *PeopleRepository) GetPerson(ctx context.Context, id string) (*domain.Person, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + personColumns + ` FROM people p WHERE p.id = $1`

	person, err := scanPerson(r.db.QueryRow(ctx, query, id))
	if err != nil {
		return nil, fmt.Errorf("get person: %w", err)
	}
	return person, nil
}

// CreatePerson adds a directory entry.
func (r *PeopleRepository) CreatePerson(ctx context.Context, req *domain.CreatePersonRequest) (*domain.Person, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `INSERT INTO people (user_id, display_name, person_role, school,
			designation, bio, email, office_location, profile_url,
			research_interests, accepting_students)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''),
			NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), $10, $11)
		RETURNING id`

	var id string
	if err := r.db.QueryRow(ctx, query, req.UserID, req.DisplayName, req.Role,
		req.School, req.Designation, req.Bio, req.Email, req.OfficeLocation,
		req.ProfileURL, req.ResearchInterests, req.AcceptingStudents).Scan(&id); err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		if isUniqueViolation(err) {
			return nil, ErrDuplicateKey
		}
		if isCheckViolation(err) {
			return nil, fmt.Errorf("unknown person role %q", req.Role)
		}
		return nil, fmt.Errorf("create person: %w", err)
	}
	return r.GetPerson(ctx, id)
}

// UpdatePerson patches a directory entry. Nil fields are left alone; an empty
// slice clears the interest list, which is how an entry is corrected.
func (r *PeopleRepository) UpdatePerson(ctx context.Context, id string, req *domain.UpdatePersonRequest) (*domain.Person, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `UPDATE people SET
			display_name = COALESCE($2, display_name),
			person_role = COALESCE($3, person_role),
			school = COALESCE($4, school),
			designation = COALESCE($5, designation),
			bio = COALESCE($6, bio),
			email = COALESCE($7, email),
			office_location = COALESCE($8, office_location),
			profile_url = COALESCE($9, profile_url),
			research_interests = COALESCE($10, research_interests),
			accepting_students = COALESCE($11, accepting_students),
			status = COALESCE($12, status),
			updated_at = now()
		WHERE id = $1
		RETURNING id`

	var updated string
	if err := r.db.QueryRow(ctx, query, id, req.DisplayName, req.Role, req.School,
		req.Designation, req.Bio, req.Email, req.OfficeLocation, req.ProfileURL,
		req.ResearchInterests, req.AcceptingStudents, req.Status).Scan(&updated); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		if isCheckViolation(err) {
			return nil, fmt.Errorf("unknown person role or status")
		}
		return nil, fmt.Errorf("update person: %w", err)
	}
	return r.GetPerson(ctx, id)
}

// ListSchools returns the schools represented in the directory so the browse
// filter offers real values.
func (r *PeopleRepository) ListSchools(ctx context.Context) ([]domain.CategoryCount, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT COALESCE(NULLIF(school, ''), 'UNSPECIFIED'), count(*)
		FROM people WHERE status = 'ACTIVE'
		GROUP BY 1 ORDER BY 2 DESC, 1 ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list schools: %w", err)
	}
	defer rows.Close()

	out := make([]domain.CategoryCount, 0, 16)
	for rows.Next() {
		var entry domain.CategoryCount
		if err := rows.Scan(&entry.Label, &entry.Count); err != nil {
			return nil, fmt.Errorf("scan school: %w", err)
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

// ListResearchInterests returns the topics people in the directory work on,
// most common first, so a member can pick a topic rather than free-type one.
func (r *PeopleRepository) ListResearchInterests(ctx context.Context, limit int) ([]domain.CategoryCount, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT interest, count(*) AS people
		FROM people p, unnest(p.research_interests) AS interest
		WHERE p.status = 'ACTIVE'
		GROUP BY interest ORDER BY people DESC, interest ASC
		LIMIT $1`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("list research interests: %w", err)
	}
	defer rows.Close()

	out := make([]domain.CategoryCount, 0, limit)
	for rows.Next() {
		var entry domain.CategoryCount
		if err := rows.Scan(&entry.Label, &entry.Count); err != nil {
			return nil, fmt.Errorf("scan research interest: %w", err)
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

// FindPeopleByInterest is the collaboration entry point: everyone in the
// directory working on a topic, optionally only those taking students.
func (r *PeopleRepository) FindPeopleByInterest(ctx context.Context, interest string, acceptingOnly bool, limit int) ([]domain.Person, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + personColumns + `
		FROM people p
		WHERE p.status = 'ACTIVE'
			AND ($2 = ANY(p.research_interests))
			AND (NOT $3 OR p.accepting_students)
		ORDER BY p.display_name ASC LIMIT $1`

	rows, err := r.db.Query(ctx, query, limit, interest, acceptingOnly)
	if err != nil {
		return nil, fmt.Errorf("find people by interest: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Person, 0, limit)
	for rows.Next() {
		person, err := scanPerson(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *person)
	}
	return out, rows.Err()
}

// ------------------------------------------------ member research interests

// ListMemberResearchInterests is the signed-in member's own research topics.
func (r *PeopleRepository) ListMemberResearchInterests(ctx context.Context, userID string) ([]string, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT interest FROM user_research_interests
		WHERE user_id = $1 ORDER BY interest ASC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list member research interests: %w", err)
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

// AddMemberResearchInterest records a topic. Entries are stored trimmed so the
// same topic typed two ways does not create a duplicate the member has to clean
// up.
func (r *PeopleRepository) AddMemberResearchInterest(ctx context.Context, userID, interest string) error {
	if err := r.ready(); err != nil {
		return err
	}
	interest = strings.TrimSpace(interest)
	if interest == "" {
		return fmt.Errorf("a research interest cannot be blank")
	}

	const query = `INSERT INTO user_research_interests (user_id, interest)
		VALUES ($1, $2) ON CONFLICT DO NOTHING`

	if _, err := r.db.Exec(ctx, query, userID, interest); err != nil {
		if isForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		return fmt.Errorf("add research interest: %w", err)
	}
	return nil
}

func (r *PeopleRepository) RemoveMemberResearchInterest(ctx context.Context, userID, interest string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`DELETE FROM user_research_interests WHERE user_id = $1 AND interest = $2`,
		userID, strings.TrimSpace(interest))
	if err != nil {
		return fmt.Errorf("remove research interest: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// FindMembersByResearchInterest finds members who listed a topic. It is the
// student-side mirror of FindPeopleByInterest: a member looking for collaborators
// on a topic, or a project team to join. Only the fields a collaborator needs are
// returned, so this is not a member directory.
func (r *PeopleRepository) FindMembersByResearchInterest(ctx context.Context, userID, interest, program string, limit int) ([]domain.MemberMatch, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT u.id, COALESCE(u.display_name, ''),
			COALESCE(u.headline, ''), COALESCE(u.program, ''), COALESCE(u.branch, ''),
			u.year_of_study, COALESCE(u.avatar_url, ''), u.open_to_opportunities,
			ARRAY(SELECT i.interest FROM user_research_interests i
				WHERE i.user_id = u.id
					AND ($2 = ANY((SELECT j.interest FROM user_research_interests j
						WHERE j.user_id = $1)))
				ORDER BY i.interest)
		FROM users u
		WHERE u.id <> $1 AND u.status = 'ACTIVE'
			AND EXISTS (SELECT 1 FROM user_research_interests i
				WHERE i.user_id = u.id AND i.interest = $2)
			AND ($3 = '' OR COALESCE(u.program, '') = $3)
		ORDER BY u.display_name ASC LIMIT $4`

	rows, err := r.db.Query(ctx, query, userID, interest, program, limit)
	if err != nil {
		return nil, fmt.Errorf("find members by interest: %w", err)
	}
	defer rows.Close()

	out := make([]domain.MemberMatch, 0, limit)
	for rows.Next() {
		var m domain.MemberMatch
		if err := rows.Scan(&m.UserID, &m.DisplayName, &m.Headline, &m.Program,
			&m.Branch, &m.YearOfStudy, &m.AvatarURL, &m.OpenToOpportunities,
			&m.SharedInterests); err != nil {
			return nil, fmt.Errorf("scan member match: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
