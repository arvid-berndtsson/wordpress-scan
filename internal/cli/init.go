package cli

import (
	"fmt"

	"github.com/example/wphunter/internal/config"
	"github.com/example/wphunter/internal/wpprobe"
	"github.com/spf13/cobra"
)

func newInitCmd(loader *config.Loader) *cobra.Command {
	flags := &runtimeFlagSet{}
	var skipBinaryCheck bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Validate the execution environment and configuration",
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

			if !skipBinaryCheck && !cfg.DryRun {
				runner := wpprobe.NewRunner()
				if err := runner.EnsureBinary(); err != nil {
					return err
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Environment looks good. Output will be stored in %s\n", cfg.OutputDir)
			return nil
		},
	}

	bindRuntimeFlags(cmd, flags)
	cmd.Flags().BoolVar(&skipBinaryCheck, "skip-wpprobe-check", false, "Allow init to pass even if wpprobe is missing (useful for dry-run mode)")

	return cmd
}
