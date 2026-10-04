package handler

import (
	"net/http"
	"strings"

	"github.com/campuscare/api/internal/delivery/http/middleware"
	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/internal/repository/postgres"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

// ContentHandler serves the surfaces that turn CampusCare AI from a read-only
// catalogue into something a campus runs itself: organisers create events, clubs
// and opportunities; members save, follow and apply; announcements carry
// campus-wide notices.
//
// Ownership is enforced in the handler rather than the repository: an organiser
// may only change content they created, while an administrator may change any of
// it. Centralising that decision here keeps one rule for every write path.
type ContentHandler struct {
	events        *postgres.EventRepository
	clubs         *postgres.ClubRepository
	opportunities *postgres.OpportunityRepository
	announcements *postgres.AnnouncementRepository
	notifications *postgres.NotificationRepository
}

func NewContentHandler(events *postgres.EventRepository, clubs *postgres.ClubRepository,
	opportunities *postgres.OpportunityRepository, announcements *postgres.AnnouncementRepository,
	notifications *postgres.NotificationRepository) *ContentHandler {
	return &ContentHandler{
		events:        events,
		clubs:         clubs,
		opportunities: opportunities,
		announcements: announcements,
		notifications: notifications,
	}
}

// adminRoles are the roles that may act on any content, not just their own.
var adminRoles = []string{"ADMIN", "SUPER_ADMIN"}

func isAdmin(c *gin.Context) bool {
	role := middleware.Role(c)
	for _, candidate := range adminRoles {
		if role == candidate {
			return true
		}
	}
	return false
}

// ownedByCaller reports whether the caller may modify content created by owner.
// Admins always may; anyone else only for their own content.
func ownedByCaller(c *gin.Context, owner string) bool {
	return isAdmin(c) || owner == middleware.UserID(c)
}

// ------------------------------------------------------------------ events

// CreateEvent is the organiser write path. An event may be drafted or published
// in the same call, so an organiser can prepare something and check it before it
// appears in the calendar.
func (h *ContentHandler) CreateEvent(c *gin.Context) {
	var req domain.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	event, err := h.events.CreateEvent(c.Request.Context(), &req, middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_CREATE_FAILED")
		return
	}

	response.Success(c, http.StatusCreated, "Event created", event)
}

// UpdateEvent edits an event the caller organises.
func (h *ContentHandler) UpdateEvent(c *gin.Context) {
	var req domain.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	event, err := h.events.GetEvent(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_FETCH_FAILED")
		return
	}
	if !ownedByCaller(c, event.OrganizerID) {
		response.Error(c, http.StatusForbidden, "only the organiser who posted this event can change it", nil)
		return
	}

	updated, err := h.events.UpdateEvent(c.Request.Context(), event.ID, &req)
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_UPDATE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Event updated", updated)
}

// CancelEvent withdraws an event and tells everyone who had a place, because a
// cancelled event is exactly the message a member needs to receive.
func (h *ContentHandler) CancelEvent(c *gin.Context) {
	event, err := h.events.GetEvent(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_FETCH_FAILED")
		return
	}
	if !ownedByCaller(c, event.OrganizerID) {
		response.Error(c, http.StatusForbidden, "only the organiser who posted this event can cancel it", nil)
		return
	}

	if err := h.events.CancelEvent(c.Request.Context(), event.ID); err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_CANCEL_FAILED")
		return
	}

	notified, _ := h.notifications.NotifyEventAttendees(c.Request.Context(), event.ID,
		"Cancelled: "+event.Title,
		"This event is no longer taking place. Your registration has been released.",
		"/events/"+event.ID)

	response.Success(c, http.StatusOK, "Event cancelled", gin.H{
		"event_id":           event.ID,
		"attendees_notified": notified,
	})
}

// ListOrganiserEvents is the organiser's own dashboard, drafts included.
func (h *ContentHandler) ListOrganiserEvents(c *gin.Context) {
	events, err := h.events.ListEventsByOrganizer(c.Request.Context(), middleware.UserID(c),
		pageSize(c.Query("limit")), parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Your events retrieved", gin.H{
		"events": events,
		"count":  len(events),
	})
}

