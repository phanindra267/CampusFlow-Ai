package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type OpportunityHandler struct{}

func NewOpportunityHandler() *OpportunityHandler {
	return &OpportunityHandler{}
}

func (h *OpportunityHandler) CreateOpportunity(c *gin.Context) {
	var req domain.Opportunity
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid opportunity data", err)
		return
	}

	organizerID, _ := c.Get("userID")
	req.OrganizerID = organizerID.(string)

	response.Success(c, http.StatusCreated, "Opportunity created successfully", req)
}

func (h *OpportunityHandler) ListOpportunities(c *gin.Context) {
	response.Success(c, http.StatusOK, "Opportunities retrieved", []domain.Opportunity{})
}

func (h *OpportunityHandler) SaveOpportunity(c *gin.Context) {
	oppID := c.Param("id")
	userID, _ := c.Get("userID")

	response.Success(c, http.StatusOK, "Opportunity saved successfully", gin.H{
		"opportunity_id": oppID,
		"user_id":        userID,
	})
}
