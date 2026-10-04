package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type AdminInsightsHandler struct{}

func NewAdminInsightsHandler() *AdminInsightsHandler {
	return &AdminInsightsHandler{}
}

func (h *AdminInsightsHandler) GetKnowledgeGaps(c *gin.Context) {
	gaps := []domain.KnowledgeGap{
		{
			ID:           "gap-1",
			QueryPattern: "When are AI internships open?",
			Frequency:    42,
			Status:       "DETECTED",
		},
	}

	response.Success(c, http.StatusOK, "Knowledge gaps retrieved", gaps)
}
