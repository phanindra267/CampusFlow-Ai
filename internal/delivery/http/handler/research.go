package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type ResearchHandler struct{}

func NewResearchHandler() *ResearchHandler {
	return &ResearchHandler{}
}

func (h *ResearchHandler) DiscoverResearch(c *gin.Context) {
	pubs := []domain.ResearchPublication{
		{
			ID:    "res-1",
			Title: "Federated Learning over Edge Devices in Smart Campuses",
			DOI:   "10.1234/smart.campus.2026",
		},
	}

	response.Success(c, http.StatusOK, "Research intelligence discovery retrieved", pubs)
}
