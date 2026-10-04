
CREATE TABLE IF NOT EXISTS dt_population_cohorts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    cohort_name VARCHAR(255) NOT NULL,
    dimensions JSONB NOT NULL,
    current_state JSONB NOT NULL,
    historical_states JSONB,
    version VARCHAR(50) DEFAULT '1.0',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dt_system_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_name VARCHAR(255) NOT NULL,
    baseline_cohorts JSONB NOT NULL,
    baseline_economics JSONB NOT NULL,
    random_seed BIGINT,
    provenance JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dt_simulations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_id UUID REFERENCES dt_system_snapshots(id),
    owner_id VARCHAR(255) NOT NULL,
    scenario_type VARCHAR(100), -- MONTE_CARLO, SENSITIVITY, DETERMINISTIC
    assumptions JSONB NOT NULL,
    shocks JSONB,
    status VARCHAR(50) DEFAULT 'QUEUED', -- QUEUED, RUNNING, COMPLETED, FAILED, CANCELLED
    progress FLOAT DEFAULT 0.0,
    results JSONB,
    uncertainty_metrics JSONB,
    parent_simulation_id UUID REFERENCES dt_simulations(id), -- For branching
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS dt_interventions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    simulation_id UUID REFERENCES dt_simulations(id) ON DELETE CASCADE,
    target_entity VARCHAR(255) NOT NULL,
    intervention_type VARCHAR(100),
    parameters JSONB NOT NULL,
    expected_mechanism TEXT,
    human_approval_status VARCHAR(50) DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dt_early_warnings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    signal_type VARCHAR(100) NOT NULL, -- SKILL_SHORTAGE, BOTTLENECK, CAPACITY_LOSS
    evidence JSONB NOT NULL,
    confidence FLOAT NOT NULL,
    time_window VARCHAR(100),
    status VARCHAR(50) DEFAULT 'DETECTED', -- DETECTED, INVESTIGATING, MITIGATED, FALSE_POSITIVE
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dt_model_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    purpose TEXT NOT NULL,
    intended_use TEXT NOT NULL,
    prohibited_use TEXT NOT NULL,
    assumptions JSONB NOT NULL,
    limitations JSONB NOT NULL,
    validation_metrics JSONB,
    approval_status VARCHAR(50) DEFAULT 'DRAFT', -- DRAFT, APPROVED, RETIRED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
