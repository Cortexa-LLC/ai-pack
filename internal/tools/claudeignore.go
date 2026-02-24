package tools

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ClaudeIgnore handles .claudeignore pattern matching
type ClaudeIgnore struct {
	patterns []ignorePattern
	rootDir  string
}

type ignorePattern struct {
	pattern string
	negate  bool
}

// LoadClaudeIgnore loads and parses .claudeignore files hierarchically
// It loads the root .claudeignore first, then any .claudeignore files in subdirectories
// along the path to the target file, allowing subdirectories to override root patterns
func LoadClaudeIgnore(rootDir string) (*ClaudeIgnore, error) {
	ci := &ClaudeIgnore{
		patterns: make([]ignorePattern, 0),
		rootDir:  rootDir,
	}

	// Load root .claudeignore first
	rootIgnorePath := filepath.Join(rootDir, ".claudeignore")
	if err := loadIgnoreFile(rootIgnorePath, ci); err != nil {
		// Ignore errors for missing files, but return other errors
		if !os.IsNotExist(err) {
			return ci, err
		}
	}

	return ci, nil
}

// LoadClaudeIgnoreForPath loads .claudeignore files hierarchically for a specific path
// It loads the root .claudeignore, then any .claudeignore files in subdirectories
// along the path, allowing closer .claudeignore files to override parent patterns
func LoadClaudeIgnoreForPath(rootDir string, targetPath string) (*ClaudeIgnore, error) {
	ci := &ClaudeIgnore{
		patterns: make([]ignorePattern, 0),
		rootDir:  rootDir,
	}

	// Load root .claudeignore first
	rootIgnorePath := filepath.Join(rootDir, ".claudeignore")
	if err := loadIgnoreFile(rootIgnorePath, ci); err != nil {
		if !os.IsNotExist(err) {
			return ci, err
		}
	}

	// Walk up the directory tree from target to root, loading .claudeignore files
	currentDir := filepath.Dir(targetPath)
	for {
		// Don't go above root directory
		if !strings.HasPrefix(currentDir, rootDir) || currentDir == rootDir {
			break
		}

		ignorePath := filepath.Join(currentDir, ".claudeignore")
		if err := loadIgnoreFile(ignorePath, ci); err != nil {
			if !os.IsNotExist(err) {
				return ci, err
			}
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}
		currentDir = parentDir
	}

	return ci, nil
}

// loadIgnoreFile loads patterns from a single .claudeignore file
func loadIgnoreFile(ignorePath string, ci *ClaudeIgnore) error {
	// If .claudeignore doesn't exist, skip it
	if _, err := os.Stat(ignorePath); os.IsNotExist(err) {
		return err
	}

	file, err := os.Open(ignorePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		pattern := ignorePattern{}

		// Check for negation
		if strings.HasPrefix(line, "!") {
			pattern.negate = true
			line = strings.TrimPrefix(line, "!")
		}

		pattern.pattern = line
		ci.patterns = append(ci.patterns, pattern)
	}

	return scanner.Err()
}

// ShouldIgnore checks if a path should be ignored based on .claudeignore patterns
func (ci *ClaudeIgnore) ShouldIgnore(path string) bool {
	if len(ci.patterns) == 0 {
		return false
	}

	// Make path relative to root directory for matching
	relPath, err := filepath.Rel(ci.rootDir, path)
	if err != nil {
		// If we can't make it relative, use as-is
		relPath = path
	}

	// Clean the path for consistent matching
	relPath = filepath.Clean(relPath)

	ignored := false

	// Process patterns in order - later patterns can override earlier ones
	for _, p := range ci.patterns {
		matched := matchPattern(p.pattern, relPath)

		if matched {
			if p.negate {
				ignored = false
			} else {
				ignored = true
			}
		}
	}

	return ignored
}

