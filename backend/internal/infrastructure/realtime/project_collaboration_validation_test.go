package realtime

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestParseProjectCollaborationDraftStateValidatesAllAggregatePaths(t *testing.T) {
	tests := []struct {
		aggregate string
		path      string
	}{
		{"project", "name"},
		{"control_cabinet", "control_cabinet_nr"},
		{"sps_controller", "device_name"},
		{"sps_controller_system_type", "number"},
		{"field_device", "apparat_nr"},
		{"specification", "specification_brand"},
		{"bacnet_object", "software_type"},
		{"alarm", "values.high_limit"},
	}

	for _, tt := range tests {
		t.Run(tt.aggregate+"/"+tt.path, func(t *testing.T) {
			aggregateID := uuid.New().String()
			parsed := mustParseProjectCollaborationMessage(t, map[string]any{
				"type": "draft_state",
				"entries": []map[string]any{{
					"aggregate_type": tt.aggregate,
					"aggregate_id":   aggregateID,
					"action":         "update",
					"base_version":   7,
					"fields":         []map[string]any{{"path": tt.path, "value": "draft"}},
				}},
			})

			if parsed.Type != projectCollaborationMessageDraftState || len(parsed.Entries) != 1 {
				t.Fatalf("parsed = %+v", parsed)
			}
			entry := parsed.Entries[0]
			if entry.AggregateType != tt.aggregate || entry.AggregateID != aggregateID || entry.Fields[0].Path != tt.path {
				t.Fatalf("entry = %+v", entry)
			}
		})
	}
}

func TestParseProjectCollaborationCreateDraftRequiresDraftID(t *testing.T) {
	assertInvalidProjectCollaborationMessage(t, map[string]any{
		"type": "draft_state",
		"entries": []map[string]any{{
			"aggregate_type": "field_device",
			"action":         "create",
			"base_version":   0,
			"fields":         []map[string]any{{"path": "bmk", "value": "draft"}},
		}},
	})
}

func TestParseProjectCollaborationDraftRejectsCrossAggregatePath(t *testing.T) {
	assertInvalidProjectCollaborationMessage(t, map[string]any{
		"type": "draft_state",
		"entries": []map[string]any{{
			"aggregate_type": "project",
			"aggregate_id":   uuid.NewString(),
			"action":         "update",
			"base_version":   1,
			"fields":         []map[string]any{{"path": "software_type", "value": "ai"}},
		}},
	})
}

func TestParseProjectCollaborationDraftClearValidatesSelector(t *testing.T) {
	id := uuid.NewString()
	parsed := mustParseProjectCollaborationMessage(t, map[string]any{
		"type":           "draft_clear",
		"aggregate_type": "bacnet_object",
		"aggregate_id":   id,
	})
	if parsed.Clear == nil || parsed.Clear.AggregateType != "bacnet_object" || parsed.Clear.AggregateID != id {
		t.Fatalf("clear = %+v", parsed.Clear)
	}
}

func TestParseProjectCollaborationRejectsBrowserAuthoredCommittedMessages(t *testing.T) {
	for _, messageType := range []string{"entity_delta", "refresh_request", "project_change", "revision"} {
		t.Run(messageType, func(t *testing.T) {
			assertInvalidProjectCollaborationMessage(t, map[string]any{"type": messageType})
		})
	}
}

func TestParseProjectCollaborationRejectsUnknownJSONFields(t *testing.T) {
	assertInvalidProjectCollaborationMessage(t, map[string]any{
		"type": "draft_clear", "aggregate_type": "project", "aggregate_id": uuid.NewString(), "admin": true,
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
