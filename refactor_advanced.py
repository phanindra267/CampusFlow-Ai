import os

BASE_DIR = r"c:\Users\phani\Downloads\CampusCare AI"

def write_file(path, content):
    full_path = os.path.join(BASE_DIR, path)
    os.makedirs(os.path.dirname(full_path), exist_ok=True)
    with open(full_path, "w", encoding="utf-8") as f:
        f.write(content.strip())
    print(f"Written {path}")

# 1. OAuth & OTP Configurations
write_file("internal/config/oauth.go", """
package config

import (
	"os"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GetGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}
""")

write_file("internal/delivery/http/handler/oauth.go", """
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
""")

# 2. OTP Authentication
write_file("internal/delivery/http/handler/otp.go", """
package handler

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type OTPHandler struct {
	store sync.Map
}

func NewOTPHandler() *OTPHandler {
	return &OTPHandler{}
}

func (h *OTPHandler) RequestOTP(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		response.Error(c, http.StatusBadRequest, "phone is required", nil)
		return
	}

	val, _ := rand.Int(rand.Reader, big.NewInt(900000))
	otp := fmt.Sprintf("%06d", val.Int64()+100000)

	h.store.Store(phone, otp)

	// Simulate sending SMS via external provider
	fmt.Printf("SIMULATED SMS: Sending OTP %s to %s\\n", otp, phone)

	go func() {
		time.Sleep(5 * time.Minute)
		h.store.Delete(phone) // expire OTP
	}()

	response.Success(c, http.StatusOK, "otp sent", nil)
}

func (h *OTPHandler) VerifyOTP(c *gin.Context) {
	phone := c.Query("phone")
	otp := c.Query("otp")

	storedOTP, ok := h.store.Load(phone)
	if !ok || storedOTP.(string) != otp {
		response.Error(c, http.StatusUnauthorized, "invalid or expired otp", nil)
		return
	}

	h.store.Delete(phone)
	response.Success(c, http.StatusOK, "otp verified", nil)
}
""")

# 3. Audit Logging Middleware
write_file("internal/delivery/http/middleware/audit.go", """
package middleware

import (
	"bytes"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuditLog(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		// Only log state-mutating requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" || c.Request.Method == "PATCH" {
			userID, _ := c.Get("userID")
			uidStr, _ := userID.(string)

			query := `INSERT INTO audit_logs (user_id, method, path, status, latency_ms, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
			_, err := db.Exec(context.Background(), query, uidStr, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start).Milliseconds(), time.Now())
			if err != nil {
				log.Printf("Failed to write audit log: %v", err)
			}
		}
	}
}
""")

# 4. Async Event Bus / Notifications
write_file("internal/eventbus/bus.go", """
package eventbus

import (
	"log"
	"sync"
)

type Event struct {
	Type    string
	Payload interface{}
}

type EventBus struct {
	subscribers map[string][]chan Event
	rm          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan Event),
	}
}

func (b *EventBus) Subscribe(eventType string, ch chan Event) {
	b.rm.Lock()
	defer b.rm.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], ch)
}

func (b *EventBus) Publish(event Event) {
	b.rm.RLock()
	defer b.rm.RUnlock()
	if chans, found := b.subscribers[event.Type]; found {
		for _, ch := range chans {
			select {
			case ch <- event:
			default:
				log.Printf("EventBus: subscriber channel for %s is full, dropping event", event.Type)
			}
		}
	}
}
""")

# 5. Weaviate RAG Implementation
write_file("internal/service/rag_service.go", """
package service

import (
	"context"
	"fmt"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
)

type RAGService struct {
	client *weaviate.Client
}

func NewRAGService(client *weaviate.Client) *RAGService {
	return &RAGService{client: client}
}

func (s *RAGService) RetrieveContext(ctx context.Context, query string, limit int) (string, error) {
	// Real Weaviate vector similarity search
	result, err := s.client.GraphQL().Get().
		WithClassName("DocumentChunk").
		WithFields(graphql.Field{Name: "content"}, graphql.Field{Name: "source"}).
		WithNearText(s.client.GraphQL().NearTextArgBuilder().WithConcepts([]string{query})).
		WithLimit(limit).
		Do(ctx)

	if err != nil {
		return "", fmt.Errorf("weaviate retrieval failed: %w", err)
	}

	// Construct context string
	contextStr := ""
	if result.Data != nil {
		// Type asserting through Weaviate's deeply nested map[string]interface{}
		if get, ok := result.Data["Get"].(map[string]interface{}); ok {
			if chunks, ok := get["DocumentChunk"].([]interface{}); ok {
				for _, chunk := range chunks {
					if cMap, ok := chunk.(map[string]interface{}); ok {
						content := cMap["content"].(string)
						source := cMap["source"].(string)
						contextStr += fmt.Sprintf("Source: %s\\n%s\\n\\n", source, content)
					}
				}
			}
		}
	}

	return contextStr, nil
}
""")

# 6. Autonomous Agent Loop
write_file("internal/agent/orchestrator.go", """
package agent

import (
	"context"
	"fmt"
	"strings"
)

type Orchestrator struct {
	registry *Registry
}

func NewOrchestrator(registry *Registry) *Orchestrator {
	return &Orchestrator{registry: registry}
}

// ExecuteTask runs a basic ReAct loop simulation
func (o *Orchestrator) ExecuteTask(ctx context.Context, prompt string) (string, error) {
	// 1. Parse intent (Simulated LLM call)
	intent := parseIntent(prompt)

	// 2. Select Tool
	var result string
	var err error
	
	switch intent {
	case "check_attendance":
		result, err = o.registry.Execute(ctx, "GetAttendanceTool", map[string]interface{}{})
	case "find_events":
		result, err = o.registry.Execute(ctx, "GetEventsTool", map[string]interface{}{})
	default:
		result = "I couldn't determine the appropriate tool for that request."
	}

	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	// 3. Format Response (Simulated LLM synthesis)
	return fmt.Sprintf("Based on the data retrieved: %s", result), nil
}

func parseIntent(prompt string) string {
	p := strings.ToLower(prompt)
	if strings.Contains(p, "attendance") {
		return "check_attendance"
	}
	if strings.Contains(p, "event") || strings.Contains(p, "hackathon") {
		return "find_events"
	}
	return "unknown"
}
""")

print("Advanced Implementation Scripts Written.")
