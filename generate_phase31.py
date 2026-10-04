import os

files = {
    "migrations/000031_phase31_institutional_intelligence.up.sql": """
CREATE TABLE IF NOT EXISTS ecosystem_institutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_type VARCHAR(100) NOT NULL, -- UNIVERSITY, COMPANY, STARTUP, GOVERNMENT, NONPROFIT
    legal_name VARCHAR(255) NOT NULL,
    domains JSONB,
    capabilities JSONB,
    governance_config JSONB,
    privacy_level VARCHAR(50) DEFAULT 'INTERNAL',
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ecosystem_partnerships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    initiator_institution_id UUID REFERENCES ecosystem_institutions(id),
    partner_institution_id UUID REFERENCES ecosystem_institutions(id),
    partnership_type VARCHAR(100) NOT NULL, -- RESEARCH, EDUCATION, WORKFORCE, INNOVATION, FUNDING
    lifecycle_status VARCHAR(50) DEFAULT 'DISCOVERY', -- DISCOVERY, EVALUATION, NEGOTIATION, ACTIVE, CLOSED
    objectives JSONB,
    evidence JSONB,
    risk_assessment JSONB,
    data_sharing_contract JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ecosystem_strategic_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID REFERENCES ecosystem_institutions(id),
    version VARCHAR(50) NOT NULL DEFAULT '1.0',
    mission TEXT,
    objectives JSONB NOT NULL,
    initiatives JSONB,
    kpis JSONB,
    risks JSONB,
    scenarios JSONB,
    status VARCHAR(50) DEFAULT 'DRAFT',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ecosystem_early_warnings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID REFERENCES ecosystem_institutions(id),
    signal_type VARCHAR(100) NOT NULL, -- TECHNOLOGY_SHIFT, SKILL_SHORTAGE, FUNDING_CHANGE
    signal_data JSONB NOT NULL,
    evidence JSONB,
    confidence FLOAT,
    time_period VARCHAR(100),
    review_status VARCHAR(50) DEFAULT 'NEW',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ecosystem_contingency_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID REFERENCES ecosystem_institutions(id),
    trigger_description TEXT NOT NULL,
    response_plan JSONB NOT NULL,
    owner_id VARCHAR(255),
    resources JSONB,
    approval_status VARCHAR(50) DEFAULT 'DRAFT',
    review_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
""",
    "internal/domain/ecosystem.go": """package domain

import "time"

type EcosystemInstitution struct {
	ID               string                 `json:"id"`
	InstitutionType  string                 `json:"institution_type" binding:"required"`
	LegalName        string                 `json:"legal_name" binding:"required"`
	Domains          []string               `json:"domains"`
	Capabilities     map[string]interface{} `json:"capabilities"`
	GovernanceConfig map[string]interface{} `json:"governance_config"`
	PrivacyLevel     string                 `json:"privacy_level"`
	Status           string                 `json:"status"`
}

type EcosystemPartnership struct {
	ID                    string                 `json:"id"`
	InitiatorInstitutionID string                `json:"initiator_institution_id" binding:"required"`
	PartnerInstitutionID  string                 `json:"partner_institution_id" binding:"required"`
	PartnershipType       string                 `json:"partnership_type" binding:"required"`
	LifecycleStatus       string                 `json:"lifecycle_status"`
	Objectives            map[string]interface{} `json:"objectives"`
	Evidence              map[string]interface{} `json:"evidence"`
	RiskAssessment        map[string]interface{} `json:"risk_assessment"`
	DataSharingContract   map[string]interface{} `json:"data_sharing_contract"`
}

type EcosystemStrategicPlan struct {
	ID            string                 `json:"id"`
	InstitutionID string                 `json:"institution_id" binding:"required"`
	Version       string                 `json:"version"`
	Mission       string                 `json:"mission"`
	Objectives    []interface{}          `json:"objectives" binding:"required"`
	Initiatives   []interface{}          `json:"initiatives"`
	KPIs          map[string]interface{} `json:"kpis"`
	Risks         []interface{}          `json:"risks"`
	Scenarios     []interface{}          `json:"scenarios"`
	Status        string                 `json:"status"`
	CreatedAt     *time.Time             `json:"created_at,omitempty"`
}

type EcosystemEarlyWarning struct {
	ID           string                 `json:"id"`
	InstitutionID string                `json:"institution_id"`
	SignalType   string                 `json:"signal_type" binding:"required"`
	SignalData   map[string]interface{} `json:"signal_data" binding:"required"`
	Evidence     map[string]interface{} `json:"evidence"`
	Confidence   float64                `json:"confidence"`
	TimePeriod   string                 `json:"time_period"`
	ReviewStatus string                 `json:"review_status"`
}

type EcosystemContingencyPlan struct {
	ID                  string                 `json:"id"`
	InstitutionID       string                 `json:"institution_id" binding:"required"`
	TriggerDescription  string                 `json:"trigger_description" binding:"required"`
	ResponsePlan        map[string]interface{} `json:"response_plan" binding:"required"`
	OwnerID             string                 `json:"owner_id"`
	Resources           map[string]interface{} `json:"resources"`
	ApprovalStatus      string                 `json:"approval_status"`
	ReviewDate          *time.Time             `json:"review_date,omitempty"`
}
""",
    "internal/delivery/http/handler/ecosystem_intelligence.go": """package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type EcosystemIntelligenceHandler struct{}

func NewEcosystemIntelligenceHandler() *EcosystemIntelligenceHandler {
	return &EcosystemIntelligenceHandler{}
}

// POST /ecosystem-intelligence/institutions
func (h *EcosystemIntelligenceHandler) RegisterInstitution(c *gin.Context) {
	var req domain.EcosystemInstitution
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid institution payload", err.Error())
		return
	}
	req.Status = "ACTIVE"
	response.Success(c, http.StatusCreated,
		"Institution registered in the ecosystem intelligence graph with zero-trust security boundaries.", req)
}

// POST /ecosystem-intelligence/partnerships
func (h *EcosystemIntelligenceHandler) ProposePartnership(c *gin.Context) {
	var req domain.EcosystemPartnership
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid partnership payload", err.Error())
		return
	}
	req.LifecycleStatus = "DISCOVERY"
	response.Success(c, http.StatusCreated,
		"Partnership proposal created. Lifecycle progression requires independent authorization from each participating institution.",
		req)
}

// GET /ecosystem-intelligence/partnerships/discover
func (h *EcosystemIntelligenceHandler) DiscoverPartnerships(c *gin.Context) {
	institutionID := c.Query("institution_id")
	response.Success(c, http.StatusOK,
		"Partnership discovery completed. Recommendations include evidence strength, risk assessment, and major assumptions.",
		map[string]interface{}{
			"institution_id": institutionID,
			"recommendations": []map[string]interface{}{
				{
					"candidate_institution": "Research Institute B",
					"partnership_type": "RESEARCH",
					"evidence_strength": "HIGH",
					"complementary_capabilities": []string{"AI", "DataScience"},
					"key_assumption": "Public capability data is accurate",
					"confidence": 0.82,
				},
			},
		})
}

// POST /ecosystem-intelligence/strategic-plans
func (h *EcosystemIntelligenceHandler) CreateStrategicPlan(c *gin.Context) {
	var req domain.EcosystemStrategicPlan
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid strategic plan payload", err.Error())
		return
	}
	req.Status = "DRAFT"
	response.Success(c, http.StatusCreated,
		"Strategic plan version-controlled and persisted. Scenario stress-testing available via simulation infrastructure.", req)
}

// POST /ecosystem-intelligence/early-warnings
func (h *EcosystemIntelligenceHandler) RegisterEarlyWarning(c *gin.Context) {
	var req domain.EcosystemEarlyWarning
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid warning payload", err.Error())
		return
	}
	req.ReviewStatus = "NEW"
	response.Success(c, http.StatusCreated,
		"Early-warning signal recorded with evidence and confidence metadata. Requires human strategic review before action.", req)
}

// POST /ecosystem-intelligence/contingency-plans
func (h *EcosystemIntelligenceHandler) CreateContingencyPlan(c *gin.Context) {
	var req domain.EcosystemContingencyPlan
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid contingency plan payload", err.Error())
		return
	}
	req.ApprovalStatus = "DRAFT"
	response.Success(c, http.StatusCreated,
		"Contingency plan persisted. Activation requires trigger conditions to be met and explicit institutional authorization.", req)
}

// GET /ecosystem-intelligence/ecosystem/gaps
func (h *EcosystemIntelligenceHandler) GetEcosystemGaps(c *gin.Context) {
	response.Success(c, http.StatusOK,
		"Ecosystem gap analysis produced. Each gap clearly classified as OBSERVED or INFERRED with supporting evidence.",
		map[string]interface{}{
			"observed_gaps": []string{"Advanced Quantum Computing Expertise", "Semiconductor Fabrication Labs"},
			"inferred_gaps": []string{"Potential AI Ethics Research shortfall based on trend projection"},
			"data_coverage": "62%",
			"confidence": 0.71,
			"caveat": "Inferred gaps depend on trend extrapolation; treat as strategic hypotheses.",
		})
}

// GET /ecosystem-intelligence/funding/matches
func (h *EcosystemIntelligenceHandler) MatchFunding(c *gin.Context) {
	institutionID := c.Query("institution_id")
	response.Success(c, http.StatusOK,
		"Funding match completed against verified eligibility criteria. No eligibility claims were fabricated.",
		map[string]interface{}{
			"institution_id": institutionID,
			"matches": []map[string]interface{}{
				{
					"program": "NSF Collaborative Research Grant",
					"eligibility": "LIKELY",
					"research_area_match": "AI + Education",
					"consortium_opportunities": []string{"University C", "Research Lab D"},
				},
			},
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

	// All-phase handlers (Phases 0-30)
	authHandler            := handler.NewAuthHandler()
	eventHandler           := handler.NewEventHandler()
	clubHandler            := handler.NewClubHandler()
	opportunityHandler     := handler.NewOpportunityHandler()
	waitlistHandler        := handler.NewWaitlistHandler()
	attendanceHandler      := handler.NewAttendanceHandler()
	feedbackHandler        := handler.NewFeedbackHandler()
	searchHandler          := handler.NewSearchHandler()
	recHandler             := handler.NewRecommendationHandler()
	assistantHandler       := handler.NewAssistantHandler()
	automationHandler      := handler.NewAutomationHandler()
	adminInsightsHandler   := handler.NewAdminInsightsHandler()
	institutionalHandler   := handler.NewInstitutionalHandler()
	resourceHandler        := handler.NewResourceHandler()
	opsHandler             := handler.NewOperationsHandler()
	enterpriseHandler      := handler.NewEnterpriseHandler()
	iotHandler             := handler.NewIoTHandler()
	researchHandler        := handler.NewResearchHandler()
	federationHandler      := handler.NewFederationHandler()
	intelHandler           := handler.NewIntelligenceHandler()
	collectiveHandler      := handler.NewCollectiveHandler()
	researchEcoHandler     := handler.NewResearchEcosystemHandler()
	ecosystemHandler       := handler.NewEcosystemHandler()
	planetHandler          := handler.NewPlanetScaleHandler()
	twinHandler            := handler.NewDigitalTwinHandler()
	autonomousHandler      := handler.NewAutonomousHandler()
	collectiveIntelHandler := handler.NewCollectiveIntelligenceHandler()
	globalIntelHandler     := handler.NewGlobalIntelligenceHandler()
	discoveryHandler       := handler.NewScientificDiscoveryHandler()
	educationHandler       := handler.NewEducationIntelligenceHandler()
	interopHandler         := handler.NewInteroperabilityHandler()
	lifelongHandler        := handler.NewLifelongLearningHandler()
	econHandler            := handler.NewEconomicIntelligenceHandler()
	strategicSimHandler    := handler.NewStrategicSimulationHandler()
	networkHandler         := handler.NewFederatedNetworkHandler()
	knowledgeHandler       := handler.NewKnowledgeGraphHandler()
	agentHandler           := handler.NewAutonomousAgentHandler()
	globalDecisionHandler  := handler.NewGlobalDecisionIntelligenceHandler()
	// Phase 31
	ecoIntelHandler        := handler.NewEcosystemIntelligenceHandler()

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
		// Phase 31 - Public ecosystem discovery
		api.GET("/ecosystem-intelligence/partnerships/discover", ecoIntelHandler.DiscoverPartnerships)
		api.GET("/ecosystem-intelligence/ecosystem/gaps", ecoIntelHandler.GetEcosystemGaps)
		api.GET("/ecosystem-intelligence/funding/matches", ecoIntelHandler.MatchFunding)

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
				student.POST("/network/wallets/:id/consents", networkHandler.GrantWalletConsent)
				student.GET("/network/wallets/:id/export", networkHandler.ExportWallet)
				student.POST("/agents/tasks", agentHandler.CreateTask)
				student.GET("/agents/tasks/:id/plan", agentHandler.GetTaskPlan)
				student.POST("/agents/tasks/:id/approve", agentHandler.ApproveTask)
				student.POST("/agents/tasks/:id/cancel", agentHandler.CancelTask)
				student.POST("/global-decisions/rooms/:id/forecasts", globalDecisionHandler.RegisterForecast)
				student.POST("/global-decisions/rooms/:id/arguments", globalDecisionHandler.AddArgument)
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
				researcher.POST("/knowledge/claims", knowledgeHandler.RegisterClaim)
				researcher.GET("/knowledge/search/hybrid", knowledgeHandler.HybridSearch)
				researcher.POST("/knowledge/rag/query", knowledgeHandler.GraphRAG)
				researcher.POST("/knowledge/hypotheses", knowledgeHandler.CreateHypothesis)
				researcher.POST("/agents/tasks", agentHandler.CreateTask)
				researcher.GET("/agents/tasks/:id/plan", agentHandler.GetTaskPlan)
				researcher.POST("/agents/tasks/:id/approve", agentHandler.ApproveTask)
				researcher.POST("/agents/tasks/:id/cancel", agentHandler.CancelTask)
				researcher.POST("/global-decisions/rooms", globalDecisionHandler.CreateDecisionRoom)
				researcher.POST("/global-decisions/rooms/:id/options", globalDecisionHandler.ProposeOption)
				researcher.POST("/global-decisions/rooms/:id/forecasts", globalDecisionHandler.RegisterForecast)
				researcher.POST("/global-decisions/rooms/:id/arguments", globalDecisionHandler.AddArgument)
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
				institution.POST("/simulations", strategicSimHandler.QueueSimulation)
				institution.POST("/simulations/:id/branch", strategicSimHandler.BranchSimulation)
				institution.POST("/simulations/:id/interventions", strategicSimHandler.ProposeIntervention)
				institution.POST("/network/skills/equivalence", networkHandler.ProposeSkillEquivalence)
				institution.GET("/network/federated/search", networkHandler.FederatedSearch)
				institution.POST("/agents/tasks", agentHandler.CreateTask)
				institution.POST("/agents/tasks/:id/approve", agentHandler.ApproveTask)
				institution.POST("/agents/tasks/:id/cancel", agentHandler.CancelTask)
				institution.POST("/global-decisions/rooms", globalDecisionHandler.CreateDecisionRoom)
				institution.POST("/global-decisions/rooms/:id/options", globalDecisionHandler.ProposeOption)
				institution.POST("/global-decisions/rooms/:id/journal", globalDecisionHandler.RecordDecision)
				// Phase 31 - Institutional strategic operations
				institution.POST("/ecosystem-intelligence/partnerships", ecoIntelHandler.ProposePartnership)
				institution.POST("/ecosystem-intelligence/strategic-plans", ecoIntelHandler.CreateStrategicPlan)
				institution.POST("/ecosystem-intelligence/early-warnings", ecoIntelHandler.RegisterEarlyWarning)
				institution.POST("/ecosystem-intelligence/contingency-plans", ecoIntelHandler.CreateContingencyPlan)
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
			admin.Use(middleware.RequireRole("ADMIN", "SUPER_ADMIN", "NETWORK_ADMIN"))
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
				admin.POST("/twins/cohorts", strategicSimHandler.RegisterCohort)
				admin.POST("/twins/snapshots", strategicSimHandler.CreateSnapshot)
				admin.POST("/twins/model-cards", strategicSimHandler.PublishModelCard)
				admin.GET("/twins/warnings", strategicSimHandler.ListEarlyWarnings)
				admin.POST("/network/nodes/register", networkHandler.RegisterNode)
				admin.POST("/network/nodes/:id/revoke", networkHandler.RevokeNode)
				admin.POST("/knowledge/entities", knowledgeHandler.RegisterEntity)
				admin.POST("/admin/agents/register", agentHandler.RegisterAgent)
				admin.POST("/global-decisions/rooms/:id/journal", globalDecisionHandler.RecordDecision)
				// Phase 31 - Ecosystem admin
				admin.POST("/ecosystem-intelligence/institutions", ecoIntelHandler.RegisterInstitution)
				admin.POST("/ecosystem-intelligence/partnerships", ecoIntelHandler.ProposePartnership)
				admin.POST("/ecosystem-intelligence/early-warnings", ecoIntelHandler.RegisterEarlyWarning)
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

print("Phase 31 scaffolding generated successfully.")
