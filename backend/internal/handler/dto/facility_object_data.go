package dto

import (
	"time"

	"github.com/google/uuid"
)

// Facility DTOs - ObjectData

type CreateObjectDataRequest struct {
	Description   string               `json:"description" binding:"required,max=250"`
	Version       string               `json:"version" binding:"required,max=100"`
	IsActive      *bool                `json:"is_active"`
	ProjectID     *uuid.UUID           `json:"project_id"`
	BacnetObjects *[]BacnetObjectInput `json:"bacnet_objects" binding:"omitempty,dive"`
}

type UpdateObjectDataRequest struct {
	Description *string    `json:"description"`
	Version     *string    `json:"version"`
	IsActive    *bool      `json:"is_active"`
	ProjectID   *uuid.UUID `json:"project_id"`
}

type ObjectDataResponse struct {
	ID            uuid.UUID              `json:"id"`
	Description   string                 `json:"description"`
	Version       string                 `json:"version"`
	IsActive      bool                   `json:"is_active"`
	ProjectID     *uuid.UUID             `json:"project_id"`
	BacnetObjects []BacnetObjectResponse `json:"bacnet_objects"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type ObjectDataListResponse struct {
	Items      []ObjectDataResponse `json:"items"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	TotalPages int                  `json:"total_pages"`
}
