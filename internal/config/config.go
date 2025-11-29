package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath = "worker.config.yml"

	envTargets     = "WORKER_TARGETS"
	envTargetsFile = "WORKER_TARGETS_FILE"
	envMode        = "WORKER_MODE"
	envThreads     = "WORKER_THREADS"
	envOutputDir   = "WORKER_OUTPUT_DIR"
	envFormats     = "WORKER_FORMATS"
	envDryRun      = "WORKER_DRY_RUN"
	envSummaryFile = "WORKER_SUMMARY_FILE"
)

// Loader merges configuration coming from files, environment variables, and CLI flags.
type Loader struct {
	ConfigPath string
}

// RuntimeConfig contains the fully merged settings required by worker sub-commands.
type RuntimeConfig struct {
	Targets     []string
	Mode        string
	Threads     int
	OutputDir   string
	Formats     []string
	DryRun      bool
	SummaryFile string
}

// Overrides captures values coming from env vars or CLI flags.
type Overrides struct {
	Targets     []string
	TargetsFile string
	Mode        string
	Threads     int
	ThreadsSet  bool
	OutputDir   string
	Formats     []string
	DryRun      *bool
	SummaryFile string
}

// DefaultRuntimeConfig returns the baseline configuration when no overrides are provided.
func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		Mode:      "hybrid",
		Threads:   10,
		OutputDir: "scan-results",
		Formats:   []string{"json", "csv"},
	}
}

// Load resolves the final runtime configuration.
func (l Loader) Load(override Overrides) (RuntimeConfig, error) {
	cfg := DefaultRuntimeConfig()
	path := l.ConfigPath
	if path == "" {
		path = DefaultConfigPath
	}

	if fileExists(path) {
		fileOv, err := loadFromFile(path)
		if err != nil {
			return cfg, err
		}
		if err := cfg.apply(fileOv); err != nil {
			return cfg, err
		}
	}

	if err := cfg.apply(overridesFromEnv()); err != nil {
		return cfg, err
	}

	if err := cfg.apply(override); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Validate ensures the config contains the minimum required data for scan/init commands.
func (c RuntimeConfig) Validate() error {
	if len(c.Targets) == 0 {
		return errors.New("no targets configured; provide --targets, --targets-file, or set WORKER_TARGETS")
	}

	if c.Threads < 1 || c.Threads > 64 {
		return fmt.Errorf("threads must be between 1 and 64 (got %d)", c.Threads)
	}

	if c.Mode == "" {
		return errors.New("scan mode must be specified")
	}

	if len(c.Formats) == 0 {
		return errors.New("at least one output format must be specified")
	}

	if c.OutputDir == "" {
		return errors.New("output directory cannot be empty")
	}

	return nil
}

func (c *RuntimeConfig) apply(src Overrides) error {
	if len(src.Targets) > 0 {
		c.Targets = cleanList(src.Targets)
	}

	if src.TargetsFile != "" {
		values, err := readTargetsFile(src.TargetsFile)
		if err != nil {
			return err
		}
		c.Targets = values
	}

	if src.Mode != "" {
		c.Mode = src.Mode
	}

	if src.ThreadsSet {
		c.Threads = src.Threads
	}

	if src.OutputDir != "" {
		c.OutputDir = src.OutputDir
	}

	if len(src.Formats) > 0 {
		c.Formats = cleanList(src.Formats)
	}

	if src.DryRun != nil {
		c.DryRun = *src.DryRun
	}

	if src.SummaryFile != "" {
		c.SummaryFile = src.SummaryFile
	}

	return nil
}

func loadFromFile(path string) (Overrides, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Overrides{}, err
	}

	type rawConfig struct {
		Targets     targetList `yaml:"targets"`
		TargetsFile string     `yaml:"targetsFile"`
		Mode        string     `yaml:"mode"`
		Threads     *int       `yaml:"threads"`
		OutputDir   string     `yaml:"outputDir"`
		Formats     []string   `yaml:"formats"`
		DryRun      *bool      `yaml:"dryRun"`
		SummaryFile string     `yaml:"summaryFile"`
	}

	var raw rawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Overrides{}, err
	}

	over := Overrides{
		Targets:     raw.Targets,
		TargetsFile: raw.TargetsFile,
		Mode:        raw.Mode,
		OutputDir:   raw.OutputDir,
		Formats:     raw.Formats,
		SummaryFile: raw.SummaryFile,
	}

	if raw.Threads != nil {
		over.Threads = *raw.Threads
		over.ThreadsSet = true
	}

	if raw.DryRun != nil {
		over.DryRun = raw.DryRun
	}

	return over, nil
}

func overridesFromEnv() Overrides {
	ov := Overrides{}

	if value := os.Getenv(envTargets); value != "" {
		ov.Targets = ParseTargetsList(value)
	}

	if value := os.Getenv(envTargetsFile); value != "" {
		ov.TargetsFile = value
	}

	if value := os.Getenv(envMode); value != "" {
		ov.Mode = value
	}

	if value := os.Getenv(envThreads); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			ov.Threads = parsed
			ov.ThreadsSet = true
		}
	}

	if value := os.Getenv(envOutputDir); value != "" {
		ov.OutputDir = value
	}

	if value := os.Getenv(envFormats); value != "" {
		ov.Formats = ParseFormats(value)
	}

	if value := os.Getenv(envDryRun); value != "" {
		parsed := strings.EqualFold(value, "true") || value == "1"
		ov.DryRun = &parsed
	}

	if value := os.Getenv(envSummaryFile); value != "" {
		ov.SummaryFile = value
	}

	return ov
}

// ParseTargetsList turns comma or newline separated input into individual targets.
func ParseTargetsList(input string) []string {
	return splitOnDelimiters(input, []rune{',', '\n', '\r'})
}

// ParseFormats splits comma separated format strings.
func ParseFormats(input string) []string {
	return splitOnDelimiters(input, []rune{',', '\n', '\r', ' '})
}

func splitOnDelimiters(input string, delims []rune) []string {
	if input == "" {
		return nil
	}

	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil
	}

	separator := func(r rune) bool {
		for _, d := range delims {
			if r == d {
				return true
			}
		}
		return false
	}

	parts := strings.FieldsFunc(trimmed, separator)
	return cleanList(parts)
}

func cleanList(values []string) []string {
	var out []string
	for _, v := range values {
		candidate := strings.TrimSpace(v)
		if candidate != "" {
			out = append(out, candidate)
		}
	}
	return out
}

func readTargetsFile(path string) ([]string, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var targets []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		targets = append(targets, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return targets, nil
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// targetList enables YAML fields that can be specified as a scalar or sequence.
type targetList []string

func (t *targetList) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.SequenceNode:
		var out []string
		for _, node := range value.Content {
			out = append(out, strings.TrimSpace(node.Value))
		}
		*t = cleanList(out)
	case yaml.ScalarNode:
		*t = ParseTargetsList(value.Value)
	default:
		return fmt.Errorf("unsupported YAML type for targets")
	}
	return nil
}
