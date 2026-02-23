package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/google/uuid"
)

// normalizeTextIndividual converts empty-string pointers to nil so the DB
// always stores either NULL (disabled) or a non-empty value (enabled).
func normalizeTextIndividual(s *string) *string {
	if s != nil && *s == "" {
		return nil
	}
	return s
}

func toBacnetObjectModel(req dto.CreateBacnetObjectRequest) *domainFacility.BacnetObject {
	return &domainFacility.BacnetObject{
		TextFix:             req.TextFix,
		Description:         req.Description,
		GMSVisible:          req.GMSVisible,
		Optional:            req.Optional,
		TextIndividual:      normalizeTextIndividual(req.TextIndividual),
		SoftwareType:        domainFacility.BacnetSoftwareType(req.SoftwareType),
		SoftwareNumber:      uint16(req.SoftwareNumber),
		HardwareType:        domainFacility.BacnetHardwareType(req.HardwareType),
		HardwareQuantity:    uint8(req.HardwareQuantity),
		SoftwareReferenceID: req.SoftwareReferenceID,
		StateTextID:         req.StateTextID,
		NotificationClassID: req.NotificationClassID,
		AlarmDefinitionID:   req.AlarmDefinitionID,
		AlarmTypeID:         req.AlarmTypeID,
	}
}

func applyBacnetObjectPatch(target *domainFacility.BacnetObject, req dto.BacnetObjectPatchInput) {
	if req.TextFix != nil {
		target.TextFix = *req.TextFix
	}
	if req.Description != nil {
		target.Description = req.Description
	}
	if req.GMSVisible != nil {
		target.GMSVisible = *req.GMSVisible
	}
	if req.Optional != nil {
		target.Optional = *req.Optional
	}
	if req.TextIndividual != nil {
		target.TextIndividual = normalizeTextIndividual(req.TextIndividual)
	}
	if req.SoftwareType != nil {
		target.SoftwareType = domainFacility.BacnetSoftwareType(*req.SoftwareType)
	}
	if req.SoftwareNumber != nil {
		target.SoftwareNumber = uint16(*req.SoftwareNumber)
	}
	if req.HardwareType != nil {
		target.HardwareType = domainFacility.BacnetHardwareType(*req.HardwareType)
	}
	if req.HardwareQuantity != nil {
		target.HardwareQuantity = uint8(*req.HardwareQuantity)
	}
	if req.SoftwareReferenceID != nil {
		target.SoftwareReferenceID = req.SoftwareReferenceID
	}
	if req.StateTextID != nil {
		target.StateTextID = req.StateTextID
	}
	if req.NotificationClassID != nil {
		target.NotificationClassID = req.NotificationClassID
	}
	if req.AlarmDefinitionID != nil {
		target.AlarmDefinitionID = req.AlarmDefinitionID
	}
	if req.AlarmTypeID != nil {
		target.AlarmTypeID = req.AlarmTypeID
	}
}

func toBacnetObjectPatches(inputs []dto.BacnetObjectBulkPatchInput) []domainFacility.BacnetObjectPatch {
	items := make([]domainFacility.BacnetObjectPatch, 0, len(inputs))
	for _, input := range inputs {
		patch := domainFacility.BacnetObjectPatch{
			ID:                  input.ID,
			TextFix:             input.TextFix,
			Description:         input.Description,
			GMSVisible:          input.GMSVisible,
			Optional:            input.Optional,
			TextIndividual:      normalizeTextIndividual(input.TextIndividual),
			SoftwareReferenceID: input.SoftwareReferenceID,
			StateTextID:         input.StateTextID,
			NotificationClassID: input.NotificationClassID,
			AlarmDefinitionID:   input.AlarmDefinitionID,
			AlarmTypeID:         input.AlarmTypeID,
		}
		if input.SoftwareType != nil {
			st := domainFacility.BacnetSoftwareType(*input.SoftwareType)
			patch.SoftwareType = &st
		}
		if input.SoftwareNumber != nil {
			num := uint16(*input.SoftwareNumber)
			patch.SoftwareNumber = &num
		}
		if input.HardwareType != nil {
			ht := domainFacility.BacnetHardwareType(*input.HardwareType)
			patch.HardwareType = &ht
		}
		if input.HardwareQuantity != nil {
			qty := uint8(*input.HardwareQuantity)
			patch.HardwareQuantity = &qty
		}
		items = append(items, patch)
	}
	return items
}

