package domain

import "time"

type GlobalTrustEntity struct {
	ID                string                 `json:"id"`
	EntityType        string                 `json:"entity_type" binding:"required"`
	EntityID          string                 `json:"entity_id" binding:"required"`
	TrustScore        float64                `json:"trust_score"`
	IdentityAssurance string                 `json:"identity_assurance"`
	SecurityPosture   string                 `json:"security_posture"`
	GovernanceStatus  string                 `json:"governance_status"`
	Evidence          map[string]interface{} `json:"evidence"`
	LastEvaluatedAt   *time.Time             `json:"last_evaluated_at,omitempty"`
}

type GlobalAgent struct {
	ID                   string                 `json:"id"`
	OwnerTenantID        string                 `json:"owner_tenant_id"`
	Name                 string                 `json:"name" binding:"required"`
	Version              string                 `json:"version" binding:"required"`
	CertificationStatus  string                 `json:"certification_status"`
	Capabilities         map[string]interface{} `json:"capabilities"`
	TrustEntityID        string                 `json:"trust_entity_id,omitempty"`
	ResourceRequirements map[string]interface{} `json:"resource_requirements"`
	SupportedProtocols   []interface{}          `json:"supported_protocols"`
	AuthorityLevel       string                 `json:"authority_level"`
	Status               string                 `json:"status"`
}

type HumanAITeam struct {
	ID                 string                 `json:"id"`
	TenantID           string                 `json:"tenant_id"`
	Name               string                 `json:"name" binding:"required"`
	Objective          string                 `json:"objective" binding:"required"`
	Members            []interface{}          `json:"members" binding:"required"`
	Permissions        map[string]interface{} `json:"permissions"`
	Deadline           *time.Time             `json:"deadline,omitempty"`
	SuccessCriteria    map[string]interface{} `json:"success_criteria"`
	PerformanceMetrics map[string]interface{} `json:"performance_metrics"`
	Status             string                 `json:"status"`
}

type CollectiveReasoningJob struct {
	ID                  string                 `json:"id"`
	InitiatorTenantID   string                 `json:"initiator_tenant_id"`
	Topic               string                 `json:"topic" binding:"required"`
	Hypotheses          []interface{}          `json:"hypotheses"`
	Evidence            []interface{}          `json:"evidence"`
	AssignedTeams       []interface{}          `json:"assigned_teams"`
	ConsensusState      string                 `json:"consensus_state"`
	AdversarialReviews  []interface{}          `json:"adversarial_reviews"`
	FinalRecommendation map[string]interface{} `json:"final_recommendation"`
	Confidence          float64                `json:"confidence"`
	DissentingOpinions  []interface{}          `json:"dissenting_opinions"`
	Status              string                 `json:"status"`
}

type GlobalForecast struct {
	ID                  string                 `json:"id"`
	Subject             string                 `json:"subject" binding:"required"`
	EnsembleMembers     []interface{}          `json:"ensemble_members"`
	Prediction          map[string]interface{} `json:"prediction"`
	Confidence          float64                `json:"confidence"`
	CalibrationScore    float64                `json:"calibration_score"`
	DisagreementDetails map[string]interface{} `json:"disagreement_details"`
	ValidUntil          *time.Time             `json:"valid_until,omitempty"`
	Status              string                 `json:"status"`
}

type GlobalRisk struct {
	ID                   string                 `json:"id"`
	Threat               string                 `json:"threat" binding:"required"`
	Probability          float64                `json:"probability"`
	Impact               string                 `json:"impact"`
	Exposure             map[string]interface{} `json:"exposure"`
	Mitigation           string                 `json:"mitigation"`
	SystemicDependencies []interface{}          `json:"systemic_dependencies"`
	OwnerID              string                 `json:"owner_id"`
	Confidence           float64                `json:"confidence"`
	Status               string                 `json:"status"`
}
