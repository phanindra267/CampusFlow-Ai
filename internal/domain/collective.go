package domain

import "time"

// ── Phase 13 — Collective Intelligence types ─────────────────────────────────

type DeliberationRoom struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	Topic           string                 `json:"topic" binding:"required"`
	Description     string                 `json:"description"`
	Status          string                 `json:"status"`
	OwnerID         string                 `json:"owner_id"`
	DecisionOwnerID string                 `json:"decision_owner_id"`
	Metadata        map[string]interface{} `json:"metadata"`
}

type ArgumentGraph struct {
	ID            string        `json:"id"`
	RoomID        string        `json:"room_id"`
	AuthorID      string        `json:"author_id"`
	ContributedBy string        `json:"contributed_by"`
	Claim         string        `json:"claim" binding:"required"`
	Type          string        `json:"type"`
	Evidence      []interface{} `json:"evidence"`
	ParentID      string        `json:"parent_id,omitempty"`
}

type DecisionOutcome struct {
	ID         string                 `json:"id"`
	RoomID     string                 `json:"room_id" binding:"required"`
	Decision   string                 `json:"decision" binding:"required"`
	Rationale  string                 `json:"rationale"`
	Evidence   map[string]interface{} `json:"evidence"`
	RecordedBy string                 `json:"recorded_by"`
	Status     string                 `json:"status"`
}

type InstitutionalPlaybook struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Steps       []interface{}          `json:"steps"`
	Triggers    []interface{}          `json:"triggers"`
	Policies    map[string]interface{} `json:"policies"`
	Version     string                 `json:"version"`
	Status      string                 `json:"status"`
	OwnerID     string                 `json:"owner_id"`
}

type PredictionLedgerEntry struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	Subject      string                 `json:"subject" binding:"required"`
	ModelName    string                 `json:"model_name"`
	ModelVersion string                 `json:"model_version"`
	Prediction   map[string]interface{} `json:"prediction" binding:"required"`
	Confidence   float64                `json:"confidence"`
	Evidence     map[string]interface{} `json:"evidence"`
	ValidUntil   time.Time              `json:"valid_until"`
	AuthorID     string                 `json:"author_id"`
}

// ── Phase 19 — Collective Intelligence & Federation types ────────────────────

type FederationMember struct {
	ID                string                 `json:"id"`
	FederationID      string                 `json:"federation_id,omitempty"`
	TenantID          string                 `json:"tenant_id"`
	InstitutionName   string                 `json:"institution_name" binding:"required"`
	TrustLevel        string                 `json:"trust_level"`
	DataSharingPolicy map[string]interface{} `json:"data_sharing_policy"`
	CapabilityProfile map[string]interface{} `json:"capability_profile"`
	GovernancePolicy  map[string]interface{} `json:"governance_policy"`
	ComplianceConfig  map[string]interface{} `json:"compliance_config"`
	Status            string                 `json:"status"`
	JoinedAt          *time.Time             `json:"joined_at,omitempty"`
}

type FederationAgreement struct {
	ID                 string                 `json:"id"`
	FederationID       string                 `json:"federation_id,omitempty"`
	InitiatingTenantID string                 `json:"initiating_tenant_id"`
	Parties            []interface{}          `json:"parties" binding:"required"`
	Scope              map[string]interface{} `json:"scope"`
	Obligations        map[string]interface{} `json:"obligations"`
	Constraints        map[string]interface{} `json:"constraints"`
	DurationDays       int                    `json:"duration_days"`
	Status             string                 `json:"status"`
	Version            int                    `json:"version"`
	ApprovedBy         []interface{}          `json:"approved_by"`
	ExpiresAt          *time.Time             `json:"expires_at,omitempty"`
}

type FederatedKnowledgeClaim struct {
	ID              string                 `json:"id"`
	SourceTenantID  string                 `json:"source_tenant_id"`
	Statement       string                 `json:"statement" binding:"required"`
	SourceReference string                 `json:"source_reference"`
	Evidence        map[string]interface{} `json:"evidence"`
	Confidence      float64                `json:"confidence"`
	Visibility      string                 `json:"visibility"`
	Status          string                 `json:"status"`
	Contradictions  []interface{}          `json:"contradictions"`
	Version         int                    `json:"version"`
	OwnerID         string                 `json:"owner_id"`
}

type GovernanceProposal struct {
	ID                 string                 `json:"id"`
	FederationID       string                 `json:"federation_id,omitempty"`
	Title              string                 `json:"title" binding:"required"`
	Description        string                 `json:"description"`
	ProposedByTenantID string                 `json:"proposed_by_tenant_id"`
	ProposalType       string                 `json:"proposal_type" binding:"required"`
	Content            map[string]interface{} `json:"content"`
	Votes              map[string]interface{} `json:"votes"`
	QuorumRequired     int                    `json:"quorum_required"`
	Status             string                 `json:"status"`
	DecisionRationale  string                 `json:"decision_rationale,omitempty"`
	Version            int                    `json:"version"`
}

type FederationDispute struct {
	ID               string                 `json:"id"`
	FederationID     string                 `json:"federation_id,omitempty"`
	RaisedByTenantID string                 `json:"raised_by_tenant_id"`
	AgainstTenantID  string                 `json:"against_tenant_id,omitempty"`
	Subject          string                 `json:"subject" binding:"required"`
	Description      string                 `json:"description" binding:"required"`
	Evidence         map[string]interface{} `json:"evidence"`
	Status           string                 `json:"status"`
	Resolution       string                 `json:"resolution,omitempty"`
}
