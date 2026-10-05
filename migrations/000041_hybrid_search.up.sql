-- 000041_hybrid_search.up.sql
--
-- Hybrid retrieval: lexical (BM25) scoring over the campus corpus.
--
-- The AI layer fuses a lexical arm with the existing Weaviate vector arm.
-- BM25 needs term frequencies and document frequencies, so this adds
-- generated tsvector columns plus GIN indexes for the searchable entities.
-- ts_rank_cd gives a cheap candidate filter; the authoritative BM25 ordering
-- is computed in Go (internal/service/search) over this same corpus.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Events -----------------------------------------------------------------
ALTER TABLE events
	ADD COLUMN IF NOT EXISTS search_vector tsvector
	GENERATED ALWAYS AS (
		setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
		setweight(to_tsvector('english', coalesce(description, '')), 'B') ||
		setweight(to_tsvector('english', coalesce(venue, '')), 'C') ||
		setweight(to_tsvector('english', coalesce(category, '')), 'C')
	) STORED;

CREATE INDEX IF NOT EXISTS idx_events_search_vector
	ON events USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_events_title_trgm
	ON events USING GIN (title gin_trgm_ops);

-- Clubs ------------------------------------------------------------------
ALTER TABLE clubs
	ADD COLUMN IF NOT EXISTS search_vector tsvector
	GENERATED ALWAYS AS (
		setweight(to_tsvector('english', coalesce(name, '')), 'A') ||
		setweight(to_tsvector('english', coalesce(description, '')), 'B') ||
		setweight(to_tsvector('english', coalesce(category, '')), 'C')
	) STORED;

CREATE INDEX IF NOT EXISTS idx_clubs_search_vector
	ON clubs USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_clubs_name_trgm
	ON clubs USING GIN (name gin_trgm_ops);

-- Opportunities ----------------------------------------------------------
ALTER TABLE opportunities
	ADD COLUMN IF NOT EXISTS search_vector tsvector
	GENERATED ALWAYS AS (
		setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
		setweight(to_tsvector('english', coalesce(description, '')), 'B') ||
		setweight(to_tsvector('english', coalesce(type, '')), 'C')
	) STORED;

CREATE INDEX IF NOT EXISTS idx_opportunities_search_vector
	ON opportunities USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_opportunities_title_trgm
	ON opportunities USING GIN (title gin_trgm_ops);

-- Announcements ----------------------------------------------------------
ALTER TABLE announcements
	ADD COLUMN IF NOT EXISTS search_vector tsvector
	GENERATED ALWAYS AS (
		setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
		setweight(to_tsvector('english', coalesce(body, '')), 'B')
	) STORED;

CREATE INDEX IF NOT EXISTS idx_announcements_search_vector
	ON announcements USING GIN (search_vector);

-- Campus resources -------------------------------------------------------
ALTER TABLE campus_resources
	ADD COLUMN IF NOT EXISTS search_vector tsvector
	GENERATED ALWAYS AS (
		setweight(to_tsvector('english', coalesce(name, '')), 'A') ||
		setweight(to_tsvector('english', coalesce(description, '')), 'B') ||
		setweight(to_tsvector('english', coalesce(location, '')), 'C') ||
		setweight(to_tsvector('english', coalesce(resource_type, '')), 'C')
	) STORED;

CREATE INDEX IF NOT EXISTS idx_campus_resources_search_vector
	ON campus_resources USING GIN (search_vector);

-- Discussions ------------------------------------------------------------
ALTER TABLE discussions
	ADD COLUMN IF NOT EXISTS search_vector tsvector
	GENERATED ALWAYS AS (
		setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
		setweight(to_tsvector('english', coalesce(body, '')), 'B') ||
		setweight(to_tsvector('english', coalesce(category, '')), 'C')
	) STORED;

CREATE INDEX IF NOT EXISTS idx_discussions_search_vector
	ON discussions USING GIN (search_vector);