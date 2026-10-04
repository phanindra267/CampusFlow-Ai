
CREATE TABLE IF NOT EXISTS agent_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL DEFAULT '1.0',
    purpose TEXT NOT NULL,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    autonomy_level INT NOT NULL DEFAULT 0,
    capabilities JSONB,
    permissions JSONB,
    tool_ids JSONB,
    resource_limits JSONB,
    cost_budget NUMERIC(18, 2),
    safety_policy JSONB,
    status VARCHAR(50) DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS autonomy_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    allowed_actions JSONB NOT NULL,
    forbidden_actions JSONB NOT NULL,
    required_approvals JSONB,
    resource_limits JSONB,
    cost_limits JSONB,
    risk_limits JSONB,
    retry_limits JSONB,
    escalation_conditions JSONB,
    status VARCHAR(50) DEFAULT 'DRAFT',
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS action_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    agent_id UUID REFERENCES agent_registry(id) ON DELETE SET NULL,
    goal TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    policy_version_id UUID REFERENCES autonomy_policies(id) ON DELETE SET NULL,
    steps JSONB NOT NULL,
    preconditions JSONB,
    constraints JSONB,
    expected_outcome JSONB,
    risks JSONB,
    rollback_strategy JSONB,
    status VARCHAR(50) DEFAULT 'DRAFT',
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_at TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS decision_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    agent_id UUID REFERENCES agent_registry(id) ON DELETE SET NULL,
    plan_id UUID REFERENCES action_plans(id) ON DELETE SET NULL,
    decision_type VARCHAR(100) NOT NULL,
    evidence JSONB,
    policy_version_id UUID REFERENCES autonomy_policies(id) ON DELETE SET NULL,
    explanation TEXT,
    outcome JSONB,
    human_approved BOOLEAN DEFAULT FALSE,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS lessons_learned (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subject VARCHAR(255) NOT NULL,
    incident_ref VARCHAR(255),
    observation TEXT NOT NULL,
    root_cause TEXT,
    corrective_action TEXT,
    validation_status VARCHAR(50) DEFAULT 'DRAFT',
    validated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
