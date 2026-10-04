package domain

import "time"

type GlobalDecisionRoom struct {
	ID             string     `json:"id"`
	Topic          string     `json:"topic" binding:"required"`
	Objective      string     `json:"objective"`
	OrganizationID string     `json:"organization_id"`
	Status         string     `json:"status"`
	PrivacyLevel   string     `json:"privacy_level"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
}

type GlobalDecisionArgument struct {
	ID               string                 `json:"id"`
	RoomID           string                 `json:"room_id" binding:"required"`
	AuthorID         string                 `json:"author_id"`
	AuthorType       string                 `json:"author_type"`
	ArgumentText     string                 `json:"argument_text" binding:"required"`
	ArgumentType     string                 `json:"argument_type" binding:"required"`
	ParentArgumentID string                 `json:"parent_argument_id,omitempty"`
	Evidence         map[string]interface{} `json:"evidence"`
	Confidence       float64                `json:"confidence"`
	CreatedAt        *time.Time             `json:"created_at,omitempty"`
}

type GlobalDecisionForecast struct {
	ID            string                 `json:"id"`
	RoomID        string                 `json:"room_id" binding:"required"`
	PredictorID   string                 `json:"predictor_id"`
	PredictorType string                 `json:"predictor_type"`
	Question      string                 `json:"question" binding:"required"`
	ForecastValue float64                `json:"forecast_value"`
	Probability   float64                `json:"probability"`
	TimeHorizon   *time.Time             `json:"time_horizon"`
	Evidence      map[string]interface{} `json:"evidence"`
	Status        string                 `json:"status"`
	ActualOutcome map[string]interface{} `json:"actual_outcome,omitempty"`
}

type GlobalDecisionOption struct {
	ID             string                 `json:"id"`
	RoomID         string                 `json:"room_id" binding:"required"`
	ProposerID     string                 `json:"proposer_id"`
	OptionTitle    string                 `json:"option_title" binding:"required"`
	Description    string                 `json:"description"`
	Assumptions    map[string]interface{} `json:"assumptions"`
	RiskAssessment map[string]interface{} `json:"risk_assessment"`
	Status         string                 `json:"status"`
}

type GlobalDecisionJournal struct {
	ID               string                 `json:"id"`
	RoomID           string                 `json:"room_id" binding:"required"`
	DeciderID        string                 `json:"decider_id"`
	SelectedOptionID string                 `json:"selected_option_id"`
	Rationale        string                 `json:"rationale" binding:"required"`
	ExpectedOutcome  map[string]interface{} `json:"expected_outcome"`
	Postmortem       map[string]interface{} `json:"postmortem,omitempty"`
	Status           string                 `json:"status"`
	CreatedAt        *time.Time             `json:"created_at,omitempty"`
}
