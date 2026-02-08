package main

import (
	"bytes"
	"sync"
)

// LogBuffer is a thread-safe ring buffer that implements io.Writer.
// It captures log output and notifies listeners when new entries arrive.
type LogBuffer struct {
	mu       sync.Mutex
	lines    []string
	maxLines int
	onChange func() // called when new lines are appended (can be nil)
}

// NewLogBuffer creates a new log buffer with the specified maximum number of lines.
// When the buffer exceeds maxLines, the oldest lines are discarded.
func NewLogBuffer(maxLines int) *LogBuffer {
	return &LogBuffer{
		lines:    make([]string, 0, maxLines),
		maxLines: maxLines,
	}
}

// Write implements io.Writer. It appends p to the buffer, splitting on newlines.
// Calls onChange callback if set (from a goroutine, so onChange should be thread-safe).
func (b *LogBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Split on newlines
	lines := bytes.Split(p, []byte("\n"))

	for i, line := range lines {
		// Skip empty lines (common from trailing newlines)
		if len(line) == 0 && i == len(lines)-1 {
			continue
		}

		str := string(line)
		b.lines = append(b.lines, str)

		// Enforce max size (ring buffer behavior)
		if len(b.lines) > b.maxLines {
			b.lines = b.lines[len(b.lines)-b.maxLines:]
		}
	}

	// Notify listener of new content (if registered)
	if b.onChange != nil {
		// Call onChange without lock to avoid deadlock if onChange tries to read
		onChange := b.onChange
		go onChange()
	}

	return len(p), nil
}

// Lines returns a copy of all buffered log lines.
// Safe to call concurrently with Write.
func (b *LogBuffer) Lines() []string {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Return a copy to avoid concurrent modification
	result := make([]string, len(b.lines))
	copy(result, b.lines)
	return result
}

// SetOnChange registers a callback invoked when new log lines arrive.
// The callback is called from a goroutine, so it should be thread-safe.
// Pass nil to unregister.
func (b *LogBuffer) SetOnChange(fn func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onChange = fn
}
