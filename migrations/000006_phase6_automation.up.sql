
CREATE TABLE IF NOT EXISTS automation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    trigger_type VARCHAR(100) NOT NULL, -- e.g., OPPORTUNITY_CREATED, DEADLINE_APPROACHING
    conditions JSONB, -- JSON representation of rule conditions
    action_type VARCHAR(100) NOT NULL, -- e.g., CREATE_NOTIFICATION, EMAIL_DIGEST
    action_payload JSONB, -- Details of the action to execute
    status VARCHAR(50) DEFAULT 'ACTIVE',
    execution_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_executed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS automation_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES automation_rules(id) ON DELETE CASCADE,
    trigger_event_id VARCHAR(255),
    status VARCHAR(50) NOT NULL, -- SUCCESS, FAILED
    error_message TEXT,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    email_enabled BOOLEAN DEFAULT TRUE,
    push_enabled BOOLEAN DEFAULT TRUE,
    digest_frequency VARCHAR(50) DEFAULT 'DAILY', -- NONE, DAILY, WEEKLY
    fatigue_threshold INT DEFAULT 5, -- Max notifications per day before squashing
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ai_decision_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    decision_type VARCHAR(100) NOT NULL, -- NOTIFICATION_DECISION, RECOMMENDATION_FILTER
    trigger_source VARCHAR(255),
    relevant_signals JSONB,
    outcome VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS knowledge_gaps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_pattern VARCHAR(255) NOT NULL,
    frequency INT DEFAULT 1,
    status VARCHAR(50) DEFAULT 'DETECTED', -- DETECTED, RESOLVED
    first_detected_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_detected_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
