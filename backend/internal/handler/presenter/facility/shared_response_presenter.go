package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
)

func ToControlCabinetResponse(controlCabinet domainFacility.ControlCabinet) dto.ControlCabinetResponse {
	return dto.ControlCabinetResponse{
		ID:               controlCabinet.ID,
		BuildingID:       controlCabinet.BuildingID,
		ControlCabinetNr: controlCabinet.ControlCabinetNr,
		CreatedAt:        controlCabinet.CreatedAt,
		UpdatedAt:        controlCabinet.UpdatedAt,
	}
}

func ToSPSControllerResponse(controller domainFacility.SPSController) dto.SPSControllerResponse {
	return dto.SPSControllerResponse{
		ID:                controller.ID,
		ControlCabinetID:  controller.ControlCabinetID,
		GADevice:          controller.GADevice,
		DeviceName:        controller.DeviceName,
		DeviceDescription: controller.DeviceDescription,
		DeviceLocation:    controller.DeviceLocation,
		IPAddress:         controller.IPAddress,
		Subnet:            controller.Subnet,
		Gateway:           controller.Gateway,
		Vlan:              controller.Vlan,
		CreatedAt:         controller.CreatedAt,
		UpdatedAt:         controller.UpdatedAt,
	}
}

func ToSPSControllerSystemTypeResponse(item domainFacility.SPSControllerSystemType) dto.SPSControllerSystemTypeResponse {
	return dto.SPSControllerSystemTypeResponse{
		ID:                item.ID,
		SPSControllerID:   item.SPSControllerID,
		SystemTypeID:      item.SystemTypeID,
		SPSControllerName: item.SPSController.DeviceName,
		SystemTypeName:    item.SystemType.Name,
		Number:            item.Number,
		DocumentName:      item.DocumentName,
		FieldDevicesCount: item.FieldDevicesCount,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

func ToFieldDeviceOptionsResponse(options *domainFacility.FieldDeviceOptions) dto.FieldDeviceOptionsResponse {
	apparats := make([]dto.ApparatResponse, len(options.Apparats))
	for i, apparat := range options.Apparats {
		apparats[i] = toApparatResponse(apparat)
	}

	systemParts := make([]dto.SystemPartResponse, len(options.SystemParts))
	for i, systemPart := range options.SystemParts {
		systemParts[i] = toSystemPartResponse(systemPart)
	}

	objectDatas := make([]dto.ObjectDataResponse, len(options.ObjectDatas))
	for i, objectData := range options.ObjectDatas {
		objectDatas[i] = toObjectDataResponse(objectData)
	}

	apparatToSystemPart := make(map[string][]string, len(options.ApparatToSystemPart))
	for apparatID, systemPartIDs := range options.ApparatToSystemPart {
		ids := make([]string, len(systemPartIDs))
		for i, id := range systemPartIDs {
			ids[i] = id.String()
		}
		apparatToSystemPart[apparatID.String()] = ids
	}

	objectDataToApparat := make(map[string][]string, len(options.ObjectDataToApparat))
	for objectDataID, apparatIDs := range options.ObjectDataToApparat {
		ids := make([]string, len(apparatIDs))
		for i, id := range apparatIDs {
			ids[i] = id.String()
		}
		objectDataToApparat[objectDataID.String()] = ids
	}

	return dto.FieldDeviceOptionsResponse{
		Apparats:            apparats,
		SystemParts:         systemParts,
		ObjectDatas:         objectDatas,
		ApparatToSystemPart: apparatToSystemPart,
		ObjectDataToApparat: objectDataToApparat,
	}
}

func toApparatResponse(apparat domainFacility.Apparat) dto.ApparatResponse {
	systemParts := make([]dto.SystemPartResponse, 0, len(apparat.SystemParts))
	for _, systemPart := range apparat.SystemParts {
		if systemPart == nil {
			continue
		}
		systemParts = append(systemParts, toSystemPartResponse(*systemPart))
	}

	return dto.ApparatResponse{
		ID:          apparat.ID,
		ShortName:   apparat.ShortName,
		Name:        apparat.Name,
		Description: apparat.Description,
		SystemParts: systemParts,
		CreatedAt:   apparat.CreatedAt,
		UpdatedAt:   apparat.UpdatedAt,
	}
}

func toSystemPartResponse(systemPart domainFacility.SystemPart) dto.SystemPartResponse {
	return dto.SystemPartResponse{
		ID:          systemPart.ID,
		ShortName:   systemPart.ShortName,
		Name:        systemPart.Name,
		Description: systemPart.Description,
		CreatedAt:   systemPart.CreatedAt,
		UpdatedAt:   systemPart.UpdatedAt,
	}
}

func toObjectDataResponse(objectData domainFacility.ObjectData) dto.ObjectDataResponse {
	bacnetObjects := make([]domainFacility.BacnetObject, 0, len(objectData.BacnetObjects))
	for _, bacnetObject := range objectData.BacnetObjects {
		if bacnetObject == nil {
			continue
		}
		bacnetObjects = append(bacnetObjects, *bacnetObject)
	}

	apparats := make([]dto.ApparatResponse, 0, len(objectData.Apparats))
	for _, apparat := range objectData.Apparats {
		if apparat == nil {
			continue
		}
		apparats = append(apparats, toApparatResponse(*apparat))
	}

	return dto.ObjectDataResponse{
		ID:            objectData.ID,
		Description:   objectData.Description,
		Version:       objectData.Version,
		IsActive:      objectData.IsActive,
		ProjectID:     objectData.ProjectID,
		Apparats:      apparats,
		BacnetObjects: toBacnetObjectResponses(bacnetObjects),
		CreatedAt:     objectData.CreatedAt,
		UpdatedAt:     objectData.UpdatedAt,
	}
}

func toBacnetObjectResponses(items []domainFacility.BacnetObject) []dto.BacnetObjectResponse {
	responses := make([]dto.BacnetObjectResponse, len(items))
	for i, item := range items {
		responses[i] = dto.BacnetObjectResponse{
			ID:                  item.ID.String(),
			TextFix:             item.TextFix,
			Description:         item.Description,
			GMSVisible:          item.GMSVisible,
			Optional:            item.Optional,
			TextIndividual:      item.TextIndividual,
			SoftwareType:        string(item.SoftwareType),
			SoftwareNumber:      int(item.SoftwareNumber),
			HardwareType:        string(item.HardwareType),
			HardwareQuantity:    int(item.HardwareQuantity),
			FieldDeviceID:       item.FieldDeviceID,
			SoftwareReferenceID: item.SoftwareReferenceID,
			StateTextID:         item.StateTextID,
			NotificationClassID: item.NotificationClassID,
			AlarmDefinitionID:   item.AlarmDefinitionID,
			AlarmTypeID:         item.AlarmTypeID,
			CreatedAt:           item.CreatedAt,
			UpdatedAt:           item.UpdatedAt,
		}
	}
	return responses
}
