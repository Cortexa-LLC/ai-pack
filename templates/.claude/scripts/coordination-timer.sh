#!/bin/bash
# Coordination Timer - Triggers periodic coordinator check-ins
# Usage: bash .claude/scripts/coordination-timer.sh [interval_seconds] [max_checks]

INTERVAL_SECONDS=${1:-30}  # Default: 30 seconds
MAX_CHECKS=${2:-1200}       # Default: 1200 checks (10 hours at 30s intervals)
CHECKPOINT_FILE=".claude/.coordination-checkpoint"

echo "Starting coordination timer..."
echo "Interval: ${INTERVAL_SECONDS} seconds"
echo "Checkpoint file: ${CHECKPOINT_FILE}"
echo "Max checks: ${MAX_CHECKS}"

# Initialize checkpoint file
mkdir -p .claude
echo "0" > "${CHECKPOINT_FILE}"

# Run timer loop
CHECK_COUNT=0
while [ ${CHECK_COUNT} -lt ${MAX_CHECKS} ]; do
    # Sleep for the interval
    sleep ${INTERVAL_SECONDS}

    # Increment check count
    CHECK_COUNT=$((CHECK_COUNT + 1))

    # Update checkpoint file with timestamp and count
    TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
    echo "${CHECK_COUNT}" > "${CHECKPOINT_FILE}"
    echo "${TIMESTAMP}" >> "${CHECKPOINT_FILE}"
    echo "⏰ Coordination checkpoint ${CHECK_COUNT} triggered at ${TIMESTAMP}" >> "${CHECKPOINT_FILE}"

    echo "[$(date)] Checkpoint ${CHECK_COUNT}/${MAX_CHECKS} triggered"
done

echo "Coordination timer completed after ${MAX_CHECKS} checks"
