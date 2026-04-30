package project

import (
	"encoding/json"
	"maps"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
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
	projectCollaborationMessageEntityDelta    = "entity_delta"
	projectCollaborationMessageRefreshRequest = "refresh_request"
	projectCollaborationMessageEditState      = "edit_state"
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
	DeviceID      string         `json:"device_id"`
	ChangedFields []string       `json:"changed_fields"`
	FieldValues   map[string]any `json:"field_values,omitempty"`
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
	EntityIDs []string  `json:"entity_ids,omitempty"`
	DeviceIDs []string  `json:"device_ids,omitempty"`
	At        time.Time `json:"at"`
}

type projectCollaborationControlCabinet struct {
	ID               uuid.UUID `json:"id"`
	BuildingID       uuid.UUID `json:"building_id"`
	ControlCabinetNr *string   `json:"control_cabinet_nr"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type projectCollaborationSPSController struct {
	ID                uuid.UUID `json:"id"`
	ControlCabinetID  uuid.UUID `json:"control_cabinet_id"`
	GADevice          *string   `json:"ga_device"`
	DeviceName        string    `json:"device_name"`
	DeviceDescription *string   `json:"device_description,omitempty"`
	DeviceLocation    *string   `json:"device_location,omitempty"`
	IPAddress         *string   `json:"ip_address,omitempty"`
	Subnet            *string   `json:"subnet,omitempty"`
	Gateway           *string   `json:"gateway,omitempty"`
	Vlan              *string   `json:"vlan,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type projectCollaborationEntityDeltaMessage struct {
	Type            string                               `json:"type"`
	ProjectID       uuid.UUID                            `json:"project_id"`
	Scope           string                               `json:"scope"`
	ActorID         string                               `json:"actor_id,omitempty"`
	ControlCabinets []projectCollaborationControlCabinet `json:"control_cabinets,omitempty"`
	SPSControllers  []projectCollaborationSPSController  `json:"sps_controllers,omitempty"`
	FieldDevices    []map[string]any                     `json:"field_devices,omitempty"`
	At              time.Time                            `json:"at"`
}

