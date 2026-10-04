-- Community layer for CampusCare AI.
--
-- CampusCare AI is a community platform, not an academic records system. This
-- migration adds the collaboration primitives the community product needs:
-- threaded discussions, member notifications, and opportunity applications.
-- It deliberately contains no attendance, grades, courses or timetable tables.

-- A discussion thread. A NULL club_id means the thread is campus-wide.
CREATE TABLE IF NOT EXISTS discussions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_id UUID REFERENCES clubs(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'GENERAL'
        CHECK (category IN ('GENERAL', 'ANNOUNCEMENT', 'QUESTION', 'EVENT', 'OPPORTUNITY', 'RESEARCH')),
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN'
        CHECK (status IN ('OPEN', 'LOCKED', 'ARCHIVED')),
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    reply_count INT NOT NULL DEFAULT 0 CHECK (reply_count >= 0),
    last_activity_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_discussions_club ON discussions (club_id, last_activity_at DESC);
CREATE INDEX IF NOT EXISTS idx_discussions_campus ON discussions (last_activity_at DESC)
    WHERE club_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_discussions_author ON discussions (author_id);

-- Replies support one level of threading via parent_id.
CREATE TABLE IF NOT EXISTS discussion_replies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    discussion_id UUID NOT NULL REFERENCES discussions(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES discussion_replies(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT discussion_replies_not_self_parent CHECK (parent_id IS DISTINCT FROM id)
);

CREATE INDEX IF NOT EXISTS idx_discussion_replies_discussion
    ON discussion_replies (discussion_id, created_at);

-- In-app notifications. read_at being NULL means unread.
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL
        CHECK (type IN ('EVENT', 'CLUB', 'OPPORTUNITY', 'DISCUSSION', 'SYSTEM')),
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    link VARCHAR(512),
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
    ON notifications (user_id, created_at DESC)
    WHERE read_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_notifications_user_all
    ON notifications (user_id, created_at DESC);

-- Applications to internships, research roles, workshops and so on.
CREATE TABLE IF NOT EXISTS opportunity_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'SUBMITTED'
        CHECK (status IN ('SUBMITTED', 'UNDER_REVIEW', 'SHORTLISTED', 'REJECTED', 'WITHDRAWN')),
    cover_note TEXT,
    resume_url VARCHAR(512),
    applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (opportunity_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_opportunity_applications_user
    ON opportunity_applications (user_id, applied_at DESC);
CREATE INDEX IF NOT EXISTS idx_opportunity_applications_opportunity
    ON opportunity_applications (opportunity_id, status);

-- Denormalised counters keep club and opportunity listings to a single query.
ALTER TABLE clubs
    ADD COLUMN IF NOT EXISTS member_count INT NOT NULL DEFAULT 0;

ALTER TABLE opportunities
    ADD COLUMN IF NOT EXISTS application_count INT NOT NULL DEFAULT 0;