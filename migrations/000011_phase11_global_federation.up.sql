
CREATE TABLE IF NOT EXISTS federation_trusts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    target_tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    trust_level VARCHAR(50) DEFAULT 'NONE', -- NONE, DISCOVERY, COLLABORATION, FULL
    allowed_scopes JSONB,
    status VARCHAR(50) DEFAULT 'PENDING', -- PENDING, ACTIVE, SUSPENDED, REVOKED
    established_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source_tenant_id, target_tenant_id)
);

CREATE TABLE IF NOT EXISTS research_workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    visibility VARCHAR(50) DEFAULT 'PRIVATE', -- PRIVATE, FEDERATED_RESTRICTED, PUBLIC
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workspace_members (
    workspace_id UUID REFERENCES research_workspaces(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE, -- Allows federated members
    role VARCHAR(50) NOT NULL, -- VIEWER, CONTRIBUTOR, ADMIN
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (workspace_id, user_id)
);

CREATE TABLE IF NOT EXISTS research_experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES research_workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    configuration JSONB,
    dataset_version VARCHAR(100),
    model_version VARCHAR(100),
    status VARCHAR(50) DEFAULT 'PLANNED', -- PLANNED, RUNNING, COMPLETED, FAILED
    results JSONB,
    executed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ai_incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    incident_type VARCHAR(100) NOT NULL, -- HALLUCINATION, DATA_LEAK, POLICY_VIOLATION, TOOL_ABUSE
    severity VARCHAR(50) NOT NULL, -- LOW, MEDIUM, HIGH, CRITICAL
    affected_agent_id UUID, -- Optional link to agent
    description TEXT,
    evidence JSONB,
    status VARCHAR(50) DEFAULT 'OPEN', -- OPEN, INVESTIGATING, CONTAINED, RESOLVED
    detected_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
