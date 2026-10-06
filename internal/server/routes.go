package server

import (
	"context"
	"log"
	"time"

	"github.com/campuscare/api/internal/config"
	"github.com/campuscare/api/internal/delivery/http/handler"
	"github.com/campuscare/api/internal/delivery/http/middleware"
	"github.com/campuscare/api/internal/repository/postgres"
	"github.com/campuscare/api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
)

// newWeaviateClient dials the vector store once at start-up. A failure is
// logged and returned as nil rather than fatal: the semantic arm is an
// enhancement, and CampusCare must stay fully usable with BM25-only search.
func newWeaviateClient(cfg *config.Config) *weaviate.Client {
	host := cfg.Weaviate.Host
	if host == "" {
		return nil
	}

	client, err := weaviate.NewClient(weaviate.Config{
		Host:    host,
		Scheme:  cfg.Weaviate.Scheme,
		Timeout: 5 * time.Second,
		Headers: map[string]string{},
	})

	if err != nil {
		log.Printf("Weaviate unavailable at %s (%v). Semantic search disabled; BM25 hybrid search remains active.", host, err)
		return nil
	}

	// Confirm the store answers and actually holds the retrieval class. The two
	// failures are reported separately: an unreachable store is an outage,
	// whereas a reachable store with no chunks simply means nothing has been
	// indexed yet.
	probe, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := client.GraphQL().Get().
		WithClassName(service.DocumentChunkClass).
		WithFields(graphql.Field{Name: "source"}).
		WithLimit(1).
		Do(probe)
	if err != nil {
		log.Printf("Weaviate reachable at %s but not responding (%v). Semantic search disabled; BM25 lexical search remains active.", host, err)
		return nil
	}

	if len(result.Errors) > 0 || len(result.Data) == 0 {
		log.Printf(
			"Weaviate is running at %s but the %s class is not indexed. "+
				"Semantic search is inactive; BM25 lexical search remains active.",
			host, service.DocumentChunkClass,
		)
		return nil
	}

	return client
}

// Roles recognised by the authorisation middleware. There is no FACULTY role
// because CampusCare AI is a community platform rather than an academic one.
var (
	adminRoles   = []string{"ADMIN", "SUPER_ADMIN"}
	creatorRoles = []string{"ORGANIZER", "ADMIN", "SUPER_ADMIN"}
)

// handlers bundles every HTTP handler so routes can be registered in one place.
type handlers struct {
	auth      *handler.AuthHandler
	oauth     *handler.OAuthHandler
	readiness gin.HandlerFunc

	campus    *handler.CampusHandler
	community *handler.CommunityHandler
	content   *handler.ContentHandler
	service   *handler.ServiceHandler
	people    *handler.PeopleHandler
	admin     *handler.AdminHandler

	institutional *handler.InstitutionalHandler
	ai            *handler.AIHandler
}

func newHandlers(cfg *config.Config, db *pgxpool.Pool, userRepo *postgres.UserRepository) *handlers {
	authCfg := handler.AuthConfig{
		Secret:    cfg.JWT.Secret,
		Issuer:    cfg.JWT.Issuer,
		AccessTTL: cfg.JWT.TTL,
	}

	notifications := postgres.NewNotificationRepository(db)

	return &handlers{
		auth:      handler.NewAuthHandler(authCfg, userRepo),
		oauth:     handler.NewOAuthHandler(authCfg, userRepo),
		readiness: handler.ReadinessCheck(db),

		campus:    handler.NewCampusHandler(postgres.NewCampusRepository(db), notifications),
		community: handler.NewCommunityHandler(postgres.NewCommunityRepository(db)),
		content: handler.NewContentHandler(
			postgres.NewEventRepository(db),
			postgres.NewClubRepository(db),
			postgres.NewOpportunityRepository(db),
			postgres.NewAnnouncementRepository(db),
			notifications,
		),
		service: handler.NewServiceHandler(postgres.NewServiceRepository(db), notifications),
		people:  handler.NewPeopleHandler(postgres.NewPeopleRepository(db)),
		admin:   handler.NewAdminHandler(userRepo, postgres.NewClubRepository(db)),

		institutional: handler.NewInstitutionalHandler(
			postgres.NewAnalyticsRepository(db),
			postgres.NewGraphRepository(db),
		),

		// Retrieval for the assistant. The Weaviate client is optional: without
		// it the semantic arm is skipped and hybrid search runs BM25 only.
		ai: handler.NewAIHandler(
			service.NewRAGService(newWeaviateClient(cfg), service.NewCorpusSource(db)),
		),
	}
}

