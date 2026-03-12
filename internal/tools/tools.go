package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/claude"
	"github.com/cortexa-llc/ai-pack/internal/streaming"
)

const (
	errReadingFile = "Error reading file: %v"
	errWritingFile = "Error writing file: %v"
)

// DefineTools returns all native tools in provider-agnostic format.
func DefineTools() []streaming.Tool {
	return []streaming.Tool{
		// File operations
		defineReadTool(),
		defineWriteTool(),
		defineEditTool(),
		defineMultiEditTool(),

		// Search operations
		defineGrepTool(),
		defineGlobTool(),

		// Shell operations
		defineBashTool(),

		// Task lifecycle
		defineTaskCompleteTool(),

		// Cross-project knowledge graph
		defineSearchKnowledgeInProjectTool(),

		// NOTE: Web operations disabled - not yet implemented
		// defineWebSearchTool(),
		// defineWebFetchTool(),
	}
}

// defineBashTool creates the Bash tool (matches Claude Code)
func defineBashTool() streaming.Tool {
	return streaming.Tool{
		Name:        "Bash",
		Description: "Execute bash commands in the working directory. Use this to run git, npm, tests, or any other shell commands.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The bash command to execute",
				},
			},
			"required": []string{"command"},
		},
	}
}

// defineReadTool creates the Read tool (matches Claude Code)
func defineReadTool() streaming.Tool {
	return streaming.Tool{
		Name:        "Read",
		Description: "Read the contents of a file.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to read (relative to working directory)",
				},
			},
			"required": []string{"file_path"},
		},
	}
}

// defineWriteTool creates the Write tool (matches Claude Code)
func defineWriteTool() streaming.Tool {
	return streaming.Tool{
		Name:        "Write",
		Description: "Create a new file with the given content. Will overwrite if file exists.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to write (relative to working directory)",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Content to write to the file",
				},
			},
			"required": []string{"file_path", "content"},
		},
	}
}

// defineEditTool creates the Edit tool (matches Claude Code)
func defineEditTool() streaming.Tool {
	return streaming.Tool{
		Name:        "Edit",
		Description: "Edit a file by replacing an exact string match. The old_string must match exactly (including whitespace and indentation).",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to edit (relative to working directory)",
				},
				"old_string": map[string]interface{}{
					"type":        "string",
					"description": "The exact string to replace in the file",
				},
				"new_string": map[string]interface{}{
					"type":        "string",
					"description": "The string to replace it with",
				},
			},
			"required": []string{"file_path", "old_string", "new_string"},
		},
	}
}

// defineMultiEditTool creates the MultiEdit tool for making multiple edits at once
func defineMultiEditTool() streaming.Tool {
	return streaming.Tool{
		Name:        "MultiEdit",
		Description: "Make multiple edits to a file in a single operation. Each edit replaces an exact string match.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to edit (relative to working directory)",
				},
				"edits": map[string]interface{}{
					"type":        "array",
					"description": "Array of edits to perform in order",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"old_string": map[string]interface{}{
								"type":        "string",
								"description": "The exact string to replace in the file",
							},
							"new_string": map[string]interface{}{
								"type":        "string",
								"description": "The string to replace it with",
							},
						},
						"required": []string{"old_string", "new_string"},
					},
				},
			},
			"required": []string{"file_path", "edits"},
		},
	}
}

// defineGrepTool creates the Grep tool (matches Claude Code)
func defineGrepTool() streaming.Tool {
	return streaming.Tool{
		Name:        "Grep",
		Description: "Search for patterns in files. Returns matching lines with file paths.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "The pattern to search for (regex supported)",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Optional path to search in (defaults to current directory)",
				},
			},
			"required": []string{"pattern"},
		},
	}
}

// defineGlobTool creates the Glob tool (matches Claude Code)
func defineGlobTool() streaming.Tool {
	return streaming.Tool{
		Name:        "Glob",
		Description: "Find files matching a glob pattern. Returns list of file paths.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Glob pattern to match files (e.g., **/*.go, src/**/*.ts)",
				},
			},
			"required": []string{"pattern"},
		},
	}
}

// defineWebSearchTool creates the WebSearch tool (matches Claude Code)
func defineWebSearchTool() streaming.Tool {
	return streaming.Tool{
		Name:        "WebSearch",
		Description: "Search the web for information. Returns search results with titles, URLs, and snippets.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The search query",
				},
			},
			"required": []string{"query"},
		},
	}
}

