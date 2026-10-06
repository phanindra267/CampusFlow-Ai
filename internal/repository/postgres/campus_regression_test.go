package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/campuscare/api/internal/config"
	"github.com/campuscare/api/internal/database"
	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/internal/delivery/http/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// This file holds the regression tests for the Phase 0 correctness fixes. Unlike
// the rest of the suite these need a real PostgreSQL, because the bugs they guard
// are properties of the database's isolation level rather than of Go code: an
// overbooking race only exists if two transactions really do interleave, and a
// counter only drifts if the database really does commit two writes.
//
// They therefore skip unless TEST_DATABASE_URL (or the usual DB_* variables) is
// set, which keeps `go test ./...` working on a machine with no database while
// still being runnable in CI:
//
//	TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/campuscare_test' go test ./...
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		cfg, err := config.Load()
		if err != nil {
			t.Skipf("no TEST_DATABASE_URL and config unavailable: %v", err)
		}
		// Point at a scratch database by default: these tests write rows.
		if cfg.Database.Name == "" {
			t.Skip("TEST_DATABASE_URL is not set and DB_NAME is unset")
		}
		dsn = cfg.Database.DSN()
	}

	pool, err := database.NewPostgresPool(config.DatabaseConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     envOr("DB_PORT", "5432"),
		User:     envOr("DB_USER", "postgres"),
		Password: envOr("DB_PASSWORD", "postgres"),
		Name:     envOr("DB_NAME", "campuscare"),
		SSLMode:  envOr("DB_SSLMODE", "disable"),
	})
	if err != nil {
		if dsn != "" {
			t.Skipf("PostgreSQL is not reachable, skipping database regression tests: %v", err)
		}
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// eventFixture is a self-contained event plus the users it needs, torn down
// afterwards. Each test builds its own so the tests stay independent of seeded
// data and of each other.
type eventFixture struct {
	eventID     string
	organiserID string
	members     []string
}

func newEventFixture(t *testing.T, pool *pgxpool.Pool, capacity int) *eventFixture {
	t.Helper()
	ctx := context.Background()

	f := &eventFixture{
		eventID:     newUUID(t),
		organiserID: insertUser(t, pool),
	}
	// Two spare members for the waitlist and the "not registered" case.
	for i := 0; i < 6; i++ {
		f.members = append(f.members, insertUser(t, pool))
	}

	_, err := pool.Exec(ctx,
		`INSERT INTO events (id, title, description, category, organizer_id, venue,
			is_online, start_time, end_time, capacity, status)
		VALUES ($1, 'Regression event', 'for tests', 'TEST', $2, 'Hall', false,
			now() + interval '1 day', now() + interval '1 day 2 hours', $3, 'PUBLISHED')`,
		f.eventID, f.organiserID, capacity)
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}

	t.Cleanup(func() {
		// Cascades clear registrations, waitlist entries, engagement and
		// attendance; the users are removed explicitly because they are shared.
		_, _ = pool.Exec(context.Background(), `DELETE FROM events WHERE id = $1`, f.eventID)
		ids := append([]string{f.organiserID}, f.members...)
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = ANY($1)`, ids)
	})

	return f
}

func newUUID(t *testing.T) string {
	t.Helper()
	// gen_random_uuid() is what the schema uses, so ask the database rather than
	// carrying a UUID dependency into the test.
	return queryScalar(t, nil, `SELECT gen_random_uuid()::text`)
}

func insertUser(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	id := newUUID(t)
	_, err := pool.Exec(context.Background(),
		`INSERT INTO users (id, email, password_hash, display_name, role)
		VALUES ($1, $2, 'x', 'Test User', 'MEMBER')`, id, id+"@example.test")
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func queryScalar(t *testing.T, pool *pgxpool.Pool, q string) string {
	t.Helper()
	if pool == nil {
		t.Fatal("queryScalar called with no pool")
	}
	var out string
	if err := pool.QueryRow(context.Background(), q).Scan(&out); err != nil {
		t.Fatalf("query %q: %v", q, err)
	}
	return out
}

func registeredCount(t *testing.T, pool *pgxpool.Pool, eventID string) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(),
		`SELECT count(*) FROM event_registrations
		WHERE event_id = $1 AND status = 'REGISTERED'`, eventID).Scan(&n); err != nil {
		t.Fatalf("count registrations: %v", err)
	}
	return n
}

func engagementCount(t *testing.T, pool *pgxpool.Pool, memberID, eventID, eventType string) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(),
		`SELECT count(*) FROM engagement_activities
		WHERE member_id = $1 AND activity_type = 'EVENT' AND event_type = $2
			AND activity_id = $3`, memberID, eventType, eventID).Scan(&n); err != nil {
		t.Fatalf("count engagement: %v", err)
	}
	return n
}

// TestRegisterForEventDoesNotOverbook is the Phase 0.2 regression. Capacity 1
// with three simultaneous requests: exactly one must win and the other two must be
// told the event is full. Without the SELECT ... FOR UPDATE that serialises on the
// event row, all three read "0 of 1 taken" and all three insert.
func TestRegisterForEventDoesNotOverbook(t *testing.T) {
	pool := testPool(t)
	f := newEventFixture(t, pool, 1)
	repo := NewCampusRepository(pool)

	var wg sync.WaitGroup
	results := make([]error, len(f.members))
	start := make(chan struct{})

	for i, member := range f.members {
		wg.Add(1)
		go func(idx int, id string) {
			defer wg.Done()
			<-start // release all three at once to maximise the chance of interleaving
			_, _, err := repo.RegisterForEvent(context.Background(), f.eventID, id)
			results[idx] = err
		}(i, member)
	}
	close(start)
	wg.Wait()

	var succeeded, full int
	for _, err := range results {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, ErrEventFull):
			full++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if got := registeredCount(t, pool, f.eventID); got != 1 {
		t.Errorf("capacity is 1 but %d members hold a place: the event was overbooked", got)
	}
	if succeeded != 1 {
		t.Errorf("expected exactly 1 success, got %d (full: %d)", succeeded, full)
	}
}

// TestRegisterForEventRefreshesCountAfterWithdrawal is the accounting half of
// 0.2: the capacity check must count what is actually committed, so a member who
// withdraws frees a place that a later registration can take.
func TestRegisterForEventRefreshesCountAfterWithdrawal(t *testing.T) {
	pool := testPool(t)
	f := newEventFixture(t, pool, 1)
	repo := NewCampusRepository(pool)
	ctx := context.Background()

	if _, _, err := repo.RegisterForEvent(ctx, f.eventID, f.members[0]); err != nil {
		t.Fatalf("first registration: %v", err)
	}
	if _, _, err := repo.RegisterForEvent(ctx, f.eventID, f.members[1]); !errors.Is(err, ErrEventFull) {
		t.Fatalf("expected the second registration to be refused, got %v", err)
	}

	if _, err := repo.CancelRegistration(ctx, f.eventID, f.members[0]); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if _, _, err := repo.RegisterForEvent(ctx, f.eventID, f.members[1]); err != nil {
		t.Fatalf("the freed place should have been available, got %v", err)
	}
	if got := registeredCount(t, pool, f.eventID); got != 1 {
		t.Errorf("expected 1 registration after the swap, got %d", got)
	}
}

// TestCancelRegistrationPromotesWaitlist is the Phase 0.10 auto-promote
// regression. Freeing a place must hand it to whoever waited longest, in the same
// transaction, rather than leaving the seat empty until an organiser notices.
func TestCancelRegistrationPromotesWaitlist(t *testing.T) {
	pool := testPool(t)
	f := newEventFixture(t, pool, 1)
	repo := NewCampusRepository(pool)
	ctx := context.Background()

	holder, waiting := f.members[0], f.members[1]

	if _, _, err := repo.RegisterForEvent(ctx, f.eventID, holder); err != nil {
		t.Fatalf("register holder: %v", err)
	}
	// The event is full, so this member lands on the waitlist.
	if _, err := repo.JoinWaitlist(ctx, &domain.WaitlistEntry{
		EventID: f.eventID, MemberID: waiting,
	}); err != nil {
		t.Fatalf("join waitlist: %v", err)
	}

	promoted, err := repo.CancelRegistration(ctx, f.eventID, holder)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if promoted == nil {
		t.Fatal("expected the waiting member to be promoted, got nobody")
	}
	if promoted.MemberID != waiting {
		t.Errorf("promoted the wrong member: got %s, want %s", promoted.MemberID, waiting)
	}
	if got := registeredCount(t, pool, f.eventID); got != 1 {
		t.Errorf("expected exactly the promoted member to hold the place, got %d", got)
	}

	var status string
	if err := pool.QueryRow(ctx,
		`SELECT status FROM waitlist_entries WHERE event_id = $1 AND member_id = $2`,
		f.eventID, waiting).Scan(&status); err != nil {
		t.Fatalf("read waitlist entry: %v", err)
	}
	if status != "PROMOTED" {
		t.Errorf("waitlist entry status is %q, want PROMOTED", status)
	}
}

// TestCancelRegistrationPromotesInWaitlistOrder checks the fairness property:
// with two people waiting, the one who waited longest gets the seat.
func TestCancelRegistrationPromotesInWaitlistOrder(t *testing.T) {
	pool := testPool(t)
	f := newEventFixture(t, pool, 1)
	repo := NewCampusRepository(pool)
	ctx := context.Background()

	holder, first, second := f.members[0], f.members[1], f.members[2]

	if _, _, err := repo.RegisterForEvent(ctx, f.eventID, holder); err != nil {
		t.Fatalf("register holder: %v", err)
	}
	// joined_at defaults to now(), so separate the two entries explicitly to make
	// the ordering the test relies on deterministic rather than timing-dependent.
	for i, member := range []string{first, second} {
		if _, err := repo.JoinWaitlist(ctx, &domain.WaitlistEntry{
			EventID: f.eventID, MemberID: member,
		}); err != nil {
			t.Fatalf("join waitlist (%d): %v", i, err)
		}
		if _, err := pool.Exec(ctx,
			`UPDATE waitlist_entries SET joined_at = now() + make_interval(secs => $3)
			WHERE event_id = $1 AND member_id = $2`,
			f.eventID, member, i); err != nil {
			t.Fatalf("order waitlist entry (%d): %v", i, err)
		}
	}

	promoted, err := repo.CancelRegistration(ctx, f.eventID, holder)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if promoted == nil || promoted.MemberID != first {
		t.Fatalf("promoted %v, want the first member to have waited longest (%s)", promoted, first)
	}
}

// TestCancelRegistrationClearsEngagement is the other half of 0.10. Cancelling
// must retract the activity-feed record written at registration, otherwise the
// member keeps seeing themselves as going to an event they withdrew from.
func TestCancelRegistrationClearsEngagement(t *testing.T) {
	pool := testPool(t)
	f := newEventFixture(t, pool, 5)
	repo := NewCampusRepository(pool)
	ctx := context.Background()

	member := f.members[0]
	if _, _, err := repo.RegisterForEvent(ctx, f.eventID, member); err != nil {
		t.Fatalf("register: %v", err)
	}
	if got := engagementCount(t, pool, member, f.eventID, "REGISTERED_EVENT"); got != 1 {
		t.Fatalf("expected registration to record 1 engagement row, got %d", got)
	}

	if _, err := repo.CancelRegistration(ctx, f.eventID, member); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if got := engagementCount(t, pool, member, f.eventID, "REGISTERED_EVENT"); got != 0 {
		t.Errorf("cancelling left %d engagement rows behind; the feed now claims the "+
			"member is still going", got)
	}
}

// TestCancelRegistrationOnEmptyWaitlistIsNotAnError pins the distinction the two
// callers depend on: a cancellation with nobody waiting succeeds and reports no
// promotion, whereas PromoteWaitlist on the same event reports ErrNotFound.
func TestCancelRegistrationOnEmptyWaitlistIsNotAnError(t *testing.T) {
	pool := testPool(t)
	f := newEventFixture(t, pool, 5)
	repo := NewCampusRepository(pool)
	ctx := context.Background()

	member := f.members[0]
	if _, _, err := repo.RegisterForEvent(ctx, f.eventID, member); err != nil {
		t.Fatalf("register: %v", err)
	}

	promoted, err := repo.CancelRegistration(ctx, f.eventID, member)
	if err != nil {
		t.Fatalf("cancelling with an empty waitlist should succeed, got %v", err)
	}
	if promoted != nil {
		t.Errorf("nothing was waiting, so nobody should be promoted, got %+v", promoted)
	}

	// Cancelling again has nothing to cancel.
	if _, err := repo.CancelRegistration(ctx, f.eventID, member); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on a second cancel, got %v", err)
	}
}

// TestPromoteWaitlistRespectsCapacity is the Phase 0.3 regression: an organiser
// pressing promote on an already-full event must be refused rather than
// overbooking it.
func TestPromoteWaitlistRespectsCapacity(t *testing.T) {
	pool := testPool(t)
	f := newEventFixture(t, pool, 1)
	campus := NewCampusRepository(pool)
	events := NewEventRepository(pool)
	ctx := context.Background()

	holder, waiting := f.members[0], f.members[1]

	if _, _, err := campus.RegisterForEvent(ctx, f.eventID, holder); err != nil {
		t.Fatalf("register holder: %v", err)
	}
	if _, err := campus.JoinWaitlist(ctx, &domain.WaitlistEntry{
		EventID: f.eventID, MemberID: waiting,
	}); err != nil {
		t.Fatalf("join waitlist: %v", err)
	}

	// The event is full, so promoting must be refused even though somebody waits.
	if _, err := events.PromoteWaitlist(ctx, f.eventID); !errors.Is(err, ErrEventFull) {
		t.Fatalf("expected ErrEventFull for a full event, got %v", err)
	}
	if got := registeredCount(t, pool, f.eventID); got != 1 {
		t.Errorf("a refused promotion changed the headcount to %d", got)
	}

	// Freeing the seat lets the promotion through.
	if _, err := campus.CancelRegistration(ctx, f.eventID, holder); err != nil {
		t.Fatalf("cancel: %v", err)
	}
}

// TestPromoteWaitlistOnEmptyWaitlistIsNotFound contrasts with the cancellation
// case above: the organiser's explicit request must fail loudly.
func TestPromoteWaitlistOnEmptyWaitlistIsNotFound(t *testing.T) {
	pool := testPool(t)
	f := newEventFixture(t, pool, 5)
	events := NewEventRepository(pool)

	if _, err := events.PromoteWaitlist(context.Background(), f.eventID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for an empty waitlist, got %v", err)
	}
}

// TestCheckInRequiresRegistration is the Phase 0.4 regression. The QR poster at
// the door is a location, not a capability: photographing it must not let an
// unregistered member mark themselves present.
func TestCheckInRequiresRegistration(t *testing.T) {
	pool := testPool(t)
	f := newEventFixture(t, pool, 10)
	repo := NewCampusRepository(pool)
	ctx := context.Background()

	sessionID := insertSession(t, pool, f.eventID)
	registered, stranger := f.members[0], f.members[1]

	if _, _, err := repo.RegisterForEvent(ctx, f.eventID, registered); err != nil {
		t.Fatalf("register: %v", err)
	}

	session := &domain.CheckInSession{
		EventSessionID: sessionID,
		OrganizerID:    f.organiserID,
		ExpiresAt:      time.Now().Add(15 * time.Minute),
	}
	if err := repo.OpenCheckIn(ctx, session); err != nil {
		t.Fatalf("open check-in: %v", err)
	}

	_, err := repo.CheckInByToken(ctx, session.QRToken, stranger, "QR")
	if !errors.Is(err, ErrNotRegistered) {
		t.Fatalf("an unregistered member was able to check in: got %v, want ErrNotRegistered", err)
	}

	// The same token works for somebody who is actually registered.
	if _, err := repo.CheckInByToken(ctx, session.QRToken, registered, "QR"); err != nil {
		t.Fatalf("a registered member was refused: %v", err)
	}
}

// TestCheckInIsIdempotent covers the replay case: a member scanning twice is
// recorded once.
func TestCheckInIsIdempotent(t *testing.T) {
	pool := testPool(t)
	f := newEventFixture(t, pool, 10)
	repo := NewCampusRepository(pool)
	ctx := context.Background()

	sessionID := insertSession(t, pool, f.eventID)
	member := f.members[0]
	if _, _, err := repo.RegisterForEvent(ctx, f.eventID, member); err != nil {
		t.Fatalf("register: %v", err)
	}

	session := &domain.CheckInSession{
		EventSessionID: sessionID,
		OrganizerID:    f.organiserID,
		ExpiresAt:      time.Now().Add(15 * time.Minute),
	}
	if err := repo.OpenCheckIn(ctx, session); err != nil {
		t.Fatalf("open check-in: %v", err)
	}

	if _, err := repo.CheckInByToken(ctx, session.QRToken, member, "QR"); err != nil {
		t.Fatalf("first check-in: %v", err)
	}
	if _, err := repo.CheckInByToken(ctx, session.QRToken, member, "QR"); !errors.Is(err, ErrCheckInRejected) {
		t.Errorf("expected the second scan to be rejected, got %v", err)
	}

	var n int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM attendance_records WHERE event_session_id = $1 AND member_id = $2`,
		sessionID, member).Scan(&n); err != nil {
		t.Fatalf("count attendance: %v", err)
	}
	if n != 1 {
		t.Errorf("expected exactly 1 attendance row after two scans, got %d", n)
	}
}

// TestSessionOwnedByIsPerEvent is the Phase 0.5 regression: holding the
// ORGANIZER role is not permission to open a check-in window on somebody else's
// event.
func TestSessionOwnedByIsPerEvent(t *testing.T) {
	pool := testPool(t)
	f := newEventFixture(t, pool, 10)
	repo := NewCampusRepository(pool)

	sessionID := insertSession(t, pool, f.eventID)
	otherOrganiser := insertUser(t, pool)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, otherOrganiser)
	})

	owned, err := repo.SessionOwnedBy(context.Background(), sessionID, f.organiserID)
	if err != nil {
		t.Fatalf("SessionOwnedBy: %v", err)
	}
	if !owned {
		t.Error("the organiser who owns the event was told they do not own the session")
	}

	owned, err = repo.SessionOwnedBy(context.Background(), sessionID, otherOrganiser)
	if err != nil {
		t.Fatalf("SessionOwnedBy: %v", err)
	}
	if owned {
		t.Error("a different organiser was told they own somebody else's session")
	}
}

