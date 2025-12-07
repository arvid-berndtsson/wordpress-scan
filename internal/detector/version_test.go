package detector

import (
	"context"
	"fmt"
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

func TestVersionDetectorHandlesHTTPErrorStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "400 Bad Request",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "401 Unauthorized",
			statusCode: http.StatusUnauthorized,
		},
		{
			name:       "403 Forbidden",
			statusCode: http.StatusForbidden,
		},
		{
			name:       "404 Not Found",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "502 Bad Gateway",
			statusCode: http.StatusBadGateway,
		},
		{
			name:       "503 Service Unavailable",
			statusCode: http.StatusServiceUnavailable,
		},
		{
			name:       "504 Gateway Timeout",
			statusCode: http.StatusGatewayTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(`<html><body>Error page</body></html>`))
			}))
			defer ts.Close()

			detector := NewVersionDetector(ts.Client())
			_, err := detector.Detect(context.Background(), ts.URL)
			if err == nil {
				t.Fatalf("expected error for status code %d, got nil", tt.statusCode)
			}

			expectedErrorMsg := fmt.Sprintf("unexpected status code %d", tt.statusCode)
			if err.Error() != expectedErrorMsg {
				t.Fatalf("expected error message %q, got %q", expectedErrorMsg, err.Error())
			}
		})
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
		{
			name:     "whitespace only with tabs",
			input:    "\t\t",
			expected: "\t\t",
		},
		{
			name:     "whitespace only with newlines",
			input:    "\n\n",
			expected: "\n\n",
		},
		{
			name:     "whitespace only with mixed whitespace",
			input:    " \t\n ",
			expected: " \t\n ",
		},
		{
			name:     "URL with uppercase HTTP scheme",
			input:    "HTTP://example.com",
			expected: "https://HTTP://example.com",
		},
		{
			name:     "URL with uppercase HTTPS scheme",
			input:    "HTTPS://example.com",
			expected: "https://HTTPS://example.com",
		},
		{
			name:     "URL with mixed case HTTP scheme",
			input:    "Http://example.com",
			expected: "https://Http://example.com",
		},
		{
			name:     "URL with mixed case HTTPS scheme",
			input:    "Https://example.com",
			expected: "https://Https://example.com",
		},
		{
			name:     "URL without scheme with port",
			input:    "example.com:8080",
			expected: "https://example.com:8080",
		},
		{
			name:     "URL with http scheme and port",
			input:    "http://example.com:8080",
			expected: "http://example.com:8080",
		},
		{
			name:     "URL without scheme with query string",
			input:    "example.com?param=value",
			expected: "https://example.com?param=value",
		},
		{
			name:     "URL without scheme with fragment",
			input:    "example.com#section",
			expected: "https://example.com#section",
		},
		{
			name:     "URL without scheme with path, query, and fragment",
			input:    "example.com/path?param=value#section",
			expected: "https://example.com/path?param=value#section",
		},
		{
			name:     "URL with tabs and newlines around it",
			input:    "\t\nexample.com\n\t",
			expected: "https://example.com",
		},
		{
			name:     "URL with http scheme and tabs/newlines",
			input:    "\t\nhttp://example.com\n\t",
			expected: "http://example.com",
		},
		{
			name:     "single space",
			input:    " ",
			expected: " ",
		},
		{
			name:     "single tab",
			input:    "\t",
			expected: "\t",
		},
		{
			name:     "single newline",
			input:    "\n",
			expected: "\n",
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
