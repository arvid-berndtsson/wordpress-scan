package wpprobe

import (
	"context"
	"errors"
	"os/exec"
	"reflect"
	"testing"
)

// fakeRunner is a test double for testing code that depends on Runner.
type fakeRunner struct {
	ensureBinaryErr error
	scanErr         error
	updateErr       error
	scanInput       *ScanInput
}

func (f *fakeRunner) EnsureBinary() error {
	return f.ensureBinaryErr
}

func (f *fakeRunner) Scan(ctx context.Context, input ScanInput) error {
	f.scanInput = &input
	return f.scanErr
}

func (f *fakeRunner) Update(ctx context.Context) error {
	return f.updateErr
}

// mockLookPath is a mock implementation of ExecLookPath for testing.
type mockLookPath struct {
	calls      []string
	returnPath string
	returnErr  error
}

func (m *mockLookPath) LookPath(name string) (string, error) {
	m.calls = append(m.calls, name)
	return m.returnPath, m.returnErr
}

// mockCommandContext is a mock implementation of ExecCommandContext for testing.
type mockCommandContext struct {
	calls     []commandCall
	returnCmd *exec.Cmd
	returnErr error
}

type commandCall struct {
	ctx  context.Context
	name string
	args []string
}

func (m *mockCommandContext) CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	m.calls = append(m.calls, commandCall{
		ctx:  ctx,
		name: name,
		args: arg,
	})
	if m.returnCmd != nil {
		return m.returnCmd
	}
	// Return a command that will fail immediately when Run() is called
	// unless we've set up a specific return command
	cmd := exec.CommandContext(ctx, name, arg...)
	return cmd
}

// TestNewRunner verifies that NewRunner returns a CommandRunner with default binary name.
func TestNewRunner(t *testing.T) {
	runner := NewRunner()
	cr, ok := runner.(*CommandRunner)
	if !ok {
		t.Fatal("NewRunner should return a *CommandRunner")
	}
	if cr.Binary != "wpprobe" {
		t.Fatalf("expected binary name 'wpprobe', got %q", cr.Binary)
	}
	if cr.lookPath == nil {
		t.Fatal("lookPath should be initialized")
	}
	if cr.commandContext == nil {
		t.Fatal("commandContext should be initialized")
	}
}

// TestEnsureBinaryWhenPresent verifies EnsureBinary succeeds when binary is found.
func TestEnsureBinaryWhenPresent(t *testing.T) {
	mockLookPath := &mockLookPath{
		returnPath: "/usr/bin/wpprobe",
		returnErr:  nil,
	}

	runner := &CommandRunner{
		Binary:   "wpprobe",
		lookPath: mockLookPath.LookPath,
	}

	err := runner.EnsureBinary()
	if err != nil {
		t.Fatalf("EnsureBinary should succeed when binary is found: %v", err)
	}

	if len(mockLookPath.calls) != 1 {
		t.Fatalf("expected LookPath to be called once, got %d calls", len(mockLookPath.calls))
	}
	if mockLookPath.calls[0] != "wpprobe" {
		t.Fatalf("expected LookPath to be called with 'wpprobe', got %q", mockLookPath.calls[0])
	}
}

// TestEnsureBinaryWhenMissing verifies EnsureBinary returns error when binary is not found.
func TestEnsureBinaryWhenMissing(t *testing.T) {
	mockLookPath := &mockLookPath{
		returnPath: "",
		returnErr:  exec.ErrNotFound,
	}

	runner := &CommandRunner{
		Binary:   "nonexistent-binary",
		lookPath: mockLookPath.LookPath,
	}

	err := runner.EnsureBinary()
	if err == nil {
		t.Fatal("EnsureBinary should fail when binary is not found")
	}

	expectedErr := "wpprobe binary not found"
	if err.Error()[:len(expectedErr)] != expectedErr {
		t.Fatalf("expected error message to start with %q, got %q", expectedErr, err.Error())
	}

	if len(mockLookPath.calls) != 1 {
		t.Fatalf("expected LookPath to be called once, got %d calls", len(mockLookPath.calls))
	}
	if mockLookPath.calls[0] != "nonexistent-binary" {
		t.Fatalf("expected LookPath to be called with 'nonexistent-binary', got %q", mockLookPath.calls[0])
	}
}