func registerRoutes(router *gin.Engine, cfg *config.Config, h *handlers) {
	api := router.Group("/api/v1")

	api.GET("/health/live", handler.HealthCheck)
	api.GET("/health/ready", h.readiness)

	// ---- Authentication ------------------------------------------------
	//
	// The credential endpoints are the only unauthenticated surface that lets a
	// caller choose what the server hashes, so they are the only surface worth
	// throttling. Everything here is rate limited per client IP.
	//
	// login and refresh are the brute-force surface: both take an attacker-supplied
	// secret and both cost a bcrypt comparison on success or a token signature on
	// failure, so they get the strict budget.
	strict := api.Group("/", middleware.RateLimit(
		middleware.NewRateLimiter(middleware.RateLimitConfig{
			PerMinute: cfg.RateLimit.AuthPerMinute,
			Burst:     cfg.RateLimit.AuthBurst,
		}), middleware.KeyByClientIP))

	strict.POST("/auth/login", h.auth.Login)
	// Refresh carries the refresh token in the body, so it is reachable without
	// a valid (already expired) access token.
	strict.POST("/auth/refresh", h.auth.Refresh)

	// register creates an account and is therefore a mass-account surface, but it
	// does not take a guessable secret, so it gets a looser budget than login.
	// Without this, one person registering legitimately from a shared campus NAT
	// could exhaust the login budget for everyone behind the same address.
	relaxed := api.Group("/", middleware.RateLimit(
		middleware.NewRateLimiter(middleware.RateLimitConfig{
			PerMinute: max(cfg.RateLimit.AuthPerMinute*3, 1),
			Burst:     max(cfg.RateLimit.AuthBurst*3, 1),
		}), middleware.KeyByClientIP))

	relaxed.POST("/auth/register", h.auth.Register)

	api.GET("/auth/google", h.oauth.GoogleLogin)
	api.GET("/auth/google/callback", h.oauth.GoogleCallback)

	// Everything below requires a valid, non-expired access token.
	protected := api.Group("/")
	protected.Use(middleware.RequireAuth(cfg.JWT.Secret, cfg.JWT.Issuer))

	protected.GET("/auth/me", h.auth.Me)
	protected.PATCH("/auth/me", h.auth.UpdateProfile)
	protected.GET("/auth/profile", h.auth.Profile)

	registerCommunityRoutes(protected, h)
	registerCampusRoutes(protected, h)
	registerOrganiserRoutes(protected, h)
	registerContentRoutes(protected, h)
	registerServiceRoutes(protected, h)
	registerPeopleRoutes(protected, h)
	registerAIRoutes(protected, h)
	registerInstitutionalRoutes(protected, h)
}

// registerCommunityRoutes exposes the collaboration layer: discussions,
// notifications, campus services, bookings and search.
func registerCommunityRoutes(r *gin.RouterGroup, h *handlers) {
	community := r.Group("/community")
	{
		community.GET("/search", h.community.Search)

		community.GET("/discussions", h.community.ListDiscussions)
		community.GET("/discussions/:id", h.community.GetDiscussion)
		community.POST("/discussions", h.community.CreateDiscussion)
		community.POST("/discussions/:id/replies", h.community.AddReply)

		community.GET("/notifications", h.community.ListNotifications)
		community.POST("/notifications/read-all", h.community.MarkAllNotificationsRead)
		community.POST("/notifications/:id/read", h.community.MarkNotificationRead)

		community.GET("/bookings", h.community.ListMyBookings)
		community.DELETE("/bookings/:id", h.community.CancelBooking)

		community.GET("/services", h.community.ListResources)
		community.POST("/services/:id/bookings", h.community.BookResource)
	}

	// Personal collections are namespaced under /me so they cannot collide with
	// the :id routes above. Club and follow lists live in registerContentRoutes,
	// which reads them in a single query.
	me := r.Group("/me")
	{
		me.GET("/notifications", h.community.ListNotifications)
		me.GET("/opportunities", h.community.ListMyApplications)
		me.GET("/bookings", h.community.ListMyBookings)
		me.GET("/activity", h.campus.MyActivity)
	}
}

