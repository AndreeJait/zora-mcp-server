# Zora MCP Tools — Registration Reference

Register each tool via `POST /api/v1/tools`, then upload the script to MinIO bucket `zora-scripts`.

Replace `{{MCP_LLM_BASE_URL}}`, `{{MCP_LLM_API_KEY}}`, `{{MCP_LLM_MODEL}}`, `{{MCP_WAHA_BASE_URL}}`, `{{MCP_WAHA_API_KEY}}`, `{{MCP_WAHA_SESSION}}`, `{{MCP_SEARXNG_BASE_URL}}`, `{{MCP_STORAGE_BASE_URL}}`, and `{{MCP_STORAGE_API_KEY}}` with your actual values.

> **Note:** `waha-reaction`, `waha-typing`, and `waha-send-seen` are **built-in tools** in zora-core — they are injected automatically when the agent processes a WhatsApp message. They do NOT need to be registered as MCP tools. The scripts below are only for standalone MCP usage if needed.

> **Note:** `waha-send-text` is **no longer a built-in tool**. Text delivery to WhatsApp is handled automatically by zora-core. The `waha-send-text.py` MCP script is only for standalone MCP usage.

---

## 1. save-into-file

Save content as a downloadable file. Use this tool when the user asks to save, export, or download generated code or text content.

### Register Tool

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/tools \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "save-into-file",
    "description": "Save content as a downloadable file. Use this tool when the user asks to save, export, or download generated code or text content. Provide the content, desired filename, and optionally a MIME type. Returns a download URL for the saved file.",
    "language": "python",
    "object_key": "save-into-file.py",
    "bucket": "zora-scripts",
    "parameters": {
      "type": "object",
      "required": ["content", "filename"],
      "properties": {
        "content": {"type": "string", "description": "The text content to save to the file"},
        "filename": {"type": "string", "description": "The filename for the saved file (e.g. 'quicksort.go', 'analysis.md')"},
        "content_type": {"type": "string", "description": "MIME type of the content (optional, auto-detected from filename if omitted)"}
      }
    }
  }'
```

### Call via MCP

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/mcp/tools/call \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "save-into-file",
    "arguments": {
      "content": "package main\n\nfunc main() {\n    println(\"hello\")\n}",
      "filename": "main.go"
    }
  }'
```

---

## 2. generate-code

Generate code based on a prompt. Use this tool when the user asks to write, create, or generate code. Returns the generated code. To save the generated code as a file, use the save-into-file tool afterwards.

### Register Tool

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/tools \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "generate-code",
    "description": "Generate code based on a prompt. Use this tool when the user asks to write, create, or generate code. Returns the generated code. To save the generated code as a file, use the save-into-file tool afterwards.",
    "language": "python",
    "object_key": "generate-code.py",
    "bucket": "zora-scripts",
    "parameters": {
      "type": "object",
      "required": ["prompt"],
      "properties": {
        "prompt": {"type": "string", "description": "Description of the code to generate (e.g. 'a quicksort function', 'a REST API handler in Go')"},
        "language": {"type": "string", "description": "Programming language for the generated code (e.g. 'go', 'python', 'javascript')"}
      }
    },
    "env": {
      "MCP_LLM_BASE_URL": "{{MCP_LLM_BASE_URL}}",
      "MCP_LLM_API_KEY": "{{MCP_LLM_API_KEY}}",
      "MCP_LLM_MODEL": "{{MCP_LLM_MODEL}}"
    }
  }'
```

### Call via MCP

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/mcp/tools/call \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "generate-code",
    "arguments": {
      "prompt": "generate a quicksort algorithm",
      "language": "go"
    }
  }'
```

---

## 3. analyze-code

Analyze, review, debug, or explain code. Use this tool when the user asks to understand, review, find bugs in, or get explanations about code. Does not save files — use save-into-file for that.

### Register Tool

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/tools \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "analyze-code",
    "description": "Analyze, review, debug, or explain code. Use this tool when the user asks to understand, review, find bugs in, or get explanations about code. Does not save files — use save-into-file for that.",
    "language": "python",
    "object_key": "analyze-code.py",
    "bucket": "zora-scripts",
    "parameters": {
      "type": "object",
      "required": ["prompt"],
      "properties": {
        "prompt": {"type": "string", "description": "What to analyze (e.g. 'find bugs in this code', 'explain what this function does')"},
        "code": {"type": "string", "description": "The source code to analyze (optional — code can also be pasted in the prompt)"},
        "language": {"type": "string", "description": "Programming language (e.g. 'python', 'go', 'javascript')"}
      }
    },
    "env": {
      "MCP_LLM_BASE_URL": "{{MCP_LLM_BASE_URL}}",
      "MCP_LLM_API_KEY": "{{MCP_LLM_API_KEY}}",
      "MCP_LLM_MODEL": "{{MCP_LLM_MODEL}}"
    }
  }'
