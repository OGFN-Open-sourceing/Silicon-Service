package Backend

import (
	"context"
	"sync"
	"time"
)

// Status is the backend's GET /status answer (API.md).
type Status struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	UptimeSeconds int64  `json:"uptimeSeconds"`
	Online        int    `json:"online"`
	InQueue       int    `json:"inQueue"`
	InMatch       int    `json:"inMatch"`
	Servers       int    `json:"servers"`
}

type Sample struct {
	At     time.Time `json:"at"`
	Up     bool      `json:"up"`
	Status *Status   `json:"status,omitempty"`
	Error  string    `json:"error,omitempty"`
}

// Heartbeat polls the backend's status and keeps the last day of samples for the activity chart.
type Heartbeat struct {
	client   *Client
	interval time.Duration

	mu      sync.RWMutex
	samples []Sample
}

const maxSamples = 8640 // a day at the default 10 s

func NewHeartbeat(client *Client, interval time.Duration) *Heartbeat {
	return &Heartbeat{client: client, interval: interval}
}

func (h *Heartbeat) Run() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()
	for {
		h.poll()
		<-ticker.C
	}
}

func (h *Heartbeat) poll() {
	ctx, cancel := context.WithTimeout(context.Background(), h.interval)
	defer cancel()
	sample := Sample{At: time.Now()}
	var st Status
	if err := h.client.GetJSON(ctx, "/status", &st); err != nil {
		sample.Error = err.Error()
	} else {
		sample.Up = true
		sample.Status = &st
	}
	h.mu.Lock()
	h.samples = append(h.samples, sample)
	if len(h.samples) > maxSamples {
		h.samples = h.samples[len(h.samples)-maxSamples:]
	}
	h.mu.Unlock()
}

func (h *Heartbeat) Latest() (Sample, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.samples) == 0 {
		return Sample{}, false
	}
	return h.samples[len(h.samples)-1], true
}

// Since returns samples newer than t, thinned to at most limit points.
func (h *Heartbeat) Since(t time.Time, limit int) []Sample {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var out []Sample
	for _, s := range h.samples {
		if s.At.After(t) {
			out = append(out, s)
		}
	}
	if limit > 0 && len(out) > limit {
		step := float64(len(out)) / float64(limit)
		thinned := make([]Sample, 0, limit)
		for i := 0; i < limit; i++ {
			thinned = append(thinned, out[int(float64(i)*step)])
		}
		out = thinned
	}
	return out
}
