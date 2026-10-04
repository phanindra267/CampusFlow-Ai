package domain

import "time"

// Announcement is a campus-wide notice. Announcements are how a member finds
// out about a deadline or a closure without subscribing to anything. PublishAt
// and ExpiresAt schedule the visibility window; IsPinned keeps a notice at the
// top of the list.
type Announcement struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	Category    string     `json:"category"`
	Audience    string     `json:"audience"`
	IsPinned    bool       `json:"is_pinned"`
	Status      string     `json:"status"`
	PublishAt   *time.Time `json:"publish_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	AuthorID    string     `json:"author_id"`
	AuthorName  string     `json:"author_name,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateAnnouncementRequest is the admin write model. Status defaults to
// PUBLISHED so a same-day notice does not need a second call.
type CreateAnnouncementRequest struct {
	Title     string     `json:"title" binding:"required,min=5,max=255"`
	Body      string     `json:"body" binding:"required,min=20,max=20000"`
	Category  string     `json:"category" binding:"required,max=60"`
	Audience  string     `json:"audience" binding:"omitempty,oneof=ALL MEMBERS STUDENTS FACULTY STAFF CLUBS ORGANISERS ADMINS"`
	IsPinned  bool       `json:"is_pinned"`
	Status    string     `json:"status" binding:"omitempty,oneof=DRAFT PUBLISHED ARCHIVED"`
	PublishAt *time.Time `json:"publish_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// UpdateAnnouncementRequest patches an announcement. Pointer fields let an
// editor clear a pinned flag or an expiry without resending the whole body.
type UpdateAnnouncementRequest struct {
	Title     *string    `json:"title,omitempty" binding:"omitempty,min=5,max=255"`
	Body      *string    `json:"body,omitempty" binding:"omitempty,min=20,max=20000"`
	Category  *string    `json:"category,omitempty" binding:"omitempty,max=60"`
	Audience  *string    `json:"audience,omitempty" binding:"omitempty,oneof=ALL MEMBERS STUDENTS FACULTY STAFF CLUBS ORGANISERS ADMINS"`
	IsPinned  *bool      `json:"is_pinned,omitempty"`
	Status    *string    `json:"status,omitempty" binding:"omitempty,oneof=DRAFT PUBLISHED ARCHIVED"`
	PublishAt *time.Time `json:"publish_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
