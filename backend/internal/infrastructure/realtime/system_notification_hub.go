package realtime

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"github.com/google/uuid"
)

const (
	systemNotificationWriteWait  = 10 * time.Second
	systemNotificationPongWait   = 60 * time.Second
	systemNotificationPingPeriod = 25 * time.Second
	systemNotificationMaxMessage = 4096

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

type systemNotificationClient struct {
	hub         *SystemNotificationHub
	recipientID uuid.UUID
	socket      *WebSocketClient
}

type SystemNotificationHub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*systemNotificationClient]struct{}
}

func NewSystemNotificationHub() *SystemNotificationHub {
	return &SystemNotificationHub{clients: make(map[uuid.UUID]map[*systemNotificationClient]struct{})}
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
	socket, err := AcceptWebSocket(w, r, systemNotificationSocketConfig, nil, func() {
		h.unregister(client)
	})
	if err != nil {
		return
	}
	client.socket = socket

	h.register(client)
	socket.Run()
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
	h.broadcast(change.RecipientID, event)
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

	h.mu.RLock()
	recipientClients := h.clients[recipientID]
	clients := make([]*systemNotificationClient, 0, len(recipientClients))
	for client := range recipientClients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if client.socket == nil || !client.socket.SendBytes(b) {
			go h.unregister(client)
		}
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
