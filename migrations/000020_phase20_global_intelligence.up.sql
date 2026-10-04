
CREATE TABLE IF NOT EXISTS global_trust_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL, -- INSTITUTION, AGENT, MODEL, USER
    entity_id VARCHAR(255) NOT NULL,
    trust_score FLOAT, -- Dynamic, multi-dimensional score representation
    identity_assurance VARCHAR(50) DEFAULT 'STANDARD',
    security_posture VARCHAR(50) DEFAULT 'UNKNOWN',
    governance_status VARCHAR(50) DEFAULT 'ACTIVE',
    evidence JSONB,
    last_evaluated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS global_agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    certification_status VARCHAR(50) DEFAULT 'REGISTERED',
    capabilities JSONB,
    trust_entity_id UUID REFERENCES global_trust_entities(id) ON DELETE SET NULL,
    resource_requirements JSONB,
    supported_protocols JSONB,
    authority_level VARCHAR(50) DEFAULT 'OBSERVE',
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS human_ai_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    objective TEXT NOT NULL,
    members JSONB NOT NULL, -- Array of human and agent references
    permissions JSONB,
    deadline TIMESTAMP WITH TIME ZONE,
    success_criteria JSONB,
    performance_metrics JSONB,
    status VARCHAR(50) DEFAULT 'FORMED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS collective_reasoning_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    initiator_tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    topic TEXT NOT NULL,
    hypotheses JSONB,
    evidence JSONB,
    assigned_teams JSONB,
    consensus_state VARCHAR(50) DEFAULT 'GATHERING',
    adversarial_reviews JSONB,
    final_recommendation JSONB,
    confidence FLOAT,
    dissenting_opinions JSONB,
    status VARCHAR(50) DEFAULT 'IN_PROGRESS',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS global_forecasts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject VARCHAR(255) NOT NULL,
    ensemble_members JSONB,
    prediction JSONB,
    confidence FLOAT,
    calibration_score FLOAT,
    disagreement_details JSONB,
    valid_until TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS global_risks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    threat VARCHAR(255) NOT NULL,
    probability FLOAT,
    impact VARCHAR(50),
    exposure JSONB,
    mitigation TEXT,
    systemic_dependencies JSONB,
    owner_id VARCHAR(255),
    confidence FLOAT,
    status VARCHAR(50) DEFAULT 'IDENTIFIED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
