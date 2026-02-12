package handler

import (
	"encoding/json"
	"log"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512 KB
)

// Client represents a WebSocket client connection
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	userID    uuid.UUID
	projectID uuid.UUID
	user      *user.User
	handler   MessageHandler // Interface for handling incoming messages
}

// MessageHandler defines the interface for handling WebSocket messages
type MessageHandler interface {
	HandleMessage(client *Client, message []byte) error
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID, projectID uuid.UUID, usr *user.User, handler MessageHandler) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		send:      make(chan []byte, 256),
		userID:    userID,
		projectID: projectID,
		user:      usr,
		handler:   handler,
	}
}

// readPump pumps messages from the WebSocket connection to the hub
// The application runs readPump in a per-connection goroutine
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle the message through the handler interface
		if err := c.handler.HandleMessage(c, message); err != nil {
			log.Printf("Error handling message from user %s: %v", c.userID, err)

			// Send error response back to client
			errResp := dto.WSMessage{
				Action:    "ERROR",
				ProjectID: c.projectID,
				Payload: mustMarshal(dto.WSErrorResponse{
					Error:   "message_handling_failed",
					Message: err.Error(),
				}),
			}

			if errData, marshalErr := json.Marshal(errResp); marshalErr == nil {
				select {
				case c.send <- errData:
				default:
					log.Printf("Could not send error response to client %s", c.userID)
				}
			}
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
// The application runs writePump in a per-connection goroutine
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage sends a message to this specific client
func (c *Client) SendMessage(message []byte) error {
	select {
	case c.send <- message:
		return nil
	default:
		return ErrSendChannelFull
	}
}

// ErrSendChannelFull is returned when the client's send channel is full
var ErrSendChannelFull = &ClientError{Code: "send_channel_full", Message: "Client send channel is full"}

// ClientError represents a client-specific error
type ClientError struct {
	Code    string
	Message string
}

func (e *ClientError) Error() string {
	return e.Message
}
