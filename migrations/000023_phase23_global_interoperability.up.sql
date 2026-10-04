
CREATE TABLE IF NOT EXISTS edu_institutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    official_name VARCHAR(255) NOT NULL,
    institution_type VARCHAR(100),
    jurisdiction VARCHAR(100),
    accreditation_info JSONB,
    public_keys JSONB,
    api_endpoints JSONB,
    trust_level VARCHAR(50) DEFAULT 'UNVERIFIED',
    capabilities JSONB,
    status VARCHAR(50) DEFAULT 'REGISTERED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS edu_credential_schemas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    schema_definition JSONB NOT NULL,
    issuer_id UUID REFERENCES edu_institutions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS edu_digital_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_id UUID REFERENCES edu_credential_schemas(id) ON DELETE CASCADE,
    issuer_id UUID REFERENCES edu_institutions(id) ON DELETE CASCADE,
    recipient_id VARCHAR(255) NOT NULL, -- User/Student ID
    credential_type VARCHAR(100) NOT NULL,
    claims JSONB NOT NULL,
    evidence JSONB,
    digital_signature TEXT NOT NULL,
    issued_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    revocation_reason TEXT,
    status VARCHAR(50) DEFAULT 'VALID' -- VALID, REVOKED, EXPIRED
);

CREATE TABLE IF NOT EXISTS edu_data_consents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    requester_id VARCHAR(255) NOT NULL, -- Institution, Employer, etc.
    requested_data JSONB NOT NULL,
    purpose TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'GRANTED', -- GRANTED, REVOKED, EXPIRED
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS edu_transfer_credits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id VARCHAR(255) NOT NULL,
    source_institution_id UUID REFERENCES edu_institutions(id),
    target_institution_id UUID REFERENCES edu_institutions(id),
    source_course_info JSONB NOT NULL,
    equivalency_mapping JSONB,
    confidence FLOAT,
    status VARCHAR(50) DEFAULT 'PROPOSED', -- PROPOSED, APPROVED, REJECTED
    reviewed_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS edu_mobility_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id VARCHAR(255) NOT NULL UNIQUE,
    linked_identities JSONB,
    verified_skills JSONB,
    exchange_history JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
