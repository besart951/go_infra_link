package realtime

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestParseProjectCollaborationEditStateMessageValidatesAndNormalizes(t *testing.T) {
	deviceID := uuid.New().String()
	bacnetObjectID := uuid.New().String()
	message := map[string]any{
		"type": "edit_state",
		"devices": []map[string]any{
			{
				"device_id": deviceID,
				"changed_fields": []string{
					" text_fix ",
					"specification.specification_brand",
					"bacnet_objects." + bacnetObjectID + ".software_type",
				},
				"field_values": map[string]any{
					"text_fix":                          "TF-1",
					"specification.specification_brand": "Brand",
					"bacnet_objects." + bacnetObjectID + ".software_type": "ai",
				},
			},
		},
	}

	parsed := mustParseProjectCollaborationMessage(t, message)

	if parsed.Type != projectCollaborationMessageEditState {
		t.Fatalf("expected edit_state type, got %q", parsed.Type)
	}
	if len(parsed.Devices) != 1 {
		t.Fatalf("expected one device, got %+v", parsed.Devices)
	}
	if parsed.Devices[0].DeviceID != deviceID {
		t.Fatalf("expected normalized device id, got %q", parsed.Devices[0].DeviceID)
	}
	if !containsString(parsed.Devices[0].ChangedFields, "text_fix") {
		t.Fatalf("expected trimmed text_fix field, got %+v", parsed.Devices[0].ChangedFields)
	}
	if parsed.Devices[0].FieldValues["text_fix"] != "TF-1" {
		t.Fatalf("expected field value to survive validation, got %+v", parsed.Devices[0].FieldValues)
	}
}

func TestParseProjectCollaborationMessageRejectsInvalidType(t *testing.T) {
	assertInvalidProjectCollaborationMessage(t, map[string]any{"type": "delete_everything"})
}

func TestParseProjectCollaborationEditStateRejectsInvalidUUID(t *testing.T) {
	assertInvalidProjectCollaborationMessage(t, map[string]any{
		"type": "edit_state",
		"devices": []map[string]any{
			{
				"device_id":      "not-a-uuid",
				"changed_fields": []string{"text_fix"},
			},
		},
	})
}

func TestParseProjectCollaborationEditStateRejectsUnknownFieldName(t *testing.T) {
	assertInvalidProjectCollaborationMessage(t, map[string]any{
		"type": "edit_state",
		"devices": []map[string]any{
			{
				"device_id":      uuid.New().String(),
				"changed_fields": []string{"admin_role"},
			},
		},
	})
}

func TestParseProjectCollaborationEditStateRejectsExcessiveFieldValues(t *testing.T) {
	fieldValues := make(map[string]any, projectCollaborationMaxFieldValuesPerDevice+1)
	changedFields := make([]string, 0, projectCollaborationMaxFieldValuesPerDevice+1)
	for range projectCollaborationMaxFieldValuesPerDevice + 1 {
		field := "bacnet_objects." + uuid.New().String() + ".text_fix"
		changedFields = append(changedFields, field)
		fieldValues[field] = strings.Repeat("x", 4)
	}

	assertInvalidProjectCollaborationMessage(t, map[string]any{
		"type": "edit_state",
		"devices": []map[string]any{
			{
				"device_id":      uuid.New().String(),
				"changed_fields": changedFields,
				"field_values":   fieldValues,
			},
		},
	})
}

func TestParseProjectCollaborationRefreshRequestValidatesScopeAndIDs(t *testing.T) {
	id := uuid.New().String()
	parsed := mustParseProjectCollaborationMessage(t, map[string]any{
		"type":       "refresh_request",
		"scope":      "field_device",
		"device_ids": []string{id},
	})

	if parsed.Scope != projectCollaborationRefreshScopeFieldDevice {
		t.Fatalf("expected field_device scope, got %q", parsed.Scope)
	}
	if len(parsed.DeviceIDs) != 1 || parsed.DeviceIDs[0] != id {
		t.Fatalf("expected validated device id, got %+v", parsed.DeviceIDs)
	}
}

func TestParseProjectCollaborationRefreshRequestRejectsProjectScopeFromClient(t *testing.T) {
	assertInvalidProjectCollaborationMessage(t, map[string]any{
		"type":  "refresh_request",
		"scope": "project_users",
	})
}

func TestParseProjectCollaborationEntityDeltaSanitizesFieldDeviceDeltas(t *testing.T) {
	fieldDeviceID := uuid.New().String()
	apparatID := uuid.New().String()
	parsed := mustParseProjectCollaborationMessage(t, map[string]any{
		"type":  "entity_delta",
		"scope": "field_device",
		"field_devices": []map[string]any{
			{
				"id":         fieldDeviceID,
				"apparat_id": apparatID,
				"text_fix":   "TF-2",
			},
		},
	})

	if parsed.Scope != projectCollaborationRefreshScopeFieldDevice {
		t.Fatalf("expected field_device scope, got %q", parsed.Scope)
	}
	if len(parsed.FieldDevices) != 1 {
		t.Fatalf("expected one field device delta, got %+v", parsed.FieldDevices)
	}
	if parsed.FieldDevices[0]["id"] != fieldDeviceID {
		t.Fatalf("expected sanitized id, got %+v", parsed.FieldDevices[0])
	}
}

func TestParseProjectCollaborationEntityDeltaRejectsUnknownFieldDeviceField(t *testing.T) {
	assertInvalidProjectCollaborationMessage(t, map[string]any{
		"type":  "entity_delta",
		"scope": "field_device",
		"field_devices": []map[string]any{
			{
				"id":         uuid.New().String(),
				"admin_note": "should not pass",
			},
		},
	})
}

func mustParseProjectCollaborationMessage(t *testing.T, message map[string]any) projectCollaborationClientMessage {
	t.Helper()
	data, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}
	parsed, err := parseProjectCollaborationClientMessage(data)
	if err != nil {
		t.Fatalf("parse message: %v", err)
	}
	return parsed
}

func assertInvalidProjectCollaborationMessage(t *testing.T, message map[string]any) {
	t.Helper()
	data, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}
	if _, err := parseProjectCollaborationClientMessage(data); err == nil {
		t.Fatalf("expected message to be rejected")
	}
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
