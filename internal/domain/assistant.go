package domain

import "time"

type KnowledgeSource struct {
	ID         string    `json:"id"`
	Title      string    `json:"title" binding:"required"`
	SourceType string    `json:"source_type"`
	Visibility string    `json:"visibility"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type Conversation struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id" binding:"required"`
	Role           string    `json:"role" binding:"required"` // USER, ASSISTANT
	Content        string    `json:"content" binding:"required"`
	CreatedAt      time.Time `json:"created_at"`
}

type Citation struct {
	ID             string  `json:"id"`
	MessageID      string  `json:"message_id"`
	DocumentID     string  `json:"document_id"`
	RelevanceScore float64 `json:"relevance_score"`
}
