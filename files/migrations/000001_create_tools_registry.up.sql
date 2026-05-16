CREATE EXTENSION IF NOT EXISTS vector;

-- Tools registry with pgvector embedding for semantic search
CREATE TABLE IF NOT EXISTS zora_tools_registry (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    description     TEXT NOT NULL,
    language        VARCHAR(50) NOT NULL CHECK (language IN ('python', 'go', 'bash')),
    object_key      VARCHAR(500) NOT NULL,
    bucket          VARCHAR(255) NOT NULL DEFAULT 'zora-scripts',
    parameters      JSONB NOT NULL DEFAULT '{}',
    embedding       VECTOR(768),
    version         INT NOT NULL DEFAULT 1,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_by      VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata        JSONB NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX uq_tools_name ON zora_tools_registry (name);
CREATE INDEX idx_tools_active ON zora_tools_registry (is_active) WHERE is_active = true;
CREATE INDEX idx_tools_embedding_cosine ON zora_tools_registry
    USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_tools_language ON zora_tools_registry (language);

-- Tool execution audit log
CREATE TABLE IF NOT EXISTS zora_tool_executions (
    id              BIGSERIAL PRIMARY KEY,
    tool_id         UUID NOT NULL REFERENCES zora_tools_registry(id) ON DELETE CASCADE,
    trace_id        VARCHAR(255) NOT NULL,
    parameters      JSONB,
    result          JSONB,
    status          VARCHAR(50) NOT NULL CHECK (status IN ('success', 'failed', 'timeout')),
    error_msg       TEXT,
    duration_ms     INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_executions_tool_id ON zora_tool_executions (tool_id);
CREATE INDEX idx_executions_trace_id ON zora_tool_executions (trace_id);
CREATE INDEX idx_executions_created ON zora_tool_executions (created_at DESC);
