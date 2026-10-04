package domain

import "time"

type Opportunity struct {
	ID                   string    `json:"id"`
	Title                string    `json:"title" binding:"required"`
	Description          string    `json:"description" binding:"required"`
	Type                 string    `json:"type" binding:"required"` // WORKSHOP, HACKATHON, etc.
	Category             string    `json:"category" binding:"required"`
	OrganizerID          string    `json:"organizer_id"`
	ClubID               *string   `json:"club_id,omitempty"`
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
	RegistrationDeadline time.Time `json:"registration_deadline"`
	Location             string    `json:"location"`
	DeliveryMode         string    `json:"delivery_mode"`
	Capacity             int       `json:"capacity"`
	Status               string    `json:"status"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type SavedOpportunity struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	OpportunityID string    `json:"opportunity_id"`
	CreatedAt     time.Time `json:"created_at"`
}
