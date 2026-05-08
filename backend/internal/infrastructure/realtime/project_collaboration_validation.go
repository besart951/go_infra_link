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
	projectCollaborationMaxClientDevices          = 100
	projectCollaborationMaxChangedFieldsPerDevice = 64
	projectCollaborationMaxFieldValuesPerDevice   = 64
	projectCollaborationMaxRefreshIDs             = 100
	projectCollaborationMaxFieldDeviceDeltas      = 100
	projectCollaborationMaxDeltaRootFields        = 20
	projectCollaborationMaxDeltaNestedFields      = 80
	projectCollaborationMaxDeltaNestedItems       = 250
	projectCollaborationMaxFieldNameBytes         = 180
	projectCollaborationMaxStringValueBytes       = 2048
	projectCollaborationMaxMessageTypeBytes       = 64
)

var (
	errProjectCollaborationInvalidMessage = errors.New("invalid project collaboration message")

	projectCollaborationAllowedBaseEditFields = map[string]struct{}{
		"bmk":                           {},
		"description":                   {},
		"text_fix":                      {},
		"apparat_nr":                    {},
		"apparat_id":                    {},
		"system_part_id":                {},
		"sps_controller_system_type_id": {},
		"specification_id":              {},
	}

	projectCollaborationAllowedSpecificationEditFields = map[string]struct{}{
		"specification_supplier":                       {},
		"specification_brand":                          {},
		"specification_type":                           {},
		"additional_info_motor_valve":                  {},
		"additional_info_size":                         {},
		"additional_information_installation_location": {},
		"electrical_connection_ph":                     {},
		"electrical_connection_acdc":                   {},
		"electrical_connection_amperage":               {},
		"electrical_connection_power":                  {},
		"electrical_connection_rotation":               {},
	}

	projectCollaborationAllowedBacnetEditFields = map[string]struct{}{
		"text_fix":              {},
		"description":           {},
		"gms_visible":           {},
		"optional":              {},
		"text_individual":       {},
		"software_type":         {},
		"software_number":       {},
		"hardware_type":         {},
		"hardware_quantity":     {},
		"software_reference_id": {},
		"state_text_id":         {},
		"notification_class_id": {},
		"alarm_type_id":         {},
	}

	projectCollaborationAllowedFieldDeviceDeltaFields = map[string]struct{}{
		"id":                            {},
		"bmk":                           {},
		"description":                   {},
		"text_fix":                      {},
		"apparat_nr":                    {},
		"sps_controller_system_type_id": {},
		"system_part_id":                {},
		"specification_id":              {},
		"apparat_id":                    {},
		"created_at":                    {},
		"updated_at":                    {},
		"sps_controller_system_type":    {},
		"apparat":                       {},
		"system_part":                   {},
		"specification":                 {},
		"bacnet_objects":                {},
	}
)

type projectCollaborationMessageEnvelopeDTO struct {
	Type string `json:"type"`
}

type projectCollaborationEditStateMessageDTO struct {
	Type    string                              `json:"type"`
	Devices []projectCollaborationDeviceEditDTO `json:"devices,omitempty"`
}

type projectCollaborationDeviceEditDTO struct {
	DeviceID      string         `json:"device_id"`
	ChangedFields []string       `json:"changed_fields"`
	FieldValues   map[string]any `json:"field_values,omitempty"`
}

type projectCollaborationRefreshRequestDTO struct {
	Type      string   `json:"type"`
	Scope     string   `json:"scope,omitempty"`
	EntityIDs []string `json:"entity_ids,omitempty"`
	DeviceIDs []string `json:"device_ids,omitempty"`
}

