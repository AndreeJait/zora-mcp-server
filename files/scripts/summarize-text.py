#!/usr/bin/env python3
"""Summarize Text Tool — Summarize or process text using an LLM.

Takes text content and a prompt, sends them to an LLM, and returns the result.
Useful for summarizing search results, articles, papers, or any text content.

Env vars:
  MCP_LLM_BASE_URL   - LLM API base URL (required, e.g. http://localhost:11434/v1)
  MCP_LLM_API_KEY    - API key (optional, for OpenAI-compatible APIs)
  MCP_LLM_MODEL      - Model name (default: "qwen3:14b")

Args (JSON via sys.argv[1]):
  text      - The text content to summarize (required)
  prompt    - Instructions for how to process the text (optional, default: "Summarize the following text concisely.")
              e.g. "Summarize the key findings", "Extract main points with references",
              "Create a structured report with sections: summary, key findings, references"
  format    - Output format hint (optional, e.g. "markdown", "bullet-points", "paragraph")

Returns:
  JSON with the LLM-generated summary/result.
"""

import json
import os
import sys
import urllib.request
import urllib.error

SYSTEM_PROMPT = """You are Zora, an expert research assistant specialized in text summarization and analysis.
When processing text:
- Follow the user's prompt instructions precisely
- Preserve key facts, numbers, and references from the source text
- When summarizing research or papers, include source URLs as references
- Structure output clearly with headings and bullet points when appropriate
- Be concise but comprehensive — do not omit important details
- When references/URLs are present in the source text, always include them in your output
- If the user asks for a specific format, follow it exactly"""

MAX_TOKENS = 8192
MAX_INPUT_CHARS = 50000  # ~12k tokens input limit


def call_llm(text: str, prompt: str, fmt: str = None) -> str:
    """Call LLM to summarize or process text."""
    base_url = os.environ["MCP_LLM_BASE_URL"].rstrip("/")
    api_key = os.environ.get("MCP_LLM_API_KEY", "")
    model = os.environ.get("MCP_LLM_MODEL", "qwen3:14b")

    # Truncate if input is too long
    if len(text) > MAX_INPUT_CHARS:
        text = text[:MAX_INPUT_CHARS] + "\n\n[... content truncated due to length ...]"

    fmt_hint = f"\n\nOutput format: {fmt}" if fmt else ""

    user_content = f"""{prompt}{fmt_hint}

--- TEXT TO PROCESS ---
{text}
--- END TEXT ---"""

    messages = [
        {"role": "system", "content": SYSTEM_PROMPT},
        {"role": "user", "content": user_content},
    ]

    body = json.dumps({
        "model": model,
        "messages": messages,
        "temperature": 0.3,
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
    text = args.get("text", "")
    if not text:
        print(json.dumps({"error": "text is required"}))
        sys.exit(1)

    prompt = args.get("prompt", "Summarize the following text concisely, preserving key facts and references.")
    fmt = args.get("format")

    try:
        result_text = call_llm(text, prompt, fmt)
        print(json.dumps({"summary": result_text}))
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)


if __name__ == "__main__":
    main()