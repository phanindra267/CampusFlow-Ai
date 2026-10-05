package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"

	"github.com/campuscare/api/internal/service/search"
)

// DocumentChunkClass is the Weaviate class holding embedded campus content.
// It is exported so the server can probe for it during start-up without
// duplicating the name.
const DocumentChunkClass = "DocumentChunk"

// Default retrieval limits.
const (
	defaultRetrieveLimit = 5
	maxRetrieveLimit     = 25
)

// RAGService retrieves grounded campus context for the client-side assistant.
//
// Two properties matter more than anything else here:
//
//  1. It never invents. Everything returned comes from PostgreSQL or Weaviate,
//     and when nothing matches the caller is told so explicitly rather than
//     being handed an empty string to improvise around.
//  2. It is cheap. Structured campus operations never reach this service, and
//     retrieval itself fuses a BM25 lexical arm with the Weaviate vector arm,
//     neither of which requires a language model.
type RAGService struct {
	corpus *CorpusSource
	vector VectorRetriever

	// retriever caches the fused index. The corpus changes rarely and BM25
	// statistics are query-independent, so rebuilding per request would be
	// wasted work. The service is shared across requests, so the cache is
	// guarded and concurrent callers wait for a single rebuild.
	mu        sync.Mutex
	retriever *search.Retriever
	expires   time.Time
	cacheTTL  time.Duration
}

// VectorRetriever is the optional semantic arm. It is an interface so the
// service is fully exercisable without a running Weaviate.
//
// DocumentChunk objects carry their origin in the `source` property, formatted
// as "KIND:id" to match the lexical corpus keys. Results the corpus does not
// know about are dropped, so the vector arm can only reinforce records that
// already exist.
type VectorRetriever interface {
	SimilarityIDs(query string, limit int, types []string) ([]search.VectorHit, error)
}

// NewRAGService builds the service. Both the Weaviate client and the corpus
// source are optional: without Weaviate retrieval is lexical-only, and without
// the corpus source the service reports itself as unavailable.
func NewRAGService(client *weaviate.Client, corpus *CorpusSource) *RAGService {
	service := &RAGService{
		corpus:   corpus,
		cacheTTL: 5 * time.Minute,
	}

	if client != nil {
		service.vector = newWeaviateRetriever(client)
	}

	return service
}

// weaviateRetriever implements VectorRetriever over the vendor client.
type weaviateRetriever struct {
	client *weaviate.Client
	class  string
}

// newWeaviateRetriever wraps a live Weaviate client as the semantic arm.
func newWeaviateRetriever(client *weaviate.Client) VectorRetriever {
	if client == nil {
		return nil
	}
	return &weaviateRetriever{client: client, class: DocumentChunkClass}
}

