package dto

import (
	"time"

	"github.com/google/uuid"
)

// Unit DTOs

type UnitResponse struct {
	ID     uuid.UUID `json:"id"`
	Code   string    `json:"code"`
	Symbol string    `json:"symbol"`
	Name   string    `json:"name"`
}

type CreateUnitRequest struct {
	Code   string `json:"code" binding:"required,max=30"`
	Symbol string `json:"symbol" binding:"required,max=20"`
	Name   string `json:"name" binding:"required,max=100"`
}

type UpdateUnitRequest struct {
	Code   *string `json:"code" binding:"omitempty,max=30"`
	Symbol *string `json:"symbol" binding:"omitempty,max=20"`
	Name   *string `json:"name" binding:"omitempty,max=100"`
}

type UnitListResponse struct {
	Items      []UnitResponse `json:"items"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	TotalPages int            `json:"total_pages"`
}

// AlarmField DTOs

type AlarmFieldResponse struct {
	ID              uuid.UUID `json:"id"`
	Key             string    `json:"key"`
	Label           string    `json:"label"`
	DataType        string    `json:"data_type"`
	DefaultUnitCode *string   `json:"default_unit_code,omitempty"`
}

type CreateAlarmFieldRequest struct {
	Key             string  `json:"key" binding:"required,max=100"`
	Label           string  `json:"label" binding:"required,max=150"`
	DataType        string  `json:"data_type" binding:"required,max=30"`
	DefaultUnitCode *string `json:"default_unit_code" binding:"omitempty,max=30"`
}

type UpdateAlarmFieldRequest struct {
	Key             *string `json:"key" binding:"omitempty,max=100"`
	Label           *string `json:"label" binding:"omitempty,max=150"`
	DataType        *string `json:"data_type" binding:"omitempty,max=30"`
	DefaultUnitCode *string `json:"default_unit_code" binding:"omitempty,max=30"`
}

type AlarmFieldListResponse struct {
	Items      []AlarmFieldResponse `json:"items"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	TotalPages int                  `json:"total_pages"`
}

// AlarmTypeField DTOs

type AlarmTypeFieldResponse struct {
	ID               uuid.UUID           `json:"id"`
	AlarmTypeID      uuid.UUID           `json:"alarm_type_id"`
	AlarmFieldID     uuid.UUID           `json:"alarm_field_id"`
	AlarmField       *AlarmFieldResponse `json:"alarm_field,omitempty"`
	DisplayOrder     int                 `json:"display_order"`
	IsRequired       bool                `json:"is_required"`
	IsUserEditable   bool                `json:"is_user_editable"`
	DefaultValueJSON *string             `json:"default_value_json,omitempty"`
	ValidationJSON   *string             `json:"validation_json,omitempty"`
	DefaultUnitID    *uuid.UUID          `json:"default_unit_id,omitempty"`
	DefaultUnit      *UnitResponse       `json:"default_unit,omitempty"`
	UIGroup          *string             `json:"ui_group,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

type CreateAlarmTypeFieldRequest struct {
	AlarmFieldID     uuid.UUID  `json:"alarm_field_id" binding:"required"`
	DisplayOrder     int        `json:"display_order" binding:"min=0"`
	IsRequired       bool       `json:"is_required"`
	IsUserEditable   bool       `json:"is_user_editable"`
	DefaultValueJSON *string    `json:"default_value_json"`
	ValidationJSON   *string    `json:"validation_json"`
	DefaultUnitID    *uuid.UUID `json:"default_unit_id"`
	UIGroup          *string    `json:"ui_group" binding:"omitempty,max=80"`
}

type UpdateAlarmTypeFieldRequest struct {
	DisplayOrder     *int       `json:"display_order" binding:"omitempty,min=0"`
	IsRequired       *bool      `json:"is_required"`
	IsUserEditable   *bool      `json:"is_user_editable"`
	DefaultValueJSON *string    `json:"default_value_json"`
	ValidationJSON   *string    `json:"validation_json"`
	DefaultUnitID    *uuid.UUID `json:"default_unit_id"`
	UIGroup          *string    `json:"ui_group" binding:"omitempty,max=80"`
}

// AlarmType DTOs

type CreateAlarmTypeRequest struct {
	Code string `json:"code" binding:"required,max=80"`
	Name string `json:"name" binding:"required,max=120"`
}

type UpdateAlarmTypeRequest struct {
	Name *string `json:"name" binding:"omitempty,max=120"`
}

type AlarmTypeResponse struct {
	ID        uuid.UUID                `json:"id"`
	Code      string                   `json:"code"`
	Name      string                   `json:"name"`
	Fields    []AlarmTypeFieldResponse `json:"fields,omitempty"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
}

type AlarmTypeListResponse struct {
	Items      []AlarmTypeResponse `json:"items"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	TotalPages int                 `json:"total_pages"`
}
