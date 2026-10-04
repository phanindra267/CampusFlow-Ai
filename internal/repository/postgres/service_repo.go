package postgres

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/campuscare/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ServiceRepository is the campus service desk. A member asks a question or
// files a complaint against a campus service; staff move the ticket through a
// status timeline until it is resolved. Attachment metadata is recorded, but the
// API never hosts the file itself.
type ServiceRepository struct {
	dbBase
}

func NewServiceRepository(db *pgxpool.Pool) *ServiceRepository {
	return &ServiceRepository{dbBase: dbBase{db: db}}
}

// serviceRequestColumns joins the service and the two people names a member sees
// on a ticket.
const serviceRequestColumns = `sr.id, sr.reference, sr.resource_id,
	COALESCE(cr.name, ''), sr.user_id, COALESCE(u.display_name, ''), sr.kind,
	sr.subject, sr.description, sr.status, sr.priority, sr.assigned_to,
	COALESCE(assignee.display_name, ''), COALESCE(sr.resolution, ''),
	sr.resolved_at, sr.feedback_rating, COALESCE(sr.feedback_comment, ''),
	sr.created_at, sr.updated_at`

const serviceRequestJoins = `FROM service_requests sr
	JOIN users u ON u.id = sr.user_id
	LEFT JOIN campus_resources cr ON cr.id = sr.resource_id
	LEFT JOIN users assignee ON assignee.id = sr.assigned_to`

func scanServiceRequest(row pgx.Row) (*domain.ServiceRequest, error) {
	var r domain.ServiceRequest
	if err := row.Scan(&r.ID, &r.Reference, &r.ServiceID, &r.ServiceName, &r.UserID,
		&r.RequesterName, &r.Kind, &r.Subject, &r.Description, &r.Status,
		&r.Priority, &r.AssignedTo, &r.AssignedName, &r.Resolution, &r.ResolvedAt,
		&r.Rating, &r.RatingComment, &r.CreatedAt, &r.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan service request: %w", err)
	}
	return &r, nil
}

// CreateServiceRequest opens a ticket. The reference is generated here rather
// than accepted from the client: it is the short code a member reads out at a
// counter, so it has to be unambiguous and cannot be chosen to collide with
// another ticket.
func (r *ServiceRepository) CreateServiceRequest(ctx context.Context, userID string, req *domain.CreateServiceRequestRequest) (*domain.ServiceRequest, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	priority := req.Priority
	if priority == "" {
		priority = string(domain.ServicePriorityNormal)
	}

	reference, err := newRequestReference()
	if err != nil {
		return nil, err
	}

	const query = `INSERT INTO service_requests (reference, user_id, resource_id,
			kind, subject, description, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	var id string
	if err := r.db.QueryRow(ctx, query, reference, userID, req.ServiceID, req.Kind,
		req.Subject, req.Description, priority).Scan(&id); err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		if isCheckViolation(err) {
			return nil, fmt.Errorf("unknown request kind or priority")
		}
		return nil, fmt.Errorf("create service request: %w", err)
	}

	// The opening entry makes the timeline self-explanatory: a member who opens
	// the ticket later sees when it was raised and by whom.
	if err := r.appendTimeline(ctx, id, &userID, nil, string(domain.ServiceStatusOpen),
		"Request submitted"); err != nil {
		return nil, err
	}

	return r.GetServiceRequest(ctx, id, userID)
}

// GetServiceRequest returns one ticket. Ownership is checked by the caller: the
// handler compares the returned user against the session before showing it.
func (r *ServiceRepository) GetServiceRequest(ctx context.Context, id, requesterID string) (*domain.ServiceRequest, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT ` + serviceRequestColumns + ` ` + serviceRequestJoins + `
		WHERE sr.id = $1`

	request, err := scanServiceRequest(r.db.QueryRow(ctx, query, id))
	if err != nil {
		return nil, fmt.Errorf("get service request: %w", err)
	}
	if requesterID != "" && request.UserID != requesterID {
		// A member may only read their own ticket; staff go through the queue
		// endpoint, which has no such restriction.
		return nil, ErrNotFound
	}
	return request, nil
}

