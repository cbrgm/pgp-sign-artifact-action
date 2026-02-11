package main

import "fmt"

// SignOptions contains the options for signing a file.
type SignOptions struct {
	Armor      bool // Create ASCII armored output
	DetachSign bool // Make a detached signature
	ClearSign  bool // Make a clear text signature
}

// Signer defines the interface for GPG signing operations.
type Signer interface {
	// SignFile signs a file and writes the signature.
	// For detached signatures, creates a .sig or .asc file.
	// For clear signatures, creates a .asc file with the clear-signed content.
	// For normal signatures, creates a .gpg or .asc file.
	SignFile(filePath string, opts SignOptions) error
}

// NewSigner creates a new Signer based on the specified backend.
func NewSigner(backend SignerBackend, privateKey, passphrase string) (Signer, error) {
	switch backend {
	case BackendGoPGP:
		return NewGoPGPSigner(privateKey, passphrase)
	case BackendGnuPG:
		return NewGnuPGSigner(privateKey, passphrase)
	default:
		return nil, fmt.Errorf("unknown signer backend: %s", backend)
	}
}

// getOutputExtension returns the appropriate file extension for the signature.
func getOutputExtension(opts SignOptions) string {
	// Clear sign always produces armored output
	if opts.ClearSign {
		return ".asc"
	}

	if opts.Armor {
		return ".asc"
	}

	if opts.DetachSign {
		return ".sig"
	}

	return ".gpg"
}
