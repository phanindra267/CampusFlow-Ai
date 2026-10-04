package domain

import "time"

// CampusCare AI's service desk answers two kinds of question: "how do I get this
// done on campus" (a request) and "something here is wrong" (a complaint). Both
// are the same ticket with a shared status timeline, because what a member wants
// either way is to know where their issue currently stands.
type ServiceRequestKind string

const (
	ServiceRequestKindRequest   ServiceRequestKind = "REQUEST"
	ServiceRequestKindComplaint ServiceRequestKind = "COMPLAINT"
)

type ServiceRequestStatus string

const (
	ServiceStatusOpen       ServiceRequestStatus = "OPEN"
	ServiceStatusInProgress ServiceRequestStatus = "IN_PROGRESS"
	ServiceStatusResolved   ServiceRequestStatus = "RESOLVED"
	ServiceStatusRejected   ServiceRequestStatus = "REJECTED"
	ServiceStatusClosed     ServiceRequestStatus = "CLOSED"
)

// ServiceRequestPriority is urgency, set by the member at submission and
// reviewed by staff, not a grade or an academic standing.
type ServiceRequestPriority string

const (
	ServicePriorityLow    ServiceRequestPriority = "LOW"
	ServicePriorityNormal ServiceRequestPriority = "NORMAL"
	ServicePriorityHigh   ServiceRequestPriority = "HIGH"
	ServicePriorityUrgent ServiceRequestPriority = "URGENT"
)

// ServiceRequest is one ticket in the campus service desk. Reference is the
// short code a member quotes at a counter; ID stays internal.
type ServiceRequest struct {
	ID            string                 `json:"id"`
	Reference     string                 `json:"reference"`
	ServiceID     *string                `json:"service_id,omitempty"`
	ServiceName   string                 `json:"service_name,omitempty"`
	UserID        string                 `json:"user_id"`
	RequesterName string                 `json:"requester_name,omitempty"`
	Kind          ServiceRequestKind     `json:"kind"`
	Subject       string                 `json:"subject"`
	Description   string                 `json:"description"`
	Status        ServiceRequestStatus   `json:"status"`
	Priority      ServiceRequestPriority `json:"priority"`
	AssignedTo    *string                `json:"assigned_to,omitempty"`
	AssignedName  string                 `json:"assigned_name,omitempty"`
	Resolution    string                 `json:"resolution,omitempty"`
	ResolvedAt    *time.Time             `json:"resolved_at,omitempty"`
	Rating        *int                   `json:"rating,omitempty"`
	RatingComment string                 `json:"rating_comment,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// CreateServiceRequestRequest is what a member submits. The service is chosen
// from the campus service catalogue; kind decides whether this is a question or
// a complaint.
type CreateServiceRequestRequest struct {
	ServiceID   *string `json:"service_id,omitempty"`
	Kind        string  `json:"kind" binding:"required,oneof=REQUEST COMPLAINT"`
	Subject     string  `json:"subject" binding:"required,min=5,max=255"`
	Description string  `json:"description" binding:"required,min=20,max=5000"`
	Priority    string  `json:"priority" binding:"omitempty,oneof=LOW NORMAL HIGH URGENT"`
}

// UpdateServiceStatusRequest moves a ticket along. Staff own the status and the
// resolution; a member may only close or reopen their own ticket, which the
// handler enforces.
type UpdateServiceStatusRequest struct {
	Status     string  `json:"status" binding:"required,oneof=OPEN IN_PROGRESS RESOLVED REJECTED CLOSED"`
	Note       string  `json:"note" binding:"omitempty,max=2000"`
	Resolution string  `json:"resolution" binding:"omitempty,max=4000"`
	AssignedTo *string `json:"assigned_to,omitempty"`
}

// AddServiceRequestMessageRequest adds a note to the ticket timeline without
// changing its status.
type AddServiceRequestMessageRequest struct {
	Message string `json:"message" binding:"required,min=2,max=2000"`
}

// ServiceRequestEvent is one entry in a ticket's immutable timeline.
type ServiceRequestEvent struct {
	ID         string    `json:"id"`
	RequestID  string    `json:"request_id"`
	EventType  string    `json:"event_type"`
	FromStatus string    `json:"from_status,omitempty"`
	ToStatus   string    `json:"to_status,omitempty"`
	ActorID    *string   `json:"actor_id,omitempty"`
	ActorName  string    `json:"actor_name,omitempty"`
	IsStaff    bool      `json:"is_staff"`
	Message    string    `json:"message,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// ServiceRequestAttachment records that a file belongs to a ticket.
// CampusCare AI stores the descriptor and where the file lives; it does not
// host upload bytes, so storage_url is required and validated as a URL.
type ServiceRequestAttachment struct {
	ID         string    `json:"id"`
	RequestID  string    `json:"request_id"`
	FileName   string    `json:"file_name"`
	FileURL    string    `json:"file_url"`
	MimeType   string    `json:"mime_type,omitempty"`
	SizeBytes  *int64    `json:"size_bytes,omitempty"`
	UploadedBy string    `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

// AddServiceAttachmentRequest registers an attachment descriptor against a
// ticket. FileURL points at an already-uploaded object.
type AddServiceAttachmentRequest struct {
	FileName  string `json:"file_name" binding:"required,min=1,max=255"`
	FileURL   string `json:"file_url" binding:"required,url,max=512"`
	MimeType  string `json:"mime_type" binding:"omitempty,max=120"`
	SizeBytes *int64 `json:"size_bytes,omitempty" binding:"omitempty,min=1"`
}

// ServiceRatingRequest is a member's satisfaction score for a handled ticket.
type ServiceRatingRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment" binding:"omitempty,max=2000"`
}

// ServiceFAQ is an entry on a service page. Position controls display order so
// the most-asked question can be pinned first.
type ServiceFAQ struct {
	ID        string    `json:"id"`
	ServiceID string    `json:"service_id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateServiceFAQRequest adds an FAQ entry to a service page.
type CreateServiceFAQRequest struct {
	Question string `json:"question" binding:"required,min=5,max=500"`
	Answer   string `json:"answer" binding:"required,min=10,max=5000"`
	Position int    `json:"position" binding:"omitempty,min=0,max=999"`
}

// ServiceQueueStats powers the service desk dashboard.
type ServiceQueueStats struct {
	Open               int     `json:"open"`
	InProgress         int     `json:"in_progress"`
	Resolved           int     `json:"resolved"`
	Closed             int     `json:"closed"`
	Rejected           int     `json:"rejected"`
	Complaints         int     `json:"complaints"`
	UrgentOpen         int     `json:"urgent_open"`
	AvgResolutionHours float64 `json:"avg_resolution_hours"`
}