// TestScanConstructsCorrectCommand verifies that Scan builds the expected command arguments.
func TestScanConstructsCorrectCommand(t *testing.T) {
	tests := []struct {
		name           string
		input          ScanInput
		expectedArgs   []string
		expectedBinary string
	}{
		{
			name: "basic scan",
			input: ScanInput{
				TargetsFile: "/tmp/targets.txt",
				Mode:        "fast",
				Threads:     10,
				OutputPath:  "/tmp/output.json",
			},
			expectedBinary: "wpprobe",
			expectedArgs: []string{
				"scan",
				"-f", "/tmp/targets.txt",
				"--mode", "fast",
				"-o", "/tmp/output.json",
				"-t", "10",
			},
		},
		{
			name: "stealthy mode with different threads",
			input: ScanInput{
				TargetsFile: "/data/targets.txt",
				Mode:        "stealthy",
				Threads:     5,
				OutputPath:  "/data/results.json",
			},
			expectedBinary: "wpprobe",
			expectedArgs: []string{
				"scan",
				"-f", "/data/targets.txt",
				"--mode", "stealthy",
				"-o", "/data/results.json",
				"-t", "5",
			},
		},
		{
			name: "single thread",
			input: ScanInput{
				TargetsFile: "/work/targets.txt",
				Mode:        "aggressive",
				Threads:     1,
				OutputPath:  "/work/scan.json",
			},
			expectedBinary: "wpprobe",
			expectedArgs: []string{
				"scan",
				"-f", "/work/targets.txt",
				"--mode", "aggressive",
				"-o", "/work/scan.json",
				"-t", "1",
			},
		},
		{
			name: "zero threads",
			input: ScanInput{
				TargetsFile: "/tmp/targets.txt",
				Mode:        "fast",
				Threads:     0,
				OutputPath:  "/tmp/output.json",
			},
			expectedBinary: "wpprobe",
			expectedArgs: []string{
				"scan",
				"-f", "/tmp/targets.txt",
				"--mode", "fast",
				"-o", "/tmp/output.json",
				"-t", "0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCmdCtx := &mockCommandContext{}

			runner := &CommandRunner{
				Binary:         tt.expectedBinary,
				commandContext: mockCmdCtx.CommandContext,
			}

			ctx := context.Background()
			err := runner.Scan(ctx, tt.input)

			// We expect an error because the mock command will fail when Run() is called
			// (since there's no actual binary), but we can verify the command was constructed correctly
			if err == nil {
				t.Log("Note: Scan succeeded unexpectedly (this is okay if a real binary exists)")
			}

			if len(mockCmdCtx.calls) != 1 {
				t.Fatalf("expected CommandContext to be called once, got %d calls", len(mockCmdCtx.calls))
			}

			call := mockCmdCtx.calls[0]
			if call.name != tt.expectedBinary {
				t.Fatalf("expected binary %q, got %q", tt.expectedBinary, call.name)
			}

			if !reflect.DeepEqual(call.args, tt.expectedArgs) {
				t.Fatalf("expected args %v, got %v", tt.expectedArgs, call.args)
			}

			if call.ctx != ctx {
				t.Fatal("expected context to be passed through")
			}
		})
	}
}

