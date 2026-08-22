package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	apprealtime "github.com/besart951/go_infra_link/backend/internal/application/realtime"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

const (
	projectCollaborationWriteWait   = 10 * time.Second
	projectCollaborationPongWait    = 60 * time.Second
	projectCollaborationPingPeriod  = 25 * time.Second
	projectCollaborationMaxMessage  = 32 * 1024
	projectCollaborationPublishWait = 2 * time.Second

	projectCollaborationMessageSnapshot      = "snapshot"
	projectCollaborationMessagePresence      = "presence"
	projectCollaborationMessageDraftStates   = "draft_states"
	projectCollaborationMessageDraftState    = "draft_state"
	projectCollaborationMessageDraftClear    = "draft_clear"
	projectCollaborationMessageProjectChange = "project_change"
	projectCollaborationMessageRevision      = "revision"

	projectCollaborationBusKindPayload    = "payload"
	projectCollaborationBusKindDraftState = "draft_state"
)

var ErrInvalidProjectChange = errors.New("invalid project change")

var projectCollaborationSocketConfig = WebSocketConfig{
	WriteWait:       projectCollaborationWriteWait,
	PongWait:        projectCollaborationPongWait,
	PingPeriod:      projectCollaborationPingPeriod,
	MaxMessageBytes: projectCollaborationMaxMessage,
}

type ProjectCollaboratorPresence struct {
	UserID      uuid.UUID `json:"user_id"`
	ConnectedAt time.Time `json:"connected_at"`
	LastSeenAt  time.Time `json:"last_seen_at"`
}

type ProjectDraftSelector struct {
	AggregateType string `json:"aggregate_type"`
	AggregateID   string `json:"aggregate_id,omitempty"`
	DraftID       string `json:"draft_id,omitempty"`
}

func (s ProjectDraftSelector) selectorKey() string {
	return s.AggregateType + ":" + s.AggregateID + ":" + s.DraftID
}

type ProjectDraftField struct {
	Path  string `json:"path"`
	Value any    `json:"value"`
}

type ProjectDraftEntry struct {
	ProjectDraftSelector
	Action      string              `json:"action"`
	BaseVersion int64               `json:"base_version"`
	Fields      []ProjectDraftField `json:"fields"`
}

