package project

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	projectCollaborationWriteWait  = 10 * time.Second
	projectCollaborationPongWait   = 60 * time.Second
	projectCollaborationPingPeriod = 25 * time.Second
	projectCollaborationMaxMessage = 32 * 1024

	projectCollaborationMessageSnapshot       = "snapshot"
	projectCollaborationMessagePresence       = "presence"
	projectCollaborationMessageEditStates     = "edit_states"
	projectCollaborationMessageRefreshRequest = "refresh_request"
	projectCollaborationMessageEditState      = "edit_state"

	projectCollaborationScopeFieldDevice = "field_device"
)

var projectCollaborationUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}

		originURL, err := url.Parse(origin)
		if err != nil {
			return false
		}

		requestHost := r.Host
		if parsedHost, err := url.Parse("http://" + r.Host); err == nil && parsedHost.Hostname() != "" {
			requestHost = parsedHost.Hostname()
		}

		return strings.EqualFold(originURL.Hostname(), requestHost)
	},
}

type ProjectCollaboratorPresence struct {
	UserID      uuid.UUID `json:"user_id"`
	ConnectedAt time.Time `json:"connected_at"`
	LastSeenAt  time.Time `json:"last_seen_at"`
}

type ProjectFieldDeviceByFields struct {
	DeviceID      string                 `json:"device_id"`
	ChangedFields []string               `json:"changed_fields"`
	FieldValues   map[string]interface{} `json:"field_values,omitempty"`
}

type ProjectFieldDeviceEditState struct {
	UserID    uuid.UUID                    `json:"user_id"`
	Devices   []ProjectFieldDeviceByFields `json:"devices"`
	UpdatedAt time.Time                    `json:"updated_at"`
}

type projectCollaborationSnapshotMessage struct {
	Type       string                        `json:"type"`
	ProjectID  uuid.UUID                     `json:"project_id"`
	Presence   []ProjectCollaboratorPresence `json:"presence"`
	EditStates []ProjectFieldDeviceEditState `json:"edit_states"`
	At         time.Time                     `json:"at"`
}

type projectCollaborationPresenceMessage struct {
	Type      string                        `json:"type"`
	ProjectID uuid.UUID                     `json:"project_id"`
	Presence  []ProjectCollaboratorPresence `json:"presence"`
	At        time.Time                     `json:"at"`
}

type projectCollaborationEditStatesMessage struct {
	Type       string                        `json:"type"`
	ProjectID  uuid.UUID                     `json:"project_id"`
	EditStates []ProjectFieldDeviceEditState `json:"edit_states"`
	At         time.Time                     `json:"at"`
}

type ProjectCollaborationRefreshMessage struct {
	Type      string    `json:"type"`
	ProjectID uuid.UUID `json:"project_id"`
	Scope     string    `json:"scope"`
	ActorID   string    `json:"actor_id,omitempty"`
	DeviceIDs []string  `json:"device_ids,omitempty"`
	At        time.Time `json:"at"`
}

type projectCollaborationClientMessage struct {
	Type      string                       `json:"type"`
	Devices   []ProjectFieldDeviceByFields `json:"devices,omitempty"`
	Scope     string                       `json:"scope,omitempty"`
	DeviceIDs []string                     `json:"device_ids,omitempty"`
}

type projectCollaborationClient struct {
	hub       *ProjectCollaborationHub
	projectID uuid.UUID
	userID    uuid.UUID
	conn      *websocket.Conn
	send      chan []byte
	closed    sync.Once
}

type projectCollaborationRoom struct {
	clients        map[*projectCollaborationClient]struct{}
	connectionByID map[uuid.UUID]int
	presence       map[uuid.UUID]ProjectCollaboratorPresence
	editStates     map[uuid.UUID][]ProjectFieldDeviceByFields
}

type ProjectCollaborationHub struct {
	mu    sync.RWMutex
	rooms map[uuid.UUID]*projectCollaborationRoom
}

func NewProjectCollaborationHub() *ProjectCollaborationHub {
	return &ProjectCollaborationHub{rooms: make(map[uuid.UUID]*projectCollaborationRoom)}
}

