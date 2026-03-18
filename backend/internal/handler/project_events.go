package handler

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	projectEventName = "project.change"
)

type ProjectEvent struct {
	Type      string    `json:"type"`
	ProjectID uuid.UUID `json:"project_id"`
	At        time.Time `json:"at"`
	ActorID   string    `json:"actor_id,omitempty"`
}

type ProjectEventHub struct {
	mu          sync.RWMutex
	subscribers map[uuid.UUID]map[chan ProjectEvent]struct{}
}

func NewProjectEventHub() *ProjectEventHub {
	return &ProjectEventHub{
		subscribers: make(map[uuid.UUID]map[chan ProjectEvent]struct{}),
	}
}

func (h *ProjectEventHub) Subscribe(projectID uuid.UUID) (<-chan ProjectEvent, func()) {
	ch := make(chan ProjectEvent, 16)

	h.mu.Lock()
	if _, ok := h.subscribers[projectID]; !ok {
		h.subscribers[projectID] = make(map[chan ProjectEvent]struct{})
	}
	h.subscribers[projectID][ch] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		projectSubs, ok := h.subscribers[projectID]
		if !ok {
			return
		}

		if _, exists := projectSubs[ch]; exists {
			delete(projectSubs, ch)
			close(ch)
		}

		if len(projectSubs) == 0 {
			delete(h.subscribers, projectID)
		}
	}

	return ch, unsubscribe
}

func (h *ProjectEventHub) Publish(projectID uuid.UUID, eventType string, actorID *uuid.UUID) {
	h.mu.RLock()
	projectSubs, ok := h.subscribers[projectID]
	if !ok {
		h.mu.RUnlock()
		return
	}

	subscribers := make([]chan ProjectEvent, 0, len(projectSubs))
	for ch := range projectSubs {
		subscribers = append(subscribers, ch)
	}
	h.mu.RUnlock()

	actor := ""
	if actorID != nil {
		actor = actorID.String()
	}

	event := ProjectEvent{
		Type:      eventType,
		ProjectID: projectID,
		At:        time.Now().UTC(),
		ActorID:   actor,
	}

	for _, ch := range subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

func formatSSE(eventName string, payload any) (string, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, string(b)), nil
}
