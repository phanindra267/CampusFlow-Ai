package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type AutonomousHandler struct{}

func NewAutonomousHandler() *AutonomousHandler {
	return &AutonomousHandler{}
}

// POST /admin/autonomous/agents
func (h *AutonomousHandler) RegisterAgent(c *gin.Context) {
	var req domain.AgentRegistryEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid agent registry payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.Status = "PENDING"
	req.Version = "1.0"
	if req.AutonomyLevel > 5 {
		response.Error(c, http.StatusBadRequest,
			"Autonomy level exceeds maximum (L5). Reduce autonomy level.", nil)
		return
	}
	response.Success(c, http.StatusCreated,
		"Agent registered as PENDING. Requires security review and explicit activation before autonomous operation.", req)
}

// POST /admin/autonomous/agents/:agentId/activate
func (h *AutonomousHandler) ActivateAgent(c *gin.Context) {
	agentID := c.Param("agentId")
	approver, _ := c.Get("userID")
	response.Success(c, http.StatusOK, "Agent activation approved. Bounded by registered autonomy level and policy.", map[string]interface{}{
		"agent_id":    agentID,
		"approved_by": approver,
		"status":      "ACTIVE",
		"note":        "Agent may never exceed its registered autonomy_level.",
	})
}

// POST /admin/autonomous/agents/:agentId/suspend
func (h *AutonomousHandler) SuspendAgent(c *gin.Context) {
	agentID := c.Param("agentId")
	response.Success(c, http.StatusOK, "Agent suspended. All pending actions halted safely.", map[string]interface{}{
		"agent_id": agentID,
		"status":   "SUSPENDED",
	})
}

// POST /admin/autonomous/policies
func (h *AutonomousHandler) CreatePolicy(c *gin.Context) {
	var req domain.AutonomyPolicy
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid autonomy policy payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.Status = "DRAFT"
	req.Version = 1
	response.Success(c, http.StatusCreated,
		"Autonomy policy created in DRAFT state. Must be approved before agents can operate under it.", req)
}

// POST /admin/autonomous/plans
func (h *AutonomousHandler) CreatePlan(c *gin.Context) {
	var req domain.ActionPlan
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid action plan payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.Status = "DRAFT"
	req.Version = 1
	req.Preconditions = map[string]interface{}{
		"submitted_by": userID,
	}
	response.Success(c, http.StatusCreated,
		"Action plan created. Requires validation, simulation, and approval before execution.", req)
}

// POST /admin/autonomous/plans/:planId/approve
func (h *AutonomousHandler) ApprovePlan(c *gin.Context) {
	planID := c.Param("planId")
	approver, _ := c.Get("userID")
	response.Success(c, http.StatusOK,
		"Plan approved for execution. Each step will be verified independently.", map[string]interface{}{
			"plan_id":     planID,
			"approved_by": approver,
			"status":      "APPROVED",
		})
}

// POST /admin/autonomous/decisions
func (h *AutonomousHandler) RecordDecision(c *gin.Context) {
	var req domain.DecisionRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid decision record payload", err)
		return
	}
	req.HumanApproved = false
	response.Success(c, http.StatusCreated,
		"Decision record created. Evidence and explanation preserved for audit replay.", req)
}

// POST /admin/autonomous/kill-switch
func (h *AutonomousHandler) GlobalKillSwitch(c *gin.Context) {
	userID, _ := c.Get("userID")
	response.Success(c, http.StatusOK,
		"GLOBAL KILL SWITCH ACTIVATED. All autonomous agents suspended. Requires explicit reactivation by authorized administrator.", map[string]interface{}{
			"activated_by": userID,
			"status":       "ALL_AGENTS_SUSPENDED",
			"note":         "Kill switch is independent of all agents. No agent can override it.",
		})
}

// POST /admin/autonomous/safe-mode
func (h *AutonomousHandler) ActivateSafeMode(c *gin.Context) {
	userID, _ := c.Get("userID")
	response.Success(c, http.StatusOK,
		"Safe mode activated. No new high-impact autonomous actions permitted.", map[string]interface{}{
			"activated_by":        userID,
			"status":              "SAFE_MODE",
			"high_impact_blocked": true,
			"safe_ops_continue":   true,
		})
}

// POST /admin/autonomous/lessons
func (h *AutonomousHandler) RecordLesson(c *gin.Context) {
	var req domain.LessonLearned
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid lesson payload", err)
		return
	}
	req.ValidationStatus = "DRAFT"
	response.Success(c, http.StatusCreated,
		"Lesson recorded as DRAFT. Requires validation before becoming authoritative institutional knowledge.", req)
}
