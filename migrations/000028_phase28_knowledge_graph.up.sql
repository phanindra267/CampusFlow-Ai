
CREATE TABLE IF NOT EXISTS kg_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(100) NOT NULL, -- PERSON, ORGANIZATION, COURSE, SKILL, PUBLICATION
    canonical_name VARCHAR(255) NOT NULL,
    aliases JSONB,
    source_id UUID,
    confidence FLOAT,
    version VARCHAR(50) DEFAULT '1.0',
    provenance JSONB,
    status VARCHAR(50) DEFAULT 'ACTIVE', -- ACTIVE, MERGED, DEPRECATED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS kg_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_entity_id UUID REFERENCES kg_entities(id),
    target_entity_id UUID REFERENCES kg_entities(id),
    relation_type VARCHAR(100) NOT NULL, -- CITES, DEPENDS_ON, TEACHES, RESEARCHES
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    confidence FLOAT,
    provenance JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS kg_claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_id UUID REFERENCES kg_entities(id),
    claim_text TEXT NOT NULL,
    claim_type VARCHAR(100), -- FACT, INFERENCE, EXTRACTION
    evidence JSONB NOT NULL,
    confidence FLOAT,
    is_machine_extracted BOOLEAN DEFAULT FALSE,
    verification_status VARCHAR(50) DEFAULT 'CANDIDATE', -- CANDIDATE, SUPPORTED, DISPUTED, REJECTED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS kg_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_uri VARCHAR(1000) NOT NULL,
    publisher VARCHAR(255),
    license_info TEXT,
    trust_score FLOAT,
    ingestion_contract JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS kg_hypotheses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id VARCHAR(255) NOT NULL,
    hypothesis_text TEXT NOT NULL,
    supporting_evidence JSONB,
    counter_evidence JSONB,
    is_ai_generated BOOLEAN DEFAULT TRUE,
    lifecycle_status VARCHAR(50) DEFAULT 'PROPOSED', -- PROPOSED, INVESTIGATING, SUPPORTED, REFUTED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
