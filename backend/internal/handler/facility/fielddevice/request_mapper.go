package fielddevice

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
)

// CreateItems maps the shared HTTP request contract used by global and
// project-scoped field-device creation to domain input. Keeping the mapping
// here prevents the two entry points from drifting apart.
func CreateItems(requests []dto.CreateFieldDeviceRequest) []domainFacility.FieldDeviceCreateItem {
	items := make([]domainFacility.FieldDeviceCreateItem, len(requests))
	for i, request := range requests {
		apparatNr := 0
		if request.ApparatNr != nil {
			apparatNr = *request.ApparatNr
		}

		items[i] = domainFacility.FieldDeviceCreateItem{
			FieldDevice: &domainFacility.FieldDevice{
				BMK:                       request.BMK,
				Description:               request.Description,
				TextIndividuell:           request.TextIndividuell,
				ApparatNr:                 apparatNr,
				SPSControllerSystemTypeID: request.SPSControllerSystemTypeID,
				SystemPartID:              request.SystemPartID,
				ApparatID:                 request.ApparatID,
			},
			ObjectDataID:  request.ObjectDataID,
			BacnetObjects: ToBacnetObjects(request.BacnetObjects),
		}
	}
	return items
}

// ToBacnetObjects maps nested field-device BACnet input consistently for all
// creation and template flows.
func ToBacnetObjects(inputs []dto.BacnetObjectInput) []domainFacility.BacnetObject {
	items := make([]domainFacility.BacnetObject, 0, len(inputs))
	for _, input := range inputs {
		items = append(items, domainFacility.BacnetObject{
			TextFix:             input.TextFix,
			Description:         input.Description,
			GMSVisible:          input.GMSVisible,
			Optional:            input.Optional,
			TextIndividual:      NormalizeTextIndividual(input.TextIndividual),
			SoftwareType:        domainFacility.BacnetSoftwareType(input.SoftwareType),
			SoftwareNumber:      uint16(input.SoftwareNumber),
			HardwareType:        domainFacility.BacnetHardwareType(input.HardwareType),
			HardwareQuantity:    uint8(input.HardwareQuantity),
			SoftwareReferenceID: input.SoftwareReferenceID,
			StateTextID:         input.StateTextID,
			NotificationClassID: input.NotificationClassID,
			AlarmDefinitionID:   input.AlarmDefinitionID,
			AlarmTypeID:         input.AlarmTypeID,
		})
	}
	return items
}

// NormalizeTextIndividual stores disabled individual text as NULL instead of
// an empty string. It applies to every BACnet input nested under a field device.
func NormalizeTextIndividual(value *string) *string {
	if value != nil && *value == "" {
		return nil
	}
	return value
}
