package domain

import "time"

type ResearchProject struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	Name          string    `json:"name" binding:"required"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	MaturityStage string    `json:"maturity_stage"`
	OwnerID       string    `json:"owner_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type ResearchHypothesis struct {
	ID                 string                 `json:"id"`
	ProjectID          string                 `json:"project_id"`
	Statement          string                 `json:"statement" binding:"required"`
	Variables          map[string]interface{} `json:"variables"`
	SupportingEvidence []interface{}          `json:"supporting_evidence"`
	CounterEvidence    []interface{}          `json:"counter_evidence"`
	Status             string                 `json:"status"`
	Version            int                    `json:"version"`
	ChangeReason       string                 `json:"change_reason"`
}

type ResearchDataset struct {
	ID                    string                 `json:"id"`
	TenantID              string                 `json:"tenant_id"`
	Name                  string                 `json:"name" binding:"required"`
	Version               string                 `json:"version"`
	SchemaDefinition      map[string]interface{} `json:"schema_definition"`
	Source                string                 `json:"source"`
	License               string                 `json:"license"`
	PrivacyClassification string                 `json:"privacy_classification"`
	QualityMetadata       map[string]interface{} `json:"quality_metadata"`
	OwnerID               string                 `json:"owner_id"`
}

type ResearchExperiment struct {
	ID                  string                 `json:"id"`
	ProjectID           string                 `json:"project_id"`
	HypothesisID        string                 `json:"hypothesis_id,omitempty"`
	Name                string                 `json:"name" binding:"required"`
	Objective           string                 `json:"objective"`
	DatasetID           string                 `json:"dataset_id"`
	DatasetVersion      string                 `json:"dataset_version"`
	Parameters          map[string]interface{} `json:"parameters"`
	EnvironmentSnapshot map[string]interface{} `json:"environment_snapshot"`
	Status              string                 `json:"status"`
	Results             map[string]interface{} `json:"results,omitempty"`
}

type ResearchArtifact struct {
	ID                 string `json:"id"`
	ProjectID          string `json:"project_id"`
	ArtifactType       string `json:"artifact_type"`
	Name               string `json:"name" binding:"required"`
	Version            string `json:"version"`
	Description        string `json:"description"`
	IsImmutable        bool   `json:"is_immutable"`
	LinkedExperimentID string `json:"linked_experiment_id,omitempty"`
	OwnerID            string `json:"owner_id"`
}

type InventionDisclosure struct {
	ID           string        `json:"id"`
	TenantID     string        `json:"tenant_id"`
	ProjectID    string        `json:"project_id,omitempty"`
	Title        string        `json:"title" binding:"required"`
	Description  string        `json:"description" binding:"required"`
	Contributors []interface{} `json:"contributors"`
	Status       string        `json:"status"`
}
