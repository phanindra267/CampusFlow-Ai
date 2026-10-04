package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type AssistantHandler struct{}

func NewAssistantHandler() *AssistantHandler {
	return &AssistantHandler{}
}

func (h *AssistantHandler) CreateConversation(c *gin.Context) {
	userID, _ := c.Get("userID")

	response.Success(c, http.StatusCreated, "Conversation started", gin.H{
		"conversation_id": "mock-conv-id",
		"user_id":         userID,
	})
}

func (h *AssistantHandler) SendMessage(c *gin.Context) {
	var req domain.Message
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid message format", err)
		return
	}

	// Simulated RAG processing structure
	response.Success(c, http.StatusOK, "Message processed successfully", gin.H{
		"response": "This is a grounded assistant response.",
		"citations": []domain.Citation{
			{
				DocumentID:     "mock-doc-id",
				RelevanceScore: 0.98,
			},
		},
	})
}

func (h *AssistantHandler) UploadKnowledge(c *gin.Context) {
	var req domain.KnowledgeSource
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid knowledge metadata", err)
		return
	}

	response.Success(c, http.StatusCreated, "Document uploaded and queued for embedding", req)
}
