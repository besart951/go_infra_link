package dto

import (
	"time"

	"github.com/google/uuid"
)

// BacnetObjectAlarmValue DTOs

type AlarmValueInput struct {
	AlarmTypeFieldID uuid.UUID  `json:"alarm_type_field_id" binding:"required"`
	ValueNumber      *float64   `json:"value_number,omitempty"`
	ValueInteger     *int64     `json:"value_integer,omitempty"`
	ValueBoolean     *bool      `json:"value_boolean,omitempty"`
	ValueString      *string    `json:"value_string,omitempty"`
	ValueJSON        *string    `json:"value_json,omitempty"`
	UnitID           *uuid.UUID `json:"unit_id,omitempty"`
	Source           string     `json:"source,omitempty"`
}

type PutAlarmValuesRequest struct {
	Values []AlarmValueInput `json:"values" binding:"required"`
}

type AlarmValueResponse struct {
	ID               uuid.UUID  `json:"id"`
	BacnetObjectID   uuid.UUID  `json:"bacnet_object_id"`
	AlarmTypeFieldID uuid.UUID  `json:"alarm_type_field_id"`
	ValueNumber      *float64   `json:"value_number,omitempty"`
	ValueInteger     *int64     `json:"value_integer,omitempty"`
	ValueBoolean     *bool      `json:"value_boolean,omitempty"`
	ValueString      *string    `json:"value_string,omitempty"`
	ValueJSON        *string    `json:"value_json,omitempty"`
	UnitID           *uuid.UUID `json:"unit_id,omitempty"`
	Source           string     `json:"source"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type AlarmValuesResponse struct {
	Items []AlarmValueResponse `json:"items"`
}
