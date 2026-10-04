
CREATE TABLE IF NOT EXISTS trusted_institutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    institution_name VARCHAR(255) NOT NULL,
    verification_status VARCHAR(50) DEFAULT 'UNVERIFIED',
    federation_status VARCHAR(50) DEFAULT 'NONE',
    shared_capabilities JSONB,
    trust_policy JSONB,
    security_status VARCHAR(50) DEFAULT 'UNKNOWN',
    admin_owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS agent_tools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    purpose TEXT,
    risk_level VARCHAR(50) DEFAULT 'LOW',
    required_permission VARCHAR(255),
    resource_cost JSONB,
    failure_behavior TEXT,
    version VARCHAR(50) NOT NULL DEFAULT '1.0',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS service_catalog (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    service_name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    availability_sla VARCHAR(100),
    api_endpoint VARCHAR(500),
    access_policy JSONB,
    cost_model JSONB,
    health_status VARCHAR(50) DEFAULT 'UNKNOWN',
    version VARCHAR(50) NOT NULL DEFAULT '1.0',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS automation_loop_guards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    workflow_id VARCHAR(255) NOT NULL,
    invocation_count INT DEFAULT 0,
    max_invocations INT DEFAULT 10,
    window_seconds INT DEFAULT 3600,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS governance_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subject VARCHAR(255) NOT NULL,
    subject_type VARCHAR(100) NOT NULL,
    review_criteria JSONB,
    approvers JSONB,
    evidence JSONB,
    decision VARCHAR(50),
    conditions JSONB,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