```

### Call via MCP

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/mcp/tools/call \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "analyze-code",
    "arguments": {
      "prompt": "find bugs",
      "code": "def add(a, b):\n    return a - b",
      "language": "python"
    }
  }'
```

---

## 4. waha-reaction (Built-in in zora-core)

React to a WhatsApp message with an emoji. Use 💭 when thinking, empty string to remove.

> This is a built-in tool in zora-core. Register as MCP tool only for standalone usage.

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/tools \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "waha-reaction",
    "description": "React to a WhatsApp message with an emoji. Use empty emoji to remove a reaction. Use 💭 when the agent is thinking and remove it when done.",
    "language": "python",
    "object_key": "waha-reaction.py",
    "bucket": "zora-scripts",
    "parameters": {
      "type": "object",
      "required": ["chatId", "messageId", "emoji"],
      "properties": {
        "chatId": {"type": "string", "description": "WhatsApp chat ID (e.g. 62812xxx@c.us)"},
        "messageId": {"type": "string", "description": "WhatsApp message ID to react to"},
        "emoji": {"type": "string", "description": "Emoji reaction (empty string to remove)"}
      }
    },
    "env": {
      "MCP_WAHA_BASE_URL": "{{MCP_WAHA_BASE_URL}}",
      "MCP_WAHA_API_KEY": "{{MCP_WAHA_API_KEY}}",
      "MCP_WAHA_SESSION": "{{MCP_WAHA_SESSION}}"
    }
  }'
```

---

## 5. waha-typing (Built-in in zora-core)

Start or stop typing indicator in a WhatsApp chat.

> This is a built-in tool in zora-core. Register as MCP tool only for standalone usage.

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/tools \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "waha-typing",
    "description": "Start or stop typing indicator in a WhatsApp chat.",
    "language": "python",
    "object_key": "waha-typing.py",
    "bucket": "zora-scripts",
    "parameters": {
      "type": "object",
      "required": ["chatId", "action"],
      "properties": {
        "chatId": {"type": "string", "description": "WhatsApp chat ID"},
        "action": {"type": "string", "enum": ["start", "stop"], "description": "'start' or 'stop'"}
      }
    },
    "env": {
      "MCP_WAHA_BASE_URL": "{{MCP_WAHA_BASE_URL}}",
      "MCP_WAHA_API_KEY": "{{MCP_WAHA_API_KEY}}",
      "MCP_WAHA_SESSION": "{{MCP_WAHA_SESSION}}"
    }
  }'
```

---

## 6. waha-send-seen (Built-in in zora-core)

Mark messages as read in a WhatsApp chat.

> This is a built-in tool in zora-core. Register as MCP tool only for standalone usage.

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/tools \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "waha-send-seen",
    "description": "Mark messages as read in a WhatsApp chat.",
    "language": "python",
    "object_key": "waha-send-seen.py",
    "bucket": "zora-scripts",
    "parameters": {
      "type": "object",
      "required": ["chatId"],
      "properties": {
        "chatId": {"type": "string", "description": "WhatsApp chat ID"}
      }
    },
    "env": {
      "MCP_WAHA_BASE_URL": "{{MCP_WAHA_BASE_URL}}",
      "MCP_WAHA_API_KEY": "{{MCP_WAHA_API_KEY}}",
      "MCP_WAHA_SESSION": "{{MCP_WAHA_SESSION}}"
    }
  }'
```

---

## 7. waha-chit-chat

Handle casual conversations using an LLM. Maintains in-memory session-based chat history.

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/tools \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "waha-chit-chat",
    "description": "Handle casual conversations using an LLM. Processes incoming messages and generates natural responses.",
    "language": "python",
    "object_key": "waha-chit-chat.py",
    "bucket": "zora-scripts",
    "parameters": {
      "type": "object",
      "required": ["message"],
      "properties": {
        "message": {"type": "string", "description": "The incoming message text"},
        "sender": {"type": "string", "description": "Sender name for context"},
        "sessionId": {"type": "string", "description": "Conversation session ID for chat history"}
      }
    },
    "env": {
      "MCP_LLM_BASE_URL": "{{MCP_LLM_BASE_URL}}",
      "MCP_LLM_API_KEY": "{{MCP_LLM_API_KEY}}",
      "MCP_LLM_MODEL": "{{MCP_LLM_MODEL}}"
    }
  }'
```

