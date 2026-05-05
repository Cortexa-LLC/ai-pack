# Dead Code Review - May 5, 2026

## Executive Summary

Comprehensive dead code analysis completed using `deadcode` static analysis tool and manual review. **68 items identified** requiring cleanup across 5 categories.

**Total estimated cleanup effort:** 5-7 days across 3 weeks

---

## Critical Findings

### 1. Obsolete "Beads" References (533+ instances) 🔴

Despite recent cleanup efforts (tasks ai-pack-fe5, ai-pack-416, ai-pack-4d1, ai-pack-c84), **533+ references to "beads"** remain throughout the codebase.

**Primary locations:**
- `scripts/github-integration.py` - Functions `sync_beads_to_github()`, `import_github_to_beads()`
- `scripts/README.md` - Reference to non-existent `backup-beads.py`
- ~531 other references in docs and code

**Action Required:** Complete beads removal effort OR clearly document if this is still a supported integration.

---

### 2. Entire MCP Server Module Unused 🔴

**File:** `internal/mcp/server.go`

All 5 exported functions are unreachable:
- `NewServer` (line 20)
- `Server.send` (line 29)
- `Server.sendResult` (line 33)
- `Server.sendError` (line 37)
- `Server.Serve` (line 48)

**Action Required:** Delete entire module OR document planned usage with clear TODOs.

---

### 3. Config Validator Module Completely Unused 🔴

**File:** `internal/server/config_validator.go`

All 6 validation functions are unreachable:
- `NewConfigValidator`
- `ValidateAllConfigs`
- `validateAgentDirectory`
- `validateConfig`
- `validateModelAvailability`
- `PrintValidationReport`

**Risk:** Configuration errors discovered at runtime instead of startup.

**Action Required:** Delete OR integrate validation into server initialization.

---

## Major Findings

### 4. Dead GraphQL Resolver TODOs 🟡

**File:** `internal/graphql/schema.resolvers.go`

4+ resolver functions return "TODO: Implement" errors:
- Task status update logic
- Retry task logic
- Logs query
- Additional unimplemented resolvers

**Impact:** GraphQL API advertises endpoints that don't work.

**Action:** Implement the resolvers OR remove from schema.

---

### 5. Unreachable Functions (33 total) 🟡

Detected via `deadcode -test ./...`:

**cmd/agent/commands/spawn.go:**
- `wrapText` (line 394)
- `minInt` (line 423)

**internal/auth/apikey.go:**
- `GetBaseURL` (line 120) - duplicate implementation

**internal/graphql/scalars.go:**
- `MarshalTime` (line 13)
- `UnmarshalTime` (line 23)

**internal/server/chat_handler.go:**
- `generateFollowUpSuggestion` (line 568)
- `loadProjectContext` (line 603)

**internal/server/performance_handlers.go:**
- `registerPerformanceRoutes` (line 298)

**internal/server/task_metadata.go:**
- `createTaskPacket` (line 17)
- `updateTaskPacketMetadata` (line 63)
- `buildSystemPrompt` (line 289)

**internal/tools/tools.go:**
- `defineWebSearchTool` (line 210)
- `defineWebFetchTool` (line 228)

**Plus 20 more** - see full `deadcode` output for complete list.

**Action:** Review each function, remove if truly unused, or add TODO comments if planned for future.

---

### 6. Orphaned Test Files (16 files) 🟡

Test files without corresponding source files:
- `archival_test.go`
- `complexity_risk_analyzer_test.go`
- `detect_out_of_scope_test.go`
- `export_test.go`
- `grade_fields_test.go`
- `mcp_schema_test.go`
- `message_truncation_test.go`
- `preflight_related_test.go`
- `protocol_test.go`
- `resume_test.go`
- `role_extends_test.go`
- `route_contract_test.go`
- `search_knowledge_cross_project_test.go`
- `task_description_test.go`
- `taskdb_integration_test.go`
- `timeout_test.go`

**Action:** Review each test. If testing removed features, delete. If source file has different name, rename or document.

---

## Minor Findings

### 7. Obsolete OpenAI Codex Reference 🟢

**File:** `internal/streaming/openai_adapter.go:290`  
**Function:** `IsCodexModel`

Codex has been deprecated by OpenAI.

**Action:** Remove function.

---

