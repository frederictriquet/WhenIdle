package logbuf

import (
	"strings"
	"sync"
	"testing"
)

func TestLogBufferWrite(t *testing.T) {
	buf := NewLogBuffer(10)

	n, err := buf.Write([]byte("line 1\n"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 7 {
		t.Errorf("Expected n=7, got %d", n)
	}

	lines := buf.Lines()
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}
	if lines[0] != "line 1" {
		t.Errorf("Expected 'line 1', got %q", lines[0])
	}
}

func TestLogBufferMultipleLines(t *testing.T) {
	buf := NewLogBuffer(10)

	buf.Write([]byte("line 1\nline 2\nline 3\n"))

	lines := buf.Lines()
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line 1" || lines[1] != "line 2" || lines[2] != "line 3" {
		t.Errorf("Unexpected lines: %v", lines)
	}
}

func TestLogBufferRingBehavior(t *testing.T) {
	buf := NewLogBuffer(3)

	buf.Write([]byte("line 1\n"))
	buf.Write([]byte("line 2\n"))
	buf.Write([]byte("line 3\n"))
	buf.Write([]byte("line 4\n"))

	lines := buf.Lines()
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines (ring buffer), got %d", len(lines))
	}

	if lines[0] != "line 2" || lines[1] != "line 3" || lines[2] != "line 4" {
		t.Errorf("Expected [line 2, line 3, line 4], got %v", lines)
	}
}

func TestLogBufferOnChange(t *testing.T) {
	buf := NewLogBuffer(10)

	callCount := 0
	buf.SetOnChange(func() {
		callCount++
	})

	buf.Write([]byte("line 1\n"))
	buf.Write([]byte("line 2\n"))

	if callCount != 2 {
		t.Errorf("Expected onChange called 2 times, got %d", callCount)
	}
}

func TestLogBufferConcurrent(t *testing.T) {
	buf := NewLogBuffer(100)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				buf.Write([]byte("line from goroutine " + string(rune('0'+id)) + "\n"))
			}
		}(i)
	}

	wg.Wait()

	lines := buf.Lines()
	if len(lines) == 0 {
		t.Error("Expected some lines after concurrent writes")
	}
}

func TestLogBufferLinesReturnsACopy(t *testing.T) {
	buf := NewLogBuffer(10)
	buf.Write([]byte("line 1\n"))

	lines1 := buf.Lines()
	buf.Write([]byte("line 2\n"))
	lines2 := buf.Lines()

	if len(lines1) != 1 {
		t.Errorf("Expected lines1 to still have 1 element, got %d", len(lines1))
	}
	if len(lines2) != 2 {
		t.Errorf("Expected lines2 to have 2 elements, got %d", len(lines2))
	}
}

func TestLogBufferWithMultiWriter(t *testing.T) {
	buf := NewLogBuffer(10)
	var sb strings.Builder

	data := []byte("test line\n")
	buf.Write(data)
	sb.Write(data)

	lines := buf.Lines()
	if len(lines) != 1 || lines[0] != "test line" {
		t.Errorf("Buffer: expected 'test line', got %v", lines)
	}

	if sb.String() != "test line\n" {
		t.Errorf("StringBuilder: expected 'test line\\n', got %q", sb.String())
	}
}

func TestLogBufferClear(t *testing.T) {
	buf := NewLogBuffer(10)
	buf.Write([]byte("line 1\n"))
	buf.Write([]byte("line 2\n"))

	if len(buf.Lines()) != 2 {
		t.Fatalf("Expected 2 lines before clear")
	}

	buf.Clear()

	if len(buf.Lines()) != 0 {
		t.Errorf("Expected 0 lines after clear, got %d", len(buf.Lines()))
	}
}

func TestLogBufferClearCallsOnChange(t *testing.T) {
	buf := NewLogBuffer(10)
	buf.Write([]byte("line 1\n"))

	called := false
	buf.SetOnChange(func() { called = true })
	buf.Clear()

	if !called {
		t.Error("Expected onChange to be called on Clear")
	}
}

func TestLogBufferSetOnChangeNil(t *testing.T) {
	buf := NewLogBuffer(10)

	buf.SetOnChange(func() {})
	buf.SetOnChange(nil) // unregister — must not panic

	buf.Write([]byte("line 1\n")) // should not call nil callback
	buf.Clear()                   // should not call nil callback
}

func TestLogBufferWriteMiddleEmptyLine(t *testing.T) {
	buf := NewLogBuffer(10)

	// Empty line in the middle (not a trailing newline) is kept
	buf.Write([]byte("line 1\n\nline 3\n"))

	lines := buf.Lines()
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines (including middle empty), got %d: %v", len(lines), lines)
	}
	if lines[0] != "line 1" || lines[1] != "" || lines[2] != "line 3" {
		t.Errorf("Expected [line 1, , line 3], got %v", lines)
	}
}
