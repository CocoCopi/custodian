// Package ws implements the WebSocket hub used to stream live build and
// deployment logs to the dashboard and CLI.
package ws

import (
	"sync"

	"github.com/CocoCopi/custodian/internal/models"
)

// Hub fans log lines out to subscribers keyed by service and deployment id.
type Hub struct {
	mu   sync.RWMutex
	subs map[string]map[chan models.LogEntry]struct{}
}

// NewHub returns an empty hub.
func NewHub() *Hub {
	return &Hub{subs: map[string]map[chan models.LogEntry]struct{}{}}
}

// Subscribe registers a channel for a topic (service id or deployment id).
// The returned unsubscribe func must be called when the client disconnects.
func (h *Hub) Subscribe(topic string) (chan models.LogEntry, func()) {
	ch := make(chan models.LogEntry, 256)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subs[topic] == nil {
		h.subs[topic] = map[chan models.LogEntry]struct{}{}
	}
	h.subs[topic][ch] = struct{}{}
	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if m, ok := h.subs[topic]; ok {
			delete(m, ch)
			if len(m) == 0 {
				delete(h.subs, topic)
			}
		}
		close(ch)
	}
}

// Publish delivers a log entry to every subscriber of both its service and
// deployment topics. Slow subscribers are dropped rather than blocking.
func (h *Hub) Publish(e models.LogEntry) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, topic := range []string{e.ServiceID, e.DeploymentID} {
		for ch := range h.subs[topic] {
			select {
			case ch <- e:
			default: // drop for slow consumers
			}
		}
	}
}

// Log implements jobs.Logger, bridging the worker to the hub.
func (h *Hub) Log(serviceID, deploymentID, stream, message string) {
	h.Publish(models.LogEntry{
		ServiceID:    serviceID,
		DeploymentID: deploymentID,
		Stream:       stream,
		Message:      message,
	})
}
