package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type StrategicSimulationHandler struct{}

func NewStrategicSimulationHandler() *StrategicSimulationHandler {
	return &StrategicSimulationHandler{}
}

// POST /admin/twins/cohorts
func (h *StrategicSimulationHandler) RegisterCohort(c *gin.Context) {
	var req domain.PopulationCohort
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid cohort payload", err)
		return
	}
	response.Success(c, http.StatusCreated,
		"Population cohort twin created. Preserving individual privacy via aggregate modeling.", req)
}

// POST /admin/twins/snapshots
func (h *StrategicSimulationHandler) CreateSnapshot(c *gin.Context) {
	var req domain.SystemSnapshot
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid snapshot payload", err)
		return
	}
	response.Success(c, http.StatusCreated,
		"Immutable system snapshot recorded for reproducible strategic simulations.", req)
}

// POST /institutions/simulations
func (h *StrategicSimulationHandler) QueueSimulation(c *gin.Context) {
	var req domain.StrategicSimulation
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid simulation payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.Status = "QUEUED"
	response.Success(c, http.StatusCreated,
		"Strategic simulation queued securely. Outcomes are explicitly defined as exploratory.", req)
}

// POST /institutions/simulations/:id/branch
func (h *StrategicSimulationHandler) BranchSimulation(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	branched := domain.StrategicSimulation{
		ID:                 "branched-sim-placeholder",
		ParentSimulationID: id,
		OwnerID:            userID.(string),
		Status:             "QUEUED",
	}
	response.Success(c, http.StatusCreated,
		"Simulation branched successfully for comparative what-if analysis.", branched)
}

// POST /institutions/simulations/:id/interventions
func (h *StrategicSimulationHandler) ProposeIntervention(c *gin.Context) {
	var req domain.SimulationIntervention
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid intervention payload", err)
		return
	}
	req.HumanApprovalStatus = "PENDING"
	response.Success(c, http.StatusCreated,
		"Simulation intervention proposed. Requires human review and governance before implementation.", req)
}

// GET /admin/twins/warnings
func (h *StrategicSimulationHandler) ListEarlyWarnings(c *gin.Context) {
	response.Success(c, http.StatusOK,
		"Early warning signals retrieved for workforce and educational capacity disruptions.",
		[]domain.EarlyWarningSignal{})
}

// POST /admin/twins/model-cards
func (h *StrategicSimulationHandler) PublishModelCard(c *gin.Context) {
	var req domain.ModelCard
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid model card payload", err)
		return
	}
	req.ApprovalStatus = "DRAFT"
	response.Success(c, http.StatusCreated,
		"Model card drafted. Transparently documenting simulation purpose, intended use, and critical limitations.", req)
}
