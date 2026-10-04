-- Community-facing extensions to campus facilities and services.
--
-- 000007 created campus_resources for institutional inventory (resource_type of
-- LAB/ROOM/EQUIPMENT/VENUE) with a free-form metadata column. A community needs
-- to browse and book these, which needs typed, filterable columns rather than
-- keys buried in JSONB. These are added rather than replacing the table so the
-- existing reservations and inventory data keep working.

ALTER TABLE campus_resources
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS location VARCHAR(255),
    ADD COLUMN IF NOT EXISTS is_bookable BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS requires_auth BOOLEAN NOT NULL DEFAULT FALSE;

-- resource_type previously held LAB/ROOM/EQUIPMENT/VENUE. Widen it to the
-- service categories a community browses, keeping the original values valid.
ALTER TABLE campus_resources
    DROP CONSTRAINT IF EXISTS campus_resources_resource_type_check;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'campus_resources_resource_type_check'
    ) THEN
        ALTER TABLE campus_resources
            ADD CONSTRAINT campus_resources_resource_type_check
            CHECK (resource_type IN (
                'LAB', 'ROOM', 'EQUIPMENT', 'VENUE',
                'STUDY_SPACE', 'FOOD', 'TRANSPORT', 'HEALTH', 'LIBRARY', 'SPORTS', 'SERVICE'
            ));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_campus_resources_type_status
    ON campus_resources (resource_type, status);

-- Member bookings for bookable resources. Kept separate from
-- resource_reservations, which records event-linked holds made by organisers.
CREATE TABLE IF NOT EXISTS resource_bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_id UUID NOT NULL REFERENCES campus_resources(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    note VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'CONFIRMED'
        CHECK (status IN ('CONFIRMED', 'CANCELLED')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT resource_bookings_valid_window CHECK (end_time > start_time)
);

-- A member cannot double-book the same resource, and no two bookings for one
-- resource may overlap. The exclusion constraint enforces this in the database
-- rather than relying on a racy read-then-write check in the handler.
CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE resource_bookings
    DROP CONSTRAINT IF EXISTS resource_bookings_no_overlap;

ALTER TABLE resource_bookings
    ADD CONSTRAINT resource_bookings_no_overlap
    EXCLUDE USING gist (
        resource_id WITH =,
        tstzrange(start_time, end_time, '[)') WITH &&
    )
    WHERE (status = 'CONFIRMED');

CREATE INDEX IF NOT EXISTS idx_resource_bookings_user
    ON resource_bookings (user_id, start_time DESC);