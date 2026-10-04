package domain

type ScientificHypothesis struct {
	ID                    string        `json:"id"`
	TenantID              string        `json:"tenant_id"`
	Statement             string        `json:"statement" binding:"required"`
	Context               string        `json:"context"`
	SupportingEvidence    []interface{} `json:"supporting_evidence"`
	ContradictingEvidence []interface{} `json:"contradicting_evidence"`
	Assumptions           []interface{} `json:"assumptions"`
	ExpectedPredictions   []interface{} `json:"expected_predictions"`
	Confidence            float64       `json:"confidence"`
	OwnerID               string        `json:"owner_id"`
	Status                string        `json:"status"`
}

type ScientificExperiment struct {
	ID                    string                 `json:"id"`
	HypothesisID          string                 `json:"hypothesis_id" binding:"required"`
	Objective             string                 `json:"objective" binding:"required"`
	Variables             map[string]interface{} `json:"variables"`
	Controls              map[string]interface{} `json:"controls"`
	Measurements          map[string]interface{} `json:"measurements"`
	SampleRequirements    map[string]interface{} `json:"sample_requirements"`
	ExpectedOutcomes      []interface{}          `json:"expected_outcomes"`
	Risks                 []interface{}          `json:"risks"`
	StopConditions        []interface{}          `json:"stop_conditions"`
	ReproducibilityConfig map[string]interface{} `json:"reproducibility_config"`
	Status                string                 `json:"status"`
}

type IntelligenceImprovement struct {
	ID              string                 `json:"id"`
	TargetComponent string                 `json:"target_component" binding:"required"`
	CurrentState    map[string]interface{} `json:"current_state"`
	ProposedChange  map[string]interface{} `json:"proposed_change"`
	ExpectedBenefit string                 `json:"expected_benefit"`
	ExpectedRisk    string                 `json:"expected_risk"`
	ValidationPlan  map[string]interface{} `json:"validation_plan"`
	RollbackPlan    map[string]interface{} `json:"rollback_plan"`
	SandboxResults  map[string]interface{} `json:"sandbox_results"`
	Status          string                 `json:"status"`
}

type CausalRelationship struct {
	ID               string                 `json:"id"`
	SourceEntity     string                 `json:"source_entity" binding:"required"`
	TargetEntity     string                 `json:"target_entity" binding:"required"`
	RelationshipType string                 `json:"relationship_type"`
	EvidenceStrength float64                `json:"evidence_strength"`
	Provenance       map[string]interface{} `json:"provenance"`
}

type ScientificArtifact struct {
	ID                  string                 `json:"id"`
	ArtifactType        string                 `json:"artifact_type" binding:"required"`
	Name                string                 `json:"name" binding:"required"`
	Version             string                 `json:"version" binding:"required"`
	Provenance          map[string]interface{} `json:"provenance"`
	StorageRef          string                 `json:"storage_ref"`
	SecurityScanResults map[string]interface{} `json:"security_scan_results"`
	Status              string                 `json:"status"`
}
