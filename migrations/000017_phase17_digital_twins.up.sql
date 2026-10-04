
CREATE TABLE IF NOT EXISTS digital_twins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    twin_type VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    lifecycle_state VARCHAR(50) DEFAULT 'ACTIVE',
    version INT NOT NULL DEFAULT 1,
    current_state JSONB,
    desired_state JSONB,
    data_sources JSONB,
    relationships JSONB,
    policies JSONB,
    confidence FLOAT,
    provenance JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS twin_state_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    twin_id UUID NOT NULL REFERENCES digital_twins(id) ON DELETE CASCADE,
    state_type VARCHAR(50) NOT NULL,
    state_data JSONB NOT NULL,
    source VARCHAR(100),
    confidence FLOAT,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS simulation_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    twin_id UUID REFERENCES digital_twins(id) ON DELETE SET NULL,
    scenario_name VARCHAR(255) NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    model_version VARCHAR(50) NOT NULL,
    parameters JSONB,
    assumptions JSONB,
    random_seed BIGINT,
    status VARCHAR(50) DEFAULT 'QUEUED',
    priority INT DEFAULT 5,
    cpu_limit_cores FLOAT,
    memory_limit_mb INT,
    runtime_limit_seconds INT,
    cost_budget NUMERIC(18, 2),
    results JSONB,
    provenance JSONB,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS simulation_scenarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    baseline_twin_id UUID REFERENCES digital_twins(id) ON DELETE SET NULL,
    version INT NOT NULL DEFAULT 1,
    parameters JSONB,
    assumptions JSONB,
    is_ai_generated BOOLEAN DEFAULT FALSE,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
