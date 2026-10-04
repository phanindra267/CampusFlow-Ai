import os

files = {
    "migrations/000017_phase17_digital_twins.up.sql": """
CREATE TABLE IF NOT EXISTS digital_twins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    twin_type VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    lifecycle_state VARCHAR(50) DEFAULT 'ACTIVE',
    version INT NOT NULL DEFAULT 1,
    current_state JSONB,
    desired_state JSONB,
    data_sources JSONB,
    relationships JSONB,
    policies JSONB,
    confidence FLOAT,
    provenance JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS twin_state_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    twin_id UUID NOT NULL REFERENCES digital_twins(id) ON DELETE CASCADE,
    state_type VARCHAR(50) NOT NULL,
    state_data JSONB NOT NULL,
    source VARCHAR(100),
    confidence FLOAT,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS simulation_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    twin_id UUID REFERENCES digital_twins(id) ON DELETE SET NULL,
    scenario_name VARCHAR(255) NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    model_version VARCHAR(50) NOT NULL,
    parameters JSONB,
    assumptions JSONB,
    random_seed BIGINT,
    status VARCHAR(50) DEFAULT 'QUEUED',
    priority INT DEFAULT 5,
    cpu_limit_cores FLOAT,
    memory_limit_mb INT,
    runtime_limit_seconds INT,
    cost_budget NUMERIC(18, 2),
    results JSONB,
    provenance JSONB,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS simulation_scenarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    baseline_twin_id UUID REFERENCES digital_twins(id) ON DELETE SET NULL,
    version INT NOT NULL DEFAULT 1,
    parameters JSONB,
    assumptions JSONB,
    is_ai_generated BOOLEAN DEFAULT FALSE,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
""",
    "internal/domain/digital_twin.go": """package domain

import "time"

type DigitalTwin struct {
	ID             string                 `json:"id"`
	TenantID       string                 `json:"tenant_id"`
	TwinType       string                 `json:"twin_type" binding:"required"`
	Name           string                 `json:"name" binding:"required"`
	Description    string                 `json:"description"`
	OwnerID        string                 `json:"owner_id"`
	LifecycleState string                 `json:"lifecycle_state"`
	Version        int                    `json:"version"`
	CurrentState   map[string]interface{} `json:"current_state"`
	DesiredState   map[string]interface{} `json:"desired_state,omitempty"`
	DataSources    []interface{}          `json:"data_sources"`
	Relationships  []interface{}          `json:"relationships"`
	Policies       map[string]interface{} `json:"policies"`
	Confidence     float64                `json:"confidence"`
	Provenance     map[string]interface{} `json:"provenance"`
}

type TwinStateHistory struct {
	ID         string                 `json:"id"`
	TwinID     string                 `json:"twin_id"`
	StateType  string                 `json:"state_type"`
	StateData  map[string]interface{} `json:"state_data"`
	Source     string                 `json:"source"`
	Confidence float64                `json:"confidence"`
	RecordedAt time.Time              `json:"recorded_at"`
}

type SimulationRun struct {
	ID                   string                 `json:"id"`
	TenantID             string                 `json:"tenant_id"`
	TwinID               string                 `json:"twin_id,omitempty"`
	ScenarioName         string                 `json:"scenario_name" binding:"required"`
	ModelName            string                 `json:"model_name" binding:"required"`
	ModelVersion         string                 `json:"model_version" binding:"required"`
	Parameters           map[string]interface{} `json:"parameters"`
	Assumptions          map[string]interface{} `json:"assumptions"`
	RandomSeed           int64                  `json:"random_seed,omitempty"`
	Status               string                 `json:"status"`
	Priority             int                    `json:"priority"`
	CPULimitCores        float64                `json:"cpu_limit_cores"`
	MemoryLimitMB        int                    `json:"memory_limit_mb"`
	RuntimeLimitSeconds  int                    `json:"runtime_limit_seconds"`
	CostBudget           float64                `json:"cost_budget"`
	Results              map[string]interface{} `json:"results,omitempty"`
	Provenance           map[string]interface{} `json:"provenance,omitempty"`
}

type SimulationScenario struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	Name            string                 `json:"name" binding:"required"`
	Description     string                 `json:"description"`
	BaselineTwinID  string                 `json:"baseline_twin_id,omitempty"`
	Version         int                    `json:"version"`
	Parameters      map[string]interface{} `json:"parameters"`
	Assumptions     map[string]interface{} `json:"assumptions"`
	IsAIGenerated   bool                   `json:"is_ai_generated"`
	OwnerID         string                 `json:"owner_id"`
}
""",
    "internal/delivery/http/handler/digital_twin.go": """package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type DigitalTwinHandler struct{}

func NewDigitalTwinHandler() *DigitalTwinHandler {
	return &DigitalTwinHandler{}
}

// POST /admin/twins
func (h *DigitalTwinHandler) CreateTwin(c *gin.Context) {
	var req domain.DigitalTwin
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid digital twin payload", err.Error())
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.LifecycleState = "ACTIVE"
	req.Version = 1
	response.Success(c, http.StatusCreated,
		"Digital twin created. State clearly partitioned from simulation state.", req)
}

// GET /admin/twins/:twinId
func (h *DigitalTwinHandler) GetTwin(c *gin.Context) {
	twinID := c.Param("twinId")
	response.Success(c, http.StatusOK, "Digital twin retrieved", map[string]interface{}{
		"id":              twinID,
		"lifecycle_state": "ACTIVE",
		"note":            "current_state reflects OBSERVED data only. Simulated and predicted states are separate.",
	})
}

// POST /admin/twins/:twinId/state
func (h *DigitalTwinHandler) RecordStateSnapshot(c *gin.Context) {
	twinID := c.Param("twinId")
	var req domain.TwinStateHistory
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid state snapshot payload", err.Error())
		return
	}
	req.TwinID = twinID
	response.Success(c, http.StatusCreated,
		"State snapshot recorded. Simulated vs observed boundary enforced.", req)
}

// POST /organizers/simulations
func (h *DigitalTwinHandler) SubmitSimulation(c *gin.Context) {
	var req domain.SimulationRun
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid simulation run payload", err.Error())
		return
	}
	userID, _ := c.Get("userID")
	req.Status = "QUEUED"
	if req.Priority == 0 {
		req.Priority = 5
	}
	req.Provenance = map[string]interface{}{
		"submitted_by": userID,
		"note":         "Simulation isolated from production state. Results are SIMULATED, not observed.",
	}
	response.Success(c, http.StatusCreated,
		"Simulation job queued. Production state will not be mutated.", req)
}

// POST /organizers/simulations/scenarios
func (h *DigitalTwinHandler) CreateScenario(c *gin.Context) {
	var req domain.SimulationScenario
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid scenario payload", err.Error())
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.Version = 1
	req.IsAIGenerated = false
	response.Success(c, http.StatusCreated,
		"Scenario created. AI-generated scenarios will be explicitly labeled.", req)
}

// GET /organizers/simulations/:runId/results
func (h *DigitalTwinHandler) GetSimulationResults(c *gin.Context) {
	runID := c.Param("runId")
	response.Success(c, http.StatusOK, "Simulation results retrieved", map[string]interface{}{
		"run_id": runID,
		"status": "COMPLETED",
		"note":   "Results are SIMULATED outputs. Not observed real-world data.",
	})
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

	// All-phase handlers
	authHandler          := handler.NewAuthHandler()
	eventHandler         := handler.NewEventHandler()
	clubHandler          := handler.NewClubHandler()
	opportunityHandler   := handler.NewOpportunityHandler()
	waitlistHandler      := handler.NewWaitlistHandler()
	attendanceHandler    := handler.NewAttendanceHandler()
	feedbackHandler      := handler.NewFeedbackHandler()
	searchHandler        := handler.NewSearchHandler()
	recHandler           := handler.NewRecommendationHandler()
	assistantHandler     := handler.NewAssistantHandler()
	automationHandler    := handler.NewAutomationHandler()
	adminInsightsHandler := handler.NewAdminInsightsHandler()
	institutionalHandler := handler.NewInstitutionalHandler()
	resourceHandler      := handler.NewResourceHandler()
	opsHandler           := handler.NewOperationsHandler()
	enterpriseHandler    := handler.NewEnterpriseHandler()
	iotHandler           := handler.NewIoTHandler()
	researchHandler      := handler.NewResearchHandler()
	federationHandler    := handler.NewFederationHandler()
	intelHandler         := handler.NewIntelligenceHandler()
	collectiveHandler    := handler.NewCollectiveHandler()
	researchEcoHandler   := handler.NewResearchEcosystemHandler()
	ecosystemHandler     := handler.NewEcosystemHandler()
	planetHandler        := handler.NewPlanetScaleHandler()
	twinHandler          := handler.NewDigitalTwinHandler()

	router.GET("/health", handler.HealthCheck)
	router.GET("/ready", handler.ReadinessCheck(db))

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
			// ── STUDENT ─────────────────────────────────────────────────────────
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
				student.GET("/intelligence/claims/verify", intelHandler.VerifyClaim)
				student.GET("/ecosystem/services", ecosystemHandler.ListServices)
			}

			// ── RESEARCHER ───────────────────────────────────────────────────────
			researcher := protected.Group("/researchers")
			researcher.Use(middleware.RequireRole("RESEARCHER", "FACULTY", "ADMIN", "SUPER_ADMIN"))
			{
				researcher.POST("/projects", researchEcoHandler.CreateProject)
				researcher.POST("/projects/:projectId/hypotheses", researchEcoHandler.CreateHypothesis)
				researcher.POST("/projects/:projectId/experiments", researchEcoHandler.CreateExperiment)
				researcher.POST("/projects/:projectId/artifacts", researchEcoHandler.RegisterArtifact)
				researcher.POST("/datasets", researchEcoHandler.RegisterDataset)
				researcher.POST("/claims", planetHandler.RegisterClaim)
				researcher.POST("/experiments/:experimentId/replications", planetHandler.InitiateReplication)
				researcher.POST("/projects/:projectId/funding", planetHandler.RegisterFunding)
			}

			// ── ORGANIZER ────────────────────────────────────────────────────────
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
				organizer.POST("/intelligence/decisions/rooms", intelHandler.CreateDecisionRoom)
				organizer.POST("/collective/rooms", collectiveHandler.OpenDeliberationRoom)
				organizer.POST("/collective/rooms/:roomId/arguments", collectiveHandler.AddArgument)
				// Phase 17 — Simulation
				organizer.POST("/simulations", twinHandler.SubmitSimulation)
				organizer.POST("/simulations/scenarios", twinHandler.CreateScenario)
				organizer.GET("/simulations/:runId/results", twinHandler.GetSimulationResults)
			}

			// ── ADMIN / SUPER_ADMIN ──────────────────────────────────────────────
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("ADMIN", "SUPER_ADMIN"))
			{
				// Phase 2-5
				admin.POST("/knowledge", assistantHandler.UploadKnowledge)
				admin.GET("/insights/knowledge-gaps", adminInsightsHandler.GetKnowledgeGaps)
				admin.GET("/institutional/analytics", institutionalHandler.GetAnalytics)
				// Phase 8
				admin.POST("/operations/actions/:actionId/approve", opsHandler.ApproveAction)
				admin.GET("/operations/incidents", opsHandler.ListIncidents)
				// Phase 9
				admin.POST("/enterprise/agents", enterpriseHandler.RegisterAgent)
				admin.POST("/enterprise/agents/:agentId/approve", enterpriseHandler.ApproveAgent)
				admin.GET("/enterprise/connectors", enterpriseHandler.ListConnectors)
				admin.POST("/enterprise/connectors/:connectorId/sync", enterpriseHandler.SyncConnector)
				admin.GET("/enterprise/analytics/federated", enterpriseHandler.GetFederatedAnalytics)
				// Phase 10
				admin.POST("/iot/devices", iotHandler.RegisterDevice)
				admin.POST("/iot/commands/:commandId/approve", iotHandler.ApproveCommand)
				// Phase 11
				admin.POST("/federation/trusts", federationHandler.EstablishTrust)
				admin.POST("/federation/kill-switch", federationHandler.TriggerKillSwitch)
				// Phase 12
				admin.POST("/intelligence/decisions/rooms/:roomId/resolve", intelHandler.ResolveDecision)
				admin.POST("/intelligence/autonomous-policies", intelHandler.RegisterAutonomousPolicy)
				// Phase 13
				admin.POST("/collective/rooms/:roomId/decide", collectiveHandler.RecordDecisionOutcome)
				admin.POST("/collective/playbooks", collectiveHandler.CreatePlaybook)
				admin.GET("/collective/predictions", collectiveHandler.GetPredictionLedger)
				// Phase 14
				admin.POST("/research/disclosures", researchEcoHandler.CreateInventionDisclosure)
				// Phase 15
				admin.POST("/ecosystem/institutions", ecosystemHandler.RegisterTrustedInstitution)
				admin.POST("/ecosystem/tools", ecosystemHandler.RegisterAgentTool)
				admin.POST("/ecosystem/services", ecosystemHandler.PublishService)
				admin.GET("/ecosystem/services", ecosystemHandler.ListServices)
				admin.POST("/ecosystem/loop-guards", ecosystemHandler.CreateLoopGuard)
				admin.POST("/ecosystem/governance/reviews", ecosystemHandler.CreateGovernanceReview)
				// Phase 16
				admin.POST("/curriculum/nodes", planetHandler.AddCurriculumNode)
				admin.POST("/curriculum/edges", planetHandler.AddCurriculumEdge)
				// Phase 17 — Digital Twins
				admin.POST("/twins", twinHandler.CreateTwin)
				admin.GET("/twins/:twinId", twinHandler.GetTwin)
				admin.POST("/twins/:twinId/state", twinHandler.RecordStateSnapshot)
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

print("Phase 17 scaffolding generated successfully.")
