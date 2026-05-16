#!/usr/bin/env python3
"""Generate Code Tool — Generate code using an LLM.

Sends a code generation prompt to an LLM and returns the generated code.
For saving the generated code to a file, use the save-into-file tool separately.

Env vars:
  MCP_LLM_BASE_URL   - LLM API base URL (required, e.g. http://localhost:11434/v1)
  MCP_LLM_API_KEY    - API key (optional, for OpenAI-compatible APIs)
  MCP_LLM_MODEL      - Model name (default: "qwen3:14b")

Args (JSON via sys.argv[1]):
  prompt    - What code to generate (required)
              e.g. "generate a quicksort function", "write a REST API handler in Go"
  language  - Programming language (optional, e.g. "python", "go", "javascript")

Returns:
  JSON with the LLM-generated code.
"""

import json
import os
import sys
import urllib.request
import urllib.error

SYSTEM_PROMPT = """You are Zora, an expert software engineer specialized in code generation.
When generating code:
- Write clean, well-structured, production-ready code
- Include necessary imports and package declarations
- Add comments for complex logic
- Always specify the language in code blocks
- Provide complete, runnable code — not fragments
- If the user specifies a language, use that language
- Focus solely on code generation; do not add excessive explanation"""

MAX_TOKENS = 4096


def call_llm(prompt: str, language: str = None) -> str:
    """Call LLM to generate code."""
    base_url = os.environ["MCP_LLM_BASE_URL"].rstrip("/")
    api_key = os.environ.get("MCP_LLM_API_KEY", "")
    model = os.environ.get("MCP_LLM_MODEL", "qwen3:14b")

    lang_hint = f" in {language}" if language else ""

    user_content = f"{prompt}{lang_hint}"

    messages = [
        {"role": "system", "content": SYSTEM_PROMPT},
        {"role": "user", "content": user_content},
    ]

    body = json.dumps({
        "model": model,
        "messages": messages,
        "temperature": 0.2,
        "max_tokens": MAX_TOKENS,
    }).encode()

    req = urllib.request.Request(
        f"{base_url}/chat/completions",
        data=body,
        method="POST",
        headers={"Content-Type": "application/json"},
    )
    if api_key:
        req.add_header("Authorization", f"Bearer {api_key}")

    try:
        with urllib.request.urlopen(req, timeout=120) as resp:
            result = json.loads(resp.read())
            return result["choices"][0]["message"]["content"]
    except urllib.error.HTTPError as e:
        err_body = e.read().decode() if e.fp else ""
        raise RuntimeError(f"LLM HTTP {e.code}: {err_body}")
    except urllib.error.URLError as e:
        raise RuntimeError(f"LLM connection failed: {e.reason}")


def main():
    args = json.loads(sys.argv[1])
    prompt = args["prompt"]
    language = args.get("language")

    try:
        result_text = call_llm(prompt, language)
        print(json.dumps({"code": result_text}))
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)


if __name__ == "__main__":
    main()