// defineWebFetchTool creates the WebFetch tool (matches Claude Code)
func defineWebFetchTool() streaming.Tool {
	return streaming.Tool{
		Name:        "WebFetch",
		Description: "Fetch and parse a web page. Returns the page content, optionally filtered by prompt.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "The URL to fetch",
				},
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "Optional prompt describing what to extract from the page",
				},
			},
			"required": []string{"url"},
		},
	}
}

// defineSearchKnowledgeInProjectTool creates the search_knowledge_in_project
// synthetic tool. It allows agents to query the knowledge graph of a different
// project registered with the AI-Pack server at runtime.
func defineSearchKnowledgeInProjectTool() streaming.Tool {
	return streaming.Tool{
		Name: "search_knowledge_in_project",
		Description: "Hybrid search for entities and observations in another project's knowledge graph. " +
			"Use this when working across multiple projects (e.g. patching a dependency) to " +
			"retrieve relevant functions, types, files, and topics from that project's KG. " +
			"The project must be registered with the AI-Pack server (i.e. it has a .kg/ directory). " +
			"Use short, specific terms (1–3 words); each whitespace-separated token is matched independently (OR logic).",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the other project's root directory.",
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The search query to match against entity names, types, and observation content.",
				},
			},
			"required": []string{"project_path", "query"},
		},
	}
}

// defineTaskCompleteTool creates the TaskComplete signal tool. Agents MUST call
// this to end the task. Text-only responses are never accepted as completion —
// the loop nudges the model to continue until this tool is called.
func defineTaskCompleteTool() streaming.Tool {
	return streaming.Tool{
		Name: "TaskComplete",
		Description: "Signal that the task is fully complete. " +
			"Call this tool when ALL work described in the task is done and verified. " +
			"Provide a clear summary of what was accomplished (files changed, decisions made). " +
			"This is the ONLY way to end the task — text-only responses will be nudged to continue.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"summary": map[string]interface{}{
					"type":        "string",
					"description": "Concise summary of what was accomplished: key files changed, approach taken, and any caveats.",
				},
			},
			"required": []string{"summary"},
		},
	}
}

func ExecuteTool(toolName string, toolInput map[string]interface{}, workingDir string, settings *claude.Settings) (string, error) {
	// Normalize to PascalCase so role files can declare tools in any case (e.g. "read" == "Read").
	if len(toolName) > 0 {
		toolName = strings.ToUpper(toolName[:1]) + toolName[1:]
	}
	switch toolName {
	case "Bash":
		return executeBash(toolInput, workingDir, settings)
	case "Read":
		return executeRead(toolInput, workingDir, settings)
	case "Write":
		return executeWrite(toolInput, workingDir, settings)
	case "Edit":
		return executeEdit(toolInput, workingDir, settings)
	case "MultiEdit":
		return executeMultiEdit(toolInput, workingDir, settings)
	case "Grep":
		return executeGrep(toolInput, workingDir, settings)
	case "Glob":
		return executeGlob(toolInput, workingDir, settings)
	case "WebSearch":
		return executeWebSearch(toolInput, workingDir, settings)
	case "WebFetch":
		return executeWebFetch(toolInput, workingDir, settings)
	case "TaskComplete":
		// The agentic loop intercepts TaskComplete before dispatch.
		// This path is a safety net in case it is called outside the loop.
		summary, _ := toolInput["summary"].(string)
		if summary == "" {
			return "", fmt.Errorf("TaskComplete requires a non-empty summary")
		}
		return fmt.Sprintf("Task marked complete: %s", summary), nil
	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

func executeBash(input map[string]interface{}, workingDir string, settings *claude.Settings) (string, error) {
	command, ok := input["command"].(string)
	if !ok {
		return "", fmt.Errorf("invalid bash command")
	}

	// ALWAYS block known dangerous commands (built-in safety)
	dangerousPatterns := []string{
		"rm -rf /",
		"rm -rf ~",
		"rm -rf /*",
		"rm -rf ~/*",
		":(){ :|:& };:", // fork bomb
		"mkfs",
		"dd if=/dev/zero",
		"dd if=/dev/random",
		"> /dev/sda",
		"> /dev/sd",
		"curl | sh", // arbitrary code execution
		"wget | sh",
		"curl | bash",
		"wget | bash",
		"chmod -r", // recursive permission changes can be dangerous
		"chown -r",
	}

	cmdLower := strings.ToLower(command)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(cmdLower, pattern) {
			return fmt.Sprintf("❌ Dangerous command blocked (built-in safety): %s", pattern), nil
		}
	}

	// Check against user's Claude settings deny patterns
	if allowed, reason := settings.IsOperationAllowed("Bash", command); !allowed {
		return fmt.Sprintf("❌ Command blocked by Claude settings: %s", reason), nil
	}

	// Enforce a per-command timeout so a hung or misbehaving process cannot
	// block the goroutine indefinitely.
	const bashTimeout = 2 * time.Minute
	cmdCtx, cancelCmd := context.WithTimeout(context.Background(), bashTimeout)
	defer cancelCmd()

	cmd := exec.CommandContext(cmdCtx, "bash", "-c", command)
	cmd.Dir = workingDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Sprintf("Command failed: %v\nOutput: %s", err, string(output)), nil
	}

	return string(output), nil
}

