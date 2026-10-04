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

// CampusHandler serves the community surfaces members use most: groups and their
// memberships, opportunity listings, event registration, waitlists, event
// check-in and feedback.
type CampusHandler struct {
	db *postgres.CampusRepository
}

func NewCampusHandler(db *postgres.CampusRepository) *CampusHandler {
	return &CampusHandler{db: db}
}

// ----------------------------------------------------------------- clubs

func (h *CampusHandler) GetClub(c *gin.Context) {
	club, memberCount, err := h.db.GetClub(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_FETCH_FAILED")
		return
	}

	payload := gin.H{"club": club, "member_count": memberCount}

	members, err := h.db.ListClubMembers(c.Request.Context(), club.ID)
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_MEMBERS_FAILED")
		return
	}
	payload["members"] = members

	response.Success(c, http.StatusOK, "Club retrieved", payload)
}

func (h *CampusHandler) JoinClub(c *gin.Context) {
	membership, err := h.db.JoinClub(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_JOIN_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Joined club", membership)
}

func (h *CampusHandler) LeaveClub(c *gin.Context) {
	err := h.db.LeaveClub(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "MEMBERSHIP_NOT_FOUND", "CLUB_LEAVE_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Left club", nil)
}

// --------------------------------------------------------- opportunities

func (h *CampusHandler) ListOpportunities(c *gin.Context) {
	opportunities, err := h.db.ListOpportunities(c.Request.Context(), c.Query("category"),
		c.Query("q"), c.Query("saved") == "true", middleware.UserID(c),
		pageSize(c.Query("limit")), parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_LIST_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Opportunities retrieved", gin.H{
		"opportunities": opportunities,
		"count":         len(opportunities),
	})
}

func (h *CampusHandler) GetOpportunity(c *gin.Context) {
	opportunity, applications, err := h.db.GetOpportunity(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_FETCH_FAILED")
		return
	}

	average, total, err := h.db.AverageFeedback(c.Request.Context(), "OPPORTUNITY", opportunity.ID)
	if err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "FEEDBACK_FETCH_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Opportunity retrieved", gin.H{
		"opportunity":       opportunity,
		"application_count": applications,
		"average_rating":    average,
		"feedback_count":    total,
	})
}

func (h *CampusHandler) SaveOpportunity(c *gin.Context) {
	err := h.db.SaveOpportunity(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_SAVE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Opportunity saved", nil)
}

func (h *CampusHandler) UnsaveOpportunity(c *gin.Context) {
	err := h.db.UnsaveOpportunity(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_UNSAVE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Opportunity removed from saved", nil)
}

// ---------------------------------------------------------------- events

func (h *CampusHandler) ListEvents(c *gin.Context) {
	events, err := h.db.ListEvents(c.Request.Context(), c.Query("category"), c.Query("q"),
		middleware.UserID(c), pageSize(c.Query("limit")), parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_LIST_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Events retrieved", gin.H{
		"events": events,
		"count":  len(events),
	})
}

func (h *CampusHandler) GetEvent(c *gin.Context) {
	event, err := h.db.GetEvent(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_FETCH_FAILED")
		return
	}

	average, total, err := h.db.AverageFeedback(c.Request.Context(), "EVENT", event.ID)
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "FEEDBACK_FETCH_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Event retrieved", gin.H{
		"event":          event,
		"average_rating": average,
		"feedback_count": total,
	})
}

// RegisterForEvent books a place, falling back to the waitlist when the event is
// full so a member always gets a useful outcome instead of an error page.
func (h *CampusHandler) RegisterForEvent(c *gin.Context) {
	userID := middleware.UserID(c)
	registration, _, err := h.db.RegisterForEvent(c.Request.Context(), c.Param("id"), userID)
	if err == nil {
		response.Success(c, http.StatusCreated, "Registered for event", registration)
		return
	}

	if !errors.Is(err, postgres.ErrEventFull) {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "REGISTRATION_FAILED")
		return
	}

	position, waitErr := h.db.JoinWaitlist(c.Request.Context(), &domain.WaitlistEntry{
		EventID:  c.Param("id"),
		MemberID: userID,
	})
	if waitErr != nil {
		writeRepoError(c, waitErr, "EVENT_NOT_FOUND", "WAITLIST_JOIN_FAILED")
		return
	}

	response.Success(c, http.StatusAccepted, "Event is full, you joined the waitlist", gin.H{
		"waitlist_position": position,
		"event_id":          c.Param("id"),
	})
}

