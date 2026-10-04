package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/delivery/http/middleware"
	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/internal/repository/postgres"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

// ServiceHandler is the campus service desk. A member reads a service page to
// find out how something works, then files a request or a complaint and follows it
// to resolution. Staff work the queue on the other side of the same tickets.
type ServiceHandler struct {
	db            *postgres.ServiceRepository
	notifications *postgres.NotificationRepository
}

func NewServiceHandler(db *postgres.ServiceRepository, notifications *postgres.NotificationRepository) *ServiceHandler {
	return &ServiceHandler{db: db, notifications: notifications}
}

// CreateServiceRequest opens a ticket against a campus service.
func (h *ServiceHandler) CreateServiceRequest(c *gin.Context) {
	var req domain.CreateServiceRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	request, err := h.db.CreateServiceRequest(c.Request.Context(), middleware.UserID(c), &req)
	if err != nil {
		writeRepoError(c, err, "SERVICE_NOT_FOUND", "SERVICE_REQUEST_CREATE_FAILED")
		return
	}

	// The reference goes back in the response body because it is what the member
	// will quote at a counter or over the phone.
	response.Success(c, http.StatusCreated, "Request submitted", request)
}

// ListMyServiceRequests returns the signed-in member's own tickets.
func (h *ServiceHandler) ListMyServiceRequests(c *gin.Context) {
	requests, err := h.db.ListServiceRequests(c.Request.Context(), middleware.UserID(c),
		c.Query("status"), c.Query("kind"), pageSize(c.Query("limit")),
		parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "SERVICE_REQUEST_NOT_FOUND", "SERVICE_REQUEST_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Your requests retrieved", gin.H{
		"requests": requests,
		"count":    len(requests),
	})
}

// GetServiceRequest returns one ticket with its timeline. A member may only read
// their own ticket, which the repository enforces.
func (h *ServiceHandler) GetServiceRequest(c *gin.Context) {
	request, err := h.db.GetServiceRequest(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "SERVICE_REQUEST_NOT_FOUND", "SERVICE_REQUEST_FETCH_FAILED")
		return
	}

	timeline, err := h.db.ListServiceRequestTimeline(c.Request.Context(), request.ID)
	if err != nil {
		writeRepoError(c, err, "SERVICE_REQUEST_NOT_FOUND", "SERVICE_TIMELINE_FAILED")
		return
	}

	attachments, err := h.db.ListServiceAttachments(c.Request.Context(), request.ID)
	if err != nil {
		writeRepoError(c, err, "SERVICE_REQUEST_NOT_FOUND", "SERVICE_ATTACHMENT_LIST_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Request retrieved", gin.H{
		"request":     request,
		"timeline":    timeline,
		"attachments": attachments,
	})
}

// AddServiceRequestMessage lets a member add context to their own ticket without
// changing its status.
func (h *ServiceHandler) AddServiceRequestMessage(c *gin.Context) {
	request, err := h.db.GetServiceRequest(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "SERVICE_REQUEST_NOT_FOUND", "SERVICE_REQUEST_FETCH_FAILED")
		return
	}

	var req domain.AddServiceRequestMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	event, err := h.db.AddServiceRequestMessage(c.Request.Context(), request.ID,
		middleware.UserID(c), req.Message)
	if err != nil {
		writeRepoError(c, err, "SERVICE_REQUEST_NOT_FOUND", "SERVICE_MESSAGE_FAILED")
		return
	}
	response.Success(c, http.StatusCreated, "Message added", event)
}

// RateServiceRequest records the member's satisfaction with how their ticket was
// handled.
func (h *ServiceHandler) RateServiceRequest(c *gin.Context) {
	var req domain.ServiceRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	request, err := h.db.RateServiceRequest(c.Request.Context(), c.Param("id"),
		middleware.UserID(c), req.Rating, req.Comment)
	if err != nil {
		writeRepoError(c, err, "SERVICE_REQUEST_NOT_FOUND", "SERVICE_RATING_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Thanks for the feedback", request)
}

// AddServiceAttachment registers an attachment descriptor against the member's
// own ticket. The file itself is expected to exist already at the given URL.
func (h *ServiceHandler) AddServiceAttachment(c *gin.Context) {
	request, err := h.db.GetServiceRequest(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "SERVICE_REQUEST_NOT_FOUND", "SERVICE_REQUEST_FETCH_FAILED")
		return
	}

	var req domain.AddServiceAttachmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	attachment, err := h.db.AddServiceAttachment(c.Request.Context(), request.ID,
		middleware.UserID(c), &req)
	if err != nil {
		writeRepoError(c, err, "SERVICE_REQUEST_NOT_FOUND", "SERVICE_ATTACHMENT_FAILED")
		return
	}
	response.Success(c, http.StatusCreated, "Attachment recorded", attachment)
}

