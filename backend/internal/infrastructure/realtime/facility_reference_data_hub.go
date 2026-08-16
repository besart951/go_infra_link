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
	facilityChangedEvent              = "facility.changed"
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

// FacilityChangeEvent is the shared facility-stream invalidation contract.
// Resource names are intentionally plural so one event can invalidate a list,
// a detail view and any matching client-side cache.
type FacilityChangeEvent struct {
	Type     string      `json:"type"`
	Resource string      `json:"resource"`
	Action   string      `json:"action"`
	IDs      []uuid.UUID `json:"ids"`
	ActorID  *uuid.UUID  `json:"actor_id,omitempty"`
	At       time.Time   `json:"at"`
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
	hub               *FacilityReferenceDataHub
	userID            uuid.UUID
	readableResources map[string]struct{}
	socket            *WebSocketClient
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

func (h *FacilityReferenceDataHub) Stream(w http.ResponseWriter, r *http.Request, userID uuid.UUID, readableResources map[string]struct{}) {
	client := &facilityReferenceDataClient{hub: h, userID: userID, readableResources: cloneFacilityReadableResources(readableResources)}
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

	event := FacilityReferenceDataEvent{
		Type:      facilityReferenceDataEventChanged,
		Resources: normalizedResources,
		At:        time.Now().UTC(),
	}
	h.broadcastReferenceDataEvent(event)
	h.publishEvent(event)
}

func (h *FacilityReferenceDataHub) BroadcastFacilityChange(_ context.Context, resource, action string, ids []uuid.UUID, actorID *uuid.UUID) {
	if !isFacilityRealtimeResource(resource) || !isFacilityRealtimeAction(action) {
		return
	}
	event := FacilityChangeEvent{
		Type: facilityChangedEvent, Resource: resource, Action: action,
		IDs: uniqueUUIDs(ids), ActorID: actorID, At: time.Now().UTC(),
	}
	h.broadcastFacilityChangeEvent(event)
	h.publishEvent(event)
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
		h.broadcastReferenceDataEvent(payload)
	case facilityChangedEvent:
		var payload FacilityChangeEvent
		if err := json.Unmarshal(busEvent.Payload, &payload); err != nil || !isFacilityRealtimeResource(payload.Resource) || !isFacilityRealtimeAction(payload.Action) {
			slog.Warn("ignored invalid facility change event")
			return
		}
		h.broadcastFacilityChangeEvent(payload)
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

func (h *FacilityReferenceDataHub) broadcastReferenceDataEvent(event FacilityReferenceDataEvent) {
	resources := normalizeFacilityReferenceDataResources(event.Resources)
	if len(resources) == 0 {
		return
	}
	event.Resources = resources
	h.forEachClient(func(client *facilityReferenceDataClient) {
		allowed := make([]string, 0, len(resources))
		for _, resource := range resources {
			if client.canRead(resource) {
				allowed = append(allowed, resource)
			}
		}
		if len(allowed) == 0 {
			return
		}
		eventForClient := event
		eventForClient.Resources = allowed
		h.sendEvent(client, eventForClient)
	})
}

func (h *FacilityReferenceDataHub) broadcastFacilityChangeEvent(event FacilityChangeEvent) {
	h.forEachClient(func(client *facilityReferenceDataClient) {
		if client.canRead(event.Resource) {
			h.sendEvent(client, event)
		}
	})
}

func (h *FacilityReferenceDataHub) sendEvent(client *facilityReferenceDataClient, event any) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	if client.socket == nil || !client.socket.SendBytes(payload) {
		h.unregister(client)
	}
}

func (h *FacilityReferenceDataHub) forEachClient(fn func(*facilityReferenceDataClient)) {
	h.mu.RLock()
	clients := make([]*facilityReferenceDataClient, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		fn(client)
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

func (h *FacilityReferenceDataHub) publishEvent(event any) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.publishPayload(payload)
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

func (c *facilityReferenceDataClient) canRead(resource string) bool {
	if c.readableResources == nil {
		return true
	}
	_, ok := c.readableResources[resource]
	return ok
}

func cloneFacilityReadableResources(resources map[string]struct{}) map[string]struct{} {
	if resources == nil {
		return nil
	}
	copy := make(map[string]struct{}, len(resources))
	for resource := range resources {
		if isFacilityRealtimeResource(resource) {
			copy[resource] = struct{}{}
		}
	}
	return copy
}

func uniqueUUIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return []uuid.UUID{}
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func isFacilityRealtimeResource(resource string) bool {
	switch resource {
	case "buildings", "system_types", "system_parts", "apparats", "control_cabinets", "sps_controllers", "sps_controller_system_types", "field_devices", "bacnet_objects", "object_data", "state_texts", "notification_classes", "alarm_definitions", "alarm_types", "alarm_type_fields", "alarm_fields", "units":
		return true
	default:
		return false
	}
}

func isFacilityRealtimeAction(action string) bool {
	switch action {
	case "created", "updated", "deleted", "copied", "bulk_created", "bulk_updated", "bulk_deleted":
		return true
	default:
		return false
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
