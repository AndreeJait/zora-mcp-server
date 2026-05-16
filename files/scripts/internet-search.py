#!/usr/bin/env python3
"""Internet Search Tool — Search the web using SearXNG.

Sends a search query to a SearXNG instance and returns formatted results
with titles, URLs, content snippets, and references.

Env vars:
  MCP_SEARXNG_BASE_URL - SearXNG instance URL (required, e.g. https://searxng.example.com)

Args (JSON via sys.argv[1]):
  query      - Search query string (required)
  categories - Comma-separated categories: general, news, images, videos, science,
               it, files, social media, scientific publications, etc. (optional, default: general)
  language   - Language code, e.g. "en", "id" (optional)
  time_range - Time filter: "day", "week", "month", "year" (optional)
  pageno     - Page number, 1-based (optional, default: 1)

Returns:
  JSON with query, result_count, results array, answers, and suggestions.
"""

import json
import os
import sys
import urllib.request
import urllib.error
import urllib.parse


def search_searxng(base_url: str, query: str, categories: str = None,
                   language: str = None, time_range: str = None, pageno: int = 1) -> dict:
    """Execute a search query against SearXNG."""
    params = {
        "q": query,
        "format": "json",
        "pageno": str(pageno),
    }
    if categories:
        params["categories"] = categories
    if language:
        params["language"] = language
    if time_range:
        params["time_range"] = time_range

    url = f"{base_url.rstrip('/')}/search?{urllib.parse.urlencode(params)}"

    req = urllib.request.Request(url, method="GET")
    req.add_header("Accept", "application/json")

    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return json.loads(resp.read())
    except urllib.error.HTTPError as e:
        err_body = e.read().decode() if e.fp else ""
        raise RuntimeError(f"SearXNG HTTP {e.code}: {err_body}")
    except urllib.error.URLError as e:
        raise RuntimeError(f"SearXNG connection failed: {e.reason}")


def format_results(data: dict) -> dict:
    """Format SearXNG response into a clean result structure."""
    results = []
    for i, r in enumerate(data.get("results", []), 1):
        entry = {
            "index": i,
            "title": r.get("title", ""),
            "url": r.get("url", ""),
            "content": r.get("content", ""),
            "engine": r.get("engine", ""),
            "category": r.get("category", ""),
        }
        if r.get("publishedDate"):
            entry["published_date"] = r["publishedDate"]
        results.append(entry)

    return {
        "query": data.get("query", ""),
        "result_count": len(results),
        "results": results,
        "answers": data.get("answers", []),
        "suggestions": data.get("suggestions", []),
    }


def main():
    args = json.loads(sys.argv[1])
    query = args.get("query", "")
    if not query:
        print(json.dumps({"error": "query is required"}))
        sys.exit(1)

    base_url = os.environ.get("MCP_SEARXNG_BASE_URL", "")
    if not base_url:
        print(json.dumps({"error": "MCP_SEARXNG_BASE_URL env var is not set"}))
        sys.exit(1)

    categories = args.get("categories")
    language = args.get("language")
    time_range = args.get("time_range")
    pageno = args.get("pageno", 1)
    if isinstance(pageno, str):
        pageno = int(pageno)

    try:
        raw = search_searxng(base_url, query, categories, language, time_range, pageno)
        result = format_results(raw)
        print(json.dumps(result))
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)


if __name__ == "__main__":
    main()