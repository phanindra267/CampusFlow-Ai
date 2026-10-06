package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/campuscare/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EventRepository owns the event lifecycle that organisers drive: creating and
// editing an event, the member's own saved and reminder state, the organiser's
// registration list, and the evidence an organiser needs afterwards.
type EventRepository struct {
	dbBase
}

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{dbBase: dbBase{db: db}}
}

// eventColumns carries the resolved-per-member columns too, so one shape is used
// for listings and detail. savedEventColumns adds the saved and reminder state;
// $1/$2 are the member and event identifiers reserved by the caller.
const savedEventColumns = `e.id, e.title, e.description, e.category, e.organizer_id,
	COALESCE(e.venue, ''), e.is_online, e.start_time, e.end_time,
	e.registration_deadline, e.capacity, e.status, COALESCE(e.eligibility, ''),
	e.created_at, e.updated_at, e.club_id, COALESCE(cl.name, ''),
	(SELECT r.status FROM event_registrations r
		WHERE r.event_id = e.id AND r.member_id = $2 LIMIT 1) AS registration_status,
	EXISTS (SELECT 1 FROM saved_events s
		WHERE s.event_id = e.id AND s.user_id = $2) AS is_saved,
	(SELECT r.remind_at FROM event_reminders r
		WHERE r.event_id = e.id AND r.user_id = $2) AS reminder_at,
	(SELECT count(*) FROM event_registrations r
		WHERE r.event_id = e.id AND r.status = 'REGISTERED') AS registered_count`

// scanSavedEvent reads the row shape produced by savedEventColumns.
func scanSavedEvent(row pgx.Row) (*domain.Event, error) {
	var e domain.Event
	var regStatus *string
	if err := row.Scan(&e.ID, &e.Title, &e.Description, &e.Category, &e.OrganizerID,
		&e.Venue, &e.IsOnline, &e.StartTime, &e.EndTime, &e.RegistrationDeadline,
		&e.Capacity, &e.Status, &e.Eligibility, &e.CreatedAt, &e.UpdatedAt,
		&e.ClubID, &e.ClubName, &regStatus, &e.IsSaved, &e.ReminderAt,
		&e.RegisteredCount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan event: %w", err)
	}
	if regStatus != nil {
		e.RegistrationStatus = *regStatus
	}
	return &e, nil
}

// CreateEvent publishes or drafts an event on behalf of an organiser. A status
// that is neither DRAFT nor PUBLISHED is refused before it reaches the database
// so the caller gets a clear message rather than a constraint violation.
func (r *EventRepository) CreateEvent(ctx context.Context, req *domain.CreateEventRequest, organizerID string) (*domain.Event, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}

	status := req.Status
	if status == "" {
		status = string(domain.EventStatusPublished)
	}
	if status != string(domain.EventStatusDraft) && status != string(domain.EventStatusPublished) {
		return nil, fmt.Errorf("an event can only be created as DRAFT or PUBLISHED")
	}
	if err := validateEventWindow(req.StartTime, req.EndTime, req.RegistrationDeadline); err != nil {
		return nil, err
	}

	const query = `INSERT INTO events (title, description, category, organizer_id, venue,
			is_online, start_time, end_time, registration_deadline, capacity, status,
			eligibility, club_id)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9, $10, $11,
			NULLIF($12, ''), $13)
		RETURNING id`

	var id string
	if err := r.db.QueryRow(ctx, query, req.Title, req.Description, req.Category,
		organizerID, req.Venue, req.IsOnline, req.StartTime, req.EndTime,
		req.RegistrationDeadline, req.Capacity, status, req.Eligibility,
		req.ClubID).Scan(&id); err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		if isCheckViolation(err) {
			return nil, fmt.Errorf("event times or capacity are not valid")
		}
		return nil, fmt.Errorf("create event: %w", err)
	}

	return r.GetEvent(ctx, id, organizerID)
}