// registerCampusRoutes exposes what members browse and join: groups, listings
// and events.
func registerCampusRoutes(r *gin.RouterGroup, h *handlers) {
	clubs := r.Group("/clubs")
	{
		clubs.GET("", h.content.ListClubs)
		clubs.GET("/categories", h.content.ClubCategories)
		clubs.GET("/:id", h.campus.GetClub)
		clubs.POST("/:id/join", h.campus.JoinClub)
		clubs.POST("/:id/leave", h.campus.LeaveClub)
		clubs.POST("/:id/follow", h.content.FollowClub)
		clubs.DELETE("/:id/follow", h.content.UnfollowClub)
	}

	opportunities := r.Group("/opportunities")
	{
		opportunities.GET("", h.campus.ListOpportunities)
		opportunities.GET("/categories", h.content.OpportunityCategories)
		opportunities.GET("/types", h.content.OpportunityTypes)
		opportunities.GET("/:id", h.campus.GetOpportunity)
		opportunities.POST("/:id/save", h.campus.SaveOpportunity)
		opportunities.DELETE("/:id/save", h.campus.UnsaveOpportunity)
		opportunities.POST("/:id/apply", h.content.ApplyToOpportunity)
		opportunities.DELETE("/:id/apply", h.content.WithdrawApplication)
	}

	events := r.Group("/events")
	{
		events.GET("", h.campus.ListEvents)
		events.GET("/categories", h.content.EventCategories)
		events.GET("/:id", h.campus.GetEvent)
		events.POST("/:id/register", h.campus.RegisterForEvent)
		events.POST("/:id/cancel", h.campus.CancelRegistration)
		events.POST("/:id/save", h.content.SaveEvent)
		events.DELETE("/:id/save", h.content.UnsaveEvent)
		events.PUT("/:id/reminder", h.content.SetEventReminder)
		events.DELETE("/:id/reminder", h.content.ClearEventReminder)
		events.GET("/:id/waitlist", h.campus.ListWaitlist)

		// Scanned by attendees at the door.
		events.POST("/check-in", h.campus.CheckIn)
	}

	feedback := r.Group("/feedback")
	{
		feedback.POST("/:type/:id", h.campus.SubmitFeedback)
	}
}

// registerOrganiserRoutes covers event operations that organisers and admins
// perform.
func registerOrganiserRoutes(r *gin.RouterGroup, h *handlers) {
	organiser := r.Group("/organiser")
	organiser.Use(middleware.RequireRole(creatorRoles...))
	{
		organiser.POST("/sessions/:id/check-in/open", h.campus.OpenCheckIn)
		organiser.GET("/sessions/:id/attendance", h.campus.ListAttendance)

		organiser.GET("/events", h.content.ListOrganiserEvents)
		organiser.GET("/events/:id/analytics", h.content.EventAnalytics)
		organiser.GET("/events/:id/registrations", h.content.ListEventRegistrations)
		organiser.POST("/events/:id/waitlist/promote", h.content.PromoteWaitlist)

		organiser.GET("/opportunities/:id/applications", h.content.ListOpportunityApplications)
		organiser.PATCH("/opportunities/:id/applications/:applicationId",
			h.content.UpdateOpportunityApplication)
		organiser.GET("/opportunities/:id/analytics", h.content.OpportunityAnalytics)

		organiser.GET("/clubs/:id/analytics", h.content.ClubAnalytics)
	}
}

