package handler

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/google/uuid"
)

// Hub maintains active WebSocket connections organized by project rooms
type Hub struct {
	// Rooms indexed by project ID
	rooms map[uuid.UUID]*Room

	// Channel operations for thread-safe communication
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage

	// Mutex for thread-safe map access
	mu sync.RWMutex
}

// Room represents a project-specific room with connected clients
type Room struct {
	projectID uuid.UUID
	clients   map[*Client]bool
	mu        sync.RWMutex
}

// BroadcastMessage represents a message to be sent to all clients in a room
type BroadcastMessage struct {
	projectID uuid.UUID
	message   []byte
	exclude   *Client // Optional: exclude originator from broadcast
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[uuid.UUID]*Room),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *BroadcastMessage, 256),
	}
}

// Run starts the hub's main event loop
// This should be run in a separate goroutine: go hub.Run()
func (h *Hub) Run() {
	log.Println("WebSocket Hub started")

	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case msg := <-h.broadcast:
			h.handleBroadcast(msg)
		}
	}
}

// handleRegister adds a client to the appropriate room and notifies other users
func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	room := h.getOrCreateRoomLocked(client.projectID)
	h.mu.Unlock()

	room.mu.Lock()
	room.clients[client] = true
	room.mu.Unlock()

	log.Printf("Client registered: user=%s, project=%s, total_clients=%d",
		client.userID, client.projectID, len(room.clients))

	// Send current presence list to the new client
	h.sendPresenceList(client)

	// Notify other clients about the new user
	h.broadcastPresenceJoin(client)
}

// handleUnregister removes a client from their room and notifies other users
func (h *Hub) handleUnregister(client *Client) {
	h.mu.RLock()
	room, exists := h.rooms[client.projectID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.Lock()
	if _, ok := room.clients[client]; ok {
		delete(room.clients, client)
		close(client.send)
	}
	clientCount := len(room.clients)
	room.mu.Unlock()

	log.Printf("Client unregistered: user=%s, project=%s, remaining_clients=%d",
		client.userID, client.projectID, clientCount)

	// Notify other clients about the user leaving
	h.broadcastPresenceLeave(client)

	// Clean up empty rooms
	if clientCount == 0 {
		h.mu.Lock()
		delete(h.rooms, client.projectID)
		h.mu.Unlock()
		log.Printf("Room cleaned up: project=%s", client.projectID)
	}
}

// handleBroadcast sends a message to all clients in a room (except excluded)
func (h *Hub) handleBroadcast(msg *BroadcastMessage) {
	h.mu.RLock()
	room, exists := h.rooms[msg.projectID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for client := range room.clients {
		// Skip the originator if specified
		if msg.exclude != nil && client == msg.exclude {
			continue
		}

		select {
		case client.send <- msg.message:
		default:
			// Client's send channel is full, skip this message
			log.Printf("Warning: Client send channel full, message dropped: user=%s", client.userID)
		}
	}
}

// getOrCreateRoomLocked retrieves or creates a room for the given project ID
// Caller must hold h.mu lock
func (h *Hub) getOrCreateRoomLocked(projectID uuid.UUID) *Room {
	room, exists := h.rooms[projectID]
	if !exists {
		room = &Room{
			projectID: projectID,
			clients:   make(map[*Client]bool),
		}
		h.rooms[projectID] = room
		log.Printf("Room created: project=%s", projectID)
	}
	return room
}

// getUsersInRoom returns a list of user presence data for all clients in a room
func (h *Hub) getUsersInRoom(projectID uuid.UUID) []dto.WSPresenceUser {
	h.mu.RLock()
	room, exists := h.rooms[projectID]
	h.mu.RUnlock()

	if !exists {
		return []dto.WSPresenceUser{}
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	users := make([]dto.WSPresenceUser, 0, len(room.clients))
	seen := make(map[uuid.UUID]bool) // Deduplicate by user ID

	for client := range room.clients {
		if !seen[client.userID] {
			users = append(users, dto.WSPresenceUser{
				UserID:    client.userID,
				FirstName: client.user.FirstName,
				LastName:  client.user.LastName,
				Email:     client.user.Email,
			})
			seen[client.userID] = true
		}
	}

	return users
}

// sendPresenceList sends the current list of users in the room to a specific client
func (h *Hub) sendPresenceList(client *Client) {
	users := h.getUsersInRoom(client.projectID)

	message := dto.WSMessage{
		Action:    dto.ActionPresenceList,
		ProjectID: client.projectID,
		Payload:   mustMarshal(dto.WSPresenceListPayload{Users: users}),
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling presence list: %v", err)
		return
	}

	select {
	case client.send <- data:
	default:
		log.Printf("Warning: Could not send presence list to client: user=%s", client.userID)
	}
}

// broadcastPresenceJoin notifies all clients in a room about a new user joining
func (h *Hub) broadcastPresenceJoin(client *Client) {
	message := dto.WSMessage{
		Action:    dto.ActionPresenceJoin,
		ProjectID: client.projectID,
		Payload: mustMarshal(dto.WSPresenceUserPayload{
			User: dto.WSPresenceUser{
				UserID:    client.userID,
				FirstName: client.user.FirstName,
				LastName:  client.user.LastName,
				Email:     client.user.Email,
			},
		}),
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling presence join: %v", err)
		return
	}

	// Broadcast to all except the joining client
	h.broadcast <- &BroadcastMessage{
		projectID: client.projectID,
		message:   data,
		exclude:   client,
	}
}

// broadcastPresenceLeave notifies all clients in a room about a user leaving
func (h *Hub) broadcastPresenceLeave(client *Client) {
	message := dto.WSMessage{
		Action:    dto.ActionPresenceLeave,
		ProjectID: client.projectID,
		Payload: mustMarshal(dto.WSPresenceLeavePayload{
			UserID: client.userID,
		}),
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling presence leave: %v", err)
		return
	}

	h.broadcast <- &BroadcastMessage{
		projectID: client.projectID,
		message:   data,
		exclude:   nil, // Send to all remaining clients
	}
}

// BroadcastToRoom sends a message to all clients in a room (excluding originator)
func (h *Hub) BroadcastToRoom(projectID uuid.UUID, message []byte, exclude *Client) {
	h.broadcast <- &BroadcastMessage{
		projectID: projectID,
		message:   message,
		exclude:   exclude,
	}
}

// mustMarshal marshals a value to JSON or panics (used for internal messages)
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		log.Panicf("Failed to marshal internal message: %v", err)
	}
	return data
}
