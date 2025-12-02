package wpprobe

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
)

// Runner defines the operations needed to drive wpprobe.
type Runner interface {
	EnsureBinary() error
	Scan(ctx context.Context, input ScanInput) error
	Update(ctx context.Context) error
}

// CommandRunner executes the real wpprobe binary present on the worker.
type CommandRunner struct {
	Binary string
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
	return &CommandRunner{Binary: "wpprobe"}
}

// EnsureBinary verifies that the wpprobe binary is discoverable on PATH.
func (r *CommandRunner) EnsureBinary() error {
	_, err := exec.LookPath(r.Binary)
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

	// Binary path is controlled by the application and args are constructed
	// programmatically from validated inputs, making command injection impossible.
	cmd := exec.CommandContext(ctx, r.Binary, args...) // #nosec G204
	cmd.Stdout = input.Stdout
	cmd.Stderr = input.Stderr

	return cmd.Run()
}

// Update runs `wpprobe update` to refresh vulnerability databases.
func (r *CommandRunner) Update(ctx context.Context) error {
	// Binary path is controlled by the application and the argument is a constant string,
	// making command injection impossible.
	cmd := exec.CommandContext(ctx, r.Binary, "update") // #nosec G204
	return cmd.Run()
}
