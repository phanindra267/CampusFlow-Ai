-- Reverses 000002_phase2_clubs_opportunities.

DROP TABLE IF EXISTS saved_opportunities;
DROP TABLE IF EXISTS opportunities;
DROP TABLE IF EXISTS club_announcements;
DROP TABLE IF EXISTS club_memberships;

ALTER TABLE events DROP COLUMN IF EXISTS club_id;

DROP TABLE IF EXISTS clubs;