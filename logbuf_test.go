package main

import (
	"strings"
	"sync"
	"testing"
)

func TestLogBufferWrite(t *testing.T) {
	buf := NewLogBuffer(10)

	// Write a single line
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
	buf := NewLogBuffer(3) // max 3 lines

	buf.Write([]byte("line 1\n"))
	buf.Write([]byte("line 2\n"))
	buf.Write([]byte("line 3\n"))
	buf.Write([]byte("line 4\n")) // Should evict line 1

	lines := buf.Lines()
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines (ring buffer), got %d", len(lines))
	}

	// Oldest line (line 1) should be gone
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

	// 10 goroutines writing concurrently
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
	// Should have 100 lines total (or fewer if ring buffer kicked in)
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

	// lines1 should not have been mutated
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

	// Simulating io.MultiWriter behavior
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
