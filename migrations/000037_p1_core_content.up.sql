-- P1 core content for CampusCare AI.
--
-- Completes the first-priority product loop: events, clubs, opportunities and
-- campus services become fully manageable entities with saving, reminders,
-- eligibility, announcements and a real service-request helpdesk.
--
-- This migration adds no marks, grades, attendance or timetable structures.

-- ---------------------------------------------------------------- events ---
-- Who may join an event. Free text on purpose: eligibility rules differ per
-- event and are written by organisers, not validated against an academic record.
ALTER TABLE events ADD COLUMN IF NOT EXISTS eligibility TEXT;

-- Saved events are a personal shortlist, separate from registering.
CREATE TABLE IF NOT EXISTS saved_events (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, event_id)
);

CREATE INDEX IF NOT EXISTS idx_saved_events_user
    ON saved_events (user_id, created_at DESC);

-- A single reminder time per member per event. The reminders themselves are
-- materialised into notifications by the notification fan-out.
CREATE TABLE IF NOT EXISTS event_reminders (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    remind_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, event_id)
);

CREATE INDEX IF NOT EXISTS idx_event_reminders_due
    ON event_reminders (remind_at)
    WHERE remind_at IS NOT NULL;

-- Registration status is read back on every event listing, so it needs to be a
-- closed set rather than free text.
UPDATE event_registrations
SET status = 'REGISTERED'
WHERE status IS NULL OR status NOT IN ('REGISTERED', 'CANCELLED', 'ATTENDED', 'WAITLISTED');

ALTER TABLE event_registrations
    DROP CONSTRAINT IF EXISTS event_registrations_status_check;
ALTER TABLE event_registrations
    ADD CONSTRAINT event_registrations_status_check
    CHECK (status IN ('REGISTERED', 'CANCELLED', 'ATTENDED', 'WAITLISTED'));

-- ----------------------------------------------------------------- clubs ---
-- Following is interest-only and does not make someone a member.
CREATE TABLE IF NOT EXISTS club_follows (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, club_id)
);

CREATE INDEX IF NOT EXISTS idx_club_follows_club ON club_follows (club_id);

ALTER TABLE clubs ADD COLUMN IF NOT EXISTS follower_count INT NOT NULL DEFAULT 0;
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS logo_url VARCHAR(512);
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS contact_email VARCHAR(255);

ALTER TABLE club_memberships
    DROP CONSTRAINT IF EXISTS club_memberships_role_check;
ALTER TABLE club_memberships
    ADD CONSTRAINT club_memberships_role_check
    CHECK (role IN ('OWNER', 'ADMIN', 'OFFICER', 'MEMBER'));

ALTER TABLE club_memberships
    DROP CONSTRAINT IF EXISTS club_memberships_status_check;
ALTER TABLE club_memberships
    ADD CONSTRAINT club_memberships_status_check
    CHECK (status IN ('PENDING', 'ACTIVE', 'INACTIVE'));

-- Activities a club runs independently of the campus event calendar.
CREATE TABLE IF NOT EXISTS club_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    activity_type VARCHAR(100),
    location VARCHAR(255),
    starts_at TIMESTAMP WITH TIME ZONE,
    ends_at TIMESTAMP WITH TIME ZONE,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_club_activities_club
    ON club_activities (club_id, starts_at DESC);

-- Projects a club is working on.
CREATE TABLE IF NOT EXISTS club_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('PLANNED', 'ACTIVE', 'COMPLETED', 'PAUSED')),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_club_projects_club
    ON club_projects (club_id, created_at DESC);

-- ------------------------------------------------------- announcements ----
-- Campus-wide announcements published by administrators.
CREATE TABLE IF NOT EXISTS announcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'GENERAL'
        CHECK (category IN ('GENERAL', 'EVENTS', 'CLUBS', 'OPPORTUNITIES', 'SERVICES', 'CAREERS', 'RESEARCH')),
    audience VARCHAR(50) NOT NULL DEFAULT 'ALL'
        CHECK (audience IN ('ALL', 'MEMBERS', 'ORGANISERS', 'ADMINS')),
    status VARCHAR(20) NOT NULL DEFAULT 'PUBLISHED'
        CHECK (status IN ('DRAFT', 'PUBLISHED', 'ARCHIVED')),
    author_id UUID NOT NULL REFERENCES users(id),
    published_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_announcements_published
    ON announcements (published_at DESC)
    WHERE status = 'PUBLISHED';

-- -------------------------------------------------------- opportunities ---
-- The scope document requires one unified opportunity ecosystem, so the type is
-- a closed set that covers hackathons through placements and scholarships.
ALTER TABLE opportunities ADD COLUMN IF NOT EXISTS eligibility TEXT;
ALTER TABLE opportunities ADD COLUMN IF NOT EXISTS source VARCHAR(255);
ALTER TABLE opportunities ADD COLUMN IF NOT EXISTS apply_url VARCHAR(512);

