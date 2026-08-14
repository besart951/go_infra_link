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

type facilityReferenceDataBusEvent struct {
	Payload json.RawMessage `json:"payload"`
}

type facilityReferenceDataClient struct {
	hub    *FacilityReferenceDataHub
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

func (h *FacilityReferenceDataHub) Stream(w http.ResponseWriter, r *http.Request) {
	client := &facilityReferenceDataClient{hub: h}
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

	var payload FacilityReferenceDataEvent
	if err := json.Unmarshal(busEvent.Payload, &payload); err != nil {
		slog.Warn("ignored invalid facility reference data event payload", "err", err)
		return
	}
	if payload.Type != facilityReferenceDataEventChanged || len(normalizeFacilityReferenceDataResources(payload.Resources)) == 0 {
		slog.Warn("ignored invalid facility reference data event")
		return
	}

	h.broadcastBytes(busEvent.Payload)
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
