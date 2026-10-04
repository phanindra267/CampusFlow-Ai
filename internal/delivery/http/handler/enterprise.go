package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type EnterpriseHandler struct{}

func NewEnterpriseHandler() *EnterpriseHandler {
	return &EnterpriseHandler{}
}

func (h *EnterpriseHandler) RegisterAgent(c *gin.Context) {
	var req domain.AgentDefinition
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid agent definition", err)
		return
	}

	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.Status = "DRAFT"

	response.Success(c, http.StatusCreated, "Agent registered successfully. Pending approval for production deployment.", req)
}

func (h *EnterpriseHandler) ApproveAgent(c *gin.Context) {
	agentID := c.Param("agentId")
	response.Success(c, http.StatusOK, "Agent approved for execution", gin.H{
		"agent_id": agentID,
		"status":   "ACTIVE",
	})
}

func (h *EnterpriseHandler) ListConnectors(c *gin.Context) {
	response.Success(c, http.StatusOK, "Enterprise connectors listed", []domain.EnterpriseConnector{
		{
			ID:           "conn-1",
			ProviderName: "CANVAS_LMS",
			Status:       "ACTIVE",
		},
	})
}

func (h *EnterpriseHandler) SyncConnector(c *gin.Context) {
	connID := c.Param("connectorId")
	response.Success(c, http.StatusOK, "Synchronization triggered successfully", gin.H{
		"connector_id": connID,
		"status":       "SYNCING",
	})
}

func (h *EnterpriseHandler) GetFederatedAnalytics(c *gin.Context) {
	response.Success(c, http.StatusOK, "Federated institutional analytics retrieved", gin.H{
		"aggregated_active_users": 15400,
		"campuses_included":       3,
		"cross_campus_events":     12,
	})
}
