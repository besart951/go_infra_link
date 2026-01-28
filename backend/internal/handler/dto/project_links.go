package dto

import (
	"time"

	"github.com/google/uuid"
)

// Project link DTOs

type CreateProjectUserRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

type ProjectUserResponse struct {
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
}

type CreateProjectControlCabinetRequest struct {
	ControlCabinetID uuid.UUID `json:"control_cabinet_id" binding:"required"`
}

type UpdateProjectControlCabinetRequest struct {
	ControlCabinetID uuid.UUID `json:"control_cabinet_id" binding:"required"`
}

type ProjectControlCabinetResponse struct {
	ID               uuid.UUID `json:"id"`
	ProjectID        uuid.UUID `json:"project_id"`
	ControlCabinetID uuid.UUID `json:"control_cabinet_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ProjectControlCabinetListResponse struct {
	Items      []ProjectControlCabinetResponse `json:"items"`
	Total      int64                           `json:"total"`
	Page       int                             `json:"page"`
	TotalPages int                             `json:"total_pages"`
}

type CreateProjectSPSControllerRequest struct {
	SPSControllerID uuid.UUID `json:"sps_controller_id" binding:"required"`
}

type UpdateProjectSPSControllerRequest struct {
	SPSControllerID uuid.UUID `json:"sps_controller_id" binding:"required"`
}

type ProjectSPSControllerResponse struct {
	ID              uuid.UUID `json:"id"`
	ProjectID       uuid.UUID `json:"project_id"`
	SPSControllerID uuid.UUID `json:"sps_controller_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ProjectSPSControllerListResponse struct {
	Items      []ProjectSPSControllerResponse `json:"items"`
	Total      int64                          `json:"total"`
	Page       int                            `json:"page"`
	TotalPages int                            `json:"total_pages"`
}

type CreateProjectFieldDeviceRequest struct {
	FieldDeviceID uuid.UUID `json:"field_device_id" binding:"required"`
}

type UpdateProjectFieldDeviceRequest struct {
	FieldDeviceID uuid.UUID `json:"field_device_id" binding:"required"`
}

type ProjectFieldDeviceResponse struct {
	ID            uuid.UUID `json:"id"`
	ProjectID     uuid.UUID `json:"project_id"`
	FieldDeviceID uuid.UUID `json:"field_device_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ProjectFieldDeviceListResponse struct {
	Items      []ProjectFieldDeviceResponse `json:"items"`
	Total      int64                        `json:"total"`
	Page       int                          `json:"page"`
	TotalPages int                          `json:"total_pages"`
}