func executeRead(input map[string]interface{}, workingDir string, settings *claude.Settings) (string, error) {
	filePath, _ := input["file_path"].(string)

	// Make path absolute relative to working directory
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(workingDir, filePath)
	}

	// Check .claudeignore (hierarchical: root + subdirectories)
	claudeIgnore, err := LoadClaudeIgnoreForPath(workingDir, filePath)
	if err == nil && claudeIgnore.ShouldIgnore(filePath) {
		return fmt.Sprintf("⚠️  File ignored by .claudeignore: %s", filePath), nil
	}

	// Check if operation is allowed by Claude settings
	if allowed, reason := settings.IsOperationAllowed("Read", filePath); !allowed {
		return fmt.Sprintf("❌ Read blocked by Claude settings on %s: %s", filePath, reason), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Sprintf(errReadingFile, err), nil
	}

	return string(data), nil
}

func executeWrite(input map[string]interface{}, workingDir string, settings *claude.Settings) (string, error) {
	filePath, _ := input["file_path"].(string)
	content, _ := input["content"].(string)

	// Make path absolute relative to working directory
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(workingDir, filePath)
	}

	// Check if operation is allowed by Claude settings
	if allowed, reason := settings.IsOperationAllowed("Write", filePath); !allowed {
		return fmt.Sprintf("❌ Write blocked by Claude settings on %s: %s", filePath, reason), nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Sprintf("Error creating directory: %v", err), nil
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Sprintf(errWritingFile, err), nil
	}

	return fmt.Sprintf("File written: %s (%d bytes)", filePath, len(content)), nil
}

func executeEdit(input map[string]interface{}, workingDir string, settings *claude.Settings) (string, error) {
	filePath, _ := input["file_path"].(string)
	oldString, _ := input["old_string"].(string)
	newString, _ := input["new_string"].(string)

	// Make path absolute relative to working directory
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(workingDir, filePath)
	}

	// Check if operation is allowed by Claude settings
	if allowed, reason := settings.IsOperationAllowed("Edit", filePath); !allowed {
		return fmt.Sprintf("❌ Edit blocked by Claude settings on %s: %s", filePath, reason), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Sprintf(errReadingFile, err), nil
	}

	content := string(data)
	if !strings.Contains(content, oldString) {
		return fmt.Sprintf("Error: old_string not found in file.\n\nSearched for:\n%s", oldString), nil
	}

	// Replace first occurrence
	newContent := strings.Replace(content, oldString, newString, 1)

	// Detect no-op before writing — old_string == new_string means nothing changed.
	if newContent == content {
		return fmt.Sprintf("No changes made to %s (old_string and new_string are identical)", filePath), nil
	}

	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Sprintf(errWritingFile, err), nil
	}

	oldLines := strings.Count(oldString, "\n") + 1
	newLines := strings.Count(newString, "\n") + 1
	return fmt.Sprintf("File edited: %s (replaced %d lines with %d lines)", filePath, oldLines, newLines), nil
}

func executeMultiEdit(input map[string]interface{}, workingDir string, settings *claude.Settings) (string, error) {
	filePath, _ := input["file_path"].(string)
	editsRaw, ok := input["edits"].([]interface{})
	if !ok {
		return "", fmt.Errorf("invalid edits parameter: must be an array")
	}

	// Make path absolute relative to working directory
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(workingDir, filePath)
	}

	// Check if operation is allowed by Claude settings
	if allowed, reason := settings.IsOperationAllowed("Edit", filePath); !allowed {
		return fmt.Sprintf("❌ Edit blocked by Claude settings on %s: %s", filePath, reason), nil
	}

	// Read file once
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Sprintf(errReadingFile, err), nil
	}

	content := string(data)
	originalContent := content
	editsApplied := 0

	// Apply each edit in sequence
	for i, editRaw := range editsRaw {
		editMap, ok := editRaw.(map[string]interface{})
		if !ok {
			return fmt.Sprintf("Error: edit #%d is not a valid object", i+1), nil
		}

		oldString, _ := editMap["old_string"].(string)
		newString, _ := editMap["new_string"].(string)

		if oldString == "" {
			return fmt.Sprintf("Error: edit #%d has empty old_string", i+1), nil
		}

		// Check if old_string exists in current content
		if !strings.Contains(content, oldString) {
			return fmt.Sprintf("Error: edit #%d - old_string not found in file after %d previous edits.\n\nSearched for:\n%s", i+1, editsApplied, oldString), nil
		}

		// Replace first occurrence
		content = strings.Replace(content, oldString, newString, 1)
		editsApplied++
	}

	// Only write if content changed
	if content == originalContent {
		return fmt.Sprintf("No changes made to %s (all edits were no-ops)", filePath), nil
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Sprintf(errWritingFile, err), nil
	}

	return fmt.Sprintf("File edited: %s (%d edits applied)", filePath, editsApplied), nil
}