// ListServiceRequests is either a member's own tickets or the staff queue. The
// two cases differ only in whose rows are returned, so they share one statement.
func (r *ServiceRepository) ListServiceRequests(ctx context.Context, requesterID, status, kind string, limit, offset int) ([]domain.ServiceRequest, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	// An empty requester means the staff queue: every ticket, not just one
	// member's.
	const query = `SELECT ` + serviceRequestColumns + ` ` + serviceRequestJoins + `
		WHERE ($2 = '' OR sr.status = NULLIF($2, ''))
			AND ($3 = '' OR sr.kind = NULLIF($3, ''))
			AND (NULLIF($1, '')::uuid IS NULL OR sr.user_id = NULLIF($1, '')::uuid)
		ORDER BY
			CASE sr.priority WHEN 'URGENT' THEN 0 WHEN 'HIGH' THEN 1
				WHEN 'NORMAL' THEN 2 ELSE 3 END,
			sr.created_at ASC
		LIMIT $4 OFFSET $5`

	rows, err := r.db.Query(ctx, query, requesterID, status, kind, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list service requests: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ServiceRequest, 0, limit)
	for rows.Next() {
		request, err := scanServiceRequest(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *request)
	}
	return out, rows.Err()
}

// UpdateServiceRequest moves a ticket and appends the change to its timeline in
// the same transaction, so the history cannot disagree with the current status.
// Resolving or rejecting a ticket stamps resolved_at and keeps the resolution
// text; both are required so a closed ticket always says what happened.
func (r *ServiceRepository) UpdateServiceRequest(ctx context.Context, id, actorID, status, note, resolution string, assignedTo *string) (*domain.ServiceRequest, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}

	closing := status == string(domain.ServiceStatusResolved) ||
		status == string(domain.ServiceStatusRejected) ||
		status == string(domain.ServiceStatusClosed)
	if closing && strings.TrimSpace(resolution) == "" && status != string(domain.ServiceStatusClosed) {
		return nil, fmt.Errorf("a %s ticket needs a resolution explaining what was done", strings.ToLower(status))
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin update request: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const current = `SELECT status FROM service_requests WHERE id = $1 FOR UPDATE`
	var fromStatus string
	if err := tx.QueryRow(ctx, current, id).Scan(&fromStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read request status: %w", err)
	}

	const query = `UPDATE service_requests SET
			status = $2,
			resolution = COALESCE(NULLIF($3, ''), resolution),
			assigned_to = COALESCE($4, assigned_to),
			resolved_at = CASE WHEN $5 THEN now() ELSE resolved_at END,
			updated_at = now()
		WHERE id = $1
		RETURNING id`
	if err := tx.QueryRow(ctx, query, id, status, resolution, assignedTo, closing).
		Scan(new(string)); err != nil {
		if isCheckViolation(err) {
			return nil, fmt.Errorf("unknown status %q", status)
		}
		return nil, fmt.Errorf("update service request: %w", err)
	}

	if err := appendTimelineTx(ctx, tx, id, &actorID, &fromStatus, status, note); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update request: %w", err)
	}

	return r.GetServiceRequest(ctx, id, "")
}

// AddServiceRequestMessage lets either side add context without changing status.
func (r *ServiceRepository) AddServiceRequestMessage(ctx context.Context, id, actorID, message string) (*domain.ServiceRequestEvent, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	if err := r.appendTimeline(ctx, id, &actorID, nil, "", message); err != nil {
		return nil, err
	}
	return r.latestTimelineEntry(ctx, id)
}

