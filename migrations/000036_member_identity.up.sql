-- Member identity and terminology cleanup for the community platform.
--
-- 1. users has no display name, yet every community surface (discussion authors,
--    group rosters, event check-in) needs to show who someone is. Adding
--    display_name here is what makes those joins possible at all.
--
-- 2. The engagement tables still carry "student_id" naming inherited from an
--    academic model. These tables track event check-in, waitlists, feedback and
--    participation, so the column is a member of the community, not a student
--    in a class. They are renamed to member_id to match how the product is
--    described.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS display_name VARCHAR(120);

-- Backfill from the local part of the email address so existing accounts render
-- a sensible name immediately instead of an empty one.
UPDATE users
SET display_name = COALESCE(
    NULLIF(split_part(email, '@', 1), ''),
    'Member'
)
WHERE display_name IS NULL;

ALTER TABLE users
    ALTER COLUMN display_name SET DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_users_display_name ON users (display_name);

-- Rename the academic "student_id" columns on the participation tables.
ALTER TABLE event_registrations RENAME COLUMN student_id TO member_id;
ALTER TABLE attendance_records   RENAME COLUMN student_id TO member_id;
ALTER TABLE waitlist_entries     RENAME COLUMN student_id TO member_id;
ALTER TABLE feedback             RENAME COLUMN student_id TO member_id;
ALTER TABLE engagement_activities RENAME COLUMN student_id TO member_id;