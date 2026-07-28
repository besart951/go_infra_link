package fielddevice

import (
	"sort"
	"strings"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

const (
	itemErrorCodeValidation    = "validation_failed"
	itemErrorCodeNotFound      = "not_found"
	itemErrorCodeConflict      = "conflict"
	itemErrorCodeNotConfigured = "not_configured"
	itemErrorCodeOperation     = "operation_failed"
	itemErrorCodeInvalid       = "invalid_argument"
)

func normalizeMultiCreateResult(
	result *domainFacility.FieldDeviceMultiCreateResult,
	items []domainFacility.FieldDeviceCreateItem,
) {
	if result == nil {
		return
	}
	for index := range result.Results {
		item := &result.Results[index]
		if item.FieldDevice != nil && item.FieldDevice.ID != uuid.Nil {
			item.ID = item.FieldDevice.ID
		} else if index < len(items) && items[index].FieldDevice != nil {
			item.ID = items[index].FieldDevice.ID
		}
		if item.Success {
			item.ErrorCode = ""
			item.Reason = ""
			continue
		}
		item.Reason = item.Error
		if item.ErrorCode == "" {
			item.ErrorCode = classifyItemError(item.Error, item.ErrorField)
		}
	}
}

func normalizeBulkResult(
	result *domainFacility.BulkOperationResult,
	ids []uuid.UUID,
) {
	if result == nil {
		return
	}
	for index := range result.Results {
		item := &result.Results[index]
		if item.ID == uuid.Nil && index < len(ids) {
			item.ID = ids[index]
		}
		if item.Success {
			item.ErrorCode = ""
			item.ErrorField = ""
			item.Reason = ""
			continue
		}
		if item.ErrorField == "" && len(item.Fields) > 0 {
			fields := make([]string, 0, len(item.Fields))
			for field := range item.Fields {
				fields = append(fields, field)
			}
			sort.Strings(fields)
			item.ErrorField = fields[0]
			if item.Error == "" {
				item.Error = item.Fields[item.ErrorField]
			}
		}
		item.Reason = item.Error
		if item.ErrorCode == "" {
			item.ErrorCode = classifyItemError(item.Error, item.ErrorField)
		}
	}
}

func classifyItemError(reason, field string) string {
	normalized := strings.ToLower(strings.TrimSpace(reason))
	switch {
	case strings.Contains(normalized, "not configured"):
		return itemErrorCodeNotConfigured
	case strings.Contains(normalized, "not found"):
		return itemErrorCodeNotFound
	case strings.Contains(normalized, "conflict"),
		strings.Contains(normalized, "already"),
		strings.Contains(normalized, "bereits vergeben"):
		return itemErrorCodeConflict
	case strings.TrimSpace(field) != "":
		return itemErrorCodeValidation
	case strings.Contains(normalized, "invalid"):
		return itemErrorCodeInvalid
	default:
		return itemErrorCodeOperation
	}
}
