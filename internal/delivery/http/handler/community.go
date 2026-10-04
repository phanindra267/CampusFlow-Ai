package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/campuscare/api/internal/delivery/http/middleware"
	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/internal/repository/postgres"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

// CommunityHandler serves the collaboration surfaces of the community platform:
// discussion threads, member notifications, opportunity applications, campus
// services and their bookings, and cross-entity search.
type CommunityHandler struct {
	db *postgres.CommunityRepository
}

func NewCommunityHandler(db *postgres.CommunityRepository) *CommunityHandler {
	return &CommunityHandler{db: db}
}

// defaultPageSize and maxPageSize bound list endpoints so a client cannot ask
// for an unbounded result set.
const (
	defaultPageSize = 25
	maxPageSize     = 100
)

func pageSize(raw string) int {
	return min(parsePositiveInt(raw, defaultPageSize), maxPageSize)
}

// writeRepoError maps repository sentinels onto HTTP responses.
func writeRepoError(c *gin.Context, err error, notFoundCode, failureCode string) {
	switch {
	case errors.Is(err, postgres.ErrNotFound):
		response.Error(c, http.StatusNotFound, notFoundCode, err)
	case errors.Is(err, postgres.ErrDuplicateKey):
		response.Error(c, http.StatusConflict, "already exists", err)
	case errors.Is(err, postgres.ErrInvalidReference):
		response.Error(c, http.StatusBadRequest, "referenced resource does not exist", err)
	case errors.Is(err, postgres.ErrSlotUnavailable):
		response.Error(c, http.StatusConflict, "that slot is already booked", err)
	case errors.Is(err, postgres.ErrEventFull):
		response.Error(c, http.StatusConflict, "this event is full", err)
	case errors.Is(err, postgres.ErrNoDatabase):
		response.Error(c, http.StatusServiceUnavailable, "database unavailable", err)
	default:
		response.Error(c, http.StatusInternalServerError, failureCode, err)
	}
}

// ------------------------------------------------------------ discussions

