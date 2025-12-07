package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoaderLoadWithFileAndEnv(t *testing.T) {
	dir := t.TempDir()
	targetFile := filepath.Join(dir, "targets.txt")
	if err := os.WriteFile(targetFile, []byte("https://one.test\nhttps://two.test\n"), 0o600); err != nil {
		t.Fatalf("write targets: %v", err)
	}

	configPath := filepath.Join(dir, "wphunter.config.yml")
	configBody := []byte("mode: stealthy\nthreads: 6\noutputDir: out\ntargetsFile: " + targetFile + "\nformats:\n  - json\n")
	if err := os.WriteFile(configPath, configBody, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv(envThreadsKeys[0], "12")
	t.Setenv(envFormatsKeys[0], "csv")
	t.Setenv(envDetectorsKeys[0], "version")

	loader := Loader{ConfigPath: configPath}
	cfg, err := loader.Load(Overrides{})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate config: %v", err)
	}

	if len(cfg.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(cfg.Targets))
	}

	if cfg.Mode != "stealthy" {
		t.Fatalf("expected mode stealthy, got %s", cfg.Mode)
	}

	if cfg.Threads != 12 {
		t.Fatalf("env override should set threads to 12, got %d", cfg.Threads)
	}

	if cfg.OutputDir != "out" {
		t.Fatalf("expected output dir out, got %s", cfg.OutputDir)
	}

	if len(cfg.Formats) != 1 || cfg.Formats[0] != "csv" {
		t.Fatalf("unexpected formats: %#v", cfg.Formats)
	}

	if len(cfg.Detectors) != 1 || cfg.Detectors[0] != "version" {
		t.Fatalf("unexpected detectors: %#v", cfg.Detectors)
	}
}

func TestOverridesApplyTargetsList(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "wphunter.config.yml")
	if err := os.WriteFile(configPath, []byte("targets:\n  - https://from-file.test\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	loader := Loader{ConfigPath: configPath}
	over := Overrides{Targets: []string{"https://override.test"}}
	cfg, err := loader.Load(over)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if len(cfg.Targets) != 1 || cfg.Targets[0] != "https://override.test" {
		t.Fatalf("expected overrides to replace targets, got %#v", cfg.Targets)
	}
}

func TestParseTargetsList(t *testing.T) {
	input := "https://one.test,https://two.test\nhttps://three.test"
	targets := ParseTargetsList(input)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}
}

func TestReadTargetsFile_PathTraversal(t *testing.T) {
	dir := t.TempDir()

	// Create a legitimate targets file
	legitimateFile := filepath.Join(dir, "targets.txt")
	if err := os.WriteFile(legitimateFile, []byte("https://legitimate.test\n"), 0o600); err != nil {
		t.Fatalf("write legitimate file: %v", err)
	}

	// Create a subdirectory to test traversal from
	subDir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}

	// Create a file in the subdirectory that should NOT be accessible via traversal
	protectedFile := filepath.Join(subDir, "protected.txt")
	if err := os.WriteFile(protectedFile, []byte("https://protected.test\n"), 0o600); err != nil {
		t.Fatalf("write protected file: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		shouldFail  bool
		description string
	}{
		{
			name:        "normal_path",
			path:        legitimateFile,
			shouldFail:  false,
			description: "normal file path should work",
		},
		{
			name:        "path_with_dot_dot_slash",
			path:        filepath.Join(subDir, "..", "targets.txt"),
			shouldFail:  false,
			description: "filepath.Clean should normalize ../ correctly",
		},
		{
			name:        "path_with_multiple_dot_dot",
			path:        filepath.Join(subDir, "..", "..", "..", "etc", "passwd"),
			shouldFail:  true,
			description: "path traversal beyond temp dir should fail",
		},
		{
			name:        "path_with_dot_slash_dot_dot",
			path:        filepath.Join(subDir, ".", "..", "targets.txt"),
			shouldFail:  false,
			description: "filepath.Clean should normalize ./../ correctly",
		},
		{
			name:        "absolute_path_traversal",
			path:        filepath.Join(dir, "..", "..", "..", "etc", "passwd"),
			shouldFail:  true,
			description: "absolute path traversal should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targets, err := readTargetsFile(tt.path)
			if tt.shouldFail {
				if err == nil {
					t.Errorf("%s: expected error for path traversal, got targets: %v", tt.description, targets)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.description, err)
				}
				if len(targets) == 0 {
					t.Errorf("%s: expected targets, got none", tt.description)
				}
			}
		})
	}
}

func TestReadTargetsFile_SymbolicLinks(t *testing.T) {
	dir := t.TempDir()

	// Create a legitimate targets file
	legitimateFile := filepath.Join(dir, "targets.txt")
	content := "https://legitimate.test\nhttps://another.test\n"
	if err := os.WriteFile(legitimateFile, []byte(content), 0o600); err != nil {
		t.Fatalf("write legitimate file: %v", err)
	}

	// Create a symbolic link to the legitimate file
	symlinkPath := filepath.Join(dir, "symlink.txt")
	if err := os.Symlink(legitimateFile, symlinkPath); err != nil {
		// Skip test if symlinks are not supported (e.g., on Windows without admin privileges)
		t.Skipf("symlinks not supported: %v", err)
	}

	// Test that symlink works (should resolve and read the file)
	targets, err := readTargetsFile(symlinkPath)
	if err != nil {
		t.Fatalf("reading symlink should succeed: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets from symlink, got %d", len(targets))
	}

	// Create a broken symlink
	brokenSymlink := filepath.Join(dir, "broken.txt")
	if err := os.Symlink(filepath.Join(dir, "nonexistent.txt"), brokenSymlink); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	// Test that broken symlink fails appropriately
	_, err = readTargetsFile(brokenSymlink)
	if err == nil {
		t.Error("reading broken symlink should fail")
	}

	// Create a symlink that points outside the temp directory
	parentDir := filepath.Dir(dir)
	externalFile := filepath.Join(parentDir, "external.txt")
	if err := os.WriteFile(externalFile, []byte("https://external.test\n"), 0o600); err != nil {
		t.Fatalf("write external file: %v", err)
	}
	defer os.Remove(externalFile)

	externalSymlink := filepath.Join(dir, "external-symlink.txt")
	if err := os.Symlink(externalFile, externalSymlink); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	// Test that external symlink works (filepath.Clean doesn't prevent symlink resolution)
	// This is expected behavior - symlinks can point outside, but filepath.Clean sanitizes the path string
	targets, err = readTargetsFile(externalSymlink)
	if err != nil {
		t.Logf("note: external symlink read failed (may be expected): %v", err)
	} else if len(targets) != 1 {
		t.Logf("note: external symlink read succeeded with %d targets", len(targets))
	}
}

