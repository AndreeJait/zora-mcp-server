-- Add tags column for hybrid retrieval (tag-based filtering alongside vector similarity).
-- Uses PostgreSQL text[] array type for efficient overlap queries with GIN index.
ALTER TABLE zora_tools_registry ADD COLUMN tags TEXT[] NOT NULL DEFAULT '{}';

-- Add GIN index for efficient tag overlap queries (tags && :tags).
CREATE INDEX idx_tools_tags ON zora_tools_registry USING GIN (tags);