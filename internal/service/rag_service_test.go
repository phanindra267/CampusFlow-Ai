package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/campuscare/api/internal/service/search"
)

// newTestService builds an RAGService over a fixed in-memory corpus, so the
// retrieval logic can be exercised without PostgreSQL.
func newTestService(docs []search.Document, vector VectorRetriever) *RAGService {
	return &RAGService{
		corpus: &CorpusSource{
			cached:  docs,
			expires: time.Now().Add(time.Hour),
			ttl:     time.Hour,
		},
		vector: vector,
	}
}

func sampleCorpus() []search.Document {
	return []search.Document{
		{
			ID: "EVENT:1", Title: "Intro to Cloud Computing",
			Body: "A hands-on workshop on cloud computing and containers.",
			Kind: "EVENT", Source: "EVENT", URL: "/events/1",
			Payload: map[string]any{"entity_id": "1", "event_at": "2026-01-05T10:00:00Z"},
		},
		{
			ID: "CLUB:1", Title: "Coding Club",
			Body: "Weekly peer programming.",
			Kind: "CLUB", Source: "CLUB", URL: "/clubs/1",
			Payload: map[string]any{"entity_id": "1"},
		},
		{
			ID: "OPPORTUNITY:1", Title: "Cloud Engineering Internship",
			Body: "Six month internship on cloud infrastructure.",
			Kind: "OPPORTUNITY", Source: "OPPORTUNITY", URL: "/opportunities/1",
			Payload: map[string]any{"entity_id": "1"},
		},
	}
}

// stubVector is a deterministic semantic arm.
type stubVector struct {
	hits []search.VectorHit
	err  error
}

func (s *stubVector) SimilarityIDs(string, int, []string) ([]search.VectorHit, error) {
	return s.hits, s.err
}

func TestRetrieveContextReturnsGroundedRecords(t *testing.T) {
	service := newTestService(sampleCorpus(), nil)

	context, err := service.RetrieveContext(context.Background(), "cloud computing", 5)
	if err != nil {
		t.Fatalf("RetrieveContext returned an error: %v", err)
	}

	if strings.TrimSpace(context) == "" {
		t.Fatal("expected a non-empty context block")
	}
	if !strings.Contains(context, "Intro to Cloud Computing") {
		t.Fatalf("expected the matching event title in the context, got:\n%s", context)
	}
	if !strings.Contains(context, "/events/1") {
		t.Fatalf("expected the record URL for attribution, got:\n%s", context)
	}
}

func TestRetrieveContextIsEmptyWhenNothingMatches(t *testing.T) {
	service := newTestService(sampleCorpus(), nil)

	context, err := service.RetrieveContext(context.Background(), "medieval falconry", 5)
	if err != nil {
		t.Fatalf("a query with no matches must not error, got: %v", err)
	}

	if context != "" {
		t.Fatalf("expected an empty context so the assistant says the data is unavailable, got:\n%s", context)
	}
}

func TestBuildContextInstructsAgainstFabrication(t *testing.T) {
	service := newTestService(sampleCorpus(), nil)

	context, err := service.RetrieveContext(context.Background(), "cloud computing", 5)
	if err != nil {
		t.Fatalf("RetrieveContext returned an error: %v", err)
	}

	// The anti-fabrication guard has to be in the prompt itself, because the
	// model cannot be trusted to infer it.
	for _, expected := range []string{"not available", "Never invent"} {
		if !strings.Contains(context, expected) {
			t.Fatalf("expected context to contain %q, got:\n%s", expected, context)
		}
	}
}

func TestRetrieveDocumentsFusesVectorAndLexical(t *testing.T) {
	// The vector arm matches an opportunity the words in the query do not.
	vector := &stubVector{hits: []search.VectorHit{{ID: "OPPORTUNITY:1", Score: 0.9}}}
	service := newTestService(sampleCorpus(), vector)

	result, err := service.RetrieveDocuments(context.Background(), "cloud computing", 5, nil)
	if err != nil {
		t.Fatalf("RetrieveDocuments returned an error: %v", err)
	}

	if result.Empty() {
		t.Fatal("expected results")
	}
	if result.LexicalHits == 0 {
		t.Error("expected lexical hits")
	}
	if result.VectorHits != 1 {
		t.Errorf("expected 1 vector hit, got %d", result.VectorHits)
	}
}

func TestRetrieveDocumentsSurvivesVectorOutage(t *testing.T) {
	vector := &stubVector{err: context.DeadlineExceeded}
	service := newTestService(sampleCorpus(), vector)

	result, err := service.RetrieveDocuments(context.Background(), "cloud", 5, nil)
	if err != nil {
		t.Fatalf("a vector outage must not fail retrieval, got: %v", err)
	}

	if result.Empty() {
		t.Fatal("expected lexical results to survive a vector outage")
	}
}

func TestRetrieveDocumentsClampsLimit(t *testing.T) {
	service := newTestService(sampleCorpus(), nil)

	result, err := service.RetrieveDocuments(context.Background(), "cloud", 1000, nil)
	if err != nil {
		t.Fatalf("RetrieveDocuments returned an error: %v", err)
	}

	if len(result.Items) > maxRetrieveLimit {
		t.Fatalf("expected at most %d results, got %d", maxRetrieveLimit, len(result.Items))
	}
}

func TestRetrieveDocumentsFiltersByType(t *testing.T) {
	service := newTestService(sampleCorpus(), nil)

	result, err := service.RetrieveDocuments(context.Background(), "cloud computing", 10, []string{"CLUB"})
	if err != nil {
		t.Fatalf("RetrieveDocuments returned an error: %v", err)
	}

	for _, item := range result.Items {
		if item.Document.Kind != "CLUB" {
			t.Fatalf("expected only CLUB results, got %s", item.Document.Kind)
		}
	}
}

func TestRAGServiceWithoutCorpusIsUnavailable(t *testing.T) {
	// A nil corpus must report unavailability rather than panic.
	service := &RAGService{}

	if _, err := service.RetrieveDocuments(context.Background(), "cloud", 5, nil); err == nil {
		t.Fatal("expected an error when retrieval is not configured")
	}
}

func TestFilterVectorHitsDropsUnnamespacedIDs(t *testing.T) {
	// IDs without a "KIND:id" shape cannot be matched against the lexical
	// corpus, so they must be discarded.
	hits := []search.VectorHit{
		{ID: "EVENT:1", Score: 1},
		{ID: "bare-id", Score: 0.5},
	}

	filtered := filterVectorHits(hits, nil)

	if len(filtered) != 1 || filtered[0].ID != "EVENT:1" {
		t.Fatalf("expected only the namespaced ID to survive, got %v", filtered)
	}
}

func TestFilterVectorHitsAppliesTypeFilter(t *testing.T) {
	hits := []search.VectorHit{
		{ID: "EVENT:1", Score: 1},
		{ID: "CLUB:1", Score: 0.9},
	}

	filtered := filterVectorHits(hits, []string{"club"})

	if len(filtered) != 1 || filtered[0].ID != "CLUB:1" {
		t.Fatalf("expected only the CLUB hit, got %v", filtered)
	}
}
