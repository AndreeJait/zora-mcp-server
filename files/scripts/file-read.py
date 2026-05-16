#!/usr/bin/env python3
"""File Read Tool — Read a file from object storage by its object key.

Fetches a presigned URL from the storage API, then downloads and returns the file content.

Env vars:
  MCP_STORAGE_BASE_URL - Zora MCP server base URL (required, e.g. http://localhost:8081)
  MCP_STORAGE_API_KEY   - API key for the storage API (required)

Args (JSON via sys.argv[1]):
  object_key - The object key (path) of the file in storage (required)
  bucket     - Storage bucket (optional, default: "zora-files")

Returns:
  JSON with object_key, bucket, content_type, size, content (text content as string),
  and url (presigned download URL).
"""

import json
import os
import sys
import urllib.request
import urllib.error
import urllib.parse


MAX_FILE_SIZE = 2 * 1024 * 1024  # 2 MB limit for text files


def call_storage_api(base_url: str, api_key: str, object_key: str, bucket: str) -> dict:
    """Call the storage API to get a presigned URL for the file."""
    path = urllib.parse.quote(object_key, safe="/")
    url = f"{base_url.rstrip('/')}/api/v1/storage/files{path}?bucket={urllib.parse.quote(bucket)}"

    req = urllib.request.Request(url, method="GET")
    if api_key:
        req.add_header("X-API-Key", api_key)
    req.add_header("Accept", "application/json")

    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            return json.loads(resp.read())
    except urllib.error.HTTPError as e:
        err_body = e.read().decode() if e.fp else ""
        raise RuntimeError(f"Storage API HTTP {e.code}: {err_body}")
    except urllib.error.URLError as e:
        raise RuntimeError(f"Storage API connection failed: {e.reason}")


def download_file(presigned_url: str) -> tuple:
    """Download file content from a presigned URL. Returns (content_bytes, content_type)."""
    req = urllib.request.Request(presigned_url, method="GET")
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            content_type = resp.headers.get("Content-Type", "application/octet-stream")
            data = resp.read(MAX_FILE_SIZE + 1)
            if len(data) > MAX_FILE_SIZE:
                raise RuntimeError(f"File too large (exceeds {MAX_FILE_SIZE // (1024*1024)}MB limit for text reading)")
            return data, content_type
    except urllib.error.HTTPError as e:
        err_body = e.read().decode() if e.fp else ""
        raise RuntimeError(f"Download HTTP {e.code}: {err_body}")
    except urllib.error.URLError as e:
        raise RuntimeError(f"Download failed: {e.reason}")


def extract_data(api_response: dict) -> dict:
    """Extract the data field from the API response envelope."""
    if "data" in api_response:
        return api_response["data"]
    return api_response


def main():
    args = json.loads(sys.argv[1])
    object_key = args.get("object_key", "")
    if not object_key:
        print(json.dumps({"error": "object_key is required"}))
        sys.exit(1)

    base_url = os.environ.get("MCP_STORAGE_BASE_URL", "")
    api_key = os.environ.get("MCP_STORAGE_API_KEY", "")
    if not base_url:
        print(json.dumps({"error": "MCP_STORAGE_BASE_URL env var is not set"}))
        sys.exit(1)

    bucket = args.get("bucket", "zora-files")

    # Ensure object_key starts with /
    if not object_key.startswith("/"):
        object_key = "/" + object_key

    try:
        # Step 1: Get presigned URL from storage API
        api_resp = call_storage_api(base_url, api_key, object_key, bucket)
        file_info = extract_data(api_resp)
        presigned_url = file_info.get("url", "")

        if not presigned_url:
            print(json.dumps({"error": "No presigned URL returned", "object_key": object_key}))
            sys.exit(1)

        # Step 2: Download the file content
        content_bytes, content_type = download_file(presigned_url)

        # Try to decode as text
        try:
            content = content_bytes.decode("utf-8")
        except UnicodeDecodeError:
            content = f"[Binary file, {len(content_bytes)} bytes, type: {content_type}]"

        print(json.dumps({
            "object_key": object_key.lstrip("/"),
            "bucket": bucket,
            "content_type": content_type,
            "size": len(content_bytes),
            "content": content,
            "url": presigned_url,
        }))

    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)


if __name__ == "__main__":
    main()