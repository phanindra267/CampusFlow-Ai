package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/campuscare/api/internal/service"
	"github.com/campuscare/api/pkg/response"
)

// AIHandler serves retrieval for the on-device assistant.
//
// Scope is deliberately narrow: this endpoint returns grounded campus records
// and a context block, and nothing else. It never accepts a prompt to proxy and
// never calls a language model — generation happens in the browser via WebLLM
// on WebGPU. Structured campus operations stay on their existing REST routes.
type AIHandler struct {
	rag *service.RAGService
}

// NewAIHandler builds the handler.
func NewAIHandler(rag *service.RAGService) *AIHandler {
	return &AIHandler{rag: rag}
}

// aiContextResponse is the payload the assistant uses to ground a reply.
type aiContextResponse struct {
	Query       string            `json:"query"`
	Context     string            `json:"context"`
	Found       bool              `json:"found"`
	Records     []aiContextRecord `json:"records"`
	LexicalHits int               `json:"lexical_hits"`
	VectorHits  int               `json:"vector_hits"`
	Count       int               `json:"count"`
}

// aiContextRecord is one retrieved campus record.
type aiContextRecord struct {
	EntityType  string            `json:"entity_type"`
	EntityID    string            `json:"entity_id"`
	Title       string            `json:"title"`
	Summary     string            `json:"summary"`
	URL         string            `json:"url"`
	EventAt     string            `json:"event_at,omitempty"`
	RetrievedBy string            `json:"retrieved_by"`
	Meta        map[string]string `json:"meta,omitempty"`
}

// Context returns grounded records for a natural-language query.
//
// A query that matches nothing is a successful, empty result rather than an
// error: the assistant needs to be able to tell a member that the information
// is unavailable, which requires a 200 with `found: false`.
func (h *AIHandler) Context(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		response.Error(c, http.StatusBadRequest, "QUERY_REQUIRED", nil)
		return
	}

	if h.rag == nil {
		response.Error(c, http.StatusServiceUnavailable, "RETRIEVAL_UNAVAILABLE",
			errors.New("retrieval is not configured on this server"))
		return
	}

	types := splitTypes(c.Query("types"))
	limit := pageSize(c.Query("limit"))

	result, err := h.rag.RetrieveDocuments(c.Request.Context(), query, limit, types)
	if err != nil {
		response.Error(c, http.StatusServiceUnavailable, "RETRIEVAL_FAILED", err)
		return
	}

	payload := aiContextResponse{
		Query:       result.Query,
		Context:     service.BuildContext(result),
		Found:       !result.Empty(),
		Records:     make([]aiContextRecord, 0, len(result.Items)),
		LexicalHits: result.LexicalHits,
		VectorHits:  result.VectorHits,
		Count:       len(result.Items),
	}

	for _, item := range result.Items {
		record := aiContextRecord{
			EntityType:  item.Document.Kind,
			Title:       item.Document.Title,
			Summary:     item.Document.Body,
			URL:         item.Document.URL,
			RetrievedBy: string(item.Source),
		}

		if entityID, ok := item.Document.Payload["entity_id"].(string); ok {
			record.EntityID = entityID
		}
		if eventAt, ok := item.Document.Payload["event_at"].(string); ok {
			record.EventAt = eventAt
		}
		if category, ok := item.Document.Payload["category"].(string); ok && category != "" {
			record.Meta = map[string]string{"category": category}
		}

		payload.Records = append(payload.Records, record)
	}

	response.Success(c, http.StatusOK, "Context retrieved", payload)
}

// splitTypes parses the optional comma-separated entity filter.
func splitTypes(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	types := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			types = append(types, trimmed)
		}
	}

	if len(types) == 0 {
		return nil
	}
	return types
}
