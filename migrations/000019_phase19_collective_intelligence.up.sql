
CREATE TABLE IF NOT EXISTS federation_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    federation_id UUID,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    institution_name VARCHAR(255) NOT NULL,
    trust_level VARCHAR(50) DEFAULT 'UNTRUSTED',
    data_sharing_policy JSONB,
    capability_profile JSONB,
    governance_policy JSONB,
    compliance_config JSONB,
    status VARCHAR(50) DEFAULT 'PENDING',
    joined_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS federation_agreements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    federation_id UUID,
    initiating_tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    parties JSONB NOT NULL,
    scope JSONB,
    obligations JSONB,
    constraints JSONB,
    duration_days INT,
    status VARCHAR(50) DEFAULT 'PROPOSED',
    version INT NOT NULL DEFAULT 1,
    approved_by JSONB,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS federated_knowledge_claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    statement TEXT NOT NULL,
    source_reference TEXT,
    evidence JSONB,
    confidence FLOAT,
    visibility VARCHAR(50) DEFAULT 'INSTITUTION',
    status VARCHAR(50) DEFAULT 'PROPOSED',
    contradictions JSONB,
    version INT NOT NULL DEFAULT 1,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS governance_proposals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    federation_id UUID,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    proposed_by_tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    proposal_type VARCHAR(100) NOT NULL,
    content JSONB,
    votes JSONB,
    quorum_required INT DEFAULT 2,
    status VARCHAR(50) DEFAULT 'OPEN',
    decision_rationale TEXT,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS federation_disputes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    federation_id UUID,
    raised_by_tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    against_tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
    subject VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    evidence JSONB,
    status VARCHAR(50) DEFAULT 'RAISED',
    resolution TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
