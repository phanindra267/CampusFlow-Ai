// Command seed loads a small, honest demo campus so the application is
// navigable the moment it starts.
//
// Everything here is demonstration data for one fictional campus. It deliberately
// invents no official facts: no real institution's clubs, staff or portals are
// named, and every person in the directory is labelled as a placeholder. Replace
// it with real content before showing the platform to anyone.
//
// The seed is idempotent: every row is keyed on a natural column and skipped when
// it already exists, so running it twice changes nothing.
//
// Usage:
//
//	seed              seed an empty database
//	seed -force       seed even if content already exists
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/campuscare/api/internal/config"
	"github.com/campuscare/api/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var force = flag.Bool("force", false, "seed even if content already exists")

func main() {
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("seed: load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := database.NewPostgresPool(cfg.Database)
	if err != nil {
		log.Fatalf("seed: connect: %v", err)
	}
	defer pool.Close()

	s := &seeder{pool: pool}

	if err := s.run(ctx); err != nil {
		log.Fatalf("seed: %v", err)
	}
}

type seeder struct {
	pool *pgxpool.Pool
}

func (s *seeder) run(ctx context.Context) error {
	empty, err := s.isEmpty(ctx)
	if err != nil {
		return err
	}
	if !empty && !*force {
		log.Println("seed: content already present, nothing to do (use -force to continue)")
		return nil
	}

	adminID, err := s.seedAccounts(ctx)
	if err != nil {
		return fmt.Errorf("accounts: %w", err)
	}

	services, err := s.seedServices(ctx)
	if err != nil {
		return fmt.Errorf("campus services: %w", err)
	}
	if err := s.seedServiceFAQs(ctx, services); err != nil {
		return fmt.Errorf("service faqs: %w", err)
	}

	clubs, err := s.seedClubs(ctx, adminID)
	if err != nil {
		return fmt.Errorf("clubs: %w", err)
	}
	if err := s.seedClubActivity(ctx, adminID, clubs); err != nil {
		return fmt.Errorf("club activity: %w", err)
	}
	if err := s.seedEvents(ctx, adminID, clubs); err != nil {
		return fmt.Errorf("events: %w", err)
	}
	if err := s.seedOpportunities(ctx, adminID, clubs); err != nil {
		return fmt.Errorf("opportunities: %w", err)
	}
	if err := s.seedPeople(ctx); err != nil {
		return fmt.Errorf("people directory: %w", err)
	}
	if err := s.seedAnnouncements(ctx, adminID); err != nil {
		return fmt.Errorf("announcements: %w", err)
	}

	log.Println("seed: done. sign in with one of:")
	log.Println("  admin@campuscare.test / ChangeMe123! (ADMIN)")
	log.Println("  organiser@campuscare.test / ChangeMe123! (ORGANIZER)")
	log.Println("  student@campuscare.test / ChangeMe123! (MEMBER)")
	return nil
}

// isEmpty reports whether the campus content tables are untouched, which is how
// a fresh database is told from one that already has data.
func (s *seeder) isEmpty(ctx context.Context) (bool, error) {
	const query = `SELECT
		(SELECT count(*) FROM campus_resources) +
		(SELECT count(*) FROM clubs) +
		(SELECT count(*) FROM events) +
		(SELECT count(*) FROM opportunities)`

	var total int
	if err := s.pool.QueryRow(ctx, query).Scan(&total); err != nil {
		return false, fmt.Errorf("check existing content: %w", err)
	}
	return total == 0, nil
}

// ---------------------------------------------------------------- accounts

// demoPassword is shared by the demo accounts. It is deliberately obvious so
// nobody runs a deployment with it.
const demoPassword = "ChangeMe123!"

func (s *seeder) seedAccounts(ctx context.Context) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(demoPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash demo password: %w", err)
	}

	accounts := []struct {
		email string
		name  string
		role  string
	}{
		{"admin@campuscare.test", "Campus Administrator", "ADMIN"},
		{"organiser@campuscare.test", "Events Organiser", "ORGANIZER"},
		{"student@campuscare.test", "Student Member", "MEMBER"},
	}

	var adminID string
	for _, account := range accounts {
		const query = `INSERT INTO users (email, display_name, password_hash, role, status)
			VALUES ($1, $2, $3, $4, 'ACTIVE')
			ON CONFLICT (email) DO NOTHING
			RETURNING id`

		var id string
		err := s.pool.QueryRow(ctx, query, account.email, account.name,
			string(hashed), account.role).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			const lookup = `SELECT id FROM users WHERE email = $1`
			if err := s.pool.QueryRow(ctx, lookup, account.email).Scan(&id); err != nil {
				return "", fmt.Errorf("%s: %w", account.email, err)
			}
		} else if err != nil {
			return "", fmt.Errorf("%s: %w", account.email, err)
		}

		if account.role == "ADMIN" {
			adminID = id
		}
	}
	return adminID, nil
}

