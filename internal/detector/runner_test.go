package detector

import (
	"context"
	"errors"
	"testing"
)

type fakeDetector struct {
	name   string
	result Result
	err    error
}

func (f fakeDetector) Name() string { return f.name }

func (f fakeDetector) Detect(ctx context.Context, target string) (Result, error) {
	return f.result, f.err
}

func TestRunAggregatesResults(t *testing.T) {
	dets := []Detector{
		fakeDetector{name: "one", result: Result{Target: "https://example", Detector: "one"}},
		fakeDetector{name: "two", err: errors.New("boom")},
	}

	results, err := Run(context.Background(), dets, []string{"https://example"})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestRegistryBuildDetectors(t *testing.T) {
	r := Registry{
		"fake": func() Detector { return fakeDetector{name: "fake"} },
	}

	dets, err := r.BuildDetectors([]string{"fake"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dets) != 1 || dets[0].Name() != "fake" {
		t.Fatalf("unexpected detectors: %#v", dets)
	}
}
