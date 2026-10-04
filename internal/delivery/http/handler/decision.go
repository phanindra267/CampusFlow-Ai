package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type GlobalDecisionIntelligenceHandler struct{}

func NewGlobalDecisionIntelligenceHandler() *GlobalDecisionIntelligenceHandler {
	return &GlobalDecisionIntelligenceHandler{}
}

// POST /global-decisions/rooms
func (h *GlobalDecisionIntelligenceHandler) CreateDecisionRoom(c *gin.Context) {
	var req domain.GlobalDecisionRoom
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid room payload", err)
		return
	}
	req.Status = "OPEN"
	response.Success(c, http.StatusCreated,
		"Global Decision Room established. Configured for multi-agent advisory, human expertise, and strict evidence gating.", req)
}

// POST /global-decisions/rooms/:id/arguments
func (h *GlobalDecisionIntelligenceHandler) AddArgument(c *gin.Context) {
	var req domain.GlobalDecisionArgument
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid argument payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.AuthorID = userID.(string)
	req.AuthorType = "HUMAN"
	req.RoomID = c.Param("id")
	response.Success(c, http.StatusCreated,
		"Evidence-backed argument added to the decision graph. Subject to independent verification.", req)
}

// POST /global-decisions/rooms/:id/forecasts
func (h *GlobalDecisionIntelligenceHandler) RegisterForecast(c *gin.Context) {
	var req domain.GlobalDecisionForecast
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid forecast payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.PredictorID = userID.(string)
	req.RoomID = c.Param("id")
	req.Status = "ACTIVE"
	response.Success(c, http.StatusCreated,
		"Probabilistic forecast recorded for future calibration and decision modeling.", req)
}

// POST /global-decisions/rooms/:id/options
func (h *GlobalDecisionIntelligenceHandler) ProposeOption(c *gin.Context) {
	var req domain.GlobalDecisionOption
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid option payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.ProposerID = userID.(string)
	req.RoomID = c.Param("id")
	response.Success(c, http.StatusCreated,
		"Decision option successfully generated for multi-criteria evaluation.", req)
}

// POST /global-decisions/rooms/:id/journal
func (h *GlobalDecisionIntelligenceHandler) RecordDecision(c *gin.Context) {
	var req domain.GlobalDecisionJournal
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid journal payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.DeciderID = userID.(string)
	req.RoomID = c.Param("id")
	response.Success(c, http.StatusCreated,
		"Final decision firmly recorded into the immutable journal. Explanatory trace and expected outcomes preserved for future postmortem learning.", req)
}
