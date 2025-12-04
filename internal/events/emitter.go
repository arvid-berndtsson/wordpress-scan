package events

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

// Event represents a single NDJSON record for worker-friendly logs.
type Event struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Emitter writes NDJSON events to an io.Writer safely across goroutines.
type Emitter struct {
	writer io.Writer
	mu     sync.Mutex
}

// NewEmitter returns a new NDJSON emitter.
func NewEmitter(w io.Writer) *Emitter {
	return &Emitter{writer: w}
}

// Emit serializes the event to JSON and appends a newline.
func (e *Emitter) Emit(evt Event) error {
	if evt.Timestamp.IsZero() {
		evt.Timestamp = time.Now().UTC()
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, err := e.writer.Write(append(payload, '\n')); err != nil {
		return err
	}

	return nil
}
