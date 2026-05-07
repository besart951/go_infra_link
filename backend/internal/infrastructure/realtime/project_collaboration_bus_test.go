package realtime

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProjectCollaborationBroadcastRoutesAcrossInstances(t *testing.T) {
	bus := NewInMemoryBus()
	defer bus.Close()

	hubA := NewProjectCollaborationHub(WithProjectCollaborationBus(bus, "node-a"))
	defer hubA.Close()
	hubB := NewProjectCollaborationHub(WithProjectCollaborationBus(bus, "node-b"))
	defer hubB.Close()

	projectID := uuid.New()
	client := registerProjectTestClient(hubB, projectID, uuid.New(), 8)
	drainSocket(client.socket)

	entityID := uuid.New().String()
	hubA.BroadcastRefreshRequest(projectID, nil, projectCollaborationRefreshScopeFieldDevice, []string{entityID})

	message := receiveSocketMessageOfType(t, client.socket, projectCollaborationMessageRefreshRequest)
	if message["project_id"] != projectID.String() {
		t.Fatalf("project_id = %v, want %s", message["project_id"], projectID)
	}
	if message["scope"] != projectCollaborationRefreshScopeFieldDevice {
		t.Fatalf("scope = %v", message["scope"])
	}
}

func TestProjectCollaborationEditStateRoutesAcrossInstances(t *testing.T) {
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

	deviceID := uuid.New().String()
	hubA.UpdateEditState(projectID, editorID, []ProjectFieldDeviceByFields{
		{
			DeviceID:      deviceID,
			ChangedFields: []string{"text_fix"},
			FieldValues:   map[string]any{"text_fix": "TF-100"},
		},
	})

	message := receiveSocketMessageOfType(t, viewer.socket, projectCollaborationMessageEditStates)
	editStates, ok := message["edit_states"].([]any)
	if !ok || len(editStates) != 1 {
		t.Fatalf("edit_states = %#v", message["edit_states"])
	}
	state := editStates[0].(map[string]any)
	if state["user_id"] != editorID.String() {
		t.Fatalf("user_id = %v, want %s", state["user_id"], editorID)
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

	hub.BroadcastRefreshRequest(projectID, nil, projectCollaborationRefreshScopeFieldDevice, []string{uuid.New().String()})

	_ = receiveSocketMessageOfType(t, client.socket, projectCollaborationMessageRefreshRequest)
	assertNoSocketMessageOfType(t, client.socket, projectCollaborationMessageRefreshRequest)
}

func registerProjectTestClient(hub *ProjectCollaborationHub, projectID, userID uuid.UUID, buffer int) *projectCollaborationClient {
	client := &projectCollaborationClient{
		hub:       hub,
		projectID: projectID,
		userID:    userID,
		socket:    newTestSocket(buffer),
	}
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