// ----------------------------------------------------------------- staff

// ListServiceQueue is the staff view: every ticket, most urgent first.
func (h *ServiceHandler) ListServiceQueue(c *gin.Context) {
	requests, err := h.db.ListServiceRequests(c.Request.Context(), "",
		c.Query("status"), c.Query("kind"), pageSize(c.Query("limit")),
		parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "SERVICE_REQUEST_NOT_FOUND", "SERVICE_QUEUE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Service queue retrieved", gin.H{
		"requests": requests,
		"count":    len(requests),
	})
}

// UpdateServiceRequest moves a ticket along and records the change on its
// timeline. Resolving or rejecting requires a resolution so a closed ticket
// always says what was done.
func (h *ServiceHandler) UpdateServiceRequest(c *gin.Context) {
	var req domain.UpdateServiceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	userID := middleware.UserID(c)
	request, err := h.db.UpdateServiceRequest(c.Request.Context(), c.Param("id"), userID,
		req.Status, req.Note, req.Resolution, req.AssignedTo)
	if err != nil {
		writeRepoError(c, err, "SERVICE_REQUEST_NOT_FOUND", "SERVICE_REQUEST_UPDATE_FAILED")
		return
	}

	// The member is told the outcome. A closed ticket is the one they most need
	// to hear about, since they are the one waiting.
	_ = h.notifications.NotifyUser(c.Request.Context(), request.UserID,
		domain.NotificationTypeService,
		"Your request "+request.Reference+" is now "+req.Status,
		resolutionOrStatus(request), "/support/requests/"+request.ID)

	response.Success(c, http.StatusOK, "Request updated", request)
}

// QueueStats is the service desk dashboard.
func (h *ServiceHandler) QueueStats(c *gin.Context) {
	stats, err := h.db.QueueStats(c.Request.Context())
	if err != nil {
		writeRepoError(c, err, "SERVICE_QUEUE_NOT_FOUND", "SERVICE_QUEUE_STATS_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Service queue statistics retrieved", stats)
}

// ------------------------------------------------------------ service pages

// ListServiceFAQs returns the questions and answers attached to one service page.
func (h *ServiceHandler) ListServiceFAQs(c *gin.Context) {
	faqs, err := h.db.ListServiceFAQs(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeRepoError(c, err, "SERVICE_NOT_FOUND", "SERVICE_FAQ_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Service FAQs retrieved", gin.H{
		"faqs":  faqs,
		"count": len(faqs),
	})
}

// AddServiceFAQ adds a question to a service page.
func (h *ServiceHandler) AddServiceFAQ(c *gin.Context) {
	var req domain.CreateServiceFAQRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	faq, err := h.db.AddServiceFAQ(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		writeRepoError(c, err, "SERVICE_NOT_FOUND", "SERVICE_FAQ_CREATE_FAILED")
		return
	}
	response.Success(c, http.StatusCreated, "FAQ entry added", faq)
}

func (h *ServiceHandler) DeleteServiceFAQ(c *gin.Context) {
	if err := h.db.DeleteServiceFAQ(c.Request.Context(), c.Param("id"), c.Param("faqId")); err != nil {
		writeRepoError(c, err, "FAQ_NOT_FOUND", "SERVICE_FAQ_DELETE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "FAQ entry removed", nil)
}

// resolutionOrStatus builds the message body for an outcome notification: the
// resolution when staff wrote one, and the status on its own when they did not.
func resolutionOrStatus(request *domain.ServiceRequest) string {
	if request.Resolution != "" {
		return request.Resolution
	}
	return "Current status: " + string(request.Status)
}
