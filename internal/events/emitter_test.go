package events

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// errorWriter is a writer that always returns an error.
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write failed")
}

// limitedWriter is a writer that fails after a certain number of bytes.
type limitedWriter struct {
	maxBytes int
	written  int
}

func (l *limitedWriter) Write(p []byte) (int, error) {
	if l.written >= l.maxBytes {
		return 0, errors.New("write limit exceeded")
	}
	remaining := l.maxBytes - l.written
	if len(p) <= remaining {
		l.written += len(p)
		return len(p), nil
	}
	// Can't write all data, return error per io.Writer contract
	l.written = l.maxBytes
	return remaining, errors.New("write limit exceeded")
}

func TestNewEmitter(t *testing.T) {
	buf := &bytes.Buffer{}
	emitter := NewEmitter(buf)

	if emitter == nil {
		t.Fatal("NewEmitter returned nil")
	}
	if emitter.writer != buf {
		t.Error("Emitter writer not set correctly")
	}
}

func TestEmit_AutomaticTimestampAssignment(t *testing.T) {
	tests := []struct {
		name      string
		event     Event
		wantZero  bool
		checkFunc func(t *testing.T, evt Event)
	}{
		{
			name: "zero timestamp gets assigned",
			event: Event{
				Type:    "test",
				Message: "test message",
			},
			wantZero: false,
			checkFunc: func(t *testing.T, evt Event) {
				if evt.Timestamp.IsZero() {
					t.Error("Expected timestamp to be assigned, but it's zero")
				}
				// Check that timestamp is recent (within last second)
				now := time.Now().UTC()
				diff := now.Sub(evt.Timestamp)
				if diff < 0 {
					diff = -diff
				}
				if diff > time.Second {
					t.Errorf("Timestamp is too old: %v, expected within last second", evt.Timestamp)
				}
			},
		},
		{
			name: "non-zero timestamp is preserved",
			event: Event{
				Type:      "test",
				Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				Message:   "test message",
			},
			wantZero: false,
			checkFunc: func(t *testing.T, evt Event) {
				expected := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
				if !evt.Timestamp.Equal(expected) {
					t.Errorf("Expected timestamp %v, got %v", expected, evt.Timestamp)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			emitter := NewEmitter(buf)

			err := emitter.Emit(tt.event)
			if err != nil {
				t.Fatalf("Emit() error = %v", err)
			}

			// Parse the written JSON to verify timestamp
			lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
			if len(lines) != 1 {
				t.Fatalf("Expected 1 line, got %d", len(lines))
			}

			var writtenEvent Event
			if err := json.Unmarshal([]byte(lines[0]), &writtenEvent); err != nil {
				t.Fatalf("Failed to unmarshal written event: %v", err)
			}

			tt.checkFunc(t, writtenEvent)
		})
	}
}

func TestEmit_JSONMarshalingEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		event   Event
		wantErr bool
	}{
		{
			name: "special characters in message",
			event: Event{
				Type:    "test",
				Message: "Hello\nWorld\tTab\"Quote\\Backslash",
			},
			wantErr: false,
		},
		{
			name: "unicode characters",
			event: Event{
				Type:    "test",
				Message: "Hello 世界 🌍",
			},
			wantErr: false,
		},
		{
			name: "empty message",
			event: Event{
				Type:    "test",
				Message: "",
			},
			wantErr: false,
		},
		{
			name: "nil fields map",
			event: Event{
				Type:   "test",
				Fields: nil,
			},
			wantErr: false,
		},
		{
			name: "empty fields map",
			event: Event{
				Type:   "test",
				Fields: map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name: "fields with various types",
			event: Event{
				Type: "test",
				Fields: map[string]interface{}{
					"string": "value",
					"int":    42,
					"float":  3.14,
					"bool":   true,
					"nil":    nil,
					"array":  []interface{}{1, 2, 3},
					"nested": map[string]interface{}{"key": "value"},
				},
			},
			wantErr: false,
		},
		{
			name: "very long message",
			event: Event{
				Type:    "test",
				Message: strings.Repeat("a", 10000),
			},
			wantErr: false,
		},
		{
			name: "empty type",
			event: Event{
				Type: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			emitter := NewEmitter(buf)

			err := emitter.Emit(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("Emit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the output is valid JSON
				lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
				if len(lines) != 1 {
					t.Fatalf("Expected 1 line, got %d", len(lines))
				}

				var writtenEvent Event
				if err := json.Unmarshal([]byte(lines[0]), &writtenEvent); err != nil {
					t.Errorf("Failed to unmarshal written event: %v. Output: %s", err, lines[0])
				}

				// Verify the output ends with newline
				if !strings.HasSuffix(buf.String(), "\n") {
					t.Error("Output should end with newline")
				}
			}
		})
	}
}

func TestEmit_ConcurrentEmission(t *testing.T) {
	buf := &bytes.Buffer{}
	emitter := NewEmitter(buf)

	const numGoroutines = 100
	const eventsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch multiple goroutines that emit events concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				event := Event{
					Type:    "concurrent_test",
					Message: "goroutine",
					Fields: map[string]interface{}{
						"goroutine_id": id,
						"event_id":     j,
					},
				}
				if err := emitter.Emit(event); err != nil {
					t.Errorf("Emit() error in goroutine %d: %v", id, err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all events were written
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	expectedLines := numGoroutines * eventsPerGoroutine
	if len(lines) != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, len(lines))
	}

	// Verify all events are valid JSON
	eventCounts := make(map[string]int)
	for _, line := range lines {
		if line == "" {
			continue
		}
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Errorf("Failed to unmarshal line: %v. Line: %s", err, line)
			continue
		}
		if evt.Type != "concurrent_test" {
			t.Errorf("Unexpected event type: %s", evt.Type)
		}
		eventCounts[line]++
	}

	// Verify no duplicate events (each should appear exactly once)
	for line, count := range eventCounts {
		if count != 1 {
			t.Errorf("Line appears %d times (expected 1): %s", count, line)
		}
	}

	// Verify timestamps are assigned
	for _, line := range lines {
		if line == "" {
			continue
		}
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			continue
		}
		if evt.Timestamp.IsZero() {
			t.Errorf("Event has zero timestamp: %s", line)
		}
	}
}

