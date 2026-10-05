package search

import (
	"sort"
	"strings"
)

// BM25RFK is the reciprocal-rank-fusion constant. 60 is the value from the
// original Cormack et al. paper and works well for result lists of this size.
const BM25RFK = 60

// RetrievalSource records which arm of hybrid retrieval produced a result, so
// handlers and the UI can be explicit about how an answer was found.
type RetrievalSource string

const (
	SourceLexical RetrievalSource = "lexical"
	SourceVector  RetrievalSource = "vector"
	SourceHybrid  RetrievalSource = "hybrid"
)

// Document is one retrievable unit. ID must be stable across lexical and
// vector retrieval so the two result sets can be fused.
type Document struct {
	ID      string
	Title   string
	Body    string
	Source  string
	URL     string
	Kind    string
	Payload map[string]any
}

// VectorHit is one semantic result.
type VectorHit struct {
	ID    string
	Score float64
}

// VectorSearcher is the semantic half of retrieval. Weaviate satisfies this
// interface in production; tests use a stub.
type VectorSearcher interface {
	// SimilarityIDs returns document IDs ordered most similar first, with a
	// normalised score for each.
	SimilarityIDs(query string, limit int, types []string) ([]VectorHit, error)
}

// Retrieved is one ranked result. LexicalScore and VectorScore are the raw
// per-arm scores; FusedScore is the reciprocal-rank-fusion value.
type Retrieved struct {
	Document Document
	Source   RetrievalSource

	LexicalScore float64
	VectorScore  float64
	FusedScore   float64
	LexicalRank  int
	VectorRank   int
}

// Result is the full outcome of a hybrid query.
type Result struct {
	Query       string
	Items       []Retrieved
	LexicalHits int
	VectorHits  int
}

// HybridQuery is one request to the hybrid retriever.
type HybridQuery struct {
	Text string
	// Limit caps the number of fused results returned.
	Limit int
	// Types optionally restricts results to these entity kinds.
	Types []string
	// CandidatesPerArm caps each arm before fusion.
	CandidatesPerArm int
}

// Retriever combines lexical BM25 scoring with vector similarity over one
// shared corpus, keyed by Document.ID. That shared key is what makes fusion
// meaningful: the same campus record is matched by both arms and its two ranks
// combine.
//
// A vector-arm failure degrades to lexical-only rather than erroring, so search
// keeps working when Weaviate is unavailable.
type Retriever struct {
	docs   []Document
	byID   map[string]int
	bm25   *BM25
	vector VectorSearcher
}

// NewRetriever indexes the corpus lexically and attaches a vector arm. The
// vector arm may be nil, in which case retrieval is lexical-only.
func NewRetriever(docs []Document, vector VectorSearcher) *Retriever {
	byID := make(map[string]int, len(docs))
	for i, doc := range docs {
		byID[doc.ID] = i
	}

	return &Retriever{
		docs:   docs,
		byID:   byID,
		bm25:   NewBM25(docs),
		vector: vector,
	}
}

// Documents exposes the indexed corpus so callers can build prompt context.
func (r *Retriever) Documents() []Document {
	return r.docs
}

// Retrieve runs both arms and fuses them.
func (r *Retriever) Retrieve(query HybridQuery) (*Result, error) {
	text := strings.TrimSpace(query.Text)
	limit := query.Limit
	if limit <= 0 {
		limit = 10
	}
	perArm := query.CandidatesPerArm
	if perArm <= 0 {
		perArm = limit * 3
	}

	result := &Result{Query: text, Items: make([]Retrieved, 0, limit)}
	if text == "" || len(r.docs) == 0 {
		return result, nil
	}

	allowed := typeFilter(query.Types)

	// ---- lexical arm ----
	lexicalRanks := make(map[string]int)
	var lexicalIDs []string
	for rank, docIndex := range r.lexicalRank(text, perArm, allowed) {
		id := r.docs[docIndex].ID
		lexicalRanks[id] = rank
		lexicalIDs = append(lexicalIDs, id)
	}
	result.LexicalHits = len(lexicalIDs)

	// ---- vector arm ----
	vectorRanks := make(map[string]int)
	vectorScores := make(map[string]float64)
	var vectorIDs []string

	if r.vector != nil {
		hits, err := r.vector.SimilarityIDs(text, perArm, query.Types)
		if err != nil {
			// A vector-store outage must not take search down with it.
			hits = nil
		}
		for rank, hit := range hits {
			docIndex, known := r.byID[hit.ID]
			if !known {
				// The vector store may hold documents this corpus does not.
				continue
			}
			if allowed != nil && !allowed(r.docs[docIndex].Kind) {
				continue
			}
			vectorRanks[hit.ID] = rank
			vectorScores[hit.ID] = hit.Score
			vectorIDs = append(vectorIDs, hit.ID)
		}
	}
	result.VectorHits = len(vectorIDs)

	// ---- fusion ----
	lexicalScores := r.bm25.ScoreMap(text)

	for _, id := range ReciprocalRankFusion([][]string{lexicalIDs, vectorIDs}, BM25RFK) {
		if len(result.Items) >= limit {
			break
		}

		docIndex, known := r.byID[id]
		if !known {
			continue
		}

		item := Retrieved{Document: r.docs[docIndex]}

		lexicalRank, inLexical := lexicalRanks[id]
		vectorRank, inVector := vectorRanks[id]

		switch {
		case inLexical && inVector:
			item.Source = SourceHybrid
		case inVector:
			item.Source = SourceVector
		default:
			item.Source = SourceLexical
		}

		if inLexical {
			item.LexicalRank = lexicalRank
			item.LexicalScore = lexicalScores[id]
		}
		if inVector {
			item.VectorRank = vectorRank
			item.VectorScore = vectorScores[id]
		}

		item.FusedScore = fusedScore(lexicalRank, inLexical, vectorRank, inVector, BM25RFK)

		result.Items = append(result.Items, item)
	}

	return result, nil
}

func fusedScore(lexicalRank int, inLexical bool, vectorRank int, inVector bool, k int) float64 {
	var score float64
	if inLexical {
		score += 1 / float64(k+lexicalRank+1)
	}
	if inVector {
		score += 1 / float64(k+vectorRank+1)
	}
	return score
}

func (r *Retriever) lexicalRank(text string, limit int, allowed func(string) bool) []int {
	terms := Tokenize(text)
	if len(terms) == 0 {
		return nil
	}

	candidates := make([]scoredDoc, 0, len(r.docs))
	for i, doc := range r.docs {
		if allowed != nil && !allowed(doc.Kind) {
			continue
		}
		if score := r.bm25.Score(i, terms); score > 0 {
			candidates = append(candidates, scoredDoc{index: i, score: score})
		}
	}

	// Highest score first; document ID breaks ties so ordering is stable.
	sort.SliceStable(candidates, func(a, b int) bool {
		if candidates[a].score != candidates[b].score {
			return candidates[a].score > candidates[b].score
		}
		return r.docs[candidates[a].index].ID < r.docs[candidates[b].index].ID
	})

	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	order := make([]int, len(candidates))
	for i, candidate := range candidates {
		order[i] = candidate.index
	}
	return order
}

type scoredDoc struct {
	index int
	score float64
}

func typeFilter(types []string) func(string) bool {
	if len(types) == 0 {
		return nil
	}

	allowed := make(map[string]bool, len(types))
	for _, t := range types {
		allowed[strings.ToUpper(strings.TrimSpace(t))] = true
	}

	return func(kind string) bool {
		return allowed[strings.ToUpper(kind)]
	}
}
