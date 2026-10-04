package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type OperationsHandler struct{}

func NewOperationsHandler() *OperationsHandler {
	return &OperationsHandler{}
}

func (h *OperationsHandler) CreateScenario(c *gin.Context) {
	var req domain.OpsSimulationScenario
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid scenario configuration", err)
		return
	}

	userID, _ := c.Get("userID")
	req.CreatedBy = userID.(string)

	// Mock successful simulation run
	req.Results = map[string]interface{}{
		"optimized_venue":  "Main Auditorium",
		"conflict_avoided": true,
	}

	response.Success(c, http.StatusCreated, "Simulation scenario completed", req)
}

func (h *OperationsHandler) RequestAction(c *gin.Context) {
	var req domain.ActionExecution
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid action request", err)
		return
	}

	userID, _ := c.Get("userID")
	req.RequestedBy = userID.(string)
	req.Status = "PENDING_APPROVAL"

	response.Success(c, http.StatusCreated, "High-impact action queued for human approval", req)
}

func (h *OperationsHandler) ApproveAction(c *gin.Context) {
	actionID := c.Param("actionId")
	userID, _ := c.Get("userID")

	response.Success(c, http.StatusOK, "Action approved and executed safely", gin.H{
		"action_id":   actionID,
		"approved_by": userID,
		"status":      "EXECUTED",
	})
}

func (h *OperationsHandler) ListIncidents(c *gin.Context) {
	response.Success(c, http.StatusOK, "Operational incidents listed", []domain.OperationalIncident{
		{
			ID:       "inc-1",
			Severity: "HIGH",
			Status:   "INVESTIGATING",
			Symptoms: map[string]interface{}{"description": "Registration spike leading to DB latency"},
		},
	})
}
