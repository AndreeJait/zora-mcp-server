#!/usr/bin/env python3
"""Save Into File Tool — Save content as a downloadable file.

Takes text content and saves it to a file that can be downloaded via a presigned URL.
Uses the _files convention so the MCP server handles MinIO upload automatically.

Args (JSON via sys.argv[1]):
  content       - The text content to save (required)
  filename      - The filename for the saved file (required, e.g. "quicksort.go")
  content_type  - MIME type of the content (optional, auto-detected from filename if omitted)

Returns:
  JSON with filename, content_type, size_bytes, and _files array.
  The server replaces _files with saved_files containing presigned download URLs.
"""

import json
import os
import sys


# Extension-to-MIME mapping
EXT_MAP = {
    ".go": "text/x-go",
    ".py": "text/x-python",
    ".js": "text/javascript",
    ".ts": "text/typescript",
    ".java": "text/x-java",
    ".rs": "text/x-rust",
    ".sh": "text/x-shellscript",
    ".sql": "text/x-sql",
    ".html": "text/html",
    ".css": "text/css",
    ".yaml": "text/yaml",
    ".yml": "text/yaml",
    ".json": "application/json",
    ".xml": "text/xml",
    ".md": "text/markdown",
    ".txt": "text/plain",
    ".rb": "text/x-ruby",
    ".cpp": "text/x-c++src",
    ".c": "text/x-csrc",
    ".php": "text/x-php",
    ".swift": "text/x-swift",
    ".kt": "text/x-kotlin",
    ".dart": "text/x-dart",
}


def guess_content_type(filename: str) -> str:
    """Guess MIME content type from filename extension."""
    _, ext = os.path.splitext(filename)
    return EXT_MAP.get(ext.lower(), "text/plain")


def main():
    args = json.loads(sys.argv[1])
    content = args.get("content", "")
    filename = args.get("filename", "output.txt")
    content_type = args.get("content_type") or guess_content_type(filename)

    if not content:
        print(json.dumps({"error": "content is required and cannot be empty"}))
        sys.exit(1)

    if not filename:
        print(json.dumps({"error": "filename is required"}))
        sys.exit(1)

    # Use _files convention — the MCP server will upload to MinIO
    # and replace with saved_files containing presigned URLs
    response = {
        "filename": filename,
        "content_type": content_type,
        "size_bytes": len(content.encode("utf-8")),
        "_files": [
            {
                "filename": filename,
                "content": content,
                "content_type": content_type,
            }
        ],
    }

    print(json.dumps(response))


if __name__ == "__main__":
    main()