package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureOutputDir(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*testing.T) string
		path      string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "empty path returns error",
			path:      "",
			wantError: true,
			errorMsg:  "output directory cannot be empty",
		},
		{
			name: "successful directory creation",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "output")
			},
			wantError: false,
		},
		{
			name: "nested directory creation",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "level1", "level2", "level3")
			},
			wantError: false,
		},
		{
			name: "idempotent behavior - directory already exists",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				existingDir := filepath.Join(tmpDir, "existing")
				if err := os.MkdirAll(existingDir, 0o755); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				return existingDir
			},
			wantError: false,
		},
		{
			name: "permission denied scenario",
			setup: func(t *testing.T) string {
				if os.Getuid() == 0 {
					t.Skip("skipping permission test when running as root")
				}
				tmpDir := t.TempDir()
				restrictedDir := filepath.Join(tmpDir, "restricted")
				if err := os.MkdirAll(restrictedDir, 0o755); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				// Make the directory read-only
				if err := os.Chmod(restrictedDir, 0o444); err != nil {
					t.Fatalf("setup failed to set permissions: %v", err)
				}
				// Try to create a subdirectory in the read-only directory
				return filepath.Join(restrictedDir, "subdir")
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.setup != nil {
				path = tt.setup(t)
			} else {
				path = tt.path
			}

			err := ensureOutputDir(path)

			if tt.wantError {
				if err == nil {
					t.Errorf("ensureOutputDir() expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("ensureOutputDir() error = %q, want %q", err.Error(), tt.errorMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("ensureOutputDir() unexpected error = %v", err)
				return
			}

			// Verify directory was created
			if path != "" {
				info, err := os.Stat(path)
				if err != nil {
					t.Errorf("directory not created: %v", err)
					return
				}
				if !info.IsDir() {
					t.Errorf("path exists but is not a directory")
				}
			}
		})
	}
}

func TestEnsureOutputDirIdempotence(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "idempotent")

	// First call - create directory
	if err := ensureOutputDir(outputPath); err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Verify directory was created
	info1, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info1.IsDir() {
		t.Fatalf("path exists but is not a directory")
	}

	// Second call - should succeed without error
	if err := ensureOutputDir(outputPath); err != nil {
		t.Errorf("second call failed: %v", err)
	}

	// Verify directory still exists
	info2, err := os.Stat(outputPath)
	if err != nil {
		t.Errorf("directory no longer exists: %v", err)
	}
	if !info2.IsDir() {
		t.Errorf("path is no longer a directory")
	}
}
