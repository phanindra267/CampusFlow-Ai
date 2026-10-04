package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type ResourceHandler struct{}

func NewResourceHandler() *ResourceHandler {
	return &ResourceHandler{}
}

func (h *ResourceHandler) ListResources(c *gin.Context) {
	resources := []domain.CampusResource{
		{
			ID:           "mock-res-1",
			Name:         "Main Auditorium",
			ResourceType: "VENUE",
			Capacity:     500,
			Status:       "AVAILABLE",
		},
	}

	response.Success(c, http.StatusOK, "Campus resources listed", resources)
}
