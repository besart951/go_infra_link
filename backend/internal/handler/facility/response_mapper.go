package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/google/uuid"
)

func toBuildingResponse(building domainFacility.Building) dto.BuildingResponse {
	return dto.BuildingResponse{
		ID:            building.ID,
		IWSCode:       building.IWSCode,
		BuildingGroup: building.BuildingGroup,
		CreatedAt:     building.CreatedAt,
		UpdatedAt:     building.UpdatedAt,
	}
}

func toBuildingResponses(buildings []domainFacility.Building) []dto.BuildingResponse {
	items := make([]dto.BuildingResponse, len(buildings))
	for i, item := range buildings {
		items[i] = toBuildingResponse(item)
	}
	return items
}

func toBuildingListResponse(list *domain.PaginatedList[domainFacility.Building]) dto.BuildingListResponse {
	return dto.BuildingListResponse{
		Items:      toBuildingResponses(list.Items),
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
	}
}

func toSystemTypeResponse(systemType domainFacility.SystemType) dto.SystemTypeResponse {
	return dto.SystemTypeResponse{
		ID:        systemType.ID,
		NumberMin: systemType.NumberMin,
		NumberMax: systemType.NumberMax,
		Name:      systemType.Name,
		CreatedAt: systemType.CreatedAt,
		UpdatedAt: systemType.UpdatedAt,
	}
}

