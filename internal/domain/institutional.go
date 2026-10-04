package domain

import "time"

type GraphNode struct {
	ID         string                 `json:"id"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Label      string                 `json:"label"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type GraphEdge struct {
	ID               string  `json:"id"`
	SourceNodeID     string  `json:"source_node_id"`
	TargetNodeID     string  `json:"target_node_id"`
	RelationshipType string  `json:"relationship_type"`
	Weight           float64 `json:"weight"`
}

type CampusResource struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	ResourceType string                 `json:"resource_type"`
	Capacity     int                    `json:"capacity"`
	Status       string                 `json:"status"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type Skill struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type InstitutionalMetric struct {
	MetricName  string                 `json:"metric_name"`
	MetricValue float64                `json:"metric_value"`
	Dimensions  map[string]interface{} `json:"dimensions"`
	PeriodStart time.Time              `json:"period_start"`
	PeriodEnd   time.Time              `json:"period_end"`
}