---

## 8. waha-analyze-paper

Analyze documents/papers using an LLM with vision capabilities.

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/tools \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "waha-analyze-paper",
    "description": "Analyze documents and papers using an LLM with vision capabilities. Accepts image URLs or base64-encoded files.",
    "language": "python",
    "object_key": "waha-analyze-paper.py",
    "bucket": "zora-scripts",
    "parameters": {
      "type": "object",
      "required": ["prompt"],
      "properties": {
        "prompt": {"type": "string", "description": "Analysis prompt (e.g. 'Summarize this paper', 'Extract key findings')"},
        "fileUrl": {"type": "string", "description": "URL to the document/image"},
        "fileBase64": {"type": "string", "description": "Base64-encoded file content"},
        "fileType": {"type": "string", "description": "MIME type of the file (required with fileBase64)"}
      }
    },
    "env": {
      "MCP_LLM_BASE_URL": "{{MCP_LLM_BASE_URL}}",
      "MCP_LLM_API_KEY": "{{MCP_LLM_API_KEY}}",
      "MCP_LLM_MODEL": "{{MCP_LLM_MODEL}}"
    }
  }'
```

---

## 9. waha-analyze-code (Legacy)

> **Note:** This tool is deprecated. Use `analyze-code` for code analysis and `generate-code` for code generation. Use `save-into-file` for saving files to MinIO.

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/tools \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "waha-analyze-code",
    "description": "Analyze, generate, and save code using an LLM. DEPRECATED: use analyze-code, generate-code, and save-into-file instead.",
    "language": "python",
    "object_key": "waha-analyze-code.py",
    "bucket": "zora-scripts",
    "parameters": {
      "type": "object",
      "required": ["prompt"],
      "properties": {
        "prompt": {"type": "string", "description": "What to do"},
        "code": {"type": "string", "description": "Source code to analyze"},
        "language": {"type": "string", "description": "Programming language"},
        "save": {"type": "boolean", "description": "If true, save generated code to MinIO"},
        "filename": {"type": "string", "description": "Filename for saved code"}
      }
    },
    "env": {
      "MCP_LLM_BASE_URL": "{{MCP_LLM_BASE_URL}}",
      "MCP_LLM_API_KEY": "{{MCP_LLM_API_KEY}}",
      "MCP_LLM_MODEL": "{{MCP_LLM_MODEL}}"
    }
  }'
```

---

## 10. internet-search

Search the internet using a SearXNG instance. Returns formatted results with titles, URLs, content snippets, and references.

### Register Tool

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/tools \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "internet-search",
    "description": "Search the internet using SearXNG. Returns formatted results with titles, URLs, content snippets, and references. Use this when you need to find current information from the web. Supports categories: general, news, images, videos, science, it, files, social media, scientific publications.",
    "language": "python",
    "object_key": "internet-search.py",
    "bucket": "zora-scripts",
    "parameters": {
      "type": "object",
      "required": ["query"],
      "properties": {
        "query": {"type": "string", "description": "Search query string"},
        "categories": {"type": "string", "description": "Comma-separated categories: general, news, images, videos, science, it, files, social media, scientific publications. Default: general"},
        "language": {"type": "string", "description": "Language code (e.g. en, id). Default: auto"},
        "time_range": {"type": "string", "enum": ["day", "week", "month", "year"], "description": "Time range filter. Default: none"},
        "pageno": {"type": "integer", "description": "Page number (1-based). Default: 1"}
      }
    },
    "env": {
      "MCP_SEARXNG_BASE_URL": "{{MCP_SEARXNG_BASE_URL}}"
    }
  }'
```

### Call via MCP

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/mcp/tools/call \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "internet-search",
    "arguments": {
      "query": "papers about trading using LLM",
      "categories": "scientific publications,general"
    }
  }'
```

---

## 11. file-read

Read the content of a file stored in object storage by its object key. Returns the file content as text.

### Register Tool

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/tools \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "file-read",
    "description": "Read the content of a file stored in object storage by its object_key. Returns the file content as text. Use this to view existing files before editing them.",
    "language": "python",
    "object_key": "file-read.py",
    "bucket": "zora-scripts",
    "parameters": {
      "type": "object",
      "required": ["object_key"],
      "properties": {
        "object_key": {"type": "string", "description": "The object key (path) of the file in storage"},
        "bucket": {"type": "string", "description": "Storage bucket (default: zora-files)"}
      }
    },
    "env": {
      "MCP_STORAGE_BASE_URL": "{{MCP_STORAGE_BASE_URL}}",
      "MCP_STORAGE_API_KEY": "{{MCP_STORAGE_API_KEY}}"
    }
  }'
