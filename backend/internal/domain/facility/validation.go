package facility

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

func (c ControlCabinet) Validate() error {
	validation := domain.NewValidationError()
	requireUUID(validation, "controlcabinet.building_id", c.BuildingID)
	requireStringPtr(validation, "controlcabinet.control_cabinet_nr", c.ControlCabinetNr, 11)
	return validationResult(validation)
}

func (c SPSController) Validate() error {
	validation := domain.NewValidationError()
	requireUUID(validation, "spscontroller.control_cabinet_id", c.ControlCabinetID)
	requireString(validation, "spscontroller.device_name", c.DeviceName, 100)

	if c.GADevice == nil || strings.TrimSpace(*c.GADevice) == "" {
		validation.AddCode("spscontroller.ga_device", "required", "ga_device is required")
	} else if !validGADevice(*c.GADevice) {
		validation.AddCode("spscontroller.ga_device", "format", "ga_device must be exactly 3 uppercase letters (A-Z)")
	}
	optionalStringMax(validation, "spscontroller.device_description", c.DeviceDescription, 250)
	optionalStringMax(validation, "spscontroller.device_location", c.DeviceLocation, 250)
	validateIPv4(validation, "spscontroller.ip_address", c.IPAddress)
	validateIPv4(validation, "spscontroller.gateway", c.Gateway)
	validateSubnet(validation, c.Subnet)
	validateVLAN(validation, c.Vlan)
	return validationResult(validation)
}

func (s SPSControllerSystemType) Validate(path string) error {
	if path == "" {
		path = "spscontroller.system_types"
	}
	validation := domain.NewValidationError()
	requireUUID(validation, path+".sps_controller_id", s.SPSControllerID)
	requireUUID(validation, path+".system_type_id", s.SystemTypeID)
	if s.Number != nil && *s.Number <= 0 {
		validation.AddCode(path+".number", "min", "number must be greater than zero")
	}
	optionalStringMax(validation, path+".document_name", s.DocumentName, 250)
	return validationResult(validation)
}

func (f FieldDevice) Validate() error {
	validation := domain.NewValidationError()
	requireUUID(validation, "fielddevice.sps_controller_system_type_id", f.SPSControllerSystemTypeID)
	requireUUID(validation, "fielddevice.system_part_id", f.SystemPartID)
	requireUUID(validation, "fielddevice.apparat_id", f.ApparatID)
	if f.ApparatNr < 1 || f.ApparatNr > 99 {
		validation.AddCode("fielddevice.apparat_nr", "range", "apparat_nr must be between 1 and 99")
	}
	optionalStringMax(validation, "fielddevice.bmk", f.BMK, 10)
	optionalStringMax(validation, "fielddevice.description", f.Description, 250)
	optionalStringMax(validation, "fielddevice.text_fix", f.TextIndividuell, 250)
	return validationResult(validation)
}

func (b BacnetObject) Validate(path string) error {
	if path == "" {
		path = "bacnetobject"
	}
	validation := domain.NewValidationError()
	requireString(validation, path+".text_fix", b.TextFix, 250)
	if !b.SoftwareType.Valid() {
		validation.AddCode(path+".software_type", "oneof", "software_type is invalid")
	}
	if !b.HardwareType.Valid() {
		validation.AddCode(path+".hardware_type", "oneof", "hardware_type is invalid")
	}
	optionalStringMax(validation, path+".text_individual", b.TextIndividual, 250)
	return validationResult(validation)
}

func (t BacnetSoftwareType) Valid() bool {
	switch t {
	case BacnetSoftwareTypeAI, BacnetSoftwareTypeAO, BacnetSoftwareTypeAV,
		BacnetSoftwareTypeBI, BacnetSoftwareTypeBO, BacnetSoftwareTypeBV,
		BacnetSoftwareTypeMI, BacnetSoftwareTypeMO, BacnetSoftwareTypeMV,
		BacnetSoftwareTypeCA, BacnetSoftwareTypeEE, BacnetSoftwareTypeLP,
		BacnetSoftwareTypeNC, BacnetSoftwareTypeSC, BacnetSoftwareTypeTL:
		return true
	default:
		return false
	}
}

func (t BacnetHardwareType) Valid() bool {
	switch t {
	case BacnetHardwareTypeEMPTY, BacnetHardwareTypeDO, BacnetHardwareTypeAO,
		BacnetHardwareTypeDI, BacnetHardwareTypeAI:
		return true
	default:
		return false
	}
}

func validationResult(validation *domain.ValidationError) error {
	if len(validation.Fields) == 0 {
		return nil
	}
	return validation
}

func requireUUID(validation *domain.ValidationError, path string, value uuid.UUID) {
	if value == uuid.Nil {
		validation.AddCode(path, "required", fieldName(path)+" is required")
	}
}

func requireString(validation *domain.ValidationError, path, value string, max int) {
	name := fieldName(path)
	if strings.TrimSpace(value) == "" {
		validation.AddCode(path, "required", name+" is required")
		return
	}
	if len(value) > max {
		validation.AddCode(path, "max", fmt.Sprintf("%s must be at most %d characters", name, max))
	}
}

func requireStringPtr(validation *domain.ValidationError, path string, value *string, max int) {
	if value == nil {
		validation.AddCode(path, "required", fieldName(path)+" is required")
		return
	}
	requireString(validation, path, *value, max)
}

func optionalStringMax(validation *domain.ValidationError, path string, value *string, max int) {
	if value != nil && len(*value) > max {
		validation.AddCode(path, "max", fmt.Sprintf("%s must be at most %d characters", fieldName(path), max))
	}
}

func fieldName(path string) string {
	return path[strings.LastIndex(path, ".")+1:]
}

func validGADevice(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 3 {
		return false
	}
	for _, char := range value {
		if char < 'A' || char > 'Z' {
			return false
		}
	}
	return true
}

func validateIPv4(validation *domain.ValidationError, path string, value *string) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return
	}
	ip := net.ParseIP(strings.TrimSpace(*value))
	if ip == nil || ip.To4() == nil {
		validation.AddCode(path, "ipv4", path[strings.LastIndex(path, ".")+1:]+" must be a valid IPv4 address")
	}
}

func validateSubnet(validation *domain.ValidationError, value *string) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return
	}
	ip := net.ParseIP(strings.TrimSpace(*value))
	if ip == nil || ip.To4() == nil {
		validation.AddCode("spscontroller.subnet", "subnet", "subnet must be a valid IPv4 subnet mask")
		return
	}
	ones, bits := net.IPMask(ip.To4()).Size()
	if bits != 32 || ones <= 0 {
		validation.AddCode("spscontroller.subnet", "subnet", "subnet must be a valid IPv4 subnet mask")
	}
}

func validateVLAN(validation *domain.ValidationError, value *string) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return
	}
	vlan, err := strconv.Atoi(strings.TrimSpace(*value))
	if err != nil || vlan < 1 || vlan > 4094 {
		validation.AddCode("spscontroller.vlan", "range", "vlan must be a number between 1 and 4094")
	}
}
