package logger

import (
	"bytes"
	"sync"
)

// Ring captures the last N lines written to it.
type Ring struct {
	mu     sync.RWMutex
	lines  []string
	size   int
	head   int
	count  int
}

func NewRing(size int) *Ring {
	return &Ring{
		lines: make([]string, size),
		size:  size,
	}
}

func (r *Ring) Write(p []byte) (n int, err error) {
	// Split on newlines; keep partial line for next write.
	r.mu.Lock()
	defer r.mu.Unlock()

	parts := bytes.Split(p, []byte("\n"))
	for i, part := range parts {
		if i == len(parts)-1 && len(part) == 0 {
			continue
		}
		r.lines[r.head] = string(part)
		r.head = (r.head + 1) % r.size
		if r.count < r.size {
			r.count++
		}
	}
	return len(p), nil
}

// Lines returns the captured lines in chronological order (oldest first).
func (r *Ring) Lines() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.count == 0 {
		return nil
	}

	out := make([]string, 0, r.count)
	start := 0
	if r.count == r.size {
		start = r.head
	}
	for i := 0; i < r.count; i++ {
		idx := (start + i) % r.size
		out = append(out, r.lines[idx])
	}
	return out
}

// Clear empties the buffer.
func (r *Ring) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.head = 0
	r.count = 0
	for i := range r.lines {
		r.lines[i] = ""
	}
}
