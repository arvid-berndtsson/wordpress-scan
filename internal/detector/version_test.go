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
