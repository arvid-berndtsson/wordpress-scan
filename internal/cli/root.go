package cli

import (
	"github.com/example/wphunter/internal/config"
	"github.com/spf13/cobra"
)

// Execute builds the root command tree and runs the CLI.
func Execute() error {
	loader := &config.Loader{ConfigPath: config.DefaultConfigPath}
	rootOpts := &rootOptions{}

	rootCmd := &cobra.Command{
		Use:           "wphunter",
		Short:         "Red/blue WordPress scanner with modular detectors",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}
	rootCmd.SetVersionTemplate("wphunter version {{.Version}}\n")

	rootCmd.PersistentFlags().StringVar(&rootOpts.ConfigPath, "config", config.DefaultConfigPath, "Path to wphunter.config.yml (optional)")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if rootOpts.ConfigPath != "" {
			loader.ConfigPath = rootOpts.ConfigPath
		}
	}

	rootCmd.AddCommand(
		newInitCmd(loader),
		newScanCmd(loader),
		newReportCmd(),
		newDoctorCmd(loader),
	)

	return rootCmd.Execute()
}

type rootOptions struct {
	ConfigPath string
}
