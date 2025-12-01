package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/example/wphunter/internal/config"
	"github.com/example/wphunter/internal/detector"
)

func TestScanCommandDryRunCreatesArtifacts(t *testing.T) {
	outputDir := t.TempDir()
	summaryPath := filepath.Join(outputDir, "summary.json")

	loader := &config.Loader{ConfigPath: ""}
	cmd := newScanCmd(loader)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--targets=https://one.test,https://two.test",
		"--dry-run",
		"--detectors", "",
		"--output-dir", outputDir,
		"--formats", "json",
		"--summary-file", summaryPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("scan command failed: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(outputDir, "scan_*.json"))
	if err != nil {
		t.Fatalf("glob artifacts: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected one artifact, found %d (%v)", len(files), files)
	}

	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}

	if !bytes.Contains(data, []byte("dry-run placeholder")) {
		t.Fatalf("artifact should mention dry-run placeholder, got %s", string(data))
	}

	if _, err := os.Stat(summaryPath); err != nil {
		t.Fatalf("summary not created: %v", err)
	}
}

func TestWritePlaceholderArtifactCSV(t *testing.T) {
	outputDir := t.TempDir()
	path := filepath.Join(outputDir, "scan.csv")
	targets := []string{"https://one.test", "https://two.test"}

	if err := writePlaceholderArtifact(path, "csv", targets); err != nil {
		t.Fatalf("write placeholder csv: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}

	content := string(data)
	for _, target := range targets {
		if !bytes.Contains(data, []byte(target)) {
			t.Fatalf("csv missing target %s: %s", target, content)
		}
	}
}

func TestWriteSummary(t *testing.T) {
	targets := []string{"https://one.test"}
	cfg := config.RuntimeConfig{
		Targets: targets,
		Mode:    "hybrid",
		DryRun:  true,
	}

	outputDir := t.TempDir()
	summaryPath := filepath.Join(outputDir, "summary.json")

	artifacts := []string{"scan.json"}
	var detections []detector.Result
	if err := writeSummary(summaryPath, cfg, artifacts, detections); err != nil {
		t.Fatalf("write summary: %v", err)
	}

	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse summary json: %v", err)
	}

	if parsed["dryRun"] != true {
		t.Fatalf("summary missing dryRun flag: %+v", parsed)
	}
}
