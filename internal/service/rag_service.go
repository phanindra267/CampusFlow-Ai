package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

const documentChunkClass = "DocumentChunk"

type RAGService struct {
	client *weaviate.Client
}

func NewRAGService(client *weaviate.Client) *RAGService {
	return &RAGService{client: client}
}

// RetrieveContext performs a vector similarity search and flattens the results
// into a prompt-ready context string.
func (s *RAGService) RetrieveContext(ctx context.Context, query string, limit int) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("rag: weaviate client is not configured")
	}
	if strings.TrimSpace(query) == "" {
		return "", fmt.Errorf("rag: query must not be empty")
	}
	if limit <= 0 {
		limit = 5
	}

	result, err := s.client.GraphQL().Get().
		WithClassName(documentChunkClass).
		WithFields(graphql.Field{Name: "content"}, graphql.Field{Name: "source"}).
		WithNearText(s.client.GraphQL().NearTextArgBuilder().WithConcepts([]string{query})).
		WithLimit(limit).
		Do(ctx)
	if err != nil {
		return "", fmt.Errorf("weaviate retrieval failed: %w", err)
	}

	if result == nil || len(result.Errors) > 0 {
		return "", fmt.Errorf("weaviate retrieval returned errors: %v", result.Errors)
	}

	return buildContext(result.Data), nil
}

// buildContext walks Weaviate's dynamically typed GraphQL payload. Every type
// assertion is checked because optional fields are omitted from the response
// when they hold no value.
func buildContext(data map[string]models.JSONObject) string {
	var builder strings.Builder

	root, ok := data["Get"].(map[string]interface{})
	if !ok {
		return ""
	}

	chunks, ok := root[documentChunkClass].([]interface{})
	if !ok {
		return ""
	}

	for _, raw := range chunks {
		chunk, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		content, _ := chunk["content"].(string)
		if strings.TrimSpace(content) == "" {
			continue
		}

		source, _ := chunk["source"].(string)
		if source == "" {
			source = "unknown source"
		}

		builder.WriteString("Source: ")
		builder.WriteString(source)
		builder.WriteString("\n")
		builder.WriteString(content)
		builder.WriteString("\n\n")
	}

	return builder.String()
}