func executeGrep(input map[string]interface{}, workingDir string, settings *claude.Settings) (string, error) {
	pattern, _ := input["pattern"].(string)
	searchPath, _ := input["path"].(string)

	if searchPath == "" {
		searchPath = workingDir
	} else if !filepath.IsAbs(searchPath) {
		searchPath = filepath.Join(workingDir, searchPath)
	}

	// Load .claudeignore and build grep exclusion flags so grep never scans ignored paths.
	claudeIgnore, ignoreErr := LoadClaudeIgnore(workingDir)
	grepArgs := []string{"-r", "-n", "-I"} // -I skips binary files
	if ignoreErr == nil {
		grepArgs = append(grepArgs, claudeIgnore.GrepExcludeArgs()...)
	}
	grepArgs = append(grepArgs, pattern, searchPath)

	cmd := exec.Command("grep", grepArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// grep returns exit code 1 when no matches found
		if len(output) == 0 {
			return "No matches found", nil
		}
	}

	// Post-filter output lines for complex patterns that couldn't be expressed as grep flags.
	if ignoreErr == nil && len(claudeIgnore.patterns) > 0 {
		lines := strings.Split(string(output), "\n")
		filtered := make([]string, 0, len(lines))

		for _, line := range lines {
			if line == "" {
				continue
			}

			// Extract filename from grep output (format: "filename:line:content")
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 0 {
				filename := parts[0]
				absPath := filename
				if !filepath.IsAbs(filename) {
					absPath = filepath.Join(workingDir, filename)
				}

				if !claudeIgnore.ShouldIgnore(absPath) {
					filtered = append(filtered, line)
				}
			}
		}

		if len(filtered) == 0 {
			return "No matches found (after applying .claudeignore)", nil
		}

		return strings.Join(filtered, "\n"), nil
	}

	return string(output), nil
}

func executeGlob(input map[string]interface{}, workingDir string, settings *claude.Settings) (string, error) {
	pattern, _ := input["pattern"].(string)

	// Build an absolute pattern so filepath.Glob works without os.Chdir.
	// os.Chdir is process-wide and causes a race condition when multiple agents
	// run concurrently — each call would overwrite the working directory for all
	// goroutines. Using filepath.Join avoids that entirely.
	absPattern := filepath.Join(workingDir, pattern)

	matches, err := filepath.Glob(absPattern)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	// Filter out ignored files
	claudeIgnore, err := LoadClaudeIgnore(workingDir)
	if err == nil && len(claudeIgnore.patterns) > 0 {
		filtered := make([]string, 0, len(matches))
		for _, match := range matches {
			if !claudeIgnore.ShouldIgnore(match) {
				filtered = append(filtered, match)
			}
		}
		matches = filtered
	}

	// Return paths relative to workingDir so callers see the same short paths
	// they would have seen before (e.g. "src/foo.go" instead of
	// "/abs/path/to/src/foo.go").
	rel := make([]string, 0, len(matches))
	for _, match := range matches {
		r, err := filepath.Rel(workingDir, match)
		if err != nil {
			r = match // fall back to absolute if Rel fails
		}
		rel = append(rel, r)
	}
	matches = rel

	if len(matches) == 0 {
		return "No files found matching pattern", nil
	}

	return strings.Join(matches, "\n"), nil
}

func executeWebSearch(input map[string]interface{}, workingDir string, settings *claude.Settings) (string, error) {
	// NOTE: WebSearch requires API integration (Google, Bing, etc.)
	// For now, return a placeholder
	query, _ := input["query"].(string)
	return fmt.Sprintf("WebSearch not yet implemented. Query: %s", query), nil
}

func executeWebFetch(input map[string]interface{}, workingDir string, settings *claude.Settings) (string, error) {
	// NOTE: WebFetch requires HTTP client and HTML parsing
	// For now, return a placeholder
	url, _ := input["url"].(string)
	return fmt.Sprintf("WebFetch not yet implemented. URL: %s", url), nil
}