type ProjectDraftState struct {
	UserID    uuid.UUID           `json:"user_id"`
	Entries   []ProjectDraftEntry `json:"entries"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type projectCollaborationSnapshotMessage struct {
	Type            string                        `json:"type"`
	ProjectID       uuid.UUID                     `json:"project_id"`
	CurrentRevision int64                         `json:"current_revision"`
	Presence        []ProjectCollaboratorPresence `json:"presence"`
	DraftStates     []ProjectDraftState           `json:"draft_states"`
	At              time.Time                     `json:"at"`
}

type projectCollaborationPresenceMessage struct {
	Type      string                        `json:"type"`
	ProjectID uuid.UUID                     `json:"project_id"`
	Presence  []ProjectCollaboratorPresence `json:"presence"`
	At        time.Time                     `json:"at"`
}

type projectCollaborationDraftStatesMessage struct {
	Type        string              `json:"type"`
	ProjectID   uuid.UUID           `json:"project_id"`
	DraftStates []ProjectDraftState `json:"draft_states"`
	At          time.Time           `json:"at"`
}

type ProjectChange struct {
	ProjectID     uuid.UUID         `json:"project_id"`
	Revision      int64             `json:"revision"`
	EventID       uuid.UUID         `json:"event_id"`
	AggregateType string            `json:"aggregate_type"`
	AggregateID   uuid.UUID         `json:"aggregate_id"`
	Action        string            `json:"action"`
	ActorID       *uuid.UUID        `json:"actor_id"`
	ChangedFields []string          `json:"changed_fields"`
	ParentRefs    map[string]string `json:"parent_refs"`
	OccurredAt    time.Time         `json:"occurred_at"`
}

// ProjectChangeFromDomain is the single adapter from the durable project
// change model to its realtime representation.
func ProjectChangeFromDomain(change domainProject.Change) (ProjectChange, bool) {
	if change.AggregateID == nil {
		return ProjectChange{}, false
	}
	parentRefs := make(map[string]string, len(change.ParentRefs))
	for key, value := range change.ParentRefs {
		parentRefs[key] = value.String()
	}
	return ProjectChange{
		ProjectID: change.ProjectID, Revision: int64(change.Revision), EventID: change.EventID,
		AggregateType: change.AggregateType, AggregateID: *change.AggregateID,
		Action: string(change.Action), ActorID: change.ActorID, ChangedFields: change.ChangedFields,
		ParentRefs: parentRefs, OccurredAt: change.OccurredAt,
	}, true
}

type projectCollaborationProjectChangeMessage struct {
	Type string `json:"type"`
	ProjectChange
}

type projectCollaborationRevisionMessage struct {
	Type            string    `json:"type"`
	ProjectID       uuid.UUID `json:"project_id"`
	CurrentRevision int64     `json:"current_revision"`
	At              time.Time `json:"at"`
}

type ProjectRevisionSource interface {
	CurrentRevision(ctx context.Context, projectID uuid.UUID) (int64, error)
}

type ProjectDraftStore interface {
	LoadProjectDraftStates(ctx context.Context, projectID uuid.UUID) ([]ProjectDraftState, error)
	SaveProjectDraftState(ctx context.Context, projectID, userID uuid.UUID, entries []ProjectDraftEntry) error
	ClearProjectDraft(ctx context.Context, projectID, userID uuid.UUID, selector *ProjectDraftSelector) error
}

type ProjectPresenceStore interface {
	SaveProjectPresence(ctx context.Context, connectionID, projectID, userID uuid.UUID, connectedAt time.Time) error
	DeleteProjectPresence(ctx context.Context, connectionID uuid.UUID) error
	LoadProjectPresence(ctx context.Context, projectID uuid.UUID) ([]ProjectCollaboratorPresence, error)
}

type projectCollaborationClientMessage struct {
	Type    string
	Entries []ProjectDraftEntry
	Clear   *ProjectDraftSelector
}

type projectCollaborationBusEvent struct {
	Kind      string              `json:"kind"`
	ProjectID uuid.UUID           `json:"project_id"`
	UserID    uuid.UUID           `json:"user_id,omitempty"`
	Entries   []ProjectDraftEntry `json:"entries,omitempty"`
	Payload   json.RawMessage     `json:"payload,omitempty"`
}

type projectCollaborationClient struct {
	hub          *ProjectCollaborationHub
	projectID    uuid.UUID
	userID       uuid.UUID
	connectionID uuid.UUID
	connectedAt  time.Time
	socket       *WebSocketClient
	unregister   sync.Once
}

type projectCollaborationRoom struct {
	clients        map[*projectCollaborationClient]struct{}
	connectionByID map[uuid.UUID]int
	presence       map[uuid.UUID]ProjectCollaboratorPresence
	draftStates    map[uuid.UUID][]ProjectDraftEntry
	revision       int64
}

type ProjectCollaborationHub struct {
	mu             sync.RWMutex
	rooms          map[uuid.UUID]*projectCollaborationRoom
	bus            apprealtime.Bus
	revisions      ProjectRevisionSource
	drafts         ProjectDraftStore
	presenceStore  ProjectPresenceStore
	nodeID         string
	publishTimeout time.Duration
	ctx            context.Context
	cancel         context.CancelFunc
	closeOnce      sync.Once
}

type ProjectCollaborationHubOption func(*ProjectCollaborationHub)

func WithProjectCollaborationBus(bus apprealtime.Bus, nodeID string) ProjectCollaborationHubOption {
	return func(h *ProjectCollaborationHub) {
		h.bus = bus
		h.nodeID = strings.TrimSpace(nodeID)
	}
}

func WithProjectRevisionSource(source ProjectRevisionSource) ProjectCollaborationHubOption {
	return func(h *ProjectCollaborationHub) {
		h.revisions = source
	}
}

func WithProjectDraftStore(store ProjectDraftStore) ProjectCollaborationHubOption {
	return func(h *ProjectCollaborationHub) {
		h.drafts = store
	}
}

func WithProjectPresenceStore(store ProjectPresenceStore) ProjectCollaborationHubOption {
	return func(h *ProjectCollaborationHub) { h.presenceStore = store }
}

func NewProjectCollaborationHub(options ...ProjectCollaborationHubOption) *ProjectCollaborationHub {
	ctx, cancel := context.WithCancel(context.Background())
	h := &ProjectCollaborationHub{
		rooms:          make(map[uuid.UUID]*projectCollaborationRoom),
		nodeID:         uuid.NewString(),
		publishTimeout: projectCollaborationPublishWait,
		ctx:            ctx,
		cancel:         cancel,
	}
	for _, option := range options {
		if option != nil {
			option(h)
		}
	}
	if h.nodeID == "" {
		h.nodeID = uuid.NewString()
	}
	h.startBusSubscription()
	h.startRevisionWatermarks()
	return h
}

func (h *ProjectCollaborationHub) startRevisionWatermarks() {
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-h.ctx.Done():
				return
			case <-ticker.C:
				h.mu.RLock()
				projectIDs := make([]uuid.UUID, 0, len(h.rooms))
				for projectID := range h.rooms {
					projectIDs = append(projectIDs, projectID)
				}
				h.mu.RUnlock()
				for _, projectID := range projectIDs {
					ctx, cancel := context.WithTimeout(h.ctx, h.publishTimeout)
					revision := h.currentRevision(ctx, projectID)
					presence := h.loadPresence(ctx, projectID)
					cancel()
					h.observeRevision(projectID, revision)
					h.broadcast(projectID, projectCollaborationRevisionMessage{Type: projectCollaborationMessageRevision, ProjectID: projectID, CurrentRevision: revision, At: time.Now().UTC()})
					if h.presenceStore != nil {
						h.broadcast(projectID, projectCollaborationPresenceMessage{Type: projectCollaborationMessagePresence, ProjectID: projectID, Presence: presence, At: time.Now().UTC()})
					}
				}
			}
		}
	}()
}

func (h *ProjectCollaborationHub) Close() {
	h.closeOnce.Do(func() {
		h.cancel()
	})
}

func (h *ProjectCollaborationHub) startBusSubscription() {
	if h.bus == nil {
		return
	}

	events, err := h.bus.Subscribe(h.ctx, apprealtime.TopicProjectCollaboration)
	if err != nil {
		slog.Warn("project collaboration realtime bus subscription disabled", "err", err)
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

func (h *ProjectCollaborationHub) handleBusEvent(event apprealtime.Event) {
	if event.Source == h.nodeID {
		return
	}

	var payload projectCollaborationBusEvent
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		slog.Warn("ignored invalid project collaboration bus event", "err", err)
		return
	}

	switch payload.Kind {
	case projectCollaborationBusKindPayload:
		h.observeRemoteRevision(payload.ProjectID, payload.Payload)
		h.broadcastBytes(payload.ProjectID, payload.Payload)
	case projectCollaborationBusKindDraftState:
		h.applyRemoteDraftState(payload.ProjectID, payload.UserID, payload.Entries)
	}
}

func (h *ProjectCollaborationHub) observeRemoteRevision(projectID uuid.UUID, payload []byte) {
	var message struct {
		Type            string `json:"type"`
		Revision        int64  `json:"revision"`
		CurrentRevision int64  `json:"current_revision"`
	}
	if projectID == uuid.Nil || json.Unmarshal(payload, &message) != nil {
		return
	}

	revision := message.Revision
	if message.Type == projectCollaborationMessageRevision {
		revision = message.CurrentRevision
	} else if message.Type != projectCollaborationMessageProjectChange {
		return
	}
	if revision < 0 {
		return
	}

	h.mu.Lock()
	if room := h.rooms[projectID]; room != nil {
		room.revision = max(room.revision, revision)
	}
	h.mu.Unlock()
}

func (h *ProjectCollaborationHub) Register(client *projectCollaborationClient) {
	ctx, cancel := context.WithTimeout(h.ctx, h.publishTimeout)
	defer cancel()
	revision := h.currentRevision(ctx, client.projectID)
	now := time.Now().UTC()
	if client.connectedAt.IsZero() {
		client.connectedAt = now
	}
	h.savePresence(ctx, client)
	storedPresence := h.loadPresence(ctx, client.projectID)
	h.mu.RLock()
	_, roomAlreadyLoaded := h.rooms[client.projectID]
	h.mu.RUnlock()
	var storedDrafts []ProjectDraftState
	if !roomAlreadyLoaded {
		storedDrafts = h.loadDraftStates(ctx, client.projectID)
	}

	h.mu.Lock()
	room := h.ensureRoomLocked(client.projectID)
	room.revision = max(room.revision, revision)
	if !roomAlreadyLoaded {
		for _, state := range storedDrafts {
			if state.UserID != uuid.Nil && len(state.Entries) > 0 {
				room.draftStates[state.UserID] = cloneProjectDraftEntries(state.Entries)
			}
		}
	}
	room.clients[client] = struct{}{}
	room.connectionByID[client.userID] += 1
	if len(storedPresence) > 0 {
		room.presence = make(map[uuid.UUID]ProjectCollaboratorPresence, len(storedPresence))
		for _, item := range storedPresence {
			room.presence[item.UserID] = item
		}
	} else if _, exists := room.presence[client.userID]; !exists {
		room.presence[client.userID] = ProjectCollaboratorPresence{
			UserID:      client.userID,
			ConnectedAt: now,
			LastSeenAt:  now,
		}
	} else {
		presence := room.presence[client.userID]
		presence.LastSeenAt = now
		room.presence[client.userID] = presence
	}
	presence := snapshotPresence(room)
	draftStates := snapshotDraftStates(room)
	currentRevision := room.revision
	h.mu.Unlock()

	h.sendToClient(client, projectCollaborationSnapshotMessage{
		Type:            projectCollaborationMessageSnapshot,
		ProjectID:       client.projectID,
		CurrentRevision: currentRevision,
		Presence:        presence,
		DraftStates:     draftStates,
		At:              now,
	})
	h.broadcastDistributed(client.projectID, projectCollaborationPresenceMessage{
		Type:      projectCollaborationMessagePresence,
		ProjectID: client.projectID,
		Presence:  presence,
		At:        now,
	})
}

func (h *ProjectCollaborationHub) Unregister(client *projectCollaborationClient) {
	if client == nil {
		return
	}
	client.unregister.Do(func() { h.unregister(client) })
}

func (h *ProjectCollaborationHub) unregister(client *projectCollaborationClient) {
	var (
		presence          []ProjectCollaboratorPresence
		draftStates       []ProjectDraftState
		now               = time.Now().UTC()
		shouldSend        bool
		shouldClearDrafts bool
	)

	ctx, cancel := context.WithTimeout(context.Background(), h.publishTimeout)
	defer cancel()
	h.deletePresence(ctx, client.connectionID)
	storedPresence := h.loadPresence(ctx, client.projectID)
	userPresentOnAnotherNode := false
	for _, item := range storedPresence {
		if item.UserID == client.userID {
			userPresentOnAnotherNode = true
			break
		}
	}
	h.mu.Lock()
	room, ok := h.rooms[client.projectID]
	if ok {
		delete(room.clients, client)
		if len(storedPresence) > 0 {
			room.presence = make(map[uuid.UUID]ProjectCollaboratorPresence, len(storedPresence))
			for _, item := range storedPresence {
				room.presence[item.UserID] = item
			}
		}
		if room.connectionByID[client.userID] > 1 {
			room.connectionByID[client.userID] -= 1
			presenceState := room.presence[client.userID]
			presenceState.LastSeenAt = now
			room.presence[client.userID] = presenceState
		} else {
			delete(room.connectionByID, client.userID)
			if !userPresentOnAnotherNode {
				delete(room.presence, client.userID)
				_, shouldClearDrafts = room.draftStates[client.userID]
				delete(room.draftStates, client.userID)
			}
		}

		if len(room.clients) == 0 {
			delete(h.rooms, client.projectID)
		} else {
			presence = snapshotPresence(room)
			draftStates = snapshotDraftStates(room)
			shouldSend = true
		}
	}
	h.mu.Unlock()

	if shouldSend {
		h.broadcastDistributed(client.projectID, projectCollaborationPresenceMessage{
			Type:      projectCollaborationMessagePresence,
			ProjectID: client.projectID,
			Presence:  presence,
			At:        now,
		})
		h.broadcast(client.projectID, projectCollaborationDraftStatesMessage{
			Type:        projectCollaborationMessageDraftStates,
			ProjectID:   client.projectID,
			DraftStates: draftStates,
			At:          now,
		})
	} else if h.presenceStore != nil {
		h.broadcastDistributed(client.projectID, projectCollaborationPresenceMessage{
			Type: projectCollaborationMessagePresence, ProjectID: client.projectID,
			Presence: storedPresence, At: now,
		})
	}
	if shouldClearDrafts {
		h.publishDraftState(client.projectID, client.userID, nil)
		h.clearStoredDraft(client.projectID, client.userID, nil)
	}

	if client.socket != nil {
		client.socket.CloseSend()
	}
}

func (h *ProjectCollaborationHub) UpdateDraftState(projectID, userID uuid.UUID, entries []ProjectDraftEntry) {
	if projectID == uuid.Nil || userID == uuid.Nil || len(entries) == 0 {
		return
	}
	now := time.Now().UTC()
	normalized := cloneProjectDraftEntries(entries)

	h.mu.Lock()
	room, ok := h.rooms[projectID]
	if !ok {
		h.mu.Unlock()
		return
	}
	room.draftStates[userID] = normalized
	draftStates := snapshotDraftStates(room)
	h.mu.Unlock()

	h.broadcast(projectID, projectCollaborationDraftStatesMessage{
		Type:        projectCollaborationMessageDraftStates,
		ProjectID:   projectID,
		DraftStates: draftStates,
		At:          now,
	})
	h.publishDraftState(projectID, userID, normalized)
	h.saveStoredDraft(projectID, userID, normalized)
}

func (h *ProjectCollaborationHub) ClearDraft(projectID, userID uuid.UUID, selector ProjectDraftSelector) {
	if projectID == uuid.Nil || userID == uuid.Nil {
		return
	}
	now := time.Now().UTC()

	h.mu.Lock()
	room, ok := h.rooms[projectID]
	if !ok {
		h.mu.Unlock()
		return
	}
	entries := room.draftStates[userID]
	kept := entries[:0]
	for _, entry := range entries {
		if entry.ProjectDraftSelector.selectorKey() != selector.selectorKey() {
			kept = append(kept, entry)
		}
	}
	if len(kept) == 0 {
		delete(room.draftStates, userID)
	} else {
		room.draftStates[userID] = cloneProjectDraftEntries(kept)
	}
	current := cloneProjectDraftEntries(room.draftStates[userID])
	draftStates := snapshotDraftStates(room)
	h.mu.Unlock()

	h.broadcast(projectID, projectCollaborationDraftStatesMessage{
		Type:        projectCollaborationMessageDraftStates,
		ProjectID:   projectID,
		DraftStates: draftStates,
		At:          now,
	})
	h.publishDraftState(projectID, userID, current)
	h.clearStoredDraft(projectID, userID, &selector)
}

func (h *ProjectCollaborationHub) BroadcastProjectChange(ctx context.Context, change ProjectChange) error {
	if err := validateServerProjectChange(change); err != nil {
		return err
	}
	change.ChangedFields = normalizeIDs(change.ChangedFields)
	if change.ChangedFields == nil {
		change.ChangedFields = []string{}
	}
	if change.ParentRefs == nil {
		change.ParentRefs = map[string]string{}
	}
	if change.OccurredAt.IsZero() {
		change.OccurredAt = time.Now().UTC()
	}

	h.observeRevision(change.ProjectID, change.Revision)

	h.broadcastDistributed(change.ProjectID, projectCollaborationProjectChangeMessage{
		Type:          projectCollaborationMessageProjectChange,
		ProjectChange: change,
	})
	h.BroadcastRevision(change.ProjectID, change.Revision)
	return nil
}

func (h *ProjectCollaborationHub) BroadcastRevision(projectID uuid.UUID, revision int64) {
	if projectID == uuid.Nil || revision < 0 {
		return
	}
	now := time.Now().UTC()
	current := h.observeRevision(projectID, revision)
	h.broadcastDistributed(projectID, projectCollaborationRevisionMessage{
		Type:            projectCollaborationMessageRevision,
		ProjectID:       projectID,
		CurrentRevision: current,
		At:              now,
	})
}

func (h *ProjectCollaborationHub) observeRevision(projectID uuid.UUID, revision int64) int64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.rooms[projectID]
	if room == nil {
		return revision
	}
	room.revision = max(room.revision, revision)
	return room.revision
}

func validateServerProjectChange(change ProjectChange) error {
	if change.ProjectID == uuid.Nil || change.Revision <= 0 || change.EventID == uuid.Nil ||
		strings.TrimSpace(change.AggregateType) == "" || change.AggregateID == uuid.Nil || strings.TrimSpace(change.Action) == "" {
		return ErrInvalidProjectChange
	}
	return nil
}

// Deprecated compatibility methods translate existing server-side notifications
// into v2 messages. Browser-authored deltas and refresh requests are rejected.
func (h *ProjectCollaborationHub) BroadcastRefreshRequest(projectID uuid.UUID, _ *uuid.UUID, _ string, _ []string) {
	h.BroadcastRevision(projectID, h.nextLocalRevision(projectID))
}

func (h *ProjectCollaborationHub) BroadcastControlCabinetDelta(projectID uuid.UUID, actorID *uuid.UUID, cabinet domainFacility.ControlCabinet) {
	h.broadcastLegacyProjectChange(projectID, actorID, "control_cabinet", cabinet.ID, "updated", []string{"building_id", "control_cabinet_nr"})
}

func (h *ProjectCollaborationHub) BroadcastSPSControllerDelta(projectID uuid.UUID, actorID *uuid.UUID, controller domainFacility.SPSController) {
	h.broadcastLegacyProjectChange(projectID, actorID, "sps_controller", controller.ID, "updated", []string{"control_cabinet_id", "ga_device", "device_name", "device_description", "device_location", "ip_address", "subnet", "gateway", "vlan"})
}

func (h *ProjectCollaborationHub) BroadcastFieldDeviceDelta(projectID uuid.UUID, actorID *uuid.UUID, devices []map[string]any) {
	for _, device := range devices {
		id, _ := device["id"].(string)
		entityID, err := uuid.Parse(id)
		if err != nil || entityID == uuid.Nil {
			continue
		}
		fields := make([]string, 0, len(device)-1)
		for field := range device {
			if field != "id" {
				fields = append(fields, field)
			}
		}
		h.broadcastLegacyProjectChange(projectID, actorID, "field_device", entityID, "updated", fields)
	}
}

func (h *ProjectCollaborationHub) broadcastLegacyProjectChange(projectID uuid.UUID, actorID *uuid.UUID, aggregateType string, aggregateID uuid.UUID, action string, fields []string) {
	_ = h.BroadcastProjectChange(context.Background(), ProjectChange{
		ProjectID: projectID, Revision: h.nextLocalRevision(projectID), EventID: uuid.New(), AggregateType: aggregateType,
		AggregateID: aggregateID, Action: action, ActorID: actorID, ChangedFields: fields, ParentRefs: map[string]string{}, OccurredAt: time.Now().UTC(),
	})
}

func (h *ProjectCollaborationHub) ensureRoomLocked(projectID uuid.UUID) *projectCollaborationRoom {
	room, ok := h.rooms[projectID]
	if !ok {
		room = &projectCollaborationRoom{
			clients:        make(map[*projectCollaborationClient]struct{}),
			connectionByID: make(map[uuid.UUID]int),
			presence:       make(map[uuid.UUID]ProjectCollaboratorPresence),
			draftStates:    make(map[uuid.UUID][]ProjectDraftEntry),
		}
		h.rooms[projectID] = room
	}
	return room
}

func (h *ProjectCollaborationHub) broadcast(projectID uuid.UUID, payload any) {
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.broadcastBytes(projectID, b)
}

func (h *ProjectCollaborationHub) broadcastDistributed(projectID uuid.UUID, payload any) {
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.broadcastBytes(projectID, b)
	h.publishPayload(projectID, b)
}

func (h *ProjectCollaborationHub) broadcastBytes(projectID uuid.UUID, b []byte) {
	h.mu.RLock()
	room, ok := h.rooms[projectID]
	if !ok {
		h.mu.RUnlock()
		return
	}
	clients := make([]*projectCollaborationClient, 0, len(room.clients))
	for client := range room.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if client.socket == nil || !client.socket.SendBytes(b) {
			h.Unregister(client)
		}
	}
}

func (h *ProjectCollaborationHub) sendToClient(client *projectCollaborationClient, payload any) {
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	if client.socket == nil || !client.socket.SendBytes(b) {
		h.Unregister(client)
	}
}

func (h *ProjectCollaborationHub) publishPayload(projectID uuid.UUID, payload []byte) {
	h.publishBusEvent(projectCollaborationBusEvent{
		Kind:      projectCollaborationBusKindPayload,
		ProjectID: projectID,
		Payload:   append(json.RawMessage(nil), payload...),
	})
}

func (h *ProjectCollaborationHub) publishDraftState(projectID, userID uuid.UUID, entries []ProjectDraftEntry) {
	h.publishBusEvent(projectCollaborationBusEvent{
		Kind:      projectCollaborationBusKindDraftState,
		ProjectID: projectID,
		UserID:    userID,
		Entries:   cloneProjectDraftEntries(entries),
	})
}

func (h *ProjectCollaborationHub) publishBusEvent(payload projectCollaborationBusEvent) {
	if h.bus == nil {
		return
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), h.publishTimeout)
	defer cancel()
	if err := h.bus.Publish(ctx, apprealtime.NewEvent(apprealtime.TopicProjectCollaboration, h.nodeID, b)); err != nil {
		slog.Warn("project collaboration realtime bus publish failed", "err", err)
	}
}

func (h *ProjectCollaborationHub) applyRemoteDraftState(projectID, userID uuid.UUID, entries []ProjectDraftEntry) {
	if projectID == uuid.Nil || userID == uuid.Nil {
		return
	}

	now := time.Now().UTC()
	h.mu.Lock()
	room, ok := h.rooms[projectID]
	if !ok {
		h.mu.Unlock()
		return
	}
	if len(entries) == 0 {
		delete(room.draftStates, userID)
	} else {
		room.draftStates[userID] = cloneProjectDraftEntries(entries)
	}
	draftStates := snapshotDraftStates(room)
	h.mu.Unlock()

	h.broadcast(projectID, projectCollaborationDraftStatesMessage{
		Type:        projectCollaborationMessageDraftStates,
		ProjectID:   projectID,
		DraftStates: draftStates,
		At:          now,
	})
}

func snapshotPresence(room *projectCollaborationRoom) []ProjectCollaboratorPresence {
	items := make([]ProjectCollaboratorPresence, 0, len(room.presence))
	for _, item := range room.presence {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].ConnectedAt.Equal(items[j].ConnectedAt) {
			return items[i].UserID.String() < items[j].UserID.String()
		}
		return items[i].ConnectedAt.Before(items[j].ConnectedAt)
	})
	return items
}

func snapshotDraftStates(room *projectCollaborationRoom) []ProjectDraftState {
	items := make([]ProjectDraftState, 0, len(room.draftStates))
	now := time.Now().UTC()
	type draftStateWithTime struct {
		userID uuid.UUID
		state  ProjectDraftState
	}
	var itemsWithTime []draftStateWithTime

	for userID, entries := range room.draftStates {
		itemsWithTime = append(itemsWithTime, draftStateWithTime{
			userID: userID,
			state: ProjectDraftState{
				UserID:    userID,
				Entries:   cloneProjectDraftEntries(entries),
				UpdatedAt: now,
			},
		})
	}

	sort.Slice(itemsWithTime, func(i, j int) bool {
		return itemsWithTime[i].userID.String() < itemsWithTime[j].userID.String()
	})

	for _, item := range itemsWithTime {
		items = append(items, item.state)
	}

	return items
}

func normalizeIDs(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ids))
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	sort.Strings(result)
	return result
}

func cloneProjectDraftEntries(entries []ProjectDraftEntry) []ProjectDraftEntry {
	if len(entries) == 0 {
		return nil
	}
	cloned := make([]ProjectDraftEntry, len(entries))
	for i, entry := range entries {
		cloned[i] = entry
		cloned[i].Fields = make([]ProjectDraftField, len(entry.Fields))
		for j, field := range entry.Fields {
			cloned[i].Fields[j] = ProjectDraftField{Path: field.Path, Value: cloneDraftValue(field.Value)}
		}
	}
	return cloned
}

func cloneDraftValue(value any) any {
	switch typed := value.(type) {
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = cloneDraftValue(item)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = cloneDraftValue(item)
		}
		return out
	default:
		return value
	}
}

func (h *ProjectCollaborationHub) currentRevision(ctx context.Context, projectID uuid.UUID) int64 {
	if h.revisions != nil {
		revision, err := h.revisions.CurrentRevision(ctx, projectID)
		if err == nil && revision >= 0 {
			return revision
		}
		if err != nil {
			slog.Warn("project collaboration revision lookup failed", "project_id", projectID, "err", err)
		}
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if room := h.rooms[projectID]; room != nil {
		return room.revision
	}
	return 0
}

func (h *ProjectCollaborationHub) nextLocalRevision(projectID uuid.UUID) int64 {
	current := int64(0)
	if h.revisions != nil {
		ctx, cancel := context.WithTimeout(context.Background(), h.publishTimeout)
		current = h.currentRevision(ctx, projectID)
		cancel()
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.ensureRoomLocked(projectID)
	room.revision = max(room.revision, current)
	room.revision++
	return room.revision
}

func (h *ProjectCollaborationHub) loadDraftStates(ctx context.Context, projectID uuid.UUID) []ProjectDraftState {
	if h.drafts == nil {
		return nil
	}
	states, err := h.drafts.LoadProjectDraftStates(ctx, projectID)
	if err != nil {
		slog.Warn("project collaboration draft lookup failed", "project_id", projectID, "err", err)
		return nil
	}
	return states
}

func (h *ProjectCollaborationHub) saveStoredDraft(projectID, userID uuid.UUID, entries []ProjectDraftEntry) {
	if h.drafts == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), h.publishTimeout)
	defer cancel()
	if err := h.drafts.SaveProjectDraftState(ctx, projectID, userID, cloneProjectDraftEntries(entries)); err != nil {
		slog.Warn("project collaboration draft save failed", "project_id", projectID, "user_id", userID, "err", err)
	}
}

func (h *ProjectCollaborationHub) clearStoredDraft(projectID, userID uuid.UUID, selector *ProjectDraftSelector) {
	if h.drafts == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), h.publishTimeout)
	defer cancel()
	if err := h.drafts.ClearProjectDraft(ctx, projectID, userID, selector); err != nil {
		slog.Warn("project collaboration draft clear failed", "project_id", projectID, "user_id", userID, "err", err)
	}
}

func (h *ProjectCollaborationHub) savePresence(ctx context.Context, client *projectCollaborationClient) {
	if h.presenceStore == nil {
		return
	}
	if err := h.presenceStore.SaveProjectPresence(ctx, client.connectionID, client.projectID, client.userID, client.connectedAt); err != nil {
		slog.Warn("project collaboration presence save failed", "project_id", client.projectID, "err", err)
	}
}

func (h *ProjectCollaborationHub) deletePresence(ctx context.Context, connectionID uuid.UUID) {
	if h.presenceStore == nil {
		return
	}
	if err := h.presenceStore.DeleteProjectPresence(ctx, connectionID); err != nil {
		slog.Warn("project collaboration presence delete failed", "connection_id", connectionID, "err", err)
	}
}

func (h *ProjectCollaborationHub) loadPresence(ctx context.Context, projectID uuid.UUID) []ProjectCollaboratorPresence {
	if h.presenceStore == nil {
		return nil
	}
	items, err := h.presenceStore.LoadProjectPresence(ctx, projectID)
	if err != nil {
		slog.Warn("project collaboration presence lookup failed", "project_id", projectID, "err", err)
		return nil
	}
	return items
}

func (h *ProjectCollaborationHub) renewPresence(ctx context.Context, client *projectCollaborationClient) {
	if h.presenceStore == nil {
		return
	}
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			callCtx, cancel := context.WithTimeout(ctx, h.publishTimeout)
			h.savePresence(callCtx, client)
			h.mu.RLock()
			room := h.rooms[client.projectID]
			var entries []ProjectDraftEntry
			if room != nil {
				entries = cloneProjectDraftEntries(room.draftStates[client.userID])
			}
			h.mu.RUnlock()
			if len(entries) > 0 && h.drafts != nil {
				if err := h.drafts.SaveProjectDraftState(callCtx, client.projectID, client.userID, entries); err != nil {
					slog.Warn("project collaboration draft renewal failed", "project_id", client.projectID, "user_id", client.userID, "err", err)
				}
			}
			cancel()
		}
	}
}

func (c *projectCollaborationClient) handleMessage(data []byte) {
	message, err := parseProjectCollaborationClientMessage(data)
	if err != nil {
		logInvalidProjectCollaborationMessage(data, err)
		return
	}

	switch message.Type {
	case projectCollaborationMessageDraftState:
		c.hub.UpdateDraftState(c.projectID, c.userID, message.Entries)
	case projectCollaborationMessageDraftClear:
		if message.Clear != nil {
			c.hub.ClearDraft(c.projectID, c.userID, *message.Clear)
		}
	}
}

func (h *ProjectCollaborationHub) Stream(w http.ResponseWriter, r *http.Request, projectID, userID uuid.UUID) error {
	client := &projectCollaborationClient{
		hub:          h,
		projectID:    projectID,
		userID:       userID,
		connectionID: uuid.New(),
		connectedAt:  time.Now().UTC(),
	}
	socket, err := AcceptWebSocket(w, r, projectCollaborationSocketConfig, client.handleMessage, func() {
		h.Unregister(client)
	})
	if err != nil {
		return err
	}
	client.socket = socket

	h.Register(client)
	leaseCtx, cancelLease := context.WithCancel(r.Context())
	defer cancelLease()
	go h.renewPresence(leaseCtx, client)
	socket.Run()
	return nil
}
