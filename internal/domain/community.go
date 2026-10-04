package domain

import "time"

// CampusCare AI models a community: the people on campus, the groups they
// organise, what they gather for, and how they talk to each other. This file
// holds the collaboration primitives that the community product is built on.

type DiscussionCategory string

const (
	DiscussionCategoryGeneral      DiscussionCategory = "GENERAL"
	DiscussionCategoryAnnouncement DiscussionCategory = "ANNOUNCEMENT"
	DiscussionCategoryQuestion     DiscussionCategory = "QUESTION"
	DiscussionCategoryEvent        DiscussionCategory = "EVENT"
	DiscussionCategoryOpportunity  DiscussionCategory = "OPPORTUNITY"
	DiscussionCategoryResearch     DiscussionCategory = "RESEARCH"
)

type DiscussionStatus string

const (
	DiscussionStatusOpen     DiscussionStatus = "OPEN"
	DiscussionStatusLocked   DiscussionStatus = "LOCKED"
	DiscussionStatusArchived DiscussionStatus = "ARCHIVED"
)

// Discussion is a thread. A nil ClubID marks a campus-wide thread that is not
// scoped to a single club or group.
type Discussion struct {
	ID             string             `json:"id"`
	ClubID         *string            `json:"club_id,omitempty"`
	AuthorID       string             `json:"author_id"`
	AuthorName     string             `json:"author_name,omitempty"`
	Title          string             `json:"title" binding:"required"`
	Body           string             `json:"body" binding:"required"`
	Category       DiscussionCategory `json:"category"`
	Status         DiscussionStatus   `json:"status"`
	IsPinned       bool               `json:"is_pinned"`
	ReplyCount     int                `json:"reply_count"`
	LastActivityAt time.Time          `json:"last_activity_at"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// DiscussionReply is a response to a thread. ParentID supports one level of
// threading so replies can be grouped without unbounded nesting.
type DiscussionReply struct {
	ID           string    `json:"id"`
	DiscussionID string    `json:"discussion_id"`
	AuthorID     string    `json:"author_id"`
	AuthorName   string    `json:"author_name,omitempty"`
	ParentID     *string   `json:"parent_id,omitempty"`
	Body         string    `json:"body" binding:"required"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type NotificationType string

const (
	NotificationTypeEvent        NotificationType = "EVENT"
	NotificationTypeClub         NotificationType = "CLUB"
	NotificationTypeOpportunity  NotificationType = "OPPORTUNITY"
	NotificationTypeDiscussion   NotificationType = "DISCUSSION"
	NotificationTypeSystem       NotificationType = "SYSTEM"
	NotificationTypeService      NotificationType = "SERVICE"
	NotificationTypeAnnouncement NotificationType = "ANNOUNCEMENT"
)

type Notification struct {
	ID        string           `json:"id"`
	UserID    string           `json:"user_id"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Body      string           `json:"body"`
	Link      *string          `json:"link,omitempty"`
	ReadAt    *time.Time       `json:"read_at,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

type ApplicationStatus string

const (
	ApplicationStatusSubmitted   ApplicationStatus = "SUBMITTED"
	ApplicationStatusUnderReview ApplicationStatus = "UNDER_REVIEW"
	ApplicationStatusShortlisted ApplicationStatus = "SHORTLISTED"
	ApplicationStatusAccepted    ApplicationStatus = "ACCEPTED"
	ApplicationStatusRejected    ApplicationStatus = "REJECTED"
	ApplicationStatusWithdrawn   ApplicationStatus = "WITHDRAWN"
)

// OpportunityApplication is a member's application to an internship, research
// role, workshop or similar listing.
type OpportunityApplication struct {
	ID            string            `json:"id"`
	OpportunityID string            `json:"opportunity_id"`
	UserID        string            `json:"user_id"`
	Title         string            `json:"title,omitempty"`
	Status        ApplicationStatus `json:"status"`
	CoverNote     string            `json:"cover_note"`
	ResumeURL     *string           `json:"resume_url,omitempty"`
	AppliedAt     time.Time         `json:"applied_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// CampusResource is a bookable or browsable facility or service on campus.
// CampusResource is a bookable or browsable facility or service on campus.
// ResourceType carries the catalogue value (ROOM, LAB, SPORTS, FOOD, ...);
// Description, Location and the booking flags are community-facing columns
// added in migration 000035.
type CampusResource struct {
	ID           string    `json:"id"`
	Name         string    `json:"name" binding:"required"`
	ResourceType string    `json:"resource_type" binding:"required"`
	Description  string    `json:"description"`
	Location     string    `json:"location"`
	Capacity     int       `json:"capacity"`
	IsBookable   bool      `json:"is_bookable"`
	RequiresAuth bool      `json:"requires_auth"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ResourceBookingStatus string

const (
	BookingStatusConfirmed ResourceBookingStatus = "CONFIRMED"
	BookingStatusCancelled ResourceBookingStatus = "CANCELLED"
)

type ResourceBooking struct {
	ID         string                `json:"id"`
	ResourceID string                `json:"resource_id"`
	ResourceNm string                `json:"resource_name,omitempty"`
	UserID     string                `json:"user_id"`
	UserName   string                `json:"user_name,omitempty"`
	StartTime  time.Time             `json:"start_time" binding:"required"`
	EndTime    time.Time             `json:"end_time" binding:"required"`
	Note       string                `json:"note"`
	Status     ResourceBookingStatus `json:"status"`
	CreatedAt  time.Time             `json:"created_at"`
}

type SearchResult struct {
	EntityType string            `json:"entity_type"`
	EntityID   string            `json:"entity_id"`
	Title      string            `json:"title"`
	Summary    string            `json:"summary"`
	Category   string            `json:"category"`
	URL        string            `json:"url"`
	EventAt    *time.Time        `json:"event_at,omitempty"`
	Meta       map[string]string `json:"meta,omitempty"`
	Score      float64           `json:"score"`
}

type SearchResults struct {
	Query   string         `json:"query"`
	Total   int            `json:"total"`
	Results []SearchResult `json:"results"`
	ByType  map[string]int `json:"by_type"`
}
