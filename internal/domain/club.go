package domain

import "time"

type ClubStatus string

const (
	ClubStatusPending   ClubStatus = "PENDING"
	ClubStatusActive    ClubStatus = "ACTIVE"
	ClubStatusSuspended ClubStatus = "SUSPENDED"
	ClubStatusArchived  ClubStatus = "ARCHIVED"
)

type Club struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name" binding:"required"`
	Slug               string     `json:"slug" binding:"required"`
	Description        string     `json:"description"`
	Category           string     `json:"category"`
	VerificationStatus string     `json:"verification_status"`
	Status             ClubStatus `json:"status"`
	CreatedBy          string     `json:"created_by"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	// Presentation and contact fields so a club page looks like the club
	// rather than a name and a paragraph.
	LogoURL       string `json:"logo_url,omitempty"`
	ContactEmail  string `json:"contact_email,omitempty"`
	MemberCount   int    `json:"member_count"`
	FollowerCount int    `json:"follower_count"`
	IsFollowing   bool   `json:"is_following"`
}

// CreateClubRequest is the organiser-facing write model for a club. Slug is
// derived from the name when omitted so callers do not hand-pick URLs.
type CreateClubRequest struct {
	Name         string `json:"name" binding:"required,min=3,max=255"`
	Description  string `json:"description" binding:"required,min=20"`
	Category     string `json:"category" binding:"required,max=100"`
	LogoURL      string `json:"logo_url" binding:"omitempty,max=512"`
	ContactEmail string `json:"contact_email" binding:"omitempty,email,max=255"`
}

// ClubActivity is something a club did independently of the campus event
// calendar: a meeting, a workshop it ran, a recruitment drive.
type ClubActivity struct {
	ID          string     `json:"id"`
	ClubID      string     `json:"club_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Type        string     `json:"type,omitempty"`
	Location    string     `json:"location,omitempty"`
	StartsAt    *time.Time `json:"starts_at,omitempty"`
	EndsAt      *time.Time `json:"ends_at,omitempty"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateClubActivityRequest records one club activity.
type CreateClubActivityRequest struct {
	Title       string     `json:"title" binding:"required,min=3,max=255"`
	Description string     `json:"description" binding:"omitempty,max=5000"`
	Type        string     `json:"type" binding:"omitempty,max=100"`
	Location    string     `json:"location" binding:"omitempty,max=255"`
	StartsAt    *time.Time `json:"starts_at,omitempty"`
	EndsAt      *time.Time `json:"ends_at,omitempty"`
}

// ClubProject is a body of work a club is running.
type ClubProject struct {
	ID          string    `json:"id"`
	ClubID      string    `json:"club_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateClubProjectRequest records one club project.
type CreateClubProjectRequest struct {
	Title       string `json:"title" binding:"required,min=3,max=255"`
	Description string `json:"description" binding:"omitempty,max=5000"`
	Status      string `json:"status" binding:"omitempty,oneof=PLANNED ACTIVE COMPLETED PAUSED"`
}

// ClubAnalytics is the picture a club office gets before its next meeting.
type ClubAnalytics struct {
	ClubID              string          `json:"club_id"`
	MemberCount         int             `json:"member_count"`
	FollowerCount       int             `json:"follower_count"`
	ActiveMembers30d    int             `json:"active_members_30d"`
	Activities          int             `json:"activities"`
	EventsHosted        int             `json:"events_hosted"`
	EventRegistrations  int             `json:"event_registrations"`
	OpportunitiesPosted int             `json:"opportunities_posted"`
	AverageRating       float64         `json:"average_rating"`
	RatingsCount        int             `json:"ratings_count"`
	TopCategories       []CategoryCount `json:"top_categories"`
	MostActiveMonths    []CategoryCount `json:"most_active_months"`
}

type CategoryCount struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type ClubMembershipRole string

const (
	ClubRoleOwner   ClubMembershipRole = "OWNER"
	ClubRoleAdmin   ClubMembershipRole = "ADMIN"
	ClubRoleOfficer ClubMembershipRole = "OFFICER"
	ClubRoleMember  ClubMembershipRole = "MEMBER"
)

type ClubMembership struct {
	ID        string             `json:"id"`
	ClubID    string             `json:"club_id"`
	UserID    string             `json:"user_id"`
	Role      ClubMembershipRole `json:"role"`
	Status    string             `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
}
