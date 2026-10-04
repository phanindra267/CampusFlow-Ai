package domain

type StrategicPlan struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Objectives  map[string]interface{} `json:"objectives"`
	Assumptions map[string]interface{} `json:"assumptions"`
	Status      string                 `json:"status"`
	OwnerID     string                 `json:"owner_id"`
}

type DecisionRoom struct {
	ID                string                 `json:"id"`
	TenantID          string                 `json:"tenant_id"`
	Topic             string                 `json:"topic"`
	Evidence          map[string]interface{} `json:"evidence"`
	AIRecommendations map[string]interface{} `json:"ai_recommendations"`
	FinalDecision     map[string]interface{} `json:"final_decision,omitempty"`
	Status            string                 `json:"status"`
}

type EvidenceGraph struct {
	ID              string  `json:"id"`
	Claim           string  `json:"claim"`
	SourceReference string  `json:"source_reference"`
	DatasetVersion  string  `json:"dataset_version"`
	Confidence      float64 `json:"confidence"`
	TenantID        string  `json:"tenant_id"`
}

type AutonomousPolicy struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	PolicyName   string                 `json:"policy_name"`
	AllowedScope map[string]interface{} `json:"allowed_scope"`
	Mode         string                 `json:"mode"`
	RollbackPlan string                 `json:"rollback_plan"`
}

type SearchQuery struct {
	Query      string                 `json:"query"`
	Filters    map[string]interface{} `json:"filters"`
	SearchMode string                 `json:"search_mode"`
}

type Recommendation struct {
	ID         string  `json:"id"`
	EntityID   string  `json:"entity_id"`
	EntityType string  `json:"entity_type"`
	Score      float64 `json:"score"`
	Reason     string  `json:"reason"`
}

type RecommendationFeedback struct {
	ID               string `json:"id"`
	RecommendationID string `json:"recommendation_id"`
	UserID           string `json:"user_id"`
	Feedback         string `json:"feedback"`
}
