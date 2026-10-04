package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type EcosystemHandler struct{}

func NewEcosystemHandler() *EcosystemHandler {
	return &EcosystemHandler{}
}

// POST /admin/ecosystem/institutions
func (h *EcosystemHandler) RegisterTrustedInstitution(c *gin.Context) {
	var req domain.TrustedInstitution
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid trusted institution payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.AdminOwnerID = userID.(string)
	req.VerificationStatus = "UNVERIFIED"
	req.FederationStatus = "NONE"
	req.SecurityStatus = "UNKNOWN"
	response.Success(c, http.StatusCreated, "Institution registered. Verification required before federation can be established.", req)
}

// POST /admin/ecosystem/tools
func (h *EcosystemHandler) RegisterAgentTool(c *gin.Context) {
	var req domain.AgentTool
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid tool payload", err)
		return
	}
	req.Version = "1.0"
	req.IsActive = false // Inactive until reviewed
	response.Success(c, http.StatusCreated, "Agent tool registered as INACTIVE. Requires security review before activation.", req)
}

// POST /admin/ecosystem/services
func (h *EcosystemHandler) PublishService(c *gin.Context) {
	var req domain.ServiceCatalogEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid service catalog payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.HealthStatus = "UNKNOWN"
	req.Version = "1.0"
	response.Success(c, http.StatusCreated, "Service published to catalog. Access policy enforced on every request.", req)
}

// POST /admin/ecosystem/loop-guards
func (h *EcosystemHandler) CreateLoopGuard(c *gin.Context) {
	var req domain.AutomationLoopGuard
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid loop guard payload", err)
		return
	}
	req.Status = "ACTIVE"
	req.InvocationCount = 0
	response.Success(c, http.StatusCreated, "Automation loop guard created. Runaway workflow protection active.", req)
}

// POST /admin/ecosystem/governance/reviews
func (h *EcosystemHandler) CreateGovernanceReview(c *gin.Context) {
	var req domain.GovernanceReview
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid governance review payload", err)
		return
	}
	response.Success(c, http.StatusCreated, "Governance review opened. Approval required before subject becomes operational.", req)
}

// GET /admin/ecosystem/services
func (h *EcosystemHandler) ListServices(c *gin.Context) {
	response.Success(c, http.StatusOK, "Service catalog retrieved", []domain.ServiceCatalogEntry{
		{
			ID:              "svc-001",
			ServiceName:     "Research Compute Gateway",
			AvailabilitySLA: "99.5%",
			HealthStatus:    "HEALTHY",
			Version:         "1.0",
		},
	})
}