// TestApplicationCountIgnoresWithdrawn is the Phase 0.10 counter regression. The
// denormalised count must agree with the applications table under the same rule
// everywhere: withdrawn applications do not count.
func TestApplicationCountIgnoresWithdrawn(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	opportunityID := newUUID(t)
	_, err := pool.Exec(ctx,
		`INSERT INTO opportunities (id, title, description, category, posted_by,
			application_count, status)
		VALUES ($1, 'Regression role', 'for tests', 'TEST', $2, 0, 'OPEN')`,
		opportunityID, insertUser(t, pool))
	if err != nil {
		t.Skipf("opportunities table does not accept this insert, skipping: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM opportunities WHERE id = $1`, opportunityID)
	})

	repo := NewOpportunityRepository(pool)

	resume := ""
	if _, err := repo.ApplyToOpportunity(ctx, opportunityID, f0User(t, pool), "", &resume); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if got := storedApplicationCount(t, pool, opportunityID); got != 1 {
		t.Fatalf("expected application_count 1 after one application, got %d", got)
	}

	if err := repo.WithdrawApplication(ctx, f0User(t, pool), opportunityID); err != nil {
		t.Fatalf("withdraw: %v", err)
	}
	if got := storedApplicationCount(t, pool, opportunityID); got != 0 {
		t.Errorf("application_count is %d after withdrawing the only application; the "+
			"counter and the table disagree", got)
	}
}

