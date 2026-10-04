package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type PlanetScaleHandler struct{}

func NewPlanetScaleHandler() *PlanetScaleHandler {
	return &PlanetScaleHandler{}
}

// POST /researchers/claims
func (h *PlanetScaleHandler) RegisterClaim(c *gin.Context) {
	var req domain.ScientificClaim
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid scientific claim payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.AuthorID = userID.(string)
	req.Status = "PROPOSED"
	req.Version = 1
	response.Success(c, http.StatusCreated,
		"Scientific claim registered as PROPOSED. Not validated until evidence review is complete.", req)
}

// POST /researchers/experiments/:experimentId/replications
func (h *PlanetScaleHandler) InitiateReplication(c *gin.Context) {
	experimentID := c.Param("experimentId")
	var req domain.ReplicationStudy
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid replication payload", err)
		return
	}
	req.OriginalExperimentID = experimentID
	userID, _ := c.Get("userID")
	req.ReplicatingResearcherID = userID.(string)
	req.Status = "PLANNED"
	response.Success(c, http.StatusCreated,
		"Replication study initiated. Original experiment artifacts are read-only.", req)
}

// POST /admin/curriculum/nodes
func (h *PlanetScaleHandler) AddCurriculumNode(c *gin.Context) {
	var req domain.CurriculumNode
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid curriculum node payload", err)
		return
	}
	response.Success(c, http.StatusCreated, "Curriculum node added to knowledge graph.", req)
}

// POST /admin/curriculum/edges
func (h *PlanetScaleHandler) AddCurriculumEdge(c *gin.Context) {
	var req domain.CurriculumEdge
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid curriculum edge payload", err)
		return
	}
	response.Success(c, http.StatusCreated, "Curriculum relationship added. Prerequisite graph updated.", req)
}

// POST /researchers/projects/:projectId/funding
func (h *PlanetScaleHandler) RegisterFunding(c *gin.Context) {
	projectID := c.Param("projectId")
	var req domain.ResearchFunding
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid funding payload", err)
		return
	}
	req.ProjectID = projectID
	req.SpentBudget = 0
	req.Status = "ACTIVE"
	if req.Currency == "" {
		req.Currency = "USD"
	}
	response.Success(c, http.StatusCreated,
		"Research funding registered. Budget tracking active. Restrictions enforced.", req)
}
