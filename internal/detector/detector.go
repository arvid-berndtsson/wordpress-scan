package detector

import "context"

// Result represents a single detector finding for a target.
type Result struct {
	Target     string                 `json:"target"`
	Detector   string                 `json:"detector"`
	Severity   string                 `json:"severity"`
	Summary    string                 `json:"summary"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Confidence float64                `json:"confidence,omitempty"`
}

// Detector is implemented by modules that can analyze a target.
type Detector interface {
	Name() string
	Detect(ctx context.Context, target string) (Result, error)
}
