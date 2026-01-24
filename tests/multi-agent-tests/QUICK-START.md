# Multi-Agent Tests - Quick Start Guide

**For running validation tests on a new machine**

---

## Quick Setup (New Machine)

**Run the automated setup script:**

```bash
python3 tests/tools/setup-multi-agent-environment.py
```

This will:
- Create `.claude/` directory if missing
- Create/update `.claude/settings.json` with required permissions
- Create `.ai/test-artifacts/` directory
- Verify test files are present

**Then restart Claude Code** for settings to take effect.

---

## Manual Setup (Alternative)

If you prefer manual setup, create `.claude/settings.json`:

```bash
mkdir -p .claude
cat > .claude/settings.json <<'EOF'
{
  "permissions": {
    "allow": [
      "Write(*)",
      "Edit(*)",
      "Read(*)"
    ],
    "defaultMode": "bypassPermissions"
  }
}
EOF
```

**Why:** Spawned Agents need file write permissions to create test artifacts.

**Reference:** `docs/CLAUDE-CODE-CONFIGURATION.md`

---

## Verify Configuration

Run the environment diagnostic to check everything:

```bash
python3 tests/tools/check-multi-agent-environment.py
```

This checks:
- ✓ Claude Code settings.json (CRITICAL)
- ✓ Working directory
- ✓ Git state
- ✓ Python version
- ✓ Test files present
- ✓ Test imports work

---

## Running Multi-Agent Tests

### Individual Test Files

```bash
# Simple execution test
python3 tests/test_multi-agent_real_execution.py

# Sequential workflow test
python3 tests/test_multi-agent_sequential_workflows.py

# Parallel execution test (5 agents)
python3 tests/test_multi-agent_parallel_execution.py

# Complex mixed workflow test (8 agents)
python3 tests/test_multi-agent_complex_mixed_workflows.py

# Token limit stress tests
python3 tests/test_multi-agent_token_limits.py

# Watchdog and timer tests
python3 tests/test_multi-agent_watchdog_and_timers.py
```

### Run All Multi-Agent Tests

```bash
# Using test runner
python3 tests/run_tests.py --multi-agent

# Or manually
for test in tests/test_multi-agent*.py; do
    echo "Running $test..."
    python3 "$test"
done
```

---

## Command for Claude Agent

**To have a Claude agent run tier 2 tests, use:**

```
Run the tier 2 validation tests:

python3 tests/test_multi-agent_real_execution.py
```

**Or for all tests:**

```
Run all tier 2 validation tests:

python3 tests/run_tests.py --multi-agent
```

---

## What Gets Created

Tests create artifacts in `.ai/test-artifacts/`:

```
.ai/test-artifacts/
├── multi-agent-real-{timestamp}/
│   ├── output/
│   │   ├── model.py
│   │   ├── service.py
│   │   └── test_service.py
│   └── tasks/
│       └── 2026-01-15_multi-agent-simple-task/
│           └── 00-contract.md
│
├── multi-agent-sequential-{timestamp}/
│   ├── output/
│   │   ├── requirements.md
│   │   ├── architecture.md
│   │   ├── profile/
│   │   ├── tests/
│   │   └── review.md
│   └── stages/
│
├── multi-agent-parallel-{timestamp}/
│   ├── output/
│   │   ├── feature-1-auth/
│   │   ├── feature-2-api/
│   │   ├── feature-3-cache/
│   │   ├── feature-4-validator/
│   │   └── feature-5-logger/
│   └── tasks/
│
└── multi-agent-complex-mixed-{timestamp}/
    ├── output/
    │   ├── module-a-cart/
    │   ├── module-b-pricing/
    │   ├── module-c-checkout/
    │   └── integration_test_results.md
    └── stages/
```

---

## Troubleshooting

### Test fails with "Permission denied"

**Problem:** Spawned Agents cannot write files

**Solution:**
1. Verify `.claude/settings.json` exists with correct permissions
2. Restart Claude Code after creating settings.json
3. Re-run diagnostic: `python3 tests/tools/check-multi-agent-environment.py`

### Test files not found

**Problem:** Not in ai-pack directory

**Solution:**
```bash
cd path/to/ai-pack
python3 tests/test_multi-agent_real_execution.py
```

### Import errors

**Problem:** Python path issues

**Solution:**
```bash
# Run from ai-pack root directory
pwd  # Should end in /ai-pack
python3 tests/test_multi-agent_real_execution.py
```

### Tests hang indefinitely

**Problem:** Agent waiting for user input or permissions

**Solution:**
1. Check `.claude/settings.json` has `"defaultMode": "bypassPermissions"`
2. Restart Claude Code
3. Re-run test

---

## Expected Results

### test_multi-agent_real_execution.py
- ✅ Spawns spawned agent
- ✅ Creates model.py, service.py, test_service.py
- ✅ Verifies all files exist
- ✅ Duration: ~30-60 seconds

### test_multi-agent_sequential_workflows.py
- ✅ Runs 5-stage sequential workflow (PRD → Architect → Engineer → Reviewer → Tester)
- ✅ Each stage completes before next starts
- ✅ Creates comprehensive artifacts
- ✅ Duration: ~3-5 minutes

### test_multi-agent_parallel_execution.py
- ✅ Spawns 5 spawned agents simultaneously
- ✅ All agents run independently
- ✅ Creates 5 feature directories
- ✅ Duration: ~2-3 minutes

### test_multi-agent_complex_mixed_workflows.py
- ✅ Orchestrator spawns sequential + parallel agents (8 total)
- ✅ 3 parallel engineers → 3 parallel reviewers → 1 integration tester
- ✅ All artifacts created
- ✅ Duration: ~5-8 minutes

### test_multi-agent_token_limits.py
- ✅ 15-file batch: PASS (acceptable, high token usage)
- ⚠️ 20-file batch: MARGINAL (very high token usage)
- ❌ 25-file batch: FAIL (exceeds token limit)

### test_multi-agent_watchdog_and_timers.py
- ✅ Watchdog detects timeout
- ✅ Progress monitoring works
- ✅ Graceful failure handling
- ✅ Duration: ~2-3 minutes

---

## Common Issues Between Machines

### Works on Machine A, fails on Machine B

**Check:**

1. **Settings differ:**
   ```bash
   # On both machines:
   cat .claude/settings.json
   # Should be identical
   ```

2. **Different branches:**
   ```bash
   git branch --show-current
   # Should be: main or validation-tests
   ```

3. **Claude Code version:**
   - Ensure both machines have same Claude Code version
   - Features may vary between versions

4. **Python version:**
   ```bash
   python3 --version
   # Should be Python 3.8+
   ```

5. **Working directory:**
   ```bash
   pwd
   # Must be in ai-pack root, not subdirectory
   ```

---

## Reference Documentation

- **Configuration:** `docs/CLAUDE-CODE-CONFIGURATION.md`
- **Test Guide:** `tests/test-results.md`
- **Test Results:** `tests/test-results.md`
- **Environment Check:** `tests/tools/check-multi-agent-environment.py`

---

**Last Updated:** 2026-01-15