// ListEventRegistrations is the organiser's attendee list.
func (h *ContentHandler) ListEventRegistrations(c *gin.Context) {
	event, err := h.events.GetEvent(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_FETCH_FAILED")
		return
	}
	if !ownedByCaller(c, event.OrganizerID) {
		response.Error(c, http.StatusForbidden, "only the organiser who posted this event can see its registrations", nil)
		return
	}

	registrations, err := h.events.ListEventRegistrations(c.Request.Context(), event.ID,
		pageSize(c.Query("limit")), parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "REGISTRATION_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Registrations retrieved", gin.H{
		"registrations": registrations,
		"count":         len(registrations),
	})
}

// EventAnalytics is the evidence an organiser needs after running an event.
func (h *ContentHandler) EventAnalytics(c *gin.Context) {
	event, err := h.events.GetEvent(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_FETCH_FAILED")
		return
	}
	if !ownedByCaller(c, event.OrganizerID) {
		response.Error(c, http.StatusForbidden, "only the organiser who posted this event can see its analytics", nil)
		return
	}

	stats, err := h.events.OrganizerEventAnalytics(c.Request.Context(), event.ID)
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_ANALYTICS_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Event analytics retrieved", stats)
}

// PromoteWaitlist fills a place that just freed up from the waitlist.
func (h *ContentHandler) PromoteWaitlist(c *gin.Context) {
	event, err := h.events.GetEvent(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_FETCH_FAILED")
		return
	}
	if !ownedByCaller(c, event.OrganizerID) {
		response.Error(c, http.StatusForbidden, "only the organiser who posted this event can promote from the waitlist", nil)
		return
	}

	registration, err := h.events.PromoteWaitlist(c.Request.Context(), event.ID)
	if err != nil {
		writeRepoError(c, err, "WAITLIST_EMPTY", "WAITLIST_PROMOTE_FAILED")
		return
	}

	// The promoted member is told directly, because the offer only stands until
	// they accept it.
	_ = h.notifications.NotifyUser(c.Request.Context(), registration.MemberID,
		domain.NotificationTypeEvent, "A place opened up: "+event.Title,
		"You have been moved off the waitlist into a confirmed place.",
		"/events/"+event.ID)

	response.Success(c, http.StatusCreated, "Waitlist promoted", registration)
}

