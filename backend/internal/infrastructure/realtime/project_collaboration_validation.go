package realtime

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	projectCollaborationMaxDraftEntries     = 100
	projectCollaborationMaxDraftFields      = 100
	projectCollaborationMaxFieldNameBytes   = 180
	projectCollaborationMaxStringValueBytes = 2048
	projectCollaborationMaxMessageTypeBytes = 64
)

var (
	errProjectCollaborationInvalidMessage = errors.New("invalid project collaboration message")

	projectCollaborationDraftFields = map[string]map[string]struct{}{
		"project":         fieldSet("name", "description", "status", "phase_id", "start_date"),
		"control_cabinet": fieldSet("building_id", "control_cabinet_nr"),
		"sps_controller": fieldSet(
			"control_cabinet_id", "ga_device", "device_name", "device_description", "device_location",
			"ip_address", "subnet", "gateway", "vlan",
		),
		"sps_controller_system_type": fieldSet("sps_controller_id", "system_type_id", "number", "document_name"),
		"field_device": fieldSet(
			"bmk", "description", "apparat_nr", "text_individuell", "sps_controller_system_type_id",
			"system_part_id", "specification_id", "apparat_id",
		),
		"specification": fieldSet(
			"specification_supplier", "specification_brand", "specification_type", "additional_info_motor_valve",
			"additional_info_size", "additional_information_installation_location", "electrical_connection_ph",
			"electrical_connection_acdc", "electrical_connection_amperage", "electrical_connection_power",
			"electrical_connection_rotation",
		),
		"bacnet_object": fieldSet(
			"text_fix", "description", "gms_visible", "optional", "text_individual", "software_type",
			"software_number", "hardware_type", "hardware_quantity", "field_device_id", "software_reference_id",
			"state_text_id", "notification_class_id", "alarm_type_id", "alarm_definition_id",
		),
		"alarm": fieldSet(
			"alarm_type_id", "alarm_definition_id", "notification_class_id", "state_text_id", "enabled",
			"priority", "limit", "deadband", "delay", "message", "value", "unit_id",
		),
		"alarm_definition": fieldSet("alarm_type_id", "name", "description", "enabled", "delay", "deadband"),
		"alarm_value":      fieldSet("alarm_definition_id", "alarm_type_field_id", "value", "unit_id"),
		"alarm_type":       fieldSet("name", "description", "validation_json"),
	}
)

type projectCollaborationMessageEnvelopeDTO struct {
	Type string `json:"type"`
}

type projectCollaborationDraftStateDTO struct {
	Type    string                         `json:"type"`
	Entries []projectCollaborationDraftDTO `json:"entries"`
}

type projectCollaborationDraftClearDTO struct {
	Type          string  `json:"type"`
	AggregateType string  `json:"aggregate_type"`
	AggregateID   *string `json:"aggregate_id,omitempty"`
	DraftID       *string `json:"draft_id,omitempty"`
}

type projectCollaborationDraftDTO struct {
	AggregateType string                           `json:"aggregate_type"`
	AggregateID   *string                          `json:"aggregate_id,omitempty"`
	DraftID       *string                          `json:"draft_id,omitempty"`
	Action        string                           `json:"action"`
	BaseVersion   int64                            `json:"base_version"`
	Fields        []projectCollaborationDraftField `json:"fields"`
}

type projectCollaborationDraftField struct {
	Path  string `json:"path"`
	Value any    `json:"value"`
}

func parseProjectCollaborationClientMessage(data []byte) (projectCollaborationClientMessage, error) {
	if len(data) == 0 {
		return projectCollaborationClientMessage{}, fmt.Errorf("%w: empty payload", errProjectCollaborationInvalidMessage)
	}
	if int64(len(data)) > projectCollaborationMaxMessage {
		return projectCollaborationClientMessage{}, fmt.Errorf("%w: payload exceeds %d bytes", errProjectCollaborationInvalidMessage, projectCollaborationMaxMessage)
	}

	var envelope projectCollaborationMessageEnvelopeDTO
	if err := json.Unmarshal(data, &envelope); err != nil {
		return projectCollaborationClientMessage{}, fmt.Errorf("%w: malformed JSON", errProjectCollaborationInvalidMessage)
	}

	switch strings.TrimSpace(envelope.Type) {
	case projectCollaborationMessageDraftState:
		var dto projectCollaborationDraftStateDTO
		if err := decodeProjectCollaborationJSON(data, &dto); err != nil {
			return projectCollaborationClientMessage{}, err
		}
		return validateProjectCollaborationDraftState(dto)
	case projectCollaborationMessageDraftClear:
		var dto projectCollaborationDraftClearDTO
		if err := decodeProjectCollaborationJSON(data, &dto); err != nil {
			return projectCollaborationClientMessage{}, err
		}
		selector, err := validateProjectCollaborationDraftSelector(dto.AggregateType, dto.AggregateID, dto.DraftID)
		if err != nil {
			return projectCollaborationClientMessage{}, err
		}
		return projectCollaborationClientMessage{Type: projectCollaborationMessageDraftClear, Clear: &selector}, nil
	default:
		return projectCollaborationClientMessage{}, fmt.Errorf("%w: unsupported message type", errProjectCollaborationInvalidMessage)
	}
}