func (h *ProjectCollaborationHub) Register(client *projectCollaborationClient) {
	h.mu.Lock()
	room := h.ensureRoomLocked(client.projectID)
	room.clients[client] = struct{}{}
	room.connectionByID[client.userID] += 1
	now := time.Now().UTC()
	if _, exists := room.presence[client.userID]; !exists {
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
	editStates := snapshotEditStates(room)
	h.mu.Unlock()

	h.sendToClient(client, projectCollaborationSnapshotMessage{
		Type:       projectCollaborationMessageSnapshot,
		ProjectID:  client.projectID,
		Presence:   presence,
		EditStates: editStates,
		At:         now,
	})
	h.broadcast(client.projectID, projectCollaborationPresenceMessage{
		Type:      projectCollaborationMessagePresence,
		ProjectID: client.projectID,
		Presence:  presence,
		At:        now,
	})
}

func (h *ProjectCollaborationHub) Unregister(client *projectCollaborationClient) {
	var (
		presence   []ProjectCollaboratorPresence
		editStates []ProjectFieldDeviceEditState
		now        = time.Now().UTC()
		shouldSend bool
	)

	h.mu.Lock()
	room, ok := h.rooms[client.projectID]
	if ok {
		if _, exists := room.clients[client]; exists {
			delete(room.clients, client)
		}
		if room.connectionByID[client.userID] > 1 {
			room.connectionByID[client.userID] -= 1
			presenceState := room.presence[client.userID]
			presenceState.LastSeenAt = now
			room.presence[client.userID] = presenceState
		} else {
			delete(room.connectionByID, client.userID)
			delete(room.presence, client.userID)
			delete(room.editStates, client.userID)
		}

		if len(room.clients) == 0 {
			delete(h.rooms, client.projectID)
		} else {
			presence = snapshotPresence(room)
			editStates = snapshotEditStates(room)
			shouldSend = true
		}
	}
	h.mu.Unlock()

	if shouldSend {
		h.broadcast(client.projectID, projectCollaborationPresenceMessage{
			Type:      projectCollaborationMessagePresence,
			ProjectID: client.projectID,
			Presence:  presence,
			At:        now,
		})
		h.broadcast(client.projectID, projectCollaborationEditStatesMessage{
			Type:       projectCollaborationMessageEditStates,
			ProjectID:  client.projectID,
			EditStates: editStates,
			At:         now,
		})
	}

	client.closeSend()
}

func (h *ProjectCollaborationHub) UpdateEditState(projectID, userID uuid.UUID, devices []ProjectFieldDeviceByFields) {
	h.mu.Lock()
	room, ok := h.rooms[projectID]
	if !ok {
		h.mu.Unlock()
		return
	}

	now := time.Now().UTC()
	if len(devices) == 0 {
		delete(room.editStates, userID)
	} else {
		userDevices := make(map[string]ProjectFieldDeviceByFields)
		for _, device := range devices {
			deviceID := strings.TrimSpace(device.DeviceID)
			if deviceID == "" {
				continue
			}
			fields := normalizeIDs(device.ChangedFields)
			if len(fields) > 0 {
				userDevices[deviceID] = ProjectFieldDeviceByFields{
					DeviceID:      deviceID,
					ChangedFields: fields,
					FieldValues:   normalizeFieldValues(device.FieldValues),
				}
			}
		}

		if len(userDevices) == 0 {
			delete(room.editStates, userID)
		} else {
			normalizedDevices := make([]ProjectFieldDeviceByFields, 0, len(userDevices))
			for _, device := range userDevices {
				normalizedDevices = append(normalizedDevices, device)
			}
			sort.Slice(normalizedDevices, func(i, j int) bool {
				return normalizedDevices[i].DeviceID < normalizedDevices[j].DeviceID
			})
			room.editStates[userID] = normalizedDevices
		}
	}

	editStates := snapshotEditStates(room)
	h.mu.Unlock()

	h.broadcast(projectID, projectCollaborationEditStatesMessage{
		Type:       projectCollaborationMessageEditStates,
		ProjectID:  projectID,
		EditStates: editStates,
		At:         now,
	})
}

func (h *ProjectCollaborationHub) BroadcastRefreshRequest(projectID uuid.UUID, actorID *uuid.UUID, scope string, deviceIDs []string) {
	actor := ""
	if actorID != nil {
		actor = actorID.String()
	}

	h.broadcast(projectID, ProjectCollaborationRefreshMessage{
		Type:      projectCollaborationMessageRefreshRequest,
		ProjectID: projectID,
		Scope:     scope,
		ActorID:   actor,
		DeviceIDs: normalizeIDs(deviceIDs),
		At:        time.Now().UTC(),
	})
}

func (h *ProjectCollaborationHub) ensureRoomLocked(projectID uuid.UUID) *projectCollaborationRoom {
	room, ok := h.rooms[projectID]
	if !ok {
		room = &projectCollaborationRoom{
			clients:        make(map[*projectCollaborationClient]struct{}),
			connectionByID: make(map[uuid.UUID]int),
			presence:       make(map[uuid.UUID]ProjectCollaboratorPresence),
			editStates:     make(map[uuid.UUID][]ProjectFieldDeviceByFields),
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
		select {
		case client.send <- b:
		default:
			go h.Unregister(client)
		}
	}
}

func (h *ProjectCollaborationHub) sendToClient(client *projectCollaborationClient, payload any) {
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	select {
	case client.send <- b:
	default:
		go h.Unregister(client)
	}
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

func snapshotEditStates(room *projectCollaborationRoom) []ProjectFieldDeviceEditState {
	items := make([]ProjectFieldDeviceEditState, 0, len(room.editStates))
	now := time.Now().UTC()

	// Collect all items with their timestamps
	type editStateWithTime struct {
		userID uuid.UUID
		state  ProjectFieldDeviceEditState
	}
	var itemsWithTime []editStateWithTime

	for userID, devices := range room.editStates {
		normalizedDevices := make([]ProjectFieldDeviceByFields, 0, len(devices))
		for _, device := range devices {
			normalizedDevices = append(normalizedDevices, ProjectFieldDeviceByFields{
				DeviceID:      device.DeviceID,
				ChangedFields: append([]string(nil), device.ChangedFields...),
				FieldValues:   normalizeFieldValues(device.FieldValues),
			})
		}
		itemsWithTime = append(itemsWithTime, editStateWithTime{
			userID: userID,
			state: ProjectFieldDeviceEditState{
				UserID:    userID,
				Devices:   normalizedDevices,
				UpdatedAt: now,
			},
		})
	}

	// Sort by user ID for consistency
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

func normalizeFieldValues(values map[string]interface{}) map[string]interface{} {
	if len(values) == 0 {
		return nil
	}

	result := make(map[string]interface{}, len(values))
	for key, value := range values {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		result[trimmedKey] = value
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func (c *projectCollaborationClient) closeSend() {
	c.closed.Do(func() {
		close(c.send)
	})
}

func (c *projectCollaborationClient) readPump() {
	defer func() {
		c.hub.Unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(projectCollaborationMaxMessage)
	_ = c.conn.SetReadDeadline(time.Now().Add(projectCollaborationPongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(projectCollaborationPongWait))
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var message projectCollaborationClientMessage
		if err := json.Unmarshal(data, &message); err != nil {
			continue
		}

		switch message.Type {
		case projectCollaborationMessageEditState:
			c.hub.UpdateEditState(c.projectID, c.userID, message.Devices)
		case projectCollaborationMessageRefreshRequest:
			scope := strings.TrimSpace(message.Scope)
			if scope == "" {
				scope = projectCollaborationScopeFieldDevice
			}
			c.hub.BroadcastRefreshRequest(c.projectID, &c.userID, scope, message.DeviceIDs)
		}
	}
}

func (c *projectCollaborationClient) writePump() {
	ticker := time.NewTicker(projectCollaborationPingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(projectCollaborationWriteWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := writer.Write(message); err != nil {
				_ = writer.Close()
				return
			}
			if err := writer.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(projectCollaborationWriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *ProjectHandler) StreamProjectCollaboration(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !h.ensureProjectAccess(c, projectID) {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}

	conn, err := projectCollaborationUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &projectCollaborationClient{
		hub:       h.collaboration,
		projectID: projectID,
		userID:    userID,
		conn:      conn,
		send:      make(chan []byte, 16),
	}

	h.collaboration.Register(client)
	go client.writePump()
	client.readPump()
}