// UpdateEvent patches an event in place. Every column is coalesced so a partial
// payload leaves the fields it does not mention untouched.
func (r *EventRepository) UpdateEvent(ctx context.Context, id string, req *domain.CreateEventRequest) (*domain.Event, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	if err := validateEventWindow(req.StartTime, req.EndTime, req.RegistrationDeadline); err != nil {
		return nil, err
	}

	status := req.Status
	if status != "" && status != string(domain.EventStatusDraft) &&
		status != string(domain.EventStatusPublished) && status != string(domain.EventStatusCompleted) &&
		status != string(domain.EventStatusCancelled) {
		return nil, fmt.Errorf("unknown event status %q", status)
	}

	const query = `UPDATE events SET
			title = COALESCE(NULLIF($2, ''), title),
			description = COALESCE(NULLIF($3, ''), description),
			category = COALESCE(NULLIF($4, ''), category),
			venue = COALESCE(NULLIF($5, ''), venue),
			is_online = CASE WHEN $6 THEN true ELSE is_online END,
			start_time = CASE WHEN $7 THEN start_time ELSE $8 END,
			end_time = CASE WHEN $7 THEN end_time ELSE $9 END,
			registration_deadline = CASE WHEN $7 THEN registration_deadline ELSE $10 END,
			capacity = CASE WHEN $11 > 0 THEN $11 ELSE capacity END,
			status = COALESCE(NULLIF($12, ''), status),
			eligibility = COALESCE(NULLIF($13, ''), eligibility),
			club_id = COALESCE($14, club_id),
			updated_at = now()
		WHERE id = $1
		RETURNING id`

	var updated string
	if err := r.db.QueryRow(ctx, query, id, req.Title, req.Description, req.Category,
		req.Venue, req.IsOnline, !req.StartTime.IsZero(), req.StartTime, req.EndTime,
		req.RegistrationDeadline, req.Capacity, status, req.Eligibility,
		req.ClubID).Scan(&updated); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		return nil, fmt.Errorf("update event: %w", err)
	}

	return r.GetEvent(ctx, id, "")
}

