package mapper

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
)

func ToObjectDataResponse(item domainFacility.ObjectData) dto.ObjectDataResponse {
	bacnetObjects := []domainFacility.BacnetObject{}
	if len(item.BacnetObjects) > 0 {
		bacnetObjects = make([]domainFacility.BacnetObject, 0, len(item.BacnetObjects))
		for _, obj := range item.BacnetObjects {
			if obj == nil {
				continue
			}
			bacnetObjects = append(bacnetObjects, *obj)
		}
	}

	return dto.ObjectDataResponse{
		ID:            item.ID,
		Description:   item.Description,
		Version:       item.Version,
		IsActive:      item.IsActive,
		ProjectID:     item.ProjectID,
		BacnetObjects: mapperBacnetObjectResponses(bacnetObjects),
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func ToObjectDataList(items []domainFacility.ObjectData) []dto.ObjectDataResponse {
	out := make([]dto.ObjectDataResponse, len(items))
	for i, item := range items {
		out[i] = ToObjectDataResponse(item)
	}
	return out
}

func mapperBacnetObjectResponses(objs []domainFacility.BacnetObject) []dto.BacnetObjectResponse {
	items := make([]dto.BacnetObjectResponse, len(objs))
	for i, obj := range objs {
		items[i] = dto.BacnetObjectResponse{
			ID:                  obj.ID.String(),
			TextFix:             obj.TextFix,
			Description:         obj.Description,
			GMSVisible:          obj.GMSVisible,
			Optional:            obj.Optional,
			TextIndividual:      obj.TextIndividual,
			SoftwareType:        string(obj.SoftwareType),
			SoftwareNumber:      int(obj.SoftwareNumber),
			HardwareType:        string(obj.HardwareType),
			HardwareQuantity:    int(obj.HardwareQuantity),
			FieldDeviceID:       obj.FieldDeviceID,
			SoftwareReferenceID: obj.SoftwareReferenceID,
			StateTextID:         obj.StateTextID,
			NotificationClassID: obj.NotificationClassID,
			AlarmDefinitionID:   obj.AlarmDefinitionID,
			AlarmTypeID:         obj.AlarmTypeID,
			CreatedAt:           obj.CreatedAt,
			UpdatedAt:           obj.UpdatedAt,
		}
	}
	return items
}
