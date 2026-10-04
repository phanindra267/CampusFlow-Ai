package domain

import "time"

type WaitlistEntry struct {
	ID         string     `json:"id"`
	EventID    string     `json:"event_id"`
	StudentID  string     `json:"student_id"`
	Status     string     `json:"status"` // WAITING, PROMOTED, EXPIRED
	JoinedAt   time.Time  `json:"joined_at"`
	PromotedAt *time.Time `json:"promoted_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

type EventSession struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Location  string    `json:"location"`
}

type CheckInSession struct {
	ID             string    `json:"id"`
	EventSessionID string    `json:"event_session_id"`
	OrganizerID    string    `json:"organizer_id"`
	QRToken        string    `json:"qr_token"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type AttendanceRecord struct {
	ID             string    `json:"id"`
	EventSessionID string    `json:"event_session_id"`
	StudentID      string    `json:"student_id"`
	CheckInMethod  string    `json:"check_in_method"`
	CheckedInAt    time.Time `json:"checked_in_at"`
	Status         string    `json:"status"`
}

type Feedback struct {
	ID           string    `json:"id"`
	ActivityType string    `json:"activity_type" binding:"required"`
	ActivityID   string    `json:"activity_id" binding:"required"`
	StudentID    string    `json:"student_id"`
	Rating       int       `json:"rating" binding:"required,min=1,max=5"`
	Comment      string    `json:"comment"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type EngagementActivity struct {
	ID           string    `json:"id"`
	StudentID    string    `json:"student_id"`
	ActivityType string    `json:"activity_type"`
	ActivityID   string    `json:"activity_id"`
	EventType    string    `json:"event_type"`
	CreatedAt    time.Time `json:"created_at"`
}