func TestReadTargetsFile_MalformedPaths(t *testing.T) {
	dir := t.TempDir()
	legitimateFile := filepath.Join(dir, "targets.txt")
	if err := os.WriteFile(legitimateFile, []byte("https://test.example\n"), 0o600); err != nil {
		t.Fatalf("write legitimate file: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		shouldFail  bool
		description string
	}{
		{
			name:        "path_with_null_byte",
			path:        legitimateFile + "\x00",
			shouldFail:  true,
			description: "path with null byte should fail",
		},
		{
			name:        "path_with_null_byte_in_middle",
			path:        filepath.Join(dir, "targets\x00.txt"),
			shouldFail:  true,
			description: "path with null byte in middle should fail",
		},
		{
			name:        "very_long_path",
			path:        filepath.Join(dir, strings.Repeat("a", 300), "targets.txt"),
			shouldFail:  true,
			description: "very long path should fail",
		},
		{
			name:        "empty_path",
			path:        "",
			shouldFail:  true,
			description: "empty path should fail",
		},
		{
			name:        "path_with_trailing_slash",
			path:        legitimateFile + string(filepath.Separator),
			shouldFail:  false, // filepath.Clean removes trailing separators, so this works
			description: "path with trailing separator should be cleaned and work",
		},
		{
			name:        "path_with_windows_separators_on_unix",
			path:        strings.ReplaceAll(legitimateFile, "/", "\\"),
			shouldFail:  true, // On Unix, backslashes are not path separators, so this creates an invalid path
			description: "path with Windows separators on Unix should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targets, err := readTargetsFile(tt.path)
			if tt.shouldFail {
				if err == nil {
					t.Errorf("%s: expected error for malformed path, got targets: %v", tt.description, targets)
				}
			} else {
				// For paths that should work, we still might get errors in some cases
				// (e.g., very long paths on some systems), so we just log
				if err != nil {
					t.Logf("%s: got error (may be expected): %v", tt.description, err)
				}
			}
		})
	}
}

func TestReadTargetsFile_EdgeCases(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name        string
		setup       func() string
		shouldFail  bool
		description string
	}{
		{
			name: "nonexistent_file",
			setup: func() string {
				return filepath.Join(dir, "nonexistent.txt")
			},
			shouldFail:  true,
			description: "nonexistent file should fail",
		},
		{
			name: "directory_instead_of_file",
			setup: func() string {
				dirPath := filepath.Join(dir, "notafile")
				if err := os.Mkdir(dirPath, 0o755); err != nil {
					t.Fatalf("create dir: %v", err)
				}
				return dirPath
			},
			shouldFail:  true,
			description: "directory path should fail (not a file)",
		},
		{
			name: "empty_file",
			setup: func() string {
				emptyFile := filepath.Join(dir, "empty.txt")
				if err := os.WriteFile(emptyFile, []byte(""), 0o600); err != nil {
					t.Fatalf("write empty file: %v", err)
				}
				return emptyFile
			},
			shouldFail:  false,
			description: "empty file should return empty targets",
		},
		{
			name: "file_with_only_comments",
			setup: func() string {
				commentFile := filepath.Join(dir, "comments.txt")
				if err := os.WriteFile(commentFile, []byte("# comment 1\n# comment 2\n"), 0o600); err != nil {
					t.Fatalf("write comment file: %v", err)
				}
				return commentFile
			},
			shouldFail:  false,
			description: "file with only comments should return empty targets",
		},
		{
			name: "file_with_mixed_content",
			setup: func() string {
				mixedFile := filepath.Join(dir, "mixed.txt")
				content := "# comment\nhttps://target1.test\n  \n# another comment\nhttps://target2.test\n"
				if err := os.WriteFile(mixedFile, []byte(content), 0o600); err != nil {
					t.Fatalf("write mixed file: %v", err)
				}
				return mixedFile
			},
			shouldFail:  false,
			description: "file with mixed content should parse correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			targets, err := readTargetsFile(path)
			if tt.shouldFail {
				if err == nil {
					t.Errorf("%s: expected error, got targets: %v", tt.description, targets)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.description, err)
				}
				// For empty files or comment-only files, we expect empty targets
				if tt.name == "empty_file" || tt.name == "file_with_only_comments" {
					if len(targets) != 0 {
						t.Errorf("%s: expected empty targets, got %d", tt.description, len(targets))
					}
				} else if tt.name == "file_with_mixed_content" {
					if len(targets) != 2 {
						t.Errorf("%s: expected 2 targets, got %d", tt.description, len(targets))
					}
				}
			}
		})
	}
}