### 8. Duplicate GetBaseURL Implementations 🟢

**Files:**
- `internal/auth/apikey.go:120`
- `internal/proxy/transport.go:115`

**Action:** Consolidate to single canonical implementation.

---

### 9. Template File Without References 🟢

**File:** `agent-server.json.template`

Only referenced in error logs, not in setup scripts or documentation.

**Action:** Integrate into setup flow OR remove if superseded by `config/agent-server.example.json`.

---

### 10. Additional Minor Dead Functions 🟢

- `internal/monitoring/execution_metrics.go:66` - `ComputeExplorationRatio`
- `internal/monitoring/logger.go:254` - `LogRateLimitExceeded`
- `internal/monitoring/model_selector.go:493` - `EstimateCost`
- `internal/tools/claudeignore.go:279` - `ClaudeIgnore.FilterPaths`
- `internal/server/server_helpers.go:906` - `loadRoleContext`
- `internal/server/server_helpers.go:938` - `findMostRecentExecutionFolderInRoot`
- `internal/server/orchestrator_session.go:278` - `OrchestratorSession.Stop`

---

## Cleanup Roadmap

### Priority 1: Critical Cleanup (Week 1, 2-3 days)

1. **Complete beads removal** (533+ references)
2. **Delete `internal/mcp/server.go`** (entire module)
3. **Delete `internal/server/config_validator.go`** (entire module)

### Priority 2: Major Cleanup (Week 2, 2-3 days)

4. **Remove 33 unreachable functions** or add TODO comments
5. **Resolve GraphQL TODO resolvers** (implement or remove)
6. **Review 16 orphaned test files** (delete or document)

### Priority 3: Minor Cleanup (Week 3, 1-2 days)

7. Remove obsolete Codex reference
8. Consolidate GetBaseURL implementations
9. Review template file usage
10. Remove remaining minor dead functions

---

## Build Verification

✅ **Go build:** `go build ./...` - PASSES  
✅ **Go vet:** `go vet ./...` - PASSES  
⚠️  **Static analysis:** `deadcode -test ./...` - Shows 33 unreachable functions (expected)

**Note:** Build passes because dead code is never called. Removing dead code will not break builds.

---

## How to Use This Report

1. **Start with Priority 1** - Highest impact, clearest action items
2. **Create separate PRs** for each major category (easier review)
3. **Run verification** after each cleanup:
   ```bash
   go build ./...
   go test ./...
   deadcode -test ./...
   go vet ./...
   ```
4. **Document uncertainty** - If unclear whether code is truly dead, ask in PR review

---

## Detailed Checklist

For a detailed, actionable checklist with bash commands and step-by-step instructions, see:

**`.ai/tasks/dead-code-review-20260505/cleanup-checklist.md`**

(Note: This file is in gitignored `.ai/` directory. Access locally or request regeneration.)

---

## Analysis Methodology

**Tools Used:**
- `deadcode -test ./...` - Go static analysis for unreachable code
- `git log --since="6 months ago"` - File usage history
- Manual code review - Cross-referencing, import analysis
- Knowledge graph context - Recent cleanup task history

**Coverage:**
- ✅ Go source files (cmd/, internal/, pkg/)
- ✅ Python scripts (scripts/, tests/)
- ✅ Shell scripts (scripts/, tests/)
- ✅ Test files (all *_test.go)
- ✅ Configuration and template files
- ❌ Excluded: vendor/, .ai/tasks/, node_modules/, historical docs

---

## Knowledge Graph Entries

Recorded in project knowledge graph:
- **Topic:** `dead-code-review-2026-05-05`
- **Files:** `internal/mcp/server.go`, `internal/server/config_validator.go`, `scripts/github-integration.py`
- **Observations:** Critical findings for each dead code module

Query with: `kg__search_knowledge({query: "dead code review"})`

---

## Next Steps

1. Review this report with team
2. Prioritize cleanup work
3. Assign tasks from checklist
4. Create tracking issues/tasks for each priority level
5. Begin Priority 1 cleanup (estimated 2-3 days)

---

**Report Generated:** May 5, 2026  
**Reviewer:** AI Agent (reviewer role)  
**Analysis Scope:** Full repository (excluding vendor, node_modules, .ai/tasks)  
**Total Items Found:** 68 items across 5 categories
