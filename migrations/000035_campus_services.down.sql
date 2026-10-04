-- Rollback of the campus services extensions from 000035.

DROP TABLE IF EXISTS resource_bookings;

ALTER TABLE campus_resources
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS location,
    DROP COLUMN IF EXISTS is_bookable,
    DROP COLUMN IF EXISTS requires_auth;

ALTER TABLE campus_resources DROP CONSTRAINT IF EXISTS campus_resources_resource_type_check;
DROP INDEX IF EXISTS idx_campus_resources_type_status;