// ListDiscussions returns recent threads. The optional club_id query parameter
// scopes the list to one group; omitting it returns campus-wide threads.
func (h *CommunityHandler) ListDiscussions(c *gin.Context) {
	var clubID *string
	if raw := c.Query("club_id"); raw != "" {
		clubID = &raw
	}

	limit := pageSize(c.Query("limit"))
	offset := c.Query("offset")

	threads, err := h.db.ListDiscussions(c.Request.Context(), clubID, limit, parsePositiveInt(offset, 0))
	if err != nil {
		writeRepoError(c, err, "DISCUSSION_NOT_FOUND", "DISCUSSION_LIST_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Discussions retrieved", gin.H{
		"discussions": threads,
		"count":       len(threads),
	})
}

func (h *CommunityHandler) GetDiscussion(c *gin.Context) {
	thread, replies, err := h.db.GetDiscussion(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeRepoError(c, err, "DISCUSSION_NOT_FOUND", "DISCUSSION_FETCH_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Discussion retrieved", gin.H{
		"discussion": thread,
		"replies":    replies,
	})
}

type createDiscussionRequest struct {
	ClubID   *string `json:"club_id"`
	Title    string  `json:"title" binding:"required,min=3,max=255"`
	Body     string  `json:"body" binding:"required,min=1"`
	Category string  `json:"category"`
}

func (h *CommunityHandler) CreateDiscussion(c *gin.Context) {
	var req createDiscussionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	category := domain.DiscussionCategory(req.Category)
	switch category {
	case domain.DiscussionCategoryGeneral, domain.DiscussionCategoryAnnouncement,
		domain.DiscussionCategoryQuestion, domain.DiscussionCategoryEvent,
		domain.DiscussionCategoryOpportunity, domain.DiscussionCategoryResearch:
	case "":
		category = domain.DiscussionCategoryGeneral
	default:
		response.Error(c, http.StatusBadRequest, "INVALID_CATEGORY", nil)
		return
	}

	thread := &domain.Discussion{
		ClubID:   req.ClubID,
		AuthorID: middleware.UserID(c),
		Title:    req.Title,
		Body:     req.Body,
		Category: category,
	}

	if err := h.db.CreateDiscussion(c.Request.Context(), thread); err != nil {
		writeRepoError(c, err, "DISCUSSION_NOT_FOUND", "DISCUSSION_CREATE_FAILED")
		return
	}

	response.Success(c, http.StatusCreated, "Discussion created", thread)
}

type createReplyRequest struct {
	Body     string  `json:"body" binding:"required,min=1"`
	ParentID *string `json:"parent_id"`
}

func (h *CommunityHandler) AddReply(c *gin.Context) {
	var req createReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	reply := &domain.DiscussionReply{
		DiscussionID: c.Param("id"),
		AuthorID:     middleware.UserID(c),
		ParentID:     req.ParentID,
		Body:         req.Body,
	}

	if err := h.db.AddReply(c.Request.Context(), reply); err != nil {
		writeRepoError(c, err, "DISCUSSION_NOT_FOUND", "REPLY_CREATE_FAILED")
		return
	}

	response.Success(c, http.StatusCreated, "Reply posted", reply)
}

// ----------------------------------------------------------- notifications

func (h *CommunityHandler) ListNotifications(c *gin.Context) {
	unreadOnly := c.Query("unread") == "true"

	items, unread, err := h.db.ListNotifications(
		c.Request.Context(), middleware.UserID(c), unreadOnly,
		pageSize(c.Query("limit")), parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "NOTIFICATION_NOT_FOUND", "NOTIFICATION_LIST_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Notifications retrieved", gin.H{
		"notifications": items,
		"count":         len(items),
		"unread":        unread,
	})
}

func (h *CommunityHandler) MarkNotificationRead(c *gin.Context) {
	err := h.db.MarkNotificationRead(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		writeRepoError(c, err, "NOTIFICATION_NOT_FOUND", "NOTIFICATION_UPDATE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Notification marked read", nil)
}

func (h *CommunityHandler) MarkAllNotificationsRead(c *gin.Context) {
	updated, err := h.db.MarkAllNotificationsRead(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "NOTIFICATION_NOT_FOUND", "NOTIFICATION_UPDATE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Notifications marked read", gin.H{"updated": updated})
}

// ----------------------------------------------------------- applications

func (h *CommunityHandler) ListMyApplications(c *gin.Context) {
	applications, err := h.db.ListMyApplications(c.Request.Context(), middleware.UserID(c),
		pageSize(c.Query("limit")), parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "APPLICATION_NOT_FOUND", "APPLICATION_LIST_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Applications retrieved", gin.H{
		"applications": applications,
		"count":        len(applications),
	})
}

// -------------------------------------------------------------- services

func (h *CommunityHandler) ListResources(c *gin.Context) {
	resources, err := h.db.ListResources(c.Request.Context(), c.Query("type"), c.Query("q"),
		pageSize(c.Query("limit")), parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "RESOURCE_NOT_FOUND", "RESOURCE_LIST_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Campus services retrieved", gin.H{
		"resources": resources,
		"count":     len(resources),
	})
}

type bookResourceRequest struct {
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
	Note      string    `json:"note"`
}

func (h *CommunityHandler) BookResource(c *gin.Context) {
	var req bookResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	if !req.EndTime.After(req.StartTime) {
		response.Error(c, http.StatusBadRequest, "END_TIME_MUST_BE_AFTER_START_TIME", nil)
		return
	}
	if req.StartTime.Before(time.Now()) {
		response.Error(c, http.StatusBadRequest, "CANNOT_BOOK_PAST_SLOT", nil)
		return
	}

	booking := &domain.ResourceBooking{
		ResourceID: c.Param("id"),
		UserID:     middleware.UserID(c),
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Note:       req.Note,
	}

	if err := h.db.BookResource(c.Request.Context(), booking); err != nil {
		writeRepoError(c, err, "RESOURCE_NOT_FOUND", "BOOKING_FAILED")
		return
	}

	response.Success(c, http.StatusCreated, "Resource booked", booking)
}

func (h *CommunityHandler) ListMyBookings(c *gin.Context) {
	bookings, err := h.db.ListMyBookings(c.Request.Context(), middleware.UserID(c),
		c.Query("upcoming") == "true")
	if err != nil {
		writeRepoError(c, err, "BOOKING_NOT_FOUND", "BOOKING_LIST_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Bookings retrieved", gin.H{
		"bookings": bookings,
		"count":    len(bookings),
	})
}

func (h *CommunityHandler) CancelBooking(c *gin.Context) {
	err := h.db.CancelBooking(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		writeRepoError(c, err, "BOOKING_NOT_FOUND", "BOOKING_CANCEL_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Booking cancelled", nil)
}

// ---------------------------------------------------------------- search

func (h *CommunityHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "QUERY_REQUIRED", nil)
		return
	}

	results, err := h.db.Search(c.Request.Context(), query, c.Query("type"),
		pageSize(c.Query("limit")))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "SEARCH_FAILED", err)
		return
	}

	response.Success(c, http.StatusOK, "Search completed", results)
}
