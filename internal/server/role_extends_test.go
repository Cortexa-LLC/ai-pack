package server

import (
	"os"
	"path/filepath"
	"testing"
)

// ── mergeRoleContent ─────────────────────────────────────────────────────────

func TestMergeRoleContent_ProjectSectionOverridesBase(t *testing.T) {
	base := `# Overview
Base overview text.

# Responsibilities
Base responsibilities.

# Guidelines
Base guidelines.
`
	project := `# Overview
Project override of overview.
`
	merged := mergeRoleContent(base, project)

	if !containsString(merged, "Project override of overview.") {
		t.Error("merged content should contain project overview text")
	}
	if containsString(merged, "Base overview text.") {
		t.Error("merged content should not contain base overview text (was overridden)")
	}
	// Sections absent in project are inherited from base.
	if !containsString(merged, "Base responsibilities.") {
		t.Error("merged content should inherit base Responsibilities section")
	}
	if !containsString(merged, "Base guidelines.") {
		t.Error("merged content should inherit base Guidelines section")
	}
}

func TestMergeRoleContent_ProjectOnlySectionAppended(t *testing.T) {
	base := `# Overview
Base overview.
`
	project := `# ProjectSpecific
Project-only section.
`
	merged := mergeRoleContent(base, project)

	if !containsString(merged, "Base overview.") {
		t.Error("merged content should contain inherited base overview")
	}
	if !containsString(merged, "Project-only section.") {
		t.Error("merged content should contain project-only section")
	}
}

func TestMergeRoleContent_EmptyProjectInheritsAll(t *testing.T) {
	base := `# Overview
Everything from base.

# Responsibilities
Base duties.
`
	merged := mergeRoleContent(base, "")

	if !containsString(merged, "Everything from base.") {
		t.Error("merged content should contain all base sections when project is empty")
	}
}

// ── mergeRoleConfigs ─────────────────────────────────────────────────────────

func makeBaseConfig() *AgentConfig {
	cfg := &AgentConfig{
		Name:        "engineer",
		Description: "Base engineer description",
		Tier:        "minimal",
		Model:       "gpt-4o-mini",
		Tools:       []string{"read", "write", "bash"},
		ExplicitFields: map[string]bool{
			configFieldTier: true,
		},
	}
	cfg.Delegation.Mode = "delegate"
	cfg.Delegation.Timeout = "30min"
	cfg.Delegation.MaxContext = 32000
	cfg.Context.RoleContent = "# Overview\nBase overview.\n\n# Responsibilities\nBase duties.\n"
	cfg.Context.RoleFile = "/path/to/roles/engineer.md"
	return cfg
}

func TestMergeRoleConfigs_TierLockedError(t *testing.T) {
	base := makeBaseConfig()
	project := &AgentConfig{
		Name:  "engineer",
		Tier:  "standard",
		Model: "gpt-4o",
		ExplicitFields: map[string]bool{
			configFieldTier:  true,
			configFieldModel: true,
		},
	}
	_, err := mergeRoleConfigs(base, project)
	if err == nil {
		t.Fatal("expected error when project file sets Tier:, got nil")
	}
}

func TestMergeRoleConfigs_OmittedFieldInheritsFromBase(t *testing.T) {
	base := makeBaseConfig()
	project := &AgentConfig{
		Name:           "engineer",
		ExplicitFields: map[string]bool{},
		// Model not set — should inherit from base
	}
	project.Context.RoleContent = ""

	merged, err := mergeRoleConfigs(base, project)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged.Model != base.Model {
		t.Errorf("merged.Model = %q, want %q (inherited from base)", merged.Model, base.Model)
	}
	if merged.Tier != base.Tier {
		t.Errorf("merged.Tier = %q, want %q (inherited from base)", merged.Tier, base.Tier)
	}
	if merged.Delegation.Timeout != base.Delegation.Timeout {
		t.Errorf("merged.Delegation.Timeout = %q, want %q", merged.Delegation.Timeout, base.Delegation.Timeout)
	}
}

func TestMergeRoleConfigs_ExplicitProjectFieldOverrides(t *testing.T) {
	base := makeBaseConfig()
	project := &AgentConfig{
		Name:        "engineer",
		Description: "Project-specific description",
		Model:       "gpt-4o",
		ExplicitFields: map[string]bool{
			configFieldDescription: true,
			configFieldModel:       true,
		},
	}
	project.Context.RoleContent = ""

	merged, err := mergeRoleConfigs(base, project)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged.Description != "Project-specific description" {
		t.Errorf("merged.Description = %q, want project override", merged.Description)
	}
	if merged.Model != "gpt-4o" {
		t.Errorf("merged.Model = %q, want gpt-4o", merged.Model)
	}
	// Tier should still come from base.
	if merged.Tier != base.Tier {
		t.Errorf("merged.Tier = %q, want %q (locked to base)", merged.Tier, base.Tier)
	}
}

