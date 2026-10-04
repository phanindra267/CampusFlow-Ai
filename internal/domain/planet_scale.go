package domain

import "time"

type ScientificClaim struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	Statement       string                 `json:"statement" binding:"required"`
	SourceReference string                 `json:"source_reference"`
	Confidence      float64                `json:"confidence"`
	Status          string                 `json:"status"`
	Context         map[string]interface{} `json:"context"`
	Evidence        []interface{}          `json:"evidence"`
	Contradictions  []interface{}          `json:"contradictions"`
	Version         int                    `json:"version"`
	AuthorID        string                 `json:"author_id"`
}

type CurriculumNode struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	NodeType    string                 `json:"node_type" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type CurriculumEdge struct {
	ID               string                 `json:"id"`
	TenantID         string                 `json:"tenant_id"`
	SourceID         string                 `json:"source_id" binding:"required"`
	TargetID         string                 `json:"target_id" binding:"required"`
	RelationshipType string                 `json:"relationship_type" binding:"required"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type ResearchFunding struct {
	ID           string        `json:"id"`
	ProjectID    string        `json:"project_id"`
	GrantName    string        `json:"grant_name" binding:"required"`
	Sponsor      string        `json:"sponsor"`
	TotalBudget  float64       `json:"total_budget"`
	SpentBudget  float64       `json:"spent_budget"`
	Currency     string        `json:"currency"`
	Restrictions []interface{} `json:"restrictions"`
	StartDate    *time.Time    `json:"start_date"`
	EndDate      *time.Time    `json:"end_date"`
	Status       string        `json:"status"`
}

type ReplicationStudy struct {
	ID                      string                 `json:"id"`
	OriginalExperimentID    string                 `json:"original_experiment_id" binding:"required"`
	ReplicatingResearcherID string                 `json:"replicating_researcher_id"`
	Protocol                map[string]interface{} `json:"protocol"`
	EnvironmentSnapshot     map[string]interface{} `json:"environment_snapshot"`
	Status                  string                 `json:"status"`
	Results                 map[string]interface{} `json:"results,omitempty"`
	Comparison              map[string]interface{} `json:"comparison,omitempty"`
	Differences             string                 `json:"differences,omitempty"`
}