// ---------------------------------------------------------- campus services

// serviceSeed is one entry in the campus service catalogue.
type serviceSeed struct {
	name        string
	kind        string
	location    string
	bookable    bool
	description string
	faq         [][2]string
}

var serviceSeeds = []serviceSeed{
	{
		name: "Hostel Support Desk", kind: "SERVICE", location: "Hostel Block A",
		description: "Room allocation, maintenance requests and hostel fee queries.",
		faq: [][2]string{
			{"How do I report a maintenance problem in my room?",
				"Open a request against this service with the room number and a short description. You will get a reference code, and you can follow the status from your profile."},
			{"Who handles a complaint about a roommate or hostel staff?",
				"File a complaint against this service. Complaints go to the hostel warden directly rather than through the general queue."},
		},
	},
	{
		name: "Campus Transport", kind: "TRANSPORT", location: "Transport Office", bookable: true,
		description: "Bus routes, passes and late-evening pickup points.",
		faq: [][2]string{
			{"How do I get a bus pass?",
				"Apply through this service and collect the pass from the transport office. Renewals keep the same pass number."},
		},
	},
	{
		name: "IT Helpdesk", kind: "SERVICE", location: "Central Library, Level 1",
		description: "Campus network accounts, Wi-Fi, lab software and device issues.",
		faq: [][2]string{
			{"My campus account stopped working. What should I do?",
				"Open a request here with your registration identifier. Password resets are handled by the helpdesk, not by email."},
			{"Can I get software installed on a lab machine?",
				"Yes. Name the software and the lab, and the helpdesk will confirm whether a licence allows it."},
		},
	},
	{
		name: "Central Library", kind: "LIBRARY", location: "Central Library", bookable: true,
		description: "Issue desk, study spaces and reading room booking.",
		faq: [][2]string{
			{"How do I reserve a discussion room?",
				"Book the library as a resource with the time window you need. Overlapping bookings are refused rather than silently double-booked."},
		},
	},
	{
		name: "Sports Complex", kind: "SPORTS", location: "Sports Complex", bookable: true,
		description: "Turf, courts, gymnasium and equipment issue.",
		faq: [][2]string{
			{"Do I need to pay to use the gym?",
				"No. Registration at the sports office is enough; bring your ID for the first visit."},
		},
	},
	{
		name: "Health Centre", kind: "HEALTH", location: "Health Centre",
		description: "First aid, campus clinic appointments and counselling.",
		faq: [][2]string{
			{"How do I book a counselling appointment?",
				"Open a request here. Only the clinic and the requester can see the details, so describe only what you are comfortable sharing."},
		},
	},
	{
		name: "Fee Desk", kind: "SERVICE", location: "Administration Block",
		description: "Fee payment, receipts and instalment questions.",
		faq: [][2]string{
			{"I paid but my receipt is missing. Who do I contact?",
				"Open a request with your transaction reference. The fee desk can reissue a receipt from the payment record."},
		},
	},
}

