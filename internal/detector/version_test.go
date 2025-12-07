package detector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersionDetectorDetectsGeneratorMeta(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><head><meta name="generator" content="WordPress 6.5.1" /></head></html>`))
	}))

	defer ts.Close()

	detector := NewVersionDetector(ts.Client())
	res, err := detector.Detect(context.Background(), ts.URL)
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}

	if res.Metadata["version"] != "6.5.1" {
		t.Fatalf("expected version 6.5.1, got %v", res.Metadata)
	}
}

func TestVersionDetectorHandlesMissingVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><head></head><body>No generator</body></html>`))
	}))
	defer ts.Close()

	detector := NewVersionDetector(ts.Client())
	if _, err := detector.Detect(context.Background(), ts.URL); err == nil {
		t.Fatalf("expected error when generator missing")
	}
}

func TestNormalizeTargetURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "   ",
		},
		{
			name:     "URL with https scheme",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "URL with http scheme",
			input:    "http://example.com",
			expected: "http://example.com",
		},
		{
			name:     "URL without scheme",
			input:    "example.com",
			expected: "https://example.com",
		},
		{
			name:     "URL with leading whitespace",
			input:    "  example.com",
			expected: "https://example.com",
		},
		{
			name:     "URL with trailing whitespace",
			input:    "example.com  ",
			expected: "https://example.com",
		},
		{
			name:     "URL with leading and trailing whitespace",
			input:    "  example.com  ",
			expected: "https://example.com",
		},
		{
			name:     "URL with https scheme and whitespace",
			input:    "  https://example.com  ",
			expected: "https://example.com",
		},
		{
			name:     "URL with http scheme and whitespace",
			input:    "  http://example.com  ",
			expected: "http://example.com",
		},
		{
			name:     "URL without scheme with path",
			input:    "example.com/path",
			expected: "https://example.com/path",
		},
		{
			name:     "URL with scheme and path",
			input:    "https://example.com/path",
			expected: "https://example.com/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeTargetURL(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeTargetURL(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
