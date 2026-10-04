package domain

import "time"

type EduInstitution struct {
	ID                string                 `json:"id"`
	OfficialName      string                 `json:"official_name" binding:"required"`
	InstitutionType   string                 `json:"institution_type"`
	Jurisdiction      string                 `json:"jurisdiction"`
	AccreditationInfo map[string]interface{} `json:"accreditation_info"`
	PublicKeys        []interface{}          `json:"public_keys"`
	APIEndpoints      map[string]interface{} `json:"api_endpoints"`
	TrustLevel        string                 `json:"trust_level"`
	Capabilities      []interface{}          `json:"capabilities"`
	Status            string                 `json:"status"`
}

type CredentialSchema struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name" binding:"required"`
	Version          string                 `json:"version" binding:"required"`
	SchemaDefinition map[string]interface{} `json:"schema_definition" binding:"required"`
	IssuerID         string                 `json:"issuer_id" binding:"required"`
}

type DigitalCredential struct {
	ID               string                 `json:"id"`
	SchemaID         string                 `json:"schema_id" binding:"required"`
	IssuerID         string                 `json:"issuer_id" binding:"required"`
	RecipientID      string                 `json:"recipient_id" binding:"required"`
	CredentialType   string                 `json:"credential_type" binding:"required"`
	Claims           map[string]interface{} `json:"claims" binding:"required"`
	Evidence         map[string]interface{} `json:"evidence"`
	DigitalSignature string                 `json:"digital_signature"`
	IssuedAt         *time.Time             `json:"issued_at,omitempty"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
	RevocationReason string                 `json:"revocation_reason,omitempty"`
	Status           string                 `json:"status"`
}

type DataConsent struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id" binding:"required"`
	RequesterID   string                 `json:"requester_id" binding:"required"`
	RequestedData map[string]interface{} `json:"requested_data" binding:"required"`
	Purpose       string                 `json:"purpose" binding:"required"`
	Status        string                 `json:"status"`
	ExpiresAt     *time.Time             `json:"expires_at,omitempty"`
}

type TransferCreditRequest struct {
	ID                  string                 `json:"id"`
	StudentID           string                 `json:"student_id" binding:"required"`
	SourceInstitutionID string                 `json:"source_institution_id" binding:"required"`
	TargetInstitutionID string                 `json:"target_institution_id" binding:"required"`
	SourceCourseInfo    map[string]interface{} `json:"source_course_info" binding:"required"`
	EquivalencyMapping  map[string]interface{} `json:"equivalency_mapping"`
	Confidence          float64                `json:"confidence"`
	Status              string                 `json:"status"`
	ReviewedBy          string                 `json:"reviewed_by,omitempty"`
}

type MobilityProfile struct {
	ID               string        `json:"id"`
	StudentID        string        `json:"student_id" binding:"required"`
	LinkedIdentities []interface{} `json:"linked_identities"`
	VerifiedSkills   []interface{} `json:"verified_skills"`
	ExchangeHistory  []interface{} `json:"exchange_history"`
}