func toSPSControllerModel(req dto.CreateSPSControllerRequest) *domainFacility.SPSController {
	return &domainFacility.SPSController{
		ControlCabinetID:  req.ControlCabinetID,
		GADevice:          req.GADevice,
		DeviceName:        req.DeviceName,
		DeviceDescription: req.DeviceDescription,
		DeviceLocation:    req.DeviceLocation,
		IPAddress:         req.IPAddress,
		Subnet:            req.Subnet,
		Gateway:           req.Gateway,
		Vlan:              req.Vlan,
	}
}

func applySPSControllerUpdate(target *domainFacility.SPSController, req dto.UpdateSPSControllerRequest) {
	if req.ControlCabinetID != uuid.Nil {
		target.ControlCabinetID = req.ControlCabinetID
	}
	if req.GADevice != nil {
		target.GADevice = req.GADevice
	}
	if req.DeviceName != "" {
		target.DeviceName = req.DeviceName
	}
	if req.DeviceDescription != nil {
		target.DeviceDescription = req.DeviceDescription
	}
	if req.DeviceLocation != nil {
		target.DeviceLocation = req.DeviceLocation
	}
	if req.IPAddress != nil {
		target.IPAddress = req.IPAddress
	}
	if req.Subnet != nil {
		target.Subnet = req.Subnet
	}
	if req.Gateway != nil {
		target.Gateway = req.Gateway
	}
	if req.Vlan != nil {
		target.Vlan = req.Vlan
	}
}

func toSPSControllerSystemTypes(inputs []dto.SPSControllerSystemTypeInput) []domainFacility.SPSControllerSystemType {
	items := make([]domainFacility.SPSControllerSystemType, 0, len(inputs))
	for _, item := range inputs {
		items = append(items, domainFacility.SPSControllerSystemType{
			SystemTypeID: item.SystemTypeID,
			Number:       item.Number,
			DocumentName: item.DocumentName,
		})
	}
	return items
}

func toFieldDeviceModel(req dto.CreateFieldDeviceRequest) *domainFacility.FieldDevice {
	var apparatNr int
	if req.ApparatNr != nil {
		apparatNr = *req.ApparatNr
	}
	systemPartID := req.SystemPartID

	return &domainFacility.FieldDevice{
		BMK:                       req.BMK,
		Description:               req.Description,
		ApparatNr:                 apparatNr,
		SPSControllerSystemTypeID: req.SPSControllerSystemTypeID,
		SystemPartID:              systemPartID,
		ApparatID:                 req.ApparatID,
	}
}

func applyFieldDeviceUpdate(target *domainFacility.FieldDevice, req dto.UpdateFieldDeviceRequest) {
	if req.BMK != nil {
		target.BMK = req.BMK
	}
	if req.Description != nil {
		target.Description = req.Description
	}
	if req.ApparatNr != nil {
		target.ApparatNr = *req.ApparatNr
	}
	if req.SPSControllerSystemTypeID != uuid.Nil {
		target.SPSControllerSystemTypeID = req.SPSControllerSystemTypeID
	}
	if req.SystemPartID != uuid.Nil {
		target.SystemPartID = req.SystemPartID
	}
	if req.ApparatID != uuid.Nil {
		target.ApparatID = req.ApparatID
	}
}

