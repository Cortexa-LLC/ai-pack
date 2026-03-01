package server

// skill_composer.go — Phase 1 of OCP skill composition (ADR 004)
//
// Provides:
//   - SkillConfig struct
//   - parseSkillMarkdown(data []byte, name string) (*SkillConfig, error)
//   - resolveSkillPath(name, projectRoot string) string
//   - WorkflowGateConfig struct
//   - parseWorkflowConfig(projectRoot string) (*WorkflowGateConfig, error)
//   - composeSkills(config *AgentConfig, projectRoot string) error
//
// Called from loadAgentConfig in server_helpers.go immediately after the role
// file is parsed.

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// SkillConfig holds the parsed header + prompt fragment of a single skill file.
type SkillConfig struct {
	Name           string
	Version        string
	InjectAt       string   // "preamble" | "role_context" | "postamble"
	Slot           int      // sort order within InjectAt group
	Tools          []string // additional tools contributed by this skill
	Gates          []string // gate names contributed by this skill
	MaxExtraTokens int64    // extra token budget added to role budget
	Optional       bool     // if false, absence is a hard failure
	PromptFragment string   // markdown body after the separator
}

// WorkflowGateConfig holds per-role gate disable lists parsed from .ai/workflow.md.
type WorkflowGateConfig struct {
	// Disable maps role name (or "*") → set of gate names to disable.
	Disable map[string]map[string]bool
}

// parseSkillMarkdown parses a skill file (data) whose logical name is name.
//
// Expected format:
//
//	# Title
//	<!-- skills/foo.skill.md -->   ← optional HTML comment, ignored
//	**Version:** 1.0
//	**InjectAt:** role_context
//	**Slot:** 50
//	**Tools:** read, write
//	**Gates:** tdd-enforcement
//	**MaxExtraTokens:** 0
//	**Optional:** true
//
//	---
//
//	<prompt fragment markdown>
func parseSkillMarkdown(data []byte, name string) (*SkillConfig, error) {
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	skill := &SkillConfig{
		Name:     name,
		InjectAt: "role_context", // default per spec
		Slot:     50,             // default per spec
		Optional: true,           // safe default
	}

	separatorIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			separatorIdx = i
			break
		}
	}
	if separatorIdx == -1 {
		return &SkillConfig{Name: name, Optional: true}, fmt.Errorf("skill %q: missing --- separator", name)
	}

	// Parse header lines (before the separator)
	for i := 0; i < separatorIdx; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		// Skip title line (# ...) and HTML comments <!-- ... -->
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "<!--") {
			continue
		}
		// Expect **Key:** value
		if !strings.HasPrefix(line, markdownFieldStart) || !strings.Contains(line, markdownFieldEnd) {
			continue
		}
		parts := strings.SplitN(line, markdownFieldEnd, 2)
		if len(parts) != 2 {
			continue
		}
		field := strings.TrimPrefix(parts[0], markdownFieldStart)
		value := strings.TrimSpace(parts[1])

		switch field {
		case "Version":
			skill.Version = value
		case "InjectAt":
			v := strings.ToLower(value)
			switch v {
			case "preamble", "role_context", "postamble":
				skill.InjectAt = v
			default:
				// Unknown → default to role_context per spec
				skill.InjectAt = "role_context"
			}
		case "Slot":
			if n, err := strconv.Atoi(value); err == nil {
				skill.Slot = n
			}
		case "Tools":
			if value != "" && value != "(none)" {
				skill.Tools = splitAndTrim(value)
			}
		case "Gates":
			if value != "" && value != "(none)" {
				skill.Gates = splitAndTrim(value)
			}
		case "MaxExtraTokens":
			if n, err := strconv.ParseInt(value, 10, 64); err == nil {
				skill.MaxExtraTokens = n
			}
		case "Optional":
			skill.Optional = strings.EqualFold(value, "true")
		}
	}

	// Everything after the separator is the prompt fragment
	if separatorIdx+1 < len(lines) {
		fragment := strings.Join(lines[separatorIdx+1:], "\n")
		skill.PromptFragment = strings.TrimSpace(fragment)
	}

	return skill, nil
}

