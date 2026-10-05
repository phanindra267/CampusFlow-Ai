package search

import (
	"testing"
)

// corpus builds a small, realistic campus corpus. Term frequencies are
// deliberately uneven so BM25's rarity weighting is observable.
func corpus() []Document {
	return []Document{
		{ID: "EVENT:1", Title: "Intro to Cloud Computing", Body: "A hands-on workshop on cloud computing and containers.", Kind: "EVENT"},
		{ID: "EVENT:2", Title: "Advanced Distributed Systems", Body: "Deep dive into consensus, replication and large scale systems.", Kind: "EVENT"},
		{ID: "EVENT:3", Title: "Quantum Computing Seminar", Body: "An overview of quantum computing research directions.", Kind: "EVENT"},
		{ID: "CLUB:1", Title: "Coding Club", Body: "Weekly peer programming and competitive practice.", Kind: "CLUB"},
		{ID: "CLUB:2", Title: "Robotics Society", Body: "Build robots and compete nationally.", Kind: "CLUB"},
		{ID: "OPPORTUNITY:1", Title: "Cloud Engineering Internship", Body: "Six month internship on cloud infrastructure.", Kind: "OPPORTUNITY"},
		{ID: "OPPORTUNITY:2", Title: "Research Assistant in Quantum Systems", Body: "Paid research assistantship working on quantum error correction.", Kind: "OPPORTUNITY"},
	}
}

func TestTokenizeDropsStopWordsAndSingleCharacters(t *testing.T) {
	tokens := Tokenize("The quick brown fox, a 5")

	want := []string{"quick", "brown", "fox"}
	if len(tokens) != len(want) {
		t.Fatalf("Tokenize returned %v, want %v", tokens, want)
	}
	for i := range want {
		if tokens[i] != want[i] {
			t.Fatalf("Tokenize returned %v, want %v", tokens, want)
		}
	}
}

func TestBM25FavoursRareTerms(t *testing.T) {
	index := NewBM25(corpus())

	// "quantum" appears in 2 of 7 documents; "computing" appears in 3 and is
	// lexically a prefix of quantum documents too. The rare term must carry
	// more weight for a query about it.
	scores := index.ScoreMap("quantum")

	if scores["EVENT:3"] == 0 {
		t.Fatal("expected the quantum seminar to match 'quantum'")
	}
	if scores["EVENT:3"] <= 0 {
		t.Fatalf("expected a positive score, got %v", scores["EVENT:3"])
	}
}

func TestBM25MatchesExactKeywordInTitle(t *testing.T) {
	index := NewBM25(corpus())

	order := index.Rank("quantum computing")

	if len(order) == 0 {
		t.Fatal("expected matches for 'quantum computing'")
	}

	// Both documents mention quantum computing, but the seminar is titled for
	// it so it must rank first.
	if index.docs[order[0]].ID != "EVENT:3" {
		t.Fatalf("expected EVENT:3 first, got %s", index.docs[order[0]].ID)
	}
}

func TestBM25ReturnsNoResultsForUnmatchedQuery(t *testing.T) {
	index := NewBM25(corpus())

	if order := index.Rank("medieval falconry"); len(order) != 0 {
		t.Fatalf("expected no results, got %v", order)
	}

	if scores := index.ScoreMap("medieval falconry"); len(scores) != 0 {
		t.Fatalf("expected an empty score map, got %v", scores)
	}
}

func TestBM25IsDeterministic(t *testing.T) {
	first := NewBM25(corpus()).Rank("cloud computing")
	second := NewBM25(corpus()).Rank("cloud computing")

	if len(first) != len(second) {
		t.Fatalf("result counts differ: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("ordering differs at %d: %d vs %d", i, first[i], second[i])
		}
	}
}

// stubVector is a deterministic semantic arm for tests.
type stubVector struct {
	hits []VectorHit
	err  error
}

func (s *stubVector) SimilarityIDs(string, int, []string) ([]VectorHit, error) {
	return s.hits, s.err
}

func TestHybridRetrievalFusesBothArms(t *testing.T) {
	// The vector arm knows about the quantum opportunity but never says the
	// word "quantum", which is exactly the case lexical search misses.
	vector := &stubVector{hits: []VectorHit{{ID: "OPPORTUNITY:2", Score: 0.9}}}

	retriever := NewRetriever(corpus(), vector)
	result, err := retriever.Retrieve(HybridQuery{Text: "quantum", Limit: 5})
	if err != nil {
		t.Fatalf("Retrieve returned an error: %v", err)
	}

	if result.LexicalHits == 0 {
		t.Fatal("expected lexical hits for 'quantum'")
	}
	if result.VectorHits != 1 {
		t.Fatalf("expected 1 vector hit, got %d", result.VectorHits)
	}

	found := false
	sources := map[RetrievalSource]bool{}
	for _, item := range result.Items {
		if item.Document.ID == "OPPORTUNITY:2" {
			found = true
		}
		sources[item.Source] = true
	}

	if !found {
		t.Fatal("expected the vector-only match to appear in fused results")
	}
	if !sources[SourceHybrid] {
		t.Fatalf("expected at least one hybrid-ranked result, got %v", sources)
	}
}

