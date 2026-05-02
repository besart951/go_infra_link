package realtime

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultWebSocketReadBufferSize  = 1024
	defaultWebSocketWriteBufferSize = 1024
	defaultWebSocketSendBufferSize  = 16
	defaultWebSocketWriteWait       = 10 * time.Second
	defaultWebSocketPongWait        = 60 * time.Second
	defaultWebSocketPingPeriod      = 25 * time.Second
	defaultWebSocketMaxMessageBytes = 4096
)

type WebSocketConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	SendBufferSize  int
	WriteWait       time.Duration
	PongWait        time.Duration
	PingPeriod      time.Duration
	MaxMessageBytes int64
	CheckOrigin     func(*http.Request) bool
}

type WebSocketClient struct {
	conn      *websocket.Conn
	send      chan []byte
	config    WebSocketConfig
	onMessage func([]byte)
	onClose   func()
	closed    sync.Once
}

func AcceptWebSocket(
	w http.ResponseWriter,
	r *http.Request,
	config WebSocketConfig,
	onMessage func([]byte),
	onClose func(),
) (*WebSocketClient, error) {
	config = config.withDefaults()
	upgrader := websocket.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
		CheckOrigin:     config.CheckOrigin,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &WebSocketClient{
		conn:      conn,
		send:      make(chan []byte, config.SendBufferSize),
		config:    config,
		onMessage: onMessage,
		onClose:   onClose,
	}, nil
}

func (c *WebSocketClient) Run() {
	go c.writePump()
	c.readPump()
}

func (c *WebSocketClient) SendBytes(message []byte) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	select {
	case c.send <- message:
		return true
	default:
		return false
	}
}

func (c *WebSocketClient) CloseSend() {
	c.closed.Do(func() {
		close(c.send)
	})
}

func SameHostOrigin(r *http.Request) bool {
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
}

func (config WebSocketConfig) withDefaults() WebSocketConfig {
	if config.ReadBufferSize <= 0 {
		config.ReadBufferSize = defaultWebSocketReadBufferSize
	}
	if config.WriteBufferSize <= 0 {
		config.WriteBufferSize = defaultWebSocketWriteBufferSize
	}
	if config.SendBufferSize <= 0 {
		config.SendBufferSize = defaultWebSocketSendBufferSize
	}
	if config.WriteWait <= 0 {
		config.WriteWait = defaultWebSocketWriteWait
	}
	if config.PongWait <= 0 {
		config.PongWait = defaultWebSocketPongWait
	}
	if config.PingPeriod <= 0 {
		config.PingPeriod = defaultWebSocketPingPeriod
	}
	if config.MaxMessageBytes <= 0 {
		config.MaxMessageBytes = defaultWebSocketMaxMessageBytes
	}
	if config.CheckOrigin == nil {
		config.CheckOrigin = SameHostOrigin
	}
	return config
}

func (c *WebSocketClient) readPump() {
	defer func() {
		if c.onClose != nil {
			c.onClose()
		}
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(c.config.MaxMessageBytes)
	_ = c.conn.SetReadDeadline(time.Now().Add(c.config.PongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(c.config.PongWait))
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		if c.onMessage != nil {
			c.onMessage(data)
		}
	}
}

func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(c.config.PingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteWait))
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
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
