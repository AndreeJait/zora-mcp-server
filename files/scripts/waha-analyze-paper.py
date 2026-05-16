#!/usr/bin/env python3
"""WAHA Analyze Paper Tool — Analyze documents/papers using an LLM with vision.

Sends a document (image, PDF page, etc.) along with a prompt to an LLM
and returns structured analysis.

Env vars:
  MCP_LLM_BASE_URL   - LLM API base URL (required, e.g. http://localhost:11434/v1)
  MCP_LLM_API_KEY    - API key (optional, for OpenAI-compatible APIs)
  MCP_LLM_MODEL      - Model name (default: "llama3.2-vision")

Args (JSON via sys.argv[1]):
  prompt       - Analysis prompt (required)
  fileUrl      - URL to the document/image (mutually exclusive with fileBase64)
  fileBase64   - Base64-encoded file content (mutually exclusive with fileUrl)
  fileType     - MIME type of the file (required with fileBase64, e.g. "image/png")

Returns:
  JSON with the LLM analysis result.
"""

import base64
import json
import os
import sys
import urllib.request
import urllib.error

SYSTEM_PROMPT = """You are Zora, an analytical assistant specialized in analyzing documents and papers.
Provide clear, structured, and thorough analysis.
When analyzing academic papers, cover: key findings, methodology, conclusions, and potential limitations.
When analyzing images or figures, describe what you see and extract relevant information.
Be concise but comprehensive."""

MAX_TOKENS = 4096


def call_llm_vision(prompt: str, file_url: str = None, file_base64: str = None,
                    file_type: str = None) -> str:
    """Call LLM with vision capabilities using OpenAI-compatible content array."""
    base_url = os.environ["MCP_LLM_BASE_URL"].rstrip("/")
    api_key = os.environ.get("MCP_LLM_API_KEY", "")
    model = os.environ.get("MCP_LLM_MODEL", "llama3.2-vision")

    content = [{"type": "text", "text": prompt}]

    if file_url:
        content.append({
            "type": "image_url",
            "image_url": {"url": file_url},
        })
    elif file_base64:
        mime_type = file_type or "image/png"
        data_uri = f"data:{mime_type};base64,{file_base64}"
        content.append({
            "type": "image_url",
            "image_url": {"url": data_uri},
        })

    messages = [
        {"role": "system", "content": SYSTEM_PROMPT},
        {"role": "user", "content": content},
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
    prompt = args["prompt"]
    file_url = args.get("fileUrl")
    file_base64 = args.get("fileBase64")
    file_type = args.get("fileType")

    if not file_url and not file_base64:
        print(json.dumps({"error": "either fileUrl or fileBase64 is required"}))
        sys.exit(1)

    if file_base64 and not file_type:
        print(json.dumps({"error": "fileType is required when using fileBase64"}))
        sys.exit(1)

    if file_url and file_base64:
        print(json.dumps({"error": "fileUrl and fileBase64 are mutually exclusive"}))
        sys.exit(1)

    try:
        result_text = call_llm_vision(
            prompt=prompt,
            file_url=file_url,
            file_base64=file_base64,
            file_type=file_type,
        )

        print(json.dumps({"analysis": result_text}))
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)


if __name__ == "__main__":
    main()