UPDATE opportunities
SET type = 'WORKSHOP'
WHERE type IS NULL
   OR UPPER(type) NOT IN (
        'HACKATHON', 'COMPETITION', 'WORKSHOP', 'SEMINAR', 'INTERNSHIP',
        'PLACEMENT', 'RESEARCH', 'PROJECT', 'ENTREPRENEURSHIP', 'SCHOLARSHIP', 'JOB'
   );

ALTER TABLE opportunities
    DROP CONSTRAINT IF EXISTS opportunities_type_check;
ALTER TABLE opportunities
    ADD CONSTRAINT opportunities_type_check CHECK (type IN (
        'HACKATHON', 'COMPETITION', 'WORKSHOP', 'SEMINAR', 'INTERNSHIP',
        'PLACEMENT', 'RESEARCH', 'PROJECT', 'ENTREPRENEURSHIP', 'SCHOLARSHIP', 'JOB'
    ));

-- -------------------------------------------------------- campus services --
-- FAQs sit with the service they explain.
CREATE TABLE IF NOT EXISTS service_faqs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_id UUID NOT NULL REFERENCES campus_resources(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    position INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_service_faqs_resource
    ON service_faqs (resource_id, position);

-- Requests and complaints. Both are tracked as a ticket with a status timeline,
-- which is what members actually want: "where is my issue right now?".
CREATE TABLE IF NOT EXISTS service_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference VARCHAR(32) NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resource_id UUID REFERENCES campus_resources(id) ON DELETE SET NULL,
    kind VARCHAR(20) NOT NULL DEFAULT 'REQUEST'
        CHECK (kind IN ('REQUEST', 'COMPLAINT')),
    subject VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'NORMAL'
        CHECK (priority IN ('LOW', 'NORMAL', 'HIGH', 'URGENT')),
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN'
        CHECK (status IN ('OPEN', 'IN_PROGRESS', 'RESOLVED', 'REJECTED', 'CLOSED')),
    resolution TEXT,
    resolved_by UUID REFERENCES users(id),
    resolved_at TIMESTAMP WITH TIME ZONE,
    feedback_rating INT CHECK (feedback_rating IS NULL OR feedback_rating BETWEEN 1 AND 5),
    feedback_comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_service_requests_user
    ON service_requests (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_service_requests_status
    ON service_requests (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_service_requests_resource
    ON service_requests (resource_id)
    WHERE resource_id IS NOT NULL;

-- Attachment metadata only. Binary storage is intentionally out of scope: a
-- member records where the file lives rather than the API hosting it.
CREATE TABLE IF NOT EXISTS service_request_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES service_requests(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_url VARCHAR(512) NOT NULL,
    mime_type VARCHAR(120),
    size_bytes BIGINT,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_service_request_attachments_request
    ON service_request_attachments (request_id);

-- Immutable status timeline, which is the ticket-tracking surface.
CREATE TABLE IF NOT EXISTS service_request_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES service_requests(id) ON DELETE CASCADE,
    actor_id UUID REFERENCES users(id),
    from_status VARCHAR(20),
    to_status VARCHAR(20) NOT NULL,
    note TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_service_request_events_request
    ON service_request_events (request_id, created_at);

-- ----------------------------------------------------- notification prefs --
-- Members control which categories they hear about. The table was introduced
-- for e-mail and push channels, so the per-category toggles are added onto it
-- rather than replacing it.
ALTER TABLE notification_preferences
    ADD COLUMN IF NOT EXISTS event_reminders BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS opportunity_deadlines BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS application_updates BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS club_activity BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS service_updates BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS announcements BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE notification_preferences
SET digest_frequency = 'DAILY'
WHERE digest_frequency IS NULL OR digest_frequency NOT IN ('INSTANT', 'DAILY', 'WEEKLY', 'OFF');

ALTER TABLE notification_preferences
    ALTER COLUMN digest_frequency SET DEFAULT 'INSTANT';

ALTER TABLE notification_preferences
    DROP CONSTRAINT IF EXISTS notification_preferences_digest_frequency_check;
ALTER TABLE notification_preferences
    ADD CONSTRAINT notification_preferences_digest_frequency_check
    CHECK (digest_frequency IN ('INSTANT', 'DAILY', 'WEEKLY', 'OFF'));

-- Notifications now cover every category the product promises.
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_type_check;
ALTER TABLE notifications
    ADD CONSTRAINT notifications_type_check CHECK (type IN (
        'EVENT', 'CLUB', 'OPPORTUNITY', 'DISCUSSION', 'SERVICE', 'ANNOUNCEMENT', 'SYSTEM'
    ));