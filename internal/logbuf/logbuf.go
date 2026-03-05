package logbuf

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
// Calls onChange callback (if set) after releasing the lock to avoid deadlocks.
func (b *LogBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()

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

	// Capture callback before releasing lock
	onChange := b.onChange
	b.mu.Unlock()

	// Notify listener after releasing the lock to avoid deadlock
	// and without spawning a goroutine per write.
	if onChange != nil {
		onChange()
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

// Clear removes all buffered log lines and notifies the listener.
func (b *LogBuffer) Clear() {
	b.mu.Lock()
	b.lines = b.lines[:0]
	onChange := b.onChange
	b.mu.Unlock()

	if onChange != nil {
		onChange()
	}
}

// SetOnChange registers a callback invoked when new log lines arrive.
// The callback is called synchronously from the writing goroutine (after
// releasing the lock), so it should not block for long. Pass nil to unregister.
func (b *LogBuffer) SetOnChange(fn func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onChange = fn
}
