package mapper

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
)

func ToProjectControlCabinetResponse(item project.ProjectControlCabinet) dto.ProjectControlCabinetResponse {
	return dto.ProjectControlCabinetResponse{
		ID:               item.ID,
		ProjectID:        item.ProjectID,
		ControlCabinetID: item.ControlCabinetID,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}

func ToProjectControlCabinetList(items []project.ProjectControlCabinet) []dto.ProjectControlCabinetResponse {
	out := make([]dto.ProjectControlCabinetResponse, len(items))
	for i, item := range items {
		out[i] = ToProjectControlCabinetResponse(item)
	}
	return out
}

func ToProjectSPSControllerResponse(item project.ProjectSPSController) dto.ProjectSPSControllerResponse {
	return dto.ProjectSPSControllerResponse{
		ID:              item.ID,
		ProjectID:       item.ProjectID,
		SPSControllerID: item.SPSControllerID,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func ToProjectSPSControllerList(items []project.ProjectSPSController) []dto.ProjectSPSControllerResponse {
	out := make([]dto.ProjectSPSControllerResponse, len(items))
	for i, item := range items {
		out[i] = ToProjectSPSControllerResponse(item)
	}
	return out
}

func ToProjectFieldDeviceResponse(item project.ProjectFieldDevice) dto.ProjectFieldDeviceResponse {
	return dto.ProjectFieldDeviceResponse{
		ID:            item.ID,
		ProjectID:     item.ProjectID,
		FieldDeviceID: item.FieldDeviceID,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func ToProjectFieldDeviceList(items []project.ProjectFieldDevice) []dto.ProjectFieldDeviceResponse {
	out := make([]dto.ProjectFieldDeviceResponse, len(items))
	for i, item := range items {
		out[i] = ToProjectFieldDeviceResponse(item)
	}
	return out
}
