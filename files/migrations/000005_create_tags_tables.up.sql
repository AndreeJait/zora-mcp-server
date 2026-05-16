-- Normalize tags from text[] column into a dedicated zora_tags table with descriptions
-- and a many-to-many join table zora_tool_tags.

-- 1. Create zora_tags table
CREATE TABLE IF NOT EXISTS zora_tags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_tags_name ON zora_tags (name);

-- 2. Create zora_tool_tags join table
CREATE TABLE IF NOT EXISTS zora_tool_tags (
    tool_id UUID NOT NULL REFERENCES zora_tools_registry(id) ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES zora_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (tool_id, tag_id)
);

-- 3. Migrate existing text[] tags into zora_tags (empty descriptions for existing tags)
INSERT INTO zora_tags (name, description)
SELECT DISTINCT tag_name, ''
FROM (
    SELECT unnest(tags) AS tag_name
    FROM zora_tools_registry
    WHERE tags <> '{}'
) sub
ON CONFLICT (name) DO NOTHING;

-- 4. Migrate tool-tag associations into zora_tool_tags
INSERT INTO zora_tool_tags (tool_id, tag_id)
SELECT t.id, tg.id
FROM zora_tools_registry t
CROSS JOIN LATERAL unnest(t.tags) AS tag_name
JOIN zora_tags tg ON tg.name = tag_name
WHERE t.tags <> '{}';

-- 5. Drop the old tags column and its GIN index
DROP INDEX IF EXISTS idx_tools_tags;
ALTER TABLE zora_tools_registry DROP COLUMN IF EXISTS tags;