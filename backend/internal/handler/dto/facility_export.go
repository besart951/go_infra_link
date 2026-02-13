package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateFieldDeviceExportRequest struct {
	ProjectIDs        []uuid.UUID `json:"project_ids" binding:"omitempty,dive,uuid"`
	BuildingIDs       []uuid.UUID `json:"buildings_id" binding:"omitempty,dive,uuid"`
	ControlCabinetIDs []uuid.UUID `json:"control_cabinet_id" binding:"omitempty,dive,uuid"`
	SPSControllerIDs  []uuid.UUID `json:"sps_controller_id" binding:"omitempty,dive,uuid"`
	ForceAsync        bool        `json:"force_async"`
}

type FieldDeviceExportJobResponse struct {
	JobID       uuid.UUID `json:"job_id"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	Message     string    `json:"message"`
	OutputType  string    `json:"output_type,omitempty"`
	FileName    string    `json:"file_name,omitempty"`
	ContentType string    `json:"content_type,omitempty"`
	DownloadURL string    `json:"download_url,omitempty"`
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
