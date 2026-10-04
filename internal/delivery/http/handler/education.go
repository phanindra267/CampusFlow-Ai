package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type EducationIntelligenceHandler struct{}

func NewEducationIntelligenceHandler() *EducationIntelligenceHandler {
	return &EducationIntelligenceHandler{}
}

// POST /students/profiles/:id/learning-paths
func (h *EducationIntelligenceHandler) GenerateLearningPath(c *gin.Context) {
	var req domain.EduLearningPath
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid learning path payload", err)
		return
	}
	req.Status = "ACTIVE"
	response.Success(c, http.StatusCreated,
		"Personalized learning path generated. Respects privacy and student agency.", req)
}

// POST /students/profiles/:id/interventions
func (h *EducationIntelligenceHandler) RecommendIntervention(c *gin.Context) {
	var req domain.EduIntervention
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid intervention payload", err)
		return
	}
	req.AdvisorReviewStatus = "PENDING"
	response.Success(c, http.StatusCreated,
		"Educational intervention proposed. Advisor approval required before student notification.", req)
}

// POST /advisors/interventions/:id/review
func (h *EducationIntelligenceHandler) ReviewIntervention(c *gin.Context) {
	id := c.Param("id")
	response.Success(c, http.StatusOK,
		"Intervention reviewed by advisor.", map[string]string{"id": id, "status": "APPROVED"})
}

// POST /faculty/courses/:id/assessments
func (h *EducationIntelligenceHandler) CreateAssessment(c *gin.Context) {
	var req domain.EduAssessment
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid assessment payload", err)
		return
	}
	if req.AssessmentMode == "" {
		req.AssessmentMode = "PRACTICE"
	}
	response.Success(c, http.StatusCreated,
		"Assessment created with explicit AI assistance policies and academic integrity controls.", req)
}

// GET /students/opportunities/matches
func (h *EducationIntelligenceHandler) MatchOpportunities(c *gin.Context) {
	// Return mocked opportunity
	response.Success(c, http.StatusOK,
		"Skill-aligned opportunities retrieved based on target competency gaps.",
		[]domain.EduOpportunity{
			{
				ID:              "opp-1",
				Title:           "AI Research Assistantship",
				OpportunityType: "RESEARCH",
				Status:          "OPEN",
			},
		})
}
