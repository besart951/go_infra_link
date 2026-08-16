package fielddevice

import (
	"testing"

	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/google/uuid"
)

func TestCreateItemsMapsSharedFieldDeviceInput(t *testing.T) {
	empty := ""
	projectObjectDataID := uuid.New()
	systemTypeID := uuid.New()
	systemPartID := uuid.New()
	apparatID := uuid.New()

	items := CreateItems([]dto.CreateFieldDeviceRequest{{
		ApparatNr:                 intPointer(8),
		SPSControllerSystemTypeID: systemTypeID,
		SystemPartID:              systemPartID,
		ApparatID:                 apparatID,
		ObjectDataID:              &projectObjectDataID,
		BacnetObjects: []dto.BacnetObjectInput{{
			TextFix:          "AI01",
			TextIndividual:   &empty,
			SoftwareNumber:   9,
			HardwareQuantity: 1,
		}},
	}})

	if len(items) != 1 || items[0].FieldDevice == nil {
		t.Fatalf("expected one mapped field device, got %#v", items)
	}
	if items[0].FieldDevice.ApparatNr != 8 || items[0].FieldDevice.SystemPartID != systemPartID {
		t.Fatalf("expected shared scalar mapping, got %#v", items[0].FieldDevice)
	}
	if items[0].ObjectDataID == nil || *items[0].ObjectDataID != projectObjectDataID {
		t.Fatalf("expected object data link to be preserved")
	}
	if len(items[0].BacnetObjects) != 1 || items[0].BacnetObjects[0].TextIndividual != nil {
		t.Fatalf("expected empty BACnet individual text to normalize to nil, got %#v", items[0].BacnetObjects)
	}
}

func intPointer(value int) *int {
	return &value
}