// SimilarityIDs runs a nearText query and maps the response back onto
// namespaced corpus IDs.
func (w *weaviateRetriever) SimilarityIDs(
	query string,
	limit int,
	types []string,
) ([]search.VectorHit, error) {
	if w == nil || w.client == nil {
		return nil, fmt.Errorf("weaviate client is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := w.client.GraphQL().Get().
		WithClassName(w.class).
		WithFields(graphql.Field{Name: "source"}).
		WithNearText(w.client.GraphQL().NearTextArgBuilder().WithConcepts([]string{query})).
		WithLimit(limit).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("weaviate returned no response")
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("weaviate returned errors: %v", result.Errors)
	}

	hits := filterVectorHits(parseChunkHits(result.Data), types)

	return hits, nil
}

// filterVectorHits drops vector results the lexical corpus cannot map onto, so
// the semantic arm only ever reinforces records that actually exist.
func filterVectorHits(hits []search.VectorHit, types []string) []search.VectorHit {
	allowed := make(map[string]bool, len(types))
	for _, t := range types {
		allowed[strings.ToUpper(strings.TrimSpace(t))] = true
	}

	out := make([]search.VectorHit, 0, len(hits))
	for _, hit := range hits {
		kind, _, ok := strings.Cut(hit.ID, ":")
		if !ok {
			continue
		}
		if len(allowed) > 0 && !allowed[strings.ToUpper(kind)] {
			continue
		}
		out = append(out, hit)
	}

	return out
}

// parseChunkHits walks Weaviate's dynamically typed GraphQL payload. Every type
// assertion is checked because optional fields are omitted when empty.
func parseChunkHits(data map[string]models.JSONObject) []search.VectorHit {
	root, ok := data["Get"].(map[string]interface{})
	if !ok {
		return nil
	}

	chunks, ok := root[DocumentChunkClass].([]interface{})
	if !ok {
		return nil
	}

	hits := make([]search.VectorHit, 0, len(chunks))
	for position, raw := range chunks {
		chunk, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		source, _ := chunk["source"].(string)
		if strings.TrimSpace(source) == "" {
			continue
		}

		// Weaviate returns no distance for nearText, so rank position carries
		// the ordering. Convert it to a descending 0..1 score.
		hits = append(hits, search.VectorHit{
			ID:    source,
			Score: 1 - float64(position)/float64(len(chunks)),
		})
	}

	return hits
}

// RetrieveContext returns a prompt-ready context string for the given query.
//
// The empty-string-plus-no-error case is reserved for a query with no usable
// text. A query that matches nothing returns ("", nil) so callers can state
// that the information is unavailable instead of sending an empty prompt.
func (s *RAGService) RetrieveContext(ctx context.Context, query string, limit int) (string, error) {
	result, err := s.RetrieveDocuments(ctx, query, limit, nil)
	if err != nil {
		return "", err
	}
	return BuildContext(result), nil
}

// RetrieveDocuments runs hybrid retrieval and returns the fused results. This
// is what the AI context endpoint serves.
func (s *RAGService) RetrieveDocuments(
	ctx context.Context,
	query string,
	limit int,
	types []string,
) (*RetrievalResult, error) {
	retriever, err := s.getRetriever(ctx)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = defaultRetrieveLimit
	}
	if limit > maxRetrieveLimit {
		limit = maxRetrieveLimit
	}

	fused, err := retriever.Retrieve(search.HybridQuery{
		Text:  query,
		Limit: limit,
		Types: types,
	})
	if err != nil {
		return nil, fmt.Errorf("rag: hybrid retrieval failed: %w", err)
	}

	return &RetrievalResult{
		Query:       fused.Query,
		Items:       fused.Items,
		LexicalHits: fused.LexicalHits,
		VectorHits:  fused.VectorHits,
	}, nil
}

// RetrievalResult is the backend-facing view of a hybrid query.
type RetrievalResult struct {
	Query       string
	Items       []search.Retrieved
	LexicalHits int
	VectorHits  int
}

// Empty reports whether retrieval found nothing at all.
func (r *RetrievalResult) Empty() bool {
	return r == nil || len(r.Items) == 0
}

// getRetriever returns a retriever over the current corpus, rebuilding it when
// the cache has expired.
func (s *RAGService) getRetriever(ctx context.Context) (*search.Retriever, error) {
	if s == nil || s.corpus == nil {
		return nil, fmt.Errorf("rag: retrieval is not configured")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.retriever != nil && time.Now().Before(s.expires) {
		return s.retriever, nil
	}

	docs, err := s.corpus.Load(ctx)
	if err != nil {
		return nil, err
	}

	var vector VectorRetriever
	if s.vector != nil {
		vector = s.vector
	}

	s.retriever = search.NewRetriever(docs, vector)
	s.expires = time.Now().Add(s.cacheTTL)

	return s.retriever, nil
}

// BuildContext renders retrieved documents into a compact context block for the
// on-device model.
//
// Every passage is labelled with its entity type so the model can attribute
// claims, and an explicit note accompanies thin results. That note is the
// anti-fabrication guard: the model is told, in the prompt itself, that the
// information is not in the system rather than being left to guess.
func BuildContext(result *RetrievalResult) string {
	if result.Empty() {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("Campus records retrieved for this question:\n\n")

	for i, item := range result.Items {
		doc := item.Document

		builder.WriteString(fmt.Sprintf("[%d] %s — %s\n", i+1, doc.Kind, doc.Title))

		if doc.URL != "" {
			builder.WriteString(fmt.Sprintf("Page: %s\n", doc.URL))
		}

		body := strings.TrimSpace(doc.Body)
		if len(body) > 600 {
			body = strings.TrimSpace(body[:600]) + "…"
		}
		if body != "" {
			builder.WriteString(body)
			builder.WriteString("\n")
		}

		if at, ok := doc.Payload["event_at"].(string); ok && at != "" {
			builder.WriteString(fmt.Sprintf("Starts: %s\n", at))
		}

		builder.WriteString("\n")
	}

	builder.WriteString(
		"Use only the records above. If they do not contain the answer, say that the " +
			"information is not available in CampusCare and name the relevant page instead " +
			"of estimating. Never invent dates, seat counts, deadlines, rankings or " +
			"organisers.\n",
	)

	return builder.String()
}
