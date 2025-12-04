package detector

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var versionRegex = regexp.MustCompile(`WordPress\s+([0-9]+\.[0-9]+(\.[0-9]+)?)`)

// VersionDetector inspects the target homepage for WordPress generator metadata.
type VersionDetector struct {
	client       *http.Client
	maxBodyBytes int64
}

// NewVersionDetector builds a detector with an optional custom HTTP client.
func NewVersionDetector(client *http.Client) *VersionDetector {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &VersionDetector{client: client, maxBodyBytes: 1024 * 1024}
}

// Name implements Detector.
func (d *VersionDetector) Name() string {
	return "version"
}

// Detect fetches the target root document and scans for a generator meta tag.
func (d *VersionDetector) Detect(ctx context.Context, target string) (Result, error) {
	url := normalizeTargetURL(target)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return Result{}, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	reader := io.LimitReader(resp.Body, d.maxBodyBytes)
	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		return Result{}, err
	}

	matches := versionRegex.FindSubmatch(bodyBytes)
	if len(matches) < 2 {
		return Result{}, errors.New("version not discovered in generator tag")
	}

	version := string(matches[1])
	return Result{
		Target:     target,
		Detector:   d.Name(),
		Severity:   "info",
		Summary:    fmt.Sprintf("WordPress version %s detected", version),
		Metadata:   map[string]interface{}{"version": version, "source": "meta-generator"},
		Confidence: 0.85,
	}, nil
}

func normalizeTargetURL(target string) string {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return target
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}
	return "https://" + trimmed
}
