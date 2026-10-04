
CREATE TABLE IF NOT EXISTS global_decision_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic VARCHAR(255) NOT NULL,
    objective TEXT,
    organization_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'OPEN', -- OPEN, DELIBERATING, DECIDED, ARCHIVED
    privacy_level VARCHAR(50) DEFAULT 'INTERNAL',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS global_decision_arguments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID REFERENCES global_decision_rooms(id) ON DELETE CASCADE,
    author_id VARCHAR(255) NOT NULL,
    author_type VARCHAR(50) DEFAULT 'HUMAN', -- HUMAN, AGENT
    argument_text TEXT NOT NULL,
    argument_type VARCHAR(50) NOT NULL, -- SUPPORT, REFUTE, QUALIFY, PROPOSE
    parent_argument_id UUID REFERENCES global_decision_arguments(id),
    evidence JSONB,
    confidence FLOAT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS global_decision_forecasts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID REFERENCES global_decision_rooms(id) ON DELETE CASCADE,
    predictor_id VARCHAR(255) NOT NULL,
    predictor_type VARCHAR(50) DEFAULT 'HUMAN',
    question TEXT NOT NULL,
    forecast_value FLOAT,
    probability FLOAT,
    time_horizon TIMESTAMP WITH TIME ZONE,
    evidence JSONB,
    status VARCHAR(50) DEFAULT 'ACTIVE', -- ACTIVE, RESOLVED, CANCELLED
    actual_outcome JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS global_decision_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID REFERENCES global_decision_rooms(id) ON DELETE CASCADE,
    proposer_id VARCHAR(255) NOT NULL,
    option_title VARCHAR(255) NOT NULL,
    description TEXT,
    assumptions JSONB,
    risk_assessment JSONB,
    status VARCHAR(50) DEFAULT 'PROPOSED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS global_decision_journals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID REFERENCES global_decision_rooms(id) ON DELETE CASCADE,
    decider_id VARCHAR(255) NOT NULL,
    selected_option_id UUID REFERENCES global_decision_options(id),
    rationale TEXT NOT NULL,
    expected_outcome JSONB,
    postmortem JSONB,
    status VARCHAR(50) DEFAULT 'RECORDED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
