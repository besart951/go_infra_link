package facility

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/handler/facility/internal/routing"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type facilityMutation struct {
	resource       string
	action         string
	pathIDIsTarget bool
}

func facilityMutationForRoute(route routing.Definition) (facilityMutation, bool) {
	if route.Method != http.MethodPost && route.Method != http.MethodPut && route.Method != http.MethodPatch && route.Method != http.MethodDelete {
		return facilityMutation{}, false
	}

	resource := facilityRealtimeResource(route.Path)
	if resource == "" || strings.HasSuffix(route.Path, "/bulk") || strings.HasSuffix(route.Path, "/validate") || strings.Contains(route.Path, "/export") {
		return facilityMutation{}, false
	}
	action := map[string]string{
		http.MethodPost:   "created",
		http.MethodPut:    "updated",
		http.MethodPatch:  "updated",
		http.MethodDelete: "deleted",
	}[route.Method]
	if strings.HasSuffix(route.Path, "/copy") {
		action = "copied"
	}
	if strings.Contains(route.Path, "/bulk-") || strings.Contains(route.Path, "/multi-create") || strings.HasSuffix(route.Path, "/multi") {
		action = "bulk_" + action
	}
	// POST /alarm-types/:id/fields uses the alarm-type ID, not the ID of the
	// newly created mapping. Do not publish that parent ID as an
	// alarm_type_fields ID; consumers can still refresh the affected list from
	// the bundled event.
	pathIDIsTarget := !(route.Method == http.MethodPost && strings.HasPrefix(route.Path, "/alarm-types/:id/fields"))
	return facilityMutation{resource: resource, action: action, pathIDIsTarget: pathIDIsTarget}, true
}

func facilityRealtimeResource(path string) string {
	if strings.HasPrefix(path, "/alarm-types/:id/fields") || strings.HasPrefix(path, "/alarm-type-fields/") {
		return "alarm_type_fields"
	}
	segment := strings.TrimPrefix(path, "/")
	if slash := strings.IndexByte(segment, '/'); slash >= 0 {
		segment = segment[:slash]
	}
	return map[string]string{
		"buildings":                   "buildings",
		"system-types":                "system_types",
		"system-parts":                "system_parts",
		"apparats":                    "apparats",
		"control-cabinets":            "control_cabinets",
		"sps-controllers":             "sps_controllers",
		"sps-controller-system-types": "sps_controller_system_types",
		"field-devices":               "field_devices",
		"bacnet-objects":              "bacnet_objects",
		"object-data":                 "object_data",
		"state-texts":                 "state_texts",
		"notification-classes":        "notification_classes",
		"alarm-definitions":           "alarm_definitions",
		"alarm-types":                 "alarm_types",
		"alarm-type-fields":           "alarm_type_fields",
		"units":                       "units",
		"alarm-fields":                "alarm_fields",
	}[segment]
}

func withFacilityChangeBroadcast(next gin.HandlerFunc, broadcaster FacilityChangeBroadcaster, change facilityMutation) gin.HandlerFunc {
	return func(c *gin.Context) {
		originalWriter := c.Writer
		capture := &facilityChangeResponseWriter{ResponseWriter: originalWriter}
		c.Writer = capture
		next(c)
		c.Writer = originalWriter
		if capture.Status() < http.StatusOK || capture.Status() >= http.StatusMultipleChoices {
			return
		}
		var ids []uuid.UUID
		if change.pathIDIsTarget {
			ids = facilityChangeIDs(c)
		}
		if len(ids) == 0 {
			ids = facilityChangeIDsFromResponse(capture.body.Bytes())
		}
		broadcaster.BroadcastFacilityChange(c.Request.Context(), change.resource, change.action, ids, currentActorID(c))
	}
}

type facilityChangeResponseWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *facilityChangeResponseWriter) Write(data []byte) (int, error) {
	_, _ = w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *facilityChangeResponseWriter) WriteString(value string) (int, error) {
	_, _ = w.body.WriteString(value)
	return w.ResponseWriter.WriteString(value)
}

func facilityChangeIDs(c *gin.Context) []uuid.UUID {
	value := c.Param("id")
	if value == "" {
		return nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil
	}
	return []uuid.UUID{id}
}

// facilityChangeIDsFromResponse keeps the realtime event resource-scoped. It
// accepts the top-level DTO ID, conventional item arrays, and the explicit
// successful result shapes used by field-device bulk operations. It does not
// recurse through related entities, which would otherwise publish e.g. a
// system-part ID as an apparat change.
func facilityChangeIDsFromResponse(body []byte) []uuid.UUID {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}

	ids := make([]uuid.UUID, 0)
	if id, ok := facilityChangeID(payload); ok {
		ids = append(ids, id)
	}
	for _, key := range []string{"items", "results"} {
		var items []map[string]json.RawMessage
		if err := json.Unmarshal(payload[key], &items); err != nil {
			continue
		}
		for _, item := range items {
			if !facilityChangeResultSucceeded(item) {
				continue
			}
			if nested, ok := facilityChangeNestedResultID(item); ok {
				ids = append(ids, nested)
				continue
			}
			if id, ok := facilityChangeID(item); ok {
				ids = append(ids, id)
			}
		}
	}
	return uniqueFacilityChangeIDs(ids)
}

func facilityChangeID(payload map[string]json.RawMessage) (uuid.UUID, bool) {
	var raw string
	if err := json.Unmarshal(payload["id"], &raw); err != nil {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}

func facilityChangeResultSucceeded(payload map[string]json.RawMessage) bool {
	value, exists := payload["success"]
	if !exists {
		return true
	}
	var success bool
	return json.Unmarshal(value, &success) == nil && success
}

func facilityChangeNestedResultID(payload map[string]json.RawMessage) (uuid.UUID, bool) {
	for _, key := range []string{"field_device", "item"} {
		var nested map[string]json.RawMessage
		if err := json.Unmarshal(payload[key], &nested); err != nil {
			continue
		}
		if id, ok := facilityChangeID(nested); ok {
			return id, true
		}
	}
	return uuid.Nil, false
}

func uniqueFacilityChangeIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return []uuid.UUID{}
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	unique := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}
