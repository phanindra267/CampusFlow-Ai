-- Rollback of the community layer introduced by 000033.
-- Dependent objects are dropped in reverse dependency order.

DROP TABLE IF EXISTS opportunity_applications;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS discussion_replies;
DROP TABLE IF EXISTS discussions;

ALTER TABLE opportunities DROP COLUMN IF EXISTS application_count;
ALTER TABLE clubs DROP COLUMN IF EXISTS member_count;