func (h *CampusHandler) CancelRegistration(c *gin.Context) {
	err := h.db.CancelRegistration(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "REGISTRATION_NOT_FOUND", "REGISTRATION_CANCEL_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Registration cancelled", nil)
}

func (h *CampusHandler) ListWaitlist(c *gin.Context) {
	entries, err := h.db.ListWaitlist(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "WAITLIST_FETCH_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Waitlist retrieved", gin.H{
		"waitlist": entries,
		"count":    len(entries),
	})
}

// --------------------------------------------------------------- check-in

type openCheckInRequest struct {
	DurationMinutes int `json:"duration_minutes"`
}

// maxCheckInMinutes bounds how long a check-in window stays open.
const maxCheckInMinutes = 120

func (h *CampusHandler) OpenCheckIn(c *gin.Context) {
	var req openCheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	if req.DurationMinutes <= 0 {
		req.DurationMinutes = 15
	}
	if req.DurationMinutes > maxCheckInMinutes {
		response.Error(c, http.StatusBadRequest, "DURATION_TOO_LONG", nil)
		return
	}

	session := &domain.CheckInSession{
		EventSessionID: c.Param("id"),
		OrganizerID:    middleware.UserID(c),
		ExpiresAt:      time.Now().Add(time.Duration(req.DurationMinutes) * time.Minute),
	}

	if err := h.db.OpenCheckIn(c.Request.Context(), session); err != nil {
		writeRepoError(c, err, "SESSION_NOT_FOUND", "CHECK_IN_OPEN_FAILED")
		return
	}

	response.Success(c, http.StatusCreated, "Check-in window opened", session)
}

type checkInRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *CampusHandler) CheckIn(c *gin.Context) {
	var req checkInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	record, err := h.db.CheckInByToken(c.Request.Context(), req.Token, middleware.UserID(c), "QR")
	if err != nil {
		if errors.Is(err, postgres.ErrCheckInRejected) {
			response.Error(c, http.StatusConflict,
				"check-in rejected: the code is invalid, expired, or already used", err)
			return
		}
		writeRepoError(c, err, "SESSION_NOT_FOUND", "CHECK_IN_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Checked in", record)
}

func (h *CampusHandler) ListAttendance(c *gin.Context) {
	records, err := h.db.ListEventAttendance(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeRepoError(c, err, "SESSION_NOT_FOUND", "ATTENDANCE_FETCH_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Attendance retrieved", gin.H{
		"attendance": records,
		"count":      len(records),
	})
}

// -------------------------------------------------------------- feedback

type submitFeedbackRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

func (h *CampusHandler) SubmitFeedback(c *gin.Context) {
	var req submitFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	activityType := c.Param("type")
	if activityType != "EVENT" && activityType != "OPPORTUNITY" {
		response.Error(c, http.StatusBadRequest, "UNSUPPORTED_ACTIVITY_TYPE", nil)
		return
	}

	feedback := &domain.Feedback{
		ActivityType: activityType,
		ActivityID:   c.Param("id"),
		MemberID:     middleware.UserID(c),
		Rating:       req.Rating,
		Comment:      req.Comment,
	}

	if err := h.db.SubmitFeedback(c.Request.Context(), feedback); err != nil {
		writeRepoError(c, err, "ACTIVITY_NOT_FOUND", "FEEDBACK_SUBMIT_FAILED")
		return
	}

	response.Success(c, http.StatusCreated, "Feedback recorded", feedback)
}

// -------------------------------------------------------------- activity

func (h *CampusHandler) MyActivity(c *gin.Context) {
	limit := min(parsePositiveInt(c.Query("limit"), defaultPageSize), maxPageSize)

	activity, err := h.db.RecentActivity(c.Request.Context(), middleware.UserID(c), limit)
	if err != nil {
		writeRepoError(c, err, "ACTIVITY_NOT_FOUND", "ACTIVITY_FETCH_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Your activity retrieved", gin.H{
		"activity": activity,
		"count":    len(activity),
	})
}
