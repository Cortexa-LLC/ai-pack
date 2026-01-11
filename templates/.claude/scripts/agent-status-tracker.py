#!/usr/bin/env python3
"""
Agent Status Tracker
Tracks completion status of agents and generates reports for Orchestrator.
"""

import os
import json
import sys
from pathlib import Path
from datetime import datetime

STATUS_FILE = ".claude/.agent-status.json"

class AgentStatusTracker:
    def __init__(self):
        self.status_file = Path(STATUS_FILE)
        self.status_file.parent.mkdir(parents=True, exist_ok=True)
        self.data = self.load()

    def load(self):
        """Load existing status data."""
        if self.status_file.exists():
            try:
                with open(self.status_file, 'r') as f:
                    return json.load(f)
            except:
                return self.default_data()
        return self.default_data()

    def default_data(self):
        """Return default status structure."""
        return {
            "agents": {},
            "last_update": None,
            "orchestrator_id": None
        }

    def save(self):
        """Save status data."""
        with open(self.status_file, 'w') as f:
            json.dump(self.data, f, indent=2)

    def register_agent(self, agent_id, role, task, orchestrator_id):
        """Register a new agent."""
        self.data["agents"][agent_id] = {
            "role": role,
            "task": task,
            "status": "active",
            "started": datetime.now().isoformat(),
            "last_update": datetime.now().isoformat(),
            "completed": None,
            "work_log": None,
            "commits": 0,
            "blockers": []
        }
        self.data["orchestrator_id"] = orchestrator_id
        self.data["last_update"] = datetime.now().isoformat()
        self.save()
        print(f"Registered agent: {agent_id} ({role})")

    def update_agent(self, agent_id, **kwargs):
        """Update agent status."""
        if agent_id not in self.data["agents"]:
            print(f"Warning: Agent {agent_id} not registered")
            return

        agent = self.data["agents"][agent_id]
        for key, value in kwargs.items():
            agent[key] = value
        agent["last_update"] = datetime.now().isoformat()
        self.data["last_update"] = datetime.now().isoformat()
        self.save()
        print(f"Updated agent: {agent_id} - {kwargs}")

    def mark_complete(self, agent_id):
        """Mark agent as completed."""
        self.update_agent(
            agent_id,
            status="completed",
            completed=datetime.now().isoformat()
        )

    def mark_blocked(self, agent_id, blocker):
        """Mark agent as blocked."""
        if agent_id not in self.data["agents"]:
            return

        agent = self.data["agents"][agent_id]
        agent["status"] = "blocked"
        agent["blockers"].append({
            "blocker": blocker,
            "timestamp": datetime.now().isoformat()
        })
        agent["last_update"] = datetime.now().isoformat()
        self.save()

    def generate_report(self):
        """Generate status report for Orchestrator."""
        agents = self.data["agents"]

        if not agents:
            return {
                "summary": "No agents registered",
                "active": 0,
                "completed": 0,
                "blocked": 0,
                "agents": []
            }

        active = [a for a in agents.values() if a["status"] == "active"]
        completed = [a for a in agents.values() if a["status"] == "completed"]
        blocked = [a for a in agents.values() if a["status"] == "blocked"]

        return {
            "summary": f"{len(completed)}/{len(agents)} completed, {len(active)} active, {len(blocked)} blocked",
            "active_count": len(active),
            "completed_count": len(completed),
            "blocked_count": len(blocked),
            "total_count": len(agents),
            "agents": [
                {
                    "id": agent_id,
                    "role": agent["role"],
                    "task": agent["task"],
                    "status": agent["status"],
                    "commits": agent["commits"],
                    "blockers": agent["blockers"]
                }
                for agent_id, agent in agents.items()
            ]
        }

def main():
    """Main CLI interface."""
    if len(sys.argv) < 2:
        print("Usage:")
        print("  python3 agent-status-tracker.py register <agent_id> <role> <task> <orchestrator_id>")
        print("  python3 agent-status-tracker.py update <agent_id> <key=value> ...")
        print("  python3 agent-status-tracker.py complete <agent_id>")
        print("  python3 agent-status-tracker.py blocked <agent_id> <blocker>")
        print("  python3 agent-status-tracker.py report")
        sys.exit(1)

    tracker = AgentStatusTracker()
    command = sys.argv[1]

    if command == "register":
        agent_id, role, task, orchestrator_id = sys.argv[2:6]
        tracker.register_agent(agent_id, role, task, orchestrator_id)

    elif command == "update":
        agent_id = sys.argv[2]
        updates = {}
        for arg in sys.argv[3:]:
            if "=" in arg:
                key, value = arg.split("=", 1)
                # Try to parse as int
                try:
                    value = int(value)
                except:
                    pass
                updates[key] = value
        tracker.update_agent(agent_id, **updates)

    elif command == "complete":
        agent_id = sys.argv[2]
        tracker.mark_complete(agent_id)

    elif command == "blocked":
        agent_id = sys.argv[2]
        blocker = " ".join(sys.argv[3:])
        tracker.mark_blocked(agent_id, blocker)

    elif command == "report":
        report = tracker.generate_report()
        print("\n" + "="*60)
        print("AGENT STATUS REPORT")
        print("="*60)
        print(f"Summary: {report['summary']}")
        print(f"Active: {report['active_count']}")
        print(f"Completed: {report['completed_count']}")
        print(f"Blocked: {report['blocked_count']}")
        print("\nAgent Details:")
        for agent in report['agents']:
            status_icon = {
                "active": "🟢",
                "completed": "✅",
                "blocked": "🔴"
            }.get(agent['status'], "⚪")
            print(f"\n  {status_icon} {agent['id']} ({agent['role']})")
            print(f"     Task: {agent['task']}")
            print(f"     Status: {agent['status']}")
            print(f"     Commits: {agent['commits']}")
            if agent['blockers']:
                print(f"     Blockers: {len(agent['blockers'])}")
                for blocker in agent['blockers']:
                    print(f"       - {blocker['blocker']}")
        print("="*60 + "\n")

if __name__ == "__main__":
    main()