type projectCollaborationClientMessage struct {
	Type            string                               `json:"type"`
	Devices         []ProjectFieldDeviceByFields         `json:"devices,omitempty"`
	Scope           string                               `json:"scope,omitempty"`
	EntityIDs       []string                             `json:"entity_ids,omitempty"`
	DeviceIDs       []string                             `json:"device_ids,omitempty"`
	ControlCabinets []projectCollaborationControlCabinet `json:"control_cabinets,omitempty"`
	SPSControllers  []projectCollaborationSPSController  `json:"sps_controllers,omitempty"`
	FieldDevices    []map[string]any                     `json:"field_devices,omitempty"`
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
		delete(room.clients, client)
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

func (h *ProjectCollaborationHub) BroadcastRefreshRequest(projectID uuid.UUID, actorID *uuid.UUID, scope string, entityIDs []string) {
	actor := ""
	if actorID != nil {
		actor = actorID.String()
	}

	normalizedEntityIDs := normalizeIDs(entityIDs)

	h.broadcast(projectID, ProjectCollaborationRefreshMessage{
		Type:      projectCollaborationMessageRefreshRequest,
		ProjectID: projectID,
		Scope:     scope,
		ActorID:   actor,
		EntityIDs: normalizedEntityIDs,
		DeviceIDs: refreshDeviceIDs(scope, normalizedEntityIDs),
		At:        time.Now().UTC(),
	})
}

func (h *ProjectCollaborationHub) BroadcastControlCabinetDelta(projectID uuid.UUID, actorID *uuid.UUID, controlCabinet projectCollaborationControlCabinet) {
	h.broadcastEntityDelta(projectID, actorID, projectRefreshScopeControlCabinet, projectCollaborationEntityDeltaMessage{
		ControlCabinets: []projectCollaborationControlCabinet{controlCabinet},
	})
}

func (h *ProjectCollaborationHub) BroadcastSPSControllerDelta(projectID uuid.UUID, actorID *uuid.UUID, spsController projectCollaborationSPSController) {
	h.broadcastEntityDelta(projectID, actorID, projectRefreshScopeSPSController, projectCollaborationEntityDeltaMessage{
		SPSControllers: []projectCollaborationSPSController{spsController},
	})
}

func (h *ProjectCollaborationHub) BroadcastFieldDeviceDelta(projectID uuid.UUID, actorID *uuid.UUID, fieldDevices []map[string]any) {
	if len(fieldDevices) == 0 {
		return
	}

	h.broadcastEntityDelta(projectID, actorID, projectRefreshScopeFieldDevice, projectCollaborationEntityDeltaMessage{
		FieldDevices: cloneFieldDeviceDeltas(fieldDevices),
	})
}

func (h *ProjectCollaborationHub) broadcastEntityDelta(projectID uuid.UUID, actorID *uuid.UUID, scope string, payload projectCollaborationEntityDeltaMessage) {
	actor := ""
	if actorID != nil {
		actor = actorID.String()
	}

	payload.Type = projectCollaborationMessageEntityDelta
	payload.ProjectID = projectID
	payload.Scope = scope
	payload.ActorID = actor
	payload.At = time.Now().UTC()

	h.broadcast(projectID, payload)
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

func normalizeRefreshEntityIDs(scope string, entityIDs, deviceIDs []string) []string {
	if normalized := normalizeIDs(entityIDs); len(normalized) > 0 {
		return normalized
	}

	if scope == projectRefreshScopeFieldDevice {
		return normalizeIDs(deviceIDs)
	}

	return nil
}

func refreshDeviceIDs(scope string, entityIDs []string) []string {
	if scope != projectRefreshScopeFieldDevice {
		return nil
	}

	return entityIDs
}

func cloneFieldDeviceDeltas(fieldDevices []map[string]any) []map[string]any {
	if len(fieldDevices) == 0 {
		return nil
	}

	cloned := make([]map[string]any, 0, len(fieldDevices))
	for _, item := range fieldDevices {
		if item == nil {
			continue
		}

		copied := make(map[string]any, len(item))
		maps.Copy(copied, item)
		cloned = append(cloned, copied)
	}

	if len(cloned) == 0 {
		return nil
	}

	return cloned
}

func toProjectCollaborationControlCabinet(controlCabinet domainFacility.ControlCabinet) projectCollaborationControlCabinet {
	return projectCollaborationControlCabinet{
		ID:               controlCabinet.ID,
		BuildingID:       controlCabinet.BuildingID,
		ControlCabinetNr: controlCabinet.ControlCabinetNr,
		CreatedAt:        controlCabinet.CreatedAt,
		UpdatedAt:        controlCabinet.UpdatedAt,
	}
}

func toProjectCollaborationSPSController(spsController domainFacility.SPSController) projectCollaborationSPSController {
	return projectCollaborationSPSController{
		ID:                spsController.ID,
		ControlCabinetID:  spsController.ControlCabinetID,
		GADevice:          spsController.GADevice,
		DeviceName:        spsController.DeviceName,
		DeviceDescription: spsController.DeviceDescription,
		DeviceLocation:    spsController.DeviceLocation,
		IPAddress:         spsController.IPAddress,
		Subnet:            spsController.Subnet,
		Gateway:           spsController.Gateway,
		Vlan:              spsController.Vlan,
		CreatedAt:         spsController.CreatedAt,
		UpdatedAt:         spsController.UpdatedAt,
	}
}

func normalizeFieldValues(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}

	result := make(map[string]any, len(values))
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
		case projectCollaborationMessageEntityDelta:
			if strings.TrimSpace(message.Scope) != projectRefreshScopeFieldDevice || len(message.FieldDevices) == 0 {
				continue
			}

			c.hub.BroadcastFieldDeviceDelta(c.projectID, &c.userID, message.FieldDevices)
		case projectCollaborationMessageRefreshRequest:
			scope := strings.TrimSpace(message.Scope)
			if scope == "" {
				scope = projectRefreshScopeFieldDevice
			}
			c.hub.BroadcastRefreshRequest(
				c.projectID,
				&c.userID,
				scope,
				normalizeRefreshEntityIDs(scope, message.EntityIDs, message.DeviceIDs),
			)
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
