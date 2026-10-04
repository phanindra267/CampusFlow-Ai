package domain

import "time"

type AgentIdentity struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name" binding:"required"`
	Version            string                 `json:"version" binding:"required"`
	Purpose            string                 `json:"purpose"`
	OwnerID            string                 `json:"owner_id"`
	Capabilities       []string               `json:"capabilities"`
	PermissionSet      map[string]interface{} `json:"permission_set"`
	RiskClassification string                 `json:"risk_classification"`
	MaxAutonomyLevel   int                    `json:"max_autonomy_level"`
	Status             string                 `json:"status"`
}

type ToolDefinition struct {
	ID               string                 `json:"id"`
	ToolName         string                 `json:"tool_name" binding:"required"`
	Version          string                 `json:"version" binding:"required"`
	Description      string                 `json:"description"`
	InputSchema      map[string]interface{} `json:"input_schema"`
	OutputSchema     map[string]interface{} `json:"output_schema"`
	Permissions      []string               `json:"permissions"`
	RiskLevel        string                 `json:"risk_level"`
	ResourceLimits   map[string]interface{} `json:"resource_limits"`
	RequiresApproval bool                   `json:"requires_approval"`
}

type AutonomousTask struct {
	ID               string                 `json:"id"`
	RequesterID      string                 `json:"requester_id"`
	Objective        string                 `json:"objective" binding:"required"`
	Priority         string                 `json:"priority"`
	RiskLevel        string                 `json:"risk_level"`
	Budget           map[string]interface{} `json:"budget"`
	State            string                 `json:"state"`
	AssignedAgents   []string               `json:"assigned_agents"`
	ExecutionPlan    map[string]interface{} `json:"execution_plan"`
	ExecutionHistory []interface{}          `json:"execution_history"`
	TraceID          string                 `json:"trace_id"`
	Deadline         *time.Time             `json:"deadline,omitempty"`
}

type TaskApproval struct {
	ID         string                 `json:"id"`
	TaskID     string                 `json:"task_id" binding:"required"`
	ApproverID string                 `json:"approver_id"`
	Scope      map[string]interface{} `json:"scope"`
	Decision   string                 `json:"decision"`
	Reason     string                 `json:"reason"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
}

type AgentMemory struct {
	ID         string                 `json:"id"`
	AgentID    string                 `json:"agent_id" binding:"required"`
	TaskID     string                 `json:"task_id"`
	MemoryType string                 `json:"memory_type"`
	Content    map[string]interface{} `json:"content"`
	Confidence float64                `json:"confidence"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
}
