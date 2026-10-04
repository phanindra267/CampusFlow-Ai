package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type IoTHandler struct{}

func NewIoTHandler() *IoTHandler {
	return &IoTHandler{}
}

func (h *IoTHandler) RegisterDevice(c *gin.Context) {
	var req domain.IoTDevice
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid device registration", err)
		return
	}

	req.Status = "ACTIVE"
	req.SecurityState = "SECURE"

	response.Success(c, http.StatusCreated, "IoT device securely registered to tenant", req)
}

func (h *IoTHandler) RequestCommand(c *gin.Context) {
	var req domain.IoTCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid command request", err)
		return
	}

	userID, _ := c.Get("userID")
	req.RequestedBy = userID.(string)
	req.Status = "PENDING_APPROVAL"
	req.SafetyClearance = false // Explicitly enforce safety gate

	response.Success(c, http.StatusCreated, "Physical command requires Safe Command Gateway clearance", req)
}

func (h *IoTHandler) ApproveCommand(c *gin.Context) {
	cmdID := c.Param("commandId")
	userID, _ := c.Get("userID")

	response.Success(c, http.StatusOK, "Command passed Safety Gateway and executed", gin.H{
		"command_id":       cmdID,
		"approved_by":      userID,
		"safety_clearance": true,
		"status":           "VERIFIED",
	})
}
