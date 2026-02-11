package main

import (
	"os"
	"testing"
)

func TestParseMultilineInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "single line",
			input:    "file.txt",
			expected: []string{"file.txt"},
		},
		{
			name:     "multiple lines",
			input:    "file1.txt\nfile2.txt\nfile3.txt",
			expected: []string{"file1.txt", "file2.txt", "file3.txt"},
		},
		{
			name:     "with empty lines",
			input:    "file1.txt\n\nfile2.txt\n\n",
			expected: []string{"file1.txt", "file2.txt"},
		},
		{
			name:     "with whitespace",
			input:    "  file1.txt  \n  file2.txt  ",
			expected: []string{"file1.txt", "file2.txt"},
		},
		{
			name:     "glob patterns",
			input:    "dist/*\n*.tar.gz\n**/*.bin",
			expected: []string{"dist/*", "*.tar.gz", "**/*.bin"},
		},
		{
			name:     "only whitespace lines",
			input:    "   \n\t\n  \n",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMultilineInput(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d items, got %d", len(tt.expected), len(result))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("item %d: expected %q, got %q", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name        string
		args        ActionInputs
		mockSigner  *MockSigner
		mockFinder  *MockFileFinder
		expectError bool
	}{
		{
			name: "successful signing single file",
			args: ActionInputs{
				PrivateKey: "test-key",
				Files:      "*.txt",
				Armor:      true,
				DetachSign: true,
			},
			mockSigner: &MockSigner{},
			mockFinder: &MockFileFinder{
				Files: []string{"/tmp/test/file.txt"},
			},
			expectError: false,
		},
		{
			name: "successful signing multiple files",
			args: ActionInputs{
				PrivateKey: "test-key",
				Files:      "*.txt\n*.bin",
				Armor:      true,
			},
			mockSigner: &MockSigner{},
			mockFinder: &MockFileFinder{
				Files: []string{"/tmp/test/file.txt", "/tmp/test/file.bin"},
			},
			expectError: false,
		},
		{
			name: "no files matched",
			args: ActionInputs{
				PrivateKey: "test-key",
				Files:      "*.nonexistent",
			},
			mockSigner: &MockSigner{},
			mockFinder: &MockFileFinder{
				Files: []string{},
			},
			expectError: false,
		},
		{
			name: "signer error",
			args: ActionInputs{
				PrivateKey: "test-key",
				Files:      "*.txt",
			},
			mockSigner: &MockSigner{
				Err: os.ErrPermission,
			},
			mockFinder: &MockFileFinder{
				Files: []string{"/tmp/test/file.txt"},
			},
			expectError: true,
		},
		{
			name: "finder error",
			args: ActionInputs{
				PrivateKey: "test-key",
				Files:      "*.txt",
			},
			mockSigner: &MockSigner{},
			mockFinder: &MockFileFinder{
				Err: os.ErrNotExist,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.args, tt.mockSigner, tt.mockFinder, nil)
			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRunSignOptionsPassthrough(t *testing.T) {
	tests := []struct {
		name         string
		args         ActionInputs
		expectedOpts SignOptions
	}{
		{
			name: "armor only",
			args: ActionInputs{
				PrivateKey: "key",
				Files:      "file.txt",
				Armor:      true,
			},
			expectedOpts: SignOptions{
				Armor:      true,
				DetachSign: false,
				ClearSign:  false,
			},
		},
		{
			name: "detach sign with armor",
			args: ActionInputs{
				PrivateKey: "key",
				Files:      "file.txt",
				Armor:      true,
				DetachSign: true,
			},
			expectedOpts: SignOptions{
				Armor:      true,
				DetachSign: true,
				ClearSign:  false,
			},
		},
		{
			name: "clear sign",
			args: ActionInputs{
				PrivateKey: "key",
				Files:      "file.txt",
				ClearSign:  true,
			},
			expectedOpts: SignOptions{
				Armor:      false,
				DetachSign: false,
				ClearSign:  true,
			},
		},
		{
			name: "no armor binary output",
			args: ActionInputs{
				PrivateKey: "key",
				Files:      "file.txt",
				Armor:      false,
				DetachSign: true,
			},
			expectedOpts: SignOptions{
				Armor:      false,
				DetachSign: true,
				ClearSign:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSigner := &MockSigner{}
			mockFinder := &MockFileFinder{
				Files: []string{"/tmp/file.txt"},
			}

			_ = run(tt.args, mockSigner, mockFinder, nil)

			if len(mockSigner.SignedFiles) != 1 {
				t.Fatalf("expected 1 signed file, got %d", len(mockSigner.SignedFiles))
			}

			opts := mockSigner.SignedOpts[0]
			if opts.Armor != tt.expectedOpts.Armor {
				t.Errorf("Armor: expected %v, got %v", tt.expectedOpts.Armor, opts.Armor)
			}
			if opts.DetachSign != tt.expectedOpts.DetachSign {
				t.Errorf("DetachSign: expected %v, got %v", tt.expectedOpts.DetachSign, opts.DetachSign)
			}
			if opts.ClearSign != tt.expectedOpts.ClearSign {
				t.Errorf("ClearSign: expected %v, got %v", tt.expectedOpts.ClearSign, opts.ClearSign)
			}
		})
	}
}
