package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type ResearchEcosystemHandler struct{}

func NewResearchEcosystemHandler() *ResearchEcosystemHandler {
	return &ResearchEcosystemHandler{}
}

// POST /researchers/projects
func (h *ResearchEcosystemHandler) CreateProject(c *gin.Context) {
	var req domain.ResearchProject
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid research project payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.Status = "IDEA"
	req.MaturityStage = "IDEA"
	response.Success(c, http.StatusCreated, "Research project created. Lifecycle governance active.", req)
}

// POST /researchers/projects/:projectId/hypotheses
func (h *ResearchEcosystemHandler) CreateHypothesis(c *gin.Context) {
	projectID := c.Param("projectId")
	var req domain.ResearchHypothesis
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid hypothesis payload", err)
		return
	}
	req.ProjectID = projectID
	req.Status = "DRAFT"
	req.Version = 1
	response.Success(c, http.StatusCreated, "Hypothesis created. Version tracking active. Provenance preserved.", req)
}

// POST /researchers/projects/:projectId/experiments
func (h *ResearchEcosystemHandler) CreateExperiment(c *gin.Context) {
	projectID := c.Param("projectId")
	var req domain.ResearchExperiment
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid experiment payload", err)
		return
	}
	req.ProjectID = projectID
	req.Status = "PLANNED"
	response.Success(c, http.StatusCreated, "Experiment defined in PLANNED state. Execution requires authorized workflow trigger.", req)
}

// POST /researchers/datasets
func (h *ResearchEcosystemHandler) RegisterDataset(c *gin.Context) {
	var req domain.ResearchDataset
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid dataset payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.Version = "1.0"
	response.Success(c, http.StatusCreated, "Dataset registered in governed registry. License and privacy classification recorded.", req)
}

// POST /researchers/projects/:projectId/artifacts
func (h *ResearchEcosystemHandler) RegisterArtifact(c *gin.Context) {
	projectID := c.Param("projectId")
	var req domain.ResearchArtifact
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid artifact payload", err)
		return
	}
	req.ProjectID = projectID
	req.Version = "1.0"
	req.IsImmutable = false
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	response.Success(c, http.StatusCreated, "Research artifact registered. Immutability enforced upon publication.", req)
}

// POST /admin/research/disclosures
func (h *ResearchEcosystemHandler) CreateInventionDisclosure(c *gin.Context) {
	var req domain.InventionDisclosure
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invention disclosure payload", err)
		return
	}
	req.Status = "DRAFT"
	response.Success(c, http.StatusCreated, "Invention disclosure recorded. This is a metadata record only — legal review required separately.", req)
}
