package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
)

// Global variables for application metadata.
var (
	Version   string
	Revision  string
	GoVersion = runtime.Version()
	StartTime = time.Now()
)

// SignerBackend defines the available signer backends.
type SignerBackend string

const (
	BackendGoPGP  SignerBackend = "gopgp"
	BackendGnuPG  SignerBackend = "gnupg"
	DefaultBackend              = BackendGoPGP
)

// ActionInputs holds the input parameters for the GPG signing action.
type ActionInputs struct {
	PrivateKey string `arg:"--private-key,env:PRIVATE_KEY,required" help:"Private GPG key used for signing"`
	Passphrase string `arg:"--passphrase,env:PASSPHRASE" help:"Passphrase for the GPG key"`
	Armor      bool   `arg:"--armor,env:ARMOR" default:"true" help:"Create ASCII armored output"`
	DetachSign bool   `arg:"--detach-sign,env:DETACH_SIGN" default:"false" help:"Make a detached signature"`
	ClearSign  bool   `arg:"--clear-sign,env:CLEAR_SIGN" default:"false" help:"Make a clear text signature"`
	Files      string `arg:"--files,env:FILES,required" help:"List of files to sign (glob patterns, newline separated)"`
	Excludes   string `arg:"--excludes,env:EXCLUDES" help:"List of files to exclude (glob patterns, newline separated)"`
	WorkDir    string `arg:"--workdir,env:WORKDIR" help:"Working directory for file operations"`
	Backend    string `arg:"--backend,env:BACKEND" default:"gopgp" help:"Signer backend: gopgp (pure Go, default) or gnupg (system GPG)"`
	LogLevel   string `arg:"--log-level,env:LOG_LEVEL" default:"info" help:"Log level: debug, info, warn, error"`
}

// Version returns a formatted string with application version details.
func (ActionInputs) Version() string {
	return fmt.Sprintf("Version: %s %s\nBuildTime: %s\n%s\n", Revision, Version, StartTime.Format("2006-01-02"), GoVersion)
}

func main() {
	var args ActionInputs
	arg.MustParse(&args)

	log := setupLogger(args.LogLevel)

	if err := run(args, nil, nil, log); err != nil {
		log.Error("Action failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// setupLogger creates a new slog.Logger with the specified log level.
func setupLogger(level string) *slog.Logger {
	logLevel := stringToLogLevel(level)
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	return slog.New(handler)
}

// stringToLogLevel converts a string log level to slog.Level.
func stringToLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// run executes the main logic of the GPG signing action.
func run(args ActionInputs, signer Signer, finder FileFinder, log *slog.Logger) error {
	// Use a no-op logger if none provided (for testing)
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	log.Debug("Starting PGP Sign Artifact Action",
		slog.String("backend", args.Backend),
		slog.Bool("armor", args.Armor),
		slog.Bool("detach_sign", args.DetachSign),
		slog.Bool("clear_sign", args.ClearSign),
		slog.Bool("has_passphrase", args.Passphrase != ""),
	)

	// Create signer if not provided (for testing)
	if signer == nil {
		log.Debug("Creating signer", slog.String("backend", args.Backend))
		var err error
		signer, err = NewSigner(SignerBackend(args.Backend), args.PrivateKey, args.Passphrase)
		if err != nil {
			return fmt.Errorf("failed to create signer: %w", err)
		}
		log.Debug("Signer created successfully")
	}

	// Create file finder if not provided (for testing)
	if finder == nil {
		finder = &DefaultFileFinder{}
	}

	workDir := args.WorkDir
	if workDir == "" {
		workDir = os.Getenv("GITHUB_WORKSPACE")
	}
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}
	log.Debug("Working directory resolved", slog.String("workdir", workDir))

	opts := SignOptions{
		Armor:      args.Armor,
		DetachSign: args.DetachSign,
		ClearSign:  args.ClearSign,
	}

	patterns := parseMultilineInput(args.Files)
	excludes := parseMultilineInput(args.Excludes)

	log.Debug("File patterns configured",
		slog.Any("patterns", patterns),
		slog.Any("excludes", excludes),
	)

	files, err := finder.FindFiles(workDir, patterns, excludes)
	if err != nil {
		return fmt.Errorf("failed to find files: %w", err)
	}

	log.Debug("Files matched", slog.Int("count", len(files)))

	if len(files) == 0 {
		log.Warn("No files matched the specified patterns")
		return nil
	}

	log.Info("Starting to sign files", slog.Int("count", len(files)))

	for _, file := range files {
		log.Info("Signing file", slog.String("file", file))
		if err := signer.SignFile(file, opts); err != nil {
			return fmt.Errorf("failed to sign file %s: %w", file, err)
		}
		log.Debug("File signed successfully", slog.String("file", file))
	}

	log.Info("Successfully signed all files", slog.Int("count", len(files)))
	return nil
}

// parseMultilineInput splits a multiline string into a slice of trimmed, non-empty strings.
func parseMultilineInput(input string) []string {
	var result []string
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

// setActionOutput writes an output value for GitHub Actions.
func setActionOutput(name, value string) {
	outputFile := os.Getenv("GITHUB_OUTPUT")
	if outputFile == "" {
		fmt.Printf("::set-output name=%s::%s\n", name, value)
		return
	}

	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to open GITHUB_OUTPUT file: %v\n", err)
		fmt.Printf("::set-output name=%s::%s\n", name, value)
		return
	}
	defer f.Close()

	if _, err := io.WriteString(f, fmt.Sprintf("%s=%s\n", name, value)); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write to GITHUB_OUTPUT file: %v\n", err)
		fmt.Printf("::set-output name=%s::%s\n", name, value)
	}
}