func TestMergeRoleConfigs_ToolsReplacedByProject(t *testing.T) {
	base := makeBaseConfig()
	project := &AgentConfig{
		Name:           "engineer",
		Tools:          []string{"read"},
		ExplicitFields: map[string]bool{configFieldTools: true},
	}
	project.Context.RoleContent = ""

	merged, err := mergeRoleConfigs(base, project)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(merged.Tools) != 1 || merged.Tools[0] != "read" {
		t.Errorf("merged.Tools = %v, want [read]", merged.Tools)
	}
}

func TestMergeRoleConfigs_EmptyToolsInheritsFromBase(t *testing.T) {
	base := makeBaseConfig()
	project := &AgentConfig{
		Name:           "engineer",
		ExplicitFields: map[string]bool{},
		// Tools empty — should inherit from base
	}
	project.Context.RoleContent = ""

	merged, err := mergeRoleConfigs(base, project)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(merged.Tools) != len(base.Tools) {
		t.Errorf("merged.Tools = %v, want %v (inherited from base)", merged.Tools, base.Tools)
	}
}

func TestMergeRoleConfigs_ExtendsAndExplicitFieldsCleared(t *testing.T) {
	base := makeBaseConfig()
	project := &AgentConfig{
		Name:    "engineer",
		Extends: "engineer",
		ExplicitFields: map[string]bool{
			configFieldExtends: true,
		},
	}
	project.Context.RoleContent = ""

	merged, err := mergeRoleConfigs(base, project)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged.Extends != "" {
		t.Error("merged.Extends should be cleared after merge")
	}
	if merged.ExplicitFields != nil {
		t.Error("merged.ExplicitFields should be nil after merge")
	}
}

// ── loadAgentConfig with Extends: ─────────────────────────────────────────────

