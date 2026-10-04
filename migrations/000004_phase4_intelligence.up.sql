
-- Requires superuser to install pgvector extension in a real setup:
-- CREATE EXTENSION IF NOT EXISTS vector;
-- We'll simulate the vector column with an array/jsonb for fallback if vector extension isn't available on the system.
-- But using standard pgvector syntax here for reference, wrapped safely.

CREATE TABLE IF NOT EXISTS embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL, -- EVENT, OPPORTUNITY, CLUB
    entity_id UUID NOT NULL,
    embedding JSONB, -- fallback structure for vector to ensure DB runs without extension errors during scaffolding
    model VARCHAR(100) NOT NULL,
    version VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(entity_type, entity_id)
);

CREATE TABLE IF NOT EXISTS user_interests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    interest VARCHAR(100) NOT NULL,
    strength FLOAT NOT NULL DEFAULT 1.0,
    source VARCHAR(50) NOT NULL, -- EXPLICIT, BEHAVIORAL
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, interest)
);

CREATE TABLE IF NOT EXISTS search_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    query TEXT NOT NULL,
    search_mode VARCHAR(50) NOT NULL, -- KEYWORD, SEMANTIC, HYBRID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS recommendation_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    feedback_type VARCHAR(50) NOT NULL, -- CLICKED, SAVED, DISMISSED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
