
CREATE TABLE IF NOT EXISTS lifelong_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL UNIQUE,
    career_goals JSONB,
    verified_skills JSONB,
    learning_history JSONB,
    portfolio JSONB,
    privacy_controls JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workforce_skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    industry_mappings JSONB,
    related_skills JSONB,
    version VARCHAR(50) DEFAULT '1.0',
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS job_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    employer_id VARCHAR(255) NOT NULL,
    responsibilities JSONB,
    required_skills JSONB,
    preferred_skills JSONB,
    experience_requirements JSONB,
    education_requirements JSONB,
    location_preferences JSONB,
    status VARCHAR(50) DEFAULT 'OPEN',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS career_paths (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    target_role_id UUID REFERENCES job_roles(id) ON DELETE CASCADE,
    skill_gaps JSONB,
    closure_plan JSONB,
    transition_readiness FLOAT,
    estimated_effort_hours INT,
    status VARCHAR(50) DEFAULT 'SIMULATED', -- SIMULATED, ACTIVE, ABANDONED, COMPLETED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS verified_experiences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    verifier_id VARCHAR(255) NOT NULL, -- Employer or Institution
    experience_type VARCHAR(100), -- WORK, PROJECT, RESEARCH
    description TEXT,
    demonstrated_skills JSONB,
    verification_status VARCHAR(50) DEFAULT 'PENDING', -- PENDING, VERIFIED, REJECTED
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS talent_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_role_id UUID REFERENCES job_roles(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    match_score FLOAT,
    match_explanation JSONB,
    employer_status VARCHAR(50) DEFAULT 'REVIEWING',
    candidate_status VARCHAR(50) DEFAULT 'PENDING_CONSENT',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
