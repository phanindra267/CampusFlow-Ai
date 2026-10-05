// Package search implements the lexical half of hybrid retrieval.
//
// BM25 rather than plain keyword matching: PostgreSQL full-text search has no
// notion of term rarity, so a rare word like "quantum" scores the same as a
// common one like "workshop". BM25 corrects for that with inverse document
// frequency and a length-normalisation term, which is why exact keyword matches
// reliably outrank merely-longer passages.
//
// It is implemented in Go on the backend, against documents already loaded from
// PostgreSQL. That keeps ranking deterministic and testable, and avoids adding a
// separate search service to the infrastructure.
package search

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// DefaultBM25 constants. k1 controls term-frequency saturation and b controls
// how strongly document length is penalised; these are the values used by the
// original Okapi BM25 paper and by every mainstream implementation since.
const (
	DefaultK1 = 1.2
	DefaultB  = 0.75
)

// Tokenize lowercases, splits on non-alphanumerics and drops single characters
// and stop words, yielding the terms BM25 scores over.
func Tokenize(text string) []string {
	lower := strings.ToLower(text)

	fields := strings.FieldsFunc(lower, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) < 2 || stopWords[field] {
			continue
		}
		tokens = append(tokens, field)
	}

	return tokens
}

// stopWords are dropped before scoring. Every occurrence would otherwise add
// the same idf to every document, which does not change the ranking.
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true, "of": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "is": true,
	"are": true, "was": true, "were": true, "be": true, "been": true, "am": true,
	"i": true, "me": true, "my": true, "we": true, "our": true, "you": true,
	"it": true, "this": true, "that": true, "these": true, "those": true,
	"with": true, "from": true, "by": true, "as": true, "do": true, "does": true,
	"did": true, "has": true, "have": true, "had": true, "will": true, "would": true,
	"can": true, "could": true, "should": true, "show": true, "find": true,
	"list": true, "get": true, "please": true, "there": true, "any": true,
	"all": true, "some": true, "what": true, "which": true, "who": true,
}

// BM25 scores a fixed document set. Build one per corpus and reuse it across
// queries; all corpus statistics are query-independent.
type BM25 struct {
	docs       []Document
	termFreqs  []map[string]int
	docLengths []int
	avgDocLen  float64
	docFreq    map[string]int
	k1         float64
	b          float64
	weights    map[string]float64
}

// NewBM25 builds the scoring index for a corpus. The title is weighted more
// than the body, which reflects how titles in this domain are written.
func NewBM25(docs []Document) *BM25 {
	index := &BM25{
		docs:       docs,
		termFreqs:  make([]map[string]int, len(docs)),
		docLengths: make([]int, len(docs)),
		docFreq:    make(map[string]int),
		k1:         DefaultK1,
		b:          DefaultB,
		weights:    make(map[string]float64),
	}

	var totalLength int

	for i, doc := range docs {
		// The title is indexed twice so a title hit outweighs the same word
		// appearing once in the body.
		bodyTerms := Tokenize(doc.Body)
		tokens := make([]string, 0, 2*len(bodyTerms))
		tokens = append(tokens, Tokenize(doc.Title)...)
		tokens = append(tokens, bodyTerms...)
		tokens = append(tokens, Tokenize(doc.Title)...)

		freqs := make(map[string]int, len(tokens))
		for _, token := range tokens {
			freqs[token]++
		}

		index.termFreqs[i] = freqs
		index.docLengths[i] = len(tokens)
		totalLength += len(tokens)

		for term := range freqs {
			index.docFreq[term]++
		}
	}

	if len(docs) > 0 {
		index.avgDocLen = float64(totalLength) / float64(len(docs))
	}
	if index.avgDocLen == 0 {
		index.avgDocLen = 1
	}

	for term, df := range index.docFreq {
		// Robertson/Sparck-Jones IDF with the +1 smoothing that keeps the
		// score positive for terms appearing in more than half the corpus.
		idf := math.Log(1 + (float64(len(docs)-df)+0.5)/(float64(df)+0.5))
		if idf > 0 {
			index.weights[term] = idf
		}
	}

	return index
}

// lengthNormalizer is the BM25 length penalty for one document. Short
// documents are favoured for a given term frequency.
func (index *BM25) lengthNormalizer(docIndex int) float64 {
	ratio := float64(index.docLengths[docIndex]) / index.avgDocLen
	normalizer := index.k1 * (1 - index.b + index.b*ratio)
	if normalizer <= 0 {
		return index.k1
	}
	return normalizer
}

// Score returns the BM25 score of one document for the given query terms.
func (index *BM25) Score(docIndex int, queryTerms []string) float64 {
	if docIndex < 0 || docIndex >= len(index.docs) {
		return 0
	}

	var score float64
	for _, term := range queryTerms {
		weight, ok := index.weights[term]
		if !ok {
			continue
		}

		freq := float64(index.termFreqs[docIndex][term])
		if freq == 0 {
			continue
		}

		normalizer := index.lengthNormalizer(docIndex)
		score += weight * (freq * (index.k1 + 1)) / (freq + normalizer)
	}

	return score
}

// Rank returns document indices ordered by BM25 score, highest first. Ties
// break on document ID so results are reproducible run to run.
func (index *BM25) Rank(query string) []int {
	terms := Tokenize(query)
	if len(terms) == 0 || len(index.docs) == 0 {
		return nil
	}

	type scored struct {
		index int
		score float64
	}

	ranked := make([]scored, 0, len(index.docs))
	for i := range index.docs {
		if score := index.Score(i, terms); score > 0 {
			ranked = append(ranked, scored{index: i, score: score})
		}
	}

	sort.SliceStable(ranked, func(a, b int) bool {
		if ranked[a].score != ranked[b].score {
			return ranked[a].score > ranked[b].score
		}
		return index.docs[ranked[a].index].ID < index.docs[ranked[b].index].ID
	})

	order := make([]int, len(ranked))
	for i, entry := range ranked {
		order[i] = entry.index
	}
	return order
}

// ScoreMap returns BM25 scores keyed by document ID, for fusion with a
// separate vector result set.
func (index *BM25) ScoreMap(query string) map[string]float64 {
	scores := make(map[string]float64)

	terms := Tokenize(query)
	if len(terms) == 0 {
		return scores
	}

	for i, doc := range index.docs {
		if score := index.Score(i, terms); score > 0 {
			scores[doc.ID] = score
		}
	}

	return scores
}

// ReciprocalRankFusion merges ranked result lists into one ordering.
//
// RRF combines rankings by position rather than by score, which avoids having
// to reconcile BM25's unbounded scores with cosine similarity's 0..1 range. The
// constant damps the influence of the top few ranks so a document cannot win
// on one list's first place alone.
func ReciprocalRankFusion(lists [][]string, k int) []string {
	if k <= 0 {
		k = 60
	}

	positions := make(map[string]float64)
	for _, list := range lists {
		for rank, id := range list {
			positions[id] += 1 / float64(k+rank+1)
		}
	}

	fused := make([]string, 0, len(positions))
	for id := range positions {
		fused = append(fused, id)
	}

	sort.SliceStable(fused, func(a, b int) bool {
		// Descending fused score; ID ascending breaks ties deterministically.
		if positions[fused[a]] != positions[fused[b]] {
			return positions[fused[a]] > positions[fused[b]]
		}
		return fused[a] < fused[b]
	})

	return fused
}
