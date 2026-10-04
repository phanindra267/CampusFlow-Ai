package domain

import "time"

type EventStatus string

const (
	EventStatusDraft     EventStatus = "DRAFT"
	EventStatusPublished EventStatus = "PUBLISHED"
	EventStatusCancelled EventStatus = "CANCELLED"
	EventStatusCompleted EventStatus = "COMPLETED"
)

type Event struct {
	ID                   string      `json:"id"`
	Title                string      `json:"title" binding:"required"`
	Description          string      `json:"description" binding:"required"`
	Category             string      `json:"category" binding:"required"`
	OrganizerID          string      `json:"organizer_id"`
	Venue                string      `json:"venue"`
	IsOnline             bool        `json:"is_online"`
	StartTime            time.Time   `json:"start_time" binding:"required"`
	EndTime              time.Time   `json:"end_time" binding:"required"`
	RegistrationDeadline time.Time   `json:"registration_deadline" binding:"required"`
	Capacity             int         `json:"capacity" binding:"required,min=1"`
	Status               EventStatus `json:"status"`
	RegistrationStatus   string      `json:"registration_status,omitempty"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`

	// Eligibility is free text written by the organiser: "second year and
	// above", "members of any technical club". CampusCare AI does not evaluate
	// it against an academic record.
	Eligibility string `json:"eligibility,omitempty"`

	// Resolved for the requesting member so the UI can show saved and
	// reminder state without extra round trips.
	IsSaved         bool       `json:"is_saved"`
	ReminderAt      *time.Time `json:"reminder_at,omitempty"`
	RegisteredCount int        `json:"registered_count"`
	ClubID          *string    `json:"club_id,omitempty"`
	ClubName        string     `json:"club_name,omitempty"`
}

// CreateEventRequest is the organiser-facing write model for an event.
type CreateEventRequest struct {
	Title                string    `json:"title" binding:"required,min=3,max=255"`
	Description          string    `json:"description" binding:"required,min=10"`
	Category             string    `json:"category" binding:"required,max=100"`
	Venue                string    `json:"venue" binding:"max=255"`
	IsOnline             bool      `json:"is_online"`
	StartTime            time.Time `json:"start_time" binding:"required"`
	EndTime              time.Time `json:"end_time" binding:"required"`
	RegistrationDeadline time.Time `json:"registration_deadline" binding:"required"`
	Capacity             int       `json:"capacity" binding:"required,min=1"`
	Eligibility          string    `json:"eligibility" binding:"max=2000"`
	ClubID               *string   `json:"club_id,omitempty"`
	Status               string    `json:"status" binding:"omitempty,oneof=DRAFT PUBLISHED"`
}

// SetReminderRequest schedules a reminder for the requesting member.
type SetReminderRequest struct {
	RemindAt time.Time `json:"remind_at" binding:"required"`
}

// EventRegistrationDetail is one row of an organiser's registration list.
type EventRegistrationDetail struct {
	ID          string     `json:"id"`
	MemberID    string     `json:"member_id"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CheckedInAt *time.Time `json:"checked_in_at,omitempty"`
}

// OrganizerEventAnalytics is the evidence an organiser needs after an event.
type OrganizerEventAnalytics struct {
	EventID        string  `json:"event_id"`
	Title          string  `json:"title"`
	Capacity       int     `json:"capacity"`
	Registered     int     `json:"registered"`
	Cancelled      int     `json:"cancelled"`
	Attended       int     `json:"attended"`
	Waitlisted     int     `json:"waitlisted"`
	FillRate       float64 `json:"fill_rate"`
	CheckInRate    float64 `json:"check_in_rate"`
	RemindersSet   int     `json:"reminders_set"`
	SavedCount     int     `json:"saved_count"`
	AverageRating  float64 `json:"average_rating"`
	RatingsCount   int     `json:"ratings_count"`
	DemandVsSupply float64 `json:"demand_vs_supply"`
}

type RegistrationStatus string

const (
	RegStatusRegistered RegistrationStatus = "REGISTERED"
	RegStatusCancelled  RegistrationStatus = "CANCELLED"
	RegStatusAttended   RegistrationStatus = "ATTENDED"
)

type EventRegistration struct {
	ID        string             `json:"id"`
	EventID   string             `json:"event_id"`
	MemberID  string             `json:"member_id"`
	Status    RegistrationStatus `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
}