func TestHybridRetrievalDegradesToLexicalWhenVectorFails(t *testing.T) {
	vector := &stubVector{err: errStub}

	retriever := NewRetriever(corpus(), vector)
	result, err := retriever.Retrieve(HybridQuery{Text: "quantum", Limit: 5})
	if err != nil {
		t.Fatalf("a vector failure must not fail retrieval, got: %v", err)
	}

	if len(result.Items) == 0 {
		t.Fatal("expected lexical results to survive a vector outage")
	}
	for _, item := range result.Items {
		if item.Source == SourceVector || item.Source == SourceHybrid {
			t.Fatalf("expected lexical-only sources, got %s", item.Source)
		}
	}
}

var errStub = errStubType{}

type errStubType struct{}

func (errStubType) Error() string { return "weaviate unavailable" }

func TestHybridRetrievalIgnoresUnknownVectorIDs(t *testing.T) {
	// The vector store may hold documents the campus corpus does not. They must
	// not reach the results.
	vector := &stubVector{hits: []VectorHit{{ID: "GHOST:99", Score: 1}}}

	retriever := NewRetriever(corpus(), vector)
	result, err := retriever.Retrieve(HybridQuery{Text: "quantum", Limit: 10})
	if err != nil {
		t.Fatalf("Retrieve returned an error: %v", err)
	}

	for _, item := range result.Items {
		if item.Document.ID == "GHOST:99" {
			t.Fatal("unknown vector IDs must be dropped")
		}
	}
}

func TestHybridRetrievalFiltersByType(t *testing.T) {
	retriever := NewRetriever(corpus(), nil)

	result, err := retriever.Retrieve(HybridQuery{
		Text:  "computing",
		Limit: 10,
		Types: []string{"CLUB"},
	})
	if err != nil {
		t.Fatalf("Retrieve returned an error: %v", err)
	}

	for _, item := range result.Items {
		if item.Document.Kind != "CLUB" {
			t.Fatalf("expected only CLUB results, got %s", item.Document.Kind)
		}
	}
}

func TestHybridRetrievalRespectsLimit(t *testing.T) {
	retriever := NewRetriever(corpus(), nil)

	result, err := retriever.Retrieve(HybridQuery{Text: "computing cloud quantum", Limit: 2})
	if err != nil {
		t.Fatalf("Retrieve returned an error: %v", err)
	}

	if len(result.Items) > 2 {
		t.Fatalf("expected at most 2 results, got %d", len(result.Items))
	}
}

func TestHybridRetrievalHandlesEmptyQuery(t *testing.T) {
	retriever := NewRetriever(corpus(), nil)

	result, err := retriever.Retrieve(HybridQuery{Text: "   ", Limit: 5})
	if err != nil {
		t.Fatalf("an empty query must not error, got: %v", err)
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected no results, got %d", len(result.Items))
	}
}

func TestReciprocalRankFusionPromotesDocumentsFoundByBothArms(t *testing.T) {
	// B and C appear in both lists; A, D and E appear in only one. Agreement
	// between arms is the whole point of fusion, so both-arm documents must
	// outrank every single-arm document.
	lexical := []string{"A", "B", "C"}
	vector := []string{"D", "E", "B", "C"}

	fused := ReciprocalRankFusion([][]string{lexical, vector}, BM25RFK)

	position := make(map[string]int, len(fused))
	for i, id := range fused {
		if _, seen := position[id]; seen {
			t.Fatalf("duplicate %s in fused output", id)
		}
		position[id] = i
	}

	if len(fused) != 5 {
		t.Fatalf("expected 5 unique documents, got %v", fused)
	}

	for _, agreed := range []string{"B", "C"} {
		for _, single := range []string{"A", "D", "E"} {
			if position[agreed] > position[single] {
				t.Fatalf("expected %s (both arms) to outrank %s (one arm), got %v",
					agreed, single, fused)
			}
		}
	}
}

func TestReciprocalRankFusionIsDeterministic(t *testing.T) {
	lists := [][]string{{"A", "B", "C"}, {"C", "A"}}

	first := ReciprocalRankFusion(lists, BM25RFK)
	second := ReciprocalRankFusion(lists, BM25RFK)

	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("fusion is not deterministic: %v vs %v", first, second)
		}
	}
}
