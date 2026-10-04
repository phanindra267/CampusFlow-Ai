
CREATE TABLE IF NOT EXISTS agent_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    purpose TEXT,
    owner_id VARCHAR(255) NOT NULL,
    capabilities JSONB NOT NULL,
    permission_set JSONB,
    risk_classification VARCHAR(50) DEFAULT 'LOW', -- LOW, MEDIUM, HIGH, CRITICAL
    max_autonomy_level INT DEFAULT 1, -- 0: Suggest, 1: Plan, 2: Execute w/ Approval, 3: Autonomous
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tool_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tool_name VARCHAR(255) NOT NULL UNIQUE,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    input_schema JSONB NOT NULL,
    output_schema JSONB NOT NULL,
    permissions JSONB,
    risk_level VARCHAR(50) DEFAULT 'LOW',
    resource_limits JSONB,
    requires_approval BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS autonomous_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id VARCHAR(255) NOT NULL,
    objective TEXT NOT NULL,
    priority VARCHAR(50) DEFAULT 'NORMAL',
    risk_level VARCHAR(50) DEFAULT 'LOW',
    budget JSONB,
    state VARCHAR(50) DEFAULT 'CREATED', -- CREATED, PLANNING, AWAITING_APPROVAL, RUNNING, VERIFYING, COMPLETED, FAILED, CANCELLED
    assigned_agents JSONB,
    execution_plan JSONB,
    execution_history JSONB,
    trace_id VARCHAR(255),
    deadline TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS task_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID REFERENCES autonomous_tasks(id) ON DELETE CASCADE,
    approver_id VARCHAR(255),
    scope JSONB NOT NULL,
    decision VARCHAR(50) DEFAULT 'PENDING', -- PENDING, APPROVED, REJECTED, EXPIRED
    reason TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS agent_memory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID REFERENCES agent_registry(id) ON DELETE CASCADE,
    task_id UUID REFERENCES autonomous_tasks(id),
    memory_type VARCHAR(50) NOT NULL, -- TASK, SESSION, PREFERENCE
    content JSONB NOT NULL,
    confidence FLOAT,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
