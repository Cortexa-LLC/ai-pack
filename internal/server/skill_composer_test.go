package server

import (
	"os"
	"path/filepath"
	"testing"
)

// ── parseSkillMarkdown ────────────────────────────────────────────────────────

func TestParseSkillMarkdown_ValidFull(t *testing.T) {
	data := []byte(`# Test Skill
<!-- skills/test.skill.md -->

**Version:** 1.2
**InjectAt:** preamble
**Slot:** 10
**Tools:** read, write
**Gates:** tdd-enforcement, code-review
**MaxExtraTokens:** 5000
**Optional:** false

---

This is the prompt fragment.
Multiple lines.
`)
	skill, err := parseSkillMarkdown(data, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skill.Name != "test" {
		t.Errorf("Name = %q, want %q", skill.Name, "test")
	}
	if skill.Version != "1.2" {
		t.Errorf("Version = %q, want %q", skill.Version, "1.2")
	}
	if skill.InjectAt != "preamble" {
		t.Errorf("InjectAt = %q, want %q", skill.InjectAt, "preamble")
	}
	if skill.Slot != 10 {
		t.Errorf("Slot = %d, want %d", skill.Slot, 10)
	}
	if len(skill.Tools) != 2 || skill.Tools[0] != "read" || skill.Tools[1] != "write" {
		t.Errorf("Tools = %v, want [read write]", skill.Tools)
	}
	if len(skill.Gates) != 2 || skill.Gates[0] != "tdd-enforcement" {
		t.Errorf("Gates = %v, want [tdd-enforcement code-review]", skill.Gates)
	}
	if skill.MaxExtraTokens != 5000 {
		t.Errorf("MaxExtraTokens = %d, want 5000", skill.MaxExtraTokens)
	}
	if skill.Optional {
		t.Errorf("Optional = true, want false")
	}
	if skill.PromptFragment == "" {
		t.Error("PromptFragment is empty, want content")
	}
}

func TestParseSkillMarkdown_Defaults(t *testing.T) {
	data := []byte(`# Minimal Skill

**Version:** 1.0

---

Fragment here.
`)
	skill, err := parseSkillMarkdown(data, "minimal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skill.InjectAt != "role_context" {
		t.Errorf("InjectAt = %q, want %q", skill.InjectAt, "role_context")
	}
	if skill.Slot != 50 {
		t.Errorf("Slot = %d, want 50", skill.Slot)
	}
	if !skill.Optional {
		t.Error("Optional = false by default, want true")
	}
	if len(skill.Tools) != 0 {
		t.Errorf("Tools = %v, want empty", skill.Tools)
	}
	if len(skill.Gates) != 0 {
		t.Errorf("Gates = %v, want empty", skill.Gates)
	}
}

func TestParseSkillMarkdown_NoSeparator(t *testing.T) {
	data := []byte(`# Broken Skill
**Version:** 1.0
No separator here.
`)
	_, err := parseSkillMarkdown(data, "broken")
	if err == nil {
		t.Error("expected error for missing separator, got nil")
	}
}

func TestParseSkillMarkdown_NoneTokens(t *testing.T) {
	data := []byte(`# Skill
**Tools:** (none)
**Gates:** (none)
---
Fragment.
`)
	skill, err := parseSkillMarkdown(data, "nones")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skill.Tools) != 0 {
		t.Errorf("Tools = %v, want empty for (none)", skill.Tools)
	}
	if len(skill.Gates) != 0 {
		t.Errorf("Gates = %v, want empty for (none)", skill.Gates)
	}
}

func TestParseSkillMarkdown_GeneralSkillFile(t *testing.T) {
	// Test with the actual general.skill.md if present
	data, err := os.ReadFile(filepath.Join("..", "..", "skills", "general.skill.md"))
	if err != nil {
		t.Skip("skills/general.skill.md not found, skipping integration test")
	}
	skill, err := parseSkillMarkdown(data, "general")
	if err != nil {
		t.Fatalf("unexpected error parsing general.skill.md: %v", err)
	}
	if skill.InjectAt != "preamble" {
		t.Errorf("general skill InjectAt = %q, want %q", skill.InjectAt, "preamble")
	}
	if skill.PromptFragment == "" {
		t.Error("general skill has empty prompt fragment")
	}
}

// ── resolveSkillPath ─────────────────────────────────────────────────────────

func TestResolveSkillPath_FrameworkDefault(t *testing.T) {
	// Use the actual project root which has skills/
	root := filepath.Join("..", "..")
	path := resolveSkillPath("general", root)
	if path == "" {
		t.Skip("skills/general.skill.md not found in project root")
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("resolved path does not exist: %s", path)
	}
}

func TestResolveSkillPath_NotFound(t *testing.T) {
	path := resolveSkillPath("nonexistent-xyz-skill", "/tmp")
	if path != "" {
		t.Errorf("expected empty path for nonexistent skill, got %q", path)
	}
}

