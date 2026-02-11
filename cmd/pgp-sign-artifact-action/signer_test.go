package main

import (
	"testing"
)

func TestGetOutputExtension(t *testing.T) {
	tests := []struct {
		name     string
		opts     SignOptions
		expected string
	}{
		{
			name:     "armor detach sign",
			opts:     SignOptions{Armor: true, DetachSign: true},
			expected: ".asc",
		},
		{
			name:     "binary detach sign",
			opts:     SignOptions{Armor: false, DetachSign: true},
			expected: ".sig",
		},
		{
			name:     "armor inline sign",
			opts:     SignOptions{Armor: true},
			expected: ".asc",
		},
		{
			name:     "binary inline sign",
			opts:     SignOptions{Armor: false},
			expected: ".gpg",
		},
		{
			name:     "clear sign",
			opts:     SignOptions{ClearSign: true},
			expected: ".asc",
		},
		{
			name:     "clear sign ignores armor false",
			opts:     SignOptions{Armor: false, ClearSign: true},
			expected: ".asc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOutputExtension(tt.opts)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNewSigner_InvalidBackend(t *testing.T) {
	_, err := NewSigner("invalid", "key", "")
	if err == nil {
		t.Error("expected error for invalid backend")
	}
}
