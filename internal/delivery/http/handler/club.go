package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type ClubHandler struct{}

func NewClubHandler() *ClubHandler {
	return &ClubHandler{}
}

func (h *ClubHandler) CreateClub(c *gin.Context) {
	var req domain.Club
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid club data", err)
		return
	}

	userID, _ := c.Get("userID")
	req.CreatedBy = userID.(string)

	response.Success(c, http.StatusCreated, "Club created successfully", req)
}

func (h *ClubHandler) ListClubs(c *gin.Context) {
	response.Success(c, http.StatusOK, "Clubs retrieved", []domain.Club{})
}

func (h *ClubHandler) JoinClub(c *gin.Context) {
	clubID := c.Param("id")
	userID, _ := c.Get("userID")

	response.Success(c, http.StatusOK, "Membership requested", gin.H{
		"club_id": clubID,
		"user_id": userID,
		"status":  "PENDING",
	})
}
