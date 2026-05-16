-- Reverse the tags normalization: restore the text[] column and drop the new tables.

-- 1. Re-add the tags column
ALTER TABLE zora_tools_registry ADD COLUMN tags TEXT[] NOT NULL DEFAULT '{}';

-- 2. Re-populate tags from the join table
UPDATE zora_tools_registry t
SET tags = (
    SELECT ARRAY_AGG(tg.name)
    FROM zora_tool_tags tt
    JOIN zora_tags tg ON tg.id = tt.tag_id
    WHERE tt.tool_id = t.id
);

-- 3. Re-create the GIN index
CREATE INDEX idx_tools_tags ON zora_tools_registry USING GIN (tags);

-- 4. Drop the join table and tags table
DROP TABLE IF EXISTS zora_tool_tags;
DROP TABLE IF EXISTS zora_tags;