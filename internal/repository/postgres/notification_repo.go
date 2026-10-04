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

// NotificationRepository writes the notifications members receive and owns the
// per-member preference switches that decide whether they hear about a category
// at all.
type NotificationRepository struct {
	dbBase
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{dbBase: dbBase{db: db}}
}

// NotifyUser delivers one notification. Members who switched the category off
// are skipped inside the query rather than by a prior read, so a preference
// change takes effect immediately.
func (r *NotificationRepository) NotifyUser(ctx context.Context, userID string, ntype domain.NotificationType, title, body, link string) error {
	if err := r.ready(); err != nil {
		return err
	}

	const query = `INSERT INTO notifications (user_id, type, title, body, link)
		SELECT $1, $2, $3, $4, NULLIF($5, '')
		WHERE NOT EXISTS (
			SELECT 1 FROM notification_preferences p
			WHERE p.user_id = $1
				AND (
					($2 = 'EVENT' AND p.event_reminders IS NOT TRUE) OR
					($2 = 'OPPORTUNITY' AND p.opportunity_deadlines IS NOT TRUE) OR
					($2 = 'SERVICE' AND p.service_updates IS NOT TRUE) OR
					($2 = 'ANNOUNCEMENT' AND p.announcements IS NOT TRUE) OR
					($2 = 'CLUB' AND p.club_activity IS NOT TRUE)
				)
		)`

	if _, err := r.db.Exec(ctx, query, userID, string(ntype), title, body, link); err != nil {
		if isForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		return fmt.Errorf("notify user: %w", err)
	}
	return nil
}

