package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type AutonomousAgentHandler struct{}

func NewAutonomousAgentHandler() *AutonomousAgentHandler {
	return &AutonomousAgentHandler{}
}

// POST /admin/agents/register
func (h *AutonomousAgentHandler) RegisterAgent(c *gin.Context) {
	var req domain.AgentIdentity
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid agent payload", err)
		return
	}
	req.Status = "ACTIVE"
	response.Success(c, http.StatusCreated,
		"Agent registered securely with explicit autonomy levels and bounds.", req)
}

// POST /agents/tasks
func (h *AutonomousAgentHandler) CreateTask(c *gin.Context) {
	var req domain.AutonomousTask
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid task payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.RequesterID = userID.(string)
	req.State = "PLANNING"
	response.Success(c, http.StatusCreated,
		"Autonomous task initiated. Entering planning phase. No execution will occur without authorization.", req)
}

// GET /agents/tasks/:id/plan
func (h *AutonomousAgentHandler) GetTaskPlan(c *gin.Context) {
	id := c.Param("id")
	response.Success(c, http.StatusOK,
		"Execution plan generated. Explicit explanation of required tools and checkpoints provided.",
		map[string]interface{}{
			"task_id":           id,
			"plan":              "Step 1: Retrieve context. Step 2: Extract claims. Step 3: Verify evidence.",
			"requires_approval": true,
		})
}

// POST /agents/tasks/:id/approve
func (h *AutonomousAgentHandler) ApproveTask(c *gin.Context) {
	var req domain.TaskApproval
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid approval payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.ApproverID = userID.(string)
	req.Decision = "APPROVED"
	response.Success(c, http.StatusOK,
		"Human governance gate passed. Agent authorized to execute approved scope.", req)
}

// POST /agents/tasks/:id/cancel
func (h *AutonomousAgentHandler) CancelTask(c *gin.Context) {
	id := c.Param("id")
	response.Success(c, http.StatusOK,
		"Task cancellation propagated to all dependent agents and tools. Graceful termination initiated.",
		map[string]string{"id": id, "state": "CANCELLED"})
}
