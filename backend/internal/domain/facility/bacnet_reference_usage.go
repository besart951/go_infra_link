package facility

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type BacnetReferenceResource string

const (
	BacnetReferenceResourceApparat           BacnetReferenceResource = "apparat"
	BacnetReferenceResourceSystemPart        BacnetReferenceResource = "system_part"
	BacnetReferenceResourceSystemType        BacnetReferenceResource = "system_type"
	BacnetReferenceResourceStateText         BacnetReferenceResource = "state_text"
	BacnetReferenceResourceNotificationClass BacnetReferenceResource = "notification_class"
	BacnetReferenceResourceAlarmType         BacnetReferenceResource = "alarm_type"
	BacnetReferenceResourceAlarmDefinition   BacnetReferenceResource = "alarm_definition"
	BacnetReferenceResourceObjectData        BacnetReferenceResource = "object_data"
)

var ErrBacnetReferenceInUse = errors.New("bacnet reference in use")

type BacnetReferenceUsage struct {
	Resource          BacnetReferenceResource
	ID                uuid.UUID
	BacnetObjectCount int64
}

func ParseBacnetReferenceResource(value string) (BacnetReferenceResource, bool) {
	resource := BacnetReferenceResource(strings.TrimSpace(value))
	switch resource {
	case BacnetReferenceResourceApparat,
		BacnetReferenceResourceSystemPart,
		BacnetReferenceResourceSystemType,
		BacnetReferenceResourceStateText,
		BacnetReferenceResourceNotificationClass,
		BacnetReferenceResourceAlarmType,
		BacnetReferenceResourceAlarmDefinition,
		BacnetReferenceResourceObjectData:
		return resource, true
	default:
		return "", false
	}
}
