package domain

import "time"

// Phase 15 types — preserved from original ecosystem.go
type TrustedInstitution struct {
	ID                 string `json:"id"`
	Name               string `json:"name" binding:"required"`
	DID                string `json:"did"`
	AdminOwnerID       string `json:"admin_owner_id"`
	VerificationStatus string `json:"verification_status"`
	FederationStatus   string `json:"federation_status"`
	SecurityStatus     string `json:"security_status"`
}

type AgentTool struct {
	ID          string `json:"id"`
	Name        string `json:"name" binding:"required"`
	Endpoint    string `json:"endpoint" binding:"required"`
	Version     string `json:"version"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
}

type ServiceCatalogEntry struct {
	ID              string `json:"id"`
	ServiceName     string `json:"service_name" binding:"required"`
	OwnerID         string `json:"owner_id"`
	AvailabilitySLA string `json:"availability_sla"`
	HealthStatus    string `json:"health_status"`
	Version         string `json:"version"`
}

type AutomationLoopGuard struct {
	ID              string `json:"id"`
	WorkflowID      string `json:"workflow_id" binding:"required"`
	MaxInvocations  int    `json:"max_invocations"`
	InvocationCount int    `json:"invocation_count"`
	Status          string `json:"status"`
}

type GovernanceReview struct {
	ID       string `json:"id"`
	Subject  string `json:"subject" binding:"required"`
	Reviewer string `json:"reviewer"`
	Status   string `json:"status"`
}

// Phase 31 types — Global Institutional Intelligence

type EcosystemInstitution struct {
	ID               string                 `json:"id"`
	InstitutionType  string                 `json:"institution_type" binding:"required"`
	LegalName        string                 `json:"legal_name" binding:"required"`
	Domains          []string               `json:"domains"`
	Capabilities     map[string]interface{} `json:"capabilities"`
	GovernanceConfig map[string]interface{} `json:"governance_config"`
	PrivacyLevel     string                 `json:"privacy_level"`
	Status           string                 `json:"status"`
}

type EcosystemPartnership struct {
	ID                     string                 `json:"id"`
	InitiatorInstitutionID string                 `json:"initiator_institution_id" binding:"required"`
	PartnerInstitutionID   string                 `json:"partner_institution_id" binding:"required"`
	PartnershipType        string                 `json:"partnership_type" binding:"required"`
	LifecycleStatus        string                 `json:"lifecycle_status"`
	Objectives             map[string]interface{} `json:"objectives"`
	Evidence               map[string]interface{} `json:"evidence"`
	RiskAssessment         map[string]interface{} `json:"risk_assessment"`
	DataSharingContract    map[string]interface{} `json:"data_sharing_contract"`
}

type EcosystemStrategicPlan struct {
	ID            string                 `json:"id"`
	InstitutionID string                 `json:"institution_id" binding:"required"`
	Version       string                 `json:"version"`
	Mission       string                 `json:"mission"`
	Objectives    []interface{}          `json:"objectives" binding:"required"`
	Initiatives   []interface{}          `json:"initiatives"`
	KPIs          map[string]interface{} `json:"kpis"`
	Risks         []interface{}          `json:"risks"`
	Scenarios     []interface{}          `json:"scenarios"`
	Status        string                 `json:"status"`
	CreatedAt     *time.Time             `json:"created_at,omitempty"`
}

type EcosystemEarlyWarning struct {
	ID            string                 `json:"id"`
	InstitutionID string                 `json:"institution_id"`
	SignalType    string                 `json:"signal_type" binding:"required"`
	SignalData    map[string]interface{} `json:"signal_data" binding:"required"`
	Evidence      map[string]interface{} `json:"evidence"`
	Confidence    float64                `json:"confidence"`
	TimePeriod    string                 `json:"time_period"`
	ReviewStatus  string                 `json:"review_status"`
}

type EcosystemContingencyPlan struct {
	ID                 string                 `json:"id"`
	InstitutionID      string                 `json:"institution_id" binding:"required"`
	TriggerDescription string                 `json:"trigger_description" binding:"required"`
	ResponsePlan       map[string]interface{} `json:"response_plan" binding:"required"`
	OwnerID            string                 `json:"owner_id"`
	Resources          map[string]interface{} `json:"resources"`
	ApprovalStatus     string                 `json:"approval_status"`
	ReviewDate         *time.Time             `json:"review_date,omitempty"`
}
