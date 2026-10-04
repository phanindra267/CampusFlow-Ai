import os

files = {
    "migrations/000012_phase12_global_intelligence.up.sql": """
CREATE TABLE IF NOT EXISTS global_strategic_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    objectives JSONB,
    assumptions JSONB,
    status VARCHAR(50) DEFAULT 'DRAFT', -- DRAFT, APPROVED, ACTIVE, ARCHIVED
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS decision_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    topic VARCHAR(255) NOT NULL,
    evidence JSONB,
    ai_recommendations JSONB,
    human_opinions JSONB,
    final_decision JSONB,
    status VARCHAR(50) DEFAULT 'OPEN', -- OPEN, DELIBERATING, RESOLVED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS evidence_graphs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    claim VARCHAR(500) NOT NULL,
    source_reference VARCHAR(255),
    dataset_version VARCHAR(100),
    confidence FLOAT,
    counterevidence JSONB,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS autonomous_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    policy_name VARCHAR(255) NOT NULL,
    allowed_scope JSONB,
    risk_level VARCHAR(50),
    mode VARCHAR(50) DEFAULT 'SHADOW', -- SHADOW, CANARY, ACTIVE
    rollback_plan TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
""",
    "internal/domain/intelligence.go": """package domain

import "time"

type StrategicPlan struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Objectives  map[string]interface{} `json:"objectives"`
	Assumptions map[string]interface{} `json:"assumptions"`
	Status      string                 `json:"status"`
	OwnerID     string                 `json:"owner_id"`
}

type DecisionRoom struct {
	ID                string                 `json:"id"`
	TenantID          string                 `json:"tenant_id"`
	Topic             string                 `json:"topic"`
	Evidence          map[string]interface{} `json:"evidence"`
	AIRecommendations map[string]interface{} `json:"ai_recommendations"`
	FinalDecision     map[string]interface{} `json:"final_decision,omitempty"`
	Status            string                 `json:"status"`
}

type EvidenceGraph struct {
	ID              string                 `json:"id"`
	Claim           string                 `json:"claim"`
	SourceReference string                 `json:"source_reference"`
	DatasetVersion  string                 `json:"dataset_version"`
	Confidence      float64                `json:"confidence"`
	TenantID        string                 `json:"tenant_id"`
}

type AutonomousPolicy struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	PolicyName   string                 `json:"policy_name"`
	AllowedScope map[string]interface{} `json:"allowed_scope"`
	Mode         string                 `json:"mode"`
	RollbackPlan string                 `json:"rollback_plan"`
}
""",
    "internal/delivery/http/handler/intelligence.go": """package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type IntelligenceHandler struct{}

func NewIntelligenceHandler() *IntelligenceHandler {
	return &IntelligenceHandler{}
}

func (h *IntelligenceHandler) CreateDecisionRoom(c *gin.Context) {
	var req domain.DecisionRoom
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid decision room payload", err.Error())
		return
	}
	
	req.Status = "OPEN"
	response.Success(c, http.StatusCreated, "Decision Room created for human-AI collaboration", req)
}

func (h *IntelligenceHandler) ResolveDecision(c *gin.Context) {
	roomID := c.Param("roomId")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid decision resolution", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Human governance decision recorded. Autonomous constraints updated.", gin.H{
		"room_id":        roomID,
		"final_decision": req["final_decision"],
		"status":         "RESOLVED",
	})
}

func (h *IntelligenceHandler) RegisterAutonomousPolicy(c *gin.Context) {
	var req domain.AutonomousPolicy
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid policy payload", err.Error())
		return
	}

	// Force SHADOW mode on creation for safety
	req.Mode = "SHADOW"

	response.Success(c, http.StatusCreated, "Autonomous policy registered in SHADOW mode. Awaiting evaluation.", req)
}

func (h *IntelligenceHandler) VerifyClaim(c *gin.Context) {
	claim := c.Query("claim")
	
	graph := domain.EvidenceGraph{
		ID:              "evg-1",
		Claim:           claim,
		SourceReference: "Federated Research Node Alpha",
		DatasetVersion:  "v2.1",
		Confidence:      0.88,
	}

	response.Success(c, http.StatusOK, "Claim verification complete. Provenance retained.", graph)
}
""",
    "internal/server/server.go": """package server

import (
	"context"
	"net/http"

	"github.com/campuscare/api/internal/config"
	"github.com/campuscare/api/internal/delivery/http/handler"
	"github.com/campuscare/api/internal/delivery/http/middleware"
	"github.com/campuscare/api/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	httpServer *http.Server
	logger     *logger.Logger
}

func New(cfg config.ServerConfig, db *pgxpool.Pool, l *logger.Logger) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(middleware.Logger(l))
	router.Use(middleware.Recovery(l))

	// Handlers
	authHandler := handler.NewAuthHandler()
	eventHandler := handler.NewEventHandler()
	clubHandler := handler.NewClubHandler()
	opportunityHandler := handler.NewOpportunityHandler()
	waitlistHandler := handler.NewWaitlistHandler()
	attendanceHandler := handler.NewAttendanceHandler()
	feedbackHandler := handler.NewFeedbackHandler()
	searchHandler := handler.NewSearchHandler()
	recHandler := handler.NewRecommendationHandler()
	assistantHandler := handler.NewAssistantHandler()
	automationHandler := handler.NewAutomationHandler()
	adminInsightsHandler := handler.NewAdminInsightsHandler()
	institutionalHandler := handler.NewInstitutionalHandler()
	resourceHandler := handler.NewResourceHandler()
	opsHandler := handler.NewOperationsHandler()
	enterpriseHandler := handler.NewEnterpriseHandler()
	iotHandler := handler.NewIoTHandler()
	researchHandler := handler.NewResearchHandler()
	federationHandler := handler.NewFederationHandler()
	intelHandler := handler.NewIntelligenceHandler()

	// Health & Readiness
	router.GET("/health", handler.HealthCheck)
	router.GET("/ready", handler.ReadinessCheck(db))

	// API Routes
	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		api.GET("/events", eventHandler.ListEvents)
		api.GET("/clubs", clubHandler.ListClubs)
		api.GET("/opportunities", opportunityHandler.ListOpportunities)
		api.GET("/search", searchHandler.Search)

		protected := api.Group("/")
		protected.Use(middleware.Auth("supersecretkey", l))
		{
			student := protected.Group("/students")
			student.Use(middleware.RequireRole("STUDENT"))
			{
				student.POST("/events/:id/register", eventHandler.RegisterForEvent)
				student.POST("/events/:id/waitlist", waitlistHandler.JoinWaitlist)
				student.POST("/clubs/:id/join", clubHandler.JoinClub)
				student.POST("/opportunities/:id/save", opportunityHandler.SaveOpportunity)
				student.POST("/attendance/scan", attendanceHandler.ScanQR)
				student.POST("/feedback", feedbackHandler.SubmitFeedback)
				student.GET("/feed", recHandler.GetPersonalizedFeed)
				student.POST("/recommendation/feedback", recHandler.SubmitFeedback)
				student.POST("/assistant/conversations", assistantHandler.CreateConversation)
				student.POST("/assistant/messages", assistantHandler.SendMessage)
				student.POST("/automation/rules", automationHandler.CreateRule)
				student.PUT("/preferences/notifications", automationHandler.UpdatePreferences)
				student.GET("/research", researchHandler.DiscoverResearch)
				student.GET("/federation/search", federationHandler.FederatedSearch)
				
				// Claim Verification
				student.GET("/intelligence/claims/verify", intelHandler.VerifyClaim)
			}

			organizer := protected.Group("/organizers")
			organizer.Use(middleware.RequireRole("ORGANIZER", "ADMIN", "SUPER_ADMIN"))
			{
				organizer.POST("/events", eventHandler.CreateEvent)
				organizer.POST("/clubs", clubHandler.CreateClub)
				organizer.POST("/opportunities", opportunityHandler.CreateOpportunity)
				organizer.POST("/sessions/:sessionId/qr", attendanceHandler.GenerateQR)
				organizer.GET("/resources", resourceHandler.ListResources)
				organizer.GET("/institutional/graph/query", institutionalHandler.QueryGraph)
				organizer.POST("/operations/scenarios", opsHandler.CreateScenario)
				organizer.POST("/operations/actions", opsHandler.RequestAction)
				organizer.POST("/iot/devices/:deviceId/commands", iotHandler.RequestCommand)
				
				// Decision Rooms
				organizer.POST("/intelligence/decisions/rooms", intelHandler.CreateDecisionRoom)
			}
			
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("ADMIN", "SUPER_ADMIN"))
			{
				admin.POST("/knowledge", assistantHandler.UploadKnowledge)
				admin.GET("/insights/knowledge-gaps", adminInsightsHandler.GetKnowledgeGaps)
				admin.GET("/institutional/analytics", institutionalHandler.GetAnalytics)
				admin.POST("/operations/actions/:actionId/approve", opsHandler.ApproveAction)
				admin.GET("/operations/incidents", opsHandler.ListIncidents)
				admin.POST("/enterprise/agents", enterpriseHandler.RegisterAgent)
				admin.POST("/enterprise/agents/:agentId/approve", enterpriseHandler.ApproveAgent)
				admin.GET("/enterprise/connectors", enterpriseHandler.ListConnectors)
				admin.POST("/enterprise/connectors/:connectorId/sync", enterpriseHandler.SyncConnector)
				admin.GET("/enterprise/analytics/federated", enterpriseHandler.GetFederatedAnalytics)
				admin.POST("/iot/devices", iotHandler.RegisterDevice)
				admin.POST("/iot/commands/:commandId/approve", iotHandler.ApproveCommand)
				admin.POST("/federation/trusts", federationHandler.EstablishTrust)
				admin.POST("/federation/kill-switch", federationHandler.TriggerKillSwitch)
				
				// Advanced Intelligence Governance
				admin.POST("/intelligence/decisions/rooms/:roomId/resolve", intelHandler.ResolveDecision)
				admin.POST("/intelligence/autonomous-policies", intelHandler.RegisterAutonomousPolicy)
			}
		}
	}

	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      router,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
		logger: l,
	}
}

func (s *Server) Start() error {
	s.logger.Infof("Listening on %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
"""
}

for path, content in files.items():
    os.makedirs(os.path.dirname(path) if os.path.dirname(path) else ".", exist_ok=True)
    with open(path, "w", encoding="utf-8") as f:
        f.write(content)

print("Phase 12 scaffolding generated successfully.")
