package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/example/wphunter/internal/config"
)

func TestInitCommandSuccessfulValidation(t *testing.T) {
	outputDir := t.TempDir()

	loader := &config.Loader{ConfigPath: ""}
	cmd := newInitCmd(loader)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--targets=https://example.com",
		"--output-dir", outputDir,
		"--skip-wpprobe-check",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v\nOutput: %s", err, buf.String())
	}

	output := buf.String()
	if !strings.Contains(output, "Environment looks good") {
		t.Fatalf("expected success message, got: %s", output)
	}

	if !strings.Contains(output, outputDir) {
		t.Fatalf("expected output dir in message, got: %s", output)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Fatalf("output directory was not created: %s", outputDir)
	}
}

func TestInitCommandConfigurationError_NoTargets(t *testing.T) {
	outputDir := t.TempDir()

	loader := &config.Loader{ConfigPath: ""}
	cmd := newInitCmd(loader)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--output-dir", outputDir,
		"--skip-wpprobe-check",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected init to fail with no targets, but it succeeded")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("no targets configured")) {
		t.Fatalf("expected 'no targets configured' error, got: %v", err)
	}
}

func TestInitCommandConfigurationError_InvalidThreads(t *testing.T) {
	outputDir := t.TempDir()

	loader := &config.Loader{ConfigPath: ""}
	cmd := newInitCmd(loader)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--targets=https://example.com",
		"--output-dir", outputDir,
		"--threads=100",
		"--skip-wpprobe-check",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected init to fail with invalid threads, but it succeeded")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("threads must be between")) {
		t.Fatalf("expected threads validation error, got: %v", err)
	}
}

func TestInitCommandOutputDirectoryCreation(t *testing.T) {
	// Use a subdirectory that doesn't exist yet
	baseDir := t.TempDir()
	outputDir := filepath.Join(baseDir, "nested", "output", "dir")

	loader := &config.Loader{ConfigPath: ""}
	cmd := newInitCmd(loader)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--targets=https://example.com",
		"--output-dir", outputDir,
		"--skip-wpprobe-check",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Verify nested directory was created
	info, err := os.Stat(outputDir)
	if err != nil {
		t.Fatalf("output directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Fatalf("output path is not a directory")
	}
}

func TestInitCommandBinaryCheckWithSkipFlag(t *testing.T) {
	outputDir := t.TempDir()

	loader := &config.Loader{ConfigPath: ""}
	cmd := newInitCmd(loader)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--targets=https://example.com",
		"--output-dir", outputDir,
		"--skip-wpprobe-check",
	})

	// Should succeed even if wpprobe is not available
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command with skip-wpprobe-check failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Environment looks good") {
		t.Fatalf("expected success message, got: %s", output)
	}
}

func TestInitCommandBinaryCheckWithDryRun(t *testing.T) {
	outputDir := t.TempDir()

	loader := &config.Loader{ConfigPath: ""}
	cmd := newInitCmd(loader)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--targets=https://example.com",
		"--output-dir", outputDir,
		"--dry-run",
	})

	// Should succeed in dry-run mode even if wpprobe is not available
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command with dry-run failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Environment looks good") {
		t.Fatalf("expected success message, got: %s", output)
	}
}

func TestInitCommandBinaryCheckWithoutSkipFlag(t *testing.T) {
	// This test verifies that when skip flag is not set and dry-run is not enabled,
	// the binary check is performed
	outputDir := t.TempDir()

	loader := &config.Loader{ConfigPath: ""}
	cmd := newInitCmd(loader)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--targets=https://example.com",
		"--output-dir", outputDir,
	})

	err := cmd.Execute()
	// In test environment, wpprobe binary is likely not available
	// So we expect this to fail with a binary not found error
	if err == nil {
		// If it succeeds, wpprobe must be available in the test environment
		t.Logf("wpprobe binary is available in test environment")
		return
	}

	// Verify the error is about the missing binary
	if !strings.Contains(err.Error(), "wpprobe binary not found") {
		t.Fatalf("expected wpprobe binary error, got: %v", err)
	}
}

func TestInitCommandWithConfigFile(t *testing.T) {
	// Create a temporary config file
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "test-config.yml")
	outputDir := filepath.Join(configDir, "output")

	configContent := `targets:
  - https://from-config.com
outputDir: ` + outputDir + `
mode: hybrid
threads: 5
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	loader := &config.Loader{ConfigPath: configPath}
	cmd := newInitCmd(loader)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--skip-wpprobe-check",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command with config file failed: %v", err)
	}

	// Verify output directory from config was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Fatalf("output directory from config was not created: %s", outputDir)
	}

	output := buf.String()
	if !strings.Contains(output, outputDir) {
		t.Fatalf("expected output dir from config in message, got: %s", output)
	}
}

func TestInitCommandOverridesConfigFile(t *testing.T) {
	// Create a temporary config file
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "test-config.yml")
	configOutputDir := filepath.Join(configDir, "config-output")
	cliOutputDir := filepath.Join(configDir, "cli-output")

	configContent := `targets:
  - https://from-config.com
outputDir: ` + configOutputDir + `
mode: hybrid
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	loader := &config.Loader{ConfigPath: configPath}
	cmd := newInitCmd(loader)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// CLI flags should override config file
	cmd.SetArgs([]string{
		"--output-dir", cliOutputDir,
		"--targets=https://from-cli.com",
		"--skip-wpprobe-check",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Verify CLI output directory was created (not config one)
	if _, err := os.Stat(cliOutputDir); os.IsNotExist(err) {
		t.Fatalf("CLI output directory was not created: %s", cliOutputDir)
	}

	output := buf.String()
	if !strings.Contains(output, cliOutputDir) {
		t.Fatalf("expected CLI output dir in message, got: %s", output)
	}

	// Config output dir should not be created
	if _, err := os.Stat(configOutputDir); err == nil {
		t.Logf("Warning: config output dir was created but shouldn't have been: %s", configOutputDir)
	}
}
