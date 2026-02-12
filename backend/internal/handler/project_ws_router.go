package handler

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
)

// WSMessageRouter handles routing of WebSocket messages to appropriate handlers
type WSMessageRouter struct {
	hub *Hub
}

// NewWSMessageRouter creates a new WebSocket message router
func NewWSMessageRouter(hub *Hub) *WSMessageRouter {
	return &WSMessageRouter{
		hub: hub,
	}
}

// HandleMessage processes incoming WebSocket messages from clients
func (r *WSMessageRouter) HandleMessage(client *Client, message []byte) error {
	var wsMsg dto.WSMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	// Verify project ID matches client's room
	if wsMsg.ProjectID != client.projectID {
		return fmt.Errorf("project_id mismatch: expected %s, got %s", client.projectID, wsMsg.ProjectID)
	}

	log.Printf("Handling WebSocket message: action=%s, user=%s, project=%s",
		wsMsg.Action, client.userID, client.projectID)

	switch wsMsg.Action {
	case dto.ActionBulkUpdateElements:
		return r.handleBulkUpdate(client, wsMsg.Payload)
	case dto.ActionBulkDeleteElements:
		return r.handleBulkDelete(client, wsMsg.Payload)
	case dto.ActionBulkCreateElements:
		return r.handleBulkCreate(client, wsMsg.Payload)
	default:
		return fmt.Errorf("unknown action: %s", wsMsg.Action)
	}
}

// handleBulkUpdate processes bulk update operations
func (r *WSMessageRouter) handleBulkUpdate(client *Client, payload json.RawMessage) error {
	var bulkPayload dto.WSBulkUpdatePayload
	if err := json.Unmarshal(payload, &bulkPayload); err != nil {
		return fmt.Errorf("invalid bulk update payload: %w", err)
	}

	log.Printf("Bulk update: entity_type=%s, count=%d", bulkPayload.EntityType, len(bulkPayload.Updates))

	// TODO: Route to appropriate service based on entity_type
	// This will be implemented when service bulk methods are added

	// For now, just log and return success
	return fmt.Errorf("bulk update not yet implemented for entity type: %s", bulkPayload.EntityType)
}

// handleBulkDelete processes bulk delete operations
func (r *WSMessageRouter) handleBulkDelete(client *Client, payload json.RawMessage) error {
	var bulkPayload dto.WSBulkDeletePayload
	if err := json.Unmarshal(payload, &bulkPayload); err != nil {
		return fmt.Errorf("invalid bulk delete payload: %w", err)
	}

	log.Printf("Bulk delete: entity_type=%s, count=%d", bulkPayload.EntityType, len(bulkPayload.IDs))

	// TODO: Route to appropriate service based on entity_type
	// This will be implemented when service bulk methods are added

	// For now, just log and return success
	return fmt.Errorf("bulk delete not yet implemented for entity type: %s", bulkPayload.EntityType)
}

// handleBulkCreate processes bulk create operations
func (r *WSMessageRouter) handleBulkCreate(client *Client, payload json.RawMessage) error {
	var bulkPayload dto.WSBulkCreatePayload
	if err := json.Unmarshal(payload, &bulkPayload); err != nil {
		return fmt.Errorf("invalid bulk create payload: %w", err)
	}

	log.Printf("Bulk create: entity_type=%s, count=%d", bulkPayload.EntityType, len(bulkPayload.Items))

	// TODO: Route to appropriate service based on entity_type
	// This will be implemented when service bulk methods are added

	// For now, just log and return success
	return fmt.Errorf("bulk create not yet implemented for entity type: %s", bulkPayload.EntityType)
}
