#!/usr/bin/env python3
"""Analyze Code Tool — Analyze, review, and explain code using an LLM.

Sends code or a question to an LLM and returns structured analysis,
bug identification, or explanations. Does NOT save files — use the
save-into-file tool for that.

Env vars:
  MCP_LLM_BASE_URL   - LLM API base URL (required, e.g. http://localhost:11434/v1)
  MCP_LLM_API_KEY    - API key (optional, for OpenAI-compatible APIs)
  MCP_LLM_MODEL      - Model name (default: "qwen3:14b")

Args (JSON via sys.argv[1]):
  prompt    - What to analyze (required)
              e.g. "analyze this code", "what does this function do", "find bugs"
  code      - The source code to analyze (optional)
  language  - Programming language (optional, e.g. "python", "go")

Returns:
  JSON with the LLM-generated analysis.
"""

import json
import os
import sys
import urllib.request
import urllib.error

SYSTEM_PROMPT = """You are Zora, an expert software engineer and code analyst.
When analyzing code:
- Identify bugs, security issues, and anti-patterns
- Suggest improvements with clear explanations
- When fixing problems, provide the corrected code in a code block
- When recommending changes, explain the reasoning
- Keep responses structured and concise
- Always specify the language in code blocks"""

MAX_TOKENS = 4096


def call_llm(prompt: str, code: str = None, language: str = None) -> str:
    """Call LLM to analyze code or answer a coding question."""
    base_url = os.environ["MCP_LLM_BASE_URL"].rstrip("/")
    api_key = os.environ.get("MCP_LLM_API_KEY", "")
    model = os.environ.get("MCP_LLM_MODEL", "qwen3:14b")

    lang_hint = f" ({language})" if language else ""

    if code:
        user_content = f"""{prompt}{lang_hint}

```
{code}
```"""
    else:
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
    code = args.get("code")
    language = args.get("language")

    try:
        result_text = call_llm(prompt, code, language)
        print(json.dumps({"analysis": result_text}))
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)


if __name__ == "__main__":
    main()