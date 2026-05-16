#!/usr/bin/env python3
"""WAHA Send Text Tool — Send a text message to a WhatsApp chat.

Env vars:
  MCP_WAHA_BASE_URL  - WAHA server URL (required)
  MCP_WAHA_API_KEY    - API key (optional, sent as X-Api-Key header)
  MCP_WAHA_SESSION    - Session name (default: "default")

Args (JSON via sys.argv[1]):
  chatId  - WhatsApp chat ID (required)
  text    - Message text (required)
  replyTo - Message ID to reply to (optional)

WAHA endpoint: POST /api/sendText
"""

import json
import os
import sys
import urllib.request
import urllib.error


def main():
    args = json.loads(sys.argv[1])
    chat_id = args["chatId"]
    text = args["text"]

    base_url = os.environ["MCP_WAHA_BASE_URL"].rstrip("/")
    api_key = os.environ.get("MCP_WAHA_API_KEY", "")
    session = os.environ.get("MCP_WAHA_SESSION", "default")

    body_map = {
        "session": session,
        "chatId": chat_id,
        "text": text,
    }
    if args.get("replyTo"):
        body_map["reply_to"] = args["replyTo"]

    body = json.dumps(body_map).encode()

    req = urllib.request.Request(
        f"{base_url}/api/sendText",
        data=body,
        method="POST",
        headers={"Content-Type": "application/json"},
    )
    if api_key:
        req.add_header("X-Api-Key", api_key)

    try:
        with urllib.request.urlopen(req) as resp:
            result = json.loads(resp.read())
            print(json.dumps(result))
    except urllib.error.HTTPError as e:
        err_body = e.read().decode() if e.fp else ""
        print(json.dumps({"error": f"WAHA HTTP {e.code}: {err_body}"}))
        sys.exit(1)
    except urllib.error.URLError as e:
        print(json.dumps({"error": f"WAHA connection failed: {e.reason}"}))
        sys.exit(1)


if __name__ == "__main__":
    main()