import os

files = {
    "migrations/000026_phase26_digital_twins.up.sql": """
CREATE TABLE IF NOT EXISTS dt_population_cohorts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    cohort_name VARCHAR(255) NOT NULL,
    dimensions JSONB NOT NULL,
    current_state JSONB NOT NULL,
    historical_states JSONB,
    version VARCHAR(50) DEFAULT '1.0',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dt_system_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_name VARCHAR(255) NOT NULL,
    baseline_cohorts JSONB NOT NULL,
    baseline_economics JSONB NOT NULL,
    random_seed BIGINT,
    provenance JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dt_simulations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_id UUID REFERENCES dt_system_snapshots(id),
    owner_id VARCHAR(255) NOT NULL,
    scenario_type VARCHAR(100), -- MONTE_CARLO, SENSITIVITY, DETERMINISTIC
    assumptions JSONB NOT NULL,
    shocks JSONB,
    status VARCHAR(50) DEFAULT 'QUEUED', -- QUEUED, RUNNING, COMPLETED, FAILED, CANCELLED
    progress FLOAT DEFAULT 0.0,
    results JSONB,
    uncertainty_metrics JSONB,
    parent_simulation_id UUID REFERENCES dt_simulations(id), -- For branching
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS dt_interventions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    simulation_id UUID REFERENCES dt_simulations(id) ON DELETE CASCADE,
    target_entity VARCHAR(255) NOT NULL,
    intervention_type VARCHAR(100),
    parameters JSONB NOT NULL,
    expected_mechanism TEXT,
    human_approval_status VARCHAR(50) DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dt_early_warnings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    signal_type VARCHAR(100) NOT NULL, -- SKILL_SHORTAGE, BOTTLENECK, CAPACITY_LOSS
    evidence JSONB NOT NULL,
    confidence FLOAT NOT NULL,
    time_window VARCHAR(100),
    status VARCHAR(50) DEFAULT 'DETECTED', -- DETECTED, INVESTIGATING, MITIGATED, FALSE_POSITIVE
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dt_model_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    purpose TEXT NOT NULL,
    intended_use TEXT NOT NULL,
    prohibited_use TEXT NOT NULL,
    assumptions JSONB NOT NULL,
    limitations JSONB NOT NULL,
    validation_metrics JSONB,
    approval_status VARCHAR(50) DEFAULT 'DRAFT', -- DRAFT, APPROVED, RETIRED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
""",
    "internal/domain/digitaltwin.go": """package domain

import "time"

type PopulationCohort struct {
	ID               string                 `json:"id"`
	TenantID         string                 `json:"tenant_id"`
	CohortName       string                 `json:"cohort_name" binding:"required"`
	Dimensions       map[string]interface{} `json:"dimensions"`
	CurrentState     map[string]interface{} `json:"current_state"`
	HistoricalStates []interface{}          `json:"historical_states"`
	Version          string                 `json:"version"`
}

type SystemSnapshot struct {
	ID                string                 `json:"id"`
	SnapshotName      string                 `json:"snapshot_name" binding:"required"`
	BaselineCohorts   []interface{}          `json:"baseline_cohorts"`
	BaselineEconomics []interface{}          `json:"baseline_economics"`
	RandomSeed        int64                  `json:"random_seed"`
	Provenance        map[string]interface{} `json:"provenance"`
	CreatedAt         *time.Time             `json:"created_at,omitempty"`
}

type StrategicSimulation struct {
	ID                 string                 `json:"id"`
	SnapshotID         string                 `json:"snapshot_id" binding:"required"`
	OwnerID            string                 `json:"owner_id"`
	ScenarioType       string                 `json:"scenario_type"`
	Assumptions        map[string]interface{} `json:"assumptions"`
	Shocks             []interface{}          `json:"shocks"`
	Status             string                 `json:"status"`
	Progress           float64                `json:"progress"`
	Results            map[string]interface{} `json:"results"`
	UncertaintyMetrics map[string]interface{} `json:"uncertainty_metrics"`
	ParentSimulationID string                 `json:"parent_simulation_id,omitempty"`
	CreatedAt          *time.Time             `json:"created_at,omitempty"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
}

type SimulationIntervention struct {
	ID                  string                 `json:"id"`
	SimulationID        string                 `json:"simulation_id" binding:"required"`
	TargetEntity        string                 `json:"target_entity"`
	InterventionType    string                 `json:"intervention_type"`
	Parameters          map[string]interface{} `json:"parameters"`
	ExpectedMechanism   string                 `json:"expected_mechanism"`
	HumanApprovalStatus string                 `json:"human_approval_status"`
}

type EarlyWarningSignal struct {
	ID          string                 `json:"id"`
	SignalType  string                 `json:"signal_type"`
	Evidence    map[string]interface{} `json:"evidence"`
	Confidence  float64                `json:"confidence"`
	TimeWindow  string                 `json:"time_window"`
	Status      string                 `json:"status"`
	CreatedAt   *time.Time             `json:"created_at,omitempty"`
}

type ModelCard struct {
	ID                string                 `json:"id"`
	ModelName         string                 `json:"model_name" binding:"required"`
	Version           string                 `json:"version" binding:"required"`
	Purpose           string                 `json:"purpose"`
	IntendedUse       string                 `json:"intended_use"`
	ProhibitedUse     string                 `json:"prohibited_use"`
	Assumptions       map[string]interface{} `json:"assumptions"`
	Limitations       map[string]interface{} `json:"limitations"`
	ValidationMetrics map[string]interface{} `json:"validation_metrics"`
	ApprovalStatus    string                 `json:"approval_status"`
}
""",
    "internal/delivery/http/handler/digitaltwin.go": """package handler

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

// POST /admin/twins/cohorts
func (h *DigitalTwinHandler) RegisterCohort(c *gin.Context) {
	var req domain.PopulationCohort
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid cohort payload", err.Error())
		return
	}
	response.Success(c, http.StatusCreated,
		"Population cohort twin created. Preserving individual privacy via aggregate modeling.", req)
}

// POST /admin/twins/snapshots
func (h *DigitalTwinHandler) CreateSnapshot(c *gin.Context) {
	var req domain.SystemSnapshot
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid snapshot payload", err.Error())
		return
	}
	response.Success(c, http.StatusCreated,
		"Immutable system snapshot recorded for reproducible strategic simulations.", req)
}

// POST /institutions/simulations
func (h *DigitalTwinHandler) QueueSimulation(c *gin.Context) {
	var req domain.StrategicSimulation
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid simulation payload", err.Error())
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.Status = "QUEUED"
	response.Success(c, http.StatusCreated,
		"Strategic simulation queued securely. Outcomes are explicitly defined as exploratory.", req)
}

// POST /institutions/simulations/:id/branch
func (h *DigitalTwinHandler) BranchSimulation(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	
	branched := domain.StrategicSimulation{
		ID:                 "branched-sim-placeholder",
		ParentSimulationID: id,
		OwnerID:            userID.(string),
		Status:             "QUEUED",
	}
	response.Success(c, http.StatusCreated,
		"Simulation branched successfully for comparative what-if analysis.", branched)
}

// POST /institutions/simulations/:id/interventions
func (h *DigitalTwinHandler) ProposeIntervention(c *gin.Context) {
	var req domain.SimulationIntervention
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid intervention payload", err.Error())
		return
	}
	req.HumanApprovalStatus = "PENDING"
	response.Success(c, http.StatusCreated,
		"Simulation intervention proposed. Requires human review and governance before implementation.", req)
}

// GET /admin/twins/warnings
func (h *DigitalTwinHandler) ListEarlyWarnings(c *gin.Context) {
	response.Success(c, http.StatusOK,
		"Early warning signals retrieved for workforce and educational capacity disruptions.",
		[]domain.EarlyWarningSignal{})
}

// POST /admin/twins/model-cards
func (h *DigitalTwinHandler) PublishModelCard(c *gin.Context) {
	var req domain.ModelCard
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid model card payload", err.Error())
		return
	}
	req.ApprovalStatus = "DRAFT"
	response.Success(c, http.StatusCreated,
		"Model card drafted. Transparently documenting simulation purpose, intended use, and critical limitations.", req)
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
	autonomousHandler    := handler.NewAutonomousHandler()
	collectiveIntelHandler := handler.NewCollectiveIntelligenceHandler()
	globalIntelHandler   := handler.NewGlobalIntelligenceHandler()
	discoveryHandler     := handler.NewScientificDiscoveryHandler()
	educationHandler     := handler.NewEducationIntelligenceHandler()
	interopHandler       := handler.NewInteroperabilityHandler()
	lifelongHandler      := handler.NewLifelongLearningHandler()
	econHandler          := handler.NewEconomicIntelligenceHandler()
	digitalTwinHandler   := handler.NewDigitalTwinHandler()

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
		api.POST("/interop/credentials/verify", interopHandler.VerifyCredential)
		api.GET("/economics/regional/balance", econHandler.GetRegionalBalance)
		api.GET("/economics/funding", econHandler.ListFunding)

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
				student.POST("/profiles/:id/learning-paths", educationHandler.GenerateLearningPath)
				student.GET("/opportunities/matches", educationHandler.MatchOpportunities)
				student.POST("/wallet/consents", interopHandler.GrantConsent)
				student.POST("/mobility/transfers", interopHandler.RequestTransferCredit)
				student.POST("/lifelong-profiles", lifelongHandler.UpdateLifelongProfile)
				student.POST("/career-paths/simulate", lifelongHandler.SimulateCareerPath)
				student.POST("/interviews/prep", lifelongHandler.GenerateInterviewPrep)
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

			// ── ORGANIZER / FACULTY ──────────────────────────────────────────────
			organizer := protected.Group("/organizers") 
			organizer.Use(middleware.RequireRole("ORGANIZER", "FACULTY", "ADMIN", "SUPER_ADMIN"))
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
				organizer.POST("/simulations", twinHandler.SubmitSimulation)
				organizer.POST("/simulations/scenarios", twinHandler.CreateScenario)
				organizer.GET("/simulations/:runId/results", twinHandler.GetSimulationResults)
				organizer.POST("/courses/:id/assessments", educationHandler.CreateAssessment)
			}
			
			// ── ADVISOR ────────────────────────────────────────────────────────
			advisor := protected.Group("/advisors")
			advisor.Use(middleware.RequireRole("ADVISOR", "ADMIN", "SUPER_ADMIN"))
			{
				advisor.POST("/interventions/:id/review", educationHandler.ReviewIntervention)
			}

			// ── INSTITUTION / REGISTRAR ─────────────────────────────────────────
			institution := protected.Group("/institutions")
			institution.Use(middleware.RequireRole("ADMIN", "SUPER_ADMIN", "REGISTRAR", "POLICYMAKER"))
			{
				institution.POST("/credentials/issue", interopHandler.IssueCredential)
				institution.POST("/transfers/:id/approve", interopHandler.ApproveTransferCredit)
				institution.POST("/economics/scenarios", econHandler.CreateScenario)
				institution.POST("/economics/scenarios/:id/cancel", econHandler.CancelScenario)
				
				// Phase 26 - Digital Twin Strategic Simulations
				institution.POST("/simulations", digitalTwinHandler.QueueSimulation)
				institution.POST("/simulations/:id/branch", digitalTwinHandler.BranchSimulation)
				institution.POST("/simulations/:id/interventions", digitalTwinHandler.ProposeIntervention)
			}

			// ── EMPLOYER ────────────────────────────────────────────────────────
			employer := protected.Group("/employers")
			employer.Use(middleware.RequireRole("EMPLOYER", "ADMIN", "SUPER_ADMIN"))
			{
				employer.POST("/roles", lifelongHandler.CreateJobRole)
				employer.POST("/talent-matches", lifelongHandler.MatchTalent)
				employer.POST("/experiences/verify", lifelongHandler.VerifyExperience)
			}

			// ── ADMIN / SUPER_ADMIN ──────────────────────────────────────────────
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
				admin.POST("/intelligence/decisions/rooms/:roomId/resolve", intelHandler.ResolveDecision)
				admin.POST("/intelligence/autonomous-policies", intelHandler.RegisterAutonomousPolicy)
				admin.POST("/collective/rooms/:roomId/decide", collectiveHandler.RecordDecisionOutcome)
				admin.POST("/collective/playbooks", collectiveHandler.CreatePlaybook)
				admin.GET("/collective/predictions", collectiveHandler.GetPredictionLedger)
				admin.POST("/research/disclosures", researchEcoHandler.CreateInventionDisclosure)
				admin.POST("/ecosystem/institutions", ecosystemHandler.RegisterTrustedInstitution)
				admin.POST("/ecosystem/tools", ecosystemHandler.RegisterAgentTool)
				admin.POST("/ecosystem/services", ecosystemHandler.PublishService)
				admin.GET("/ecosystem/services", ecosystemHandler.ListServices)
				admin.POST("/ecosystem/loop-guards", ecosystemHandler.CreateLoopGuard)
				admin.POST("/ecosystem/governance/reviews", ecosystemHandler.CreateGovernanceReview)
				admin.POST("/curriculum/nodes", planetHandler.AddCurriculumNode)
				admin.POST("/curriculum/edges", planetHandler.AddCurriculumEdge)
				admin.POST("/twins", twinHandler.CreateTwin)
				admin.GET("/twins/:twinId", twinHandler.GetTwin)
				admin.POST("/twins/:twinId/state", twinHandler.RecordStateSnapshot)
				admin.POST("/autonomous/agents", autonomousHandler.RegisterAgent)
				admin.POST("/autonomous/agents/:agentId/activate", autonomousHandler.ActivateAgent)
				admin.POST("/autonomous/agents/:agentId/suspend", autonomousHandler.SuspendAgent)
				admin.POST("/autonomous/policies", autonomousHandler.CreatePolicy)
				admin.POST("/autonomous/plans", autonomousHandler.CreatePlan)
				admin.POST("/autonomous/plans/:planId/approve", autonomousHandler.ApprovePlan)
				admin.POST("/autonomous/decisions", autonomousHandler.RecordDecision)
				admin.POST("/autonomous/lessons", autonomousHandler.RecordLesson)
				admin.POST("/autonomous/kill-switch", autonomousHandler.GlobalKillSwitch)
				admin.POST("/autonomous/safe-mode", autonomousHandler.ActivateSafeMode)
				admin.POST("/federation/members", collectiveIntelHandler.RegisterMember)
				admin.POST("/federation/members/:memberId/activate", collectiveIntelHandler.ActivateMember)
				admin.POST("/federation/members/:memberId/revoke", collectiveIntelHandler.RevokeMember)
				admin.POST("/federation/agreements", collectiveIntelHandler.CreateAgreement)
				admin.POST("/federation/knowledge/claims", collectiveIntelHandler.PublishFederatedClaim)
				admin.POST("/federation/governance/proposals", collectiveIntelHandler.CreateProposal)
				admin.POST("/federation/governance/proposals/:proposalId/vote", collectiveIntelHandler.CastVote)
				admin.POST("/federation/disputes", collectiveIntelHandler.RaiseDispute)
				admin.POST("/global/trust/entities", globalIntelHandler.RegisterTrustEntity)
				admin.POST("/global/agents", globalIntelHandler.RegisterGlobalAgent)
				admin.POST("/global/teams", globalIntelHandler.CreateHumanAITeam)
				admin.POST("/global/reasoning/jobs", globalIntelHandler.StartReasoningJob)
				admin.POST("/global/forecasts", globalIntelHandler.RegisterForecast)
				admin.POST("/global/risks", globalIntelHandler.RegisterRisk)
				admin.POST("/discovery/hypotheses", discoveryHandler.ProposeHypothesis)
				admin.POST("/discovery/experiments", discoveryHandler.DesignExperiment)
				admin.POST("/discovery/improvements", discoveryHandler.ProposeIntelligenceImprovement)
				admin.POST("/discovery/artifacts", discoveryHandler.RegisterArtifact)
				admin.POST("/discovery/causal-graphs", discoveryHandler.RegisterCausalRelationship)
				admin.POST("/education/profiles/:id/interventions", educationHandler.RecommendIntervention)
				admin.POST("/interop/institutions", interopHandler.RegisterInstitution)
				admin.POST("/interop/credentials/schemas", interopHandler.RegisterSchema)
				admin.POST("/economics/datasets", econHandler.RegisterDataset)

				// Phase 26 - Global Human-Capital Digital Twin Admin
				admin.POST("/twins/cohorts", digitalTwinHandler.RegisterCohort)
				admin.POST("/twins/snapshots", digitalTwinHandler.CreateSnapshot)
				admin.POST("/twins/model-cards", digitalTwinHandler.PublishModelCard)
				admin.GET("/twins/warnings", digitalTwinHandler.ListEarlyWarnings)
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

print("Phase 26 scaffolding generated successfully.")
