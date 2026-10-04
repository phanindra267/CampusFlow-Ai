package domain

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

type GraphClaim struct {
	ID                 string                 `json:"id"`
	EntityID           string                 `json:"entity_id"`
	ClaimText          string                 `json:"claim_text" binding:"required"`
	ClaimType          string                 `json:"claim_type"`
	Evidence           map[string]interface{} `json:"evidence" binding:"required"`
	Confidence         float64                `json:"confidence"`
	IsMachineExtracted bool                   `json:"is_machine_extracted"`
	VerificationStatus string                 `json:"verification_status"`
}

type GraphSource struct {
	ID                string                 `json:"id"`
	SourceURI         string                 `json:"source_uri" binding:"required"`
	Publisher         string                 `json:"publisher"`
	LicenseInfo       string                 `json:"license_info"`
	TrustScore        float64                `json:"trust_score"`
	IngestionContract map[string]interface{} `json:"ingestion_contract"`
}

type GeneratedHypothesis struct {
	ID                 string        `json:"id"`
	OwnerID            string        `json:"owner_id"`
	HypothesisText     string        `json:"hypothesis_text" binding:"required"`
	SupportingEvidence []interface{} `json:"supporting_evidence"`
	CounterEvidence    []interface{} `json:"counter_evidence"`
	IsAIGenerated      bool          `json:"is_ai_generated"`
	LifecycleStatus    string        `json:"lifecycle_status"`
}