type projectCollaborationEntityDeltaDTO struct {
	Type         string           `json:"type"`
	Scope        string           `json:"scope"`
	FieldDevices []map[string]any `json:"field_devices,omitempty"`
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
	case projectCollaborationMessageEditState:
		var dto projectCollaborationEditStateMessageDTO
		if err := decodeProjectCollaborationJSON(data, &dto); err != nil {
			return projectCollaborationClientMessage{}, err
		}
		return validateProjectCollaborationEditStateDTO(dto)
	case projectCollaborationMessageEntityDelta:
		var dto projectCollaborationEntityDeltaDTO
		if err := decodeProjectCollaborationJSON(data, &dto); err != nil {
			return projectCollaborationClientMessage{}, err
		}
		return validateProjectCollaborationEntityDeltaDTO(dto)
	case projectCollaborationMessageRefreshRequest:
		var dto projectCollaborationRefreshRequestDTO
		if err := decodeProjectCollaborationJSON(data, &dto); err != nil {
			return projectCollaborationClientMessage{}, err
		}
		return validateProjectCollaborationRefreshRequestDTO(dto)
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

func validateProjectCollaborationEditStateDTO(dto projectCollaborationEditStateMessageDTO) (projectCollaborationClientMessage, error) {
	if strings.TrimSpace(dto.Type) != projectCollaborationMessageEditState {
		return projectCollaborationClientMessage{}, fmt.Errorf("%w: unsupported edit_state type", errProjectCollaborationInvalidMessage)
	}
	if len(dto.Devices) > projectCollaborationMaxClientDevices {
		return projectCollaborationClientMessage{}, fmt.Errorf("%w: too many edited devices", errProjectCollaborationInvalidMessage)
	}

	devices := make([]ProjectFieldDeviceByFields, 0, len(dto.Devices))
	for _, device := range dto.Devices {
		normalized, err := validateProjectCollaborationDeviceEditDTO(device)
		if err != nil {
			return projectCollaborationClientMessage{}, err
		}
		devices = append(devices, normalized)
	}

	return projectCollaborationClientMessage{
		Type:    projectCollaborationMessageEditState,
		Devices: devices,
	}, nil
}

func validateProjectCollaborationDeviceEditDTO(dto projectCollaborationDeviceEditDTO) (ProjectFieldDeviceByFields, error) {
	deviceID, err := validateProjectCollaborationUUIDString(dto.DeviceID, "device_id")
	if err != nil {
		return ProjectFieldDeviceByFields{}, err
	}
	if len(dto.ChangedFields) == 0 {
		return ProjectFieldDeviceByFields{}, fmt.Errorf("%w: changed_fields is required", errProjectCollaborationInvalidMessage)
	}
	if len(dto.ChangedFields) > projectCollaborationMaxChangedFieldsPerDevice {
		return ProjectFieldDeviceByFields{}, fmt.Errorf("%w: too many changed_fields", errProjectCollaborationInvalidMessage)
	}

	fieldSet := make(map[string]struct{}, len(dto.ChangedFields))
	changedFields := make([]string, 0, len(dto.ChangedFields))
	for _, field := range dto.ChangedFields {
		normalized, err := normalizeProjectCollaborationFieldName(field)
		if err != nil {
			return ProjectFieldDeviceByFields{}, err
		}
		if _, exists := fieldSet[normalized]; exists {
			continue
		}
		fieldSet[normalized] = struct{}{}
		changedFields = append(changedFields, normalized)
	}
	sort.Strings(changedFields)

	fieldValues, err := validateProjectCollaborationFieldValues(dto.FieldValues, fieldSet)
	if err != nil {
		return ProjectFieldDeviceByFields{}, err
	}

	return ProjectFieldDeviceByFields{
		DeviceID:      deviceID,
		ChangedFields: changedFields,
		FieldValues:   fieldValues,
	}, nil
}

func validateProjectCollaborationRefreshRequestDTO(dto projectCollaborationRefreshRequestDTO) (projectCollaborationClientMessage, error) {
	if strings.TrimSpace(dto.Type) != projectCollaborationMessageRefreshRequest {
		return projectCollaborationClientMessage{}, fmt.Errorf("%w: unsupported refresh_request type", errProjectCollaborationInvalidMessage)
	}

	scope, err := normalizeProjectCollaborationIncomingRefreshScope(dto.Scope)
	if err != nil {
		return projectCollaborationClientMessage{}, err
	}
	entityIDs, err := validateProjectCollaborationUUIDStrings(dto.EntityIDs, "entity_ids")
	if err != nil {
		return projectCollaborationClientMessage{}, err
	}
	deviceIDs, err := validateProjectCollaborationUUIDStrings(dto.DeviceIDs, "device_ids")
	if err != nil {
		return projectCollaborationClientMessage{}, err
	}

	return projectCollaborationClientMessage{
		Type:      projectCollaborationMessageRefreshRequest,
		Scope:     scope,
		EntityIDs: entityIDs,
		DeviceIDs: deviceIDs,
	}, nil
}

func validateProjectCollaborationEntityDeltaDTO(dto projectCollaborationEntityDeltaDTO) (projectCollaborationClientMessage, error) {
	if strings.TrimSpace(dto.Type) != projectCollaborationMessageEntityDelta {
		return projectCollaborationClientMessage{}, fmt.Errorf("%w: unsupported entity_delta type", errProjectCollaborationInvalidMessage)
	}

	scope := strings.TrimSpace(dto.Scope)
	if scope != projectCollaborationRefreshScopeFieldDevice {
		return projectCollaborationClientMessage{}, fmt.Errorf("%w: unsupported entity_delta scope", errProjectCollaborationInvalidMessage)
	}
	if len(dto.FieldDevices) == 0 {
		return projectCollaborationClientMessage{}, fmt.Errorf("%w: field_devices is required", errProjectCollaborationInvalidMessage)
	}
	if len(dto.FieldDevices) > projectCollaborationMaxFieldDeviceDeltas {
		return projectCollaborationClientMessage{}, fmt.Errorf("%w: too many field_devices", errProjectCollaborationInvalidMessage)
	}

	fieldDevices := make([]map[string]any, 0, len(dto.FieldDevices))
	for _, item := range dto.FieldDevices {
		sanitized, err := sanitizeProjectCollaborationFieldDeviceDelta(item)
		if err != nil {
			return projectCollaborationClientMessage{}, err
		}
		fieldDevices = append(fieldDevices, sanitized)
	}

	return projectCollaborationClientMessage{
		Type:         projectCollaborationMessageEntityDelta,
		Scope:        scope,
		FieldDevices: fieldDevices,
	}, nil
}

func validateProjectCollaborationFieldValues(values map[string]any, changedFields map[string]struct{}) (map[string]any, error) {
	if len(values) == 0 {
		return nil, nil
	}
	if len(values) > projectCollaborationMaxFieldValuesPerDevice {
		return nil, fmt.Errorf("%w: too many field_values", errProjectCollaborationInvalidMessage)
	}

	normalized := make(map[string]any, len(values))
	for key, value := range values {
		field, err := normalizeProjectCollaborationFieldName(key)
		if err != nil {
			return nil, err
		}
		if _, ok := changedFields[field]; !ok {
			return nil, fmt.Errorf("%w: field_values contains a key outside changed_fields", errProjectCollaborationInvalidMessage)
		}
		if err := validateProjectCollaborationScalarValue(value); err != nil {
			return nil, err
		}
		normalized[field] = value
	}
	return normalized, nil
}

func normalizeProjectCollaborationFieldName(field string) (string, error) {
	normalized := strings.TrimSpace(field)
	if normalized == "" {
		return "", fmt.Errorf("%w: empty field name", errProjectCollaborationInvalidMessage)
	}
	if len(normalized) > projectCollaborationMaxFieldNameBytes || !utf8.ValidString(normalized) {
		return "", fmt.Errorf("%w: invalid field name", errProjectCollaborationInvalidMessage)
	}
	if _, ok := projectCollaborationAllowedBaseEditFields[normalized]; ok {
		return normalized, nil
	}

	if specField, ok := strings.CutPrefix(normalized, "specification."); ok {
		if _, allowed := projectCollaborationAllowedSpecificationEditFields[specField]; allowed {
			return normalized, nil
		}
		return "", fmt.Errorf("%w: unsupported specification field", errProjectCollaborationInvalidMessage)
	}

	parts := strings.Split(normalized, ".")
	if len(parts) == 3 && parts[0] == "bacnet_objects" {
		if _, err := validateProjectCollaborationUUIDString(parts[1], "bacnet_object_id"); err != nil {
			return "", err
		}
		if _, allowed := projectCollaborationAllowedBacnetEditFields[parts[2]]; allowed {
			return normalized, nil
		}
		return "", fmt.Errorf("%w: unsupported BACnet field", errProjectCollaborationInvalidMessage)
	}

	return "", fmt.Errorf("%w: unsupported field name", errProjectCollaborationInvalidMessage)
}

func normalizeProjectCollaborationIncomingRefreshScope(scope string) (string, error) {
	normalized := strings.TrimSpace(scope)
	if normalized == "" {
		return projectCollaborationRefreshScopeFieldDevice, nil
	}

	switch normalized {
	case projectCollaborationRefreshScopeFieldDevice,
		projectCollaborationRefreshScopeControlCabinet,
		projectCollaborationRefreshScopeSPSController:
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: unsupported refresh scope", errProjectCollaborationInvalidMessage)
	}
}

func validateProjectCollaborationUUIDStrings(values []string, field string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	if len(values) > projectCollaborationMaxRefreshIDs {
		return nil, fmt.Errorf("%w: too many %s", errProjectCollaborationInvalidMessage, field)
	}

	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		parsed, err := validateProjectCollaborationUUIDString(value, field)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[parsed]; exists {
			continue
		}
		seen[parsed] = struct{}{}
		normalized = append(normalized, parsed)
	}
	sort.Strings(normalized)
	return normalized, nil
}