func toFieldDeviceBacnetObjects(inputs []dto.BacnetObjectInput) []domainFacility.BacnetObject {
	items := make([]domainFacility.BacnetObject, 0, len(inputs))
	for _, bo := range inputs {
		items = append(items, domainFacility.BacnetObject{
			TextFix:             bo.TextFix,
			Description:         bo.Description,
			GMSVisible:          bo.GMSVisible,
			Optional:            bo.Optional,
			TextIndividual:      normalizeTextIndividual(bo.TextIndividual),
			SoftwareType:        domainFacility.BacnetSoftwareType(bo.SoftwareType),
			SoftwareNumber:      uint16(bo.SoftwareNumber),
			HardwareType:        domainFacility.BacnetHardwareType(bo.HardwareType),
			HardwareQuantity:    uint8(bo.HardwareQuantity),
			SoftwareReferenceID: bo.SoftwareReferenceID,
			StateTextID:         bo.StateTextID,
			NotificationClassID: bo.NotificationClassID,
			AlarmDefinitionID:   bo.AlarmDefinitionID,
			AlarmTypeID:         bo.AlarmTypeID,
		})
	}
	return items
}

func toFieldDeviceSpecification(req dto.CreateFieldDeviceSpecificationRequest) *domainFacility.Specification {
	return &domainFacility.Specification{
		SpecificationSupplier:                     req.SpecificationSupplier,
		SpecificationBrand:                        req.SpecificationBrand,
		SpecificationType:                         req.SpecificationType,
		AdditionalInfoMotorValve:                  req.AdditionalInfoMotorValve,
		AdditionalInfoSize:                        req.AdditionalInfoSize,
		AdditionalInformationInstallationLocation: req.AdditionalInformationInstallationLocation,
		ElectricalConnectionPH:                    req.ElectricalConnectionPH,
		ElectricalConnectionACDC:                  req.ElectricalConnectionACDC,
		ElectricalConnectionAmperage:              req.ElectricalConnectionAmperage,
		ElectricalConnectionPower:                 req.ElectricalConnectionPower,
		ElectricalConnectionRotation:              req.ElectricalConnectionRotation,
	}
}

func toFieldDeviceSpecificationPatch(req dto.UpdateFieldDeviceSpecificationRequest) *domainFacility.Specification {
	return &domainFacility.Specification{
		SpecificationSupplier:                     req.SpecificationSupplier,
		SpecificationBrand:                        req.SpecificationBrand,
		SpecificationType:                         req.SpecificationType,
		AdditionalInfoMotorValve:                  req.AdditionalInfoMotorValve,
		AdditionalInfoSize:                        req.AdditionalInfoSize,
		AdditionalInformationInstallationLocation: req.AdditionalInformationInstallationLocation,
		ElectricalConnectionPH:                    req.ElectricalConnectionPH,
		ElectricalConnectionACDC:                  req.ElectricalConnectionACDC,
		ElectricalConnectionAmperage:              req.ElectricalConnectionAmperage,
		ElectricalConnectionPower:                 req.ElectricalConnectionPower,
		ElectricalConnectionRotation:              req.ElectricalConnectionRotation,
	}
}

// toSpecificationFromInput converts a SpecificationInput DTO to a domain Specification.
// Used for bulk update operations with nested specification data.
func toSpecificationFromInput(input *dto.SpecificationInput) *domainFacility.Specification {
	if input == nil {
		return nil
	}
	return &domainFacility.Specification{
		SpecificationSupplier:                     input.SpecificationSupplier,
		SpecificationBrand:                        input.SpecificationBrand,
		SpecificationType:                         input.SpecificationType,
		AdditionalInfoMotorValve:                  input.AdditionalInfoMotorValve,
		AdditionalInfoSize:                        input.AdditionalInfoSize,
		AdditionalInformationInstallationLocation: input.AdditionalInformationInstallationLocation,
		ElectricalConnectionPH:                    input.ElectricalConnectionPH,
		ElectricalConnectionACDC:                  input.ElectricalConnectionACDC,
		ElectricalConnectionAmperage:              input.ElectricalConnectionAmperage,
		ElectricalConnectionPower:                 input.ElectricalConnectionPower,
		ElectricalConnectionRotation:              input.ElectricalConnectionRotation,
	}
}

