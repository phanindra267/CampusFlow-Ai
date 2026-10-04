package domain

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
	ID         string                 `json:"id"`
	SignalType string                 `json:"signal_type"`
	Evidence   map[string]interface{} `json:"evidence"`
	Confidence float64                `json:"confidence"`
	TimeWindow string                 `json:"time_window"`
	Status     string                 `json:"status"`
	CreatedAt  *time.Time             `json:"created_at,omitempty"`
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
