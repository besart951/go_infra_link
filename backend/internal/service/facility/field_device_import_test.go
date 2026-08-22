package facility_test

import (
	"context"
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

func TestFieldDeviceImportWritesOwnedGraphInOneTransaction(t *testing.T) {
	fixture := newImportServiceFixture()
	runnerCalls := 0
	services := newTxServices(fixture.baseRepositories(), fixture.txRepositories(), &runnerCalls)

	err := services.FieldDevice.ImportAggregate(context.Background(), fixture.aggregate())

	if err != nil {
		t.Fatal(err)
	}
	if runnerCalls != 1 || len(fixture.baseDevices.items) != 0 || len(fixture.txDevices.items) != 1 {
		t.Fatalf("transaction boundary mismatch: calls=%d base=%d tx=%d", runnerCalls, len(fixture.baseDevices.items), len(fixture.txDevices.items))
	}
	if len(fixture.txObjects.items) != 2 || len(fixture.txValues.items) != 1 || len(fixture.txSpecs.items) != 1 {
		t.Fatalf("owned graph incomplete: objects=%d values=%d specs=%d", len(fixture.txObjects.items), len(fixture.txValues.items), len(fixture.txSpecs.items))
	}
	for _, object := range fixture.txObjects.items {
		if object.SoftwareReferenceID != nil && *object.SoftwareReferenceID != fixture.targetObjectID {
			t.Fatalf("software reference changed: %s", *object.SoftwareReferenceID)
		}
	}
}

func TestFieldDeviceImportFailureDoesNotEscapeAggregateTransaction(t *testing.T) {
	fixture := newImportServiceFixture()
	failure := errors.New("alarm values unavailable")
	fixture.txValues.failBulkCreate = failure
	runnerCalls := 0
	services := newTxServices(fixture.baseRepositories(), fixture.txRepositories(), &runnerCalls)

	err := services.FieldDevice.ImportAggregate(context.Background(), fixture.aggregate())

	if !errors.Is(err, failure) || runnerCalls != 1 || len(fixture.baseDevices.items) != 0 {
		t.Fatalf("failure escaped transaction: err=%v calls=%d base=%d", err, runnerCalls, len(fixture.baseDevices.items))
	}
}

type importServiceFixture struct {
	deviceID, assignmentID, systemTypeID uuid.UUID
	apparatID, systemPartID              uuid.UUID
	sourceObjectID, targetObjectID       uuid.UUID
	alarmFieldID                         uuid.UUID
	baseDevices                          *fakeFieldDeviceStore
	txDevices                            *fakeFieldDeviceStore
	txSpecs                              *fakeSpecificationStore
	txObjects                            *fakeBacnetObjectStore
	txValues                             *fakeBacnetObjectAlarmValueRepo
	assignments                          *fakeSpsControllerSystemTypeRepo
	systemTypes                          *fakeSystemTypeRepo
	apparats                             *fakeApparatRepo
	systemParts                          *fakeSystemPartRepo
}

func newImportServiceFixture() *importServiceFixture {
	fixture := &importServiceFixture{
		deviceID: uuid.New(), assignmentID: uuid.New(), systemTypeID: uuid.New(),
		apparatID: uuid.New(), systemPartID: uuid.New(), sourceObjectID: uuid.New(), targetObjectID: uuid.New(), alarmFieldID: uuid.New(),
		baseDevices: &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{}},
		txDevices:   &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{}},
		txSpecs:     &fakeSpecificationStore{items: map[uuid.UUID]*domainFacility.Specification{}},
		txObjects:   &fakeBacnetObjectStore{items: map[uuid.UUID]*domainFacility.BacnetObject{}},
		txValues:    &fakeBacnetObjectAlarmValueRepo{items: map[uuid.UUID]*domainFacility.BacnetObjectAlarmValue{}},
	}
	fixture.assignments = &fakeSpsControllerSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
		fixture.assignmentID: {Base: domain.Base{ID: fixture.assignmentID}, SystemTypeID: fixture.systemTypeID},
	}}
	fixture.systemTypes = &fakeSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SystemType{
		fixture.systemTypeID: {Base: domain.Base{ID: fixture.systemTypeID}, NumberMin: 1, NumberMax: 99},
	}}
	fixture.apparats = &fakeApparatRepo{items: map[uuid.UUID]*domainFacility.Apparat{
		fixture.apparatID: {Base: domain.Base{ID: fixture.apparatID}, ShortName: "PMP", Name: "Pump"},
	}}
	fixture.systemParts = &fakeSystemPartRepo{items: map[uuid.UUID]*domainFacility.SystemPart{
		fixture.systemPartID: {Base: domain.Base{ID: fixture.systemPartID}, ShortName: "AIR", Name: "Air"},
	}}
	return fixture
}

func (f *importServiceFixture) baseRepositories() facility.Repositories {
	return facility.Repositories{
		FieldDevices: f.baseDevices, SPSControllerSystemTypes: f.assignments,
		SystemTypes: f.systemTypes, Apparats: f.apparats, SystemParts: f.systemParts,
	}
}

func (f *importServiceFixture) txRepositories() facility.Repositories {
	return facility.Repositories{
		FieldDevices: f.txDevices, SPSControllerSystemTypes: f.assignments,
		SystemTypes: f.systemTypes, Apparats: f.apparats, SystemParts: f.systemParts,
		Specifications: f.txSpecs, BacnetObjects: f.txObjects, BacnetObjectAlarmValues: f.txValues,
	}
}

func (f *importServiceFixture) aggregate() domainFacility.FieldDeviceImportAggregate {
	device := newFieldDevice(f.deviceID, f.assignmentID, f.apparatID, f.systemPartID, 4)
	objects := []domainFacility.BacnetObject{
		{Base: domain.Base{ID: f.sourceObjectID, Version: 2}, FieldDeviceID: &f.deviceID, TextFix: "Source", SoftwareType: domainFacility.BacnetSoftwareTypeAI, SoftwareReferenceID: &f.targetObjectID,
			AlarmValues: []domainFacility.BacnetObjectAlarmValue{{Base: domain.Base{ID: uuid.New(), Version: 1}, BacnetObjectID: f.sourceObjectID, AlarmTypeFieldID: f.alarmFieldID}}},
		{Base: domain.Base{ID: f.targetObjectID, Version: 1}, FieldDeviceID: &f.deviceID, TextFix: "Target", SoftwareType: domainFacility.BacnetSoftwareTypeAV},
	}
	return domainFacility.FieldDeviceImportAggregate{
		FieldDevice:   *device,
		Specification: &domainFacility.Specification{Base: domain.Base{ID: uuid.New(), Version: 1}, FieldDeviceID: &f.deviceID},
		BacnetObjects: objects,
	}
}
