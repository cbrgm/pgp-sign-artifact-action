# PGP Sign Artifact Action

**Sign artifacts or any file with PGP signatures.**

[![GitHub release](https://img.shields.io/github/release/cbrgm/pgp-sign-artifact-action.svg)](https://github.com/cbrgm/pgp-sign-artifact-action)
[![Go Report Card](https://goreportcard.com/badge/github.com/cbrgm/pgp-sign-artifact-action)](https://goreportcard.com/report/github.com/cbrgm/pgp-sign-artifact-action)
[![go-lint-test](https://github.com/cbrgm/pgp-sign-artifact-action/actions/workflows/go-lint-test.yml/badge.svg)](https://github.com/cbrgm/pgp-sign-artifact-action/actions/workflows/go-lint-test.yml)
[![go-binaries](https://github.com/cbrgm/pgp-sign-artifact-action/actions/workflows/go-binaries.yml/badge.svg)](https://github.com/cbrgm/pgp-sign-artifact-action/actions/workflows/go-binaries.yml)
[![container](https://github.com/cbrgm/pgp-sign-artifact-action/actions/workflows/container.yml/badge.svg)](https://github.com/cbrgm/pgp-sign-artifact-action/actions/workflows/container.yml)

## Inputs

### `private_key`
**Required** - The private GPG key used for signing (armored format). Store this securely in GitHub Secrets.

### `passphrase`
**Optional** - Passphrase for the GPG key if it is encrypted.

### `armor`
**Optional** - Create ASCII armored output. Default: `true`.
- `true` - Output signature in ASCII armored format (`.asc` extension)
- `false` - Output signature in binary format (`.sig` or `.gpg` extension)

### `detach_sign`
**Optional** - Make a detached signature. Default: `false`.
- `true` - Creates a separate signature file alongside the original file
- `false` - Creates an inline signature

### `clear_sign`
**Optional** - Make a clear text signature. Default: `false`.
- `true` - Creates a clear-signed message (original content + signature in one file)
- `false` - Creates a standard signature

### `files`
**Required** - List of files to sign. Supports glob patterns, with multiple patterns separated by newlines.

Examples:
- `dist/*` - All files in dist directory
- `*.tar.gz` - All tar.gz files
- `**/*.bin` - All .bin files recursively

### `excludes`
**Optional** - List of files to exclude from signing. Supports glob patterns, with multiple patterns separated by newlines.

### `backend`
**Optional** - Signer backend to use. Default: `gopgp`.
- `gopgp` - Pure Go implementation using [gopenpgp](https://github.com/ProtonMail/gopenpgp) (default, no external dependencies)
- `gnupg` - System GnuPG installation (requires `gpg` to be available)

### `log_level`
**Optional** - Log level for output verbosity. Default: `info`.
- `debug` - Verbose output including configuration details (never exposes keys or passphrases)
- `info` - Standard output showing files being signed
- `warn` - Only warnings and errors
- `error` - Only errors

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

### Generate a New Key

```bash
# Generate a new key (follow the prompts)
gpg --full-generate-key

# Export the private key (armored format) for signing
gpg --armor --export-secret-keys your-email@example.com > private-key.asc

# Store the content of private-key.asc as a GitHub Secret (GPG_PRIVATE_KEY)
```

### Export Public Key for Verification

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

**Signature Types:**

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

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `failed to parse private key` | Invalid key format | Ensure key is ASCII armored (starts with `-----BEGIN PGP PRIVATE KEY BLOCK-----`) |
| `private key is locked but no passphrase provided` | Key requires passphrase | Provide the `passphrase` input |
| `failed to unlock private key` | Wrong passphrase | Verify the passphrase is correct |
| `no files matched the specified patterns` | Glob pattern didn't match | Check patterns and working directory; use `log_level: debug` |
| `gpg command failed` (gnupg backend) | System GPG issue | Ensure `gpg` is installed and accessible |

### Debug Tips

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