func toBuildingModel(req dto.CreateBuildingRequest) *domainFacility.Building {
	return &domainFacility.Building{
		IWSCode:       req.IWSCode,
		BuildingGroup: req.BuildingGroup,
	}
}

func applyBuildingUpdate(target *domainFacility.Building, req dto.UpdateBuildingRequest) {
	if req.IWSCode != "" {
		target.IWSCode = req.IWSCode
	}
	if req.BuildingGroup != 0 {
		target.BuildingGroup = req.BuildingGroup
	}
}

func toSystemTypeModel(req dto.CreateSystemTypeRequest) *domainFacility.SystemType {
	return &domainFacility.SystemType{
		NumberMin: req.NumberMin,
		NumberMax: req.NumberMax,
		Name:      req.Name,
	}
}

func applySystemTypeUpdate(target *domainFacility.SystemType, req dto.UpdateSystemTypeRequest) {
	if req.NumberMin != 0 {
		target.NumberMin = req.NumberMin
	}
	if req.NumberMax != 0 {
		target.NumberMax = req.NumberMax
	}
	if req.Name != "" {
		target.Name = req.Name
	}
}

func toSystemPartModel(req dto.CreateSystemPartRequest) *domainFacility.SystemPart {
	return &domainFacility.SystemPart{
		ShortName:   req.ShortName,
		Name:        req.Name,
		Description: req.Description,
	}
}

func applySystemPartUpdate(target *domainFacility.SystemPart, req dto.UpdateSystemPartRequest) {
	if req.ShortName != "" {
		target.ShortName = req.ShortName
	}
	if req.Name != "" {
		target.Name = req.Name
	}
	if req.Description != nil {
		target.Description = req.Description
	}
}

func toApparatModel(req dto.CreateApparatRequest, systemParts []*domainFacility.SystemPart) *domainFacility.Apparat {
	return &domainFacility.Apparat{
		ShortName:   req.ShortName,
		Name:        req.Name,
		Description: req.Description,
		SystemParts: systemParts,
	}
}

func applyApparatUpdate(target *domainFacility.Apparat, req dto.UpdateApparatRequest, systemParts *[]*domainFacility.SystemPart) {
	if req.ShortName != "" {
		target.ShortName = req.ShortName
	}
	if req.Name != "" {
		target.Name = req.Name
	}
	if req.Description != nil {
		target.Description = req.Description
	}
	if systemParts != nil {
		target.SystemParts = *systemParts
	}
}

func toObjectDataModel(req dto.CreateObjectDataRequest) *domainFacility.ObjectData {
	obj := &domainFacility.ObjectData{
		Description: req.Description,
		Version:     req.Version,
		ProjectID:   req.ProjectID,
	}
	if req.IsActive != nil {
		obj.IsActive = *req.IsActive
	}
	return obj
}

func applyObjectDataUpdate(target *domainFacility.ObjectData, req dto.UpdateObjectDataRequest) {
	if req.Description != nil {
		target.Description = *req.Description
	}
	if req.Version != nil {
		target.Version = *req.Version
	}
	if req.IsActive != nil {
		target.IsActive = *req.IsActive
	}
	if req.ProjectID != nil {
		target.ProjectID = req.ProjectID
	}
}

func toStateTextModel(req dto.CreateStateTextRequest) *domainFacility.StateText {
	return &domainFacility.StateText{
		RefNumber:   req.RefNumber,
		StateText1:  req.StateText1,
		StateText2:  req.StateText2,
		StateText3:  req.StateText3,
		StateText4:  req.StateText4,
		StateText5:  req.StateText5,
		StateText6:  req.StateText6,
		StateText7:  req.StateText7,
		StateText8:  req.StateText8,
		StateText9:  req.StateText9,
		StateText10: req.StateText10,
		StateText11: req.StateText11,
		StateText12: req.StateText12,
		StateText13: req.StateText13,
		StateText14: req.StateText14,
		StateText15: req.StateText15,
		StateText16: req.StateText16,
	}
}

