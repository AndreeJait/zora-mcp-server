#!/usr/bin/env python3
"""WAHA Reaction Tool — React to a WhatsApp message with an emoji.

Env vars:
  MCP_WAHA_BASE_URL  - WAHA server URL (required)
  MCP_WAHA_API_KEY    - API key (optional, sent as X-Api-Key header)
  MCP_WAHA_SESSION    - Session name (default: "default")

Args (JSON via sys.argv[1]):
  chatId    - WhatsApp chat ID (required)
  messageId - WhatsApp message ID to react to (required)
  emoji     - Emoji reaction, empty string to remove (required)

WAHA endpoint: PUT /api/reaction
"""

import json
import os
import sys
import urllib.request
import urllib.error


def main():
    args = json.loads(sys.argv[1])
    chat_id = args["chatId"]
    message_id = args["messageId"]
    emoji = args["emoji"]

    base_url = os.environ["MCP_WAHA_BASE_URL"].rstrip("/")
    api_key = os.environ.get("MCP_WAHA_API_KEY", "")
    session = os.environ.get("MCP_WAHA_SESSION", "default")

    body = json.dumps({
        "session": session,
        "chatId": chat_id,
        "messageId": message_id,
        "reaction": emoji,
    }).encode()

    req = urllib.request.Request(
        f"{base_url}/api/reaction",
        data=body,
        method="PUT",
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