-- P1 profile and people directory.
--
-- The profile stays campus and career oriented. Program, branch and year of
-- study are optional, self-reported context used for eligibility and matching,
-- not an academic record. There are no marks, grades, credits or attendance
-- columns here.

-- ------------------------------------------------- member profile fields ----
ALTER TABLE users ADD COLUMN IF NOT EXISTS identifier VARCHAR(64);
ALTER TABLE users ADD COLUMN IF NOT EXISTS program VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS branch VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS year_of_study INT
    CHECK (year_of_study IS NULL OR year_of_study BETWEEN 1 AND 10);
ALTER TABLE users ADD COLUMN IF NOT EXISTS headline VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(512);
ALTER TABLE users ADD COLUMN IF NOT EXISTS career_interests TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE users ADD COLUMN IF NOT EXISTS open_to_opportunities BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS idx_users_program ON users (program) WHERE program IS NOT NULL;

-- Research interests are tracked separately from general interests so that
-- research matching can read one list without filtering.
CREATE TABLE IF NOT EXISTS user_research_interests (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    interest VARCHAR(150) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, interest)
);

CREATE INDEX IF NOT EXISTS idx_user_research_interests_interest
    ON user_research_interests (interest);

-- ------------------------------------------------------- people directory --
-- Faculty, researchers, supervisors and mentors that members can discover for
-- research collaboration. user_id is set when the person has a CampusCare
-- account; it is nullable because directory entries may be seeded for people
-- who do not use the platform.
CREATE TABLE IF NOT EXISTS people (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE REFERENCES users(id) ON DELETE SET NULL,
    display_name VARCHAR(255) NOT NULL,
    person_role VARCHAR(50) NOT NULL DEFAULT 'FACULTY'
        CHECK (person_role IN ('FACULTY', 'RESEARCHER', 'SUPERVISOR', 'MENTOR', 'COORDINATOR')),
    school VARCHAR(255),
    designation VARCHAR(255),
    bio TEXT,
    email VARCHAR(255),
    office_location VARCHAR(255),
    profile_url VARCHAR(512),
    research_interests TEXT[] NOT NULL DEFAULT '{}',
    accepting_students BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE', 'INACTIVE')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_people_name ON people (display_name);
CREATE INDEX IF NOT EXISTS idx_people_role ON people (person_role, status);
CREATE INDEX IF NOT EXISTS idx_people_interests ON people USING GIN (research_interests);
CREATE INDEX IF NOT EXISTS idx_people_search ON people
    USING GIN (to_tsvector('english', display_name || ' ' || COALESCE(bio, '')));