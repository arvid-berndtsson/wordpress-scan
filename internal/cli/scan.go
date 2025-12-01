package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/example/wphunter/internal/config"
	"github.com/example/wphunter/internal/detector"
	"github.com/example/wphunter/internal/events"
	"github.com/example/wphunter/internal/wpprobe"
	"github.com/spf13/cobra"
)

func newScanCmd(loader *config.Loader) *cobra.Command {
	flags := &runtimeFlagSet{}

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Run wpprobe plus configured detectors against WordPress targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			overrides := flags.toOverrides(cmd)
			cfg, err := loader.Load(overrides)
			if err != nil {
				return err
			}

			if err := cfg.Validate(); err != nil {
				return err
			}

			if err := ensureOutputDir(cfg.OutputDir); err != nil {
				return err
			}

			targetsFile, err := writeTargetsTempFile(cfg.Targets)
			if err != nil {
				return err
			}
			defer os.Remove(targetsFile)

			emitter := events.NewEmitter(cmd.OutOrStdout())
			if err := emitter.Emit(events.Event{Type: "scan-start", Message: "Starting scan", Fields: map[string]interface{}{"targets": len(cfg.Targets), "mode": cfg.Mode, "dryRun": cfg.DryRun}}); err != nil {
				return err
			}

			runner := wpprobe.NewRunner()
			if !cfg.DryRun {
				if err := runner.EnsureBinary(); err != nil {
					return err
				}
			}

			timestamp := time.Now().UTC().Format("20060102_150405")
			var outputs []string
			var detectionResults []detector.Result

			for _, format := range cfg.Formats {
				format = strings.ToLower(strings.TrimSpace(format))
				if format == "" {
					continue
				}

				outputPath := filepath.Join(cfg.OutputDir, fmt.Sprintf("scan_%s.%s", timestamp, format))
				if cfg.DryRun {
					if err := writePlaceholderArtifact(outputPath, format, cfg.Targets); err != nil {
						return err
					}
				} else {
					if err := runner.Scan(cmd.Context(), wpprobe.ScanInput{
						TargetsFile: targetsFile,
						Mode:        cfg.Mode,
						Threads:     cfg.Threads,
						OutputPath:  outputPath,
						Stdout:      cmd.ErrOrStderr(),
						Stderr:      cmd.ErrOrStderr(),
					}); err != nil {
						return err
					}
				}

				outputs = append(outputs, outputPath)
				if err := emitter.Emit(events.Event{Type: "artifact-written", Fields: map[string]interface{}{"path": outputPath, "format": format}}); err != nil {
					return err
				}
			}

			if !cfg.DryRun {
				dets, err := detector.DefaultRegistry.BuildDetectors(cfg.Detectors)
				if err != nil {
					return err
				}

				if len(dets) > 0 {
					detectionResults, err = detector.Run(cmd.Context(), dets, cfg.Targets)
					if err != nil {
						return err
					}

					detectionsPath := filepath.Join(cfg.OutputDir, fmt.Sprintf("detections_%s.json", timestamp))
					if err := writeDetectionsArtifact(detectionsPath, detectionResults); err != nil {
						return err
					}

					outputs = append(outputs, detectionsPath)
					if err := emitter.Emit(events.Event{Type: "artifact-written", Fields: map[string]interface{}{"path": detectionsPath, "format": "detections"}}); err != nil {
						return err
					}

					for _, res := range detectionResults {
						if err := emitter.Emit(events.Event{
							Type:    "detection",
							Message: res.Summary,
							Fields: map[string]interface{}{
								"target":     res.Target,
								"detector":   res.Detector,
								"severity":   res.Severity,
								"confidence": res.Confidence,
							},
						}); err != nil {
							return err
						}
					}
				}
			} else if len(cfg.Detectors) > 0 {
				if err := emitter.Emit(events.Event{Type: "detectors-skipped", Message: "Detectors require live targets; skipped due to --dry-run"}); err != nil {
					return err
				}
			}

			if cfg.SummaryFile != "" {
				if err := writeSummary(cfg.SummaryFile, cfg, outputs, detectionResults); err != nil {
					return err
				}
			}

			return emitter.Emit(events.Event{Type: "scan-finished", Message: "Scan complete", Fields: map[string]interface{}{"artifacts": len(outputs)}})
		},
	}

	bindRuntimeFlags(cmd, flags)

	return cmd
}

func writeTargetsTempFile(targets []string) (string, error) {
	file, err := os.CreateTemp("", "wphunter-targets-*.txt")
	if err != nil {
		return "", err
	}

	for _, target := range targets {
		if _, err := fmt.Fprintln(file, target); err != nil {
			file.Close()
			return "", err
		}
	}

	if err := file.Close(); err != nil {
		return "", err
	}

	return file.Name(), nil
}

func writePlaceholderArtifact(path, format string, targets []string) error {
	if err := ensureOutputDir(filepath.Dir(path)); err != nil {
		return err
	}

	switch format {
	case "json":
		payload := map[string]interface{}{
			"generatedAt": time.Now().UTC().Format(time.RFC3339),
			"targets":     targets,
			"note":        "dry-run placeholder artifact",
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(path, append(data, '\n'), 0o644)
	case "csv":
		lines := []string{"target,status"}
		for _, target := range targets {
			lines = append(lines, fmt.Sprintf("%s,placeholder", target))
		}
		content := strings.Join(lines, "\n") + "\n"
		return os.WriteFile(path, []byte(content), 0o644)
	default:
		return fmt.Errorf("unsupported format %s", format)
	}
}

func writeSummary(path string, cfg config.RuntimeConfig, artifacts []string, detections []detector.Result) error {
	summary := map[string]interface{}{
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"targets":     cfg.Targets,
		"mode":        cfg.Mode,
		"artifacts":   artifacts,
		"dryRun":      cfg.DryRun,
		"detectors":   cfg.Detectors,
		"detections":  detections,
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}

	if err := ensureOutputDir(filepath.Dir(path)); err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func writeDetectionsArtifact(path string, results []detector.Result) error {
	if err := ensureOutputDir(filepath.Dir(path)); err != nil {
		return err
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0o644)
}
