package domain

import "time"

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Campus struct {
	ID       string                 `json:"id"`
	TenantID string                 `json:"tenant_id"`
	Name     string                 `json:"name"`
	Location map[string]interface{} `json:"location"`
}

type EnterpriseConnector struct {
	ID            string                 `json:"id"`
	TenantID      string                 `json:"tenant_id"`
	ProviderName  string                 `json:"provider_name"`
	Configuration map[string]interface{} `json:"configuration"`
	Status        string                 `json:"status"`
	LastSyncAt    time.Time              `json:"last_sync_at"`
}

type AgentDefinition struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Purpose      string                 `json:"purpose"`
	AllowedTools map[string]interface{} `json:"allowed_tools"`
	RiskLevel    string                 `json:"risk_level"`
	Status       string                 `json:"status"`
	OwnerID      string                 `json:"owner_id"`
}
