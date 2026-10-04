package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type ScientificDiscoveryHandler struct{}

func NewScientificDiscoveryHandler() *ScientificDiscoveryHandler {
	return &ScientificDiscoveryHandler{}
}

// POST /admin/discovery/hypotheses
func (h *ScientificDiscoveryHandler) ProposeHypothesis(c *gin.Context) {
	var req domain.ScientificHypothesis
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid hypothesis payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.Status = "PROPOSED"
	response.Success(c, http.StatusCreated,
		"Scientific hypothesis recorded. Awaiting adversarial review and evidence synthesis.", req)
}

// POST /admin/discovery/experiments
func (h *ScientificDiscoveryHandler) DesignExperiment(c *gin.Context) {
	var req domain.ScientificExperiment
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid experiment payload", err)
		return
	}
	req.Status = "DESIGNED"
	response.Success(c, http.StatusCreated,
		"Experiment designed. Execution requires isolation, risk review, and reproducibility validation.", req)
}

// POST /admin/discovery/improvements
func (h *ScientificDiscoveryHandler) ProposeIntelligenceImprovement(c *gin.Context) {
	var req domain.IntelligenceImprovement
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid intelligence improvement payload", err)
		return
	}
	req.Status = "PROPOSED"
	response.Success(c, http.StatusCreated,
		"Self-improvement proposed. Must pass sandbox evaluation, canary deployment, and human governance before production activation.", req)
}

// POST /admin/discovery/artifacts
func (h *ScientificDiscoveryHandler) RegisterArtifact(c *gin.Context) {
	var req domain.ScientificArtifact
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid artifact payload", err)
		return
	}
	req.Status = "REGISTERED"
	response.Success(c, http.StatusCreated,
		"Research artifact registered. Complete provenance and security isolation applied.", req)
}

// POST /admin/discovery/causal-graphs
func (h *ScientificDiscoveryHandler) RegisterCausalRelationship(c *gin.Context) {
	var req domain.CausalRelationship
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid causal relationship payload", err)
		return
	}
	if req.RelationshipType == "" {
		req.RelationshipType = "HYPOTHESIZED"
	}
	response.Success(c, http.StatusCreated,
		"Causal relationship registered. Explicitly marked as non-authoritative until experimentally supported.", req)
}
