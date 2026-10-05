package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/campuscare/api/internal/service/search"
)

// CorpusSource loads the retrievable campus corpus from PostgreSQL.
//
// The corpus is small and changes only when members create content, so it is
// loaded on demand and cached for a short TTL rather than being streamed per
// query. That keeps the lexical arm free of per-request database work.
//
// The service is shared by every request, so the cache is guarded: a
// concurrent burst of requests waits for one load instead of each issuing its
// own six queries.
type CorpusSource struct {
	db  *pgxpool.Pool
	ttl time.Duration

	mu      sync.Mutex
	cached  []search.Document
	expires time.Time
}

// NewCorpusSource builds a corpus loader for the given pool.
func NewCorpusSource(db *pgxpool.Pool) *CorpusSource {
	return &CorpusSource{db: db, ttl: 5 * time.Minute}
}

// Document row queries. Each one selects the member-visible records and
// projects them into a search.Document. The trigram/tsvector indexes let
// PostgreSQL narrow candidates, but the union is deliberately bounded so the
// BM25 corpus stays predictable in size.
const (
	corpusEventSQL = `
		SELECT id::text, title, coalesce(description, '') || ' ' || coalesce(venue, '') || ' ' || coalesce(category, ''),
			'/events/' || id::text, 'EVENT', category, start_time
		FROM events
		WHERE status = 'PUBLISHED'
		ORDER BY start_time DESC
		LIMIT 500`

	corpusClubSQL = `
		SELECT id::text, name, coalesce(description, '') || ' ' || coalesce(category, ''),
			'/clubs/' || id::text, 'CLUB', coalesce(category, ''), NULL::timestamptz
		FROM clubs
		WHERE status = 'ACTIVE'
		LIMIT 500`

	corpusOpportunitySQL = `
		SELECT id::text, title, coalesce(description, '') || ' ' || coalesce(type, ''),
			'/opportunities/' || id::text, 'OPPORTUNITY', type, start_date
		FROM opportunities
		WHERE status = 'OPEN'
		LIMIT 500`

	corpusAnnouncementSQL = `
		SELECT id::text, title, body,
			'/announcements/' || id::text, 'ANNOUNCEMENT', category, published_at
		FROM announcements
		WHERE status = 'PUBLISHED'
		ORDER BY published_at DESC
		LIMIT 300`

	corpusResourceSQL = `
		SELECT id::text, name,
			coalesce(description, '') || ' ' || coalesce(location, '') || ' ' || coalesce(resource_type, ''),
			'/services/' || id::text, 'RESOURCE', coalesce(resource_type, ''), NULL::timestamptz
		FROM campus_resources
		WHERE status = 'AVAILABLE'
		LIMIT 300`

	corpusDiscussionSQL = `
		SELECT id::text, title, body || ' ' || coalesce(category, ''),
			'/community?discussion=' || id::text, 'DISCUSSION', coalesce(category, ''),
			last_activity_at
		FROM discussions
		WHERE status = 'OPEN'
		ORDER BY last_activity_at DESC
		LIMIT 400`
)

// corpusQueries is the fixed set of entity queries. Keeping it a package-level
// slice means Load is a loop rather than six near-identical blocks.
var corpusQueries = []struct {
	kind string
	sql  string
}{
	{"EVENT", corpusEventSQL},
	{"CLUB", corpusClubSQL},
	{"OPPORTUNITY", corpusOpportunitySQL},
	{"ANNOUNCEMENT", corpusAnnouncementSQL},
	{"RESOURCE", corpusResourceSQL},
	{"DISCUSSION", corpusDiscussionSQL},
}

// Load returns the campus corpus, using the cached copy when it is still warm.
// A single failing entity type is skipped rather than failing the whole load,
// because search degrading to fewer entity types is far better than search
// being unavailable.
func (s *CorpusSource) Load(ctx context.Context) ([]search.Document, error) {
	if s == nil {
		return nil, fmt.Errorf("rag: corpus source is not configured")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// A warm cache needs no database, so it is served before the pool check.
	if len(s.cached) > 0 && time.Now().Before(s.expires) {
		return s.cached, nil
	}

	if s.db == nil {
		return nil, fmt.Errorf("rag: database pool is not configured")
	}

	docs := make([]search.Document, 0, 1500)

	for _, query := range corpusQueries {
		loaded, err := s.loadKind(ctx, query.kind, query.sql)
		if err != nil {
			continue
		}
		docs = append(docs, loaded...)
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("rag: campus corpus is empty")
	}

	s.cached = docs
	s.expires = time.Now().Add(s.ttl)

	return docs, nil
}

// loadKind runs one corpus query. Document IDs are namespaced by entity kind so
// lexical and vector results can be fused on a unique key even where the same
// UUID appears in two tables.
func (s *CorpusSource) loadKind(ctx context.Context, kind string, sql string) ([]search.Document, error) {
	rows, err := s.db.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("corpus %s: %w", kind, err)
	}
	defer rows.Close()

	var docs []search.Document

	for rows.Next() {
		// Every corpus query projects the same seven columns, in this order:
		// id, title, body, url, kind, category, and an optional timestamp.
		var id, title, body, url, rowKind, category string
		var at *time.Time

		if err := rows.Scan(&id, &title, &body, &url, &rowKind, &category, &at); err != nil {
			return nil, fmt.Errorf("scan corpus %s: %w", kind, err)
		}
		if rowKind == "" {
			rowKind = kind
		}

		doc := search.Document{
			ID:     rowKind + ":" + id,
			Title:  title,
			Body:   body,
			Source: rowKind,
			URL:    url,
			Kind:   rowKind,
			Payload: map[string]any{
				"entity_id": id,
				"category":  category,
			},
		}
		if at != nil {
			doc.Payload["event_at"] = at.UTC().Format(time.RFC3339)
		}

		docs = append(docs, doc)
	}

	return docs, rows.Err()
}
