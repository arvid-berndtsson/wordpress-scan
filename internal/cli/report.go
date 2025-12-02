package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/example/wphunter/internal/events"
	"github.com/spf13/cobra"
)

func newReportCmd() *cobra.Command {
	var inputPath string
	var summaryPath string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate aggregate stats from a scan artifact",
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputPath == "" {
				return errors.New("--input is required")
			}

			data, err := os.ReadFile(inputPath)
			if err != nil {
				return err
			}

			stats := map[string]interface{}{
				"input":       inputPath,
				"sizeBytes":   len(data),
				"generatedAt": time.Now().UTC().Format(time.RFC3339),
				"mentions":    bytes.Count(bytes.ToLower(data), []byte("vulnerability")),
			}

			emitter := events.NewEmitter(cmd.OutOrStdout())
			if err := emitter.Emit(events.Event{Type: "report", Message: "Report generated", Fields: stats}); err != nil {
				return err
			}

			if summaryPath != "" {
				if err := writeReportSummary(summaryPath, stats); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Summary written to %s\n", summaryPath)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&inputPath, "input", "", "Path to JSON scan artifact")
	cmd.Flags().StringVar(&summaryPath, "summary-file", "", "Optional path to store summary JSON")
	if err := cmd.MarkFlagRequired("input"); err != nil {
		panic(err)
	}

	return cmd
}

func writeReportSummary(path string, stats map[string]interface{}) error {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}