func validateProjectCollaborationUUIDString(value, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%w: %s is required", errProjectCollaborationInvalidMessage, field)
	}
	parsed, err := uuid.Parse(trimmed)
	if err != nil || parsed == uuid.Nil {
		return "", fmt.Errorf("%w: invalid %s", errProjectCollaborationInvalidMessage, field)
	}
	return parsed.String(), nil
}

func validateProjectCollaborationScalarValue(value any) error {
	switch typed := value.(type) {
	case nil, bool, float64:
		return nil
	case string:
		if len(typed) > projectCollaborationMaxStringValueBytes || !utf8.ValidString(typed) {
			return fmt.Errorf("%w: invalid field value", errProjectCollaborationInvalidMessage)
		}
		return nil
	default:
		return fmt.Errorf("%w: unsupported field value type", errProjectCollaborationInvalidMessage)
	}
}

func sanitizeProjectCollaborationFieldDeviceDelta(item map[string]any) (map[string]any, error) {
	if len(item) == 0 {
		return nil, fmt.Errorf("%w: empty field_device delta", errProjectCollaborationInvalidMessage)
	}
	if len(item) > projectCollaborationMaxDeltaRootFields {
		return nil, fmt.Errorf("%w: too many field_device fields", errProjectCollaborationInvalidMessage)
	}

	sanitized := make(map[string]any, len(item))
	for key, value := range item {
		field := strings.TrimSpace(key)
		if _, ok := projectCollaborationAllowedFieldDeviceDeltaFields[field]; !ok {
			return nil, fmt.Errorf("%w: unsupported field_device field", errProjectCollaborationInvalidMessage)
		}
		if field == "id" {
			id, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("%w: invalid field_device id", errProjectCollaborationInvalidMessage)
			}
			normalized, err := validateProjectCollaborationUUIDString(id, "field_device.id")
			if err != nil {
				return nil, err
			}
			sanitized[field] = normalized
			continue
		}
		if isProjectCollaborationFieldDeviceUUIDField(field) {
			if value == nil {
				sanitized[field] = nil
				continue
			}
			id, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("%w: invalid field_device UUID field", errProjectCollaborationInvalidMessage)
			}
			normalized, err := validateProjectCollaborationUUIDString(id, field)
			if err != nil {
				return nil, err
			}
			sanitized[field] = normalized
			continue
		}

		safeValue, err := sanitizeProjectCollaborationDeltaValue(value, 0)
		if err != nil {
			return nil, err
		}
		sanitized[field] = safeValue
	}

	if _, ok := sanitized["id"]; !ok {
		return nil, fmt.Errorf("%w: field_device id is required", errProjectCollaborationInvalidMessage)
	}
	return sanitized, nil
}

