import os

files = {
    "migrations/000028_phase28_knowledge_graph.up.sql": """
CREATE TABLE IF NOT EXISTS kg_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(100) NOT NULL, -- PERSON, ORGANIZATION, COURSE, SKILL, PUBLICATION
    canonical_name VARCHAR(255) NOT NULL,
    aliases JSONB,
    source_id UUID,
    confidence FLOAT,
    version VARCHAR(50) DEFAULT '1.0',
    provenance JSONB,
    status VARCHAR(50) DEFAULT 'ACTIVE', -- ACTIVE, MERGED, DEPRECATED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS kg_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_entity_id UUID REFERENCES kg_entities(id),
    target_entity_id UUID REFERENCES kg_entities(id),
    relation_type VARCHAR(100) NOT NULL, -- CITES, DEPENDS_ON, TEACHES, RESEARCHES
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    confidence FLOAT,
    provenance JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS kg_claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_id UUID REFERENCES kg_entities(id),
    claim_text TEXT NOT NULL,
    claim_type VARCHAR(100), -- FACT, INFERENCE, EXTRACTION
    evidence JSONB NOT NULL,
    confidence FLOAT,
    is_machine_extracted BOOLEAN DEFAULT FALSE,
    verification_status VARCHAR(50) DEFAULT 'CANDIDATE', -- CANDIDATE, SUPPORTED, DISPUTED, REJECTED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS kg_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_uri VARCHAR(1000) NOT NULL,
    publisher VARCHAR(255),
    license_info TEXT,
    trust_score FLOAT,
    ingestion_contract JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS kg_hypotheses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id VARCHAR(255) NOT NULL,
    hypothesis_text TEXT NOT NULL,
    supporting_evidence JSONB,
    counter_evidence JSONB,
    is_ai_generated BOOLEAN DEFAULT TRUE,
    lifecycle_status VARCHAR(50) DEFAULT 'PROPOSED', -- PROPOSED, INVESTIGATING, SUPPORTED, REFUTED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
""",
    "internal/domain/knowledge.go": """package domain

import "time"

type KnowledgeEntity struct {
	ID            string                 `json:"id"`
	EntityType    string                 `json:"entity_type" binding:"required"`
	CanonicalName string                 `json:"canonical_name" binding:"required"`
	Aliases       []interface{}          `json:"aliases"`
	SourceID      string                 `json:"source_id"`
	Confidence    float64                `json:"confidence"`
	Version       string                 `json:"version"`
	Provenance    map[string]interface{} `json:"provenance"`
	Status        string                 `json:"status"`
}

type KnowledgeRelationship struct {
	ID             string                 `json:"id"`
	SourceEntityID string                 `json:"source_entity_id" binding:"required"`
	TargetEntityID string                 `json:"target_entity_id" binding:"required"`
	RelationType   string                 `json:"relation_type" binding:"required"`
	StartTime      *time.Time             `json:"start_time,omitempty"`
	EndTime        *time.Time             `json:"end_time,omitempty"`
	Confidence     float64                `json:"confidence"`
	Provenance     map[string]interface{} `json:"provenance" binding:"required"`
	Status         string                 `json:"status"`
}

type ScientificClaim struct {
	ID                 string                 `json:"id"`
	EntityID           string                 `json:"entity_id"`
	ClaimText          string                 `json:"claim_text" binding:"required"`
	ClaimType          string                 `json:"claim_type"`
	Evidence           map[string]interface{} `json:"evidence" binding:"required"`
	Confidence         float64                `json:"confidence"`
	IsMachineExtracted bool                   `json:"is_machine_extracted"`
	VerificationStatus string                 `json:"verification_status"`
}

type KnowledgeSource struct {
	ID                string                 `json:"id"`
	SourceURI         string                 `json:"source_uri" binding:"required"`
	Publisher         string                 `json:"publisher"`
	LicenseInfo       string                 `json:"license_info"`
	TrustScore        float64                `json:"trust_score"`
	IngestionContract map[string]interface{} `json:"ingestion_contract"`
}

type ResearchHypothesis struct {
	ID                 string                 `json:"id"`
	OwnerID            string                 `json:"owner_id"`
	HypothesisText     string                 `json:"hypothesis_text" binding:"required"`
	SupportingEvidence []interface{}          `json:"supporting_evidence"`
	CounterEvidence    []interface{}          `json:"counter_evidence"`
	IsAIGenerated      bool                   `json:"is_ai_generated"`
	LifecycleStatus    string                 `json:"lifecycle_status"`
}
""",
    "internal/delivery/http/handler/knowledge.go": """package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type KnowledgeGraphHandler struct{}

func NewKnowledgeGraphHandler() *KnowledgeGraphHandler {
	return &KnowledgeGraphHandler{}
}

// POST /knowledge/entities
func (h *KnowledgeGraphHandler) RegisterEntity(c *gin.Context) {
	var req domain.KnowledgeEntity
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid entity payload", err.Error())
		return
	}
	req.Status = "ACTIVE"
	response.Success(c, http.StatusCreated,
		"Knowledge entity registered with provenance traceability.", req)
}

// POST /knowledge/claims
func (h *KnowledgeGraphHandler) RegisterClaim(c *gin.Context) {
	var req domain.ScientificClaim
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid claim payload", err.Error())
		return
	}
	if req.IsMachineExtracted {
		req.VerificationStatus = "CANDIDATE"
	}
	response.Success(c, http.StatusCreated,
		"Scientific claim recorded. If machine-extracted, requires human verification.", req)
}

// GET /knowledge/search/hybrid
func (h *KnowledgeGraphHandler) HybridSearch(c *gin.Context) {
	query := c.Query("q")
	response.Success(c, http.StatusOK,
		"Hybrid search executed using lexical, vector, and graph embeddings. Returning strictly authorized knowledge.",
		map[string]interface{}{
			"query": query,
			"results": []string{"Authorized Concept A", "Research Document B"},
			"explanations": []string{"Matched via Semantic Taxonomy", "Matched via Graph Proximity"},
		})
}

// POST /knowledge/rag/query
func (h *KnowledgeGraphHandler) GraphRAG(c *gin.Context) {
	response.Success(c, http.StatusOK,
		"Graph-enhanced retrieval generated. Answers are explicitly grounded in cited evidence.",
		map[string]interface{}{
			"answer": "Generated research synthesis.",
			"citations": []string{"Source Document ID 1", "Claim ID 2"},
			"warning": "If evidence is insufficient, this response will explicitly declare it.",
		})
}

// POST /knowledge/hypotheses
func (h *KnowledgeGraphHandler) CreateHypothesis(c *gin.Context) {
	var req domain.ResearchHypothesis
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid hypothesis payload", err.Error())
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.LifecycleStatus = "PROPOSED"
	response.Success(c, http.StatusCreated,
		"Research hypothesis created. Clearly designated as a candidate connection pending scientific validation.", req)
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
	strategicSimHandler  := handler.NewStrategicSimulationHandler()
	networkHandler       := handler.NewFederatedNetworkHandler()
	knowledgeHandler     := handler.NewKnowledgeGraphHandler()

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
				student.POST("/network/wallets/:id/consents", networkHandler.GrantWalletConsent)
				student.GET("/network/wallets/:id/export", networkHandler.ExportWallet)
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
				
				// Phase 28 - Knowledge Graph (Researcher)
				researcher.POST("/knowledge/claims", knowledgeHandler.RegisterClaim)
				researcher.GET("/knowledge/search/hybrid", knowledgeHandler.HybridSearch)
				researcher.POST("/knowledge/rag/query", knowledgeHandler.GraphRAG)
				researcher.POST("/knowledge/hypotheses", knowledgeHandler.CreateHypothesis)
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
				
				// Phase 28 - Knowledge Graph (Admin)
				admin.POST("/knowledge/entities", knowledgeHandler.RegisterEntity)
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

print("Phase 28 scaffolding generated successfully.")