// ListServiceRequestTimeline returns the ticket history, oldest first.
func (r *ServiceRepository) ListServiceRequestTimeline(ctx context.Context, id string) ([]domain.ServiceRequestEvent, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT e.id, e.request_id,
			CASE WHEN e.from_status IS NULL AND e.to_status <> '' THEN 'CREATED'
				WHEN e.to_status = '' THEN 'MESSAGE' ELSE 'STATUS' END,
			COALESCE(e.from_status, ''), COALESCE(e.to_status, ''),
			e.actor_id, COALESCE(u.display_name, ''), COALESCE(e.note, ''), e.created_at,
			COALESCE(u.role, '') IN ('ORGANIZER', 'ADMIN', 'SUPER_ADMIN')
		FROM service_request_events e
		LEFT JOIN users u ON u.id = e.actor_id
		WHERE e.request_id = $1
		ORDER BY e.created_at ASC, e.id ASC`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("list request timeline: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ServiceRequestEvent, 0, 16)
	for rows.Next() {
		var e domain.ServiceRequestEvent
		if err := rows.Scan(&e.ID, &e.RequestID, &e.EventType, &e.FromStatus,
			&e.ToStatus, &e.ActorID, &e.ActorName, &e.Message, &e.CreatedAt,
			&e.IsStaff); err != nil {
			return nil, fmt.Errorf("scan timeline entry: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *ServiceRepository) latestTimelineEntry(ctx context.Context, requestID string) (*domain.ServiceRequestEvent, error) {
	const query = `SELECT e.id, e.request_id,
			CASE WHEN e.from_status IS NULL AND e.to_status <> '' THEN 'CREATED'
				WHEN e.to_status = '' THEN 'MESSAGE' ELSE 'STATUS' END,
			COALESCE(e.from_status, ''), COALESCE(e.to_status, ''),
			e.actor_id, COALESCE(u.display_name, ''), COALESCE(e.note, ''), e.created_at,
			COALESCE(u.role, '') IN ('ORGANIZER', 'ADMIN', 'SUPER_ADMIN')
		FROM service_request_events e
		LEFT JOIN users u ON u.id = e.actor_id
		WHERE e.request_id = $1
		ORDER BY e.created_at DESC, e.id DESC LIMIT 1`

	event, err := scanServiceRequestEvent(r.db.QueryRow(ctx, query, requestID))
	if err != nil {
		return nil, fmt.Errorf("read timeline entry: %w", err)
	}
	return event, nil
}

func scanServiceRequestEvent(row pgx.Row) (*domain.ServiceRequestEvent, error) {
	var e domain.ServiceRequestEvent
	if err := row.Scan(&e.ID, &e.RequestID, &e.EventType, &e.FromStatus, &e.ToStatus,
		&e.ActorID, &e.ActorName, &e.Message, &e.CreatedAt, &e.IsStaff); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan timeline entry: %w", err)
	}
	return &e, nil
}

// appendTimeline records one entry outside a caller-owned transaction.
func (r *ServiceRepository) appendTimeline(ctx context.Context, requestID string, actorID *string, fromStatus *string, toStatus, note string) error {
	if err := r.ready(); err != nil {
		return err
	}
	const query = `INSERT INTO service_request_events (request_id, actor_id,
			from_status, to_status, note)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''))`

	if _, err := r.db.Exec(ctx, query, requestID, actorID, fromStatus, toStatus,
		note); err != nil {
		if isForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		return fmt.Errorf("append timeline entry: %w", err)
	}
	return nil
}

func appendTimelineTx(ctx context.Context, tx pgx.Tx, requestID string, actorID, fromStatus *string, toStatus, note string) error {
	const query = `INSERT INTO service_request_events (request_id, actor_id,
			from_status, to_status, note)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''))`

	if _, err := tx.Exec(ctx, query, requestID, actorID, fromStatus, toStatus,
		note); err != nil {
		return fmt.Errorf("append timeline entry: %w", err)
	}
	return nil
}

// ---------------------------------------------------------- attachments

// AddServiceAttachment registers an attachment descriptor. Only the descriptor
// is stored: the bytes are expected to already exist at file_url, which keeps
// binary upload (and its storage, scanning and retention policy) out of scope.
func (r *ServiceRepository) AddServiceAttachment(ctx context.Context, requestID, uploaderID string, req *domain.AddServiceAttachmentRequest) (*domain.ServiceRequestAttachment, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `INSERT INTO service_request_attachments (request_id, file_name,
			file_url, mime_type, size_bytes, uploaded_by)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6)
		RETURNING id`

	var id string
	if err := r.db.QueryRow(ctx, query, requestID, req.FileName, req.FileURL,
		req.MimeType, req.SizeBytes, uploaderID).Scan(&id); err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("add attachment: %w", err)
	}
	return r.GetServiceAttachment(ctx, requestID, id)
}

func (r *ServiceRepository) GetServiceAttachment(ctx context.Context, requestID, id string) (*domain.ServiceRequestAttachment, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT a.id, a.request_id, a.file_name, a.file_url,
			COALESCE(a.mime_type, ''), a.size_bytes, a.uploaded_by, a.created_at
		FROM service_request_attachments a WHERE a.request_id = $1 AND a.id = $2`

	var attachment domain.ServiceRequestAttachment
	if err := r.db.QueryRow(ctx, query, requestID, id).Scan(&attachment.ID,
		&attachment.RequestID, &attachment.FileName, &attachment.FileURL,
		&attachment.MimeType, &attachment.SizeBytes, &attachment.UploadedBy,
		&attachment.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get attachment: %w", err)
	}
	return &attachment, nil
}

