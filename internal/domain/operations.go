package domain

import "time"

type CampusStateSnapshot struct {
	ID           string                 `json:"id"`
	SnapshotName string                 `json:"snapshot_name"`
	SnapshotData map[string]interface{} `json:"snapshot_data"`
	CreatedAt    time.Time              `json:"created_at"`
}

type OpsSimulationScenario struct {
	ID           string                 `json:"id"`
	SnapshotID   string                 `json:"snapshot_id"`
	Name         string                 `json:"name" binding:"required"`
	Changes      map[string]interface{} `json:"changes"`
	Constraints  map[string]interface{} `json:"constraints"`
	Assumptions  map[string]interface{} `json:"assumptions"`
	Results      map[string]interface{} `json:"results"`
	CreatedBy    string                 `json:"created_by"`
	ModelVersion string                 `json:"model_version"`
}

type OperationalIncident struct {
	ID                string                 `json:"id"`
	Severity          string                 `json:"severity"`
	AffectedServices  map[string]interface{} `json:"affected_services"`
	AffectedResources map[string]interface{} `json:"affected_resources"`
	Symptoms          map[string]interface{} `json:"symptoms"`
	Status            string                 `json:"status"`
	OwnerID           string                 `json:"owner_id"`
	Resolution        string                 `json:"resolution"`
	DetectedAt        time.Time              `json:"detected_at"`
}

type ActionExecution struct {
	ID              string                 `json:"id"`
	RequestedBy     string                 `json:"requested_by"`
	ApprovedBy      string                 `json:"approved_by"`
	ActionType      string                 `json:"action_type" binding:"required"`
	Parameters      map[string]interface{} `json:"parameters"`
	Status          string                 `json:"status"`
	ExecutionResult map[string]interface{} `json:"execution_result"`
}
