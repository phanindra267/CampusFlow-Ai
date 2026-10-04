
CREATE TABLE IF NOT EXISTS student_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    student_id VARCHAR(255) NOT NULL,
    academic_standing VARCHAR(50),
    enrolled_courses JSONB,
    mastered_skills JSONB,
    learning_preferences JSONB,
    privacy_controls JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS edu_learning_paths (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_profile_id UUID REFERENCES student_profiles(id) ON DELETE CASCADE,
    objective TEXT NOT NULL,
    target_competency VARCHAR(100),
    topics JSONB,
    resources JSONB,
    expected_duration_hours INT,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS edu_interventions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_profile_id UUID REFERENCES student_profiles(id) ON DELETE CASCADE,
    detected_challenge TEXT NOT NULL,
    recommended_action TEXT,
    evidence JSONB,
    advisor_review_status VARCHAR(50) DEFAULT 'PENDING', -- PENDING, APPROVED, MODIFIED, REJECTED
    student_feedback JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS edu_assessments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    assessment_mode VARCHAR(50) DEFAULT 'PRACTICE', -- PRACTICE, OPEN_BOOK, AI_ASSISTED, RESTRICTED, PROCTORED
    ai_assistance_policy TEXT,
    questions JSONB,
    item_analytics JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS edu_opportunities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    opportunity_type VARCHAR(50), -- INTERNSHIP, RESEARCH, CLUB, WORKSHOP
    title VARCHAR(255) NOT NULL,
    required_skills JSONB,
    target_skill_gaps JSONB,
    status VARCHAR(50) DEFAULT 'OPEN',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
