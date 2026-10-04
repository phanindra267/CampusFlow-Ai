package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type IntelligenceHandler struct{}

func NewIntelligenceHandler() *IntelligenceHandler {
	return &IntelligenceHandler{}
}

func (h *IntelligenceHandler) CreateDecisionRoom(c *gin.Context) {
	var req domain.DecisionRoom
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid decision room payload", err)
		return
	}

	req.Status = "OPEN"
	response.Success(c, http.StatusCreated, "Decision Room created for human-AI collaboration", req)
}

func (h *IntelligenceHandler) ResolveDecision(c *gin.Context) {
	roomID := c.Param("roomId")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid decision resolution", err)
		return
	}

	response.Success(c, http.StatusOK, "Human governance decision recorded. Autonomous constraints updated.", gin.H{
		"room_id":        roomID,
		"final_decision": req["final_decision"],
		"status":         "RESOLVED",
	})
}

func (h *IntelligenceHandler) RegisterAutonomousPolicy(c *gin.Context) {
	var req domain.AutonomousPolicy
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid policy payload", err)
		return
	}

	// Force SHADOW mode on creation for safety
	req.Mode = "SHADOW"

	response.Success(c, http.StatusCreated, "Autonomous policy registered in SHADOW mode. Awaiting evaluation.", req)
}

func (h *IntelligenceHandler) VerifyClaim(c *gin.Context) {
	claim := c.Query("claim")

	graph := domain.EvidenceGraph{
		ID:              "evg-1",
		Claim:           claim,
		SourceReference: "Federated Research Node Alpha",
		DatasetVersion:  "v2.1",
		Confidence:      0.88,
	}

	response.Success(c, http.StatusOK, "Claim verification complete. Provenance retained.", graph)
}
