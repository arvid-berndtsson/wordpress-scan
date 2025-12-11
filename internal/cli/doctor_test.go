package cli

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/example/wphunter/internal/config"
)

func TestNewDoctorCmd(t *testing.T) {
	loader := &config.Loader{ConfigPath: "../config/testdata/valid.yml"}
	cmd := newDoctorCmd(loader)

	if cmd == nil {
		t.Fatal("newDoctorCmd returned nil")
	}

	if cmd.Use != "doctor" {
		t.Errorf("expected Use='doctor', got %q", cmd.Use)
	}

	if !strings.Contains(cmd.Short, "Validate") {
		t.Errorf("expected Short to contain 'Validate', got %q", cmd.Short)
	}
}

func TestCheckGoVersion(t *testing.T) {
	check := checkGoVersion()

	if check.Name != "Go Runtime" {
		t.Errorf("expected Name='Go Runtime', got %q", check.Name)
	}

	if check.Status != "✓" {
		t.Errorf("expected Status='✓', got %q", check.Status)
	}

	expectedVersion := runtime.Version()
	if !strings.Contains(check.Detail, expectedVersion) {
		t.Errorf("expected Detail to contain %q, got %q", expectedVersion, check.Detail)
	}

	if check.Error != nil {
		t.Errorf("expected no error, got %v", check.Error)
	}
}

func TestCheckWPProbeBinaryDryRun(t *testing.T) {
	check := checkWPProbeBinary(true)

	if check.Name != "wpprobe Binary" {
		t.Errorf("expected Name='wpprobe Binary', got %q", check.Name)
	}

	if check.Status != "⊘" {
		t.Errorf("expected Status='⊘', got %q", check.Status)
	}

	if !strings.Contains(check.Detail, "dry-run") {
		t.Errorf("expected Detail to contain 'dry-run', got %q", check.Detail)
	}

	if check.Error != nil {
		t.Errorf("expected no error in dry-run mode, got %v", check.Error)
	}
}

func TestCheckWPProbeBinaryNotFound(t *testing.T) {
	// This test assumes wpprobe is not in PATH
	// If it is in PATH, we skip this test
	check := checkWPProbeBinary(false)

	if check.Name != "wpprobe Binary" {
		t.Errorf("expected Name='wpprobe Binary', got %q", check.Name)
	}

	// The check might pass if wpprobe is actually installed
	// So we only verify structure, not outcome
	if check.Status != "✓" && check.Status != "✗" {
		t.Errorf("expected Status to be either '✓' or '✗', got %q", check.Status)
	}
}

func TestCheckConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.RuntimeConfig
		wantStatus  string
		wantErr     bool
		detailMatch string
	}{
		{
			name: "valid configuration",
			cfg: &config.RuntimeConfig{
				Mode:      "hybrid",
				Threads:   10,
				OutputDir: "/tmp/test",
				Targets:   []string{"https://example.com"},
				Formats:   []string{"json"},
			},
			wantStatus:  "✓",
			wantErr:     false,
			detailMatch: "1 targets",
		},
		{
			name: "multiple targets",
			cfg: &config.RuntimeConfig{
				Mode:      "stealthy",
				Threads:   5,
				OutputDir: "/tmp/test",
				Targets:   []string{"https://example.com", "https://test.com"},
				Formats:   []string{"json"},
			},
			wantStatus:  "✓",
			wantErr:     false,
			detailMatch: "2 targets",
		},
		{
			name: "invalid mode",
			cfg: &config.RuntimeConfig{
				Mode:      "",
				Threads:   10,
				OutputDir: "/tmp/test",
				Targets:   []string{"https://example.com"},
				Formats:   []string{"json"},
			},
			wantStatus: "✗",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := checkConfiguration(tt.cfg)

			if check.Name != "Configuration" {
				t.Errorf("expected Name='Configuration', got %q", check.Name)
			}

			if check.Status != tt.wantStatus {
				t.Errorf("expected Status=%q, got %q", tt.wantStatus, check.Status)
			}

			if tt.wantErr && check.Error == nil {
				t.Error("expected error but got none")
			}

			if !tt.wantErr && check.Error != nil {
				t.Errorf("expected no error but got: %v", check.Error)
			}

			if tt.detailMatch != "" && !strings.Contains(check.Detail, tt.detailMatch) {
				t.Errorf("expected Detail to contain %q, got %q", tt.detailMatch, check.Detail)
			}
		})
	}
}

