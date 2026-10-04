package domain

import "time"

type AgentRegistryEntry struct {
	ID             string                 `json:"id"`
	TenantID       string                 `json:"tenant_id"`
	Name           string                 `json:"name" binding:"required"`
	Version        string                 `json:"version"`
	Purpose        string                 `json:"purpose" binding:"required"`
	OwnerID        string                 `json:"owner_id"`
	AutonomyLevel  int                    `json:"autonomy_level"`
	Capabilities   map[string]interface{} `json:"capabilities"`
	Permissions    map[string]interface{} `json:"permissions"`
	ToolIDs        []interface{}          `json:"tool_ids"`
	ResourceLimits map[string]interface{} `json:"resource_limits"`
	CostBudget     float64                `json:"cost_budget"`
	SafetyPolicy   map[string]interface{} `json:"safety_policy"`
	Status         string                 `json:"status"`
}

type AutonomyPolicy struct {
	ID                   string                 `json:"id"`
	TenantID             string                 `json:"tenant_id"`
	Name                 string                 `json:"name" binding:"required"`
	Version              int                    `json:"version"`
	AllowedActions       []interface{}          `json:"allowed_actions"`
	ForbiddenActions     []interface{}          `json:"forbidden_actions"`
	RequiredApprovals    map[string]interface{} `json:"required_approvals"`
	ResourceLimits       map[string]interface{} `json:"resource_limits"`
	CostLimits           map[string]interface{} `json:"cost_limits"`
	RiskLimits           map[string]interface{} `json:"risk_limits"`
	RetryLimits          map[string]interface{} `json:"retry_limits"`
	EscalationConditions map[string]interface{} `json:"escalation_conditions"`
	Status               string                 `json:"status"`
	OwnerID              string                 `json:"owner_id"`
}

type ActionPlan struct {
	ID               string                 `json:"id"`
	TenantID         string                 `json:"tenant_id"`
	AgentID          string                 `json:"agent_id,omitempty"`
	Goal             string                 `json:"goal" binding:"required"`
	Version          int                    `json:"version"`
	PolicyVersionID  string                 `json:"policy_version_id,omitempty"`
	Steps            []interface{}          `json:"steps"`
	Preconditions    map[string]interface{} `json:"preconditions"`
	Constraints      map[string]interface{} `json:"constraints"`
	ExpectedOutcome  map[string]interface{} `json:"expected_outcome"`
	Risks            map[string]interface{} `json:"risks"`
	RollbackStrategy map[string]interface{} `json:"rollback_strategy"`
	Status           string                 `json:"status"`
	ApprovedBy       string                 `json:"approved_by,omitempty"`
	ApprovedAt       *time.Time             `json:"approved_at,omitempty"`
}

type DecisionRecord struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	AgentID         string                 `json:"agent_id,omitempty"`
	PlanID          string                 `json:"plan_id,omitempty"`
	DecisionType    string                 `json:"decision_type" binding:"required"`
	Evidence        map[string]interface{} `json:"evidence"`
	PolicyVersionID string                 `json:"policy_version_id,omitempty"`
	Explanation     string                 `json:"explanation"`
	Outcome         map[string]interface{} `json:"outcome,omitempty"`
	HumanApproved   bool                   `json:"human_approved"`
	ApprovedBy      string                 `json:"approved_by,omitempty"`
}

type LessonLearned struct {
	ID               string `json:"id"`
	TenantID         string `json:"tenant_id"`
	Subject          string `json:"subject" binding:"required"`
	IncidentRef      string `json:"incident_ref,omitempty"`
	Observation      string `json:"observation" binding:"required"`
	RootCause        string `json:"root_cause"`
	CorrectiveAction string `json:"corrective_action"`
	ValidationStatus string `json:"validation_status"`
	ValidatedBy      string `json:"validated_by,omitempty"`
}
