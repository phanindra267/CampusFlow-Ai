package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/campuscare/api/internal/domain"
	"github.com/jackc/pgx/v5"
)

// promoteNextFromWaitlist moves the longest-waiting member of an event into a
// confirmed place and returns their registration. It is the single implementation
// of "a place freed up, so give it to whoever has waited longest".
//
// It returns (nil, nil) when nobody is waiting, which is an ordinary outcome
// rather than a failure: an event with an empty waitlist has simply freed a seat
// for whoever registers next. Callers that treat an empty waitlist as their own
// error translate the nil into it.
//
// The caller must already hold the per-event row lock taken by
// `SELECT ... FROM events WHERE id = $1 FOR UPDATE` and must run this inside the
// transaction that took it. That is what stops a cancellation from promoting a
// member into a place that a simultaneous registration just claimed.
func promoteNextFromWaitlist(ctx context.Context, tx pgx.Tx, eventID string) (*domain.EventRegistration, error) {
	// Lock the candidate row so two organisers promoting at once cannot pick the
	// same waitlist entry.
	const pick = `SELECT id, member_id FROM waitlist_entries
		WHERE event_id = $1 AND status = 'WAITING'
		ORDER BY joined_at ASC LIMIT 1 FOR UPDATE`

	var entryID, memberID string
	if err := tx.QueryRow(ctx, pick, eventID).Scan(&entryID, &memberID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("pick waitlist entry: %w", err)
	}

	// ON CONFLICT rather than a bare INSERT: a member who cancelled earlier still
	// has a registration row with status CANCELLED, and reviving that row is both
	// correct and necessary, since (event_id, member_id) is unique.
	const insert = `INSERT INTO event_registrations (event_id, member_id, status)
		VALUES ($1, $2, 'REGISTERED')
		ON CONFLICT (event_id, member_id) DO UPDATE SET status = 'REGISTERED'
		RETURNING id, event_id, member_id, status, created_at`

	var reg domain.EventRegistration
	if err := tx.QueryRow(ctx, insert, eventID, memberID).
		Scan(&reg.ID, &reg.EventID, &reg.MemberID, &reg.Status, &reg.CreatedAt); err != nil {
		return nil, fmt.Errorf("promote registration: %w", err)
	}

	if _, err := tx.Exec(ctx,
		`UPDATE waitlist_entries SET status = 'PROMOTED', promoted_at = now()
		WHERE id = $1`, entryID); err != nil {
		return nil, fmt.Errorf("mark waitlist entry: %w", err)
	}

	// The promoted member is now genuinely going, so they earn the same
	// engagement record a direct registration writes. Recording it here rather
	// than only in RegisterForEvent is what keeps the activity feed from
	// under-reporting people who joined through the waitlist.
	if _, err := tx.Exec(ctx,
		`INSERT INTO engagement_activities (member_id, activity_type, activity_id, event_type)
		VALUES ($1, 'EVENT', $2, 'REGISTERED_EVENT')`, memberID, eventID); err != nil {
		return nil, fmt.Errorf("record promotion engagement: %w", err)
	}

	return &reg, nil
}
