package handler

import (
	"net/http"

	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type AttendanceHandler struct{}

func NewAttendanceHandler() *AttendanceHandler {
	return &AttendanceHandler{}
}

func (h *AttendanceHandler) GenerateQR(c *gin.Context) {
	sessionID := c.Param("sessionId")
	organizerID, _ := c.Get("userID")

	response.Success(c, http.StatusCreated, "QR Check-in session generated", gin.H{
		"event_session_id": sessionID,
		"organizer_id":     organizerID,
		"qr_token":         "secure-qr-token-mock",
	})
}

func (h *AttendanceHandler) ScanQR(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid QR token", err)
		return
	}

	studentID, _ := c.Get("userID")

	response.Success(c, http.StatusOK, "Successfully checked in", gin.H{
		"student_id": studentID,
		"status":     "ATTENDED",
	})
}
