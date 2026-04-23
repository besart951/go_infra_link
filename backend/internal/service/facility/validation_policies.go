package facility

import "github.com/besart951/go_infra_link/backend/internal/domain"

var (
	buildingIWSCodeField              = domain.ValidationField{Key: "building.iws_code", Name: "iws_code"}
	buildingGroupField                = domain.ValidationField{Key: "building.building_group", Name: "building_group"}
	controlCabinetBuildingIDField     = domain.ValidationField{Key: "controlcabinet.building_id", Name: "building_id"}
	controlCabinetNumberField         = domain.ValidationField{Key: "controlcabinet.control_cabinet_nr", Name: "control_cabinet_nr"}
	systemTypeNameField               = domain.ValidationField{Key: "systemtype.name", Name: "name"}
	systemTypeNumberMinField          = domain.ValidationField{Key: "systemtype.number_min", Name: "number_min"}
	systemTypeNumberMaxField          = domain.ValidationField{Key: "systemtype.number_max", Name: "number_max"}
	systemPartShortNameField          = domain.ValidationField{Key: "system_part.short_name", Name: "short_name"}
	systemPartNameField               = domain.ValidationField{Key: "system_part.name", Name: "name"}
	apparatShortNameField             = domain.ValidationField{Key: "apparat.short_name", Name: "short_name"}
	apparatNameField                  = domain.ValidationField{Key: "apparat.name", Name: "name"}
	spsControllerControlCabinetField  = domain.ValidationField{Key: "spscontroller.control_cabinet_id", Name: "control_cabinet_id"}
	spsControllerDeviceNameField      = domain.ValidationField{Key: "spscontroller.device_name", Name: "device_name"}
	spsControllerGADeviceField        = domain.ValidationField{Key: "spscontroller.ga_device", Name: "ga_device"}
	spsControllerIPAddressField       = domain.ValidationField{Key: "spscontroller.ip_address", Name: "ip_address"}
	spsControllerVlanField            = domain.ValidationField{Key: "spscontroller.vlan", Name: "vlan"}
	spsControllerSystemTypesField     = domain.ValidationField{Key: "spscontroller.system_types", Name: "system_types"}
	fieldDeviceSystemTypeIDField      = domain.ValidationField{Key: "fielddevice.sps_controller_system_type_id", Name: "sps_controller_system_type_id"}
	fieldDeviceApparatIDField         = domain.ValidationField{Key: "fielddevice.apparat_id", Name: "apparat_id"}
	fieldDeviceSystemPartIDField      = domain.ValidationField{Key: "fielddevice.system_part_id", Name: "system_part_id"}
	fieldDeviceApparatNrField         = domain.ValidationField{Key: "fielddevice.apparat_nr", Name: "apparat_nr"}
	fieldDeviceBMKField               = domain.ValidationField{Key: "fielddevice.bmk", Name: "bmk"}
	fieldDeviceDescriptionField       = domain.ValidationField{Key: "fielddevice.description", Name: "description"}
	fieldDeviceTextFixField           = domain.ValidationField{Key: "fielddevice.text_fix", Name: "text_fix"}
	specificationSupplierField        = domain.ValidationField{Key: "specification.specification_supplier", Name: "specification_supplier"}
	specificationBrandField           = domain.ValidationField{Key: "specification.specification_brand", Name: "specification_brand"}
	specificationTypeField            = domain.ValidationField{Key: "specification.specification_type", Name: "specification_type"}
	specificationMotorValveInfoField  = domain.ValidationField{Key: "specification.additional_info_motor_valve", Name: "additional_info_motor_valve"}
	specificationInstallLocationField = domain.ValidationField{Key: "specification.additional_information_installation_location", Name: "additional_information_installation_location"}
	specificationElectricalACDCField  = domain.ValidationField{Key: "specification.electrical_connection_acdc", Name: "electrical_connection_acdc"}
)

const (
	buildingGroupScope  = "the building group"
	buildingScope       = "the building"
	controlCabinetScope = "the control cabinet"
)
