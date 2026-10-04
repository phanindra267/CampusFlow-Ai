
CREATE TABLE IF NOT EXISTS ecosystem_institutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_type VARCHAR(100) NOT NULL, -- UNIVERSITY, COMPANY, STARTUP, GOVERNMENT, NONPROFIT
    legal_name VARCHAR(255) NOT NULL,
    domains JSONB,
    capabilities JSONB,
    governance_config JSONB,
    privacy_level VARCHAR(50) DEFAULT 'INTERNAL',
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ecosystem_partnerships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    initiator_institution_id UUID REFERENCES ecosystem_institutions(id),
    partner_institution_id UUID REFERENCES ecosystem_institutions(id),
    partnership_type VARCHAR(100) NOT NULL, -- RESEARCH, EDUCATION, WORKFORCE, INNOVATION, FUNDING
    lifecycle_status VARCHAR(50) DEFAULT 'DISCOVERY', -- DISCOVERY, EVALUATION, NEGOTIATION, ACTIVE, CLOSED
    objectives JSONB,
    evidence JSONB,
    risk_assessment JSONB,
    data_sharing_contract JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ecosystem_strategic_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID REFERENCES ecosystem_institutions(id),
    version VARCHAR(50) NOT NULL DEFAULT '1.0',
    mission TEXT,
    objectives JSONB NOT NULL,
    initiatives JSONB,
    kpis JSONB,
    risks JSONB,
    scenarios JSONB,
    status VARCHAR(50) DEFAULT 'DRAFT',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ecosystem_early_warnings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID REFERENCES ecosystem_institutions(id),
    signal_type VARCHAR(100) NOT NULL, -- TECHNOLOGY_SHIFT, SKILL_SHORTAGE, FUNDING_CHANGE
    signal_data JSONB NOT NULL,
    evidence JSONB,
    confidence FLOAT,
    time_period VARCHAR(100),
    review_status VARCHAR(50) DEFAULT 'NEW',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ecosystem_contingency_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID REFERENCES ecosystem_institutions(id),
    trigger_description TEXT NOT NULL,
    response_plan JSONB NOT NULL,
    owner_id VARCHAR(255),
    resources JSONB,
    approval_status VARCHAR(50) DEFAULT 'DRAFT',
    review_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
