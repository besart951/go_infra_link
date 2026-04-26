package shared

import (
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func TestToControlCabinetResponseCopiesFields(t *testing.T) {
	now := time.Now().UTC()
	id := uuid.New()
	buildingID := uuid.New()
	number := "CC-12"

	response := ToControlCabinetResponse(domainFacility.ControlCabinet{
		Base:             domain.Base{ID: id, CreatedAt: now, UpdatedAt: now},
		BuildingID:       buildingID,
		ControlCabinetNr: &number,
	})

	if response.ID != id {
		t.Fatalf("expected id %s, got %s", id, response.ID)
	}
	if response.BuildingID != buildingID {
		t.Fatalf("expected building id %s, got %s", buildingID, response.BuildingID)
	}
	if response.ControlCabinetNr == nil || *response.ControlCabinetNr != number {
		t.Fatalf("expected control cabinet number %q, got %#v", number, response.ControlCabinetNr)
	}
	if !response.CreatedAt.Equal(now) || !response.UpdatedAt.Equal(now) {
		t.Fatalf("expected timestamps to be preserved")
	}
}

func TestToSPSControllerResponseCopiesFields(t *testing.T) {
	now := time.Now().UTC()
	id := uuid.New()
	controlCabinetID := uuid.New()
	gaDevice := "123"
	deviceDescription := "Primary controller"

	response := ToSPSControllerResponse(domainFacility.SPSController{
		Base:              domain.Base{ID: id, CreatedAt: now, UpdatedAt: now},
		ControlCabinetID:  controlCabinetID,
		GADevice:          &gaDevice,
		DeviceName:        "SPS-A",
		DeviceDescription: &deviceDescription,
	})

	if response.ID != id {
		t.Fatalf("expected id %s, got %s", id, response.ID)
	}
	if response.ControlCabinetID != controlCabinetID {
		t.Fatalf("expected control cabinet id %s, got %s", controlCabinetID, response.ControlCabinetID)
	}
	if response.GADevice == nil || *response.GADevice != gaDevice {
		t.Fatalf("expected GA device %q, got %#v", gaDevice, response.GADevice)
	}
	if response.DeviceName != "SPS-A" {
		t.Fatalf("expected device name SPS-A, got %q", response.DeviceName)
	}
	if response.DeviceDescription == nil || *response.DeviceDescription != deviceDescription {
		t.Fatalf("expected device description %q, got %#v", deviceDescription, response.DeviceDescription)
	}
	if !response.CreatedAt.Equal(now) || !response.UpdatedAt.Equal(now) {
		t.Fatalf("expected timestamps to be preserved")
	}
}

func TestToSPSControllerSystemTypeResponseIncludesFieldDevicesCount(t *testing.T) {
	now := time.Now().UTC()
	id := uuid.New()
	spsControllerID := uuid.New()
	systemTypeID := uuid.New()
	number := 8
	documentName := "cooling.pdf"

	response := ToSPSControllerSystemTypeResponse(domainFacility.SPSControllerSystemType{
		Base:              domain.Base{ID: id, CreatedAt: now, UpdatedAt: now},
		SPSControllerID:   spsControllerID,
		SystemTypeID:      systemTypeID,
		SPSController:     domainFacility.SPSController{DeviceName: "SPS-A"},
		SystemType:        domainFacility.SystemType{Name: "Cooling"},
		Number:            &number,
		DocumentName:      &documentName,
		FieldDevicesCount: 7,
	})

	if response.FieldDevicesCount != 7 {
		t.Fatalf("expected field device count 7, got %d", response.FieldDevicesCount)
	}
	if response.ID != id || response.SPSControllerID != spsControllerID || response.SystemTypeID != systemTypeID {
		t.Fatalf("expected identifiers to be preserved")
	}
	if response.SPSControllerName != "SPS-A" || response.SystemTypeName != "Cooling" {
		t.Fatalf("expected display names to be preserved")
	}
	if response.Number == nil || *response.Number != number {
		t.Fatalf("expected number %d, got %#v", number, response.Number)
	}
	if response.DocumentName == nil || *response.DocumentName != documentName {
		t.Fatalf("expected document name %q, got %#v", documentName, response.DocumentName)
	}
	if !response.CreatedAt.Equal(now) || !response.UpdatedAt.Equal(now) {
		t.Fatalf("expected timestamps to be preserved")
	}
}

func TestToFieldDeviceOptionsResponsePreservesNestedRelationships(t *testing.T) {
	now := time.Now().UTC()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	objectDataID := uuid.New()
	bacnetObjectID := uuid.New()
	projectID := uuid.New()
	description := "Supply air"
	textIndividual := "Zone 1"

	response := ToFieldDeviceOptionsResponse(&domainFacility.FieldDeviceOptions{
		Apparats: []domainFacility.Apparat{{
			Base:        domain.Base{ID: apparatID, CreatedAt: now, UpdatedAt: now},
			ShortName:   "APP",
			Name:        "Apparat",
			Description: &description,
			SystemParts: []*domainFacility.SystemPart{{
				Base:        domain.Base{ID: systemPartID, CreatedAt: now, UpdatedAt: now},
				ShortName:   "SYS",
				Name:        "System Part",
				Description: &description,
			}},
		}},
		SystemParts: []domainFacility.SystemPart{{
			Base:        domain.Base{ID: systemPartID, CreatedAt: now, UpdatedAt: now},
			ShortName:   "SYS",
			Name:        "System Part",
			Description: &description,
		}},
		ObjectDatas: []domainFacility.ObjectData{{
			Base:        domain.Base{ID: objectDataID, CreatedAt: now, UpdatedAt: now},
			Description: "Object Data",
			Version:     "v1",
			IsActive:    true,
			ProjectID:   &projectID,
			Apparats: []*domainFacility.Apparat{{
				Base:      domain.Base{ID: apparatID, CreatedAt: now, UpdatedAt: now},
				ShortName: "APP",
				Name:      "Apparat",
			}},
			BacnetObjects: []*domainFacility.BacnetObject{{
				Base:             domain.Base{ID: bacnetObjectID, CreatedAt: now, UpdatedAt: now},
				TextFix:          "TF",
				Description:      &description,
				GMSVisible:       true,
				Optional:         false,
				TextIndividual:   &textIndividual,
				SoftwareType:     domainFacility.BacnetSoftwareTypeAI,
				SoftwareNumber:   42,
				HardwareType:     domainFacility.BacnetHardwareTypeAO,
				HardwareQuantity: 2,
			}},
		}},
		ApparatToSystemPart: map[uuid.UUID][]uuid.UUID{
			apparatID: {systemPartID},
		},
		ObjectDataToApparat: map[uuid.UUID][]uuid.UUID{
			objectDataID: {apparatID},
		},
	})

	if len(response.Apparats) != 1 || len(response.Apparats[0].SystemParts) != 1 {
		t.Fatalf("expected nested apparat and system part mappings to be preserved")
	}
	if len(response.ObjectDatas) != 1 || len(response.ObjectDatas[0].Apparats) != 1 {
		t.Fatalf("expected object data apparat mappings to be preserved")
	}
	if len(response.ObjectDatas[0].BacnetObjects) != 1 {
		t.Fatalf("expected bacnet object mappings to be preserved")
	}
	if got := response.ApparatToSystemPart[apparatID.String()]; len(got) != 1 || got[0] != systemPartID.String() {
		t.Fatalf("expected apparat to system part map to contain %s -> %s, got %#v", apparatID, systemPartID, got)
	}
	if got := response.ObjectDataToApparat[objectDataID.String()]; len(got) != 1 || got[0] != apparatID.String() {
		t.Fatalf("expected object data to apparat map to contain %s -> %s, got %#v", objectDataID, apparatID, got)
	}
	if response.ObjectDatas[0].BacnetObjects[0].ID != bacnetObjectID.String() {
		t.Fatalf("expected bacnet object id %s, got %s", bacnetObjectID, response.ObjectDatas[0].BacnetObjects[0].ID)
	}
}