func applyStateTextUpdate(target *domainFacility.StateText, req dto.UpdateStateTextRequest) {
	if req.RefNumber != nil {
		target.RefNumber = *req.RefNumber
	}
	if req.StateText1 != nil {
		target.StateText1 = req.StateText1
	}
	if req.StateText2 != nil {
		target.StateText2 = req.StateText2
	}
	if req.StateText3 != nil {
		target.StateText3 = req.StateText3
	}
	if req.StateText4 != nil {
		target.StateText4 = req.StateText4
	}
	if req.StateText5 != nil {
		target.StateText5 = req.StateText5
	}
	if req.StateText6 != nil {
		target.StateText6 = req.StateText6
	}
	if req.StateText7 != nil {
		target.StateText7 = req.StateText7
	}
	if req.StateText8 != nil {
		target.StateText8 = req.StateText8
	}
	if req.StateText9 != nil {
		target.StateText9 = req.StateText9
	}
	if req.StateText10 != nil {
		target.StateText10 = req.StateText10
	}
	if req.StateText11 != nil {
		target.StateText11 = req.StateText11
	}
	if req.StateText12 != nil {
		target.StateText12 = req.StateText12
	}
	if req.StateText13 != nil {
		target.StateText13 = req.StateText13
	}
	if req.StateText14 != nil {
		target.StateText14 = req.StateText14
	}
	if req.StateText15 != nil {
		target.StateText15 = req.StateText15
	}
	if req.StateText16 != nil {
		target.StateText16 = req.StateText16
	}
}

func toNotificationClassModel(req dto.CreateNotificationClassRequest) *domainFacility.NotificationClass {
	return &domainFacility.NotificationClass{
		EventCategory:        req.EventCategory,
		Nc:                   req.Nc,
		ObjectDescription:    req.ObjectDescription,
		InternalDescription:  req.InternalDescription,
		Meaning:              req.Meaning,
		AckRequiredNotNormal: req.AckRequiredNotNormal,
		AckRequiredError:     req.AckRequiredError,
		AckRequiredNormal:    req.AckRequiredNormal,
		NormNotNormal:        req.NormNotNormal,
		NormError:            req.NormError,
		NormNormal:           req.NormNormal,
	}
}

func applyNotificationClassUpdate(target *domainFacility.NotificationClass, req dto.UpdateNotificationClassRequest) {
	if req.EventCategory != nil {
		target.EventCategory = *req.EventCategory
	}
	if req.Nc != nil {
		target.Nc = *req.Nc
	}
	if req.ObjectDescription != nil {
		target.ObjectDescription = *req.ObjectDescription
	}
	if req.InternalDescription != nil {
		target.InternalDescription = *req.InternalDescription
	}
	if req.Meaning != nil {
		target.Meaning = *req.Meaning
	}
	if req.AckRequiredNotNormal != nil {
		target.AckRequiredNotNormal = *req.AckRequiredNotNormal
	}
	if req.AckRequiredError != nil {
		target.AckRequiredError = *req.AckRequiredError
	}
	if req.AckRequiredNormal != nil {
		target.AckRequiredNormal = *req.AckRequiredNormal
	}
	if req.NormNotNormal != nil {
		target.NormNotNormal = *req.NormNotNormal
	}
	if req.NormError != nil {
		target.NormError = *req.NormError
	}
	if req.NormNormal != nil {
		target.NormNormal = *req.NormNormal
	}
}

func toAlarmDefinitionModel(req dto.CreateAlarmDefinitionRequest) *domainFacility.AlarmDefinition {
	return &domainFacility.AlarmDefinition{
		Name:        req.Name,
		AlarmNote:   req.AlarmNote,
		AlarmTypeID: req.AlarmTypeID,
	}
}

