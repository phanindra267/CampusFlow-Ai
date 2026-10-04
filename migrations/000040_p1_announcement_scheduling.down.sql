-- Rollback of 000040. Announcement pinning, scheduling and audience widening are
-- reverted to the 000037 shape, and service-request ownership is dropped.
--
-- Note: any announcements created while the widened audience or the scheduling
-- columns were in use must be re-saved onto the narrowed vocabulary before this
-- runs, because the restored CHECK constraint rejects the wider values.

ALTER TABLE opportunities
    DROP CONSTRAINT IF EXISTS opportunities_type_check;

UPDATE opportunities
SET type = 'WORKSHOP'
WHERE type NOT IN (
    'HACKATHON', 'COMPETITION', 'WORKSHOP', 'SEMINAR', 'INTERNSHIP',
    'PLACEMENT', 'RESEARCH', 'PROJECT', 'ENTREPRENEURSHIP', 'SCHOLARSHIP', 'JOB'
);

ALTER TABLE opportunities
    ADD CONSTRAINT opportunities_type_check CHECK (type IN (
        'HACKATHON', 'COMPETITION', 'WORKSHOP', 'SEMINAR', 'INTERNSHIP',
        'PLACEMENT', 'RESEARCH', 'PROJECT', 'ENTREPRENEURSHIP', 'SCHOLARSHIP', 'JOB'
    ));

UPDATE opportunity_applications
SET status = 'REJECTED'
WHERE status = 'ACCEPTED';

ALTER TABLE opportunity_applications
    DROP CONSTRAINT IF EXISTS opportunity_applications_status_check;

ALTER TABLE opportunity_applications
    ADD CONSTRAINT opportunity_applications_status_check CHECK (status IN (
        'SUBMITTED', 'UNDER_REVIEW', 'SHORTLISTED', 'REJECTED', 'WITHDRAWN'
    ));

DROP INDEX IF EXISTS idx_service_requests_urgent_open;
DROP INDEX IF EXISTS idx_service_requests_assignee;
ALTER TABLE service_requests DROP COLUMN IF EXISTS assigned_to;

DROP INDEX IF EXISTS idx_announcements_visible;

ALTER TABLE announcements
    DROP CONSTRAINT IF EXISTS announcements_audience_check;

UPDATE announcements
SET audience = 'ALL'
WHERE audience NOT IN ('ALL', 'MEMBERS', 'ORGANISERS', 'ADMINS');

ALTER TABLE announcements
    ADD CONSTRAINT announcements_audience_check CHECK (audience IN (
        'ALL', 'MEMBERS', 'ORGANISERS', 'ADMINS'
    ));

ALTER TABLE announcements DROP COLUMN IF EXISTS expires_at;
ALTER TABLE announcements DROP COLUMN IF EXISTS publish_at;
ALTER TABLE announcements DROP COLUMN IF EXISTS is_pinned;