func TestLoadAgentConfig_ExtendsInheritance(t *testing.T) {
	// Set up a temporary directory tree that mimics a project with:
	//   roles/engineer.md         ← base role
	//   .ai/roles/engineer.md     ← project override with Extends: engineer
	dir := t.TempDir()

	baseRolePath := filepath.Join(dir, "roles", "engineer.md")
	projectRolePath := filepath.Join(dir, ".ai", "roles", "engineer.md")

	if err := os.MkdirAll(filepath.Dir(baseRolePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(projectRolePath), 0o755); err != nil {
		t.Fatal(err)
	}

	baseContent := `# Engineer
**Agent:** engineer
**Description:** Base engineer
**Model:** gpt-4o-mini
**Tier:** minimal
**Tools:** read, write, bash

---

# Overview
Base overview.

# Responsibilities
Base responsibilities.
`
	projectContent := `# Engineer (Project Override)
**Agent:** engineer
**Description:** Project engineer
**Extends:** engineer
**Model:** gpt-4o

---

# Overview
Project overview.
`

	if err := os.WriteFile(baseRolePath, []byte(baseContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectRolePath, []byte(projectContent), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := &AgentServer{rootDir: dir}
	cfg, err := srv.loadAgentConfig("engineer", dir)
	if err != nil {
		t.Fatalf("loadAgentConfig with Extends: returned error: %v", err)
	}

	// Model should be overridden by project.
	if cfg.Model != "gpt-4o" {
		t.Errorf("cfg.Model = %q, want gpt-4o", cfg.Model)
	}
	// Tier should be locked to base.
	if cfg.Tier != "minimal" {
		t.Errorf("cfg.Tier = %q, want minimal (locked to base)", cfg.Tier)
	}
	// Tools should be inherited from base (project did not set them).
	if len(cfg.Tools) == 0 {
		t.Error("cfg.Tools should be inherited from base when not set in project")
	}
	// Role content: project overview replaces base overview.
	if !containsString(cfg.Context.RoleContent, "Project overview.") {
		t.Error("cfg.Context.RoleContent should contain project overview")
	}
	if containsString(cfg.Context.RoleContent, "Base overview.") {
		t.Error("cfg.Context.RoleContent should not contain base overview (was overridden)")
	}
	// Responsibilities should be inherited from base.
	if !containsString(cfg.Context.RoleContent, "Base responsibilities.") {
		t.Error("cfg.Context.RoleContent should inherit base Responsibilities section")
	}
	// Extends should be cleared.
	if cfg.Extends != "" {
		t.Error("cfg.Extends should be cleared after merge")
	}
}

func TestLoadAgentConfig_TierInProjectFileReturnsError(t *testing.T) {
	dir := t.TempDir()

	baseRolePath := filepath.Join(dir, "roles", "engineer.md")
	projectRolePath := filepath.Join(dir, ".ai", "roles", "engineer.md")

	if err := os.MkdirAll(filepath.Dir(baseRolePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(projectRolePath), 0o755); err != nil {
		t.Fatal(err)
	}

	baseContent := `# Engineer
**Agent:** engineer
**Description:** Base engineer
**Tier:** minimal
**Model:** gpt-4o-mini

---

# Overview
Base.
`
	// Project file sets Tier: — should be rejected.
	projectContent := `# Engineer (Project Override)
**Agent:** engineer
**Description:** Project engineer
**Extends:** engineer
**Tier:** standard

---

# Overview
Project.
`

	if err := os.WriteFile(baseRolePath, []byte(baseContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectRolePath, []byte(projectContent), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := &AgentServer{rootDir: dir}
	_, err := srv.loadAgentConfig("engineer", dir)
	if err == nil {
		t.Fatal("expected error when project role sets Tier: with Extends:, got nil")
	}
}

func TestLoadAgentConfig_ChainedExtendsReturnsError(t *testing.T) {
	dir := t.TempDir()

	// Chain: project → base1 → base2 (not allowed: base1 has Extends:)
	baseRolePath := filepath.Join(dir, "roles", "engineer.md")
	projectRolePath := filepath.Join(dir, ".ai", "roles", "engineer.md")

	if err := os.MkdirAll(filepath.Dir(baseRolePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(projectRolePath), 0o755); err != nil {
		t.Fatal(err)
	}

	// Base role itself has Extends: — not allowed.
	baseContent := `# Engineer
**Agent:** engineer
**Description:** Base engineer
**Extends:** something-else
**Tier:** minimal
**Model:** gpt-4o-mini

---

# Overview
Base.
`
	projectContent := `# Engineer (Project Override)
**Agent:** engineer
**Description:** Project engineer
**Extends:** engineer

---

# Overview
Project.
`

	if err := os.WriteFile(baseRolePath, []byte(baseContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectRolePath, []byte(projectContent), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := &AgentServer{rootDir: dir}
	_, err := srv.loadAgentConfig("engineer", dir)
	if err == nil {
		t.Fatal("expected error for chained Extends: (base role has Extends:), got nil")
	}
}

func TestLoadAgentConfig_WithoutExtendsUnchanged(t *testing.T) {
	// A project override WITHOUT Extends: should behave as before (full substitution).
	dir := t.TempDir()

	baseRolePath := filepath.Join(dir, "roles", "engineer.md")
	projectRolePath := filepath.Join(dir, ".ai", "roles", "engineer.md")

	if err := os.MkdirAll(filepath.Dir(baseRolePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(projectRolePath), 0o755); err != nil {
		t.Fatal(err)
	}

	// Base role — should be ignored because project has no Extends:.
	baseContent := `# Engineer
**Agent:** engineer
**Description:** Base engineer
**Tier:** minimal
**Model:** gpt-4o-mini
**Tools:** read, write, bash

---

# Overview
Base overview.
`
	// Project file with no Extends: — should fully replace base.
	projectContent := `# Engineer (Full Project Override)
**Agent:** engineer
**Description:** Full project engineer
**Tier:** standard
**Model:** claude-opus-4-5
**Tools:** read

---

# Overview
Full project override.
`

	if err := os.WriteFile(baseRolePath, []byte(baseContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectRolePath, []byte(projectContent), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := &AgentServer{rootDir: dir}
	cfg, err := srv.loadAgentConfig("engineer", dir)
	if err != nil {
		t.Fatalf("loadAgentConfig without Extends: returned error: %v", err)
	}

	// Should use project values entirely (full substitution, Tier 3a behaviour).
	if cfg.Model != "claude-opus-4-5" {
		t.Errorf("cfg.Model = %q, want claude-opus-4-5", cfg.Model)
	}
	if cfg.Tier != "standard" {
		t.Errorf("cfg.Tier = %q, want standard", cfg.Tier)
	}
	if len(cfg.Tools) != 1 || cfg.Tools[0] != "read" {
		t.Errorf("cfg.Tools = %v, want [read]", cfg.Tools)
	}
	if !containsString(cfg.Context.RoleContent, "Full project override.") {
		t.Error("project content should fully replace base content when no Extends:")
	}
	// Base content should NOT be present.
	if containsString(cfg.Context.RoleContent, "Base overview.") {
		t.Error("base content should not appear when project has no Extends:")
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func containsString(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && (s == sub ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
