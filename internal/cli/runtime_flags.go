package cli

import (
	"fmt"

	"github.com/example/wphunter/internal/config"
	"github.com/spf13/cobra"
)

// runtimeFlagSet tracks shared scan/init flags before they are converted into config overrides.
type runtimeFlagSet struct {
	targets     string
	targetsFile string
	mode        string
	threads     int
	outputDir   string
	formats     string
	detectors   string
	dryRun      bool
	summaryFile string
}

func bindRuntimeFlags(cmd *cobra.Command, flags *runtimeFlagSet) {
	cmd.Flags().StringVar(&flags.targets, "targets", "", "Comma-separated list of targets (overrides config)")
	cmd.Flags().StringVar(&flags.targetsFile, "targets-file", "", "Path to a file with one target per line")
	cmd.Flags().StringVar(&flags.mode, "mode", "", "Scan mode: stealthy, bruteforce, or hybrid")
	cmd.Flags().IntVar(&flags.threads, "threads", 0, fmt.Sprintf("Number of concurrent threads (1-%d)", config.MaxThreads))
	cmd.Flags().StringVar(&flags.outputDir, "output-dir", "", "Directory for scan artifacts")
	cmd.Flags().StringVar(&flags.formats, "formats", "", "Comma-separated output formats (json,csv)")
	cmd.Flags().StringVar(&flags.detectors, "detectors", "", "Comma-separated detectors to run (version,plugins,...)")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Skip wpprobe execution and emit placeholder artifacts")
	cmd.Flags().StringVar(&flags.summaryFile, "summary-file", "", "Optional summary JSON output path")
}

func (f runtimeFlagSet) toOverrides(cmd *cobra.Command) config.Overrides {
	ov := config.Overrides{}
	if cmd.Flags().Changed("targets") {
		ov.Targets = config.ParseTargetsList(f.targets)
	}

	if cmd.Flags().Changed("targets-file") {
		ov.TargetsFile = f.targetsFile
	}

	if cmd.Flags().Changed("mode") {
		ov.Mode = f.mode
	}

	if cmd.Flags().Changed("threads") {
		ov.Threads = f.threads
		ov.ThreadsSet = true
	}

	if cmd.Flags().Changed("output-dir") {
		ov.OutputDir = f.outputDir
	}

	if cmd.Flags().Changed("formats") {
		ov.Formats = config.ParseFormats(f.formats)
	}

	if cmd.Flags().Changed("detectors") {
		ov.Detectors = config.ParseDetectors(f.detectors)
	}

	if cmd.Flags().Changed("dry-run") {
		ov.DryRun = &f.dryRun
	}

	if cmd.Flags().Changed("summary-file") {
		ov.SummaryFile = f.summaryFile
	}

	return ov
}