// matchPattern checks if a path matches a .claudeignore pattern
func matchPattern(pattern, path string) bool {
	// Handle common patterns

	// Pattern ending with / matches directories only
	if strings.HasSuffix(pattern, "/") {
		pattern = strings.TrimSuffix(pattern, "/")
		// For now, we'll treat all paths as potentially being directories
		// A more complete implementation would check os.Stat
	}

	// Pattern starting with / is anchored to root
	if strings.HasPrefix(pattern, "/") {
		pattern = strings.TrimPrefix(pattern, "/")
		// Direct match from root
		matched, _ := filepath.Match(pattern, path)
		if matched {
			return true
		}
		// Also check if path starts with pattern for directory matching
		if strings.HasPrefix(path, pattern) {
			return true
		}
		return false
	}

	// Handle ** (match any directory depth)
	if strings.Contains(pattern, "**") {
		// Convert ** to a simple prefix/suffix check
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			if prefix == "" && suffix == "" {
				return true // "**" matches everything
			}
			if prefix == "" {
				// "**/pattern" - match pattern anywhere
				matched, _ := filepath.Match(suffix, filepath.Base(path))
				if matched {
					return true
				}
				// Also check if path ends with suffix
				return strings.HasSuffix(path, suffix) || strings.Contains(path, suffix)
			}
			if suffix == "" {
				// "pattern/**" - match pattern as prefix
				return strings.HasPrefix(path, prefix) || strings.Contains(path, prefix)
			}
			// "prefix/**/suffix" - check both
			return strings.Contains(path, prefix) && strings.Contains(path, suffix)
		}
	}

	// Simple glob match against basename
	matched, _ := filepath.Match(pattern, filepath.Base(path))
	if matched {
		return true
	}

	// Try matching against full path
	matched, _ = filepath.Match(pattern, path)
	if matched {
		return true
	}

	// Check if any path component matches
	pathParts := strings.Split(path, string(filepath.Separator))
	for _, part := range pathParts {
		matched, _ := filepath.Match(pattern, part)
		if matched {
			return true
		}
	}

	// Check if path contains the pattern as a directory
	return strings.Contains(path, pattern)
}

// GrepExcludeArgs converts .claudeignore patterns into grep --exclude-dir / --exclude
// flags so that grep never scans ignored directories or files in the first place.
// Complex path patterns (e.g. ".beads/tasks/*/execution.log") that can't be expressed
// as simple grep flags are skipped here and still handled by post-filtering.
func (ci *ClaudeIgnore) GrepExcludeArgs() []string {
	var args []string
	seen := make(map[string]bool)

	add := func(arg string) {
		if !seen[arg] {
			seen[arg] = true
			args = append(args, arg)
		}
	}

	for _, p := range ci.patterns {
		if p.negate {
			continue
		}
		pat := p.pattern

		// Strip leading **/ prefix: "**/node_modules/" → "node_modules/"
		pat = strings.TrimPrefix(pat, "**/")

		if strings.HasSuffix(pat, "/") {
			// Directory pattern — extract the directory name.
			// Only use as --exclude-dir if there are no remaining path separators,
			// i.e. it's a bare name like "build/" not "docs/website/node_modules/".
			name := strings.TrimSuffix(pat, "/")
			if !strings.Contains(name, "/") {
				add("--exclude-dir=" + name)
			}
		} else if !strings.Contains(pat, "/") {
			// Simple file glob with no directory component: "*.log", "*.o", ".DS_Store"
			add("--exclude=" + pat)
		}
		// Skip multi-segment path patterns — post-filtering handles those.
	}

	return args
}

// FilterPaths filters a list of paths based on .claudeignore rules
func (ci *ClaudeIgnore) FilterPaths(paths []string) []string {
	if len(ci.patterns) == 0 {
		return paths
	}

	filtered := make([]string, 0, len(paths))
	for _, path := range paths {
		if !ci.ShouldIgnore(path) {
			filtered = append(filtered, path)
		}
	}
	return filtered
}
