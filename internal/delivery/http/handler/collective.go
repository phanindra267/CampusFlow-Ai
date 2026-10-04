package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type CollectiveHandler struct{}

func NewCollectiveHandler() *CollectiveHandler {
	return &CollectiveHandler{}
}

// POST /organizers/collective/rooms
func (h *CollectiveHandler) OpenDeliberationRoom(c *gin.Context) {
	var req domain.DeliberationRoom
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid deliberation room payload", err)
		return
	}

	userID, _ := c.Get("userID")
	req.DecisionOwnerID = userID.(string)
	req.Status = "OPEN"

	response.Success(c, http.StatusCreated, "Deliberation room opened. Human decision owner assigned.", req)
}

// POST /organizers/collective/rooms/:roomId/arguments
func (h *CollectiveHandler) AddArgument(c *gin.Context) {
	roomID := c.Param("roomId")
	var req domain.ArgumentGraph
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid argument graph payload", err)
		return
	}

	req.RoomID = roomID
	userID, _ := c.Get("userID")
	req.ContributedBy = userID.(string)

	response.Success(c, http.StatusCreated, "Argument added to deliberation room. Provenance preserved.", req)
}

// POST /admin/collective/rooms/:roomId/decide
func (h *CollectiveHandler) RecordDecisionOutcome(c *gin.Context) {
	roomID := c.Param("roomId")
	var req domain.DecisionOutcome
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid decision outcome payload", err)
		return
	}

	req.RoomID = roomID
	req.Status = "PENDING"

	response.Success(c, http.StatusCreated, "Decision outcome recorded. Pending real-world observation for learning loop.", req)
}

// POST /admin/collective/playbooks
func (h *CollectiveHandler) CreatePlaybook(c *gin.Context) {
	var req domain.InstitutionalPlaybook
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid playbook payload", err)
		return
	}

	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.Status = "DRAFT"
	req.Version = "1.0"

	response.Success(c, http.StatusCreated, "Playbook created in DRAFT status. Requires approval before activation.", req)
}

// GET /admin/collective/predictions
func (h *CollectiveHandler) GetPredictionLedger(c *gin.Context) {
	response.Success(c, http.StatusOK, "Prediction ledger retrieved", []domain.PredictionLedgerEntry{
		{
			ID:           "pred-1",
			ModelName:    "AttendanceForecastV2",
			ModelVersion: "2.3.1",
			Confidence:   0.82,
			Prediction:   map[string]interface{}{"expected_attendance": 450},
		},
	})
}
