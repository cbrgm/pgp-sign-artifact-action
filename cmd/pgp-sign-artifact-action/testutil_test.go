package main

import (
	"testing"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

// generateTestKeyArmored creates a test PGP key and returns its armored form.
func generateTestKeyArmored(t *testing.T, name, email, passphrase string) string {
	t.Helper()

	pgp := crypto.PGP()
	keyGenHandle := pgp.KeyGeneration().AddUserId(name, email).New()

	key, err := keyGenHandle.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	// If passphrase is provided, lock the key
	if passphrase != "" {
		lockedKey, err := pgp.LockKey(key, []byte(passphrase))
		if err != nil {
			t.Fatalf("failed to lock key with passphrase: %v", err)
		}
		key = lockedKey
	}

	armored, err := key.Armor()
	if err != nil {
		t.Fatalf("failed to armor key: %v", err)
	}

	return armored
}

// MockSigner implements Signer for testing.
type MockSigner struct {
	SignedFiles []string
	SignedOpts  []SignOptions
	Err         error
}

func (m *MockSigner) SignFile(filePath string, opts SignOptions) error {
	if m.Err != nil {
		return m.Err
	}
	m.SignedFiles = append(m.SignedFiles, filePath)
	m.SignedOpts = append(m.SignedOpts, opts)
	return nil
}

// MockFileFinder implements FileFinder for testing.
type MockFileFinder struct {
	Files       []string
	Err         error
	LastWorkDir string
}

func (m *MockFileFinder) FindFiles(workDir string, patterns []string, excludes []string) ([]string, error) {
	m.LastWorkDir = workDir
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Files, nil
}
