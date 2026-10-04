
CREATE TABLE IF NOT EXISTS network_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id VARCHAR(255) NOT NULL,
    organization_type VARCHAR(100) NOT NULL,
    region VARCHAR(255),
    trust_status VARCHAR(50) DEFAULT 'PENDING_VERIFICATION', -- PENDING, VERIFIED, REVOKED
    public_key TEXT,
    capabilities JSONB,
    health_status VARCHAR(50) DEFAULT 'UNKNOWN',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS network_data_contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_node_id UUID REFERENCES network_nodes(id) ON DELETE CASCADE,
    product_name VARCHAR(255) NOT NULL,
    schema_definition JSONB NOT NULL,
    access_requirements TEXT,
    privacy_policy TEXT,
    version VARCHAR(50) DEFAULT '1.0',
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS network_skill_equivalences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_skill_id VARCHAR(255) NOT NULL,
    target_skill_id VARCHAR(255) NOT NULL,
    source_node_id UUID REFERENCES network_nodes(id),
    target_node_id UUID REFERENCES network_nodes(id),
    equivalence_confidence FLOAT,
    evidence JSONB,
    status VARCHAR(50) DEFAULT 'PROPOSED', -- PROPOSED, ACCEPTED, REJECTED, DISPUTED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS network_learning_wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL UNIQUE,
    portable_records JSONB,
    selective_disclosure_rules JSONB,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS network_consents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID REFERENCES network_learning_wallets(id) ON DELETE CASCADE,
    target_node_id UUID REFERENCES network_nodes(id),
    scope JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE', -- ACTIVE, WITHDRAWN, EXPIRED
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
