#!/usr/bin/env python3
"""WAHA Typing Tool — Start or stop typing indicator in a WhatsApp chat.

Env vars:
  MCP_WAHA_BASE_URL  - WAHA server URL (required)
  MCP_WAHA_API_KEY    - API key (optional, sent as X-Api-Key header)
  MCP_WAHA_SESSION    - Session name (default: "default")

Args (JSON via sys.argv[1]):
  chatId - WhatsApp chat ID (required)
  action - "start" or "stop" (required)

WAHA endpoints: POST /api/startTyping, POST /api/stopTyping
"""

import json
import os
import sys
import urllib.request
import urllib.error


def main():
    args = json.loads(sys.argv[1])
    chat_id = args["chatId"]
    action = args["action"]

    if action not in ("start", "stop"):
        print(json.dumps({"error": f"invalid action '{action}', must be 'start' or 'stop'"}))
        sys.exit(1)

    base_url = os.environ["MCP_WAHA_BASE_URL"].rstrip("/")
    api_key = os.environ.get("MCP_WAHA_API_KEY", "")
    session = os.environ.get("MCP_WAHA_SESSION", "default")

    endpoint = f"{base_url}/api/{action}Typing"
    body = json.dumps({
        "session": session,
        "chatId": chat_id,
    }).encode()

    req = urllib.request.Request(
        endpoint,
        data=body,
        method="POST",
        headers={"Content-Type": "application/json"},
    )
    if api_key:
        req.add_header("X-Api-Key", api_key)

    try:
        with urllib.request.urlopen(req) as resp:
            print(json.dumps({"status": "ok", "action": f"{action}Typing"}))
    except urllib.error.HTTPError as e:
        err_body = e.read().decode() if e.fp else ""
        print(json.dumps({"error": f"WAHA HTTP {e.code}: {err_body}"}))
        sys.exit(1)
    except urllib.error.URLError as e:
        print(json.dumps({"error": f"WAHA connection failed: {e.reason}"}))
        sys.exit(1)


if __name__ == "__main__":
    main()