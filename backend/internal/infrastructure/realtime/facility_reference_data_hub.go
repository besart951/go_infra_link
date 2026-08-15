package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	apprealtime "github.com/besart951/go_infra_link/backend/internal/application/realtime"
	"github.com/google/uuid"
)

const (
	facilityReferenceDataEventChanged = "facility_reference_data.changed"
	facilityCopyJobProgressEvent      = "facility.copy_job.progress"
	facilityReferenceDataWriteWait    = 10 * time.Second
	facilityReferenceDataPongWait     = 60 * time.Second
	facilityReferenceDataPingPeriod   = 25 * time.Second
	facilityReferenceDataMaxMessage   = 4096
	facilityReferenceDataPublishWait  = 2 * time.Second

	FacilityReferenceDataResourceApparats    = "apparats"
	FacilityReferenceDataResourceSystemParts = "system_parts"
)

var facilityReferenceDataSocketConfig = WebSocketConfig{
	WriteWait:       facilityReferenceDataWriteWait,
	PongWait:        facilityReferenceDataPongWait,
	PingPeriod:      facilityReferenceDataPingPeriod,
	MaxMessageBytes: facilityReferenceDataMaxMessage,
}

type FacilityReferenceDataEvent struct {
	Type      string    `json:"type"`
	Resources []string  `json:"resources"`
	At        time.Time `json:"at"`
}

