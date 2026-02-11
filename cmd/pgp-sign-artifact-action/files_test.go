package main

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestDefaultFileFinder_FindFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"file1.txt",
		"file2.txt",
		"file.bin",
		"data.json",
		"subdir/file3.txt",
		"subdir/file4.bin",
		"dist/release.tar.gz",
		"dist/release.sha256",
	}

	for _, f := range testFiles {
		path := filepath.Join(tempDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	tests := []struct {
		name          string
		patterns      []string
		excludes      []string
		expectedFiles []string
	}{
		{
			name:     "single pattern match txt files",
			patterns: []string{"*.txt"},
			expectedFiles: []string{
				filepath.Join(tempDir, "file1.txt"),
				filepath.Join(tempDir, "file2.txt"),
			},
		},
		{
			name:     "multiple patterns",
			patterns: []string{"*.txt", "*.bin"},
			expectedFiles: []string{
				filepath.Join(tempDir, "file1.txt"),
				filepath.Join(tempDir, "file2.txt"),
				filepath.Join(tempDir, "file.bin"),
			},
		},
		{
			name:     "pattern with subdirectory",
			patterns: []string{"subdir/*.txt"},
			expectedFiles: []string{
				filepath.Join(tempDir, "subdir/file3.txt"),
			},
		},
		{
			name:     "pattern with exclusion",
			patterns: []string{"dist/*"},
			excludes: []string{"*.sha256"},
			expectedFiles: []string{
				filepath.Join(tempDir, "dist/release.tar.gz"),
			},
		},
		{
			name:     "exclude by filename",
			patterns: []string{"*.txt"},
			excludes: []string{"file1.txt"},
			expectedFiles: []string{
				filepath.Join(tempDir, "file2.txt"),
			},
		},
		{
			name:          "no matches",
			patterns:      []string{"*.nonexistent"},
			expectedFiles: []string{},
		},
		{
			name:          "empty patterns",
			patterns:      []string{},
			expectedFiles: []string{},
		},
	}

	finder := &DefaultFileFinder{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := finder.FindFiles(tempDir, tt.patterns, tt.excludes)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			sort.Strings(files)
			sort.Strings(tt.expectedFiles)

			if len(files) != len(tt.expectedFiles) {
				t.Errorf("expected %d files, got %d\nexpected: %v\ngot: %v",
					len(tt.expectedFiles), len(files), tt.expectedFiles, files)
				return
			}

			for i := range files {
				if files[i] != tt.expectedFiles[i] {
					t.Errorf("file %d: expected %q, got %q", i, tt.expectedFiles[i], files[i])
				}
			}
		})
	}
}

func TestShouldExclude(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		workDir  string
		excludes []string
		expected bool
	}{
		{
			name:     "no excludes",
			file:     "/work/file.txt",
			workDir:  "/work",
			excludes: nil,
			expected: false,
		},
		{
			name:     "match by extension",
			file:     "/work/file.sha256",
			workDir:  "/work",
			excludes: []string{"*.sha256"},
			expected: true,
		},
		{
			name:     "no match different extension",
			file:     "/work/file.txt",
			workDir:  "/work",
			excludes: []string{"*.sha256"},
			expected: false,
		},
		{
			name:     "match by exact name",
			file:     "/work/secret.txt",
			workDir:  "/work",
			excludes: []string{"secret.txt"},
			expected: true,
		},
		{
			name:     "multiple excludes first matches",
			file:     "/work/file.bak",
			workDir:  "/work",
			excludes: []string{"*.bak", "*.tmp"},
			expected: true,
		},
		{
			name:     "multiple excludes none match",
			file:     "/work/file.txt",
			workDir:  "/work",
			excludes: []string{"*.bak", "*.tmp"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldExclude(tt.file, tt.workDir, tt.excludes)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFindFiles_DirectoriesExcluded(t *testing.T) {
	tempDir := t.TempDir()

	// Create a directory that matches the pattern
	if err := os.Mkdir(filepath.Join(tempDir, "matches.txt"), 0o755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Create a file that matches the pattern
	filePath := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("content"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	finder := &DefaultFileFinder{}
	files, err := finder.FindFiles(tempDir, []string{"*.txt"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(files), files)
	}
}

func TestFindFiles_NoDuplicates(t *testing.T) {
	tempDir := t.TempDir()

	filePath := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("content"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	finder := &DefaultFileFinder{}
	files, err := finder.FindFiles(tempDir, []string{"*.txt", "file.*", "file.txt"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 file (no duplicates), got %d: %v", len(files), files)
	}
}