func applyAlarmDefinitionUpdate(target *domainFacility.AlarmDefinition, req dto.UpdateAlarmDefinitionRequest) {
	if req.Name != nil {
		target.Name = *req.Name
	}
	if req.AlarmNote != nil {
		target.AlarmNote = req.AlarmNote
	}
	if req.AlarmTypeID != nil {
		target.AlarmTypeID = req.AlarmTypeID
	}
}

func toControlCabinetModel(req dto.CreateControlCabinetRequest) *domainFacility.ControlCabinet {
	return &domainFacility.ControlCabinet{
		BuildingID:       req.BuildingID,
		ControlCabinetNr: req.ControlCabinetNr,
	}
}

func applyControlCabinetUpdate(target *domainFacility.ControlCabinet, req dto.UpdateControlCabinetRequest) {
	if req.BuildingID != uuid.Nil {
		target.BuildingID = req.BuildingID
	}
	if req.ControlCabinetNr != nil {
		target.ControlCabinetNr = req.ControlCabinetNr
	}
}

func toUnitModel(req dto.CreateUnitRequest) *domainFacility.Unit {
	return &domainFacility.Unit{
		Code:   req.Code,
		Symbol: req.Symbol,
		Name:   req.Name,
	}
}

func applyUnitUpdate(target *domainFacility.Unit, req dto.UpdateUnitRequest) {
	if req.Code != nil {
		target.Code = *req.Code
	}
	if req.Symbol != nil {
		target.Symbol = *req.Symbol
	}
	if req.Name != nil {
		target.Name = *req.Name
	}
}

func toAlarmFieldModel(req dto.CreateAlarmFieldRequest) *domainFacility.AlarmField {
	return &domainFacility.AlarmField{
		Key:             req.Key,
		Label:           req.Label,
		DataType:        req.DataType,
		DefaultUnitCode: req.DefaultUnitCode,
	}
}

func applyAlarmFieldUpdate(target *domainFacility.AlarmField, req dto.UpdateAlarmFieldRequest) {
	if req.Key != nil {
		target.Key = *req.Key
	}
	if req.Label != nil {
		target.Label = *req.Label
	}
	if req.DataType != nil {
		target.DataType = *req.DataType
	}
	if req.DefaultUnitCode != nil {
		target.DefaultUnitCode = req.DefaultUnitCode
	}
}

func toAlarmTypeFieldModel(alarmTypeID uuid.UUID, req dto.CreateAlarmTypeFieldRequest) *domainFacility.AlarmTypeField {
	return &domainFacility.AlarmTypeField{
		AlarmTypeID:      alarmTypeID,
		AlarmFieldID:     req.AlarmFieldID,
		DisplayOrder:     req.DisplayOrder,
		IsRequired:       req.IsRequired,
		IsUserEditable:   req.IsUserEditable,
		DefaultValueJSON: req.DefaultValueJSON,
		ValidationJSON:   req.ValidationJSON,
		DefaultUnitID:    req.DefaultUnitID,
		UIGroup:          req.UIGroup,
	}
}

func applyAlarmTypeFieldUpdate(target *domainFacility.AlarmTypeField, req dto.UpdateAlarmTypeFieldRequest) {
	if req.DisplayOrder != nil {
		target.DisplayOrder = *req.DisplayOrder
	}
	if req.IsRequired != nil {
		target.IsRequired = *req.IsRequired
	}
	if req.IsUserEditable != nil {
		target.IsUserEditable = *req.IsUserEditable
	}
	if req.DefaultValueJSON != nil {
		target.DefaultValueJSON = req.DefaultValueJSON
	}
	if req.ValidationJSON != nil {
		target.ValidationJSON = req.ValidationJSON
	}
	if req.DefaultUnitID != nil {
		target.DefaultUnitID = req.DefaultUnitID
	}
	if req.UIGroup != nil {
		target.UIGroup = req.UIGroup
	}
}
