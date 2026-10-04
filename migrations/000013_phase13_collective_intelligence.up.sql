
CREATE TABLE IF NOT EXISTS deliberation_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    problem TEXT NOT NULL,
    context JSONB,
    participants JSONB,
    evidence JSONB,
    arguments JSONB,
    simulations JSONB,
    status VARCHAR(50) DEFAULT 'OPEN', -- OPEN, DELIBERATING, DECIDED, CLOSED
    final_outcome JSONB,
    decision_owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS argument_graphs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES deliberation_rooms(id) ON DELETE CASCADE,
    claim TEXT NOT NULL,
    supporting_evidence JSONB,
    counter_evidence JSONB,
    assumptions JSONB,
    counterarguments JSONB,
    contributed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS decision_outcomes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES deliberation_rooms(id) ON DELETE CASCADE,
    decision_summary TEXT NOT NULL,
    expected_outcome JSONB,
    observed_outcome JSONB,
    outcome_delta JSONB,
    lessons_learned JSONB,
    status VARCHAR(50) DEFAULT 'PENDING', -- PENDING, OBSERVED, ANALYZED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS institutional_playbooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL DEFAULT '1.0',
    category VARCHAR(100),
    steps JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'DRAFT', -- DRAFT, APPROVED, ACTIVE, DEPRECATED
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS prediction_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    model_name VARCHAR(100) NOT NULL,
    model_version VARCHAR(50),
    inputs JSONB,
    prediction JSONB NOT NULL,
    confidence FLOAT,
    observed_outcome JSONB,
    error_delta FLOAT,
    predicted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    observed_at TIMESTAMP WITH TIME ZONE
);
