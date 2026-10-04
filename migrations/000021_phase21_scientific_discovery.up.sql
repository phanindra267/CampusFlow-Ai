
CREATE TABLE IF NOT EXISTS scientific_hypotheses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    statement TEXT NOT NULL,
    context TEXT,
    supporting_evidence JSONB,
    contradicting_evidence JSONB,
    assumptions JSONB,
    expected_predictions JSONB,
    confidence FLOAT,
    owner_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'PROPOSED', -- PROPOSED, INVESTIGATING, SUPPORTED, REFUTED, INCONCLUSIVE
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS scientific_experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hypothesis_id UUID REFERENCES scientific_hypotheses(id) ON DELETE CASCADE,
    objective TEXT NOT NULL,
    variables JSONB,
    controls JSONB,
    measurements JSONB,
    sample_requirements JSONB,
    expected_outcomes JSONB,
    risks JSONB,
    stop_conditions JSONB,
    reproducibility_config JSONB,
    status VARCHAR(50) DEFAULT 'DESIGNED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS intelligence_improvements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_component VARCHAR(100) NOT NULL, -- MODEL, PROMPT, POLICY, AGENT
    current_state JSONB,
    proposed_change JSONB,
    expected_benefit TEXT,
    expected_risk TEXT,
    validation_plan JSONB,
    rollback_plan JSONB,
    sandbox_results JSONB,
    status VARCHAR(50) DEFAULT 'PROPOSED', -- PROPOSED, VALIDATING, APPROVED, CANARY, PRODUCTION, REJECTED, ROLLED_BACK
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS causal_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_entity VARCHAR(255) NOT NULL,
    target_entity VARCHAR(255) NOT NULL,
    relationship_type VARCHAR(50) DEFAULT 'HYPOTHESIZED', -- OBSERVED, INFERRED, HYPOTHESIZED, SUPPORTED
    evidence_strength FLOAT,
    provenance JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS research_artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    artifact_type VARCHAR(50) NOT NULL, -- PAPER, DATASET, CODE, MODEL
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    provenance JSONB,
    storage_ref VARCHAR(255),
    security_scan_results JSONB,
    status VARCHAR(50) DEFAULT 'REGISTERED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