```

### Call via MCP

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/mcp/tools/call \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "file-read",
    "arguments": {
      "object_key": "uploads/abc-123-report.md",
      "bucket": "zora-files"
    }
  }'
```

---

## 12. file-write

Write (overwrite) content to a file in object storage by its object key. Use file-read first to see current content, then file-write with the full updated content.

### Register Tool

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/tools \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "file-write",
    "description": "Write or overwrite content to a file in object storage by its object_key. Use file-read first to see the current content, then file-write with the full updated content. Returns the object_key and a presigned download URL.",
    "language": "python",
    "object_key": "file-write.py",
    "bucket": "zora-scripts",
    "parameters": {
      "type": "object",
      "required": ["object_key", "content"],
      "properties": {
        "object_key": {"type": "string", "description": "The object key (path) of the file to write"},
        "content": {"type": "string", "description": "The full new content to write to the file"},
        "content_type": {"type": "string", "description": "MIME type of the content (default: text/plain)"}
      }
    },
    "env": {
      "MCP_STORAGE_BASE_URL": "{{MCP_STORAGE_BASE_URL}}",
      "MCP_STORAGE_API_KEY": "{{MCP_STORAGE_API_KEY}}"
    }
  }'
```

### Call via MCP

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/mcp/tools/call \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "file-write",
    "arguments": {
      "object_key": "uploads/abc-123-report.md",
      "content": "# Updated Report\n\nThis is the updated content.",
      "content_type": "text/markdown"
    }
  }'
```

---

## 13. summarize-text

Summarize or process text using an LLM. Use after internet-search to summarize results, or standalone to condense long text.

### Register Tool

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/tools \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "summarize-text",
    "description": "Summarize or process text using an LLM. Takes text content and optional instructions. Useful for summarizing search results, articles, papers, or any text. Preserves key facts and references. Use after internet-search to condense results, then save-into-file to save the output.",
    "language": "python",
    "object_key": "summarize-text.py",
    "bucket": "zora-scripts",
    "parameters": {
      "type": "object",
      "required": ["text"],
      "properties": {
        "text": {"type": "string", "description": "The text content to summarize or process"},
        "prompt": {"type": "string", "description": "Instructions for how to process the text (default: summarize concisely with references)"},
        "format": {"type": "string", "description": "Output format hint (e.g. markdown, bullet-points, paragraph)"}
      }
    },
    "env": {
      "MCP_LLM_BASE_URL": "{{MCP_LLM_BASE_URL}}",
      "MCP_LLM_API_KEY": "{{MCP_LLM_API_KEY}}",
      "MCP_LLM_MODEL": "{{MCP_LLM_MODEL}}"
    }
  }'
```

### Call via MCP

```bash
curl --request POST \
  --url http://localhost:8081/api/v1/mcp/tools/call \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "name": "summarize-text",
    "arguments": {
      "text": "1. Paper About X ... URL: ... 2. Paper About Y ... URL: ...",
      "prompt": "Summarize the key findings from these search results. Include all source URLs as references. Format as markdown with sections: Summary, Key Findings, References.",
      "format": "markdown"
    }
  }'
```

---

## Env Var Management

### Set Env Vars on a Tool

```bash
curl --request PUT \
  --url http://localhost:8081/api/v1/tools/{{TOOL_ID}}/env \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{
    "env": {
      "MCP_LLM_BASE_URL": "http://localhost:11434/v1",
      "MCP_LLM_API_KEY": "",
      "MCP_LLM_MODEL": "qwen3:14b"
    }
  }'
```

---

## Upload Scripts to MinIO

```bash
# Replace with your MinIO endpoint and credentials
MC_HOST=myminio

for script in save-into-file generate-code analyze-code waha-reaction waha-typing waha-send-seen waha-send-text waha-chit-chat waha-analyze-paper waha-analyze-code internet-search file-read file-write summarize-text; do
  mc cp ${script}.py ${MC_HOST}/zora-scripts/${script}.py
done
```

---

## Deactivating Old Tools

To deactivate the legacy `waha-analyze-code` and `waha-send-text` MCP tools (they're superseded by the new tools):

```bash
# Get tool IDs
curl http://localhost:8081/api/v1/tools -H 'X-API-Key: YOUR_API_KEY'

# Deactivate (set is_active: false)
curl --request PUT \
  --url http://localhost:8081/api/v1/tools/{{TOOL_ID}} \
  --header 'Content-Type: application/json' \
  --header 'X-API-Key: YOUR_API_KEY' \
  --data '{"is_active": false}'
```