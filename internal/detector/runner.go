package detector

import (
	"context"
	"fmt"
)

// Registry maps detector names to constructors.
type Registry map[string]Factory

// Factory builds a detector instance.
type Factory func() Detector

// DefaultRegistry contains built-in detectors.
var DefaultRegistry = Registry{
	"version": func() Detector { return NewVersionDetector(nil) },
}

// BuildDetectors instantiates detectors from the provided names.
func (r Registry) BuildDetectors(names []string) ([]Detector, error) {
	if len(names) == 0 {
		return nil, nil
	}

	var detectors []Detector
	seen := map[string]struct{}{}
	for _, name := range names {
		factory, ok := r[name]
		if !ok {
			return nil, fmt.Errorf("unknown detector: %s", name)
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		detectors = append(detectors, factory())
	}
	return detectors, nil
}

// Run executes detectors sequentially for each target.
func Run(ctx context.Context, detectors []Detector, targets []string) ([]Result, error) {
	if len(detectors) == 0 || len(targets) == 0 {
		return nil, nil
	}

	var results []Result
	for _, target := range targets {
		for _, detector := range detectors {
			select {
			case <-ctx.Done():
				return results, ctx.Err()
			default:
			}

			result, err := detector.Detect(ctx, target)
			if err != nil {
				results = append(results, Result{
					Target:   target,
					Detector: detector.Name(),
					Severity: "info",
					Summary:  fmt.Sprintf("detector error: %v", err),
				})
				continue
			}
			results = append(results, result)
		}
	}

	return results, nil
}
