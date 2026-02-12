package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

// WebSocket Message Actions
const (
	ActionBulkUpdateElements = "BULK_UPDATE_ELEMENTS"
	ActionBulkDeleteElements = "BULK_DELETE_ELEMENTS"
	ActionBulkCreateElements = "BULK_CREATE_ELEMENTS"
	ActionPresenceJoin       = "PRESENCE_JOIN"
	ActionPresenceLeave      = "PRESENCE_LEAVE"
	ActionPresenceList       = "PRESENCE_LIST"
)

// Entity Types for Bulk Operations
const (
	EntityTypeFieldDevice                = "field_device"
	EntityTypeControlCabinet             = "control_cabinet"
	EntityTypeSPSController              = "sps_controller"
	EntityTypeSPSControllerSystemType    = "sps_controller_system_type"
	EntityTypeBacnetObject               = "bacnet_object"
	EntityTypeSpecification              = "specification"
)

// WSMessage represents the top-level WebSocket message structure
type WSMessage struct {
	Action    string          `json:"action"`
	Payload   json.RawMessage `json:"payload"`
	ProjectID uuid.UUID       `json:"project_id"`
}

// WSPresenceUser represents a user's presence information
type WSPresenceUser struct {
	UserID    uuid.UUID `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}

// WSPresenceListPayload contains the list of users in a room
type WSPresenceListPayload struct {
	Users []WSPresenceUser `json:"users"`
}

// WSPresenceUserPayload contains a single user (for join events)
type WSPresenceUserPayload struct {
	User WSPresenceUser `json:"user"`
}

// WSPresenceLeavePayload contains the user ID of a leaving user
type WSPresenceLeavePayload struct {
	UserID uuid.UUID `json:"user_id"`
}

// WSBulkUpdatePayload represents a bulk update operation
type WSBulkUpdatePayload struct {
	EntityType string                   `json:"entity_type"`
	Updates    []map[string]interface{} `json:"updates"`
}

// WSBulkDeletePayload represents a bulk delete operation
type WSBulkDeletePayload struct {
	EntityType string      `json:"entity_type"`
	IDs        []uuid.UUID `json:"ids"`
}

// WSBulkCreatePayload represents a bulk create operation
type WSBulkCreatePayload struct {
	EntityType string                   `json:"entity_type"`
	Items      []map[string]interface{} `json:"items"`
}

// WSBulkOperationResult represents the result of a bulk operation
type WSBulkOperationResult struct {
	EntityType   string                    `json:"entity_type"`
	Action       string                    `json:"action"`
	TotalCount   int                       `json:"total_count"`
	SuccessCount int                       `json:"success_count"`
	FailureCount int                       `json:"failure_count"`
	Results      []WSBulkOperationItem     `json:"results"`
}

// WSBulkOperationItem represents a single item result in a bulk operation
type WSBulkOperationItem struct {
	ID      uuid.UUID         `json:"id"`
	Index   *int              `json:"index,omitempty"`   // For create operations
	Success bool              `json:"success"`
	Error   string            `json:"error,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
	Data    interface{}       `json:"data,omitempty"`    // The created/updated entity
}

// WSErrorResponse represents an error message sent via WebSocket
type WSErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
