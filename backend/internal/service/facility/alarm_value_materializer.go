package facility

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type alarmValueMaterializer struct {
	alarmTypeRepo domainFacility.AlarmTypeRepository
}

func newAlarmValueMaterializer(alarmTypeRepo domainFacility.AlarmTypeRepository) alarmValueMaterializer {
	return alarmValueMaterializer{alarmTypeRepo: alarmTypeRepo}
}

func (m alarmValueMaterializer) buildDefaultValues(ctx context.Context, bacnetObjects []*domainFacility.BacnetObject) ([]*domainFacility.BacnetObjectAlarmValue, error) {
	if len(bacnetObjects) == 0 {
		return nil, nil
	}

	alarmTypeCache := make(map[uuid.UUID]*domainFacility.AlarmType)
	values := make([]*domainFacility.BacnetObjectAlarmValue, 0)

	for _, obj := range bacnetObjects {
		if obj == nil || obj.AlarmTypeID == nil {
			continue
		}

		alarmType, ok := alarmTypeCache[*obj.AlarmTypeID]
		if !ok {
			loaded, err := m.alarmTypeRepo.GetWithFields(ctx, *obj.AlarmTypeID)
			if err != nil {
				return nil, err
			}
			if loaded == nil {
				return nil, domain.ErrNotFound
			}
			alarmType = loaded
			alarmTypeCache[*obj.AlarmTypeID] = loaded
		}

		for _, field := range alarmType.Fields {
			value := &domainFacility.BacnetObjectAlarmValue{
				BacnetObjectID:   obj.ID,
				AlarmTypeFieldID: field.ID,
				UnitID:           field.DefaultUnitID,
				Source:           domainFacility.AlarmValueSourceDefault,
			}

			if field.DefaultValueJSON != nil && field.AlarmField != nil {
				applyAlarmDefaultValue(value, field.AlarmField.DataType, *field.DefaultValueJSON)
			}

			values = append(values, value)
		}
	}

	return values, nil
}

func applyAlarmDefaultValue(value *domainFacility.BacnetObjectAlarmValue, dataType string, defaultValueJSON string) {
	if value == nil {
		return
	}

	var decoded any
	if err := json.Unmarshal([]byte(defaultValueJSON), &decoded); err != nil {
		value.ValueString = &defaultValueJSON
		return
	}

	switch strings.ToLower(strings.TrimSpace(dataType)) {
	case "number", "duration":
		if n, ok := toFloat64(decoded); ok {
			value.ValueNumber = &n
		}
	case "integer":
		if n, ok := toInt64(decoded); ok {
			value.ValueInteger = &n
		}
	case "boolean":
		if b, ok := decoded.(bool); ok {
			value.ValueBoolean = &b
		}
	case "string", "enum":
		if s, ok := decoded.(string); ok {
			value.ValueString = &s
		}
	case "state_map", "json":
		if b, err := json.Marshal(decoded); err == nil {
			raw := string(b)
			value.ValueJSON = &raw
		}
	default:
		if b, err := json.Marshal(decoded); err == nil {
			raw := string(b)
			value.ValueJSON = &raw
		}
	}
}

func toFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

func toInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	default:
		return 0, false
	}
}
