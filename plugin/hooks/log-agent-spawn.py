#!/usr/bin/env python3
"""PostToolUse hook (matcher: Agent): log subagent spawns for adherence metrics.

Appends one JSON line per Agent-tool invocation to <cwd>/.ai/logs/agent-spawns.jsonl:
    {"ts": <epoch float>, "session_id": ..., "subagent_type": ..., "description": ...}

Spawns sharing a timestamp burst indicate a parallel batch.

Rule: a logging hook must never break a session — every step is wrapped in
try/except and the script always exits 0, silently, no matter what fails.
"""
import json
import os
import sys
import time

try:
    data = json.load(sys.stdin)
    tool_input = data.get("tool_input") or {}
    record = {
        "ts": time.time(),
        "session_id": data.get("session_id"),
        "subagent_type": tool_input.get("subagent_type"),
        "description": tool_input.get("description"),
    }
    log_dir = os.path.join(os.getcwd(), ".ai", "logs")
    os.makedirs(log_dir, exist_ok=True)
    with open(os.path.join(log_dir, "agent-spawns.jsonl"), "a") as f:
        f.write(json.dumps(record) + "\n")
except Exception:
    pass  # never break the session over logging

sys.exit(0)
