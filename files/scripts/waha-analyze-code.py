#!/usr/bin/env python3
"""WAHA Analyze Code Tool — Analyze, generate, and save code using an LLM.

Sends code or a prompt to an LLM and returns structured analysis,
recommendations, or generated code. Optionally saves generated files to MinIO.

Env vars:
  MCP_LLM_BASE_URL   - LLM API base URL (required, e.g. http://localhost:11434/v1)
  MCP_LLM_API_KEY    - API key (optional, for OpenAI-compatible APIs)
  MCP_LLM_MODEL      - Model name (default: "qwen3:14b")

Args (JSON via sys.argv[1]):
  prompt     - What to do (required)
               e.g. "analyze this code", "generate a quicksort in go", "fix bug"
               Can include pasted code directly if code arg is not provided.
  code       - The source code to analyze (optional)
  language   - Programming language (optional, e.g. "python", "go")
  save       - Whether to save generated code to file (optional, default false)
  filename   - Filename for saved code (optional, e.g. "quicksort.go")

Returns:
  JSON with the LLM-generated analysis or solution.
  When save=true, includes _files array for the server to upload to MinIO.
"""

import json
import os
import re
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
- Always specify the language in code blocks

When generating code:
- Write clean, well-structured, production-ready code
- Include necessary imports and package declarations
- Add comments for complex logic
- Specify the language in code blocks"""

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


def extract_code_blocks(text: str) -> list:
    """Extract code blocks from LLM response.
    Returns list of (language, code) tuples."""
    pattern = r"```(\w*)\n(.*?)```"
    matches = re.findall(pattern, text, re.DOTALL)
    return [(lang, code) for lang, code in matches]


def guess_content_type(language: str, filename: str) -> str:
    """Guess MIME content type from language or filename."""
    ext_map = {
        "go": "text/x-go", "python": "text/x-python", "py": "text/x-python",
        "javascript": "text/javascript", "js": "text/javascript",
        "typescript": "text/typescript", "ts": "text/typescript",
        "java": "text/x-java", "rust": "text/x-rust",
        "bash": "text/x-shellscript", "sh": "text/x-shellscript",
        "sql": "text/x-sql", "html": "text/html",
        "css": "text/css", "yaml": "text/yaml", "json": "application/json",
    }
    if filename:
        ext = filename.rsplit(".", 1)[-1].lower() if "." in filename else ""
        if ext in ext_map:
            return ext_map[ext]
    if language:
        return ext_map.get(language.lower(), "text/plain")
    return "text/plain"


def main():
    args = json.loads(sys.argv[1])
    prompt = args["prompt"]
    code = args.get("code")
    language = args.get("language")
    save = args.get("save", False)
    filename = args.get("filename")

    try:
        result_text = call_llm(prompt, code, language)

        response = {"analysis": result_text}

        # When save=true, extract code blocks and include as _files
        if save:
            code_blocks = extract_code_blocks(result_text)
            files = []
            for i, (lang, block_code) in enumerate(code_blocks):
                fname = filename if filename and len(code_blocks) == 1 else None
                if not fname:
                    # Generate filename from language or index
                    ext_map = {
                        "go": ".go", "python": ".py", "javascript": ".js",
                        "typescript": ".ts", "java": ".java", "rust": ".rs",
                        "bash": ".sh", "sql": ".sql", "ruby": ".rb",
                        "cpp": ".cpp", "c": ".c", "html": ".html", "css": ".css",
                    }
                    ext = ext_map.get(lang.lower(), ".txt") if lang else ".txt"
                    fname = f"code_{i+1}{ext}"

                files.append({
                    "filename": fname,
                    "content": block_code,
                    "content_type": guess_content_type(lang, fname),
                })

            if files:
                response["_files"] = files

        print(json.dumps(response))
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)


if __name__ == "__main__":
    main()