func (s *seeder) seedServices(ctx context.Context) (map[string]string, error) {
	ids := make(map[string]string, len(serviceSeeds))

	for _, service := range serviceSeeds {
		const query = `INSERT INTO campus_resources (name, resource_type, description,
				location, is_bookable, requires_auth, status)
			SELECT $1::varchar, $2::varchar, $3::text, $4::varchar, $5::boolean, true, 'ACTIVE'
			WHERE NOT EXISTS (SELECT 1 FROM campus_resources WHERE name = $1::varchar)
			RETURNING id`

		var id string
		err := s.pool.QueryRow(ctx, query, service.name, service.kind,
			service.description, service.location, service.bookable).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			const lookup = `SELECT id FROM campus_resources WHERE name = $1`
			if err := s.pool.QueryRow(ctx, lookup, service.name).Scan(&id); err != nil {
				return nil, fmt.Errorf("%s: %w", service.name, err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("%s: %w", service.name, err)
		}
		ids[service.name] = id
	}
	return ids, nil
}

func (s *seeder) seedServiceFAQs(ctx context.Context, services map[string]string) error {
	for _, service := range serviceSeeds {
		for position, entry := range service.faq {
			const query = `INSERT INTO service_faqs (resource_id, question, answer, position)
				SELECT $1::uuid, $2::text, $3::text, $4::int
				WHERE NOT EXISTS (
					SELECT 1 FROM service_faqs
					WHERE resource_id = $1::uuid AND question = $2::text
				)`

			if _, err := s.pool.Exec(ctx, query, services[service.name],
				entry[0], entry[1], position); err != nil {
				return fmt.Errorf("%s faq %d: %w", service.name, position, err)
			}
		}
	}
	return nil
}

// ------------------------------------------------------------------- clubs

// clubSeeds names demo clubs. They are generic on purpose: the seed does not
// claim to reproduce any real institution's club list.
var clubSeeds = []struct {
	name     string
	category string
	blurb    string
	contact  string
}{
	{"Coding Club", "TECHNICAL",
		"Weekend build sessions, internal hackathons and code review help for anyone starting out.",
		"coding.club@campuscare.test"},
	{"Robotics Club", "TECHNICAL",
		"Student-built robots, an annual line-following contest and a shared workshop.",
		"robotics.club@campuscare.test"},
	{"Photography Club", "CREATIVE",
		"Photo walks, a darkroom basics course and an annual campus exhibition.",
		"photography.club@campuscare.test"},
	{"Cultural Committee", "CULTURAL",
		"Dance, music and theatre groups that rehearse on campus and perform each term.",
		"cultural.committee@campuscare.test"},
	{"Sports Council", "SPORTS",
		"Intramurals, inter-club fixtures and coaching for the university teams.",
		"sports.council@campuscare.test"},
	{"Entrepreneurship Cell", "TECHNICAL",
		"Startup mentoring, pitch practice and connections to alumni founders.",
		"ecell@campuscare.test"},
}

func (s *seeder) seedClubs(ctx context.Context, createdBy string) (map[string]string, error) {
	ids := make(map[string]string, len(clubSeeds))

	for _, club := range clubSeeds {
		slug := slugify(club.name)

		const query = `INSERT INTO clubs (name, slug, description, category, created_by,
				verification_status, status, contact_email)
			VALUES ($1, $2, $3, $4, $5, 'VERIFIED', 'ACTIVE', $6)
			ON CONFLICT (slug) DO NOTHING
			RETURNING id`

		var id string
		err := s.pool.QueryRow(ctx, query, club.name, slug, club.blurb, club.category,
			createdBy, club.contact).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			const lookup = `SELECT id FROM clubs WHERE slug = $1`
			if err := s.pool.QueryRow(ctx, lookup, slug).Scan(&id); err != nil {
				return nil, fmt.Errorf("%s: %w", club.name, err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("%s: %w", club.name, err)
		}
		ids[club.name] = id
	}
	return ids, nil
}

func (s *seeder) seedClubActivity(ctx context.Context, createdBy string, clubs map[string]string) error {
	activities := []struct {
		club     string
		title    string
		blurb    string
		kind     string
		location string
	}{
		{"Coding Club", "Intro to version control workshop", "A hands-on session on branches, merges and recovering a mistake.", "WORKSHOP", "Computer Lab 2"},
		{"Coding Club", "Internal hackathon", "A one-day build sprint open to every club member.", "HACKATHON", "Innovation Centre"},
		{"Robotics Club", "Line-following contest", "Build, tune and race over a weekend.", "COMPETITION", "Sports Complex"},
		{"Photography Club", "Golden hour photo walk", "An evening walk around campus with a mentor.", "OUTING", "Central Library"},
		{"Cultural Committee", "Rehearsal week", "Group rehearsals ahead of the term performance.", "REHEARSAL", "Auditorium"},
		{"Sports Council", "Intramural registrations open", "Team registration for the inter-club fixtures.", "NOTICE", "Sports Complex"},
		{"Entrepreneurship Cell", "Pitch clinic", "Mock pitches with feedback from alumni founders.", "WORKSHOP", "Seminar Hall"},
	}

	for _, activity := range activities {
		clubID := clubs[activity.club]
		if clubID == "" {
			continue
		}

		const query = `INSERT INTO club_activities (club_id, title, description,
				activity_type, location, starts_at, ends_at, created_by)
			SELECT $1::uuid, $2::varchar, $3::text, $4::varchar, $5::varchar,
				now(), now() + interval '2 hours', $6::uuid
			WHERE NOT EXISTS (
				SELECT 1 FROM club_activities WHERE club_id = $1::uuid AND title = $2::varchar
			)`

		if _, err := s.pool.Exec(ctx, query, clubID, activity.title, activity.blurb,
			activity.kind, activity.location, createdBy); err != nil {
			return fmt.Errorf("%s: %w", activity.title, err)
		}
	}
	return nil
}

// ------------------------------------------------------------------- events

func (s *seeder) seedEvents(ctx context.Context, organizerID string, clubs map[string]string) error {
	now := time.Now()
	inDays := func(days int, hour, minute int) time.Time {
		day := now.AddDate(0, 0, days)
		return time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, now.Location())
	}

	events := []struct {
		title     string
		blurb     string
		category  string
		venue     string
		startsAt  time.Time
		endsAt    time.Time
		deadline  time.Time
		capacity  int
		eligible  string
		club      string
		published bool
	}{
		{
			"Campus orientation walkthrough",
			"A guided walk through the campus for new members: where services live, how to book a space, and which clubs are recruiting.",
			"GENERAL", "Meet at the Main Gate", inDays(6, 10, 0), inDays(6, 12, 0), inDays(4, 23, 59),
			120, "Open to all members", "", true,
		},
		{
			"Hackathon: build something small",
			"A one-day build sprint. Teams of up to four, mentors on site, and a demo session at the end.",
			"HACKATHON", "Innovation Centre", inDays(12, 9, 0), inDays(12, 18, 0), inDays(9, 23, 59),
			80, "Open to all members; teams may be mixed across clubs", "Coding Club", true,
		},
		{
			"Robotics showcase",
			"Student robots on the floor, with a short build talk from each team.",
			"EXHIBITION", "Sports Complex", inDays(18, 11, 0), inDays(18, 15, 0), inDays(16, 23, 59),
			150, "Open to visitors and members", "Robotics Club", true,
		},
		{
			"Term performance night",
			"Dance, music and theatre groups perform what they have been rehearsing.",
			"CULTURAL", "Auditorium", inDays(25, 18, 0), inDays(25, 21, 0), inDays(23, 23, 59),
			300, "Open to all members and guests", "Cultural Committee", true,
		},
		{
			"How to write a good project report",
			"A short workshop on writing a project report people can actually read.",
			"WORKSHOP", "Seminar Hall", inDays(9, 14, 0), inDays(9, 15, 30), inDays(7, 23, 59),
			40, "Second year and above", "", true,
		},
		{
			"Service desk orientation (staff)",
			"An internal walkthrough of the service desk queue and how resolutions are recorded.",
			"GENERAL", "Admin Block", inDays(3, 16, 0), inDays(3, 17, 0), inDays(2, 23, 59),
			20, "Organisers and staff only", "", false,
		},
	}

	for _, event := range events {
		var clubID any
		if event.club != "" {
			clubID = clubs[event.club]
		}

		status := "DRAFT"
		if event.published {
			status = "PUBLISHED"
		}

		const query = `INSERT INTO events (title, description, category, organizer_id,
				venue, is_online, start_time, end_time, registration_deadline,
				capacity, status, eligibility, club_id)
			SELECT $1::varchar, $2::text, $3::varchar, $4::uuid, $5::varchar, false,
				$6::timestamptz, $7::timestamptz, $8::timestamptz, $9::int,
				$10::varchar, $11::text, $12::uuid
			WHERE NOT EXISTS (
				SELECT 1 FROM events WHERE title = $1::varchar
			)`

		if _, err := s.pool.Exec(ctx, query, event.title, event.blurb, event.category,
			organizerID, event.venue, event.startsAt, event.endsAt, event.deadline,
			event.capacity, status, event.eligible, clubID); err != nil {
			return fmt.Errorf("%s: %w", event.title, err)
		}
	}
	return nil
}

// ------------------------------------------------------------ opportunities

func (s *seeder) seedOpportunities(ctx context.Context, organizerID string, clubs map[string]string) error {
	now := time.Now()
	inDays := func(days int) time.Time {
		return now.AddDate(0, 0, days)
	}

	opportunities := []struct {
		title    string
		blurb    string
		kind     string
		category string
		deadline time.Time
		starts   time.Time
		location string
		capacity int
		eligible string
		source   string
		club     string
	}{
		{
			"Open-source internship (remote)",
			"A paid remote internship working on an open-source project. Open to students who have contributed before.",
			"INTERNSHIP", "TECHNICAL", inDays(21), inDays(35), "Remote", 10,
			"Third year and above, or anyone with a public repository history", "EXTERNAL", "",
		},
		{
			"Summer research assistantship",
			"Ten weeks of supervised research on an open problem, with a stipend. Applications close early.",
			"RESEARCH", "RESEARCH", inDays(12), inDays(30), "Research Lab", 6,
			"Third year and above; an academic reference is helpful but not required", "INTERNAL", "",
		},
		{
			"Merit scholarship applications",
			"Merit scholarships for continuing students. The form is short; the deadline is not.",
			"SCHOLARSHIP", "GENERAL", inDays(9), inDays(20), "Administration Block", 0,
			"Open to all continuing students", "INTERNAL", "",
		},
		{
			"Campus volunteering drive",
			"A weekend of campus clean-up and planting. Open to members and their families.",
			"VOLUNTEERING", "COMMUNITY", inDays(5), inDays(7), "Campus Grounds", 60,
			"Open to all members and guests", "CLUB", "Cultural Committee",
		},
		{
			"Startup founder office hours",
			"Four founders answer questions about starting a company, from idea to first customer.",
			"STARTUP", "ENTREPRENEURSHIP", inDays(7), inDays(8), "E-Cell Hub", 30,
			"Open to all members", "CLUB", "Entrepreneurship Cell",
		},
		{
			"Photography commission call",
			"Campus events need photographers. This is a paid, term-long commission for club members.",
			"PROJECT", "CREATIVE", inDays(14), inDays(18), "Varies", 5,
			"Photography Club members, or anyone with a portfolio", "CLUB", "Photography Club",
		},
		{
			"Intro to data analysis workshop",
			"A hands-on session on cleaning and plotting real data sets.",
			"WORKSHOP", "TECHNICAL", inDays(4), inDays(5), "Computer Lab 1", 30,
			"Open to all members", "INTERNAL", "",
		},
	}

	for _, opportunity := range opportunities {
		var clubID any
		if opportunity.club != "" {
			clubID = clubs[opportunity.club]
		}

		const query = `INSERT INTO opportunities (title, description, type, category,
				organizer_id, club_id, start_date, end_date, registration_deadline,
				location, delivery_mode, capacity, status, eligibility, source)
			SELECT $1::varchar, $2::text, $3::varchar, $4::varchar, $5::uuid, $6::uuid,
				$7::timestamptz, $8::timestamptz, $9::timestamptz, $10::varchar,
				'IN_PERSON', $11::int, 'OPEN', $12::text, $13::varchar
			WHERE NOT EXISTS (
				SELECT 1 FROM opportunities WHERE title = $1::varchar
			)`

		if _, err := s.pool.Exec(ctx, query, opportunity.title, opportunity.blurb,
			opportunity.kind, opportunity.category, organizerID, clubID,
			opportunity.starts, opportunity.starts.AddDate(0, 0, 1),
			opportunity.deadline, opportunity.location, opportunity.capacity,
			opportunity.eligible, opportunity.source); err != nil {
			return fmt.Errorf("%s: %w", opportunity.title, err)
		}
	}
	return nil
}

// ---------------------------------------------------------- people directory

// People are placeholders on purpose. A demo directory that names real-looking
// staff would put words in the mouths of people who never agreed to it, so each
// entry is explicitly marked as a sample.
func (s *seeder) seedPeople(ctx context.Context) error {
	entries := []struct {
		name        string
		role        string
		school      string
		designation string
		interests   []string
		accepting   bool
	}{
		{"Sample Supervisor - Computer Science", "SUPERVISOR", "Computing",
			"Associate Professor (sample entry)",
			[]string{"Machine Learning", "Computer Vision", "Distributed Systems"}, true},
		{"Sample Supervisor - Life Sciences", "FACULTY", "Life Sciences",
			"Professor (sample entry)",
			[]string{"Genomics", "Public Health", "Ecology"}, true},
		{"Sample Researcher - Materials", "RESEARCHER", "Engineering",
			"Research Scientist (sample entry)",
			[]string{"Materials Science", "Renewable Energy"}, false},
		{"Sample Mentor - Careers", "MENTOR", "Student Development",
			"Career Mentor (sample entry)",
			[]string{"Careers", "Entrepreneurship", "Internships"}, true},
		{"Sample Coordinator - Clubs", "COORDINATOR", "Student Development",
			"Clubs Coordinator (sample entry)",
			[]string{"Event Management", "Volunteering"}, true},
		{"Sample Supervisor - Mathematics", "FACULTY", "Mathematics",
			"Assistant Professor (sample entry)",
			[]string{"Applied Mathematics", "Statistics", "Machine Learning"}, true},
	}

	for _, entry := range entries {
		const query = `INSERT INTO people (display_name, person_role, school, designation,
				bio, research_interests, accepting_students, status)
			SELECT $1::varchar, $2::varchar, $3::varchar, $4::varchar, $5::text,
				$6::varchar[], $7::boolean, 'ACTIVE'
			WHERE NOT EXISTS (
				SELECT 1 FROM people WHERE display_name = $1::varchar
			)`

		if _, err := s.pool.Exec(ctx, query, entry.name, entry.role, entry.school,
			entry.designation,
			"Placeholder directory entry created by the demo seed. Replace it with real faculty details before using this in public.",
			entry.interests, entry.accepting); err != nil {
			return fmt.Errorf("%s: %w", entry.name, err)
		}
	}
	return nil
}

// ------------------------------------------------------------- announcements

func (s *seeder) seedAnnouncements(ctx context.Context, authorID string) error {
	now := time.Now()

	announcements := []struct {
		title    string
		body     string
		category string
		pinned   bool
	}{
		{
			"CampusCare AI is live",
			"This is the campus companion: events you can register for, clubs you can join, opportunities worth applying to, and a service desk when something needs fixing. Everything is free to join.",
			"GENERAL", true,
		},
		{
			"Registrations close a day before each event",
			"Every event on the calendar has a registration deadline, usually the day before. If you plan to attend, register before it closes rather than on the day.",
			"EVENTS", false,
		},
		{
			"Tell us what is not working",
			"If a booking, a listing or a club page is wrong, open a request from the service desk. Bad data is the fastest thing to fix once someone reports it.",
			"SERVICES", false,
		},
	}

	for i, announcement := range announcements {
		// Stagger the publish times so the board is not a wall of one timestamp.
		publishedAt := now.Add(-time.Duration(i) * time.Hour)

		const query = `INSERT INTO announcements (title, body, category, audience,
				is_pinned, status, published_at, author_id)
			SELECT $1::varchar, $2::text, $3::varchar, 'ALL', $4::boolean,
				'PUBLISHED', $5::timestamptz, $6::uuid
			WHERE NOT EXISTS (
				SELECT 1 FROM announcements WHERE title = $1::varchar
			)`

		if _, err := s.pool.Exec(ctx, query, announcement.title, announcement.body,
			announcement.category, announcement.pinned, publishedAt, authorID); err != nil {
			return fmt.Errorf("%s: %w", announcement.title, err)
		}
	}
	return nil
}

// slugify mirrors the repository's club slug rules so seeds and application
// writes agree on the same URL fragment.
func slugify(name string) string {
	var out []byte
	lastHyphen := true
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r >= 'A' && r <= 'Z':
			out = append(out, byte(r))
			lastHyphen = false
		default:
			if !lastHyphen {
				out = append(out, '-')
				lastHyphen = true
			}
		}
	}

	slug := string(out)
	for len(slug) > 0 && slug[0] == '-' {
		slug = slug[1:]
	}
	for len(slug) > 0 && slug[len(slug)-1] == '-' {
		slug = slug[:len(slug)-1]
	}
	if slug == "" {
		return "club"
	}
	return slug
}
