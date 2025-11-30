package cli

import (
	"github.com/example/wp-worker/internal/config"
	"github.com/spf13/cobra"
)

// Execute builds the root command tree and runs the CLI.
func Execute() error {
	loader := &config.Loader{ConfigPath: config.DefaultConfigPath}
	rootOpts := &rootOptions{}

	rootCmd := &cobra.Command{
		Use:           "wp-worker-cli",
		Short:         "Worker-friendly wrapper around wpprobe",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}
	rootCmd.SetVersionTemplate("wp-worker-cli version {{.Version}}\n")

	rootCmd.PersistentFlags().StringVar(&rootOpts.ConfigPath, "config", config.DefaultConfigPath, "Path to worker.config.yml (optional)")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if rootOpts.ConfigPath != "" {
			loader.ConfigPath = rootOpts.ConfigPath
		}
	}

	rootCmd.AddCommand(
		newInitCmd(loader),
		newScanCmd(loader),
		newReportCmd(),
	)

	return rootCmd.Execute()
}

type rootOptions struct {
	ConfigPath string
}