func decodeProjectCollaborationJSON(data []byte, dst any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("%w: invalid JSON shape", errProjectCollaborationInvalidMessage)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("%w: trailing JSON content", errProjectCollaborationInvalidMessage)
	}
	return nil
}

func validateProjectCollaborationDraftState(dto projectCollaborationDraftStateDTO) (projectCollaborationClientMessage, error) {
	if strings.TrimSpace(dto.Type) != projectCollaborationMessageDraftState {
		return projectCollaborationClientMessage{}, fmt.Errorf("%w: unsupported draft_state type", errProjectCollaborationInvalidMessage)
	}
	if len(dto.Entries) == 0 {
		return projectCollaborationClientMessage{}, fmt.Errorf("%w: entries is required", errProjectCollaborationInvalidMessage)
	}
	if len(dto.Entries) > projectCollaborationMaxDraftEntries {
		return projectCollaborationClientMessage{}, fmt.Errorf("%w: too many draft entries", errProjectCollaborationInvalidMessage)
	}

	entries := make([]ProjectDraftEntry, 0, len(dto.Entries))
	seen := make(map[string]struct{}, len(dto.Entries))
	for _, entry := range dto.Entries {
		normalized, err := validateProjectCollaborationDraftEntry(entry)
		if err != nil {
			return projectCollaborationClientMessage{}, err
		}
		key := normalized.selectorKey()
		if _, exists := seen[key]; exists {
			return projectCollaborationClientMessage{}, fmt.Errorf("%w: duplicate draft entry", errProjectCollaborationInvalidMessage)
		}
		seen[key] = struct{}{}
		entries = append(entries, normalized)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].selectorKey() < entries[j].selectorKey() })
	return projectCollaborationClientMessage{Type: projectCollaborationMessageDraftState, Entries: entries}, nil
}

func validateProjectCollaborationDraftEntry(dto projectCollaborationDraftDTO) (ProjectDraftEntry, error) {
	selector, err := validateProjectCollaborationDraftSelector(dto.AggregateType, dto.AggregateID, dto.DraftID)
	if err != nil {
		return ProjectDraftEntry{}, err
	}
	action := strings.TrimSpace(dto.Action)
	if action != "create" && action != "update" {
		return ProjectDraftEntry{}, fmt.Errorf("%w: action must be create or update", errProjectCollaborationInvalidMessage)
	}
	if action == "create" && selector.DraftID == "" {
		return ProjectDraftEntry{}, fmt.Errorf("%w: create drafts require draft_id", errProjectCollaborationInvalidMessage)
	}
	if action == "update" && selector.AggregateID == "" {
		return ProjectDraftEntry{}, fmt.Errorf("%w: update drafts require aggregate_id", errProjectCollaborationInvalidMessage)
	}
	if dto.BaseVersion < 0 {
		return ProjectDraftEntry{}, fmt.Errorf("%w: base_version cannot be negative", errProjectCollaborationInvalidMessage)
	}
	if len(dto.Fields) == 0 || len(dto.Fields) > projectCollaborationMaxDraftFields {
		return ProjectDraftEntry{}, fmt.Errorf("%w: invalid draft fields count", errProjectCollaborationInvalidMessage)
	}

	fields := make([]ProjectDraftField, 0, len(dto.Fields))
	seen := make(map[string]struct{}, len(dto.Fields))
	for _, field := range dto.Fields {
		path, err := validateProjectCollaborationDraftPath(selector.AggregateType, field.Path)
		if err != nil {
			return ProjectDraftEntry{}, err
		}
		if _, exists := seen[path]; exists {
			return ProjectDraftEntry{}, fmt.Errorf("%w: duplicate draft field", errProjectCollaborationInvalidMessage)
		}
		if err := validateProjectCollaborationDraftValue(field.Value, 0); err != nil {
			return ProjectDraftEntry{}, err
		}
		seen[path] = struct{}{}
		fields = append(fields, ProjectDraftField{Path: path, Value: field.Value})
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].Path < fields[j].Path })

	return ProjectDraftEntry{
		ProjectDraftSelector: selector,
		Action:               action,
		BaseVersion:          dto.BaseVersion,
		Fields:               fields,
	}, nil
}