func (r *ServiceRepository) ListServiceAttachments(ctx context.Context, requestID string) ([]domain.ServiceRequestAttachment, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT a.id, a.request_id, a.file_name, a.file_url,
			COALESCE(a.mime_type, ''), a.size_bytes, a.uploaded_by, a.created_at
		FROM service_request_attachments a WHERE a.request_id = $1
		ORDER BY a.created_at ASC`

	rows, err := r.db.Query(ctx, query, requestID)
	if err != nil {
		return nil, fmt.Errorf("list attachments: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ServiceRequestAttachment, 0, 8)
	for rows.Next() {
		var attachment domain.ServiceRequestAttachment
		if err := rows.Scan(&attachment.ID, &attachment.RequestID, &attachment.FileName,
			&attachment.FileURL, &attachment.MimeType, &attachment.SizeBytes,
			&attachment.UploadedBy, &attachment.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan attachment: %w", err)
		}
		out = append(out, attachment)
	}
	return out, rows.Err()
}

// RateServiceRequest records the member's satisfaction with how the ticket was
// handled. Only one rating is kept: a repeat submission updates it.
func (r *ServiceRepository) RateServiceRequest(ctx context.Context, requestID, userID string, rating int, comment string) (*domain.ServiceRequest, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `UPDATE service_requests
		SET feedback_rating = $3, feedback_comment = NULLIF($4, ''), updated_at = now()
		WHERE id = $1 AND user_id = $2
		RETURNING id`

	var id string
	if err := r.db.QueryRow(ctx, query, requestID, userID, rating, comment).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		if isCheckViolation(err) {
			return nil, fmt.Errorf("rating must be between 1 and 5")
		}
		return nil, fmt.Errorf("rate request: %w", err)
	}
	return r.GetServiceRequest(ctx, id, userID)
}

// ------------------------------------------------------------- service FAQs

// ListServiceFAQs returns the FAQ entries for one service page.
func (r *ServiceRepository) ListServiceFAQs(ctx context.Context, serviceID string) ([]domain.ServiceFAQ, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT id, resource_id, question, answer, position, created_at
		FROM service_faqs WHERE resource_id = $1 ORDER BY position ASC, created_at ASC`

	rows, err := r.db.Query(ctx, query, serviceID)
	if err != nil {
		return nil, fmt.Errorf("list service faqs: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ServiceFAQ, 0, 8)
	for rows.Next() {
		var faq domain.ServiceFAQ
		if err := rows.Scan(&faq.ID, &faq.ServiceID, &faq.Question, &faq.Answer,
			&faq.Position, &faq.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan service faq: %w", err)
		}
		out = append(out, faq)
	}
	return out, rows.Err()
}

// AddServiceFAQ adds an entry to a service page.
func (r *ServiceRepository) AddServiceFAQ(ctx context.Context, serviceID string, req *domain.CreateServiceFAQRequest) (*domain.ServiceFAQ, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `INSERT INTO service_faqs (resource_id, question, answer, position)
		VALUES ($1, $2, $3, $4)
		RETURNING id, resource_id, question, answer, position, created_at`

	var faq domain.ServiceFAQ
	if err := r.db.QueryRow(ctx, query, serviceID, req.Question, req.Answer, req.Position).
		Scan(&faq.ID, &faq.ServiceID, &faq.Question, &faq.Answer, &faq.Position,
			&faq.CreatedAt); err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrInvalidReference
		}
		return nil, fmt.Errorf("add service faq: %w", err)
	}
	return &faq, nil
}