// SaveEvent bookmarks an event. Saving is deliberately separate from
// registering: a member can shortlist a full event.
func (h *ContentHandler) SaveEvent(c *gin.Context) {
	if err := h.events.SaveEvent(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_SAVE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Event saved", nil)
}

func (h *ContentHandler) UnsaveEvent(c *gin.Context) {
	if err := h.events.UnsaveEvent(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_UNSAVE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Event removed from saved", nil)
}

// ListSavedEvents is the member's shortlist.
func (h *ContentHandler) ListSavedEvents(c *gin.Context) {
	events, err := h.events.ListSavedEvents(c.Request.Context(), middleware.UserID(c),
		pageSize(c.Query("limit")), parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "EVENT_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Saved events retrieved", gin.H{
		"events": events,
		"count":  len(events),
	})
}

// SetEventReminder schedules the member's reminder for an event they saved or
// registered for.
func (h *ContentHandler) SetEventReminder(c *gin.Context) {
	var req domain.SetReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	if err := h.events.SetEventReminder(c.Request.Context(), middleware.UserID(c),
		c.Param("id"), req.RemindAt); err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "REMINDER_SET_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Reminder set", gin.H{
		"event_id":  c.Param("id"),
		"remind_at": req.RemindAt,
	})
}

func (h *ContentHandler) ClearEventReminder(c *gin.Context) {
	if err := h.events.ClearEventReminder(c.Request.Context(), middleware.UserID(c),
		c.Param("id")); err != nil {
		writeRepoError(c, err, "EVENT_NOT_FOUND", "REMINDER_CLEAR_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Reminder cleared", nil)
}

// EventCategories gives the browse filter the categories that actually exist.
func (h *ContentHandler) EventCategories(c *gin.Context) {
	categories, err := h.events.ListEventCategories(c.Request.Context())
	if err != nil {
		writeRepoError(c, err, "CATEGORY_NOT_FOUND", "CATEGORY_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Event categories retrieved", gin.H{
		"categories": categories,
	})
}

// ------------------------------------------------------------------- clubs

// CreateClub registers a club. New clubs start PENDING for verification: a club
// page that anyone can create under any name is not something a campus wants.
func (h *ContentHandler) CreateClub(c *gin.Context) {
	var req domain.CreateClubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	club, err := h.clubs.CreateClub(c.Request.Context(), &req, middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_CREATE_FAILED")
		return
	}
	response.Success(c, http.StatusCreated, "Club created", club)
}

func (h *ContentHandler) UpdateClub(c *gin.Context) {
	var req domain.CreateClubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	existing, err := h.clubs.GetClub(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_FETCH_FAILED")
		return
	}
	if !ownedByCaller(c, existing.CreatedBy) {
		response.Error(c, http.StatusForbidden, "only the founder or an administrator can change this club", nil)
		return
	}

	club, err := h.clubs.UpdateClub(c.Request.Context(), existing.ID, &req)
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_UPDATE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Club updated", club)
}

// ArchiveClub retires a club, keeping its history.
func (h *ContentHandler) ArchiveClub(c *gin.Context) {
	existing, err := h.clubs.GetClub(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_FETCH_FAILED")
		return
	}
	if !ownedByCaller(c, existing.CreatedBy) {
		response.Error(c, http.StatusForbidden, "only the founder or an administrator can archive this club", nil)
		return
	}

	if err := h.clubs.ArchiveClub(c.Request.Context(), existing.ID); err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_ARCHIVE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Club archived", nil)
}

// FollowClub records interest without joining.
func (h *ContentHandler) FollowClub(c *gin.Context) {
	if err := h.clubs.FollowClub(c.Request.Context(), c.Param("id"), middleware.UserID(c)); err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_FOLLOW_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Following club", nil)
}

func (h *ContentHandler) UnfollowClub(c *gin.Context) {
	if err := h.clubs.UnfollowClub(c.Request.Context(), c.Param("id"), middleware.UserID(c)); err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_UNFOLLOW_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "No longer following club", nil)
}

// ListFollowedClubs is the member's follow list.
func (h *ContentHandler) ListFollowedClubs(c *gin.Context) {
	clubs, err := h.clubs.ListFollowedClubs(c.Request.Context(), middleware.UserID(c),
		pageSize(c.Query("limit")), parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Clubs you follow retrieved", gin.H{
		"clubs": clubs,
		"count": len(clubs),
	})
}

// ListMemberClubs is the member's own club list.
func (h *ContentHandler) ListMemberClubs(c *gin.Context) {
	clubs, err := h.clubs.ListMemberClubs(c.Request.Context(), middleware.UserID(c),
		pageSize(c.Query("limit")), parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Your clubs retrieved", gin.H{
		"clubs": clubs,
		"count": len(clubs),
	})
}

func (h *ContentHandler) ClubCategories(c *gin.Context) {
	categories, err := h.clubs.ListClubCategories(c.Request.Context())
	if err != nil {
		writeRepoError(c, err, "CATEGORY_NOT_FOUND", "CATEGORY_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Club categories retrieved", gin.H{"categories": categories})
}

// ----------------------------------------------------- club activity and work

// AddClubActivity records something the club did and notifies its followers.
// A club page that only lists events looks empty; this is what a member reads to
// find out the club is still running.
func (h *ContentHandler) AddClubActivity(c *gin.Context) {
	var req domain.CreateClubActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	club, err := h.requireClubOwnership(c)
	if err != nil {
		return
	}

	activity, err := h.clubs.AddClubActivity(c.Request.Context(), club.ID, &req, middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_ACTIVITY_CREATE_FAILED")
		return
	}

	notified, _ := h.notifications.NotifyClubFollowers(c.Request.Context(), club.ID,
		club.Name+" posted an update",
		activity.Title, "/clubs/"+club.ID+"/activities")

	response.Success(c, http.StatusCreated, "Club activity recorded", gin.H{
		"activity":           activity,
		"followers_notified": notified,
	})
}

func (h *ContentHandler) ListClubActivities(c *gin.Context) {
	activities, err := h.clubs.ListClubActivities(c.Request.Context(), c.Param("id"),
		pageSize(c.Query("limit")), parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_ACTIVITY_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Club activity retrieved", gin.H{
		"activities": activities,
		"count":      len(activities),
	})
}

