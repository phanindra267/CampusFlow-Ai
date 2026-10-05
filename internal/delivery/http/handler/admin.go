package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/internal/repository/postgres"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type AdminUserRepository interface {
	ListUsers(context.Context, int, int) ([]domain.User, error)
}

type AdminClubRepository interface {
	ListPendingVerification(context.Context, int, int) ([]domain.Club, error)
	SetVerificationStatus(context.Context, string, string) (*domain.Club, error)
}

type AdminHandler struct {
	users AdminUserRepository
	clubs AdminClubRepository
}

func NewAdminHandler(users AdminUserRepository, clubs AdminClubRepository) *AdminHandler {
	return &AdminHandler{users: users, clubs: clubs}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.users.ListUsers(c.Request.Context(), pageSize(c.Query("limit")),
		parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "USER_NOT_FOUND", "USER_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Users retrieved", gin.H{
		"users": users,
		"count": len(users),
	})
}

func (h *AdminHandler) ListPendingClubApprovals(c *gin.Context) {
	clubs, err := h.clubs.ListPendingVerification(c.Request.Context(), pageSize(c.Query("limit")),
		parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_APPROVAL_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Pending club approvals retrieved", gin.H{
		"clubs": clubs,
		"count": len(clubs),
	})
}

func (h *AdminHandler) UpdateClubVerification(c *gin.Context) {
	var request struct {
		Status string `json:"status" binding:"required,oneof=APPROVED REJECTED"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	club, err := h.clubs.SetVerificationStatus(c.Request.Context(), c.Param("id"), request.Status)
	if errors.Is(err, postgres.ErrNotFound) {
		response.Error(c, http.StatusNotFound, "CLUB_NOT_FOUND", err)
		return
	}
	if err != nil {
		writeRepoError(c, err, "CLUB_NOT_FOUND", "CLUB_APPROVAL_UPDATE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Club verification updated", club)
}
