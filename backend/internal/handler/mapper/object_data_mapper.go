package mapper

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
)

func ToObjectDataResponse(item domainFacility.ObjectData) dto.ObjectDataResponse {
	return dto.ObjectDataResponse{
		ID:          item.ID,
		Description: item.Description,
		Version:     item.Version,
		IsActive:    item.IsActive,
		ProjectID:   item.ProjectID,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func ToObjectDataList(items []domainFacility.ObjectData) []dto.ObjectDataResponse {
	out := make([]dto.ObjectDataResponse, len(items))
	for i, item := range items {
		out[i] = ToObjectDataResponse(item)
	}
	return out
}
