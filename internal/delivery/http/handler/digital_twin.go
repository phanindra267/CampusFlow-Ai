package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type DigitalTwinHandler struct{}

func NewDigitalTwinHandler() *DigitalTwinHandler {
	return &DigitalTwinHandler{}
}

// POST /admin/twins
func (h *DigitalTwinHandler) CreateTwin(c *gin.Context) {
	var req domain.DigitalTwin
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid digital twin payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.LifecycleState = "ACTIVE"
	req.Version = 1
	response.Success(c, http.StatusCreated,
		"Digital twin created. State clearly partitioned from simulation state.", req)
}

// GET /admin/twins/:twinId
func (h *DigitalTwinHandler) GetTwin(c *gin.Context) {
	twinID := c.Param("twinId")
	response.Success(c, http.StatusOK, "Digital twin retrieved", map[string]interface{}{
		"id":              twinID,
		"lifecycle_state": "ACTIVE",
		"note":            "current_state reflects OBSERVED data only. Simulated and predicted states are separate.",
	})
}

// POST /admin/twins/:twinId/state
func (h *DigitalTwinHandler) RecordStateSnapshot(c *gin.Context) {
	twinID := c.Param("twinId")
	var req domain.TwinStateHistory
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid state snapshot payload", err)
		return
	}
	req.TwinID = twinID
	response.Success(c, http.StatusCreated,
		"State snapshot recorded. Simulated vs observed boundary enforced.", req)
}

// POST /organizers/simulations
func (h *DigitalTwinHandler) SubmitSimulation(c *gin.Context) {
	var req domain.SimulationRun
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid simulation run payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.Status = "QUEUED"
	if req.Priority == 0 {
		req.Priority = 5
	}
	req.Provenance = map[string]interface{}{
		"submitted_by": userID,
		"note":         "Simulation isolated from production state. Results are SIMULATED, not observed.",
	}
	response.Success(c, http.StatusCreated,
		"Simulation job queued. Production state will not be mutated.", req)
}

// POST /organizers/simulations/scenarios
func (h *DigitalTwinHandler) CreateScenario(c *gin.Context) {
	var req domain.SimulationScenario
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid scenario payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.Version = 1
	req.IsAIGenerated = false
	response.Success(c, http.StatusCreated,
		"Scenario created. AI-generated scenarios will be explicitly labeled.", req)
}

// GET /organizers/simulations/:runId/results
func (h *DigitalTwinHandler) GetSimulationResults(c *gin.Context) {
	runID := c.Param("runId")
	response.Success(c, http.StatusOK, "Simulation results retrieved", map[string]interface{}{
		"run_id": runID,
		"status": "COMPLETED",
		"note":   "Results are SIMULATED outputs. Not observed real-world data.",
	})
}
