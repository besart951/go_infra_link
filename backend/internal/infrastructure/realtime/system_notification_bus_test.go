package realtime

import (
	"context"
	"testing"

	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"github.com/google/uuid"
)

func TestSystemNotificationRoutesAcrossInstances(t *testing.T) {
	bus := NewInMemoryBus()
	defer bus.Close()

	hubA := NewSystemNotificationHub(WithSystemNotificationBus(bus, "node-a"))
	defer hubA.Close()
	hubB := NewSystemNotificationHub(WithSystemNotificationBus(bus, "node-b"))
	defer hubB.Close()

	recipientID := uuid.New()
	client := &systemNotificationClient{
		hub:         hubB,
		recipientID: recipientID,
		socket:      newTestSocket(4),
	}
	hubB.register(client)

	notificationID := uuid.New()
	hubA.PublishSystemNotificationChange(context.Background(), domainNotification.SystemNotificationChange{
		Type:           domainNotification.SystemNotificationChangeDeleted,
		RecipientID:    recipientID,
		NotificationID: notificationID,
		UnreadCount:    3,
	})

	message := receiveSocketMessageOfType(t, client.socket, string(domainNotification.SystemNotificationChangeDeleted))
	if message["notification_id"] != notificationID.String() {
		t.Fatalf("notification_id = %v, want %s", message["notification_id"], notificationID)
	}
	if message["unread_count"] != float64(3) {
		t.Fatalf("unread_count = %v, want 3", message["unread_count"])
	}
}

func TestSystemNotificationDoesNotDuplicateOwnBusEventToLocalClient(t *testing.T) {
	bus := NewInMemoryBus()
	defer bus.Close()

	hub := NewSystemNotificationHub(WithSystemNotificationBus(bus, "node-a"))
	defer hub.Close()

	recipientID := uuid.New()
	client := &systemNotificationClient{
		hub:         hub,
		recipientID: recipientID,
		socket:      newTestSocket(4),
	}
	hub.register(client)

	hub.PublishSystemNotificationDeleted(context.Background(), recipientID, uuid.New(), 0)

	_ = receiveSocketMessageOfType(t, client.socket, string(domainNotification.SystemNotificationChangeDeleted))
	assertNoSocketMessageOfType(t, client.socket, string(domainNotification.SystemNotificationChangeDeleted))
}
