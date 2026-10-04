
CREATE TABLE IF NOT EXISTS scientific_claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    statement TEXT NOT NULL,
    source_reference TEXT,
    confidence FLOAT,
    status VARCHAR(50) DEFAULT 'PROPOSED',
    context JSONB,
    evidence JSONB,
    contradictions JSONB,
    version INT NOT NULL DEFAULT 1,
    author_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS curriculum_graph_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    node_type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS curriculum_graph_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    source_id UUID NOT NULL REFERENCES curriculum_graph_nodes(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES curriculum_graph_nodes(id) ON DELETE CASCADE,
    relationship_type VARCHAR(100) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS research_funding (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES research_projects(id) ON DELETE CASCADE,
    grant_name VARCHAR(255) NOT NULL,
    sponsor VARCHAR(255),
    total_budget NUMERIC(18, 2),
    spent_budget NUMERIC(18, 2) DEFAULT 0,
    currency VARCHAR(10) DEFAULT 'USD',
    restrictions JSONB,
    start_date DATE,
    end_date DATE,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS replication_studies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_experiment_id UUID NOT NULL REFERENCES research_experiments(id) ON DELETE CASCADE,
    replicating_researcher_id UUID REFERENCES users(id) ON DELETE SET NULL,
    protocol JSONB,
    environment_snapshot JSONB,
    status VARCHAR(50) DEFAULT 'PLANNED',
    results JSONB,
    comparison JSONB,
    differences TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