// NotifyUsers delivers the same notification to many members in one statement,
// which is what an announcement or a club update needs.
func (r *NotificationRepository) NotifyUsers(ctx context.Context, userIDs []string, ntype domain.NotificationType, title, body, link string) (int, error) {
	if err := r.ready(); err != nil {
		return 0, err
	}
	if len(userIDs) == 0 {
		return 0, nil
	}

	const query = `INSERT INTO notifications (user_id, type, title, body, link)
		SELECT u.id, $1, $2, $3, NULLIF($4, '')
		FROM users u
		WHERE u.id = ANY($5) AND u.status = 'ACTIVE'
			AND NOT EXISTS (
				SELECT 1 FROM notification_preferences p
				WHERE p.user_id = u.id
					AND (
						($1 = 'EVENT' AND p.event_reminders IS NOT TRUE) OR
						($1 = 'OPPORTUNITY' AND p.opportunity_deadlines IS NOT TRUE) OR
						($1 = 'SERVICE' AND p.service_updates IS NOT TRUE) OR
						($1 = 'ANNOUNCEMENT' AND p.announcements IS NOT TRUE) OR
						($1 = 'CLUB' AND p.club_activity IS NOT TRUE)
					)
			)`

	tag, err := r.db.Exec(ctx, query, string(ntype), title, body, link, userIDs)
	if err != nil {
		return 0, fmt.Errorf("notify users: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

// NotifyClubFollowers tells everyone following a club that it posted something.
func (r *NotificationRepository) NotifyClubFollowers(ctx context.Context, clubID, title, body, link string) (int, error) {
	if err := r.ready(); err != nil {
		return 0, err
	}
	const query = `INSERT INTO notifications (user_id, type, title, body, link)
		SELECT f.user_id, 'CLUB', $2, $3, NULLIF($4, '')
		FROM club_follows f
		WHERE f.club_id = $1
			AND NOT EXISTS (
				SELECT 1 FROM notification_preferences p
				WHERE p.user_id = f.user_id AND p.club_activity IS NOT TRUE
			)`

	tag, err := r.db.Exec(ctx, query, clubID, title, body, link)
	if err != nil {
		return 0, fmt.Errorf("notify club followers: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

// NotifyEventAttendees tells everyone with a place what changed about an event,
// which is how a cancellation or a room change reaches people who did not open
// the app again.
func (r *NotificationRepository) NotifyEventAttendees(ctx context.Context, eventID, title, body, link string) (int, error) {
	if err := r.ready(); err != nil {
		return 0, err
	}
	const query = `INSERT INTO notifications (user_id, type, title, body, link)
		SELECT r.member_id, 'EVENT', $2, $3, NULLIF($4, '')
		FROM event_registrations r
		WHERE r.event_id = $1 AND r.status IN ('REGISTERED', 'WAITLISTED')
			AND NOT EXISTS (
				SELECT 1 FROM notification_preferences p
				WHERE p.user_id = r.member_id AND p.event_reminders IS NOT TRUE
			)`

	tag, err := r.db.Exec(ctx, query, eventID, title, body, link)
	if err != nil {
		return 0, fmt.Errorf("notify attendees: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

// DispatchDueEventReminders materialises reminders whose time has arrived. The
// reminders are deleted in the same transaction as the notifications are written,
// so a reminder fires exactly once even if the job runs twice.
func (r *NotificationRepository) DispatchDueEventReminders(ctx context.Context, limit int) (int, error) {
	if err := r.ready(); err != nil {
		return 0, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin reminder dispatch: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const due = `SELECT er.user_id, er.event_id
		FROM event_reminders er
		JOIN events e ON e.id = er.event_id
		WHERE er.remind_at <= now() AND e.status = 'PUBLISHED'
		ORDER BY er.remind_at ASC
		LIMIT $1
		FOR UPDATE OF er SKIP LOCKED`

	rows, err := tx.Query(ctx, due, limit)
	if err != nil {
		return 0, fmt.Errorf("read due reminders: %w", err)
	}

	type dueReminder struct {
		userID  string
		eventID string
	}
	pending := make([]dueReminder, 0, limit)
	for rows.Next() {
		var reminder dueReminder
		if err := rows.Scan(&reminder.userID, &reminder.eventID); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan due reminder: %w", err)
		}
		pending = append(pending, reminder)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("read due reminders: %w", err)
	}

	const insert = `INSERT INTO notifications (user_id, type, title, body, link)
		SELECT $1, 'EVENT', $2, $3, '/events/' || $4
		WHERE COALESCE((
			SELECT p.event_reminders FROM notification_preferences p WHERE p.user_id = $1
		), true) IS NOT FALSE`

	dispatched := 0
	for _, reminder := range pending {
		// The body is rebuilt from the event row read here rather than from the
		// reminder row, so a room change between scheduling and firing is
		// reflected in the reminder the member actually sees.
		var title, venue string
		var startTime time.Time
		if err := tx.QueryRow(ctx,
			`SELECT title, start_time, COALESCE(venue, '') FROM events WHERE id = $1`,
			reminder.eventID).Scan(&title, &startTime, &venue); err != nil {
			return 0, fmt.Errorf("read event for reminder: %w", err)
		}

		body := fmt.Sprintf("Starts %s%s", startTime.Format("Mon 2 Jan, 15:04"),
			conditional(" at "+venue, venue != ""))

		// A member who switched reminders off still has the reminder cleared: it
		// has been considered, and leaving it queued would re-check it forever.
		tag, err := tx.Exec(ctx, insert, reminder.userID,
			"Reminder: "+title, body, reminder.eventID)
		if err != nil {
			return 0, fmt.Errorf("insert reminder notification: %w", err)
		}

		if _, err := tx.Exec(ctx,
			`DELETE FROM event_reminders WHERE user_id = $1 AND event_id = $2`,
			reminder.userID, reminder.eventID); err != nil {
			return 0, fmt.Errorf("clear dispatched reminder: %w", err)
		}
		dispatched += int(tag.RowsAffected())
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit reminder dispatch: %w", err)
	}
	return dispatched, nil
}

// NotificationPreferences is the member's per-category switch set.
type NotificationPreferences struct {
	DigestFrequency      string `json:"digest_frequency"`
	EmailEnabled         bool   `json:"email_enabled"`
	PushEnabled          bool   `json:"push_enabled"`
	EventReminders       bool   `json:"event_reminders"`
	OpportunityDeadlines bool   `json:"opportunity_deadlines"`
	ApplicationUpdates   bool   `json:"application_updates"`
	ClubActivity         bool   `json:"club_activity"`
	ServiceUpdates       bool   `json:"service_updates"`
	Announcements        bool   `json:"announcements"`
}

// UpdateNotificationPreferencesRequest patches the switch set. Pointers
// distinguish "leave unchanged" from "switch off".
type UpdateNotificationPreferencesRequest struct {
	DigestFrequency      *string `json:"digest_frequency,omitempty" binding:"omitempty,oneof=INSTANT DAILY WEEKLY OFF"`
	EmailEnabled         *bool   `json:"email_enabled,omitempty"`
	PushEnabled          *bool   `json:"push_enabled,omitempty"`
	EventReminders       *bool   `json:"event_reminders,omitempty"`
	OpportunityDeadlines *bool   `json:"opportunity_deadlines,omitempty"`
	ApplicationUpdates   *bool   `json:"application_updates,omitempty"`
	ClubActivity         *bool   `json:"club_activity,omitempty"`
	ServiceUpdates       *bool   `json:"service_updates,omitempty"`
	Announcements        *bool   `json:"announcements,omitempty"`
}

// GetPreferences reads the member's switches, defaulting to everything on when
// they have never changed anything.
func (r *NotificationRepository) GetPreferences(ctx context.Context, userID string) (*NotificationPreferences, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT digest_frequency, COALESCE(email_enabled, true),
			COALESCE(push_enabled, true), event_reminders, opportunity_deadlines,
			application_updates, club_activity, service_updates, announcements
		FROM notification_preferences WHERE user_id = $1`

	prefs := &NotificationPreferences{}
	err := r.db.QueryRow(ctx, query, userID).Scan(&prefs.DigestFrequency,
		&prefs.EmailEnabled, &prefs.PushEnabled, &prefs.EventReminders,
		&prefs.OpportunityDeadlines, &prefs.ApplicationUpdates, &prefs.ClubActivity,
		&prefs.ServiceUpdates, &prefs.Announcements)
	if err == nil {
		return prefs, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("notification preferences: %w", err)
	}

	// No row yet: report the documented defaults rather than failing, so the
	// settings screen renders for a brand new member.
	return &NotificationPreferences{
		DigestFrequency:      "INSTANT",
		EmailEnabled:         true,
		PushEnabled:          true,
		EventReminders:       true,
		OpportunityDeadlines: true,
		ApplicationUpdates:   true,
		ClubActivity:         true,
		ServiceUpdates:       true,
		Announcements:        true,
	}, nil
}

// UpdatePreferences writes the switches, inserting the row on first use.
func (r *NotificationRepository) UpdatePreferences(ctx context.Context, userID string, req *UpdateNotificationPreferencesRequest) (*NotificationPreferences, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	if req.DigestFrequency == nil && req.EmailEnabled == nil && req.PushEnabled == nil &&
		req.EventReminders == nil && req.OpportunityDeadlines == nil &&
		req.ApplicationUpdates == nil && req.ClubActivity == nil &&
		req.ServiceUpdates == nil && req.Announcements == nil {
		return nil, fmt.Errorf("no preference was provided")
	}

	const query = `INSERT INTO notification_preferences (user_id, digest_frequency,
			email_enabled, push_enabled, event_reminders, opportunity_deadlines,
			application_updates, club_activity, service_updates, announcements)
		VALUES ($1, COALESCE($2, 'INSTANT'),
			COALESCE($3, true), COALESCE($4, true), COALESCE($5, true),
			COALESCE($6, true), COALESCE($7, true), COALESCE($8, true),
			COALESCE($9, true), COALESCE($10, true))
		ON CONFLICT (user_id) DO UPDATE SET
			digest_frequency = COALESCE($2, notification_preferences.digest_frequency),
			email_enabled = COALESCE($3, notification_preferences.email_enabled),
			push_enabled = COALESCE($4, notification_preferences.push_enabled),
			event_reminders = COALESCE($5, notification_preferences.event_reminders),
			opportunity_deadlines = COALESCE($6, notification_preferences.opportunity_deadlines),
			application_updates = COALESCE($7, notification_preferences.application_updates),
			club_activity = COALESCE($8, notification_preferences.club_activity),
			service_updates = COALESCE($9, notification_preferences.service_updates),
			announcements = COALESCE($10, notification_preferences.announcements)`

	if _, err := r.db.Exec(ctx, query, userID, req.DigestFrequency, req.EmailEnabled,
		req.PushEnabled, req.EventReminders, req.OpportunityDeadlines,
		req.ApplicationUpdates, req.ClubActivity, req.ServiceUpdates,
		req.Announcements); err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		return nil, fmt.Errorf("update notification preferences: %w", err)
	}

	return r.GetPreferences(ctx, userID)
}

// conditional returns value when cond holds, so notification bodies read
// naturally without a branch at every call site.
func conditional(value string, cond bool) string {
	if cond {
		return value
	}
	return ""
}
