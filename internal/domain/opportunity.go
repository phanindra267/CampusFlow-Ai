package domain

import "time"

type Opportunity struct {
	ID                   string    `json:"id"`
	Title                string    `json:"title" binding:"required"`
	Description          string    `json:"description" binding:"required"`
	Type                 string    `json:"type" binding:"required,oneof=HACKATHON COMPETITION WORKSHOP SEMINAR INTERNSHIP PLACEMENT RESEARCH PROJECT ENTREPRENEURSHIP SCHOLARSHIP JOB VOLUNTEERING STARTUP"`
	Category             string    `json:"category" binding:"required,max=100"`
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

	// Eligibility is plain text from the poster. Eligibility is never
	// auto-judged against an academic record.
	Eligibility string `json:"eligibility,omitempty"`
	// Source distinguishes who is running it (INTERNAL, CLUB, EXTERNAL) and
	// ApplyURL lets external listings send a member off-site instead of
	// collecting an in-app application.
	Source           string `json:"source,omitempty"`
	ApplyURL         string `json:"apply_url,omitempty"`
	ApplicationCount int    `json:"application_count"`
	IsSaved          bool   `json:"is_saved"`
}

// Opportunity types. Kept as constants because the UI filters on them and the
// migration enforces them with a CHECK constraint, so these values must stay in
// step with opportunities_type_check.
const (
	OpportunityTypeHackathon        = "HACKATHON"
	OpportunityTypeCompetition      = "COMPETITION"
	OpportunityTypeWorkshop         = "WORKSHOP"
	OpportunityTypeSeminar          = "SEMINAR"
	OpportunityTypeInternship       = "INTERNSHIP"
	OpportunityTypePlacement        = "PLACEMENT"
	OpportunityTypeResearch         = "RESEARCH"
	OpportunityTypeProject          = "PROJECT"
	OpportunityTypeEntrepreneurship = "ENTREPRENEURSHIP"
	OpportunityTypeScholarship      = "SCHOLARSHIP"
	OpportunityTypeJob              = "JOB"
	OpportunityTypeVolunteering     = "VOLUNTEERING"
	OpportunityTypeStartup          = "STARTUP"
)

// Sources describe who is offering the opportunity, which is what a member
// weighs when deciding whether an application is worth their time.
const (
	OpportunitySourceInternal = "INTERNAL"
	OpportunitySourceClub     = "CLUB"
	OpportunitySourceExternal = "EXTERNAL"
)

// CreateOpportunityRequest is the organiser-facing write model.
type CreateOpportunityRequest struct {
	Title                string    `json:"title" binding:"required,min=3,max=255"`
	Description          string    `json:"description" binding:"required,min=20"`
	Type                 string    `json:"type" binding:"required,max=60"`
	Category             string    `json:"category" binding:"required,max=100"`
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
	RegistrationDeadline time.Time `json:"registration_deadline"`
	Location             string    `json:"location" binding:"max=255"`
	DeliveryMode         string    `json:"delivery_mode" binding:"omitempty,max=60"`
	Capacity             int       `json:"capacity" binding:"omitempty,min=0"`
	Eligibility          string    `json:"eligibility" binding:"max=2000"`
	Source               string    `json:"source" binding:"omitempty,oneof=INTERNAL CLUB EXTERNAL"`
	ApplyURL             string    `json:"apply_url" binding:"omitempty,url,max=512"`
	ClubID               *string   `json:"club_id,omitempty"`
}

// ApplicationDetail is one row of a poster's applicant list.
type ApplicationDetail struct {
	ID            string    `json:"id"`
	OpportunityID string    `json:"opportunity_id"`
	UserID        string    `json:"user_id"`
	DisplayName   string    `json:"display_name"`
	Email         string    `json:"email"`
	Program       string    `json:"program,omitempty"`
	YearOfStudy   *int      `json:"year_of_study,omitempty"`
	Status        string    `json:"status"`
	CoverNote     string    `json:"cover_note,omitempty"`
	ResumeURL     string    `json:"resume_url,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UpdateApplicationStatusRequest lets a poster move an application forward.
type UpdateApplicationStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=SUBMITTED UNDER_REVIEW SHORTLISTED ACCEPTED REJECTED WITHDRAWN"`
}

type SavedOpportunity struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	OpportunityID string    `json:"opportunity_id"`
	CreatedAt     time.Time `json:"created_at"`
}
