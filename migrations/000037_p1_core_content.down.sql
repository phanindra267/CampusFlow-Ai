-- Reverses 000037_p1_core_content.

ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_type_check;
ALTER TABLE notifications
    ADD CONSTRAINT notifications_type_check
    CHECK (type IN ('EVENT', 'CLUB', 'OPPORTUNITY', 'DISCUSSION', 'SYSTEM'));

-- Rows created under the wider category set cannot satisfy the narrower check,
-- so the extra categories are deleted before it is reinstated.
DELETE FROM notifications WHERE type IN ('SERVICE', 'ANNOUNCEMENT');

ALTER TABLE notification_preferences
    DROP CONSTRAINT IF EXISTS notification_preferences_digest_frequency_check;
ALTER TABLE notification_preferences
    ALTER COLUMN digest_frequency SET DEFAULT 'DAILY';
ALTER TABLE notification_preferences
    DROP COLUMN IF EXISTS announcements,
    DROP COLUMN IF EXISTS service_updates,
    DROP COLUMN IF EXISTS club_activity,
    DROP COLUMN IF EXISTS application_updates,
    DROP COLUMN IF EXISTS opportunity_deadlines,
    DROP COLUMN IF EXISTS event_reminders;
DROP TABLE IF EXISTS service_request_events;
DROP TABLE IF EXISTS service_request_attachments;
DROP TABLE IF EXISTS service_requests;
DROP TABLE IF EXISTS service_faqs;

ALTER TABLE opportunities DROP CONSTRAINT IF EXISTS opportunities_type_check;
ALTER TABLE opportunities DROP COLUMN IF EXISTS apply_url;
ALTER TABLE opportunities DROP COLUMN IF EXISTS source;
ALTER TABLE opportunities DROP COLUMN IF EXISTS eligibility;

DROP TABLE IF EXISTS announcements;
DROP TABLE IF EXISTS club_projects;
DROP TABLE IF EXISTS club_activities;

ALTER TABLE club_memberships DROP CONSTRAINT IF EXISTS club_memberships_status_check;
ALTER TABLE club_memberships DROP CONSTRAINT IF EXISTS club_memberships_role_check;
ALTER TABLE clubs DROP COLUMN IF EXISTS contact_email;
ALTER TABLE clubs DROP COLUMN IF EXISTS logo_url;
ALTER TABLE clubs DROP COLUMN IF EXISTS follower_count;
DROP TABLE IF EXISTS club_follows;

ALTER TABLE event_registrations DROP CONSTRAINT IF EXISTS event_registrations_status_check;
DROP TABLE IF EXISTS event_reminders;
DROP TABLE IF EXISTS saved_events;

ALTER TABLE events DROP COLUMN IF EXISTS eligibility;