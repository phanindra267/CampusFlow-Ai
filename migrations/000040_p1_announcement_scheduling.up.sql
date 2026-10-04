-- P1 refinements: announcement pinning and scheduling, service-desk ownership.
--
-- Migration 000037 gave announcements their shape, but a campus notice board
-- needs pinning (the notice everyone must read stays at the top) and a
-- scheduled publish/expire window (a notice about tomorrow's closure should not
-- appear today and should stop appearing next week).
--
-- The helpdesk also needs an owner: without assigned_to there is no way to answer
-- "who is dealing with my complaint".

-- ------------------------------------------------------- announcements ----
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS is_pinned BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS publish_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP WITH TIME ZONE;

-- The audience set is widened from the internal-role vocabulary to the groups a
-- campus actually addresses in a notice. MEMBERS, ORGANISERS and ADMINS are
-- kept so existing rows stay valid.
ALTER TABLE announcements
    DROP CONSTRAINT IF EXISTS announcements_audience_check;
ALTER TABLE announcements
    ADD CONSTRAINT announcements_audience_check CHECK (audience IN (
        'ALL', 'MEMBERS', 'STUDENTS', 'FACULTY', 'STAFF', 'CLUBS',
        'ORGANISERS', 'ADMINS'
    ));

-- published_at is retained as the record of when a notice actually went live, so
-- a scheduled notice is not treated as published before its time.
CREATE INDEX IF NOT EXISTS idx_announcements_visible
    ON announcements (is_pinned DESC, published_at DESC)
    WHERE status = 'PUBLISHED';

-- ---------------------------------------------------- service requests ----
ALTER TABLE service_requests
    ADD COLUMN IF NOT EXISTS assigned_to UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_service_requests_assignee
    ON service_requests (assigned_to, created_at DESC)
    WHERE assigned_to IS NOT NULL;

-- A complaint is worth surfacing on the dashboard when it is urgent and still
-- open after two days, so the partial index covers exactly that case.
CREATE INDEX IF NOT EXISTS idx_service_requests_urgent_open
    ON service_requests (priority, created_at)
    WHERE priority IN ('HIGH', 'URGENT')
      AND status NOT IN ('RESOLVED', 'CLOSED', 'REJECTED');

-- ------------------------------------------------- application statuses ----
-- ACCEPTED was missing from the original vocabulary, which left a poster unable
-- to record an offer they had actually made. Any status outside this set is
-- folded into WITHDRAWN first so the new CHECK can be added without failing on
-- rows written by older code paths.
UPDATE opportunity_applications
SET status = 'WITHDRAWN'
WHERE status IS NULL
   OR status NOT IN (
        'SUBMITTED', 'UNDER_REVIEW', 'SHORTLISTED', 'ACCEPTED', 'REJECTED', 'WITHDRAWN'
   );

ALTER TABLE opportunity_applications
    DROP CONSTRAINT IF EXISTS opportunity_applications_status_check;
ALTER TABLE opportunity_applications
    ADD CONSTRAINT opportunity_applications_status_check CHECK (status IN (
        'SUBMITTED', 'UNDER_REVIEW', 'SHORTLISTED', 'ACCEPTED', 'REJECTED', 'WITHDRAWN'
    ));

-- ------------------------------------------------ opportunity type set ----
-- Volunteering drives and startup offers are posted on the same surface as
-- internships, so the closed set has to include them or the API would reject a
-- legitimate listing.
ALTER TABLE opportunities
    DROP CONSTRAINT IF EXISTS opportunities_type_check;
ALTER TABLE opportunities
    ADD CONSTRAINT opportunities_type_check CHECK (type IN (
        'HACKATHON', 'COMPETITION', 'WORKSHOP', 'SEMINAR', 'INTERNSHIP',
        'PLACEMENT', 'RESEARCH', 'PROJECT', 'ENTREPRENEURSHIP', 'SCHOLARSHIP',
        'JOB', 'VOLUNTEERING', 'STARTUP'
    ));