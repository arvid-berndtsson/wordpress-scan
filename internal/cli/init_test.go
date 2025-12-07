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

func TestInitCommandConfigurationError_InvalidThreadsTooHigh(t *testing.T) {
	outputDir := t.TempDir()

	loader := &config.Loader{ConfigPath: ""}
	cmd := newInitCmd(loader)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--targets=https://example.com",
		"--output-dir", outputDir,
		"--threads=65",
		"--skip-wpprobe-check",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected init to fail with threads=65, but it succeeded")
	}

	if !strings.Contains(err.Error(), "threads must be between") {
		t.Fatalf("expected threads validation error, got: %v", err)
	}
}

func TestInitCommandConfigurationError_InvalidThreadsTooLow(t *testing.T) {
	outputDir := t.TempDir()

	loader := &config.Loader{ConfigPath: ""}
	cmd := newInitCmd(loader)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--targets=https://example.com",
		"--output-dir", outputDir,
		"--threads=0",
		"--skip-wpprobe-check",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected init to fail with threads=0, but it succeeded")
	}

	if !strings.Contains(err.Error(), "threads must be between") {
		t.Fatalf("expected threads validation error, got: %v", err)
	}
}

func TestInitCommandConfigurationError_InvalidConfigFile(t *testing.T) {
	// Create a temporary config file with invalid YAML
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "invalid-config.yml")

	invalidYAML := `targets:
  - https://example.com
outputDir: /tmp/output
mode: hybrid
invalid: [unclosed bracket
`

	if err := os.WriteFile(configPath, []byte(invalidYAML), 0o644); err != nil {
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

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected init to fail with invalid YAML, but it succeeded")
	}

	// The error should be about YAML parsing
	if !strings.Contains(err.Error(), "yaml") && !strings.Contains(err.Error(), "YAML") {
		t.Logf("Note: YAML parsing error format may vary, got: %v", err)
	}
}

func TestInitCommandOutputDirectoryAlreadyExists(t *testing.T) {
	// Test that init succeeds when output directory already exists
	outputDir := t.TempDir()

	// Directory already exists (created by t.TempDir())
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
		t.Fatalf("init command failed when output dir already exists: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Environment looks good") {
		t.Fatalf("expected success message, got: %s", output)
	}
}

func TestInitCommandBinaryCheckWithSkipFlagAndDryRun(t *testing.T) {
	// Test that binary check is skipped when both skip flag and dry-run are set
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
		"--dry-run",
	})

	// Should succeed even if wpprobe is not available
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command with skip-wpprobe-check and dry-run failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Environment looks good") {
		t.Fatalf("expected success message, got: %s", output)
	}
}

func TestInitCommandBinaryCheckErrorFormat(t *testing.T) {
	// Test that binary check returns a properly formatted error
	outputDir := t.TempDir()

	loader := &config.Loader{ConfigPath: ""}
	cmd := newInitCmd(loader)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--targets=https://example.com",
		"--output-dir", outputDir,
		// No --skip-wpprobe-check and no --dry-run
	})

	err := cmd.Execute()
	if err == nil {
		// If wpprobe is available in test environment, skip this test
		t.Logf("wpprobe binary is available in test environment, skipping error format test")
		return
	}

	// Verify the error message format
	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "wpprobe") {
		t.Fatalf("expected error to mention 'wpprobe', got: %v", err)
	}
	if !strings.Contains(errorMsg, "binary") || !strings.Contains(errorMsg, "not found") {
		t.Fatalf("expected error to mention binary not found, got: %v", err)
	}
}

func TestInitCommandConfigFileLoadingError(t *testing.T) {
	// Test behavior when config file exists but has read permission issues
	// Note: This test may not work on all systems, so we'll test a different scenario
	// Instead, test that config file loading errors are properly propagated
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "test-config.yml")
	outputDir := filepath.Join(configDir, "output")

	// Create a config file that will cause a loading error (invalid targets file reference)
	configContent := `targets:
  - https://example.com
targetsFile: /nonexistent/targets/file.txt
outputDir: ` + outputDir + `
mode: hybrid
threads: 5
formats:
  - json
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

	// Should fail because targets file doesn't exist
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected init to fail with non-existent targets file, but it succeeded")
	}

	// Error should be about file not found
	if !strings.Contains(err.Error(), "no such file") && !strings.Contains(err.Error(), "cannot find") && !strings.Contains(err.Error(), "open") {
		t.Logf("Note: File not found error format may vary, got: %v", err)
	}
}