func isProjectCollaborationFieldDeviceUUIDField(field string) bool {
	switch field {
	case "sps_controller_system_type_id", "system_part_id", "specification_id", "apparat_id":
		return true
	default:
		return false
	}
}

func sanitizeProjectCollaborationDeltaValue(value any, depth int) (any, error) {
	if depth > 2 {
		return nil, fmt.Errorf("%w: nested delta value too deep", errProjectCollaborationInvalidMessage)
	}

	switch typed := value.(type) {
	case nil, bool, float64:
		return typed, nil
	case string:
		if len(typed) > projectCollaborationMaxStringValueBytes || !utf8.ValidString(typed) {
			return nil, fmt.Errorf("%w: invalid delta string value", errProjectCollaborationInvalidMessage)
		}
		return typed, nil
	case []any:
		if len(typed) > projectCollaborationMaxDeltaNestedItems {
			return nil, fmt.Errorf("%w: too many nested delta items", errProjectCollaborationInvalidMessage)
		}
		values := make([]any, 0, len(typed))
		for _, item := range typed {
			sanitized, err := sanitizeProjectCollaborationDeltaValue(item, depth+1)
			if err != nil {
				return nil, err
			}
			values = append(values, sanitized)
		}
		return values, nil
	case map[string]any:
		if len(typed) > projectCollaborationMaxDeltaNestedFields {
			return nil, fmt.Errorf("%w: too many nested delta fields", errProjectCollaborationInvalidMessage)
		}
		values := make(map[string]any, len(typed))
		for key, item := range typed {
			field := strings.TrimSpace(key)
			if field == "" || len(field) > projectCollaborationMaxFieldNameBytes || !utf8.ValidString(field) {
				return nil, fmt.Errorf("%w: invalid nested delta field", errProjectCollaborationInvalidMessage)
			}
			sanitized, err := sanitizeProjectCollaborationDeltaValue(item, depth+1)
			if err != nil {
				return nil, err
			}
			values[field] = sanitized
		}
		return values, nil
	default:
		return nil, fmt.Errorf("%w: unsupported delta value type", errProjectCollaborationInvalidMessage)
	}
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