func (h *ContentHandler) DeleteClubActivity(c *gin.Context) {
	club, err := h.requireClubOwnership(c)
	if err != nil {
		return
	}

	if err := h.clubs.DeleteClubActivity(c.Request.Context(), club.ID, c.Param("activityId")); err != nil {
		writeRepoError(c, err, "ACTIVITY_NOT_FOUND", "CLUB_ACTIVITY_DELETE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Club activity removed", nil)
}

// AddClubProject records work the club is doing.
func (h *ContentHandler) AddClubProject(c *gin.Context) {
	var req domain.CreateClubProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	club, err := h.requireClubOwnership(c)
	if err != nil {
		return
	}

	project, err := h.clubs.AddClubProject(c.Request.Context(), club.ID, &req, middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_PROJECT_CREATE_FAILED")
		return
	}
	response.Success(c, http.StatusCreated, "Club project recorded", project)
}

func (h *ContentHandler) ListClubProjects(c *gin.Context) {
	projects, err := h.clubs.ListClubProjects(c.Request.Context(), c.Param("id"),
		pageSize(c.Query("limit")), parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_PROJECT_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Club projects retrieved", gin.H{
		"projects": projects,
		"count":    len(projects),
	})
}

func (h *ContentHandler) UpdateClubProject(c *gin.Context) {
	club, err := h.requireClubOwnership(c)
	if err != nil {
		return
	}

	var req domain.CreateClubProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}
	if strings.TrimSpace(req.Status) == "" {
		response.Error(c, http.StatusBadRequest, "STATUS_REQUIRED", nil)
		return
	}

	project, err := h.clubs.UpdateClubProjectStatus(c.Request.Context(), club.ID,
		c.Param("projectId"), req.Status)
	if err != nil {
		writeRepoError(c, err, "PROJECT_NOT_FOUND", "CLUB_PROJECT_UPDATE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Club project updated", project)
}

// ClubAnalytics is the club's own numbers, for the people who run it.
func (h *ContentHandler) ClubAnalytics(c *gin.Context) {
	club, err := h.requireClubOwnership(c)
	if err != nil {
		return
	}

	stats, err := h.clubs.ClubAnalytics(c.Request.Context(), club.ID)
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_ANALYTICS_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Club analytics retrieved", stats)
}

// requireClubOwnership loads the club in the path and refuses the request unless
// the caller may manage it. It writes the error response itself so callers can
// simply return.
func (h *ContentHandler) requireClubOwnership(c *gin.Context) (*domain.Club, error) {
	club, err := h.clubs.GetClub(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_FETCH_FAILED")
		return nil, err
	}
	if !ownedByCaller(c, club.CreatedBy) {
		response.Error(c, http.StatusForbidden,
			"only the founder or an administrator can change this club", nil)
		return nil, postgres.ErrNotFound
	}
	return club, nil
}

// ----------------------------------------------------------- opportunities

// CreateOpportunity posts a listing.
func (h *ContentHandler) CreateOpportunity(c *gin.Context) {
	var req domain.CreateOpportunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	opportunity, err := h.opportunities.CreateOpportunity(c.Request.Context(), &req, middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_CREATE_FAILED")
		return
	}
	response.Success(c, http.StatusCreated, "Opportunity posted", opportunity)
}

func (h *ContentHandler) UpdateOpportunity(c *gin.Context) {
	var req domain.CreateOpportunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	existing, err := h.opportunities.GetOpportunity(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_FETCH_FAILED")
		return
	}
	if !ownedByCaller(c, existing.OrganizerID) {
		response.Error(c, http.StatusForbidden, "only the poster can change this opportunity", nil)
		return
	}

	opportunity, err := h.opportunities.UpdateOpportunity(c.Request.Context(), existing.ID, &req)
	if err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_UPDATE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Opportunity updated", opportunity)
}

