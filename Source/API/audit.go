package API

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"
)

// AuditEntry is one admin action. Silicon logs every change it forwards, so moderation has a trail
// whichever backend it manages.
type AuditEntry struct {
	At     time.Time `json:"at"`
	User   string    `json:"user"`
	IP     string    `json:"ip"`
	Action string    `json:"action"`
	Path   string    `json:"path"`
	Body   string    `json:"body,omitempty"`
	Status int       `json:"status"`
}

type AuditLog struct {
	mu   sync.Mutex
	path string
}

func NewAuditLog(path string) *AuditLog { return &AuditLog{path: path} }

func (l *AuditLog) Add(e AuditEntry) {
	if l.path == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	data, _ := json.Marshal(e)
	f.Write(append(data, '\n'))
}

// Recent returns up to limit entries, newest first.
func (l *AuditLog) Recent(limit int) []AuditEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := []AuditEntry{}
	f, err := os.Open(l.path)
	if err != nil {
		return out
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		var e AuditEntry
		if json.Unmarshal(scanner.Bytes(), &e) == nil {
			out = append(out, e)
		}
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}
