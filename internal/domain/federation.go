package domain

type FederationTrust struct {
	ID             string                 `json:"id"`
	SourceTenantID string                 `json:"source_tenant_id"`
	TargetTenantID string                 `json:"target_tenant_id"`
	TrustLevel     string                 `json:"trust_level"`
	AllowedScopes  map[string]interface{} `json:"allowed_scopes"`
	Status         string                 `json:"status"`
}

type ResearchWorkspace struct {
	ID            string `json:"id"`
	OwnerTenantID string `json:"owner_tenant_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Visibility    string `json:"visibility"`
	Status        string `json:"status"`
}

type WorkspaceExperiment struct {
	ID             string                 `json:"id"`
	WorkspaceID    string                 `json:"workspace_id"`
	Name           string                 `json:"name"`
	Configuration  map[string]interface{} `json:"configuration"`
	DatasetVersion string                 `json:"dataset_version"`
	ModelVersion   string                 `json:"model_version"`
	Status         string                 `json:"status"`
	Results        map[string]interface{} `json:"results"`
}

type AIIncident struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	IncidentType    string                 `json:"incident_type"`
	Severity        string                 `json:"severity"`
	AffectedAgentID string                 `json:"affected_agent_id,omitempty"`
	Description     string                 `json:"description"`
	Evidence        map[string]interface{} `json:"evidence"`
	Status          string                 `json:"status"`
}
