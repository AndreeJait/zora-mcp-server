-- Drop old index and recreate for new dimension
DROP INDEX IF EXISTS idx_tools_embedding_cosine;
ALTER TABLE zora_tools_registry ALTER COLUMN embedding TYPE VECTOR(768);
CREATE INDEX idx_tools_embedding_cosine ON zora_tools_registry
    USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);