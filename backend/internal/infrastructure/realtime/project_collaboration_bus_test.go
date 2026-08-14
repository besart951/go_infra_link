package realtime

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProjectCollaborationProjectChangeRoutesAcrossInstances(t *testing.T) {
	bus := NewInMemoryBus()
	defer bus.Close()
	hubA := NewProjectCollaborationHub(WithProjectCollaborationBus(bus, "node-a"))
	defer hubA.Close()
	hubB := NewProjectCollaborationHub(WithProjectCollaborationBus(bus, "node-b"))
	defer hubB.Close()

	projectID := uuid.New()
	client := registerProjectTestClient(hubB, projectID, uuid.New(), 12)
	drainSocket(client.socket)
	change := ProjectChange{
		ProjectID: projectID, Revision: 42, EventID: uuid.New(), AggregateType: "field_device",
		AggregateID: uuid.New(), Action: "updated", ActorID: ptrUUID(uuid.New()),
		ChangedFields: []string{"bmk", "description"}, ParentRefs: map[string]string{"control_cabinet_id": uuid.NewString()},
		OccurredAt: time.Now().UTC(),
	}
	if err := hubA.BroadcastProjectChange(t.Context(), change); err != nil {
		t.Fatalf("broadcast project change: %v", err)
	}

	message := receiveSocketMessageOfType(t, client.socket, projectCollaborationMessageProjectChange)
	if len(message) != 11 {
		t.Fatalf("project_change keys = %#v", message)
	}
	if message["revision"] != float64(42) || message["event_id"] != change.EventID.String() || message["aggregate_id"] != change.AggregateID.String() {
		t.Fatalf("project_change = %#v", message)
	}
	revision := receiveSocketMessageOfType(t, client.socket, projectCollaborationMessageRevision)
	if len(revision) != 4 || revision["current_revision"] != float64(42) {
		t.Fatalf("revision = %#v", revision)
	}

	joiningClient := registerProjectTestClient(hubB, projectID, uuid.New(), 8)
	snapshot := receiveSocketMessageOfType(t, joiningClient.socket, projectCollaborationMessageSnapshot)
	if snapshot["current_revision"] != float64(42) {
		t.Fatalf("snapshot current_revision = %#v", snapshot)
	}
}

func TestProjectCollaborationDraftStateRoutesAcrossInstances(t *testing.T) {
	bus := NewInMemoryBus()
	defer bus.Close()
	hubA := NewProjectCollaborationHub(WithProjectCollaborationBus(bus, "node-a"))
	defer hubA.Close()
	hubB := NewProjectCollaborationHub(WithProjectCollaborationBus(bus, "node-b"))
	defer hubB.Close()

	projectID := uuid.New()
	editorID := uuid.New()
	viewer := registerProjectTestClient(hubB, projectID, uuid.New(), 8)
	editor := registerProjectTestClient(hubA, projectID, editorID, 8)
	drainSocket(viewer.socket)
	drainSocket(editor.socket)

	hubA.UpdateDraftState(projectID, editorID, []ProjectDraftEntry{{
		ProjectDraftSelector: ProjectDraftSelector{AggregateType: "field_device", AggregateID: uuid.NewString()},
		Action:               "update", BaseVersion: 3, Fields: []ProjectDraftField{{Path: "bmk", Value: "draft"}},
	}})

	message := receiveSocketMessageOfType(t, viewer.socket, projectCollaborationMessageDraftStates)
	states, ok := message["draft_states"].([]any)
	if !ok || len(states) != 1 || states[0].(map[string]any)["user_id"] != editorID.String() {
		t.Fatalf("draft_states = %#v", message["draft_states"])
	}
}

func TestProjectCollaborationSnapshotIncludesDurableRevision(t *testing.T) {
	projectID := uuid.New()
	hub := NewProjectCollaborationHub(WithProjectRevisionSource(staticProjectRevision{revision: 77}))
	defer hub.Close()
	client := registerProjectTestClient(hub, projectID, uuid.New(), 8)

	message := receiveSocketMessageOfType(t, client.socket, projectCollaborationMessageSnapshot)
	if message["current_revision"] != float64(77) {
		t.Fatalf("snapshot = %#v", message)
	}
}

func TestProjectCollaborationDraftStoreReceivesSaveAndClear(t *testing.T) {
	store := &recordingProjectDraftStore{}
	hub := NewProjectCollaborationHub(WithProjectDraftStore(store))
	defer hub.Close()
	projectID, userID := uuid.New(), uuid.New()
	client := registerProjectTestClient(hub, projectID, userID, 8)
	drainSocket(client.socket)
	selector := ProjectDraftSelector{AggregateType: "control_cabinet", AggregateID: uuid.NewString()}
	entry := ProjectDraftEntry{ProjectDraftSelector: selector, Action: "update", BaseVersion: 2, Fields: []ProjectDraftField{{Path: "control_cabinet_nr", Value: "AA-1"}}}

	hub.UpdateDraftState(projectID, userID, []ProjectDraftEntry{entry})
	if len(store.saved) != 1 || store.saved[0].AggregateID != selector.AggregateID {
		t.Fatalf("saved = %+v", store.saved)
	}
	hub.ClearDraft(projectID, userID, selector)
	if store.cleared == nil || store.cleared.selectorKey() != selector.selectorKey() {
		t.Fatalf("cleared = %+v", store.cleared)
	}
}

