package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type RecommendationHandler struct{}

func NewRecommendationHandler() *RecommendationHandler {
	return &RecommendationHandler{}
}

func (h *RecommendationHandler) GetPersonalizedFeed(c *gin.Context) {
	userID, _ := c.Get("userID")

	feed := []domain.Recommendation{
		{
			EntityID:   "mock-id",
			EntityType: "EVENT",
			Score:      0.95,
			Reason:     "Because you attended a similar Cloud event",
		},
	}

	response.Success(c, http.StatusOK, "Personalized feed generated", gin.H{
		"user_id": userID,
		"feed":    feed,
	})
}

func (h *RecommendationHandler) SubmitFeedback(c *gin.Context) {
	var req domain.RecommendationFeedback
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid feedback data", err)
		return
	}

	userID, _ := c.Get("userID")
	req.UserID = userID.(string)

	response.Success(c, http.StatusCreated, "Recommendation feedback recorded", req)
}
