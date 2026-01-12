#!/usr/bin/env python3
"""
Work Log Rotation
Rotates work logs when they exceed size/token limits to keep them manageable.
"""

import os
import sys
from pathlib import Path
from datetime import datetime

# Token estimation: ~4 chars per token
MAX_TOKENS = 20000
MAX_CHARS = MAX_TOKENS * 4  # 80,000 characters

def estimate_tokens(file_path):
    """Estimate token count from file size."""
    try:
        size = os.path.getsize(file_path)
        return size // 4  # Rough estimate: 4 chars per token
    except:
        return 0

def rotate_log(work_log_path, keep_tail_tokens=5000):
    """Rotate a work log if it's too large.

    Args:
        work_log_path: Path to work log file
        keep_tail_tokens: Number of tokens to copy from tail into new log for continuity
    """
    work_log = Path(work_log_path)

    if not work_log.exists():
        print(f"Work log does not exist: {work_log_path}")
        return False

    tokens = estimate_tokens(work_log)

    if tokens < MAX_TOKENS:
        print(f"Work log size OK: {tokens:,} tokens (limit: {MAX_TOKENS:,})")
        return False

    print(f"Work log too large: {tokens:,} tokens (limit: {MAX_TOKENS:,})")

    # Read current log content
    content = work_log.read_text()

    # Create archive directory
    archive_dir = work_log.parent / "archive"
    archive_dir.mkdir(exist_ok=True)

    # Generate archive filename with timestamp
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    archive_name = f"20-work-log.{timestamp}.md"
    archive_path = archive_dir / archive_name

    # Copy full log to archive
    archive_path.write_text(content)
    print(f"Archived full log: {archive_path}")

    # Extract tail for continuity (approximate by characters)
    keep_chars = keep_tail_tokens * 4  # ~4 chars per token
    tail_content = content[-keep_chars:] if len(content) > keep_chars else content

    # Find a good break point (start of a section)
    # Look for markdown headers to avoid cutting mid-section
    lines = tail_content.split('\n')

    # Find first header (## or ###) to start from
    start_idx = 0
    for i, line in enumerate(lines):
        if line.startswith('##'):
            start_idx = i
            break

    tail_content = '\n'.join(lines[start_idx:])

    # Create new work log with header + tail
    task_dir = work_log.parent
    task_name = task_dir.name

    new_content = f"""# Work Log - {task_name}

**Note:** This is a continuation log. Previous work log archived to `archive/{archive_name}`

**Recent Context (last ~{keep_tail_tokens:,} tokens):**

---

{tail_content}

---

## Session {timestamp}

Started: {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}

"""

    work_log.write_text(new_content)
    print(f"Created new work log with tail context: {work_log_path}")
    print(f"Preserved ~{estimate_tokens(work_log):,} tokens of recent context")

    return True

def rotate_all_logs(tasks_dir=".ai/tasks"):
    """Rotate all work logs in tasks directory."""
    tasks_path = Path(tasks_dir)

    if not tasks_path.exists():
        print(f"Tasks directory not found: {tasks_dir}")
        return

    rotated = 0
    for task_dir in tasks_path.iterdir():
        if not task_dir.is_dir():
            continue

        work_log = task_dir / "20-work-log.md"
        if work_log.exists():
            if rotate_log(work_log):
                rotated += 1

    print(f"\nRotated {rotated} work log(s)")

def main():
    """Main CLI interface."""
    if len(sys.argv) < 2:
        print("Usage:")
        print("  python3 rotate-work-log.py <work-log-path>  # Rotate specific log")
        print("  python3 rotate-work-log.py --all            # Rotate all logs")
        sys.exit(1)

    if sys.argv[1] == "--all":
        rotate_all_logs()
    else:
        work_log_path = sys.argv[1]
        rotate_log(work_log_path)

if __name__ == "__main__":
    main()
