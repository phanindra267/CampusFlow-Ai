package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type FeedbackHandler struct{}

func NewFeedbackHandler() *FeedbackHandler {
	return &FeedbackHandler{}
}

func (h *FeedbackHandler) SubmitFeedback(c *gin.Context) {
	var req domain.Feedback
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid feedback data", err)
		return
	}

	studentID, _ := c.Get("userID")
	req.StudentID = studentID.(string)

	response.Success(c, http.StatusCreated, "Feedback submitted successfully", req)
}
