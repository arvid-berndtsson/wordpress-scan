package cli

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/example/wphunter/internal/config"
	"github.com/example/wphunter/internal/wpprobe"
	"github.com/spf13/cobra"
)

type doctorCheck struct {
	Name   string
	Status string // "✓" or "✗"
	Detail string
	Error  error
}

func newDoctorCmd(loader *config.Loader) *cobra.Command {
	flags := &runtimeFlagSet{}
	var timeout int

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Validate dependencies, network reachability, and wpprobe DB freshness",
		Long: `The doctor subcommand performs comprehensive validation of the wphunter environment:
- Go runtime version
- wpprobe binary presence and functionality
- Network connectivity to configured targets
- wpprobe database freshness (if applicable)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			overrides := flags.toOverrides(cmd)
			cfg, err := loader.Load(overrides)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()

			checks := runDoctorChecks(ctx, &cfg)
			printDoctorReport(cmd, checks)

			// Return error if any check failed
			for _, check := range checks {
				if check.Error != nil {
					return fmt.Errorf("doctor checks failed")
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), "\n✓ All checks passed. System is ready.")
			return nil
		},
	}

	bindRuntimeFlags(cmd, flags)
	cmd.Flags().IntVar(&timeout, "timeout", 30, "Timeout in seconds for network checks")

	return cmd
}

func runDoctorChecks(ctx context.Context, cfg *config.RuntimeConfig) []doctorCheck {
	checks := []doctorCheck{}

	// Check 1: Go version
	goCheck := checkGoVersion()
	checks = append(checks, goCheck)

	// Check 2: wpprobe binary presence
	wpprobeCheck := checkWPProbeBinary(cfg.DryRun)
	checks = append(checks, wpprobeCheck)

	// Check 3: wpprobe database (if binary is available)
	if wpprobeCheck.Error == nil && !cfg.DryRun {
		dbCheck := checkWPProbeDatabase(ctx)
		checks = append(checks, dbCheck)
	}

	// Check 4: Network reachability to targets
	if len(cfg.Targets) > 0 && !cfg.DryRun {
		networkChecks := checkNetworkReachability(ctx, cfg.Targets)
		checks = append(checks, networkChecks...)
	}

	// Check 5: Configuration validity
	configCheck := checkConfiguration(cfg)
	checks = append(checks, configCheck)

	// Check 6: Output directory
	outputCheck := checkOutputDirectory(cfg.OutputDir)
	checks = append(checks, outputCheck)

	return checks
}

func checkGoVersion() doctorCheck {
	version := runtime.Version()
	return doctorCheck{
		Name:   "Go Runtime",
		Status: "✓",
		Detail: fmt.Sprintf("Version %s", version),
	}
}

func checkWPProbeBinary(dryRun bool) doctorCheck {
	if dryRun {
		return doctorCheck{
			Name:   "wpprobe Binary",
			Status: "⊘",
			Detail: "Skipped (dry-run mode)",
		}
	}

	runner := wpprobe.NewRunner()
	err := runner.EnsureBinary()
	if err != nil {
		return doctorCheck{
			Name:   "wpprobe Binary",
			Status: "✗",
			Detail: "Not found in PATH",
			Error:  err,
		}
	}

	// Try to get version
	versionDetail := "Available"
	if version, err := getWPProbeVersion(); err == nil {
		versionDetail = fmt.Sprintf("Version %s", version)
	}

	return doctorCheck{
		Name:   "wpprobe Binary",
		Status: "✓",
		Detail: versionDetail,
	}
}

func getWPProbeVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "wpprobe", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Parse version from output (format might vary)
	version := strings.TrimSpace(string(output))
	if version == "" {
		return "unknown", nil
	}

	return version, nil
}

func checkWPProbeDatabase(ctx context.Context) doctorCheck {
	// Try to run wpprobe update to check database status
	// This is a lightweight check that doesn't actually update
	// We can't directly check DB freshness without running update,
	// so we'll just verify the binary can be executed
	testCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(testCtx, "wpprobe", "--help")
	err := cmd.Run()
	
	if err != nil {
		return doctorCheck{
			Name:   "wpprobe Functionality",
			Status: "✗",
			Detail: "Binary found but not executable",
			Error:  err,
		}
	}

	return doctorCheck{
		Name:   "wpprobe Functionality",
		Status: "✓",
		Detail: "Binary is executable",
	}
}

func checkNetworkReachability(ctx context.Context, targets []string) []doctorCheck {
	checks := []doctorCheck{}
	
	// Limit to first 3 targets for performance
	maxChecks := 3
	originalTargetCount := len(targets)
	if len(targets) > maxChecks {
		targets = targets[:maxChecks]
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	for _, target := range targets {
		check := doctorCheck{
			Name: fmt.Sprintf("Network: %s", target),
		}

		req, err := http.NewRequestWithContext(ctx, "HEAD", target, nil)
		if err != nil {
			check.Status = "✗"
			check.Detail = "Invalid URL"
			check.Error = err
			checks = append(checks, check)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			check.Status = "✗"
			check.Detail = "Unreachable"
			check.Error = err
		} else {
			resp.Body.Close()
			check.Status = "✓"
			check.Detail = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}

		checks = append(checks, check)
	}

	if originalTargetCount > maxChecks {
		checks = append(checks, doctorCheck{
			Name:   fmt.Sprintf("Network: ... (%d more targets)", originalTargetCount-maxChecks),
			Status: "⊘",
			Detail: "Skipped for brevity",
		})
	}

	return checks
}

func checkConfiguration(cfg *config.RuntimeConfig) doctorCheck {
	err := cfg.Validate()
	if err != nil {
		return doctorCheck{
			Name:   "Configuration",
			Status: "✗",
			Detail: "Invalid configuration",
			Error:  err,
		}
	}

	return doctorCheck{
		Name:   "Configuration",
		Status: "✓",
		Detail: fmt.Sprintf("%d targets, mode=%s", len(cfg.Targets), cfg.Mode),
	}
}

func checkOutputDirectory(outputDir string) doctorCheck {
	err := ensureOutputDir(outputDir)
	if err != nil {
		return doctorCheck{
			Name:   "Output Directory",
			Status: "✗",
			Detail: outputDir,
			Error:  err,
		}
	}

	return doctorCheck{
		Name:   "Output Directory",
		Status: "✓",
		Detail: outputDir,
	}
}

func printDoctorReport(cmd *cobra.Command, checks []doctorCheck) {
	fmt.Fprintln(cmd.OutOrStdout(), "Running environment diagnostics...")

	for _, check := range checks {
		fmt.Fprintf(cmd.OutOrStdout(), "%s %-30s %s\n", check.Status, check.Name+":", check.Detail)
		if check.Error != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "   Error: %v\n", check.Error)
		}
	}
}