// CloseOpportunity stops new applications without losing the listing.
func (h *ContentHandler) CloseOpportunity(c *gin.Context) {
	existing, err := h.opportunities.GetOpportunity(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_FETCH_FAILED")
		return
	}
	if !ownedByCaller(c, existing.OrganizerID) {
		response.Error(c, http.StatusForbidden, "only the poster can close this opportunity", nil)
		return
	}

	if err := h.opportunities.CloseOpportunity(c.Request.Context(), existing.ID); err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_CLOSE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Opportunity closed", nil)
}

// opportunityApplyRequest is the member's application: an optional note and an
// optional link to a CV. Both are optional because not every listing asks for
// them.
type opportunityApplyRequest struct {
	CoverNote string  `json:"cover_note"`
	ResumeURL *string `json:"resume_url"`
}

// ApplyToOpportunity records an application. Applying twice returns the same
// application rather than a second row, so a double tap is not a data problem.
func (h *ContentHandler) ApplyToOpportunity(c *gin.Context) {
	var req opportunityApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	application, err := h.opportunities.ApplyToOpportunity(c.Request.Context(),
		c.Param("id"), middleware.UserID(c), req.CoverNote, req.ResumeURL)
	if err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_APPLY_FAILED")
		return
	}
	response.Success(c, http.StatusCreated, "Application submitted", application)
}

// WithdrawApplication pulls back an application the member no longer wants to be
// considered for.
func (h *ContentHandler) WithdrawApplication(c *gin.Context) {
	if err := h.opportunities.WithdrawApplication(c.Request.Context(),
		middleware.UserID(c), c.Param("id")); err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_WITHDRAW_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Application withdrawn", nil)
}

// ListApplications is the poster's applicant list.
func (h *ContentHandler) ListOpportunityApplications(c *gin.Context) {
	existing, err := h.opportunities.GetOpportunity(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_FETCH_FAILED")
		return
	}
	if !ownedByCaller(c, existing.OrganizerID) {
		response.Error(c, http.StatusForbidden, "only the poster can see these applications", nil)
		return
	}

	applications, err := h.opportunities.ListApplicationsForOpportunity(c.Request.Context(),
		existing.ID, c.Query("status"), pageSize(c.Query("limit")),
		parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "APPLICATION_NOT_FOUND", "APPLICATION_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Applications retrieved", gin.H{
		"applications": applications,
		"count":        len(applications),
	})
}

// UpdateApplication moves an application forward and tells the applicant. The
// status change is committed first: a member who was shortlisted should not be
// left uninformed because a notification write failed.
func (h *ContentHandler) UpdateOpportunityApplication(c *gin.Context) {
	existing, err := h.opportunities.GetOpportunity(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_FETCH_FAILED")
		return
	}
	if !ownedByCaller(c, existing.OrganizerID) {
		response.Error(c, http.StatusForbidden, "only the poster can change these applications", nil)
		return
	}

	var req domain.UpdateApplicationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	application, err := h.opportunities.UpdateApplicationStatus(c.Request.Context(),
		existing.ID, c.Param("applicationId"), req.Status)
	if err != nil {
		writeRepoError(c, err, "APPLICATION_NOT_FOUND", "APPLICATION_UPDATE_FAILED")
		return
	}

	_ = h.notifications.NotifyUser(c.Request.Context(), application.UserID,
		domain.NotificationTypeOpportunity,
		"Your application for "+existing.Title+" is now "+strings.ToLower(strings.ReplaceAll(req.Status, "_", " ")),
		"You can see the current status in your profile.",
		"/opportunities/"+existing.ID)

	response.Success(c, http.StatusOK, "Application updated", application)
}