// resolveSkillPath returns the first existing path for the given skill name.
// Search order (per spec):
//  1. projectRoot/.ai/skills/<name>.skill.md        — project override
//  2. projectRoot/.ai-pack/skills/<name>.skill.md   — submodule install
//  3. projectRoot/skills/<name>.skill.md            — framework default
//  4. ../skills/<name>.skill.md                     — dev: parent dir
//
// Returns "" if the skill is not found anywhere.
func resolveSkillPath(name, projectRoot string) string {
	filename := name + ".skill.md"
	candidates := []string{
		filepath.Join(projectRoot, ".ai", "skills", filename),
		filepath.Join(projectRoot, ".ai-pack", "skills", filename),
		filepath.Join(projectRoot, "skills", filename),
		filepath.Join(projectRoot, "..", "skills", filename),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// parseWorkflowConfig reads .ai/workflow.md and extracts gate disable lists per role.
//
// Supported formats:
//
//  1. Markdown table:
//
//     | Role | Action | Gates |
//     |------|--------|-------|
//     | engineer | disable | tdd-enforcement |
//
//  2. YAML front-matter (between leading --- ... ---):
//
//     skill_gates:
//     engineer:
//     disable:
//     - tdd-enforcement
//
// Returns an empty (non-nil) config if the file is absent or unparseable.
func parseWorkflowConfig(projectRoot string) (*WorkflowGateConfig, error) {
	cfg := &WorkflowGateConfig{
		Disable: make(map[string]map[string]bool),
	}

	wfPath := filepath.Join(projectRoot, ".ai", "workflow.md")
	data, err := os.ReadFile(wfPath)
	if err != nil {
		// Absent workflow.md is normal
		return cfg, nil //nolint:nilerr
	}

	text := string(data)

	// ── Try YAML front-matter ────────────────────────────────────────────────
	if strings.HasPrefix(strings.TrimSpace(text), "---") {
		if parsed := parseWorkflowYAML(text, cfg); parsed {
			return cfg, nil
		}
	}

	// ── Try Markdown table ───────────────────────────────────────────────────
	parseWorkflowTable(text, cfg)
	return cfg, nil
}

// parseWorkflowYAML handles the YAML front-matter form of workflow.md.
// It does minimal key-by-key parsing without importing a YAML library.
//
// Expected structure (indentation-sensitive):
//
//	---
//	skill_gates:
//	  engineer:
//	    disable:
//	      - tdd-enforcement
//	---
func parseWorkflowYAML(text string, cfg *WorkflowGateConfig) bool {
	// Extract content between first pair of ---
	lines := strings.Split(text, "\n")
	start, end := -1, -1
	for i, l := range lines {
		if strings.TrimSpace(l) == "---" {
			if start == -1 {
				start = i
			} else {
				end = i
				break
			}
		}
	}
	if start == -1 || end == -1 {
		return false
	}

	yamlLines := lines[start+1 : end]

	// State machine: look for skill_gates → role → disable → entries
	inSkillGates := false
	currentRole := ""
	inDisable := false
	found := false

	for _, raw := range yamlLines {
		line := strings.TrimRight(raw, " \t\r")
		trimmed := strings.TrimLeft(line, " \t")
		indent := len(line) - len(trimmed)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		switch {
		case trimmed == "skill_gates:":
			inSkillGates = true
			currentRole = ""
			inDisable = false

		case inSkillGates && indent == 2 && strings.HasSuffix(trimmed, ":"):
			// role name
			currentRole = strings.TrimSuffix(trimmed, ":")
			inDisable = false
			found = true

		case inSkillGates && currentRole != "" && indent == 4 && trimmed == "disable:":
			inDisable = true

		case inSkillGates && currentRole != "" && inDisable && indent == 6 && strings.HasPrefix(trimmed, "- "):
			gate := strings.TrimPrefix(trimmed, "- ")
			gate = strings.TrimSpace(gate)
			if gate != "" {
				if cfg.Disable[currentRole] == nil {
					cfg.Disable[currentRole] = make(map[string]bool)
				}
				cfg.Disable[currentRole][gate] = true
			}

		default:
			if indent == 0 && trimmed != "skill_gates:" {
				inSkillGates = false
			}
		}
	}

	return found
}

// parseWorkflowTable handles the Markdown table form of workflow.md.
//
// | Role     | Action  | Gates              |
// |----------|---------|--------------------|
// | engineer | disable | tdd-enforcement    |
// | *        | disable | (none)             |
func parseWorkflowTable(text string, cfg *WorkflowGateConfig) {
	scanner := bufio.NewScanner(bytes.NewReader([]byte(text)))
	inTable := false
	headerFound := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if !strings.HasPrefix(line, "|") {
			if inTable {
				inTable = false
			}
			continue
		}

		cols := splitTableRow(line)
		if len(cols) < 3 {
			continue
		}

		// Detect header row
		if strings.EqualFold(strings.TrimSpace(cols[0]), "role") &&
			strings.EqualFold(strings.TrimSpace(cols[1]), "action") {
			inTable = true
			headerFound = true
			continue
		}
		// Skip separator row (|---|---|---|)
		if strings.Contains(cols[0], "-") {
			continue
		}

		if !inTable || !headerFound {
			continue
		}

		role := strings.TrimSpace(cols[0])
		action := strings.ToLower(strings.TrimSpace(cols[1]))
		gatesRaw := strings.TrimSpace(cols[2])

		if role == "" || action != "disable" {
			continue
		}
		if gatesRaw == "" || gatesRaw == "(none)" {
			continue
		}

		gates := splitAndTrim(gatesRaw)
		for _, g := range gates {
			if cfg.Disable[role] == nil {
				cfg.Disable[role] = make(map[string]bool)
			}
			cfg.Disable[role][g] = true
		}
	}
}

// splitTableRow splits a Markdown table row on "|" and returns non-empty cells.
func splitTableRow(line string) []string {
	parts := strings.Split(line, "|")
	var cells []string
	for _, p := range parts {
		cells = append(cells, p)
	}
	// Trim leading/trailing empty strings from outer pipes
	for len(cells) > 0 && strings.TrimSpace(cells[0]) == "" {
		cells = cells[1:]
	}
	for len(cells) > 0 && strings.TrimSpace(cells[len(cells)-1]) == "" {
		cells = cells[:len(cells)-1]
	}
	return cells
}

// composeSkills resolves, parses, and merges all skills declared by the role into
// config. It mutates config in place and returns nil on success.
//
// Steps (per skill-composition.md):
//  1. Collect skill names from config.Skills (default: ["general"])
//  2. Resolve + parse each skill file
//  3. Apply workflow gate overrides for optional skills
//  4. Merge tools, gates, budget
//  5. Assemble system prompt (preamble | role body | role_context | postamble)
func composeSkills(config *AgentConfig, projectRoot string) error {
	skillNames := config.Skills
	if len(skillNames) == 0 {
		skillNames = []string{"general"}
	}

	// ── Step 2: resolve + parse ──────────────────────────────────────────────
	workflowCfg, _ := parseWorkflowConfig(projectRoot)

	var skills []*SkillConfig
	for _, name := range skillNames {
		path := resolveSkillPath(name, projectRoot)
		if path == "" {
			monitoring.Logger.Warn("skill_not_found",
				"role", config.Name,
				"skill", name,
			)
			// "general" skill missing is a soft failure (no file yet = no augmentation)
			if name == "general" {
				continue
			}
			// Non-general skill is required unless we know otherwise; fail spawn.
			return fmt.Errorf("required skill %q not found (searched in %q)", name, projectRoot)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			monitoring.Logger.Warn("skill_read_error",
				"role", config.Name,
				"skill", name,
				"error", err,
			)
			continue
		}

		skill, err := parseSkillMarkdown(data, name)
		if err != nil {
			monitoring.Logger.Warn("skill_parse_error",
				"role", config.Name,
				"skill", name,
				"error", err,
			)
			if skill != nil && !skill.Optional {
				return fmt.Errorf("required skill %q malformed: %w", name, err)
			}
			continue
		}

		skills = append(skills, skill)
	}

	// ── Step 3: apply workflow gate overrides ────────────────────────────────
	roleName := config.Name
	for _, skill := range skills {
		if !skill.Optional {
			continue // non-optional gates are never removed
		}
		// Merge wildcard and role-specific disables
		disabledGates := make(map[string]bool)
		for g := range workflowCfg.Disable["*"] {
			disabledGates[g] = true
		}
		for g := range workflowCfg.Disable[roleName] {
			disabledGates[g] = true
		}
		if len(disabledGates) == 0 {
			continue
		}
		filtered := make([]string, 0, len(skill.Gates))
		for _, g := range skill.Gates {
			if !disabledGates[g] {
				filtered = append(filtered, g)
			}
		}
		skill.Gates = filtered
	}

	// ── Step 4: merge tools + gates + budget ─────────────────────────────────
	toolSet := make(map[string]bool)
	for _, t := range config.Tools {
		toolSet[strings.ToLower(t)] = true
	}
	gateSet := make(map[string]bool)
	for _, g := range config.Context.Gates {
		gateSet[g] = true
	}
	var extraTokens int64

	for _, skill := range skills {
		for _, t := range skill.Tools {
			toolSet[strings.ToLower(t)] = true
		}
		for _, g := range skill.Gates {
			gateSet[g] = true
		}
		extraTokens += skill.MaxExtraTokens
	}

	// Build sorted tool list
	tools := make([]string, 0, len(toolSet))
	for t := range toolSet {
		tools = append(tools, t)
	}
	sort.Strings(tools)
	config.Tools = tools

	// Build sorted gate list
	gates := make([]string, 0, len(gateSet))
	for g := range gateSet {
		gates = append(gates, g)
	}
	sort.Strings(gates)
	config.Context.Gates = gates

	// Budget: role budget + skill extras, capped at 0 (unlimited) if role is 0
	if config.Delegation.MaxBudgetTokens > 0 && extraTokens > 0 {
		config.Delegation.MaxBudgetTokens += int(extraTokens)
	}

	// ── Step 5: assemble system prompt ───────────────────────────────────────
	// Sort skills by (InjectAt priority, Slot)
	sort.Slice(skills, func(i, j int) bool {
		si, sj := skills[i], skills[j]
		if si.InjectAt != sj.InjectAt {
			return injectAtOrder(si.InjectAt) < injectAtOrder(sj.InjectAt)
		}
		return si.Slot < sj.Slot
	})

	var preamble, roleContext, postamble []string
	for _, skill := range skills {
		if skill.PromptFragment == "" {
			continue
		}
		switch skill.InjectAt {
		case "preamble":
			preamble = append(preamble, skill.PromptFragment)
		case "postamble":
			postamble = append(postamble, skill.PromptFragment)
		default: // "role_context" or unknown
			roleContext = append(roleContext, skill.PromptFragment)
		}
	}

	// Existing role body (already parsed into RoleContent by loadAgentConfig)
	roleBody := config.Context.RoleContent

	// Assemble sections
	var sections []string
	if len(preamble) > 0 {
		sections = append(sections, strings.Join(preamble, "\n\n---\n\n"))
	}
	if roleBody != "" {
		sections = append(sections, roleBody)
	}
	if len(roleContext) > 0 {
		sections = append(sections, strings.Join(roleContext, "\n\n---\n\n"))
	}
	if len(postamble) > 0 {
		sections = append(sections, strings.Join(postamble, "\n\n---\n\n"))
	}

	config.Context.RoleContent = strings.Join(sections, "\n\n")

	// ── Record which skills were loaded ──────────────────────────────────────
	loaded := make([]string, 0, len(skills))
	for _, sk := range skills {
		loaded = append(loaded, sk.Name)
	}
	config.SkillsLoaded = loaded

	monitoring.Logger.Info("skills_composed",
		"role", config.Name,
		"skills", loaded,
	)

	return nil
}

// injectAtOrder returns a sort key for InjectAt values so that
// "preamble" < "role_context" < "postamble".
func injectAtOrder(injectAt string) int {
	switch injectAt {
	case "preamble":
		return 0
	case "role_context":
		return 1
	case "postamble":
		return 2
	default:
		return 1
	}
}

// splitAndTrim splits a comma-separated value string and trims each element.
// It reuses the same splitting logic used throughout the server package.
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
