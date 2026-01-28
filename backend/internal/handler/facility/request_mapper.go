package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/google/uuid"
)

func toBacnetObjectModel(req dto.CreateBacnetObjectRequest) *domainFacility.BacnetObject {
	return &domainFacility.BacnetObject{
		TextFix:             req.TextFix,
		Description:         req.Description,
		GMSVisible:          req.GMSVisible,
		Optional:            req.Optional,
		TextIndividual:      req.TextIndividual,
		SoftwareType:        domainFacility.BacnetSoftwareType(req.SoftwareType),
		SoftwareNumber:      uint16(req.SoftwareNumber),
		HardwareType:        domainFacility.BacnetHardwareType(req.HardwareType),
		HardwareQuantity:    uint8(req.HardwareQuantity),
		SoftwareReferenceID: req.SoftwareReferenceID,
		StateTextID:         req.StateTextID,
		NotificationClassID: req.NotificationClassID,
		AlarmDefinitionID:   req.AlarmDefinitionID,
	}
}

func applyBacnetObjectUpdate(target *domainFacility.BacnetObject, req dto.UpdateBacnetObjectRequest) {
	target.TextFix = req.TextFix
	target.Description = req.Description
	target.GMSVisible = req.GMSVisible
	target.Optional = req.Optional
	target.TextIndividual = req.TextIndividual
	target.SoftwareType = domainFacility.BacnetSoftwareType(req.SoftwareType)
	target.SoftwareNumber = uint16(req.SoftwareNumber)
	target.HardwareType = domainFacility.BacnetHardwareType(req.HardwareType)
	target.HardwareQuantity = uint8(req.HardwareQuantity)
	target.SoftwareReferenceID = req.SoftwareReferenceID
	target.StateTextID = req.StateTextID
	target.NotificationClassID = req.NotificationClassID
	target.AlarmDefinitionID = req.AlarmDefinitionID
	if req.FieldDeviceID != nil {
		target.FieldDeviceID = req.FieldDeviceID
	}
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
			TextIndividual:      bo.TextIndividual,
			SoftwareType:        domainFacility.BacnetSoftwareType(bo.SoftwareType),
			SoftwareNumber:      uint16(bo.SoftwareNumber),
			HardwareType:        domainFacility.BacnetHardwareType(bo.HardwareType),
			HardwareQuantity:    uint8(bo.HardwareQuantity),
			SoftwareReferenceID: bo.SoftwareReferenceID,
			StateTextID:         bo.StateTextID,
			NotificationClassID: bo.NotificationClassID,
			AlarmDefinitionID:   bo.AlarmDefinitionID,
		})
	}
	return items
}

func toSpecificationModel(req dto.CreateSpecificationRequest) *domainFacility.Specification {
	return &domainFacility.Specification{
		FieldDeviceID:                             &req.FieldDeviceID,
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

func applySpecificationUpdate(target *domainFacility.Specification, req dto.UpdateSpecificationRequest) {
	if req.SpecificationSupplier != nil {
		target.SpecificationSupplier = req.SpecificationSupplier
	}
	if req.SpecificationBrand != nil {
		target.SpecificationBrand = req.SpecificationBrand
	}
	if req.SpecificationType != nil {
		target.SpecificationType = req.SpecificationType
	}
	if req.AdditionalInfoMotorValve != nil {
		target.AdditionalInfoMotorValve = req.AdditionalInfoMotorValve
	}
	if req.AdditionalInfoSize != nil {
		target.AdditionalInfoSize = req.AdditionalInfoSize
	}
	if req.AdditionalInformationInstallationLocation != nil {
		target.AdditionalInformationInstallationLocation = req.AdditionalInformationInstallationLocation
	}
	if req.ElectricalConnectionPH != nil {
		target.ElectricalConnectionPH = req.ElectricalConnectionPH
	}
	if req.ElectricalConnectionACDC != nil {
		target.ElectricalConnectionACDC = req.ElectricalConnectionACDC
	}
	if req.ElectricalConnectionAmperage != nil {
		target.ElectricalConnectionAmperage = req.ElectricalConnectionAmperage
	}
	if req.ElectricalConnectionPower != nil {
		target.ElectricalConnectionPower = req.ElectricalConnectionPower
	}
	if req.ElectricalConnectionRotation != nil {
		target.ElectricalConnectionRotation = req.ElectricalConnectionRotation
	}
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

func toApparatModel(req dto.CreateApparatRequest) *domainFacility.Apparat {
	return &domainFacility.Apparat{
		ShortName:   req.ShortName,
		Name:        req.Name,
		Description: req.Description,
	}
}

func applyApparatUpdate(target *domainFacility.Apparat, req dto.UpdateApparatRequest) {
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
