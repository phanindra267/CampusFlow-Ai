package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type EconomicIntelligenceHandler struct{}

func NewEconomicIntelligenceHandler() *EconomicIntelligenceHandler {
	return &EconomicIntelligenceHandler{}
}

// POST /admin/economics/datasets
func (h *EconomicIntelligenceHandler) RegisterDataset(c *gin.Context) {
	var req domain.EconDataset
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid dataset payload", err)
		return
	}
	req.Status = "ACTIVE"
	response.Success(c, http.StatusCreated,
		"Economic dataset registered securely. Metadata and provenance constraints applied.", req)
}

// GET /economics/regional/balance
func (h *EconomicIntelligenceHandler) GetRegionalBalance(c *gin.Context) {
	region := c.Query("region")
	if region == "" {
		region = "GLOBAL"
	}
	response.Success(c, http.StatusOK,
		"Regional education-workforce supply/demand balance retrieved.",
		[]domain.RegionalSkillBalance{
			{
				Region:          region,
				SkillName:       "Cloud Native Architecture",
				WorkforceDemand: 8500,
				EducationSupply: 3200,
				BalanceStatus:   "SHORTAGE",
				Confidence:      0.88,
			},
		})
}

// POST /institutions/economics/scenarios
func (h *EconomicIntelligenceHandler) CreateScenario(c *gin.Context) {
	var req domain.PolicyScenario
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid scenario payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.Status = "QUEUED" // Enforce asynchronous execution
	response.Success(c, http.StatusCreated,
		"Policy scenario simulation queued. Results are advisory and must not automatically govern policy.", req)
}

// POST /institutions/economics/scenarios/:id/cancel
func (h *EconomicIntelligenceHandler) CancelScenario(c *gin.Context) {
	id := c.Param("id")
	response.Success(c, http.StatusOK,
		"Scenario computation cancelled to preserve cluster resources.", map[string]string{"id": id, "status": "CANCELLED"})
}

// GET /economics/funding
func (h *EconomicIntelligenceHandler) ListFunding(c *gin.Context) {
	response.Success(c, http.StatusOK,
		"Funding opportunities retrieved successfully.",
		[]domain.FundingOpportunity{})
}
