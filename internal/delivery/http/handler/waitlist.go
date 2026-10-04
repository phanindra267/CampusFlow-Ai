package handler

import (
	"net/http"

	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type WaitlistHandler struct{}

func NewWaitlistHandler() *WaitlistHandler {
	return &WaitlistHandler{}
}

func (h *WaitlistHandler) JoinWaitlist(c *gin.Context) {
	eventID := c.Param("id")
	studentID, _ := c.Get("userID")

	response.Success(c, http.StatusOK, "Successfully joined waitlist", gin.H{
		"event_id":   eventID,
		"student_id": studentID,
		"status":     "WAITING",
	})
}
