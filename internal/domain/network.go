package domain

import "time"

type NetworkNode struct {
	ID               string                 `json:"id"`
	OrganizationID   string                 `json:"organization_id" binding:"required"`
	OrganizationType string                 `json:"organization_type" binding:"required"`
	Region           string                 `json:"region"`
	TrustStatus      string                 `json:"trust_status"`
	PublicKey        string                 `json:"public_key"`
	Capabilities     map[string]interface{} `json:"capabilities"`
	HealthStatus     string                 `json:"health_status"`
	CreatedAt        *time.Time             `json:"created_at,omitempty"`
}

type DataContract struct {
	ID                 string                 `json:"id"`
	ProviderNodeID     string                 `json:"provider_node_id" binding:"required"`
	ProductName        string                 `json:"product_name" binding:"required"`
	SchemaDefinition   map[string]interface{} `json:"schema_definition"`
	AccessRequirements string                 `json:"access_requirements"`
	PrivacyPolicy      string                 `json:"privacy_policy"`
	Version            string                 `json:"version"`
	Status             string                 `json:"status"`
}

type SkillEquivalence struct {
	ID                    string                 `json:"id"`
	SourceSkillID         string                 `json:"source_skill_id" binding:"required"`
	TargetSkillID         string                 `json:"target_skill_id" binding:"required"`
	SourceNodeID          string                 `json:"source_node_id"`
	TargetNodeID          string                 `json:"target_node_id"`
	EquivalenceConfidence float64                `json:"equivalence_confidence"`
	Evidence              map[string]interface{} `json:"evidence"`
	Status                string                 `json:"status"`
}

type LearningWallet struct {
	ID                       string                 `json:"id"`
	UserID                   string                 `json:"user_id" binding:"required"`
	PortableRecords          []interface{}          `json:"portable_records"`
	SelectiveDisclosureRules map[string]interface{} `json:"selective_disclosure_rules"`
	Status                   string                 `json:"status"`
}

type NetworkConsent struct {
	ID           string                 `json:"id"`
	WalletID     string                 `json:"wallet_id" binding:"required"`
	TargetNodeID string                 `json:"target_node_id" binding:"required"`
	Scope        map[string]interface{} `json:"scope" binding:"required"`
	Status       string                 `json:"status"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
}
