package Console

import (
	"strings"
	"sync"
)

// LogBuffer keeps the last lines written by the standard logger for the console screen.
type LogBuffer struct {
	mu    sync.Mutex
	lines []string
}

func (l *LogBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, line := range strings.Split(strings.TrimRight(string(p), "\n"), "\n") {
		l.lines = append(l.lines, line)
	}
	if len(l.lines) > 200 {
		l.lines = l.lines[len(l.lines)-200:]
	}
	return len(p), nil
}

func (l *LogBuffer) Last(n int) []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.lines) <= n {
		return append([]string(nil), l.lines...)
	}
	return append([]string(nil), l.lines[len(l.lines)-n:]...)
}
