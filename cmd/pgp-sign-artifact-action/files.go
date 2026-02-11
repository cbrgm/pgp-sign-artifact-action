package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileFinder defines the interface for finding files.
type FileFinder interface {
	FindFiles(workDir string, patterns []string, excludes []string) ([]string, error)
}

// DefaultFileFinder implements FileFinder using the standard library.
type DefaultFileFinder struct{}

// FindFiles finds files matching patterns while excluding others.
func (f *DefaultFileFinder) FindFiles(workDir string, patterns []string, excludes []string) ([]string, error) {
	if workDir == "" {
		workDir = "."
	}

	var matchedFiles []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		fullPattern := filepath.Join(workDir, pattern)

		// Handle ** globstar patterns by walking the directory
		if strings.Contains(pattern, "**") {
			files, err := findWithGlobstar(workDir, pattern)
			if err != nil {
				return nil, err
			}
			for _, match := range files {
				if !seen[match] && !shouldExclude(match, workDir, excludes) {
					seen[match] = true
					matchedFiles = append(matchedFiles, match)
				}
			}
			continue
		}

		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}

		for _, match := range matches {
			if seen[match] {
				continue
			}

			info, err := os.Stat(match)
			if err != nil {
				continue
			}
			if info.IsDir() {
				continue
			}

			if shouldExclude(match, workDir, excludes) {
				continue
			}

			seen[match] = true
			matchedFiles = append(matchedFiles, match)
		}
	}

	return matchedFiles, nil
}

// findWithGlobstar handles patterns containing ** for recursive matching.
func findWithGlobstar(workDir, pattern string) ([]string, error) {
	var matches []string

	// Split the pattern into parts
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		// Fallback to simple glob if pattern is complex
		return filepath.Glob(filepath.Join(workDir, strings.ReplaceAll(pattern, "**", "*")))
	}

	prefix := strings.TrimSuffix(parts[0], string(filepath.Separator))
	suffix := strings.TrimPrefix(parts[1], string(filepath.Separator))

	searchDir := filepath.Join(workDir, prefix)
	if prefix == "" {
		searchDir = workDir
	}

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files/dirs we can't access
		}

		if info.IsDir() {
			return nil
		}

		// Match the suffix pattern
		if suffix != "" {
			matched, err := filepath.Match(suffix, filepath.Base(path))
			if err != nil || !matched {
				// Try matching the full relative suffix path
				relPath, _ := filepath.Rel(searchDir, path)
				matched, _ = filepath.Match(suffix, relPath)
				if !matched {
					return nil
				}
			}
		}

		matches = append(matches, path)
		return nil
	})

	return matches, err
}

// shouldExclude checks if a file matches any exclusion pattern.
func shouldExclude(file, workDir string, excludes []string) bool {
	relPath, err := filepath.Rel(workDir, file)
	if err != nil {
		relPath = file
	}

	for _, exclude := range excludes {
		exclude = strings.TrimSpace(exclude)
		if exclude == "" {
			continue
		}

		// Direct match against relative path
		if matched, _ := filepath.Match(exclude, relPath); matched {
			return true
		}

		// Match against base name
		if matched, _ := filepath.Match(exclude, filepath.Base(file)); matched {
			return true
		}

		// Handle ** patterns
		if strings.Contains(exclude, "**") {
			simplePattern := strings.ReplaceAll(exclude, "**"+string(filepath.Separator), "")
			simplePattern = strings.ReplaceAll(simplePattern, "**", "")
			if simplePattern != "" {
				if matched, _ := filepath.Match(simplePattern, filepath.Base(file)); matched {
					return true
				}
				if matched, _ := filepath.Match(simplePattern, relPath); matched {
					return true
				}
			}
		}

		// Full path match
		excludePattern := filepath.Join(workDir, exclude)
		if matched, _ := filepath.Match(excludePattern, file); matched {
			return true
		}
	}

	return false
}