func TestResolveSkillPath_ProjectOverrideTakesPrecedence(t *testing.T) {
	// Create a temp project with a .ai/skills override
	tmp := t.TempDir()
	// Create framework skill
	frameworkDir := filepath.Join(tmp, "skills")
	if err := os.MkdirAll(frameworkDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(frameworkDir, "foo.skill.md"), []byte("# Foo\n---\nframework"), 0644); err != nil {
		t.Fatal(err)
	}
	// Create project override
	overrideDir := filepath.Join(tmp, ".ai", "skills")
	if err := os.MkdirAll(overrideDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(overrideDir, "foo.skill.md"), []byte("# Foo\n---\noverride"), 0644); err != nil {
		t.Fatal(err)
	}

	path := resolveSkillPath("foo", tmp)
	expected := filepath.Join(overrideDir, "foo.skill.md")
	if path != expected {
		t.Errorf("resolveSkillPath returned %q, want %q (override path)", path, expected)
	}
}

// ── parseWorkflowConfig ───────────────────────────────────────────────────────

func TestParseWorkflowConfig_Absent(t *testing.T) {
	cfg, err := parseWorkflowConfig("/nonexistent-dir-xyz")
	if err != nil {
		t.Errorf("expected nil error for absent workflow.md, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.Disable) != 0 {
		t.Errorf("expected empty disable map, got %v", cfg.Disable)
	}
}

func TestParseWorkflowConfig_MarkdownTable(t *testing.T) {
	tmp := t.TempDir()
	aiDir := filepath.Join(tmp, ".ai")
	if err := os.MkdirAll(aiDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `# Workflow Config

| Role | Action | Gates |
|------|--------|-------|
| engineer | disable | tdd-enforcement |
| * | disable | arch-review |
`
	if err := os.WriteFile(filepath.Join(aiDir, "workflow.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := parseWorkflowConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Disable["engineer"]["tdd-enforcement"] {
		t.Error("expected engineer/tdd-enforcement to be disabled")
	}
	if !cfg.Disable["*"]["arch-review"] {
		t.Error("expected */arch-review to be disabled")
	}
}

func TestParseWorkflowConfig_YAMLFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	aiDir := filepath.Join(tmp, ".ai")
	if err := os.MkdirAll(aiDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `---
skill_gates:
  engineer:
    disable:
      - tdd-enforcement
      - code-quality-review
  *:
    disable:
      - arch-review
---

# Workflow

Some notes.
`
	if err := os.WriteFile(filepath.Join(aiDir, "workflow.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := parseWorkflowConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Disable["engineer"]["tdd-enforcement"] {
		t.Error("expected engineer/tdd-enforcement to be disabled via YAML")
	}
	if !cfg.Disable["engineer"]["code-quality-review"] {
		t.Error("expected engineer/code-quality-review to be disabled via YAML")
	}
}

// ── composeSkills integration ─────────────────────────────────────────────────

func TestComposeSkills_GeneralSkillPresent(t *testing.T) {
	// Use actual project skills dir
	root := filepath.Join("..", "..")
	if _, err := os.Stat(filepath.Join(root, "skills", "general.skill.md")); err != nil {
		t.Skip("skills/general.skill.md not present")
	}

	config := &AgentConfig{
		Name:  "test-role",
		Tools: []string{"bash"},
	}
	config.Context.Gates = []string{"existing-gate"}

	err := composeSkills(config, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Skills should be loaded
	if len(config.SkillsLoaded) == 0 {
		t.Error("SkillsLoaded is empty after composing general skill")
	}

	// Tools from general.skill.md should be merged
	toolMap := make(map[string]bool)
	for _, t := range config.Tools {
		toolMap[t] = true
	}
	if !toolMap["bash"] {
		t.Error("existing tool 'bash' was dropped during composition")
	}

	// Existing gates should be preserved
	gateMap := make(map[string]bool)
	for _, g := range config.Context.Gates {
		gateMap[g] = true
	}
	if !gateMap["existing-gate"] {
		t.Error("existing gate was dropped during composition")
	}

	// Role content should have a fragment appended
	if config.Context.RoleContent == "" {
		t.Error("RoleContent is empty after composition")
	}
}

func TestComposeSkills_MissingRequiredSkill(t *testing.T) {
	config := &AgentConfig{
		Name:   "test-role",
		Skills: []string{"nonexistent-skill-xyz"},
	}

	err := composeSkills(config, "/tmp")
	if err == nil {
		t.Error("expected error for missing required skill, got nil")
	}
}

func TestComposeSkills_SkillsMergeTools(t *testing.T) {
	// Create a temporary skill file
	tmp := t.TempDir()
	skillsDir := filepath.Join(tmp, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}
	skillContent := `# Test Skill
**Version:** 1.0
**InjectAt:** role_context
**Slot:** 50
**Tools:** mcp__kg__search_nodes, mcp__kg__open_nodes
**Gates:** kg-gate
**MaxExtraTokens:** 1000
**Optional:** true
---

## KG Access

Query the knowledge graph.
`
	if err := os.WriteFile(filepath.Join(skillsDir, "kg.skill.md"), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	config := &AgentConfig{
		Name:   "engineer",
		Skills: []string{"kg"},
		Tools:  []string{"read", "write"},
	}
	config.Context.Gates = []string{}
	config.Context.RoleContent = "Original role body."

	err := composeSkills(config, tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Tools should be merged
	toolMap := make(map[string]bool)
	for _, t := range config.Tools {
		toolMap[t] = true
	}
	if !toolMap["read"] {
		t.Error("tool 'read' was dropped")
	}
	if !toolMap["mcp__kg__search_nodes"] {
		t.Error("tool 'mcp__kg__search_nodes' from skill was not added")
	}

	// Gates should be merged
	gateMap := make(map[string]bool)
	for _, g := range config.Context.Gates {
		gateMap[g] = true
	}
	if !gateMap["kg-gate"] {
		t.Error("gate 'kg-gate' from skill was not added")
	}

	// Role content should include both original and skill fragment
	if config.Context.RoleContent == "" {
		t.Error("RoleContent is empty")
	}

	// SkillsLoaded should be recorded
	if len(config.SkillsLoaded) != 1 || config.SkillsLoaded[0] != "kg" {
		t.Errorf("SkillsLoaded = %v, want [kg]", config.SkillsLoaded)
	}
}

func TestComposeSkills_PreambleInjectAt(t *testing.T) {
	tmp := t.TempDir()
	skillsDir := filepath.Join(tmp, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	preamble := `# Pre
**InjectAt:** preamble
**Slot:** 10
---
PREAMBLE TEXT
`
	postamble := `# Post
**InjectAt:** postamble
**Slot:** 90
---
POSTAMBLE TEXT
`
	if err := os.WriteFile(filepath.Join(skillsDir, "pre.skill.md"), []byte(preamble), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "post.skill.md"), []byte(postamble), 0644); err != nil {
		t.Fatal(err)
	}

	config := &AgentConfig{
		Name:   "test",
		Skills: []string{"pre", "post"},
	}
	config.Context.RoleContent = "ROLE BODY"

	if err := composeSkills(config, tmp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := config.Context.RoleContent
	preIdx := indexOf(content, "PREAMBLE TEXT")
	roleIdx := indexOf(content, "ROLE BODY")
	postIdx := indexOf(content, "POSTAMBLE TEXT")

	if preIdx == -1 || roleIdx == -1 || postIdx == -1 {
		t.Fatalf("one or more sections missing from RoleContent:\n%s", content)
	}
	if !(preIdx < roleIdx && roleIdx < postIdx) {
		t.Errorf("wrong order: preamble=%d role=%d postamble=%d\nContent:\n%s",
			preIdx, roleIdx, postIdx, content)
	}
}

func TestComposeSkills_WorkflowGateDisable(t *testing.T) {
	tmp := t.TempDir()
	skillsDir := filepath.Join(tmp, "skills")
	aiDir := filepath.Join(tmp, ".ai")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(aiDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Skill with a gate
	skill := `# Gated Skill
**InjectAt:** role_context
**Slot:** 50
**Gates:** tdd-enforcement
**Optional:** true
---
Fragment.
`
	if err := os.WriteFile(filepath.Join(skillsDir, "gated.skill.md"), []byte(skill), 0644); err != nil {
		t.Fatal(err)
	}

	// Workflow that disables the gate for this role
	workflow := `# Workflow
| Role | Action | Gates |
|------|--------|-------|
| engineer | disable | tdd-enforcement |
`
	if err := os.WriteFile(filepath.Join(aiDir, "workflow.md"), []byte(workflow), 0644); err != nil {
		t.Fatal(err)
	}

	config := &AgentConfig{
		Name:   "engineer",
		Skills: []string{"gated"},
	}
	config.Context.Gates = []string{}

	if err := composeSkills(config, tmp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, g := range config.Context.Gates {
		if g == "tdd-enforcement" {
			t.Error("tdd-enforcement gate should have been disabled by workflow.md")
		}
	}
}

func TestComposeSkills_BudgetAccumulates(t *testing.T) {
	tmp := t.TempDir()
	skillsDir := filepath.Join(tmp, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	skill := `# Budget Skill
**MaxExtraTokens:** 3000
**Optional:** true
---
Fragment.
`
	if err := os.WriteFile(filepath.Join(skillsDir, "budget.skill.md"), []byte(skill), 0644); err != nil {
		t.Fatal(err)
	}

	config := &AgentConfig{
		Name:   "test",
		Skills: []string{"budget"},
	}
	config.Delegation.MaxBudgetTokens = 10000

	if err := composeSkills(config, tmp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.Delegation.MaxBudgetTokens != 13000 {
		t.Errorf("MaxBudgetTokens = %d, want 13000", config.Delegation.MaxBudgetTokens)
	}
}

// indexOf returns the index of substr in s, or -1 if not found.
func indexOf(s, substr string) int {
	for i := range s {
		if len(s)-i >= len(substr) && s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ── splitAndTrim ─────────────────────────────────────────────────────────────

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"read, write, bash", []string{"read", "write", "bash"}},
		{"single", []string{"single"}},
		{"  a , b , c  ", []string{"a", "b", "c"}},
		{"", nil},
	}
	for _, tt := range tests {
		got := splitAndTrim(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitAndTrim(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitAndTrim(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}
