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
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"github.com/google/uuid"
)

const (
	systemNotificationWriteWait   = 10 * time.Second
	systemNotificationPongWait    = 60 * time.Second
	systemNotificationPingPeriod  = 25 * time.Second
	systemNotificationMaxMessage  = 4096
	systemNotificationPublishWait = 2 * time.Second

	SystemNotificationEventCreated = string(domainNotification.SystemNotificationChangeCreated)
	SystemNotificationEventUpdated = string(domainNotification.SystemNotificationChangeUpdated)
	SystemNotificationEventDeleted = string(domainNotification.SystemNotificationChangeDeleted)
	SystemNotificationEventReadAll = string(domainNotification.SystemNotificationChangeReadAll)
)

var systemNotificationSocketConfig = WebSocketConfig{
	WriteWait:       systemNotificationWriteWait,
	PongWait:        systemNotificationPongWait,
	PingPeriod:      systemNotificationPingPeriod,
	MaxMessageBytes: systemNotificationMaxMessage,
}

type SystemNotificationPayload struct {
	ID           uuid.UUID         `json:"id"`
	RecipientID  uuid.UUID         `json:"recipient_id"`
	ActorID      *uuid.UUID        `json:"actor_id,omitempty"`
	EventKey     string            `json:"event_key"`
	Title        string            `json:"title"`
	Body         string            `json:"body"`
	ResourceType string            `json:"resource_type"`
	ResourceID   *uuid.UUID        `json:"resource_id,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	ReadAt       *time.Time        `json:"read_at,omitempty"`
	IsImportant  bool              `json:"is_important"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type SystemNotificationEvent struct {
	Type           domainNotification.SystemNotificationChangeType `json:"type"`
	Notification   *SystemNotificationPayload                      `json:"notification,omitempty"`
	NotificationID string                                          `json:"notification_id,omitempty"`
	UnreadCount    int64                                           `json:"unread_count"`
	At             time.Time                                       `json:"at"`
}

type systemNotificationBusEvent struct {
	RecipientID uuid.UUID       `json:"recipient_id"`
	Payload     json.RawMessage `json:"payload"`
}

type systemNotificationClient struct {
	hub         *SystemNotificationHub
	recipientID uuid.UUID
	socket      *WebSocketClient
}

type SystemNotificationHub struct {
	mu             sync.RWMutex
	clients        map[uuid.UUID]map[*systemNotificationClient]struct{}
	bus            apprealtime.Bus
	nodeID         string
	publishTimeout time.Duration
	ctx            context.Context
	cancel         context.CancelFunc
	closeOnce      sync.Once
}

type SystemNotificationHubOption func(*SystemNotificationHub)

func WithSystemNotificationBus(bus apprealtime.Bus, nodeID string) SystemNotificationHubOption {
	return func(h *SystemNotificationHub) {
		h.bus = bus
		h.nodeID = strings.TrimSpace(nodeID)
	}
}

func NewSystemNotificationHub(options ...SystemNotificationHubOption) *SystemNotificationHub {
	ctx, cancel := context.WithCancel(context.Background())
	h := &SystemNotificationHub{
		clients:        make(map[uuid.UUID]map[*systemNotificationClient]struct{}),
		nodeID:         uuid.NewString(),
		publishTimeout: systemNotificationPublishWait,
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
	return h
}

func (h *SystemNotificationHub) Close() {
	h.closeOnce.Do(func() {
		h.cancel()
	})
}

func (h *SystemNotificationHub) startBusSubscription() {
	if h.bus == nil {
		return
	}

	events, err := h.bus.Subscribe(h.ctx, apprealtime.TopicSystemNotifications)
	if err != nil {
		slog.Warn("system notification realtime bus subscription disabled", "err", err)
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

func (h *SystemNotificationHub) handleBusEvent(event apprealtime.Event) {
	if event.Source == h.nodeID {
		return
	}

	var payload systemNotificationBusEvent
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		slog.Warn("ignored invalid system notification bus event", "err", err)
		return
	}
	h.broadcastBytes(payload.RecipientID, payload.Payload)
}

func (h *SystemNotificationHub) Stream(w http.ResponseWriter, r *http.Request, recipientID uuid.UUID) {
	if recipientID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	client := &systemNotificationClient{
		hub:         h,
		recipientID: recipientID,
	}
	socket, err := AcceptWebSocket(w, r, systemNotificationSocketConfig, client.handleMessage, func() {
		h.unregister(client)
	})
	if err != nil {
		return
	}
	client.socket = socket

	h.register(client)
	socket.Run()
}

func (c *systemNotificationClient) handleMessage(data []byte) {
	slog.Debug(
		"ignored unsupported inbound system notification websocket message",
		"bytes", len(data),
	)
}

func (h *SystemNotificationHub) PublishSystemNotificationChange(_ context.Context, change domainNotification.SystemNotificationChange) {
	if change.Notification != nil {
		change.RecipientID = change.Notification.RecipientID
		change.NotificationID = change.Notification.ID
	}
	if change.RecipientID == uuid.Nil {
		return
	}
	if change.OccurredAt.IsZero() {
		change.OccurredAt = time.Now().UTC()
	}

	event := SystemNotificationEvent{
		Type:        change.Type,
		UnreadCount: change.UnreadCount,
		At:          change.OccurredAt,
	}
	if change.Notification != nil {
		event.Notification = mapSystemNotificationPayload(*change.Notification)
	}
	if change.NotificationID != uuid.Nil {
		event.NotificationID = change.NotificationID.String()
	}
	h.broadcastDistributed(change.RecipientID, event)
}

func (h *SystemNotificationHub) PublishSystemNotificationCreated(ctx context.Context, notification domainNotification.SystemNotification, unreadCount int64) {
	h.PublishSystemNotificationChange(ctx, domainNotification.SystemNotificationChange{
		Type:         domainNotification.SystemNotificationChangeCreated,
		RecipientID:  notification.RecipientID,
		Notification: &notification,
		UnreadCount:  unreadCount,
	})
}

func (h *SystemNotificationHub) PublishSystemNotificationUpdated(ctx context.Context, notification domainNotification.SystemNotification, unreadCount int64) {
	h.PublishSystemNotificationChange(ctx, domainNotification.SystemNotificationChange{
		Type:         domainNotification.SystemNotificationChangeUpdated,
		RecipientID:  notification.RecipientID,
		Notification: &notification,
		UnreadCount:  unreadCount,
	})
}

func (h *SystemNotificationHub) PublishSystemNotificationDeleted(ctx context.Context, recipientID, notificationID uuid.UUID, unreadCount int64) {
	h.PublishSystemNotificationChange(ctx, domainNotification.SystemNotificationChange{
		Type:           domainNotification.SystemNotificationChangeDeleted,
		RecipientID:    recipientID,
		NotificationID: notificationID,
		UnreadCount:    unreadCount,
	})
}

func (h *SystemNotificationHub) PublishSystemNotificationsReadAll(ctx context.Context, recipientID uuid.UUID, unreadCount int64) {
	h.PublishSystemNotificationChange(ctx, domainNotification.SystemNotificationChange{
		Type:        domainNotification.SystemNotificationChangeReadAll,
		RecipientID: recipientID,
		UnreadCount: unreadCount,
	})
}

func (h *SystemNotificationHub) register(client *systemNotificationClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.recipientID] == nil {
		h.clients[client.recipientID] = make(map[*systemNotificationClient]struct{})
	}
	h.clients[client.recipientID][client] = struct{}{}
}

func (h *SystemNotificationHub) unregister(client *systemNotificationClient) {
	h.mu.Lock()
	if recipientClients := h.clients[client.recipientID]; recipientClients != nil {
		delete(recipientClients, client)
		if len(recipientClients) == 0 {
			delete(h.clients, client.recipientID)
		}
	}
	h.mu.Unlock()

	if client.socket != nil {
		client.socket.CloseSend()
	}
}

func (h *SystemNotificationHub) broadcast(recipientID uuid.UUID, event SystemNotificationEvent) {
	b, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.broadcastBytes(recipientID, b)
}

func (h *SystemNotificationHub) broadcastDistributed(recipientID uuid.UUID, event SystemNotificationEvent) {
	b, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.broadcastBytes(recipientID, b)
	h.publishPayload(recipientID, b)
}

func (h *SystemNotificationHub) broadcastBytes(recipientID uuid.UUID, b []byte) {
	h.mu.RLock()
	recipientClients := h.clients[recipientID]
	clients := make([]*systemNotificationClient, 0, len(recipientClients))
	for client := range recipientClients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if client.socket == nil || !client.socket.SendBytes(b) {
			h.unregister(client)
		}
	}
}

func (h *SystemNotificationHub) publishPayload(recipientID uuid.UUID, payload []byte) {
	if h.bus == nil {
		return
	}
	b, err := json.Marshal(systemNotificationBusEvent{
		RecipientID: recipientID,
		Payload:     append(json.RawMessage(nil), payload...),
	})
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), h.publishTimeout)
	defer cancel()
	if err := h.bus.Publish(ctx, apprealtime.NewEvent(apprealtime.TopicSystemNotifications, h.nodeID, b)); err != nil {
		slog.Warn("system notification realtime bus publish failed", "err", err)
	}
}

func mapSystemNotificationPayload(notification domainNotification.SystemNotification) *SystemNotificationPayload {
	return &SystemNotificationPayload{
		ID:           notification.ID,
		RecipientID:  notification.RecipientID,
		ActorID:      notification.ActorID,
		EventKey:     notification.EventKey,
		Title:        notification.Title,
		Body:         notification.Body,
		ResourceType: notification.ResourceType,
		ResourceID:   notification.ResourceID,
		Metadata:     notification.Metadata,
		ReadAt:       notification.ReadAt,
		IsImportant:  notification.IsImportant,
		CreatedAt:    notification.CreatedAt,
		UpdatedAt:    notification.UpdatedAt,
	}
}
