-- 000041_hybrid_search.down.sql
--
-- Removes the generated tsvector columns and GIN indexes added for hybrid
-- retrieval. The pg_trgm extension is left in place: other migrations and
-- operators may rely on it and dropping it is not part of this rollback.

DROP INDEX IF EXISTS idx_discussions_search_vector;
DROP INDEX IF EXISTS idx_campus_resources_search_vector;
DROP INDEX IF EXISTS idx_announcements_search_vector;
DROP INDEX IF EXISTS idx_opportunities_title_trgm;
DROP INDEX IF EXISTS idx_opportunities_search_vector;
DROP INDEX IF EXISTS idx_clubs_name_trgm;
DROP INDEX IF EXISTS idx_clubs_search_vector;
DROP INDEX IF EXISTS idx_events_title_trgm;
DROP INDEX IF EXISTS idx_events_search_vector;

ALTER TABLE discussions DROP COLUMN IF EXISTS search_vector;
ALTER TABLE campus_resources DROP COLUMN IF EXISTS search_vector;
ALTER TABLE announcements DROP COLUMN IF EXISTS search_vector;
ALTER TABLE opportunities DROP COLUMN IF EXISTS search_vector;
ALTER TABLE clubs DROP COLUMN IF EXISTS search_vector;
ALTER TABLE events DROP COLUMN IF EXISTS search_vector;