// OpportunityAnalytics is the demand picture behind a listing.
func (h *ContentHandler) OpportunityAnalytics(c *gin.Context) {
	existing, err := h.opportunities.GetOpportunity(c.Request.Context(), c.Param("id"), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_FETCH_FAILED")
		return
	}
	if !ownedByCaller(c, existing.OrganizerID) {
		response.Error(c, http.StatusForbidden, "only the poster can see these analytics", nil)
		return
	}

	stats, err := h.opportunities.OpportunityAnalytics(c.Request.Context(), existing.ID)
	if err != nil {
		writeRepoError(c, err, "OPPORTUNITY_NOT_FOUND", "OPPORTUNITY_ANALYTICS_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Opportunity analytics retrieved", stats)
}

func (h *ContentHandler) OpportunityCategories(c *gin.Context) {
	categories, err := h.opportunities.ListOpportunityCategories(c.Request.Context())
	if err != nil {
		writeRepoError(c, err, "CATEGORY_NOT_FOUND", "CATEGORY_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Opportunity categories retrieved", gin.H{
		"categories": categories,
	})
}

func (h *ContentHandler) OpportunityTypes(c *gin.Context) {
	types, err := h.opportunities.ListOpportunityTypes(c.Request.Context())
	if err != nil {
		writeRepoError(c, err, "TYPE_NOT_FOUND", "TYPE_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Opportunity types retrieved", gin.H{"types": types})
}

// ----------------------------------------------------------- announcements

// ListAnnouncements returns live notices, pinned first and filtered by the
// member's own audience. Notices whose scheduled window has not opened, or has
// closed, are simply not shown.
func (h *ContentHandler) ListAnnouncements(c *gin.Context) {
	announcements, err := h.announcements.ListAnnouncements(c.Request.Context(),
		c.Query("category"), c.Query("audience"), pageSize(c.Query("limit")),
		parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "ANNOUNCEMENT_NOT_FOUND", "ANNOUNCEMENT_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Announcements retrieved", gin.H{
		"announcements": announcements,
		"count":         len(announcements),
	})
}

func (h *ContentHandler) GetAnnouncement(c *gin.Context) {
	announcement, err := h.announcements.GetAnnouncement(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeRepoError(c, err, "ANNOUNCEMENT_NOT_FOUND", "ANNOUNCEMENT_FETCH_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Announcement retrieved", announcement)
}

// CreateAnnouncement publishes a campus notice. Announcements are
// admin-only because a notice speaks for the institution, not for a club.
func (h *ContentHandler) CreateAnnouncement(c *gin.Context) {
	var req domain.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	announcement, err := h.announcements.CreateAnnouncement(c.Request.Context(), &req,
		middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "ANNOUNCEMENT_NOT_FOUND", "ANNOUNCEMENT_CREATE_FAILED")
		return
	}

	// Only a notice that is actually live is worth pushing to everyone; a draft
	// or a notice scheduled for later waits until its window opens.
	notified := 0
	if announcement.Status == "PUBLISHED" && announcement.PublishedAt != nil {
		notified, _ = h.announcements.PublishAnnouncement(c.Request.Context(), announcement.ID)
	}

	response.Success(c, http.StatusCreated, "Announcement created", gin.H{
		"announcement":     announcement,
		"members_notified": notified,
	})
}

func (h *ContentHandler) UpdateAnnouncement(c *gin.Context) {
	var req domain.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	existing, err := h.announcements.GetAnnouncement(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeRepoError(c, err, "ANNOUNCEMENT_NOT_FOUND", "ANNOUNCEMENT_FETCH_FAILED")
		return
	}
	if !ownedByCaller(c, existing.AuthorID) {
		response.Error(c, http.StatusForbidden, "only an administrator can change this announcement", nil)
		return
	}

	announcement, err := h.announcements.UpdateAnnouncement(c.Request.Context(), existing.ID, &req)
	if err != nil {
		writeRepoError(c, err, "ANNOUNCEMENT_NOT_FOUND", "ANNOUNCEMENT_UPDATE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Announcement updated", announcement)
}

// ---------------------------------------------------- notification settings

// GetNotificationPreferences returns the member's category switches.
func (h *ContentHandler) GetNotificationPreferences(c *gin.Context) {
	preferences, err := h.notifications.GetPreferences(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "PREFERENCES_NOT_FOUND", "PREFERENCES_FETCH_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Notification preferences retrieved", preferences)
}

// UpdateNotificationPreferences changes which categories reach the member.
func (h *ContentHandler) UpdateNotificationPreferences(c *gin.Context) {
	var req postgres.UpdateNotificationPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	preferences, err := h.notifications.UpdatePreferences(c.Request.Context(),
		middleware.UserID(c), &req)
	if err != nil {
		writeRepoError(c, err, "PREFERENCES_NOT_FOUND", "PREFERENCES_UPDATE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Notification preferences updated", preferences)
}