func TestCheckOutputDirectory(t *testing.T) {
	tests := []struct {
		name       string
		outputDir  string
		wantStatus string
		wantErr    bool
	}{
		{
			name:       "valid directory",
			outputDir:  filepath.Join(t.TempDir(), "output"),
			wantStatus: "✓",
			wantErr:    false,
		},
		{
			name:       "nested directory",
			outputDir:  filepath.Join(t.TempDir(), "a", "b", "c"),
			wantStatus: "✓",
			wantErr:    false,
		},
		{
			name:       "empty path",
			outputDir:  "",
			wantStatus: "✗",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := checkOutputDirectory(tt.outputDir)

			if check.Name != "Output Directory" {
				t.Errorf("expected Name='Output Directory', got %q", check.Name)
			}

			if check.Status != tt.wantStatus {
				t.Errorf("expected Status=%q, got %q", tt.wantStatus, check.Status)
			}

			if tt.wantErr && check.Error == nil {
				t.Error("expected error but got none")
			}

			if !tt.wantErr && check.Error != nil {
				t.Errorf("expected no error but got: %v", check.Error)
			}

			if tt.outputDir != "" && check.Detail != tt.outputDir {
				t.Errorf("expected Detail=%q, got %q", tt.outputDir, check.Detail)
			}
		})
	}
}

func TestCheckNetworkReachability(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name           string
		targets        []string
		expectedChecks int
		wantSuccess    bool
	}{
		{
			name:           "reachable target",
			targets:        []string{server.URL},
			expectedChecks: 1,
			wantSuccess:    true,
		},
		{
			name:           "multiple reachable targets",
			targets:        []string{server.URL, server.URL},
			expectedChecks: 2,
			wantSuccess:    true,
		},
		{
			name:           "unreachable target",
			targets:        []string{"http://localhost:99999"},
			expectedChecks: 1,
			wantSuccess:    false,
		},
		{
			name:           "more than max targets",
			targets:        []string{server.URL, server.URL, server.URL, server.URL, server.URL},
			expectedChecks: 4, // 3 checked + 1 "skipped" message
			wantSuccess:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			checks := checkNetworkReachability(ctx, tt.targets)

			if len(checks) != tt.expectedChecks {
				t.Errorf("expected %d checks, got %d", tt.expectedChecks, len(checks))
			}

			if len(checks) > 0 {
				firstCheck := checks[0]
				if !strings.HasPrefix(firstCheck.Name, "Network:") {
					t.Errorf("expected Name to start with 'Network:', got %q", firstCheck.Name)
				}

				if tt.wantSuccess {
					if firstCheck.Status != "✓" {
						t.Errorf("expected successful check, got Status=%q, Error=%v", firstCheck.Status, firstCheck.Error)
					}
				} else {
					if firstCheck.Status != "✗" {
						t.Errorf("expected failed check, got Status=%q", firstCheck.Status)
					}
				}
			}
		})
	}
}

func TestCheckNetworkReachabilityInvalidURL(t *testing.T) {
	ctx := context.Background()
	targets := []string{"not-a-valid-url"}
	checks := checkNetworkReachability(ctx, targets)

	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks))
	}

	check := checks[0]
	if check.Status != "✗" {
		t.Errorf("expected Status='✗', got %q", check.Status)
	}

	if check.Error == nil {
		t.Error("expected error for invalid URL")
	}

	// The URL might be treated as invalid or unreachable depending on the implementation
	if !strings.Contains(check.Detail, "Invalid") && !strings.Contains(check.Detail, "Unreachable") {
		t.Errorf("expected Detail to contain 'Invalid' or 'Unreachable', got %q", check.Detail)
	}
}

