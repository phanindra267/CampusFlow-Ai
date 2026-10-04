-- Rollback of the member identity changes from 000036.
--
-- display_name is dropped rather than reversed because it was populated with
-- derived values; restoring the previous NULL state would lose those.

ALTER TABLE users DROP COLUMN IF EXISTS display_name;

ALTER TABLE engagement_activities RENAME COLUMN member_id TO student_id;
ALTER TABLE feedback              RENAME COLUMN member_id TO student_id;
ALTER TABLE waitlist_entries      RENAME COLUMN member_id TO student_id;
ALTER TABLE attendance_records    RENAME COLUMN member_id TO student_id;
ALTER TABLE event_registrations   RENAME COLUMN member_id TO student_id;