type facilityCopyJobEvent struct {
	Type      string    `json:"type"`
	JobID     uuid.UUID `json:"job_id"`
	Kind      string    `json:"kind"`
	Status    string    `json:"status"`
	Progress  int       `json:"progress"`
	Stage     string    `json:"stage"`
	Error     string    `json:"error,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type facilityReferenceDataBusEvent struct {
	OwnerID uuid.UUID       `json:"owner_id,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

type facilityReferenceDataClient struct {
	hub    *FacilityReferenceDataHub
	userID uuid.UUID
	socket *WebSocketClient
}

type FacilityReferenceDataHub struct {
	mu             sync.RWMutex
	clients        map[*facilityReferenceDataClient]struct{}
	bus            apprealtime.Bus
	nodeID         string
	publishTimeout time.Duration
	ctx            context.Context
	cancel         context.CancelFunc
	closeOnce      sync.Once
}

type FacilityReferenceDataHubOption func(*FacilityReferenceDataHub)

func WithFacilityReferenceDataBus(bus apprealtime.Bus, nodeID string) FacilityReferenceDataHubOption {
	return func(h *FacilityReferenceDataHub) {
		h.bus = bus
		h.nodeID = strings.TrimSpace(nodeID)
	}
}

func NewFacilityReferenceDataHub(options ...FacilityReferenceDataHubOption) *FacilityReferenceDataHub {
	ctx, cancel := context.WithCancel(context.Background())
	hub := &FacilityReferenceDataHub{
		clients:        make(map[*facilityReferenceDataClient]struct{}),
		nodeID:         uuid.NewString(),
		publishTimeout: facilityReferenceDataPublishWait,
		ctx:            ctx,
		cancel:         cancel,
	}
	for _, option := range options {
		if option != nil {
			option(hub)
		}
	}
	if hub.nodeID == "" {
		hub.nodeID = uuid.NewString()
	}
	hub.startBusSubscription()
	return hub
}

func (h *FacilityReferenceDataHub) Close() {
	h.closeOnce.Do(func() {
		h.cancel()
	})
}

func (h *FacilityReferenceDataHub) Stream(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	client := &facilityReferenceDataClient{hub: h, userID: userID}
	socket, err := AcceptWebSocket(w, r, facilityReferenceDataSocketConfig, client.handleMessage, func() {
		h.unregister(client)
	})
	if err != nil {
		return
	}
	client.socket = socket
	h.register(client)
	socket.Run()
}

func (c *facilityReferenceDataClient) handleMessage(data []byte) {
	slog.Debug("ignored unsupported inbound facility reference data websocket message", "bytes", len(data))
}

func (h *FacilityReferenceDataHub) BroadcastFacilityReferenceDataChange(_ context.Context, resources ...string) {
	normalizedResources := normalizeFacilityReferenceDataResources(resources)
	if len(normalizedResources) == 0 {
		return
	}

	h.broadcastDistributed(FacilityReferenceDataEvent{
		Type:      facilityReferenceDataEventChanged,
		Resources: normalizedResources,
		At:        time.Now().UTC(),
	})
}

// BroadcastCopyJobProgress delivers a progress update only to browser
// connections authenticated as the owner. The same event is sent through the
// realtime bus so reconnects routed to another node receive later updates.
func (h *FacilityReferenceDataHub) BroadcastCopyJobProgress(_ context.Context, progress apprealtime.CopyJobProgressEvent) {
	if progress.JobID == uuid.Nil || progress.OwnerID == uuid.Nil {
		return
	}
	payload, err := json.Marshal(facilityCopyJobEvent{
		Type: facilityCopyJobProgressEvent, JobID: progress.JobID, Kind: progress.Kind,
		Status: progress.Status, Progress: progress.Progress, Stage: progress.Stage,
		Error: progress.Error, UpdatedAt: progress.UpdatedAt.UTC(),
	})
	if err != nil {
		return
	}
	h.broadcastCopyJobBytes(progress.OwnerID, payload)
	h.publishCopyJobPayload(progress.OwnerID, payload)
}

func (h *FacilityReferenceDataHub) startBusSubscription() {
	if h.bus == nil {
		return
	}

	events, err := h.bus.Subscribe(h.ctx, apprealtime.TopicFacilityReferenceData)
	if err != nil {
		slog.Warn("facility reference data realtime bus subscription disabled", "err", err)
		return
	}

	go func() {
		for {
			select {
			case <-h.ctx.Done():
				return
			case event, ok := <-events:
				if !ok {
					return
				}
				h.handleBusEvent(event)
			}
		}
	}()
}

func (h *FacilityReferenceDataHub) handleBusEvent(event apprealtime.Event) {
	if event.Source == h.nodeID {
		return
	}

	var busEvent facilityReferenceDataBusEvent
	if err := json.Unmarshal(event.Payload, &busEvent); err != nil {
		slog.Warn("ignored invalid facility reference data bus event", "err", err)
		return
	}

	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(busEvent.Payload, &envelope); err != nil {
		slog.Warn("ignored invalid facility reference data event payload", "err", err)
		return
	}
	switch envelope.Type {
	case facilityReferenceDataEventChanged:
		var payload FacilityReferenceDataEvent
		if err := json.Unmarshal(busEvent.Payload, &payload); err != nil || len(normalizeFacilityReferenceDataResources(payload.Resources)) == 0 {
			slog.Warn("ignored invalid facility reference data event")
			return
		}
		h.broadcastBytes(busEvent.Payload)
	case facilityCopyJobProgressEvent:
		var payload facilityCopyJobEvent
		if err := json.Unmarshal(busEvent.Payload, &payload); err != nil || payload.JobID == uuid.Nil || busEvent.OwnerID == uuid.Nil {
			slog.Warn("ignored invalid facility copy job event")
			return
		}
		// Owner identity is transport metadata and is deliberately not embedded
		// in the browser payload.
		h.broadcastCopyJobBytes(busEvent.OwnerID, busEvent.Payload)
	default:
		slog.Warn("ignored unsupported facility realtime event", "type", envelope.Type)
	}
}

func (h *FacilityReferenceDataHub) register(client *facilityReferenceDataClient) {
	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()
}

func (h *FacilityReferenceDataHub) unregister(client *facilityReferenceDataClient) {
	h.mu.Lock()
	delete(h.clients, client)
	h.mu.Unlock()

	if client.socket != nil {
		client.socket.CloseSend()
	}
}

func (h *FacilityReferenceDataHub) broadcastDistributed(event FacilityReferenceDataEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.broadcastBytes(payload)
	h.publishPayload(payload)
}

func (h *FacilityReferenceDataHub) broadcastBytes(payload []byte) {
	h.mu.RLock()
	clients := make([]*facilityReferenceDataClient, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if client.socket == nil || !client.socket.SendBytes(payload) {
			h.unregister(client)
		}
	}
}

func (h *FacilityReferenceDataHub) broadcastCopyJobBytes(ownerID uuid.UUID, payload []byte) {
	h.mu.RLock()
	clients := make([]*facilityReferenceDataClient, 0, len(h.clients))
	for client := range h.clients {
		if client.userID == ownerID {
			clients = append(clients, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if client.socket == nil || !client.socket.SendBytes(payload) {
			h.unregister(client)
		}
	}
}

func (h *FacilityReferenceDataHub) publishPayload(payload []byte) {
	if h.bus == nil {
		return
	}

	busPayload, err := json.Marshal(facilityReferenceDataBusEvent{
		Payload: append(json.RawMessage(nil), payload...),
	})
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.publishTimeout)
	defer cancel()
	if err := h.bus.Publish(ctx, apprealtime.NewEvent(apprealtime.TopicFacilityReferenceData, h.nodeID, busPayload)); err != nil {
		slog.Warn("facility reference data realtime bus publish failed", "err", err)
	}
}

func (h *FacilityReferenceDataHub) publishCopyJobPayload(ownerID uuid.UUID, payload []byte) {
	if h.bus == nil {
		return
	}

	busPayload, err := json.Marshal(facilityReferenceDataBusEvent{
		OwnerID: ownerID,
		Payload: append(json.RawMessage(nil), payload...),
	})
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.publishTimeout)
	defer cancel()
	if err := h.bus.Publish(ctx, apprealtime.NewEvent(apprealtime.TopicFacilityReferenceData, h.nodeID, busPayload)); err != nil {
		slog.Warn("facility reference data realtime bus publish failed", "err", err)
	}
}

func normalizeFacilityReferenceDataResources(resources []string) []string {
	set := make(map[string]struct{}, len(resources))
	for _, resource := range resources {
		switch resource {
		case FacilityReferenceDataResourceApparats, FacilityReferenceDataResourceSystemParts:
			set[resource] = struct{}{}
		}
	}

	result := make([]string, 0, len(set))
	for _, resource := range []string{
		FacilityReferenceDataResourceApparats,
		FacilityReferenceDataResourceSystemParts,
	} {
		if _, ok := set[resource]; ok {
			result = append(result, resource)
		}
	}
	return result
}
