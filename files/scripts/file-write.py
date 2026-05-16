#!/usr/bin/env python3
"""File Write Tool — Write (overwrite) a file in object storage by its object key.

Uploads content to a specific object key via the storage API PUT endpoint.
If the object_key already exists, it will be overwritten.

Env vars:
  MCP_STORAGE_BASE_URL - Zora MCP server base URL (required, e.g. http://localhost:8081)
  MCP_STORAGE_API_KEY   - API key for the storage API (required)

Args (JSON via sys.argv[1]):
  object_key   - The object key (path) to write to (required)
  content      - The text content to write (required)
  content_type - MIME type of the content (optional, default: text/plain)

Returns:
  JSON with object_key, bucket, url (presigned download URL).
"""

import json
import os
import sys
import urllib.request
import urllib.error
import urllib.parse


def put_file(base_url: str, api_key: str, bucket: str, object_key: str,
             content: str, content_type: str) -> dict:
    """PUT file content to object storage via the storage API."""
    path = urllib.parse.quote(object_key, safe="/")
    url = f"{base_url.rstrip('/')}/api/v1/storage/files{path}?bucket={urllib.parse.quote(bucket)}&content_type={urllib.parse.quote(content_type)}"

    content_bytes = content.encode("utf-8")
    req = urllib.request.Request(url, data=content_bytes, method="PUT")
    if api_key:
        req.add_header("X-API-Key", api_key)
    req.add_header("Content-Type", "application/octet-stream")

    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return json.loads(resp.read())
    except urllib.error.HTTPError as e:
        err_body = e.read().decode() if e.fp else ""
        raise RuntimeError(f"Storage API HTTP {e.code}: {err_body}")
    except urllib.error.URLError as e:
        raise RuntimeError(f"Storage API connection failed: {e.reason}")


def extract_data(api_response: dict) -> dict:
    """Extract the data field from the API response envelope."""
    if "data" in api_response:
        return api_response["data"]
    return api_response


def main():
    args = json.loads(sys.argv[1])
    object_key = args.get("object_key", "")
    content = args.get("content", "")
    if not object_key:
        print(json.dumps({"error": "object_key is required"}))
        sys.exit(1)
    if not content:
        print(json.dumps({"error": "content is required and cannot be empty"}))
        sys.exit(1)

    base_url = os.environ.get("MCP_STORAGE_BASE_URL", "")
    api_key = os.environ.get("MCP_STORAGE_API_KEY", "")
    if not base_url:
        print(json.dumps({"error": "MCP_STORAGE_BASE_URL env var is not set"}))
        sys.exit(1)

    bucket = args.get("bucket", "zora-files")
    content_type = args.get("content_type", "text/plain")

    # Strip leading / from object_key
    object_key = object_key.lstrip("/")

    # Ensure object_key starts with / for the API
    api_object_key = object_key if object_key.startswith("/") else "/" + object_key

    try:
        api_resp = put_file(base_url, api_key, bucket, api_object_key, content, content_type)
        result = extract_data(api_resp)

        print(json.dumps({
            "object_key": result.get("object_key", object_key),
            "bucket": result.get("bucket", bucket),
            "url": result.get("url", ""),
        }))

    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)


if __name__ == "__main__":
    main()