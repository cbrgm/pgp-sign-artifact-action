package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GnuPGSigner implements Signer using the system's GnuPG installation.
type GnuPGSigner struct {
	passphrase string
}

// NewGnuPGSigner creates a new GnuPGSigner and imports the private key.
func NewGnuPGSigner(armoredKey, passphrase string) (*GnuPGSigner, error) {
	if err := importGPGKey(armoredKey); err != nil {
		return nil, fmt.Errorf("failed to import GPG key: %w", err)
	}

	return &GnuPGSigner{
		passphrase: passphrase,
	}, nil
}

// importGPGKey imports a GPG key using the gpg command.
func importGPGKey(armoredKey string) error {
	cmd := exec.Command("gpg", "--batch", "--import", "-")
	cmd.Stdin = strings.NewReader(armoredKey)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// SignFile signs a file using the system's GnuPG.
func (s *GnuPGSigner) SignFile(filePath string, opts SignOptions) error {
	args := s.buildArgs(opts)
	args = append(args, filePath)

	cmd := exec.Command("gpg", args...)

	if s.passphrase != "" {
		cmd.Stdin = strings.NewReader(s.passphrase)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gpg command failed: %w", err)
	}

	return nil
}

// buildArgs constructs the GPG command arguments based on sign options.
func (s *GnuPGSigner) buildArgs(opts SignOptions) []string {
	args := []string{"--batch", "--yes"}

	if s.passphrase != "" {
		args = append(args, "--pinentry-mode", "loopback", "--passphrase-fd", "0")
	}

	if opts.Armor {
		args = append(args, "--armor")
	}

	if opts.DetachSign {
		args = append(args, "--detach-sign")
	} else if opts.ClearSign {
		args = append(args, "--clear-sign")
	} else {
		args = append(args, "--sign")
	}

	return args
}