// f0User returns a fresh user, named for its only job of being an applicant.
func f0User(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	id := insertUser(t, pool)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	})
	return id
}

func storedApplicationCount(t *testing.T, pool *pgxpool.Pool, opportunityID string) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(),
		`SELECT application_count FROM opportunities WHERE id = $1`, opportunityID).Scan(&n); err != nil {
		t.Fatalf("read application_count: %v", err)
	}
	return n
}

func insertSession(t *testing.T, pool *pgxpool.Pool, eventID string) string {
	t.Helper()
	id := newUUID(t)
	_, err := pool.Exec(context.Background(),
		`INSERT INTO event_sessions (id, event_id, title, start_time, end_time)
		VALUES ($1, $2, 'Regression session', now(), now() + interval '2 hours')`, id, eventID)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}
	return id
}

// TestApplicationCountUnderConcurrency guards the write-while-counting pattern:
// several members applying at once must still leave application_count equal to
// the number of live applications.
func TestApplicationCountUnderConcurrency(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	opportunityID := newUUID(t)
	_, err := pool.Exec(ctx,
		`INSERT INTO opportunities (id, title, description, category, posted_by,
			application_count, status)
		VALUES ($1, 'Concurrent role', 'for tests', 'TEST', $2, 0, 'OPEN')`,
		opportunityID, insertUser(t, pool))
	if err != nil {
		t.Skipf("opportunities table does not accept this insert, skipping: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM opportunities WHERE id = $1`, opportunityID)
	})

	applicants := make([]string, 8)
	for i := range applicants {
		applicants[i] = f0User(t, pool)
	}

	repo := NewOpportunityRepository(pool)

	var wg sync.WaitGroup
	start := make(chan struct{})
	for _, applicant := range applicants {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			<-start
			resume := ""
			if _, err := repo.ApplyToOpportunity(ctx, opportunityID, id, "", &resume); err != nil {
				t.Errorf("apply: %v", err)
			}
		}(applicant)
	}
	close(start)
	wg.Wait()

	var truth int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM opportunity_applications
		WHERE opportunity_id = $1 AND status <> 'WITHDRAWN'`, opportunityID).Scan(&truth); err != nil {
		t.Fatalf("count applications: %v", err)
	}
	if got := storedApplicationCount(t, pool, opportunityID); got != truth {
		t.Errorf("application_count is %d but %d applications are live: the counter "+
			"drifted under concurrency", got, truth)
	}
}

// TestRateLimiterIsRaceFree exercises the limiter concurrently. It belongs in the
// database-integration build only because that is where `go test -race` runs in
// CI; it needs no database and would pass either way.
func TestRateLimiterConcurrentAllow(t *testing.T) {
	t.Parallel()

	rl := newBurstLimiterForTest()

	var wg sync.WaitGroup
	var allowed int64
	var mu sync.Mutex

	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ok, _ := rl.Allow("shared-key"); ok {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if allowed > 1 {
		t.Errorf("%d concurrent requests were allowed through a single-token bucket", allowed)
	}
}

func newBurstLimiterForTest() interface{ Allow(string) (bool, time.Duration) } {
	// Use the middleware rate limiter for testing
	return middleware.NewRateLimiter(middleware.RateLimitConfig{PerMinute: 60, Burst: 1})
}

var _ = fmt.Sprintf