# PGP Sign Artifact Action

**Sign artifacts or files with PGP signatures using either a pure Go implementation or system GnuPG.**

[![GitHub release](https://img.shields.io/github/release/cbrgm/pgp-sign-artifact-action.svg)](https://github.com/cbrgm/pgp-sign-artifact-action)
[![Go Report Card](https://goreportcard.com/badge/github.com/cbrgm/pgp-sign-artifact-action)](https://goreportcard.com/report/github.com/cbrgm/pgp-sign-artifact-action)
[![go-lint-test](https://github.com/cbrgm/pgp-sign-artifact-action/actions/workflows/go-lint-test.yml/badge.svg)](https://github.com/cbrgm/pgp-sign-artifact-action/actions/workflows/go-lint-test.yml)
[![go-binaries](https://github.com/cbrgm/pgp-sign-artifact-action/actions/workflows/go-binaries.yml/badge.svg)](https://github.com/cbrgm/pgp-sign-artifact-action/actions/workflows/go-binaries.yml)
[![container](https://github.com/cbrgm/pgp-sign-artifact-action/actions/workflows/container.yml/badge.svg)](https://github.com/cbrgm/pgp-sign-artifact-action/actions/workflows/container.yml)

- [PGP Sign Artifact Action](#pgp-sign-artifact-action)
  - [Inputs](#inputs)
  - [Workflow Usage](#workflow-usage)
    - [Basic Example: Sign Release Artifacts](#basic-example-sign-release-artifacts)
    - [Example: Sign with Exclusions](#example-sign-with-exclusions)
    - [Example: Clear Sign a Changelog](#example-clear-sign-a-changelog)
    - [Example: Binary Signatures](#example-binary-signatures)
    - [Example: Using GnuPG Backend](#example-using-gnupg-backend)
    - [Example: Debug Logging](#example-debug-logging)
    - [Example: Upload Signatures as Release Assets](#example-upload-signatures-as-release-assets)
  - [Choosing a Backend](#choosing-a-backend)
  - [CLI Usage (Standalone Binary)](#cli-usage-standalone-binary)
    - [Installation](#installation)
    - [CLI Arguments](#cli-arguments)
    - [CLI Examples](#cli-examples)
  - [Generating GPG Keys](#generating-gpg-keys)
  - [Output Files](#output-files)
  - [Verifying Signatures](#verifying-signatures)
  - [Troubleshooting](#troubleshooting)
  - [Local Development](#local-development)
  - [Contributing & License](#contributing--license)

## Inputs

- `private_key`: **Required** - The private GPG key used for signing (armored format). Store this securely in GitHub Secrets.
- `passphrase`: **Optional** - Passphrase for the GPG key if it is encrypted.
- `armor`: **Optional** - Create ASCII armored output (`.asc` extension). Set to `false` for binary output (`.sig` or `.gpg` extension). Default is `true`.
- `detach_sign`: **Optional** - Make a detached signature. Creates a separate signature file alongside the original file. Default is `false`.
- `clear_sign`: **Optional** - Make a clear text signature. Creates a clear-signed message with original content and signature in one file. Default is `false`.
- `files`: **Required** - List of files to sign. Supports glob patterns (e.g., `dist/*`, `*.tar.gz`, `**/*.bin`), with multiple patterns separated by newlines.
- `excludes`: **Optional** - List of files to exclude from signing. Supports glob patterns, with multiple patterns separated by newlines.
- `backend`: **Optional** - Signer backend to use: `gopgp` (pure Go, no dependencies) or `gnupg` (system GPG). Default is `gopgp`.
- `log_level`: **Optional** - Log level for output verbosity: `debug`, `info`, `warn`, or `error`. Default is `info`.

## Workflow Usage

### Basic Example: Sign Release Artifacts

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Build
        run: |
          mkdir -p dist
          # Your build commands here
          echo "binary content" > dist/myapp-linux-amd64
          echo "binary content" > dist/myapp-darwin-amd64

      - name: Sign Artifacts
        uses: cbrgm/pgp-sign-artifact-action@v1
        with:
          private_key: ${{ secrets.GPG_PRIVATE_KEY }}
          passphrase: ${{ secrets.GPG_PASSPHRASE }}
          detach_sign: true
          files: |
            dist/*
```

### Example: Sign with Exclusions

```yaml
- name: Sign Artifacts
  uses: cbrgm/pgp-sign-artifact-action@v1
  with:
    private_key: ${{ secrets.GPG_PRIVATE_KEY }}
    passphrase: ${{ secrets.GPG_PASSPHRASE }}
    detach_sign: true
    files: |
      dist/*
    excludes: |
      *.sha256
      *.md5
```

### Example: Clear Sign a Changelog

```yaml
- name: Sign Changelog
  uses: cbrgm/pgp-sign-artifact-action@v1
  with:
    private_key: ${{ secrets.GPG_PRIVATE_KEY }}
    clear_sign: true
    files: |
      CHANGELOG.md
```

### Example: Binary Signatures

```yaml
- name: Sign with Binary Output
  uses: cbrgm/pgp-sign-artifact-action@v1
  with:
    private_key: ${{ secrets.GPG_PRIVATE_KEY }}
    armor: false
    detach_sign: true
    files: |
      dist/*.tar.gz
```

### Example: Using GnuPG Backend

```yaml
- name: Sign with System GPG
  uses: cbrgm/pgp-sign-artifact-action@v1
  with:
    private_key: ${{ secrets.GPG_PRIVATE_KEY }}
    passphrase: ${{ secrets.GPG_PASSPHRASE }}
    backend: gnupg
    detach_sign: true
    files: |
      dist/*
```

### Example: Debug Logging

```yaml
- name: Sign with Debug Output
  uses: cbrgm/pgp-sign-artifact-action@v1
  with:
    private_key: ${{ secrets.GPG_PRIVATE_KEY }}
    passphrase: ${{ secrets.GPG_PASSPHRASE }}
    detach_sign: true
    log_level: debug
    files: |
      dist/*
```

### Example: Upload Signatures as Release Assets

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Build
        run: |
          mkdir -p dist
          # Your build commands here

      - name: Sign Artifacts
        uses: cbrgm/pgp-sign-artifact-action@v1
        with:
          private_key: ${{ secrets.GPG_PRIVATE_KEY }}
          passphrase: ${{ secrets.GPG_PASSPHRASE }}
          detach_sign: true
          files: |
            dist/*

      - name: Upload Release Assets
        uses: softprops/action-gh-release@v1
        with:
          files: |
            dist/*
            dist/*.asc
```

## Choosing a Backend

| Backend | Pros | Cons | Use When |
|---------|------|------|----------|
| `gopgp` (default) | No external dependencies, runs anywhere | Pure Go implementation | Default choice, CI environments |
| `gnupg` | Uses system GPG, supports hardware tokens | Requires `gpg` installed | Need GPG agent, smart cards, or specific GPG features |

## CLI Usage (Standalone Binary)

This action can also be run as a standalone CLI tool outside of GitHub Actions.

### Installation

```bash
# Build from source
git clone https://github.com/cbrgm/pgp-sign-artifact-action.git
cd pgp-sign-artifact-action
make build

# Or install directly with Go
go install github.com/cbrgm/pgp-sign-artifact-action/cmd/pgp-sign-artifact-action@latest
```

### CLI Arguments

| Argument | Environment Variable | Required | Default | Description |
|----------|---------------------|----------|---------|-------------|
| `--private-key` | `PRIVATE_KEY` | Yes | - | Private GPG key (armored format) |
| `--passphrase` | `PASSPHRASE` | No | - | Passphrase for the key |
| `--armor` | `ARMOR` | No | `true` | ASCII armored output |
| `--detach-sign` | `DETACH_SIGN` | No | `false` | Create detached signature |
| `--clear-sign` | `CLEAR_SIGN` | No | `false` | Create clear-text signature |
| `--files` | `FILES` | Yes | - | Files to sign (glob patterns, newline-separated) |
| `--excludes` | `EXCLUDES` | No | - | Files to exclude (glob patterns) |
| `--workdir` | `WORKDIR` | No | Current dir | Working directory |
| `--backend` | `BACKEND` | No | `gopgp` | Signer backend |
| `--log-level` | `LOG_LEVEL` | No | `info` | Log level |

### CLI Examples

**Sign files with detached signatures:**

```bash
pgp-sign-artifact-action \
  --private-key "$(cat private-key.asc)" \
  --passphrase "my-passphrase" \
  --detach-sign \
  --files "dist/*"
```

**Sign using environment variables:**

```bash
export PRIVATE_KEY="$(cat private-key.asc)"
export PASSPHRASE="my-passphrase"

pgp-sign-artifact-action \
  --detach-sign \
  --files "dist/*.tar.gz"
```

**Sign with debug logging and exclusions:**

```bash
pgp-sign-artifact-action \
  --private-key "$(cat private-key.asc)" \
  --detach-sign \
  --log-level debug \
  --files "dist/*" \
  --excludes "*.sha256
*.md5"
```

**Use system GnuPG backend:**

```bash
pgp-sign-artifact-action \
  --private-key "$(cat private-key.asc)" \
  --passphrase "my-passphrase" \
  --backend gnupg \
  --detach-sign \
  --files "release/*"
```

## Generating GPG Keys

**Generate a new key:**

```bash
# Generate a new key (follow the prompts)
gpg --full-generate-key

# Export the private key (armored format) for signing
gpg --armor --export-secret-keys your-email@example.com > private-key.asc

# Store the content of private-key.asc as a GitHub Secret (GPG_PRIVATE_KEY)
```

**Export public key for verification:**

```bash
# Export public key for recipients to verify signatures
gpg --armor --export your-email@example.com > public-key.asc

# Recipients import the public key
gpg --import public-key.asc
```

## Output Files

Signatures are written alongside the original files. The output filename and extension depend on the signing options:

| Mode | Armor | Input File | Output File |
|------|-------|------------|-------------|
| `detach_sign: true` | `true` | `file.tar.gz` | `file.tar.gz.asc` |
| `detach_sign: true` | `false` | `file.tar.gz` | `file.tar.gz.sig` |
| `clear_sign: true` | (always armored) | `file.txt` | `file.txt.asc` |
| Neither (inline) | `true` | `file.txt` | `file.txt.asc` |
| Neither (inline) | `false` | `file.txt` | `file.txt.gpg` |

**Signature types:**

| Option | Description |
|--------|-------------|
| `detach_sign: true` | Separate signature file - original file unchanged, signature in new file |
| `clear_sign: true` | Clear-text signature - human-readable original with signature appended |
| Neither | Inline signature - signed message (requires decryption to read) |

## Verifying Signatures

Recipients can verify signatures using:

```bash
# For detached signatures
gpg --verify file.asc file

# For clear-signed files
gpg --verify file.asc

# For inline signatures
gpg --decrypt file.asc
```

## Troubleshooting

**Common errors:**

| Error | Cause | Solution |
|-------|-------|----------|
| `failed to parse private key` | Invalid key format | Ensure key is ASCII armored (starts with `-----BEGIN PGP PRIVATE KEY BLOCK-----`) |
| `private key is locked but no passphrase provided` | Key requires passphrase | Provide the `passphrase` input |
| `failed to unlock private key` | Wrong passphrase | Verify the passphrase is correct |
| `no files matched the specified patterns` | Glob pattern didn't match | Check patterns and working directory; use `log_level: debug` |
| `gpg command failed` (gnupg backend) | System GPG issue | Ensure `gpg` is installed and accessible |

**Debug tips:**

1. **Enable debug logging** to see configuration details:
   ```yaml
   log_level: debug
   ```

2. **Test your glob patterns** locally:
   ```bash
   ls dist/*        # Test your pattern
   ```

3. **Verify your key** is valid:
   ```bash
   gpg --show-keys private-key.asc
   ```

## Local Development

You can build this action from source using `Go`:

```bash
make build
```

Run tests:

```bash
make test
```

## Contributing & License

We welcome and value your contributions to this project! 👍 If you're interested in making improvements or adding features, please refer to our [Contributing Guide](https://github.com/cbrgm/pgp-sign-artifact-action/blob/main/CONTRIBUTING.md). This guide provides comprehensive instructions on how to submit changes, set up your development environment, and more.

Please note that this project is developed in my spare time and is available for free 🕒💻. As an open-source initiative, it is governed by the [Apache 2.0 License](https://github.com/cbrgm/pgp-sign-artifact-action/blob/main/LICENSE). This license outlines your rights and obligations when using, modifying, and distributing this software.

Your involvement, whether it's through code contributions, suggestions, or feedback, is crucial for the ongoing improvement and success of this project. Together, we can ensure it remains a useful and well-maintained resource for everyone 🌍.
