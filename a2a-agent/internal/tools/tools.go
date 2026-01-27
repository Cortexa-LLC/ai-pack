package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/claude"
)

const (
	errReadingFile = "Error reading file: %v"
	errWritingFile = "Error writing file: %v"
)

// Define tools matching Claude Code's exact naming
func DefineTools() []anthropic.ToolParam {
	return []anthropic.ToolParam{
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

		// NOTE: Web operations disabled - not yet implemented
		// defineWebSearchTool(),
		// defineWebFetchTool(),
	}
}

// defineBashTool creates the Bash tool (matches Claude Code)
func defineBashTool() anthropic.ToolParam {
	properties := map[string]interface{}{
		"command": map[string]interface{}{
			"type":        "string",
			"description": "The bash command to execute",
		},
	}

	return anthropic.ToolParam{
		Name:        "Bash",
		Description: param.NewOpt("Execute bash commands in the working directory. Use this to run git, npm, tests, or any other shell commands."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type:       "object",
			Properties: properties,
			Required:   []string{"command"},
		},
	}
}

// defineReadTool creates the Read tool (matches Claude Code)
func defineReadTool() anthropic.ToolParam {
	properties := map[string]interface{}{
		"file_path": map[string]interface{}{
			"type":        "string",
			"description": "Path to the file to read (relative to working directory)",
		},
	}

	return anthropic.ToolParam{
		Name:        "Read",
		Description: param.NewOpt("Read the contents of a file."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type:       "object",
			Properties: properties,
			Required:   []string{"file_path"},
		},
	}
}

// defineWriteTool creates the Write tool (matches Claude Code)
func defineWriteTool() anthropic.ToolParam {
	properties := map[string]interface{}{
		"file_path": map[string]interface{}{
			"type":        "string",
			"description": "Path to the file to write (relative to working directory)",
		},
		"content": map[string]interface{}{
			"type":        "string",
			"description": "Content to write to the file",
		},
	}

	return anthropic.ToolParam{
		Name:        "Write",
		Description: param.NewOpt("Create a new file with the given content. Will overwrite if file exists."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type:       "object",
			Properties: properties,
			Required:   []string{"file_path", "content"},
		},
	}
}

// defineEditTool creates the Edit tool (matches Claude Code)
func defineEditTool() anthropic.ToolParam {
	properties := map[string]interface{}{
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
	}

	return anthropic.ToolParam{
		Name:        "Edit",
		Description: param.NewOpt("Edit a file by replacing an exact string match. The old_string must match exactly (including whitespace and indentation)."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type:       "object",
			Properties: properties,
			Required:   []string{"file_path", "old_string", "new_string"},
		},
	}
}

// defineMultiEditTool creates the MultiEdit tool for making multiple edits at once
func defineMultiEditTool() anthropic.ToolParam {
	properties := map[string]interface{}{
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
	}

	return anthropic.ToolParam{
		Name:        "MultiEdit",
		Description: param.NewOpt("Make multiple edits to a file in a single operation. Each edit replaces an exact string match."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type:       "object",
			Properties: properties,
			Required:   []string{"file_path", "edits"},
		},
	}
}

// defineGrepTool creates the Grep tool (matches Claude Code)
func defineGrepTool() anthropic.ToolParam {
	properties := map[string]interface{}{
		"pattern": map[string]interface{}{
			"type":        "string",
			"description": "The pattern to search for (regex supported)",
		},
		"path": map[string]interface{}{
			"type":        "string",
			"description": "Optional path to search in (defaults to current directory)",
		},
	}

	return anthropic.ToolParam{
		Name:        "Grep",
		Description: param.NewOpt("Search for patterns in files. Returns matching lines with file paths."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type:       "object",
			Properties: properties,
			Required:   []string{"pattern"},
		},
	}
}

// defineGlobTool creates the Glob tool (matches Claude Code)
func defineGlobTool() anthropic.ToolParam {
	properties := map[string]interface{}{
		"pattern": map[string]interface{}{
			"type":        "string",
			"description": "Glob pattern to match files (e.g., **/*.go, src/**/*.ts)",
		},
	}

	return anthropic.ToolParam{
		Name:        "Glob",
		Description: param.NewOpt("Find files matching a glob pattern. Returns list of file paths."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type:       "object",
			Properties: properties,
			Required:   []string{"pattern"},
		},
	}
}

// defineWebSearchTool creates the WebSearch tool (matches Claude Code)
func defineWebSearchTool() anthropic.ToolParam {
	properties := map[string]interface{}{
		"query": map[string]interface{}{
			"type":        "string",
			"description": "The search query",
		},
	}

	return anthropic.ToolParam{
		Name:        "WebSearch",
		Description: param.NewOpt("Search the web for information. Returns search results with titles, URLs, and snippets."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type:       "object",
			Properties: properties,
			Required:   []string{"query"},
		},
	}
}

// defineWebFetchTool creates the WebFetch tool (matches Claude Code)
func defineWebFetchTool() anthropic.ToolParam {
	properties := map[string]interface{}{
		"url": map[string]interface{}{
			"type":        "string",
			"description": "The URL to fetch",
		},
		"prompt": map[string]interface{}{
			"type":        "string",
			"description": "Optional prompt describing what to extract from the page",
		},
	}

	return anthropic.ToolParam{
		Name:        "WebFetch",
		Description: param.NewOpt("Fetch and parse a web page. Returns the page content, optionally filtered by prompt."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type:       "object",
			Properties: properties,
			Required:   []string{"url"},
		},
	}
}

// ExecuteTool executes a tool call and returns the result
func ExecuteTool(toolName string, toolInput map[string]interface{}, workingDir string, settings *claude.Settings) (string, error) {
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

	cmd := exec.Command("bash", "-c", command)
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

	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Sprintf(errWritingFile, err), nil
	}

	return fmt.Sprintf("File edited: %s", filePath), nil
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

	// Use grep command
	cmd := exec.Command("grep", "-r", "-n", pattern, searchPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// grep returns exit code 1 when no matches found
		if len(output) == 0 {
			return "No matches found", nil
		}
	}

	return string(output), nil
}

func executeGlob(input map[string]interface{}, workingDir string, settings *claude.Settings) (string, error) {
	pattern, _ := input["pattern"].(string)

	// Change to working directory for glob
	oldDir, _ := os.Getwd()
	os.Chdir(workingDir)
	defer os.Chdir(oldDir)

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

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
