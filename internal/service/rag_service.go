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
						contextStr += fmt.Sprintf("Source: %s\n%s\n\n", source, content)
					}
				}
			}
		}
	}

	return contextStr, nil
}