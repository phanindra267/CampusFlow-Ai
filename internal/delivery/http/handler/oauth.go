package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/campuscare/api/internal/config"
	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/auth"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type OAuthHandler struct {
	jwtSecret string
	db        UserRepository
}

func NewOAuthHandler(jwtSecret string, db UserRepository) *OAuthHandler {
	return &OAuthHandler{jwtSecret: jwtSecret, db: db}
}

func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	url := config.GetGoogleOAuthConfig().AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	tok, err := config.GetGoogleOAuthConfig().Exchange(context.Background(), code)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to exchange token", err)
		return
	}

	client := config.GetGoogleOAuthConfig().Client(context.Background(), tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get user info", err)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to parse user info", err)
		return
	}

	user, err := h.db.GetByEmail(c.Request.Context(), userInfo.Email)
	if err != nil {
		// Auto-register
		user = &domain.User{
			ID:           uuid.New().String(),
			Email:        userInfo.Email,
			PasswordHash: "", // No password for OAuth users
			Role:         "STUDENT", // Default role
			Status:       "ACTIVE",
		}
		if createErr := h.db.Create(c.Request.Context(), user); createErr != nil {
			response.Error(c, http.StatusInternalServerError, "failed to create user", createErr)
			return
		}
	}

	jwtToken, _ := auth.GenerateToken(user.ID, user.Role, h.jwtSecret)
	response.Success(c, http.StatusOK, "oauth successful", gin.H{"token": jwtToken, "user": user})
}