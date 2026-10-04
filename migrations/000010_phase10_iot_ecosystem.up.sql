
CREATE TABLE IF NOT EXISTS iot_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    campus_id UUID REFERENCES campuses(id) ON DELETE CASCADE,
    device_type VARCHAR(100) NOT NULL, -- SENSOR, HVAC, ACCESS_CONTROL, GATEWAY
    location JSONB,
    capabilities JSONB,
    status VARCHAR(50) DEFAULT 'ACTIVE', -- ACTIVE, MAINTENANCE, SUSPENDED, DECOMMISSIONED
    last_seen_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    security_state VARCHAR(50) DEFAULT 'SECURE',
    firmware_version VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS digital_twin_state (
    entity_id UUID PRIMARY KEY, -- Maps to a resource, room, or building
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    current_state JSONB NOT NULL,
    confidence FLOAT DEFAULT 1.0,
    data_sources JSONB,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS iot_commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL REFERENCES iot_devices(id) ON DELETE CASCADE,
    requested_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    command_payload JSONB NOT NULL,
    safety_clearance BOOLEAN DEFAULT FALSE,
    status VARCHAR(50) DEFAULT 'PENDING_APPROVAL', -- PENDING_APPROVAL, APPROVED, SENT, VERIFIED, FAILED
    timeout_seconds INT DEFAULT 30,
    executed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS research_publications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    abstract TEXT,
    authors JSONB,
    doi VARCHAR(100),
    published_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
