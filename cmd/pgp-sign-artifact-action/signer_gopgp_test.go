package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGoPGPSigner_ValidKey(t *testing.T) {
	armoredKey := generateTestKeyArmored(t, "Test", "test@test.com", "")

	signer, err := NewGoPGPSigner(armoredKey, "")
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}
	if signer == nil {
		t.Fatal("signer is nil")
	}
}

func TestNewGoPGPSigner_KeyWithPassphrase(t *testing.T) {
	armoredKey := generateTestKeyArmored(t, "Test", "test@test.com", "secret123")

	signer, err := NewGoPGPSigner(armoredKey, "secret123")
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}
	if signer == nil {
		t.Fatal("signer is nil")
	}
}

func TestNewGoPGPSigner_WrongPassphrase(t *testing.T) {
	armoredKey := generateTestKeyArmored(t, "Test", "test@test.com", "secret123")

	_, err := NewGoPGPSigner(armoredKey, "wrongpass")
	if err == nil {
		t.Fatal("expected error for wrong passphrase")
	}
}

func TestNewGoPGPSigner_MissingPassphrase(t *testing.T) {
	armoredKey := generateTestKeyArmored(t, "Test", "test@test.com", "secret123")

	_, err := NewGoPGPSigner(armoredKey, "")
	if err == nil {
		t.Fatal("expected error for missing passphrase")
	}
}

func TestNewGoPGPSigner_InvalidKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "empty key", key: ""},
		{name: "invalid armor", key: "not a valid key"},
		{name: "truncated key", key: "-----BEGIN PGP PRIVATE KEY BLOCK-----\ntruncated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGoPGPSigner(tt.key, "")
			if err == nil {
				t.Error("expected error for invalid key")
			}
		})
	}
}

func TestGoPGPSigner_SignFile_DetachedArmor(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("Hello, World!"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	armoredKey := generateTestKeyArmored(t, "Test", "test@test.com", "")
	signer, err := NewGoPGPSigner(armoredKey, "")
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}

	opts := SignOptions{Armor: true, DetachSign: true}
	if err := signer.SignFile(testFile, opts); err != nil {
		t.Fatalf("failed to sign file: %v", err)
	}

	sigFile := testFile + ".asc"
	content, err := os.ReadFile(sigFile)
	if err != nil {
		t.Fatalf("failed to read signature: %v", err)
	}

	if !strings.Contains(string(content), "BEGIN PGP SIGNATURE") {
		t.Error("signature should contain PGP SIGNATURE header")
	}
}

func TestGoPGPSigner_SignFile_DetachedBinary(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("Hello, World!"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	armoredKey := generateTestKeyArmored(t, "Test", "test@test.com", "")
	signer, err := NewGoPGPSigner(armoredKey, "")
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}

	opts := SignOptions{Armor: false, DetachSign: true}
	if err := signer.SignFile(testFile, opts); err != nil {
		t.Fatalf("failed to sign file: %v", err)
	}

	sigFile := testFile + ".sig"
	content, err := os.ReadFile(sigFile)
	if err != nil {
		t.Fatalf("failed to read signature: %v", err)
	}

	if strings.Contains(string(content), "BEGIN PGP") {
		t.Error("binary signature should not contain ASCII armor")
	}
}

func TestGoPGPSigner_SignFile_ClearSign(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	armoredKey := generateTestKeyArmored(t, "Test", "test@test.com", "")
	signer, err := NewGoPGPSigner(armoredKey, "")
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}

	opts := SignOptions{ClearSign: true}
	if err := signer.SignFile(testFile, opts); err != nil {
		t.Fatalf("failed to sign file: %v", err)
	}

	sigFile := testFile + ".asc"
	content, err := os.ReadFile(sigFile)
	if err != nil {
		t.Fatalf("failed to read signature: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "BEGIN PGP SIGNED MESSAGE") {
		t.Error("clear signature should contain SIGNED MESSAGE header")
	}
	if !strings.Contains(contentStr, testContent) {
		t.Error("clear signature should contain original message")
	}
}

func TestGoPGPSigner_SignFile_InlineArmor(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("Hello, World!"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	armoredKey := generateTestKeyArmored(t, "Test", "test@test.com", "")
	signer, err := NewGoPGPSigner(armoredKey, "")
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}

	opts := SignOptions{Armor: true}
	if err := signer.SignFile(testFile, opts); err != nil {
		t.Fatalf("failed to sign file: %v", err)
	}

	sigFile := testFile + ".asc"
	content, err := os.ReadFile(sigFile)
	if err != nil {
		t.Fatalf("failed to read signature: %v", err)
	}

	if !strings.Contains(string(content), "BEGIN PGP MESSAGE") {
		t.Error("inline signature should contain PGP MESSAGE header")
	}
}

func TestGoPGPSigner_SignFile_NonexistentFile(t *testing.T) {
	armoredKey := generateTestKeyArmored(t, "Test", "test@test.com", "")
	signer, err := NewGoPGPSigner(armoredKey, "")
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}

	err = signer.SignFile("/nonexistent/path/file.txt", SignOptions{})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestGoPGPSigner_SignFile_WithPassphrase(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("Hello, World!"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	armoredKey := generateTestKeyArmored(t, "Test", "test@test.com", "secret123")
	signer, err := NewGoPGPSigner(armoredKey, "secret123")
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}

	opts := SignOptions{Armor: true, DetachSign: true}
	if err := signer.SignFile(testFile, opts); err != nil {
		t.Fatalf("failed to sign file: %v", err)
	}

	sigFile := testFile + ".asc"
	if _, err := os.Stat(sigFile); os.IsNotExist(err) {
		t.Fatalf("signature file not created")
	}
}

func TestGoPGPSigner_GetOutputPath(t *testing.T) {
	signer := &GoPGPSigner{}

	tests := []struct {
		name     string
		filePath string
		opts     SignOptions
		expected string
	}{
		{
			name:     "detached armor",
			filePath: "/path/to/file.txt",
			opts:     SignOptions{Armor: true, DetachSign: true},
			expected: "/path/to/file.txt.asc",
		},
		{
			name:     "detached binary",
			filePath: "/path/to/file.txt",
			opts:     SignOptions{Armor: false, DetachSign: true},
			expected: "/path/to/file.txt.sig",
		},
		{
			name:     "clear sign",
			filePath: "/path/to/file.txt",
			opts:     SignOptions{ClearSign: true},
			expected: "/path/to/file.txt.asc",
		},
		{
			name:     "inline armor",
			filePath: "/path/to/file.txt",
			opts:     SignOptions{Armor: true},
			expected: "/path/to/file.txt.asc",
		},
		{
			name:     "inline binary",
			filePath: "/path/to/file.txt",
			opts:     SignOptions{Armor: false},
			expected: "/path/to/file.txt.gpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := signer.getOutputPath(tt.filePath, tt.opts)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
