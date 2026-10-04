package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type AutomationHandler struct{}

func NewAutomationHandler() *AutomationHandler {
	return &AutomationHandler{}
}

func (h *AutomationHandler) CreateRule(c *gin.Context) {
	var req domain.AutomationRule
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid automation rule configuration", err)
		return
	}

	userID, _ := c.Get("userID")
	req.UserID = userID.(string)
	req.Status = "ACTIVE"

	response.Success(c, http.StatusCreated, "Automation rule created successfully", req)
}

func (h *AutomationHandler) UpdatePreferences(c *gin.Context) {
	var req domain.NotificationPreference
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid preference configuration", err)
		return
	}

	userID, _ := c.Get("userID")
	req.UserID = userID.(string)

	response.Success(c, http.StatusOK, "Notification preferences updated", req)
}
