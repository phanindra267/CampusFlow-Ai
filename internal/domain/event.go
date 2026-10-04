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
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
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
	StudentID string             `json:"student_id"`
	Status    RegistrationStatus `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
}