// errorMarshaler is a type that always fails to marshal to JSON.
type errorMarshaler struct{}

func (e errorMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errors.New("marshal error")
}

func TestEmit_ErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		writer  io.Writer
		event   Event
		wantErr bool
	}{
		{
			name:    "write error propagates",
			writer:  &errorWriter{},
			event:   Event{Type: "test", Message: "test"},
			wantErr: true,
		},
		{
			name:    "partial write error",
			writer:  &limitedWriter{maxBytes: 5},
			event:   Event{Type: "test", Message: "this is a long message"},
			wantErr: true,
		},
		{
			name:   "JSON marshaling error",
			writer: &bytes.Buffer{},
			event: Event{
				Type: "test",
				Fields: map[string]interface{}{
					"badField": errorMarshaler{},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emitter := NewEmitter(tt.writer)

			err := emitter.Emit(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("Emit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmit_OutputFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	emitter := NewEmitter(buf)

	event := Event{
		Type:    "test_type",
		Message: "test message",
		Fields: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	err := emitter.Emit(event)
	if err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	output := buf.String()
	if !strings.HasSuffix(output, "\n") {
		t.Error("Output should end with newline")
	}

	// Verify it's valid NDJSON (one JSON object per line)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}

	var writtenEvent Event
	if err := json.Unmarshal([]byte(lines[0]), &writtenEvent); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if writtenEvent.Type != "test_type" {
		t.Errorf("Expected type 'test_type', got '%s'", writtenEvent.Type)
	}
	if writtenEvent.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", writtenEvent.Message)
	}
	if writtenEvent.Fields["key1"] != "value1" {
		t.Errorf("Expected field key1='value1', got %v", writtenEvent.Fields["key1"])
	}
	if writtenEvent.Fields["key2"] != float64(42) { // JSON numbers are float64
		t.Errorf("Expected field key2=42, got %v", writtenEvent.Fields["key2"])
	}
}

func TestEmit_MultipleEvents(t *testing.T) {
	buf := &bytes.Buffer{}
	emitter := NewEmitter(buf)

	events := []Event{
		{Type: "event1", Message: "first"},
		{Type: "event2", Message: "second"},
		{Type: "event3", Message: "third"},
	}

	for _, event := range events {
		if err := emitter.Emit(event); err != nil {
			t.Fatalf("Emit() error = %v", err)
		}
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != len(events) {
		t.Errorf("Expected %d lines, got %d", len(events), len(lines))
	}

	for i, line := range lines {
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Errorf("Failed to unmarshal line %d: %v", i, err)
			continue
		}
		if evt.Type != events[i].Type {
			t.Errorf("Line %d: expected type '%s', got '%s'", i, events[i].Type, evt.Type)
		}
	}
}
