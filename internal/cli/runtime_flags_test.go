package cli

import (
	"reflect"
	"testing"

	"github.com/example/wphunter/internal/config"
	"github.com/spf13/cobra"
)

func TestRuntimeFlagSetToOverrides(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*cobra.Command, *runtimeFlagSet)
		expected config.Overrides
	}{
		{
			name: "no flags changed returns empty overrides",
			setup: func(cmd *cobra.Command, flags *runtimeFlagSet) {
				// No flags set
			},
			expected: config.Overrides{},
		},
		{
			name: "targets flag changed",
			setup: func(cmd *cobra.Command, flags *runtimeFlagSet) {
				flags.targets = "https://example.com,https://test.com"
				cmd.Flags().Set("targets", flags.targets)
			},
			expected: config.Overrides{
				Targets: []string{"https://example.com", "https://test.com"},
			},
		},
		{
			name: "targets-file flag changed",
			setup: func(cmd *cobra.Command, flags *runtimeFlagSet) {
				flags.targetsFile = "/path/to/targets.txt"
				cmd.Flags().Set("targets-file", flags.targetsFile)
			},
			expected: config.Overrides{
				TargetsFile: "/path/to/targets.txt",
			},
		},
		{
			name: "mode flag changed",
			setup: func(cmd *cobra.Command, flags *runtimeFlagSet) {
				flags.mode = "stealthy"
				cmd.Flags().Set("mode", flags.mode)
			},
			expected: config.Overrides{
				Mode: "stealthy",
			},
		},
		{
			name: "threads flag changed",
			setup: func(cmd *cobra.Command, flags *runtimeFlagSet) {
				flags.threads = 20
				cmd.Flags().Set("threads", "20")
			},
			expected: config.Overrides{
				Threads:    20,
				ThreadsSet: true,
			},
		},
		{
			name: "output-dir flag changed",
			setup: func(cmd *cobra.Command, flags *runtimeFlagSet) {
				flags.outputDir = "/custom/output"
				cmd.Flags().Set("output-dir", flags.outputDir)
			},
			expected: config.Overrides{
				OutputDir: "/custom/output",
			},
		},
		{
			name: "formats flag changed",
			setup: func(cmd *cobra.Command, flags *runtimeFlagSet) {
				flags.formats = "json,csv"
				cmd.Flags().Set("formats", flags.formats)
			},
			expected: config.Overrides{
				Formats: []string{"json", "csv"},
			},
		},
		{
			name: "detectors flag changed",
			setup: func(cmd *cobra.Command, flags *runtimeFlagSet) {
				flags.detectors = "version,plugins,themes"
				cmd.Flags().Set("detectors", flags.detectors)
			},
			expected: config.Overrides{
				Detectors: []string{"version", "plugins", "themes"},
			},
		},
		{
			name: "dry-run flag changed to true",
			setup: func(cmd *cobra.Command, flags *runtimeFlagSet) {
				flags.dryRun = true
				cmd.Flags().Set("dry-run", "true")
			},
			expected: config.Overrides{
				DryRun: boolPtr(true),
			},
		},
		{
			name: "dry-run flag changed to false",
			setup: func(cmd *cobra.Command, flags *runtimeFlagSet) {
				flags.dryRun = false
				cmd.Flags().Set("dry-run", "false")
			},
			expected: config.Overrides{
				DryRun: boolPtr(false),
			},
		},
		{
			name: "summary-file flag changed",
			setup: func(cmd *cobra.Command, flags *runtimeFlagSet) {
				flags.summaryFile = "/path/to/summary.json"
				cmd.Flags().Set("summary-file", flags.summaryFile)
			},
			expected: config.Overrides{
				SummaryFile: "/path/to/summary.json",
			},
		},
		{
			name: "multiple flags changed",
			setup: func(cmd *cobra.Command, flags *runtimeFlagSet) {
				flags.targets = "https://multi.test"
				flags.mode = "bruteforce"
				flags.threads = 32
				flags.outputDir = "/multi/output"
				flags.dryRun = true
				cmd.Flags().Set("targets", flags.targets)
				cmd.Flags().Set("mode", flags.mode)
				cmd.Flags().Set("threads", "32")
				cmd.Flags().Set("output-dir", flags.outputDir)
				cmd.Flags().Set("dry-run", "true")
			},
			expected: config.Overrides{
				Targets:    []string{"https://multi.test"},
				Mode:       "bruteforce",
				Threads:    32,
				ThreadsSet: true,
				OutputDir:  "/multi/output",
				DryRun:     boolPtr(true),
			},
		},
		{
			name: "threads set to zero should still set ThreadsSet",
			setup: func(cmd *cobra.Command, flags *runtimeFlagSet) {
				flags.threads = 0
				cmd.Flags().Set("threads", "0")
			},
			expected: config.Overrides{
				Threads:    0,
				ThreadsSet: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command and flags for each test
			cmd := &cobra.Command{
				Use: "test",
			}
			flags := &runtimeFlagSet{}
			bindRuntimeFlags(cmd, flags)

			// Setup the test case
			tt.setup(cmd, flags)

			// Call toOverrides
			result := flags.toOverrides(cmd)

			// Compare results
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("toOverrides() mismatch\nGot:      %+v\nExpected: %+v", result, tt.expected)
			}
		})
	}
}

func TestRuntimeFlagSetToOverridesUnchangedFlags(t *testing.T) {
	// Create a command with flags bound
	cmd := &cobra.Command{
		Use: "test",
	}
	flags := &runtimeFlagSet{
		targets:     "https://default.com",
		targetsFile: "/default/targets.txt",
		mode:        "hybrid",
		threads:     10,
		outputDir:   "/default/output",
		formats:     "json",
		detectors:   "version",
		dryRun:      false,
		summaryFile: "/default/summary.json",
	}
	bindRuntimeFlags(cmd, flags)

	// Don't change any flags - just bind them with default values
	// The flags should not be marked as changed

	result := flags.toOverrides(cmd)

	// All fields should be zero/empty/nil since no flags were explicitly changed
	expected := config.Overrides{}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toOverrides() should return empty overrides when no flags changed\nGot:      %+v\nExpected: %+v", result, expected)
	}
}

// Helper function to create a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}