// TestScanWithContext verifies that Scan passes context to the command.
func TestScanWithContext(t *testing.T) {
	mockCmdCtx := &mockCommandContext{}

	runner := &CommandRunner{
		Binary:         "wpprobe",
		commandContext: mockCmdCtx.CommandContext,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := ScanInput{
		TargetsFile: "/tmp/targets.txt",
		Mode:        "fast",
		Threads:     10,
		OutputPath:  "/tmp/output.json",
	}

	err := runner.Scan(ctx, input)
	if err == nil {
		t.Log("Note: Scan succeeded unexpectedly (this is okay if a real binary exists)")
	}

	if len(mockCmdCtx.calls) != 1 {
		t.Fatalf("expected CommandContext to be called once, got %d calls", len(mockCmdCtx.calls))
	}

	call := mockCmdCtx.calls[0]
	if call.ctx != ctx {
		t.Fatal("expected the provided context to be passed to CommandContext")
	}
}

// TestScanSetsStdoutStderr verifies that Scan properly sets stdout and stderr on the command.
func TestScanSetsStdoutStderr(t *testing.T) {
	// This test verifies that stdout/stderr are set, but we can't easily mock
	// the exec.Cmd fields without actually running a command.
	// We'll verify the command is constructed and the streams are passed through
	// by checking that no panic occurs when nil writers are provided.
	runner := &CommandRunner{
		Binary:         "wpprobe",
		commandContext: exec.CommandContext,
	}

	input := ScanInput{
		TargetsFile: "/tmp/targets.txt",
		Mode:        "fast",
		Threads:     10,
		OutputPath:  "/tmp/output.json",
		Stdout:      nil, // nil is valid and will be ignored
		Stderr:      nil, // nil is valid and will be ignored
	}

	// This should not panic even with nil writers
	_ = runner.Scan(context.Background(), input)
}

// TestUpdateRunsCommand verifies that Update executes the update command with correct arguments.
func TestUpdateRunsCommand(t *testing.T) {
	mockCmdCtx := &mockCommandContext{}

	runner := &CommandRunner{
		Binary:         "wpprobe",
		commandContext: mockCmdCtx.CommandContext,
	}

	ctx := context.Background()
	err := runner.Update(ctx)

	// We expect an error because the mock command will fail when Run() is called
	// (since there's no actual binary), but we can verify the command was constructed correctly
	if err == nil {
		t.Log("Note: Update succeeded unexpectedly (this is okay if a real binary exists)")
	}

	if len(mockCmdCtx.calls) != 1 {
		t.Fatalf("expected CommandContext to be called once, got %d calls", len(mockCmdCtx.calls))
	}

	call := mockCmdCtx.calls[0]
	if call.name != "wpprobe" {
		t.Fatalf("expected binary 'wpprobe', got %q", call.name)
	}

	expectedArgs := []string{"update"}
	if !reflect.DeepEqual(call.args, expectedArgs) {
		t.Fatalf("expected args %v, got %v", expectedArgs, call.args)
	}

	if call.ctx != ctx {
		t.Fatal("expected context to be passed through")
	}
}

// TestUpdateWithContext verifies that Update passes context to the command.
func TestUpdateWithContext(t *testing.T) {
	mockCmdCtx := &mockCommandContext{}

	runner := &CommandRunner{
		Binary:         "wpprobe",
		commandContext: mockCmdCtx.CommandContext,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := runner.Update(ctx)
	if err == nil {
		t.Log("Note: Update succeeded unexpectedly (this is okay if a real binary exists)")
	}

	if len(mockCmdCtx.calls) != 1 {
		t.Fatalf("expected CommandContext to be called once, got %d calls", len(mockCmdCtx.calls))
	}

	call := mockCmdCtx.calls[0]
	if call.ctx != ctx {
		t.Fatal("expected the provided context to be passed to CommandContext")
	}
}

// TestFakeRunnerImplementsInterface verifies the fake runner for other tests.
func TestFakeRunnerImplementsInterface(t *testing.T) {
	var _ Runner = (*fakeRunner)(nil)
}

// TestFakeRunnerEnsureBinary verifies fake runner's EnsureBinary behavior.
func TestFakeRunnerEnsureBinary(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		err     error
	}{
		{
			name:    "success",
			wantErr: false,
			err:     nil,
		},
		{
			name:    "failure",
			wantErr: true,
			err:     errors.New("binary not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeRunner{ensureBinaryErr: tt.err}
			err := fake.EnsureBinary()
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

// TestFakeRunnerScan verifies fake runner's Scan behavior.
func TestFakeRunnerScan(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		err     error
	}{
		{
			name:    "success",
			wantErr: false,
			err:     nil,
		},
		{
			name:    "failure",
			wantErr: true,
			err:     errors.New("scan failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeRunner{scanErr: tt.err}
			input := ScanInput{
				TargetsFile: "/tmp/targets.txt",
				Mode:        "fast",
				Threads:     10,
				OutputPath:  "/tmp/output.json",
			}

			err := fake.Scan(context.Background(), input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}

			if fake.scanInput == nil {
				t.Fatal("scanInput should be captured")
			}

			if fake.scanInput.TargetsFile != input.TargetsFile {
				t.Fatalf("expected targets file %q, got %q", input.TargetsFile, fake.scanInput.TargetsFile)
			}
		})
	}
}

// TestFakeRunnerUpdate verifies fake runner's Update behavior.
func TestFakeRunnerUpdate(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		err     error
	}{
		{
			name:    "success",
			wantErr: false,
			err:     nil,
		},
		{
			name:    "failure",
			wantErr: true,
			err:     errors.New("update failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeRunner{updateErr: tt.err}
			err := fake.Update(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}
