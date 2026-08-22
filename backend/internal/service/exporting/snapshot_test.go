package exporting

import (
	"context"
	"os"
	"testing"

	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type snapshotSourceStub struct {
	controller domainExport.Controller
	devices    []domainFacility.FieldDevice
}

func (s snapshotSourceStub) ResolveControllers(context.Context, domainExport.Request) ([]domainExport.Controller, error) {
	return []domainExport.Controller{s.controller}, nil
}

func (s snapshotSourceStub) ListFieldDevicesByControllerAfter(_ context.Context, _ uuid.UUID, _ domainExport.Request, afterID uuid.UUID, _ int) ([]domainFacility.FieldDevice, error) {
	if afterID != uuid.Nil {
		return nil, nil
	}
	return s.devices, nil
}

func (s snapshotSourceStub) WithinSnapshot(_ context.Context, consume func(domainExport.DataProvider) error) error {
	return consume(s)
}

func TestSnapshotManifestCountsAndVerifiesStreamedArtifacts(t *testing.T) {
	controller := domainExport.Controller{ID: uuid.New(), ControlCabinetID: uuid.New()}
	referenceID := uuid.New()
	device := domainFacility.FieldDevice{Specification: &domainFacility.Specification{}}
	device.ID = uuid.New()
	device.BacnetObjects = []domainFacility.BacnetObject{{
		SoftwareReferenceID: &referenceID,
		AlarmValues:         []domainFacility.BacnetObjectAlarmValue{{}},
	}}
	directory := t.TempDir()
	snapshot, err := createOrOpenSnapshot(t.Context(), directory, domainExport.Request{}, snapshotSourceStub{
		controller: controller, devices: []domainFacility.FieldDevice{device},
	}, 500, nil)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	if snapshot.manifest.Counts != (domainExport.Counts{
		FieldDevices: 1, Specifications: 1, BacnetObjects: 1, SoftwareReferences: 1, AlarmValues: 1,
	}) {
		t.Fatalf("unexpected counts: %+v", snapshot.manifest.Counts)
	}
	if err := snapshot.Close(); err != nil {
		t.Fatalf("close snapshot: %v", err)
	}

	path := snapshotControllerPath(directory, controller.ID)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open snapshot artifact: %v", err)
	}
	if _, err := file.WriteString("tampered"); err != nil {
		t.Fatalf("tamper snapshot artifact: %v", err)
	}
	_ = file.Close()
	if _, err := openSnapshot(directory); err == nil {
		t.Fatal("expected checksum mismatch for modified snapshot")
	}
}