// registerContentRoutes is the write side of the platform: organisers and admins
// create and maintain the events, clubs, opportunities and notices that make up
// the campus calendar. Ownership is checked per resource inside the handler.
func registerContentRoutes(r *gin.RouterGroup, h *handlers) {
	creators := r.Group("/")
	creators.Use(middleware.RequireRole(creatorRoles...))
	{
		events := creators.Group("/events")
		{
			events.POST("", h.content.CreateEvent)
			events.PATCH("/:id", h.content.UpdateEvent)
			events.DELETE("/:id", h.content.CancelEvent)
		}

		clubs := creators.Group("/clubs")
		{
			clubs.POST("", h.content.CreateClub)
			clubs.PATCH("/:id", h.content.UpdateClub)
			clubs.DELETE("/:id", h.content.ArchiveClub)

			clubs.POST("/:id/activities", h.content.AddClubActivity)
			clubs.DELETE("/:id/activities/:activityId", h.content.DeleteClubActivity)
			clubs.POST("/:id/projects", h.content.AddClubProject)
			clubs.PATCH("/:id/projects/:projectId", h.content.UpdateClubProject)
		}

		opportunities := creators.Group("/opportunities")
		{
			opportunities.POST("", h.content.CreateOpportunity)
			opportunities.PATCH("/:id", h.content.UpdateOpportunity)
			opportunities.DELETE("/:id", h.content.CloseOpportunity)
		}
	}

	// Reading club activity and projects is open to any member: a club page is a
	// public surface, and its history is what a prospective member wants to see.
	clubs := r.Group("/clubs")
	{
		clubs.GET("/:id/activities", h.content.ListClubActivities)
		clubs.GET("/:id/projects", h.content.ListClubProjects)
	}

	// Announcements speak for the institution, so only administrators post them.
	admin := r.Group("/admin")
	admin.Use(middleware.RequireRole(adminRoles...))
	{
		admin.GET("/users", h.admin.ListUsers)
		admin.GET("/approvals/clubs", h.admin.ListPendingClubApprovals)
		admin.PATCH("/approvals/clubs/:id", h.admin.UpdateClubVerification)

		admin.POST("/announcements", h.content.CreateAnnouncement)
		admin.PATCH("/announcements/:id", h.content.UpdateAnnouncement)

		admin.GET("/service-requests", h.service.ListServiceQueue)
		admin.PATCH("/service-requests/:id", h.service.UpdateServiceRequest)
		admin.GET("/service-requests/queue-stats", h.service.QueueStats)

		admin.POST("/people", h.people.CreatePerson)
		admin.PATCH("/people/:id", h.people.UpdatePerson)
	}

	// Announcements are read by everyone, including the scheduled and archived
	// lifecycle states that the public listing hides.
	announcements := r.Group("/announcements")
	{
		announcements.GET("", h.content.ListAnnouncements)
		announcements.GET("/:id", h.content.GetAnnouncement)
	}

	// Personal collections and settings live under /me so they cannot collide
	// with the :id routes above.
	me := r.Group("/me")
	{
		me.GET("/clubs", h.content.ListMemberClubs)
		me.GET("/following", h.content.ListFollowedClubs)
		me.GET("/saved-events", h.content.ListSavedEvents)
		me.GET("/notification-preferences", h.content.GetNotificationPreferences)
		me.PATCH("/notification-preferences", h.content.UpdateNotificationPreferences)

		me.GET("/research-interests", h.people.ListMyResearchInterests)
		me.POST("/research-interests", h.people.AddMyResearchInterest)
		me.DELETE("/research-interests/:interest", h.people.RemoveMyResearchInterest)
	}
}

// registerServiceRoutes is the campus service desk: the service pages a member
// reads, the ticket they file, and the timeline they follow it on.
func registerServiceRoutes(r *gin.RouterGroup, h *handlers) {
	requests := r.Group("/service-requests")
	{
		requests.POST("", h.service.CreateServiceRequest)
		requests.GET("/:id", h.service.GetServiceRequest)
		requests.POST("/:id/messages", h.service.AddServiceRequestMessage)
		requests.POST("/:id/rating", h.service.RateServiceRequest)
		requests.POST("/:id/attachments", h.service.AddServiceAttachment)
	}

	me := r.Group("/me")
	{
		me.GET("/service-requests", h.service.ListMyServiceRequests)
	}

	services := r.Group("/services")
	{
		services.GET("/:id/faqs", h.service.ListServiceFAQs)
	}

	// Editing a service page's FAQs is staff work: the answer is the institution's.
	staff := r.Group("/")
	staff.Use(middleware.RequireRole(creatorRoles...))
	{
		staff.POST("/services/:id/faqs", h.service.AddServiceFAQ)
		staff.DELETE("/services/:id/faqs/:faqId", h.service.DeleteServiceFAQ)
	}
}

// registerPeopleRoutes is the campus people directory plus the research profile a
// member fills in to be findable.
func registerPeopleRoutes(r *gin.RouterGroup, h *handlers) {
	people := r.Group("/people")
	{
		people.GET("", h.people.ListPeople)
		people.GET("/schools", h.people.ListSchools)
		people.GET("/research-interests", h.people.ListResearchInterests)
		people.GET("/by-interest", h.people.FindPeopleByInterest)
		people.GET("/matches", h.people.FindMembersByResearchInterest)
		people.GET("/:id", h.people.GetPerson)
	}
}

// registerAIRoutes exposes retrieval only. The assistant's generative step runs
// in the browser via WebLLM on WebGPU, so there is deliberately no endpoint
// that accepts a prompt and proxies it to a model.
func registerAIRoutes(r *gin.RouterGroup, h *handlers) {
	ai := r.Group("/ai")
	{
		ai.GET("/context", h.ai.Context)
	}
}

func registerInstitutionalRoutes(r *gin.RouterGroup, h *handlers) {
	institutional := r.Group("/institutional")
	// Campus-wide analytics and the knowledge graph are for admins only.
	institutional.Use(middleware.RequireRole(adminRoles...))
	{
		institutional.GET("/analytics", h.institutional.GetAnalytics)
		institutional.GET("/graph", h.institutional.QueryGraph)
	}
}