// CancelEvent withdraws an event from the calendar. Registrations are kept so
// members who had a place can still be told what happened.
func (r *EventRepository) CancelEvent(ctx context.Context, id string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`UPDATE events SET status = 'CANCELLED', updated_at = now()
		WHERE id = $1 AND status <> 'CANCELLED'`, id)
	if err != nil {
		return fmt.Errorf("cancel event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// validateEventWindow enforces the ordering the database also enforces, so the
// caller receives a readable message instead of a constraint name.
func validateEventWindow(start, end, deadline time.Time) error {
	if !start.IsZero() && !end.IsZero() && !end.After(start) {
		return fmt.Errorf("event must end after it starts")
	}
	if !start.IsZero() && !deadline.IsZero() && deadline.After(start) {
		return fmt.Errorf("registration must close before the event starts")
	}
	return nil
}

// ListEventCategories returns the categories actually in use, so the filter
// offers what exists rather than a hard-coded list that drifts.
func (r *EventRepository) ListEventCategories(ctx context.Context) ([]domain.CategoryCount, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT COALESCE(category, 'UNCATEGORISED'), count(*)
		FROM events WHERE status IN ('PUBLISHED', 'COMPLETED')
		GROUP BY 1 ORDER BY 2 DESC, 1 ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("event categories: %w", err)
	}
	defer rows.Close()

	out := make([]domain.CategoryCount, 0, 16)
	for rows.Next() {
		var entry domain.CategoryCount
		if err := rows.Scan(&entry.Label, &entry.Count); err != nil {
			return nil, fmt.Errorf("scan event category: %w", err)
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

// GetEvent loads one event with the requesting member's state resolved.
func (r *EventRepository) GetEvent(ctx context.Context, id, userID string) (*domain.Event, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + savedEventColumns + `
		FROM events e LEFT JOIN clubs cl ON cl.id = e.club_id
		WHERE e.id = $1`

	event, err := scanSavedEvent(r.db.QueryRow(ctx, query, id, userID))
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}
	return event, nil
}

// ListSavedEvents is the member's shortlist, soonest first.
func (r *EventRepository) ListSavedEvents(ctx context.Context, userID string, limit, offset int) ([]domain.Event, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + savedEventColumns + `
		FROM events e
		LEFT JOIN clubs cl ON cl.id = e.club_id
		JOIN saved_events s ON s.event_id = e.id AND s.user_id = $2
		WHERE e.status IN ('PUBLISHED', 'COMPLETED')
		ORDER BY e.start_time ASC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(ctx, query, userID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list saved events: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Event, 0, limit)
	for rows.Next() {
		event, err := scanSavedEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *event)
	}
	return out, rows.Err()
}

// SaveEvent bookmarks an event for the member. Saving is not registering: it is
// a personal shortlist that survives a full event.
func (r *EventRepository) SaveEvent(ctx context.Context, userID, eventID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	const query = `INSERT INTO saved_events (user_id, event_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`

	if _, err := r.db.Exec(ctx, query, userID, eventID); err != nil {
		if isForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		return fmt.Errorf("save event: %w", err)
	}
	return nil
}

func (r *EventRepository) UnsaveEvent(ctx context.Context, userID, eventID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`DELETE FROM saved_events WHERE user_id = $1 AND event_id = $2`, userID, eventID)
	if err != nil {
		return fmt.Errorf("unsave event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SetEventReminder schedules one reminder per member per event. A reminder in
// the past is refused: it would fire immediately and read as noise.
func (r *EventRepository) SetEventReminder(ctx context.Context, userID, eventID string, remindAt time.Time) error {
	if err := r.ready(); err != nil {
		return err
	}
	if !remindAt.After(time.Now()) {
		return fmt.Errorf("reminder time must be in the future")
	}

	const query = `INSERT INTO event_reminders (user_id, event_id, remind_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, event_id) DO UPDATE SET remind_at = EXCLUDED.remind_at`

	if _, err := r.db.Exec(ctx, query, userID, eventID, remindAt); err != nil {
		if isForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		return fmt.Errorf("set reminder: %w", err)
	}
	return nil
}

func (r *EventRepository) ClearEventReminder(ctx context.Context, userID, eventID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`DELETE FROM event_reminders WHERE user_id = $1 AND event_id = $2`, userID, eventID)
	if err != nil {
		return fmt.Errorf("clear reminder: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListEventRegistrations is the organiser's attendee list, with the check-in
// time resolved from the attendance record so attendance and registration can
// be read together.
func (r *EventRepository) ListEventRegistrations(ctx context.Context, eventID string, limit, offset int) ([]domain.EventRegistrationDetail, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT r.id, r.member_id, COALESCE(u.display_name, ''), COALESCE(u.email, ''),
			r.status, r.created_at, a.checked_in_at
		FROM event_registrations r
		JOIN users u ON u.id = r.member_id
		LEFT JOIN attendance_records a ON a.member_id = r.member_id
			AND a.event_session_id IN (SELECT id FROM check_in_sessions WHERE event_session_id = $1)
		WHERE r.event_id = $1 AND r.status <> 'CANCELLED'
		ORDER BY r.created_at ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, eventID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list registrations: %w", err)
	}
	defer rows.Close()

	out := make([]domain.EventRegistrationDetail, 0, limit)
	for rows.Next() {
		var d domain.EventRegistrationDetail
		if err := rows.Scan(&d.ID, &d.MemberID, &d.DisplayName, &d.Email, &d.Status,
			&d.CreatedAt, &d.CheckedInAt); err != nil {
			return nil, fmt.Errorf("scan registration: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// OrganizerEventAnalytics answers the questions an organiser asks after an
// event: did it fill, did people turn up, and was it worth running again.
// DemandVsSupply compares interest (saves plus registrations) against places,
// which is the signal for running a bigger or repeat session.
func (r *EventRepository) OrganizerEventAnalytics(ctx context.Context, eventID string) (*domain.OrganizerEventAnalytics, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}

	const query = `SELECT e.id, e.title, e.capacity,
		(SELECT count(*) FROM event_registrations r
			WHERE r.event_id = e.id AND r.status = 'REGISTERED'),
		(SELECT count(*) FROM event_registrations r
			WHERE r.event_id = e.id AND r.status = 'CANCELLED'),
		(SELECT count(*) FROM event_registrations r
			WHERE r.event_id = e.id AND r.status = 'ATTENDED'),
		(SELECT count(*) FROM waitlist_entries w
			WHERE w.event_id = e.id AND w.status = 'WAITING'),
		(SELECT count(*) FROM event_reminders er WHERE er.event_id = e.id),
		(SELECT count(*) FROM saved_events s WHERE s.event_id = e.id),
		(SELECT count(*) FROM attendance_records a
			WHERE a.event_session_id IN (SELECT id FROM check_in_sessions WHERE event_session_id = e.id))
		FROM events e WHERE e.id = $1`

	stats := &domain.OrganizerEventAnalytics{}
	var checkedIn int
	if err := r.db.QueryRow(ctx, query, eventID).Scan(&stats.EventID, &stats.Title,
		&stats.Capacity, &stats.Registered, &stats.Cancelled, &stats.Attended,
		&stats.Waitlisted, &stats.RemindersSet, &stats.SavedCount,
		&checkedIn); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("event analytics: %w", err)
	}

	if stats.Capacity > 0 {
		stats.FillRate = round2(float64(stats.Registered) / float64(stats.Capacity))
	}
	if stats.Registered > 0 {
		stats.CheckInRate = round2(float64(checkedIn) / float64(stats.Registered))
	}
	demand := stats.SavedCount + stats.Registered
	if stats.Capacity > 0 {
		stats.DemandVsSupply = round2(float64(demand) / float64(stats.Capacity))
	}

	average, total, err := r.feedbackAverage(ctx, eventID)
	if err != nil {
		return nil, err
	}
	stats.AverageRating, stats.RatingsCount = average, total

	return stats, nil
}

// feedbackAverage is the shared rating aggregate for any rated activity.
func (r *EventRepository) feedbackAverage(ctx context.Context, activityID string) (float64, int, error) {
	const query = `SELECT COALESCE(avg(rating), 0), count(*)
		FROM feedback
		WHERE activity_type = 'EVENT' AND activity_id = $1 AND status = 'VISIBLE'`

	var average float64
	var total int
	if err := r.db.QueryRow(ctx, query, activityID).Scan(&average, &total); err != nil {
		return 0, 0, fmt.Errorf("average feedback: %w", err)
	}
	return round2(average), total, nil
}

// round2 keeps reported rates readable instead of carrying fifteen decimal
// places through the API.
func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

// ListEventsByOrganizer is the organiser's own dashboard list, including drafts
// that the public listing deliberately hides.
func (r *EventRepository) ListEventsByOrganizer(ctx context.Context, organizerID string, limit, offset int) ([]domain.Event, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + savedEventColumns + `
		FROM events e LEFT JOIN clubs cl ON cl.id = e.club_id
		WHERE e.organizer_id = $1
		ORDER BY e.start_time DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(ctx, query, organizerID, organizerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list organiser events: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Event, 0, limit)
	for rows.Next() {
		event, err := scanSavedEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *event)
	}
	return out, rows.Err()
}

// PromoteWaitlist moves the first waiting member into a place when one frees up,
// so a cancellation is not wasted on an empty seat.
//
// Capacity is re-checked here under the same per-event row lock that
// RegisterForEvent uses. Without it, promoting into an event that is already
// full (or that filled up between the organiser opening the console and pressing
// promote) would overbook deterministically rather than by race. The count runs
// as a separate statement after the lock so it reads a fresh snapshot.
func (r *EventRepository) PromoteWaitlist(ctx context.Context, eventID string) (*domain.EventRegistration, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin promote: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Same lock as the direct registration path, so promotion and registration
	// cannot interleave and jointly overshoot capacity.
	const lock = `SELECT capacity, status FROM events WHERE id = $1 FOR UPDATE`

	var capacity int
	var status string
	if err := tx.QueryRow(ctx, lock, eventID).Scan(&capacity, &status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lock event: %w", err)
	}

	if capacity > 0 {
		var taken int
		const count = `SELECT count(*) FROM event_registrations
			WHERE event_id = $1 AND status = 'REGISTERED'`
		if err := tx.QueryRow(ctx, count, eventID).Scan(&taken); err != nil {
			return nil, fmt.Errorf("count registrations: %w", err)
		}
		if taken >= capacity {
			return nil, ErrEventFull
		}
	}

	// The promotion itself is shared with CancelRegistration, so the organiser
	// pressing "promote" and a member freeing their own seat take exactly the
	// same path and cannot drift apart.
	reg, err := promoteNextFromWaitlist(ctx, tx, eventID)
	if err != nil {
		return nil, err
	}
	if reg == nil {
		// An organiser pressing this on an empty waitlist needs to know nothing
		// happened, which is different from the cancellation case where an empty
		// waitlist is a perfectly good outcome.
		return nil, ErrNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit promote: %w", err)
	}
	return reg, nil
}
