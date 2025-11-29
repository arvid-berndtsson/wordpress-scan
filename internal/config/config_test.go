package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoaderLoadWithFileAndEnv(t *testing.T) {
	dir := t.TempDir()
	targetFile := filepath.Join(dir, "targets.txt")
	if err := os.WriteFile(targetFile, []byte("https://one.test\nhttps://two.test\n"), 0o600); err != nil {
		t.Fatalf("write targets: %v", err)
	}

	configPath := filepath.Join(dir, "worker.config.yml")
	configBody := []byte("mode: stealthy\nthreads: 6\noutputDir: out\ntargetsFile: " + targetFile + "\nformats:\n  - json\n")
	if err := os.WriteFile(configPath, configBody, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv(envThreads, "12")
	t.Setenv(envFormats, "csv")

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
}

func TestOverridesApplyTargetsList(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "worker.config.yml")
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