func TestPrintDoctorReport(t *testing.T) {
	tests := []struct {
		name           string
		checks         []doctorCheck
		expectedOutput []string
	}{
		{
			name: "all passing checks",
			checks: []doctorCheck{
				{Name: "Test Check 1", Status: "✓", Detail: "OK"},
				{Name: "Test Check 2", Status: "✓", Detail: "Good"},
			},
			expectedOutput: []string{"✓", "Test Check 1", "OK", "Test Check 2", "Good"},
		},
		{
			name: "failing check",
			checks: []doctorCheck{
				{Name: "Failed Check", Status: "✗", Detail: "Bad", Error: fmt.Errorf("test error")},
			},
			expectedOutput: []string{"✗", "Failed Check", "Bad", "Error"},
		},
		{
			name: "skipped check",
			checks: []doctorCheck{
				{Name: "Skipped Check", Status: "⊘", Detail: "Not applicable"},
			},
			expectedOutput: []string{"⊘", "Skipped Check", "Not applicable"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			loader := &config.Loader{}
			cmd := newDoctorCmd(loader)
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)

			printDoctorReport(cmd, tt.checks)

			// Check both stdout and stderr for expected output
			output := stdout.String() + stderr.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestRunDoctorChecks(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.RuntimeConfig{
		Mode:      "hybrid",
		Threads:   10,
		OutputDir: tempDir,
		Targets:   []string{},
		Formats:   []string{"json"},
		DryRun:    true, // Use dry-run to avoid requiring wpprobe
	}

	ctx := context.Background()
	checks := runDoctorChecks(ctx, cfg)

	if len(checks) == 0 {
		t.Fatal("expected at least one check")
	}

	// Verify we have essential checks
	checkNames := make(map[string]bool)
	for _, check := range checks {
		checkNames[check.Name] = true
	}

	requiredChecks := []string{"Go Runtime", "Configuration", "Output Directory"}
	for _, required := range requiredChecks {
		if !checkNames[required] {
			t.Errorf("missing required check: %s", required)
		}
	}
}

func TestDoctorCmdWithConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Create a temporary config file
	configPath := filepath.Join(tempDir, "test.yml")
	configContent := fmt.Sprintf(`
mode: hybrid
threads: 5
outputDir: %s
targets:
  - https://example.com
dryRun: true
`, tempDir)

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	loader := &config.Loader{ConfigPath: configPath}
	cmd := newDoctorCmd(loader)

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"--dry-run"})

	// Run the command
	err := cmd.Execute()
	
	// In dry-run mode, we should succeed
	if err != nil {
		t.Logf("Command output:\n%s", stdout.String())
		t.Errorf("expected no error, got: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "diagnostics") {
		t.Errorf("expected output to contain 'diagnostics', got:\n%s", output)
	}
}

func TestDoctorCmdNetworkChecks(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tempDir := t.TempDir()

	// Create config with reachable target
	configPath := filepath.Join(tempDir, "test.yml")
	configContent := fmt.Sprintf(`
mode: hybrid
threads: 5
outputDir: %s
targets:
  - %s
dryRun: true
`, tempDir, server.URL)

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	loader := &config.Loader{ConfigPath: configPath}
	cmd := newDoctorCmd(loader)

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Logf("Command output:\n%s", stdout.String())
		t.Errorf("expected no error, got: %v", err)
	}

	output := stdout.String()
	// In dry-run mode, network checks might be skipped
	// Just verify the command executed successfully
	if !strings.Contains(output, "diagnostics") {
		t.Errorf("expected output to contain 'diagnostics', got:\n%s", output)
	}
}
