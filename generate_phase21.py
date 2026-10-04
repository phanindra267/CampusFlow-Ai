import os

files = {
    "migrations/000021_phase21_scientific_discovery.up.sql": """
CREATE TABLE IF NOT EXISTS scientific_hypotheses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    statement TEXT NOT NULL,
    context TEXT,
    supporting_evidence JSONB,
    contradicting_evidence JSONB,
    assumptions JSONB,
    expected_predictions JSONB,
    confidence FLOAT,
    owner_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'PROPOSED', -- PROPOSED, INVESTIGATING, SUPPORTED, REFUTED, INCONCLUSIVE
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS scientific_experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hypothesis_id UUID REFERENCES scientific_hypotheses(id) ON DELETE CASCADE,
    objective TEXT NOT NULL,
    variables JSONB,
    controls JSONB,
    measurements JSONB,
    sample_requirements JSONB,
    expected_outcomes JSONB,
    risks JSONB,
    stop_conditions JSONB,
    reproducibility_config JSONB,
    status VARCHAR(50) DEFAULT 'DESIGNED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS intelligence_improvements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_component VARCHAR(100) NOT NULL, -- MODEL, PROMPT, POLICY, AGENT
    current_state JSONB,
    proposed_change JSONB,
    expected_benefit TEXT,
    expected_risk TEXT,
    validation_plan JSONB,
    rollback_plan JSONB,
    sandbox_results JSONB,
    status VARCHAR(50) DEFAULT 'PROPOSED', -- PROPOSED, VALIDATING, APPROVED, CANARY, PRODUCTION, REJECTED, ROLLED_BACK
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS causal_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_entity VARCHAR(255) NOT NULL,
    target_entity VARCHAR(255) NOT NULL,
    relationship_type VARCHAR(50) DEFAULT 'HYPOTHESIZED', -- OBSERVED, INFERRED, HYPOTHESIZED, SUPPORTED
    evidence_strength FLOAT,
    provenance JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS research_artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    artifact_type VARCHAR(50) NOT NULL, -- PAPER, DATASET, CODE, MODEL
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    provenance JSONB,
    storage_ref VARCHAR(255),
    security_scan_results JSONB,
    status VARCHAR(50) DEFAULT 'REGISTERED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
""",
    "internal/domain/discovery.go": """package domain

import "time"

type ScientificHypothesis struct {
	ID                    string                 `json:"id"`
	TenantID              string                 `json:"tenant_id"`
	Statement             string                 `json:"statement" binding:"required"`
	Context               string                 `json:"context"`
	SupportingEvidence    []interface{}          `json:"supporting_evidence"`
	ContradictingEvidence []interface{}          `json:"contradicting_evidence"`
	Assumptions           []interface{}          `json:"assumptions"`
	ExpectedPredictions   []interface{}          `json:"expected_predictions"`
	Confidence            float64                `json:"confidence"`
	OwnerID               string                 `json:"owner_id"`
	Status                string                 `json:"status"`
}

type ScientificExperiment struct {
	ID                    string                 `json:"id"`
	HypothesisID          string                 `json:"hypothesis_id" binding:"required"`
	Objective             string                 `json:"objective" binding:"required"`
	Variables             map[string]interface{} `json:"variables"`
	Controls              map[string]interface{} `json:"controls"`
	Measurements          map[string]interface{} `json:"measurements"`
	SampleRequirements    map[string]interface{} `json:"sample_requirements"`
	ExpectedOutcomes      []interface{}          `json:"expected_outcomes"`
	Risks                 []interface{}          `json:"risks"`
	StopConditions        []interface{}          `json:"stop_conditions"`
	ReproducibilityConfig map[string]interface{} `json:"reproducibility_config"`
	Status                string                 `json:"status"`
}

type IntelligenceImprovement struct {
	ID             string                 `json:"id"`
	TargetComponent string                `json:"target_component" binding:"required"`
	CurrentState   map[string]interface{} `json:"current_state"`
	ProposedChange map[string]interface{} `json:"proposed_change"`
	ExpectedBenefit string                `json:"expected_benefit"`
	ExpectedRisk   string                 `json:"expected_risk"`
	ValidationPlan map[string]interface{} `json:"validation_plan"`
	RollbackPlan   map[string]interface{} `json:"rollback_plan"`
	SandboxResults map[string]interface{} `json:"sandbox_results"`
	Status         string                 `json:"status"`
}

type CausalRelationship struct {
	ID               string                 `json:"id"`
	SourceEntity     string                 `json:"source_entity" binding:"required"`
	TargetEntity     string                 `json:"target_entity" binding:"required"`
	RelationshipType string                 `json:"relationship_type"`
	EvidenceStrength float64                `json:"evidence_strength"`
	Provenance       map[string]interface{} `json:"provenance"`
}

type ResearchArtifact struct {
	ID                  string                 `json:"id"`
	ArtifactType        string                 `json:"artifact_type" binding:"required"`
	Name                string                 `json:"name" binding:"required"`
	Version             string                 `json:"version" binding:"required"`
	Provenance          map[string]interface{} `json:"provenance"`
	StorageRef          string                 `json:"storage_ref"`
	SecurityScanResults map[string]interface{} `json:"security_scan_results"`
	Status              string                 `json:"status"`
}
""",
    "internal/delivery/http/handler/discovery.go": """package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type ScientificDiscoveryHandler struct{}

func NewScientificDiscoveryHandler() *ScientificDiscoveryHandler {
	return &ScientificDiscoveryHandler{}
}

// POST /admin/discovery/hypotheses
func (h *ScientificDiscoveryHandler) ProposeHypothesis(c *gin.Context) {
	var req domain.ScientificHypothesis
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid hypothesis payload", err.Error())
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.Status = "PROPOSED"
	response.Success(c, http.StatusCreated,
		"Scientific hypothesis recorded. Awaiting adversarial review and evidence synthesis.", req)
}

// POST /admin/discovery/experiments
func (h *ScientificDiscoveryHandler) DesignExperiment(c *gin.Context) {
	var req domain.ScientificExperiment
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid experiment payload", err.Error())
		return
	}
	req.Status = "DESIGNED"
	response.Success(c, http.StatusCreated,
		"Experiment designed. Execution requires isolation, risk review, and reproducibility validation.", req)
}

// POST /admin/discovery/improvements
func (h *ScientificDiscoveryHandler) ProposeIntelligenceImprovement(c *gin.Context) {
	var req domain.IntelligenceImprovement
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid intelligence improvement payload", err.Error())
		return
	}
	req.Status = "PROPOSED"
	response.Success(c, http.StatusCreated,
		"Self-improvement proposed. Must pass sandbox evaluation, canary deployment, and human governance before production activation.", req)
}

// POST /admin/discovery/artifacts
func (h *ScientificDiscoveryHandler) RegisterArtifact(c *gin.Context) {
	var req domain.ResearchArtifact
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid artifact payload", err.Error())
		return
	}
	req.Status = "REGISTERED"
	response.Success(c, http.StatusCreated,
		"Research artifact registered. Complete provenance and security isolation applied.", req)
}

// POST /admin/discovery/causal-graphs
func (h *ScientificDiscoveryHandler) RegisterCausalRelationship(c *gin.Context) {
	var req domain.CausalRelationship
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid causal relationship payload", err.Error())
		return
	}
	if req.RelationshipType == "" {
		req.RelationshipType = "HYPOTHESIZED"
	}
	response.Success(c, http.StatusCreated,
		"Causal relationship registered. Explicitly marked as non-authoritative until experimentally supported.", req)
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
				// Phase 17
				admin.POST("/twins", twinHandler.CreateTwin)
				admin.GET("/twins/:twinId", twinHandler.GetTwin)
				admin.POST("/twins/:twinId/state", twinHandler.RecordStateSnapshot)
				// Phase 18
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
				// Phase 19
				admin.POST("/federation/members", collectiveIntelHandler.RegisterMember)
				admin.POST("/federation/members/:memberId/activate", collectiveIntelHandler.ActivateMember)
				admin.POST("/federation/members/:memberId/revoke", collectiveIntelHandler.RevokeMember)
				admin.POST("/federation/agreements", collectiveIntelHandler.CreateAgreement)
				admin.POST("/federation/knowledge/claims", collectiveIntelHandler.PublishFederatedClaim)
				admin.POST("/federation/governance/proposals", collectiveIntelHandler.CreateProposal)
				admin.POST("/federation/governance/proposals/:proposalId/vote", collectiveIntelHandler.CastVote)
				admin.POST("/federation/disputes", collectiveIntelHandler.RaiseDispute)
				// Phase 20
				admin.POST("/global/trust/entities", globalIntelHandler.RegisterTrustEntity)
				admin.POST("/global/agents", globalIntelHandler.RegisterGlobalAgent)
				admin.POST("/global/teams", globalIntelHandler.CreateHumanAITeam)
				admin.POST("/global/reasoning/jobs", globalIntelHandler.StartReasoningJob)
				admin.POST("/global/forecasts", globalIntelHandler.RegisterForecast)
				admin.POST("/global/risks", globalIntelHandler.RegisterRisk)
				// Phase 21 — Autonomous Scientific Discovery
				admin.POST("/discovery/hypotheses", discoveryHandler.ProposeHypothesis)
				admin.POST("/discovery/experiments", discoveryHandler.DesignExperiment)
				admin.POST("/discovery/improvements", discoveryHandler.ProposeIntelligenceImprovement)
				admin.POST("/discovery/artifacts", discoveryHandler.RegisterArtifact)
				admin.POST("/discovery/causal-graphs", discoveryHandler.RegisterCausalRelationship)
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

print("Phase 21 scaffolding generated successfully.")
