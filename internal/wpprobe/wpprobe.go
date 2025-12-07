package wpprobe

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
)

// ExecLookPath is a function type for looking up executables in PATH.
// This allows us to mock exec.LookPath in tests.
type ExecLookPath func(name string) (string, error)

// ExecCommandContext is a function type for creating commands.
// This allows us to mock exec.CommandContext in tests.
type ExecCommandContext func(ctx context.Context, name string, arg ...string) *exec.Cmd

// Runner defines the operations needed to drive wpprobe.
type Runner interface {
	EnsureBinary() error
	Scan(ctx context.Context, input ScanInput) error
	Update(ctx context.Context) error
}

// CommandRunner executes the real wpprobe binary present on the worker.
type CommandRunner struct {
	Binary         string
	lookPath       ExecLookPath
	commandContext ExecCommandContext
}

// ScanInput describes a single wpprobe scan invocation.
type ScanInput struct {
	TargetsFile string
	Mode        string
	Threads     int
	OutputPath  string
	Stdout      io.Writer
	Stderr      io.Writer
}

// NewRunner returns a default command runner.
func NewRunner() Runner {
	return &CommandRunner{
		Binary:         "wpprobe",
		lookPath:       exec.LookPath,
		commandContext: exec.CommandContext,
	}
}

// EnsureBinary verifies that the wpprobe binary is discoverable on PATH.
func (r *CommandRunner) EnsureBinary() error {
	if r.lookPath == nil {
		r.lookPath = exec.LookPath
	}
	_, err := r.lookPath(r.Binary)
	if err != nil {
		return fmt.Errorf("wpprobe binary not found: %w", err)
	}
	return nil
}

// Scan executes wpprobe scan with the provided arguments.
func (r *CommandRunner) Scan(ctx context.Context, input ScanInput) error {
	args := []string{
		"scan",
		"-f", input.TargetsFile,
		"--mode", input.Mode,
		"-o", input.OutputPath,
		"-t", strconv.Itoa(input.Threads),
	}

	if r.commandContext == nil {
		r.commandContext = exec.CommandContext
	}

	// #nosec G204: Binary path is controlled by the application and args are constructed
	// programmatically from validated inputs, making command injection impossible.
	cmd := r.commandContext(ctx, r.Binary, args...)
	cmd.Stdout = input.Stdout
	cmd.Stderr = input.Stderr

	return cmd.Run()
}

// Update runs `wpprobe update` to refresh vulnerability databases.
func (r *CommandRunner) Update(ctx context.Context) error {
	if r.commandContext == nil {
		r.commandContext = exec.CommandContext
	}
	// #nosec G204: Binary path is controlled by the application and args are constructed
	// programmatically from validated inputs (here, a constant string), making command injection impossible.
	cmd := r.commandContext(ctx, r.Binary, "update")
	return cmd.Run()
}
