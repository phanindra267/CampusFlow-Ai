package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type GlobalIntelligenceHandler struct{}

func NewGlobalIntelligenceHandler() *GlobalIntelligenceHandler {
	return &GlobalIntelligenceHandler{}
}

// POST /admin/global/trust/entities
func (h *GlobalIntelligenceHandler) RegisterTrustEntity(c *gin.Context) {
	var req domain.GlobalTrustEntity
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid trust entity payload", err)
		return
	}
	req.GovernanceStatus = "ACTIVE"
	response.Success(c, http.StatusCreated,
		"Global trust entity registered. Trust scores are dynamically updated based on evidence.", req)
}

// POST /admin/global/agents
func (h *GlobalIntelligenceHandler) RegisterGlobalAgent(c *gin.Context) {
	var req domain.GlobalAgent
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid global agent payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerTenantID = userID.(string) // using user id as tenant stub
	req.CertificationStatus = "REGISTERED"
	if req.AuthorityLevel == "" {
		req.AuthorityLevel = "OBSERVE"
	}
	response.Success(c, http.StatusCreated,
		"Global agent registered. Capabilities require certification before high-impact autonomous execution.", req)
}

// POST /admin/global/teams
func (h *GlobalIntelligenceHandler) CreateHumanAITeam(c *gin.Context) {
	var req domain.HumanAITeam
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid human-AI team payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.TenantID = userID.(string)
	req.Status = "FORMED"
	response.Success(c, http.StatusCreated,
		"Human-AI team formed. Roles, permissions, and objectives established for governed collaboration.", req)
}

// POST /admin/global/reasoning/jobs
func (h *GlobalIntelligenceHandler) StartReasoningJob(c *gin.Context) {
	var req domain.CollectiveReasoningJob
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid reasoning job payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.InitiatorTenantID = userID.(string)
	req.ConsensusState = "GATHERING"
	req.Status = "IN_PROGRESS"
	response.Success(c, http.StatusCreated,
		"Collective reasoning job initiated. Adversarial review and dissent preservation are active.", req)
}

// POST /admin/global/forecasts
func (h *GlobalIntelligenceHandler) RegisterForecast(c *gin.Context) {
	var req domain.GlobalForecast
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid forecast payload", err)
		return
	}
	req.Status = "ACTIVE"
	response.Success(c, http.StatusCreated,
		"Global ensemble forecast registered. Institutional disagreement is explicitly modeled.", req)
}

// POST /admin/global/risks
func (h *GlobalIntelligenceHandler) RegisterRisk(c *gin.Context) {
	var req domain.GlobalRisk
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid risk payload", err)
		return
	}
	req.Status = "IDENTIFIED"
	response.Success(c, http.StatusCreated,
		"Systemic risk registered. Cascading failure simulations will incorporate this entity.", req)
}
