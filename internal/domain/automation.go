package domain

import "time"

type AutomationRule struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	TriggerType   string                 `json:"trigger_type" binding:"required"`
	Conditions    map[string]interface{} `json:"conditions"`
	ActionType    string                 `json:"action_type" binding:"required"`
	ActionPayload map[string]interface{} `json:"action_payload"`
	Status        string                 `json:"status"`
	CreatedAt     time.Time              `json:"created_at"`
}

type NotificationPreference struct {
	UserID           string `json:"user_id"`
	EmailEnabled     bool   `json:"email_enabled"`
	PushEnabled      bool   `json:"push_enabled"`
	DigestFrequency  string `json:"digest_frequency"`
	FatigueThreshold int    `json:"fatigue_threshold"`
}

type KnowledgeGap struct {
	ID             string    `json:"id"`
	QueryPattern   string    `json:"query_pattern"`
	Frequency      int       `json:"frequency"`
	Status         string    `json:"status"`
	LastDetectedAt time.Time `json:"last_detected_at"`
}
