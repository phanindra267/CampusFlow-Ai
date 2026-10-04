package domain

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
	ID                  string                 `json:"id"`
	TenantID            string                 `json:"tenant_id"`
	TwinID              string                 `json:"twin_id,omitempty"`
	ScenarioName        string                 `json:"scenario_name" binding:"required"`
	ModelName           string                 `json:"model_name" binding:"required"`
	ModelVersion        string                 `json:"model_version" binding:"required"`
	Parameters          map[string]interface{} `json:"parameters"`
	Assumptions         map[string]interface{} `json:"assumptions"`
	RandomSeed          int64                  `json:"random_seed,omitempty"`
	Status              string                 `json:"status"`
	Priority            int                    `json:"priority"`
	CPULimitCores       float64                `json:"cpu_limit_cores"`
	MemoryLimitMB       int                    `json:"memory_limit_mb"`
	RuntimeLimitSeconds int                    `json:"runtime_limit_seconds"`
	CostBudget          float64                `json:"cost_budget"`
	Results             map[string]interface{} `json:"results,omitempty"`
	Provenance          map[string]interface{} `json:"provenance,omitempty"`
}

type SimulationScenario struct {
	ID             string                 `json:"id"`
	TenantID       string                 `json:"tenant_id"`
	Name           string                 `json:"name" binding:"required"`
	Description    string                 `json:"description"`
	BaselineTwinID string                 `json:"baseline_twin_id,omitempty"`
	Version        int                    `json:"version"`
	Parameters     map[string]interface{} `json:"parameters"`
	Assumptions    map[string]interface{} `json:"assumptions"`
	IsAIGenerated  bool                   `json:"is_ai_generated"`
	OwnerID        string                 `json:"owner_id"`
}
