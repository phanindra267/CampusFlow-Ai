package domain

import (
	"fmt"
	"net/mail"
	"strings"
	"time"
)

// NormalizeEmail lower-cases and trims an address so that accounts cannot be
// duplicated by varying case or padding.
func NormalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

// ValidateEmail checks a normalised address for a plausible mail shape.
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return fmt.Errorf("%q is not a valid email address", email)
	}
	return nil
}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	PasswordHash string    `json:"-"` // never expose
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Profile context. All optional and self-reported: the profile helps members
	// find relevant events, opportunities and research, and it is deliberately
	// not an academic record. There are no marks, grades or attendance fields.
	Identifier          string   `json:"identifier,omitempty"`
	Program             string   `json:"program,omitempty"`
	Branch              string   `json:"branch,omitempty"`
	YearOfStudy         *int     `json:"year_of_study,omitempty"`
	Headline            string   `json:"headline,omitempty"`
	Bio                 string   `json:"bio,omitempty"`
	AvatarURL           string   `json:"avatar_url,omitempty"`
	CareerInterests     []string `json:"career_interests,omitempty"`
	OpenToOpportunities bool     `json:"open_to_opportunities"`
}

// LoginRequest carries sign-in credentials.
//
// Email is validated by the handler rather than by a binding tag, because the
// standard `email` validator rejects values with surrounding whitespace that a
// user can easily paste in.
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	// Role is accepted for backwards compatibility but ignored: public
	// registration always provisions DefaultRegistrationRole. Elevated roles
	// are granted out of band by an administrator.
	Role string `json:"role"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateProfileRequest struct {
	// DisplayName is how the member appears on discussions, group rosters and
	// event check-in lists, so it is deliberately short and plain text.
	DisplayName string `json:"display_name" binding:"required,min=2,max=120"`

	// Everything below is optional. Pointers distinguish "leave unchanged"
	// from "clear this field", which matters for a profile the member edits
	// piecemeal from the UI.
	Identifier          *string   `json:"identifier,omitempty" binding:"omitempty,max=64"`
	Program             *string   `json:"program,omitempty" binding:"omitempty,max=255"`
	Branch              *string   `json:"branch,omitempty" binding:"omitempty,max=255"`
	YearOfStudy         *int      `json:"year_of_study,omitempty" binding:"omitempty,gte=1,lte=10"`
	Headline            *string   `json:"headline,omitempty" binding:"omitempty,max=255"`
	Bio                 *string   `json:"bio,omitempty" binding:"omitempty,max=2000"`
	AvatarURL           *string   `json:"avatar_url,omitempty" binding:"omitempty,max=512"`
	CareerInterests     *[]string `json:"career_interests,omitempty" binding:"omitempty,max=20,dive,max=80"`
	OpenToOpportunities *bool     `json:"open_to_opportunities,omitempty"`
}

// Profile is the member-facing view of an account: the profile itself plus the
// collections a member accumulates by taking part.
type Profile struct {
	User               User                      `json:"user"`
	ResearchInterests  []string                  `json:"research_interests"`
	Clubs              []ProfileClub             `json:"clubs"`
	Following          []ProfileClub             `json:"following"`
	SavedEvents        []ProfileSavedEvent       `json:"saved_events"`
	SavedOpportunities []ProfileSavedOpportunity `json:"saved_opportunities"`
	Applications       []OpportunityApplication  `json:"applications"`
	Registrations      []ProfileRegistration     `json:"registrations"`
	ServiceRequests    []ServiceRequest          `json:"service_requests"`
}

type ProfileClub struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Category      string `json:"category,omitempty"`
	Role          string `json:"role,omitempty"`
	FollowerCount int    `json:"follower_count"`
	MemberCount   int    `json:"member_count"`
}

type ProfileSavedEvent struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Category  string    `json:"category,omitempty"`
	StartTime time.Time `json:"start_time"`
	Venue     string    `json:"venue,omitempty"`
}

type ProfileSavedOpportunity struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Type     string    `json:"type"`
	Category string    `json:"category,omitempty"`
	Deadline time.Time `json:"registration_deadline"`
}

type ProfileRegistration struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	Title     string    `json:"title"`
	Category  string    `json:"category,omitempty"`
	StartTime time.Time `json:"start_time"`
	Venue     string    `json:"venue,omitempty"`
	Status    string    `json:"status"`
}

// DefaultRegistrationRole is the role assigned to every self-registered
// account, regardless of the role requested by the client. CampusCare AI is a
// community platform, so the baseline role is member; organiser and admin
// privileges are granted out of band.
const DefaultRegistrationRole = "MEMBER"
