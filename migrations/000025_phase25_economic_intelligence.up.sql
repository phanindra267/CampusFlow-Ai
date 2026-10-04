
CREATE TABLE IF NOT EXISTS econ_datasets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_name VARCHAR(255) NOT NULL,
    publisher VARCHAR(255),
    version VARCHAR(50) NOT NULL,
    methodology TEXT,
    license_info TEXT,
    geographic_scope VARCHAR(255),
    update_frequency VARCHAR(50),
    provenance JSONB,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS econ_indicators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id UUID REFERENCES econ_datasets(id) ON DELETE CASCADE,
    indicator_type VARCHAR(100) NOT NULL, -- EMPLOYMENT, WAGE, SKILL_DEMAND, INNOVATION
    region VARCHAR(255),
    time_period TIMESTAMP WITH TIME ZONE NOT NULL,
    value FLOAT NOT NULL,
    uncertainty_range JSONB,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS econ_regional_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region VARCHAR(255) NOT NULL,
    skill_name VARCHAR(255) NOT NULL,
    workforce_demand FLOAT,
    education_supply FLOAT,
    balance_status VARCHAR(50), -- SHORTAGE, SURPLUS, BALANCED
    confidence FLOAT,
    last_computed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS econ_tech_trends (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    technology_name VARCHAR(255) NOT NULL,
    maturity_level VARCHAR(50), -- EMERGING, EARLY, MATURE, DECLINING
    affected_occupations JSONB,
    task_automation_risk JSONB,
    human_augmentation_potential JSONB,
    evidence JSONB,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS econ_scenarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    assumptions JSONB NOT NULL,
    model_version VARCHAR(50),
    status VARCHAR(50) DEFAULT 'QUEUED', -- QUEUED, RUNNING, COMPLETED, FAILED, CANCELLED
    outputs JSONB,
    policy_implications JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS econ_funding_opportunities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    funding_type VARCHAR(50), -- SCHOLARSHIP, RESEARCH_GRANT, STARTUP
    provider_name VARCHAR(255),
    eligibility_rules JSONB,
    deadline TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'OPEN',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