func validateProjectCollaborationDraftSelector(aggregateType string, aggregateID, draftID *string) (ProjectDraftSelector, error) {
	normalizedType := strings.TrimSpace(aggregateType)
	if _, ok := projectCollaborationDraftFields[normalizedType]; !ok {
		return ProjectDraftSelector{}, fmt.Errorf("%w: unsupported aggregate_type", errProjectCollaborationInvalidMessage)
	}

	selector := ProjectDraftSelector{AggregateType: normalizedType}
	if aggregateID != nil {
		parsed, err := uuid.Parse(strings.TrimSpace(*aggregateID))
		if err != nil || parsed == uuid.Nil {
			return ProjectDraftSelector{}, fmt.Errorf("%w: invalid aggregate_id", errProjectCollaborationInvalidMessage)
		}
		selector.AggregateID = parsed.String()
	}
	if draftID != nil {
		parsed, err := uuid.Parse(strings.TrimSpace(*draftID))
		if err != nil || parsed == uuid.Nil {
			return ProjectDraftSelector{}, fmt.Errorf("%w: invalid draft_id", errProjectCollaborationInvalidMessage)
		}
		selector.DraftID = parsed.String()
	}
	if (selector.AggregateID == "") == (selector.DraftID == "") {
		return ProjectDraftSelector{}, fmt.Errorf("%w: exactly one of aggregate_id or draft_id is required", errProjectCollaborationInvalidMessage)
	}
	return selector, nil
}

func validateProjectCollaborationDraftPath(aggregateType, path string) (string, error) {
	normalized := strings.TrimSpace(path)
	if normalized == "" || len(normalized) > projectCollaborationMaxFieldNameBytes || !utf8.ValidString(normalized) {
		return "", fmt.Errorf("%w: invalid draft field path", errProjectCollaborationInvalidMessage)
	}
	if fields := projectCollaborationDraftFields[aggregateType]; fields != nil {
		if _, ok := fields[normalized]; ok {
			return normalized, nil
		}
	}
	if aggregateType == "alarm" && strings.HasPrefix(normalized, "values.") {
		segment := strings.TrimPrefix(normalized, "values.")
		if segment != "" && !strings.ContainsAny(segment, " .[]") {
			return normalized, nil
		}
	}
	if aggregateType == "bacnet_object" && strings.HasPrefix(normalized, "alarm.") {
		segment := strings.TrimPrefix(normalized, "alarm.")
		if segment != "" && !strings.ContainsAny(segment, " []") {
			return normalized, nil
		}
	}
	return "", fmt.Errorf("%w: unsupported %s draft field", errProjectCollaborationInvalidMessage, aggregateType)
}

func validateProjectCollaborationDraftValue(value any, depth int) error {
	if depth > 2 {
		return fmt.Errorf("%w: draft value nested too deeply", errProjectCollaborationInvalidMessage)
	}
	switch typed := value.(type) {
	case nil, bool, float64:
		return nil
	case string:
		if len(typed) > projectCollaborationMaxStringValueBytes || !utf8.ValidString(typed) {
			return fmt.Errorf("%w: invalid draft string value", errProjectCollaborationInvalidMessage)
		}
		return nil
	case []any:
		if len(typed) > projectCollaborationMaxDraftFields {
			return fmt.Errorf("%w: too many draft array values", errProjectCollaborationInvalidMessage)
		}
		for _, item := range typed {
			if err := validateProjectCollaborationDraftValue(item, depth+1); err != nil {
				return err
			}
		}
		return nil
	case map[string]any:
		if len(typed) > projectCollaborationMaxDraftFields {
			return fmt.Errorf("%w: too many draft object values", errProjectCollaborationInvalidMessage)
		}
		for key, item := range typed {
			if strings.TrimSpace(key) == "" || len(key) > projectCollaborationMaxFieldNameBytes || !utf8.ValidString(key) {
				return fmt.Errorf("%w: invalid draft object key", errProjectCollaborationInvalidMessage)
			}
			if err := validateProjectCollaborationDraftValue(item, depth+1); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("%w: unsupported draft value type", errProjectCollaborationInvalidMessage)
	}
}

func fieldSet(fields ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		set[field] = struct{}{}
	}
	return set
}

func logInvalidProjectCollaborationMessage(data []byte, err error) {
	slog.Warn(
		"ignored invalid project collaboration websocket message",
		"reason", err.Error(),
		"type", safeProjectCollaborationMessageType(data),
		"bytes", len(data),
	)
}

func safeProjectCollaborationMessageType(data []byte) string {
	var envelope projectCollaborationMessageEnvelopeDTO
	if err := json.Unmarshal(data, &envelope); err != nil {
		return ""
	}
	messageType := strings.TrimSpace(envelope.Type)
	if len(messageType) > projectCollaborationMaxMessageTypeBytes || !utf8.ValidString(messageType) {
		return ""
	}
	return messageType
}