func toSystemTypeListResponse(list *domain.PaginatedList[domainFacility.SystemType]) dto.SystemTypeListResponse {
	items := make([]dto.SystemTypeResponse, len(list.Items))
	for i, item := range list.Items {
		items[i] = toSystemTypeResponse(item)
	}

	return dto.SystemTypeListResponse{
		Items:      items,
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
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

func toSystemPartListResponse(list *domain.PaginatedList[domainFacility.SystemPart]) dto.SystemPartListResponse {
	items := make([]dto.SystemPartResponse, len(list.Items))
	for i, item := range list.Items {
		items[i] = toSystemPartResponse(item)
	}

	return dto.SystemPartListResponse{
		Items:      items,
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
	}
}

func toSpecificationResponse(specification domainFacility.Specification) dto.SpecificationResponse {
	return dto.SpecificationResponse{
		ID:                       specification.ID,
		FieldDeviceID:            specification.FieldDeviceID,
		SpecificationSupplier:    specification.SpecificationSupplier,
		SpecificationBrand:       specification.SpecificationBrand,
		SpecificationType:        specification.SpecificationType,
		AdditionalInfoMotorValve: specification.AdditionalInfoMotorValve,
		AdditionalInfoSize:       specification.AdditionalInfoSize,
		AdditionalInformationInstallationLocation: specification.AdditionalInformationInstallationLocation,
		ElectricalConnectionPH:                    specification.ElectricalConnectionPH,
		ElectricalConnectionACDC:                  specification.ElectricalConnectionACDC,
		ElectricalConnectionAmperage:              specification.ElectricalConnectionAmperage,
		ElectricalConnectionPower:                 specification.ElectricalConnectionPower,
		ElectricalConnectionRotation:              specification.ElectricalConnectionRotation,
		CreatedAt:                                 specification.CreatedAt,
		UpdatedAt:                                 specification.UpdatedAt,
	}
}

func toApparatResponse(apparat domainFacility.Apparat) dto.ApparatResponse {
	systemParts := make([]dto.SystemPartResponse, 0)
	if apparat.SystemParts != nil {
		systemParts = make([]dto.SystemPartResponse, len(apparat.SystemParts))
		for i, sp := range apparat.SystemParts {
			if sp != nil {
				systemParts[i] = toSystemPartResponse(*sp)
			}
		}
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

func toApparatResponses(apparats []domainFacility.Apparat) []dto.ApparatResponse {
	items := make([]dto.ApparatResponse, len(apparats))
	for i, item := range apparats {
		items[i] = toApparatResponse(item)
	}
	return items
}

func toApparatListResponse(list *domain.PaginatedList[domainFacility.Apparat]) dto.ApparatListResponse {
	return dto.ApparatListResponse{
		Items:      toApparatResponses(list.Items),
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
	}
}

func toControlCabinetResponse(controlCabinet domainFacility.ControlCabinet) dto.ControlCabinetResponse {
	return dto.ControlCabinetResponse{
		ID:               controlCabinet.ID,
		BuildingID:       controlCabinet.BuildingID,
		ControlCabinetNr: controlCabinet.ControlCabinetNr,
		CreatedAt:        controlCabinet.CreatedAt,
		UpdatedAt:        controlCabinet.UpdatedAt,
	}
}

func toControlCabinetResponses(items []domainFacility.ControlCabinet) []dto.ControlCabinetResponse {
	responses := make([]dto.ControlCabinetResponse, len(items))
	for i, item := range items {
		responses[i] = toControlCabinetResponse(item)
	}
	return responses
}

func toControlCabinetListResponse(list *domain.PaginatedList[domainFacility.ControlCabinet]) dto.ControlCabinetListResponse {
	return dto.ControlCabinetListResponse{
		Items:      toControlCabinetResponses(list.Items),
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
	}
}

func toSPSControllerResponse(controller domainFacility.SPSController) dto.SPSControllerResponse {
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

func toSPSControllerResponses(items []domainFacility.SPSController) []dto.SPSControllerResponse {
	responses := make([]dto.SPSControllerResponse, len(items))
	for i, item := range items {
		responses[i] = toSPSControllerResponse(item)
	}
	return responses
}

func toSPSControllerListResponse(list *domain.PaginatedList[domainFacility.SPSController]) dto.SPSControllerListResponse {
	return dto.SPSControllerListResponse{
		Items:      toSPSControllerResponses(list.Items),
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
	}
}

func toFieldDeviceResponse(fieldDevice domainFacility.FieldDevice) dto.FieldDeviceResponse {
	var systemPartID *uuid.UUID
	if fieldDevice.SystemPartID != uuid.Nil {
		systemPartID = &fieldDevice.SystemPartID
	}

	resp := dto.FieldDeviceResponse{
		ID:                        fieldDevice.ID,
		BMK:                       fieldDevice.BMK,
		Description:               fieldDevice.Description,
		ApparatNr:                 &fieldDevice.ApparatNr,
		SPSControllerSystemTypeID: fieldDevice.SPSControllerSystemTypeID,
		SystemPartID:              systemPartID,
		SpecificationID:           fieldDevice.SpecificationID,
		ApparatID:                 fieldDevice.ApparatID,
		CreatedAt:                 fieldDevice.CreatedAt,
		UpdatedAt:                 fieldDevice.UpdatedAt,
	}

	// Include embedded related entities if preloaded
	if fieldDevice.SPSControllerSystemType.ID != uuid.Nil {
		spsSystemType := toSPSControllerSystemTypeResponse(fieldDevice.SPSControllerSystemType)
		resp.SPSControllerSystemType = &spsSystemType
	}

	if fieldDevice.Apparat.ID != uuid.Nil {
		apparat := toApparatResponse(fieldDevice.Apparat)
		resp.Apparat = &apparat
	}

	if fieldDevice.SystemPart.ID != uuid.Nil {
		systemPart := toSystemPartResponse(fieldDevice.SystemPart)
		resp.SystemPart = &systemPart
	}

	if fieldDevice.Specification != nil && fieldDevice.Specification.ID != uuid.Nil {
		specification := toSpecificationResponse(*fieldDevice.Specification)
		resp.Specification = &specification
	}

	// Include BacnetObjects if preloaded
	if len(fieldDevice.BacnetObjects) > 0 {
		resp.BacnetObjects = toBacnetObjectResponses(fieldDevice.BacnetObjects)
	}

	return resp
}

func toFieldDeviceListResponse(list *domain.PaginatedList[domainFacility.FieldDevice]) dto.FieldDeviceListResponse {
	items := make([]dto.FieldDeviceResponse, len(list.Items))
	for i, item := range list.Items {
		items[i] = toFieldDeviceResponse(item)
	}

	return dto.FieldDeviceListResponse{
		Items:      items,
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
	}
}

func toFieldDeviceOptionsResponse(options *domainFacility.FieldDeviceOptions) dto.FieldDeviceOptionsResponse {
	// Convert Apparats
	apparats := make([]dto.ApparatResponse, len(options.Apparats))
	for i, app := range options.Apparats {
		apparats[i] = toApparatResponse(app)
	}

	// Convert SystemParts
	systemParts := make([]dto.SystemPartResponse, len(options.SystemParts))
	for i, sp := range options.SystemParts {
		systemParts[i] = toSystemPartResponse(sp)
	}

	// Convert ObjectDatas
	objectDatas := make([]dto.ObjectDataResponse, len(options.ObjectDatas))
	for i, od := range options.ObjectDatas {
		objectDatas[i] = toObjectDataResponse(od)
	}

	// Convert UUID maps to string maps for JSON compatibility
	apparatToSystemPart := make(map[string][]string)
	for apparatID, systemPartIDs := range options.ApparatToSystemPart {
		strIDs := make([]string, len(systemPartIDs))
		for i, id := range systemPartIDs {
			strIDs[i] = id.String()
		}
		apparatToSystemPart[apparatID.String()] = strIDs
	}

	objectDataToApparat := make(map[string][]string)
	for objectDataID, apparatIDs := range options.ObjectDataToApparat {
		strIDs := make([]string, len(apparatIDs))
		for i, id := range apparatIDs {
			strIDs[i] = id.String()
		}
		objectDataToApparat[objectDataID.String()] = strIDs
	}

	return dto.FieldDeviceOptionsResponse{
		Apparats:            apparats,
		SystemParts:         systemParts,
		ObjectDatas:         objectDatas,
		ApparatToSystemPart: apparatToSystemPart,
		ObjectDataToApparat: objectDataToApparat,
	}
}

func toMultiCreateFieldDeviceResponse(result *domainFacility.FieldDeviceMultiCreateResult) dto.MultiCreateFieldDeviceResponse {
	results := make([]dto.FieldDeviceCreateResultResponse, len(result.Results))
	for i, r := range result.Results {
		var fdResponse *dto.FieldDeviceResponse
		if r.FieldDevice != nil {
			resp := toFieldDeviceResponse(*r.FieldDevice)
			fdResponse = &resp
		}
		results[i] = dto.FieldDeviceCreateResultResponse{
			Index:       r.Index,
			Success:     r.Success,
			FieldDevice: fdResponse,
			Error:       r.Error,
			ErrorField:  r.ErrorField,
		}
	}
	return dto.MultiCreateFieldDeviceResponse{
		Results:       results,
		TotalRequests: result.TotalRequests,
		SuccessCount:  result.SuccessCount,
		FailureCount:  result.FailureCount,
	}
}

func toBacnetObjectResponse(obj domainFacility.BacnetObject) dto.BacnetObjectResponse {
	return dto.BacnetObjectResponse{
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

func toBacnetObjectResponses(objs []domainFacility.BacnetObject) []dto.BacnetObjectResponse {
	items := make([]dto.BacnetObjectResponse, len(objs))
	for i, obj := range objs {
		items[i] = toBacnetObjectResponse(obj)
	}
	return items
}

func toStateTextResponse(stateText domainFacility.StateText) dto.StateTextResponse {
	return dto.StateTextResponse{
		ID:          stateText.ID,
		RefNumber:   stateText.RefNumber,
		StateText1:  stateText.StateText1,
		StateText2:  stateText.StateText2,
		StateText3:  stateText.StateText3,
		StateText4:  stateText.StateText4,
		StateText5:  stateText.StateText5,
		StateText6:  stateText.StateText6,
		StateText7:  stateText.StateText7,
		StateText8:  stateText.StateText8,
		StateText9:  stateText.StateText9,
		StateText10: stateText.StateText10,
		StateText11: stateText.StateText11,
		StateText12: stateText.StateText12,
		StateText13: stateText.StateText13,
		StateText14: stateText.StateText14,
		StateText15: stateText.StateText15,
		StateText16: stateText.StateText16,
		CreatedAt:   stateText.CreatedAt,
		UpdatedAt:   stateText.UpdatedAt,
	}
}

func toStateTextListResponse(list *domain.PaginatedList[domainFacility.StateText]) dto.StateTextListResponse {
	items := make([]dto.StateTextResponse, len(list.Items))
	for i, item := range list.Items {
		items[i] = toStateTextResponse(item)
	}

	return dto.StateTextListResponse{
		Items:      items,
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
	}
}

func toNotificationClassResponse(notificationClass domainFacility.NotificationClass) dto.NotificationClassResponse {
	return dto.NotificationClassResponse{
		ID:                   notificationClass.ID,
		EventCategory:        notificationClass.EventCategory,
		Nc:                   notificationClass.Nc,
		ObjectDescription:    notificationClass.ObjectDescription,
		InternalDescription:  notificationClass.InternalDescription,
		Meaning:              notificationClass.Meaning,
		AckRequiredNotNormal: notificationClass.AckRequiredNotNormal,
		AckRequiredError:     notificationClass.AckRequiredError,
		AckRequiredNormal:    notificationClass.AckRequiredNormal,
		NormNotNormal:        notificationClass.NormNotNormal,
		NormError:            notificationClass.NormError,
		NormNormal:           notificationClass.NormNormal,
		CreatedAt:            notificationClass.CreatedAt,
		UpdatedAt:            notificationClass.UpdatedAt,
	}
}

func toNotificationClassListResponse(list *domain.PaginatedList[domainFacility.NotificationClass]) dto.NotificationClassListResponse {
	items := make([]dto.NotificationClassResponse, len(list.Items))
	for i, item := range list.Items {
		items[i] = toNotificationClassResponse(item)
	}

	return dto.NotificationClassListResponse{
		Items:      items,
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
	}
}

func toAlarmDefinitionResponse(alarmDefinition domainFacility.AlarmDefinition) dto.AlarmDefinitionResponse {
	return dto.AlarmDefinitionResponse{
		ID:          alarmDefinition.ID,
		Name:        alarmDefinition.Name,
		AlarmNote:   alarmDefinition.AlarmNote,
		AlarmTypeID: alarmDefinition.AlarmTypeID,
		Scope:       alarmDefinition.Scope,
		CreatedAt:   alarmDefinition.CreatedAt,
		UpdatedAt:   alarmDefinition.UpdatedAt,
	}
}

func toAlarmDefinitionListResponse(list *domain.PaginatedList[domainFacility.AlarmDefinition]) dto.AlarmDefinitionListResponse {
	items := make([]dto.AlarmDefinitionResponse, len(list.Items))
	for i, item := range list.Items {
		items[i] = toAlarmDefinitionResponse(item)
	}

	return dto.AlarmDefinitionListResponse{
		Items:      items,
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
	}
}

func toObjectDataResponse(obj domainFacility.ObjectData) dto.ObjectDataResponse {
	bacnetObjects := []domainFacility.BacnetObject{}
	if len(obj.BacnetObjects) > 0 {
		bacnetObjects = make([]domainFacility.BacnetObject, 0, len(obj.BacnetObjects))
		for _, item := range obj.BacnetObjects {
			if item == nil {
				continue
			}
			bacnetObjects = append(bacnetObjects, *item)
		}
	}

	apparats := make([]dto.ApparatResponse, 0)
	if len(obj.Apparats) > 0 {
		apparats = make([]dto.ApparatResponse, 0, len(obj.Apparats))
		for _, item := range obj.Apparats {
			if item == nil {
				continue
			}
			apparats = append(apparats, toApparatResponse(*item))
		}
	}

	return dto.ObjectDataResponse{
		ID:            obj.ID,
		Description:   obj.Description,
		Version:       obj.Version,
		IsActive:      obj.IsActive,
		ProjectID:     obj.ProjectID,
		Apparats:      apparats,
		BacnetObjects: toBacnetObjectResponses(bacnetObjects),
		CreatedAt:     obj.CreatedAt,
		UpdatedAt:     obj.UpdatedAt,
	}
}

func toObjectDataListResponse(list *domain.PaginatedList[domainFacility.ObjectData]) dto.ObjectDataListResponse {
	items := make([]dto.ObjectDataResponse, len(list.Items))
	for i, item := range list.Items {
		items[i] = toObjectDataResponse(item)
	}

	return dto.ObjectDataListResponse{
		Items:      items,
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
	}
}

func toSPSControllerSystemTypeResponse(item domainFacility.SPSControllerSystemType) dto.SPSControllerSystemTypeResponse {
	return dto.SPSControllerSystemTypeResponse{
		ID:                item.ID,
		SPSControllerID:   item.SPSControllerID,
		SystemTypeID:      item.SystemTypeID,
		SPSControllerName: item.SPSController.DeviceName,
		SystemTypeName:    item.SystemType.Name,
		Number:            item.Number,
		DocumentName:      item.DocumentName,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

func toSPSControllerSystemTypeListResponse(list *domain.PaginatedList[domainFacility.SPSControllerSystemType]) dto.SPSControllerSystemTypeListResponse {
	items := make([]dto.SPSControllerSystemTypeResponse, len(list.Items))
	for i, item := range list.Items {
		items[i] = toSPSControllerSystemTypeResponse(item)
	}

	return dto.SPSControllerSystemTypeListResponse{
		Items:      items,
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
	}
}

func toUnitResponse(unit domainFacility.Unit) dto.UnitResponse {
	return dto.UnitResponse{
		ID:     unit.ID,
		Code:   unit.Code,
		Symbol: unit.Symbol,
		Name:   unit.Name,
	}
}

func toAlarmFieldResponse(field domainFacility.AlarmField) dto.AlarmFieldResponse {
	return dto.AlarmFieldResponse{
		ID:              field.ID,
		Key:             field.Key,
		Label:           field.Label,
		DataType:        field.DataType,
		DefaultUnitCode: field.DefaultUnitCode,
	}
}

func toAlarmTypeFieldResponse(atf domainFacility.AlarmTypeField) dto.AlarmTypeFieldResponse {
	r := dto.AlarmTypeFieldResponse{
		ID:               atf.ID,
		AlarmTypeID:      atf.AlarmTypeID,
		AlarmFieldID:     atf.AlarmFieldID,
		DisplayOrder:     atf.DisplayOrder,
		IsRequired:       atf.IsRequired,
		IsUserEditable:   atf.IsUserEditable,
		DefaultValueJSON: atf.DefaultValueJSON,
		ValidationJSON:   atf.ValidationJSON,
		DefaultUnitID:    atf.DefaultUnitID,
		UIGroup:          atf.UIGroup,
		CreatedAt:        atf.CreatedAt,
		UpdatedAt:        atf.UpdatedAt,
	}
	if atf.AlarmField != nil {
		af := toAlarmFieldResponse(*atf.AlarmField)
		r.AlarmField = &af
	}
	if atf.DefaultUnit != nil {
		u := toUnitResponse(*atf.DefaultUnit)
		r.DefaultUnit = &u
	}
	return r
}

func toAlarmTypeResponse(at domainFacility.AlarmType) dto.AlarmTypeResponse {
	fields := make([]dto.AlarmTypeFieldResponse, len(at.Fields))
	for i, f := range at.Fields {
		fields[i] = toAlarmTypeFieldResponse(f)
	}
	return dto.AlarmTypeResponse{
		ID:        at.ID,
		Code:      at.Code,
		Name:      at.Name,
		Fields:    fields,
		CreatedAt: at.CreatedAt,
		UpdatedAt: at.UpdatedAt,
	}
}

func toAlarmTypeListResponse(list *domain.PaginatedList[domainFacility.AlarmType]) dto.AlarmTypeListResponse {
	items := make([]dto.AlarmTypeResponse, len(list.Items))
	for i, item := range list.Items {
		items[i] = toAlarmTypeResponse(item)
	}
	return dto.AlarmTypeListResponse{
		Items:      items,
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
	}
}

func toUnitListResponse(list *domain.PaginatedList[domainFacility.Unit]) dto.UnitListResponse {
	items := make([]dto.UnitResponse, len(list.Items))
	for i, item := range list.Items {
		items[i] = toUnitResponse(item)
	}
	return dto.UnitListResponse{
		Items:      items,
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
	}
}

func toAlarmFieldListResponse(list *domain.PaginatedList[domainFacility.AlarmField]) dto.AlarmFieldListResponse {
	items := make([]dto.AlarmFieldResponse, len(list.Items))
	for i, item := range list.Items {
		items[i] = toAlarmFieldResponse(item)
	}
	return dto.AlarmFieldListResponse{
		Items:      items,
		Total:      list.Total,
		Page:       list.Page,
		TotalPages: list.TotalPages,
	}
}

func toAlarmValueResponse(v domainFacility.BacnetObjectAlarmValue) dto.AlarmValueResponse {
	return dto.AlarmValueResponse{
		ID:               v.ID,
		BacnetObjectID:   v.BacnetObjectID,
		AlarmTypeFieldID: v.AlarmTypeFieldID,
		ValueNumber:      v.ValueNumber,
		ValueInteger:     v.ValueInteger,
		ValueBoolean:     v.ValueBoolean,
		ValueString:      v.ValueString,
		ValueJSON:        v.ValueJSON,
		UnitID:           v.UnitID,
		Source:           v.Source,
		CreatedAt:        v.CreatedAt,
		UpdatedAt:        v.UpdatedAt,
	}
}

func toAlarmValuesResponse(values []domainFacility.BacnetObjectAlarmValue) dto.AlarmValuesResponse {
	items := make([]dto.AlarmValueResponse, len(values))
	for i, v := range values {
		items[i] = toAlarmValueResponse(v)
	}
	return dto.AlarmValuesResponse{Items: items}
}