func TestProjectCollaborationDoesNotDuplicateOwnBusEventToLocalClient(t *testing.T) {
	bus := NewInMemoryBus()
	defer bus.Close()
	hub := NewProjectCollaborationHub(WithProjectCollaborationBus(bus, "node-a"))
	defer hub.Close()
	projectID := uuid.New()
	client := registerProjectTestClient(hub, projectID, uuid.New(), 8)
	drainSocket(client.socket)
	change := ProjectChange{ProjectID: projectID, Revision: 1, EventID: uuid.New(), AggregateType: "project", AggregateID: projectID, Action: "updated", OccurredAt: time.Now().UTC()}
	if err := hub.BroadcastProjectChange(t.Context(), change); err != nil {
		t.Fatal(err)
	}
	_ = receiveSocketMessageOfType(t, client.socket, projectCollaborationMessageProjectChange)
	assertNoSocketMessageOfType(t, client.socket, projectCollaborationMessageProjectChange)
}

func TestProjectCollaborationUnregisterIsIdempotentAcrossTabs(t *testing.T) {
	hub := NewProjectCollaborationHub()
	defer hub.Close()
	projectID, userID := uuid.New(), uuid.New()
	first := registerProjectTestClient(hub, projectID, userID, 8)
	second := registerProjectTestClient(hub, projectID, userID, 8)
	hub.UpdateDraftState(projectID, userID, []ProjectDraftEntry{{
		ProjectDraftSelector: ProjectDraftSelector{AggregateType: "field_device", AggregateID: uuid.NewString()},
		Action:               "update", Fields: []ProjectDraftField{{Path: "bmk", Value: "draft"}},
	}})

	hub.Unregister(first)
	hub.Unregister(first)

	hub.mu.RLock()
	room := hub.rooms[projectID]
	connections := room.connectionByID[userID]
	drafts := len(room.draftStates[userID])
	hub.mu.RUnlock()
	if connections != 1 || drafts != 1 {
		t.Fatalf("second tab state was changed by duplicate unregister: connections=%d drafts=%d", connections, drafts)
	}
	hub.Unregister(second)
}

func TestProjectCollaborationBroadcastDoesNotCreateEmptyRoom(t *testing.T) {
	hub := NewProjectCollaborationHub()
	defer hub.Close()
	projectID := uuid.New()
	if err := hub.BroadcastProjectChange(t.Context(), ProjectChange{
		ProjectID: projectID, Revision: 1, EventID: uuid.New(), AggregateType: "project",
		AggregateID: projectID, Action: "updated", OccurredAt: time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}
	hub.mu.RLock()
	rooms := len(hub.rooms)
	hub.mu.RUnlock()
	if rooms != 0 {
		t.Fatalf("broadcast retained %d empty rooms", rooms)
	}
}

type staticProjectRevision struct{ revision int64 }

func (s staticProjectRevision) CurrentRevision(context.Context, uuid.UUID) (int64, error) {
	return s.revision, nil
}

type recordingProjectDraftStore struct {
	saved   []ProjectDraftEntry
	cleared *ProjectDraftSelector
}

func (s *recordingProjectDraftStore) LoadProjectDraftStates(context.Context, uuid.UUID) ([]ProjectDraftState, error) {
	return nil, nil
}

func (s *recordingProjectDraftStore) SaveProjectDraftState(_ context.Context, _, _ uuid.UUID, entries []ProjectDraftEntry) error {
	s.saved = cloneProjectDraftEntries(entries)
	return nil
}

func (s *recordingProjectDraftStore) ClearProjectDraft(_ context.Context, _, _ uuid.UUID, selector *ProjectDraftSelector) error {
	if selector != nil {
		copy := *selector
		s.cleared = &copy
	}
	return nil
}

func ptrUUID(id uuid.UUID) *uuid.UUID { return &id }

func registerProjectTestClient(hub *ProjectCollaborationHub, projectID, userID uuid.UUID, buffer int) *projectCollaborationClient {
	client := &projectCollaborationClient{hub: hub, projectID: projectID, userID: userID, socket: newTestSocket(buffer)}
	hub.Register(client)
	return client
}

func newTestSocket(buffer int) *WebSocketClient {
	return &WebSocketClient{send: make(chan []byte, buffer)}
}

func drainSocket(socket *WebSocketClient) {
	for {
		select {
		case <-socket.send:
		default:
			return
		}
	}
}

func receiveSocketMessageOfType(t *testing.T, socket *WebSocketClient, messageType string) map[string]any {
	t.Helper()
	deadline := time.After(time.Second)
	for {
		select {
		case data := <-socket.send:
			message := decodeSocketMessage(t, data)
			if message["type"] == messageType {
				return message
			}
		case <-deadline:
			t.Fatalf("timed out waiting for websocket message type %q", messageType)
			return nil
		}
	}
}

func assertNoSocketMessageOfType(t *testing.T, socket *WebSocketClient, messageType string) {
	t.Helper()
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case data := <-socket.send:
			message := decodeSocketMessage(t, data)
			if message["type"] == messageType {
				t.Fatalf("received duplicate websocket message type %q: %s", messageType, data)
			}
		case <-timeout:
			return
		}
	}
}

func decodeSocketMessage(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var message map[string]any
	if err := json.Unmarshal(data, &message); err != nil {
		t.Fatalf("decode websocket message: %v; payload=%s", err, data)
	}
	return message
}
