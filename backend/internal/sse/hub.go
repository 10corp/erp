package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Event struct {
	JobID   string      `json:"job_id,omitempty"`
	Step    string      `json:"step"`    // 'analyzing_document', 'extracting_line_items', 'validating_math', 'completed', 'failed'
	Message string      `json:"message"` // Human-readable status description
	Payload interface{} `json:"payload,omitempty"`
}

type Client chan Event

type Hub struct {
	mu         sync.RWMutex
	jobClients map[string]map[Client]bool
}

var globalHub = &Hub{
	jobClients: make(map[string]map[Client]bool),
}

func GetHub() *Hub {
	return globalHub
}

func (h *Hub) Subscribe(jobID string) Client {
	h.mu.Lock()
	defer h.mu.Unlock()

	client := make(Client, 16)
	if _, exists := h.jobClients[jobID]; !exists {
		h.jobClients[jobID] = make(map[Client]bool)
	}
	h.jobClients[jobID][client] = true
	return client
}

func (h *Hub) Unsubscribe(jobID string, client Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, exists := h.jobClients[jobID]; exists {
		delete(clients, client)
		close(client)
		if len(clients) == 0 {
			delete(h.jobClients, jobID)
		}
	}
}

func (h *Hub) Broadcast(jobID string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	event.JobID = jobID
	if clients, exists := h.jobClients[jobID]; exists {
		for client := range clients {
			select {
			case client <- event:
			default:
				// non-blocking
			}
		}
	}
}

// ServeHTTPStream handles the HTTP SSE streaming connection for a specific job_id
func (h *Hub) ServeHTTPStream(w http.ResponseWriter, r *http.Request, jobID string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	client := h.Subscribe(jobID)
	defer h.Unsubscribe(jobID, client)

	// Send initial connected event
	fmt.Fprintf(w, "event: connected\ndata: {\"job_id\":\"%s\",\"message\":\"Subscribed to live IDP processing stream\"}\n\n", jobID)
	flusher.Flush()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case event, ok := <-client:
			if !ok {
				return
			}
			bytes, _ := json.Marshal(event)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Step, string(bytes))
			flusher.Flush()
			if event.Step == "completed" || event.Step == "failed" {
				return
			}
		}
	}
}
