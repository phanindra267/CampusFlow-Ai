
CREATE TABLE IF NOT EXISTS global_strategic_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    objectives JSONB,
    assumptions JSONB,
    status VARCHAR(50) DEFAULT 'DRAFT', -- DRAFT, APPROVED, ACTIVE, ARCHIVED
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS decision_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    topic VARCHAR(255) NOT NULL,
    evidence JSONB,
    ai_recommendations JSONB,
    human_opinions JSONB,
    final_decision JSONB,
    status VARCHAR(50) DEFAULT 'OPEN', -- OPEN, DELIBERATING, RESOLVED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS evidence_graphs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    claim VARCHAR(500) NOT NULL,
    source_reference VARCHAR(255),
    dataset_version VARCHAR(100),
    confidence FLOAT,
    counterevidence JSONB,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS autonomous_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    policy_name VARCHAR(255) NOT NULL,
    allowed_scope JSONB,
    risk_level VARCHAR(50),
    mode VARCHAR(50) DEFAULT 'SHADOW', -- SHADOW, CANARY, ACTIVE
    rollback_plan TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