func (r *ServiceRepository) DeleteServiceFAQ(ctx context.Context, serviceID, faqID string) error {
	if err := r.ready(); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx,
		`DELETE FROM service_faqs WHERE id = $1 AND resource_id = $2`, faqID, serviceID)
	if err != nil {
		return fmt.Errorf("delete service faq: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// -------------------------------------------------------------- queue stats

// QueueStats is the service desk dashboard: what is outstanding, what is
// urgent, and how quickly tickets are actually being closed.
func (r *ServiceRepository) QueueStats(ctx context.Context) (*domain.ServiceQueueStats, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	const query = `SELECT
			count(*) FILTER (WHERE status = 'OPEN'),
			count(*) FILTER (WHERE status = 'IN_PROGRESS'),
			count(*) FILTER (WHERE status = 'RESOLVED'),
			count(*) FILTER (WHERE status = 'CLOSED'),
			count(*) FILTER (WHERE status = 'REJECTED'),
			count(*) FILTER (WHERE kind = 'COMPLAINT'),
			count(*) FILTER (WHERE priority IN ('HIGH', 'URGENT')
				AND status NOT IN ('RESOLVED', 'CLOSED', 'REJECTED')),
			COALESCE(AVG(EXTRACT(EPOCH FROM (resolved_at - created_at)) / 3600)
				FILTER (WHERE resolved_at IS NOT NULL), 0)
		FROM service_requests`

	stats := &domain.ServiceQueueStats{}
	if err := r.db.QueryRow(ctx, query).Scan(&stats.Open, &stats.InProgress,
		&stats.Resolved, &stats.Closed, &stats.Rejected, &stats.Complaints,
		&stats.UrgentOpen, &stats.AvgResolutionHours); err != nil {
		return nil, fmt.Errorf("service queue stats: %w", err)
	}
	stats.AvgResolutionHours = round2(stats.AvgResolutionHours)
	return stats, nil
}

// requestReferenceAlphabet excludes letters and digits that are easy to
// misread when a member reads a code across a counter.
const requestReferenceAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// newRequestReference builds a short, human-quotable ticket code. Collisions are
// astronomically unlikely at this length and the column is unique, so a clash
// surfaces as an error rather than silently merging two tickets.
func newRequestReference() (string, error) {
	const length = 8

	limit := big.NewInt(int64(len(requestReferenceAlphabet)))
	out := make([]byte, length)
	for i := range out {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("generate request reference: %w", err)
		}
		out[i] = requestReferenceAlphabet[n.Int64()]
	}
	return "SR-" + string(out), nil
}

// serviceSlaHours is how long a ticket may sit before it counts as overdue in
// the queue view. It is a reporting threshold, not a promise to the member.
const serviceSlaHours = 72

// OverdueSince reports when a ticket passed the reporting threshold, which the
// queue sorts on so stale tickets surface first.
func (r *ServiceRepository) OverdueSince(ctx context.Context, id string) (time.Time, error) {
	if err := r.ready(); err != nil {
		return time.Time{}, err
	}
	const query = `SELECT created_at + ($2 || ' hours')::interval
		FROM service_requests WHERE id = $1`

	var due time.Time
	if err := r.db.QueryRow(ctx, query, id, serviceSlaHours).Scan(&due); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, ErrNotFound
		}
		return time.Time{}, fmt.Errorf("overdue threshold: %w", err)
	}
	return due, nil
}
