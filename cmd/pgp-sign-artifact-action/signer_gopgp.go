package main

import (
	"fmt"
	"os"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

// GoPGPSigner implements Signer using the gopenpgp library (pure Go).
type GoPGPSigner struct {
	privateKey *crypto.Key
}

// NewGoPGPSigner creates a new GoPGPSigner with the provided private key and passphrase.
func NewGoPGPSigner(armoredKey, passphrase string) (*GoPGPSigner, error) {
	key, err := crypto.NewKeyFromArmored(armoredKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	if !key.IsPrivate() {
		return nil, fmt.Errorf("provided key is not a private key")
	}

	// Unlock the key if passphrase is provided
	if passphrase != "" {
		key, err = key.Unlock([]byte(passphrase))
		if err != nil {
			return nil, fmt.Errorf("failed to unlock private key: %w", err)
		}
	}

	// Check if key is locked (needs passphrase but none provided)
	locked, err := key.IsLocked()
	if err != nil {
		return nil, fmt.Errorf("failed to check key lock status: %w", err)
	}
	if locked {
		return nil, fmt.Errorf("private key is locked but no passphrase provided")
	}

	return &GoPGPSigner{
		privateKey: key,
	}, nil
}

// SignFile signs a file using gopenpgp.
func (s *GoPGPSigner) SignFile(filePath string, opts SignOptions) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	pgp := crypto.PGP()

	var signature []byte

	if opts.DetachSign {
		signature, err = s.createDetachedSignature(pgp, data, opts.Armor)
	} else if opts.ClearSign {
		signature, err = s.createClearSignature(pgp, data)
	} else {
		signature, err = s.createInlineSignature(pgp, data, opts.Armor)
	}

	if err != nil {
		return err
	}

	outputPath := s.getOutputPath(filePath, opts)
	if err := os.WriteFile(outputPath, signature, 0644); err != nil {
		return fmt.Errorf("failed to write signature: %w", err)
	}

	return nil
}

// createDetachedSignature creates a detached signature for the data.
func (s *GoPGPSigner) createDetachedSignature(pgp *crypto.PGPHandle, data []byte, armor bool) ([]byte, error) {
	signHandle, err := pgp.Sign().
		SigningKey(s.privateKey).
		Detached().
		New()
	if err != nil {
		return nil, fmt.Errorf("failed to create signing handle: %w", err)
	}

	encoding := crypto.Bytes
	if armor {
		encoding = crypto.Armor
	}

	sig, err := signHandle.Sign(data, encoding)
	if err != nil {
		return nil, fmt.Errorf("failed to create detached signature: %w", err)
	}

	return sig, nil
}

// createClearSignature creates a clear-text signature.
func (s *GoPGPSigner) createClearSignature(pgp *crypto.PGPHandle, data []byte) ([]byte, error) {
	signHandle, err := pgp.Sign().
		SigningKey(s.privateKey).
		New()
	if err != nil {
		return nil, fmt.Errorf("failed to create signing handle: %w", err)
	}

	clearSigned, err := signHandle.SignCleartext(data)
	if err != nil {
		return nil, fmt.Errorf("failed to create clear signature: %w", err)
	}

	return clearSigned, nil
}

// createInlineSignature creates an inline (attached) signature.
func (s *GoPGPSigner) createInlineSignature(pgp *crypto.PGPHandle, data []byte, armor bool) ([]byte, error) {
	signHandle, err := pgp.Sign().
		SigningKey(s.privateKey).
		New()
	if err != nil {
		return nil, fmt.Errorf("failed to create signing handle: %w", err)
	}

	encoding := crypto.Bytes
	if armor {
		encoding = crypto.Armor
	}

	signed, err := signHandle.Sign(data, encoding)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature: %w", err)
	}

	return signed, nil
}

// getOutputPath determines the output file path based on signing options.
func (s *GoPGPSigner) getOutputPath(filePath string, opts SignOptions) string {
	ext := getOutputExtension(opts)

	if opts.DetachSign {
		return filePath + ext
	}

	if opts.ClearSign {
		return filePath + ".asc"
	}

	return filePath + ext
}
