package facility

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	"github.com/google/uuid"
)

type hierarchyCopyFieldDeviceStore struct {
	domainFieldDevice.FieldDeviceStore
	source          map[uuid.UUID]*domainFacility.FieldDevice
	created         []*domainFacility.FieldDevice
	assignments     map[uuid.UUID]uuid.UUID
	pageLengths     []int
	loadedLengths   []int
	bulkCreateCalls int
	assignmentCalls int
	pageCalls       int
}

func (s *hierarchyCopyFieldDeviceStore) GetByIds(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainFacility.FieldDevice, error) {
	s.loadedLengths = append(s.loadedLengths, len(ids))
	out := make([]*domainFacility.FieldDevice, 0, len(ids))
	for _, id := range ids {
		if item := s.source[id]; item != nil {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (s *hierarchyCopyFieldDeviceStore) ListIDsBySPSControllerSystemTypeIDsAfter(
	_ context.Context,
	systemTypeIDs []uuid.UUID,
	afterID *uuid.UUID,
	limit int,
) ([]uuid.UUID, error) {
	s.pageCalls++
	parents := uuidSetForCopyTest(systemTypeIDs)
	ids := make([]uuid.UUID, 0, len(s.source))
	for id, item := range s.source {
		if item == nil {
			continue
		}
		if _, ok := parents[item.SPSControllerSystemTypeID]; ok {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].String() < ids[j].String()
	})
	out := make([]uuid.UUID, 0, limit)
	for _, id := range ids {
		if afterID != nil && id.String() <= afterID.String() {
			continue
		}
		out = append(out, id)
		if len(out) == limit {
			break
		}
	}
	s.pageLengths = append(s.pageLengths, len(out))
	return out, nil
}

func (s *hierarchyCopyFieldDeviceStore) BulkCreate(
	_ context.Context,
	entities []*domainFacility.FieldDevice,
	_ int,
) error {
	s.bulkCreateCalls++
	for _, entity := range entities {
		entity.ID = uuid.New()
		clone := *entity
		s.created = append(s.created, &clone)
	}
	return nil
}

func (s *hierarchyCopyFieldDeviceStore) AssignSpecificationIDs(
	_ context.Context,
	assignments map[uuid.UUID]uuid.UUID,
) error {
	s.assignmentCalls++
	s.assignments = cloneUUIDAssignments(assignments)
	for _, entity := range s.created {
		specificationID, ok := assignments[entity.ID]
		if !ok {
			continue
		}
		assignedID := specificationID
		entity.SpecificationID = &assignedID
	}
	return nil
}

type hierarchyCopySpecificationStore struct {
	domainFieldDevice.SpecificationStore
	source          []*domainFacility.Specification
	created         []*domainFacility.Specification
	bulkCreateCalls int
}

func (s *hierarchyCopySpecificationStore) GetByFieldDeviceIDs(
	_ context.Context,
	fieldDeviceIDs []uuid.UUID,
) ([]*domainFacility.Specification, error) {
	requested := uuidSetForCopyTest(fieldDeviceIDs)
	out := make([]*domainFacility.Specification, 0, len(s.source))
	for _, item := range s.source {
		if item == nil || item.FieldDeviceID == nil {
			continue
		}
		if _, ok := requested[*item.FieldDeviceID]; !ok {
			continue
		}
		clone := *item
		out = append(out, &clone)
	}
	return out, nil
}

func (s *hierarchyCopySpecificationStore) BulkCreate(
	_ context.Context,
	entities []*domainFacility.Specification,
	_ int,
) error {
	s.bulkCreateCalls++
	for _, entity := range entities {
		entity.ID = uuid.New()
		clone := *entity
		s.created = append(s.created, &clone)
	}
	return nil
}

type hierarchyCopyBacnetObjectStore struct {
	domainObjectData.BacnetObjectStore
	source            []*domainFacility.BacnetObject
	created           []*domainFacility.BacnetObject
	assignments       map[uuid.UUID]uuid.UUID
	bulkCreateCalls   int
	assignmentCalls   int
	individualUpdates int
}

func (s *hierarchyCopyBacnetObjectStore) GetByFieldDeviceIDs(
	_ context.Context,
	fieldDeviceIDs []uuid.UUID,
) ([]*domainFacility.BacnetObject, error) {
	requested := uuidSetForCopyTest(fieldDeviceIDs)
	out := make([]*domainFacility.BacnetObject, 0, len(s.source))
	for _, item := range s.source {
		if item == nil || item.FieldDeviceID == nil {
			continue
		}
		if _, ok := requested[*item.FieldDeviceID]; !ok {
			continue
		}
		clone := *item
		out = append(out, &clone)
	}
	return out, nil
}

func (s *hierarchyCopyBacnetObjectStore) BulkCreate(
	_ context.Context,
	entities []*domainFacility.BacnetObject,
	_ int,
) error {
	s.bulkCreateCalls++
	for _, entity := range entities {
		entity.ID = uuid.New()
		clone := *entity
		s.created = append(s.created, &clone)
	}
	return nil
}

func (s *hierarchyCopyBacnetObjectStore) AssignSoftwareReferenceIDs(
	_ context.Context,
	assignments map[uuid.UUID]uuid.UUID,
) error {
	s.assignmentCalls++
	s.assignments = cloneUUIDAssignments(assignments)
	for _, entity := range s.created {
		referenceID, ok := assignments[entity.ID]
		if !ok {
			continue
		}
		assignedID := referenceID
		entity.SoftwareReferenceID = &assignedID
	}
	return nil
}

func (s *hierarchyCopyBacnetObjectStore) Update(
	context.Context,
	*domainFacility.BacnetObject,
) error {
	s.individualUpdates++
	return nil
}

type hierarchyCopyAlarmValueStore struct {
	domainFacility.BacnetObjectAlarmValueRepository
	source          []domainFacility.BacnetObjectAlarmValue
	created         []*domainFacility.BacnetObjectAlarmValue
	readIDs         []uuid.UUID
	batchReadCalls  int
	bulkCreateCalls int
}

func (s *hierarchyCopyAlarmValueStore) GetByBacnetObjectIDs(
	_ context.Context,
	objectIDs []uuid.UUID,
) ([]domainFacility.BacnetObjectAlarmValue, error) {
	s.batchReadCalls++
	s.readIDs = append([]uuid.UUID(nil), objectIDs...)
	requested := uuidSetForCopyTest(objectIDs)
	out := make([]domainFacility.BacnetObjectAlarmValue, 0, len(s.source))
	for _, item := range s.source {
		if _, ok := requested[item.BacnetObjectID]; ok {
			out = append(out, item)
		}
	}
	return out, nil
}

func (s *hierarchyCopyAlarmValueStore) BulkCreate(
	_ context.Context,
	values []*domainFacility.BacnetObjectAlarmValue,
	_ int,
) error {
	s.bulkCreateCalls++
	for _, value := range values {
		value.ID = uuid.New()
		clone := *value
		s.created = append(s.created, &clone)
	}
	return nil
}

func TestHierarchyCopyPreservesSpecificationBacklinkAndBacnetAlarmValues(t *testing.T) {
	originalSystemTypeID := uuid.New()
	newSystemTypeID := uuid.New()
	originalFieldDeviceID := uuid.New()
	originalSpecificationID := uuid.New()
	originalInputID := uuid.New()
	originalOutputID := uuid.New()
	alarmTypeFieldID := uuid.New()
	unitID := uuid.New()
	supplier := "supplier"
	valueNumber := 42.5
	valueString := "high"

	fieldDevices := &hierarchyCopyFieldDeviceStore{}
	specifications := &hierarchyCopySpecificationStore{source: []*domainFacility.Specification{{
		Base:                  domain.Base{ID: originalSpecificationID},
		FieldDeviceID:         &originalFieldDeviceID,
		SpecificationSupplier: &supplier,
	}}}
	bacnetObjects := &hierarchyCopyBacnetObjectStore{source: []*domainFacility.BacnetObject{
		{
			Base:           domain.Base{ID: originalInputID},
			TextFix:        "AI-1",
			SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
			SoftwareNumber: 1,
			FieldDeviceID:  &originalFieldDeviceID,
		},
		{
			Base:                domain.Base{ID: originalOutputID},
			TextFix:             "AO-1",
			SoftwareType:        domainFacility.BacnetSoftwareTypeAO,
			SoftwareNumber:      2,
			FieldDeviceID:       &originalFieldDeviceID,
			SoftwareReferenceID: &originalInputID,
		},
	}}
	alarmValues := &hierarchyCopyAlarmValueStore{source: []domainFacility.BacnetObjectAlarmValue{
		{
			Base:             domain.Base{ID: uuid.New()},
			BacnetObjectID:   originalInputID,
			BacnetObject:     &domainFacility.BacnetObject{},
			AlarmTypeFieldID: alarmTypeFieldID,
			AlarmTypeField:   &domainFacility.AlarmTypeField{},
			ValueNumber:      &valueNumber,
			UnitID:           &unitID,
			Unit:             &domainFacility.Unit{},
			Source:           domainFacility.AlarmValueSourceImport,
		},
		{
			Base:             domain.Base{ID: uuid.New()},
			BacnetObjectID:   originalOutputID,
			AlarmTypeFieldID: alarmTypeFieldID,
			ValueString:      &valueString,
			Source:           domainFacility.AlarmValueSourceUser,
		},
	}}

	copier := projectFacilityCopy{
		fieldDeviceRepo:      fieldDevices,
		specificationRepo:    specifications,
		bacnetObjectRepo:     bacnetObjects,
		bacnetAlarmValueRepo: alarmValues,
	}
	copies, err := copier.copyFieldDevicesWithChildrenDetailed(
		context.Background(),
		[]domainFacility.FieldDevice{{
			Base:                      domain.Base{ID: originalFieldDeviceID},
			SPSControllerSystemTypeID: originalSystemTypeID,
			SpecificationID:           &originalSpecificationID,
		}},
		map[uuid.UUID]uuid.UUID{originalSystemTypeID: newSystemTypeID},
	)
	if err != nil {
		t.Fatalf("copy field device hierarchy: %v", err)
	}

	if len(copies) != 1 || copies[0].ID == uuid.Nil || copies[0].ID == originalFieldDeviceID {
		t.Fatalf("field device copies: got %+v", copies)
	}
	copyDevice := copies[0]
	if copyDevice.SPSControllerSystemTypeID != newSystemTypeID {
		t.Fatalf("copied system type: got %s, want %s", copyDevice.SPSControllerSystemTypeID, newSystemTypeID)
	}
	if copyDevice.SpecificationID == nil || len(specifications.created) != 1 {
		t.Fatalf("copied specification relationship: device=%+v specs=%+v", copyDevice.SpecificationID, specifications.created)
	}
	copySpecification := specifications.created[0]
	if copySpecification.ID == originalSpecificationID || copySpecification.FieldDeviceID == nil ||
		*copySpecification.FieldDeviceID != copyDevice.ID || *copyDevice.SpecificationID != copySpecification.ID {
		t.Fatalf("copied specification identity/backlink: device=%+v spec=%+v", copyDevice, copySpecification)
	}
	if fieldDevices.bulkCreateCalls != 1 || specifications.bulkCreateCalls != 1 ||
		fieldDevices.assignmentCalls != 1 {
		t.Fatalf(
			"field/specification batch calls: fields=%d specs=%d assignments=%d",
			fieldDevices.bulkCreateCalls,
			specifications.bulkCreateCalls,
			fieldDevices.assignmentCalls,
		)
	}

	if len(bacnetObjects.created) != 2 || bacnetObjects.bulkCreateCalls != 1 ||
		bacnetObjects.assignmentCalls != 1 || bacnetObjects.individualUpdates != 0 {
		t.Fatalf(
			"BACnet persistence shape: copies=%d bulk=%d assignment=%d individual=%d",
			len(bacnetObjects.created),
			bacnetObjects.bulkCreateCalls,
			bacnetObjects.assignmentCalls,
			bacnetObjects.individualUpdates,
		)
	}
	createdByTextFix := make(map[string]*domainFacility.BacnetObject, len(bacnetObjects.created))
	for _, object := range bacnetObjects.created {
		createdByTextFix[object.TextFix] = object
		if object.FieldDeviceID == nil || *object.FieldDeviceID != copyDevice.ID {
			t.Fatalf("BACnet owner: got %+v, want %s", object.FieldDeviceID, copyDevice.ID)
		}
	}
	inputCopy := createdByTextFix["AI-1"]
	outputCopy := createdByTextFix["AO-1"]
	if inputCopy == nil || outputCopy == nil || outputCopy.SoftwareReferenceID == nil ||
		*outputCopy.SoftwareReferenceID != inputCopy.ID {
		t.Fatalf("remapped software reference: input=%+v output=%+v", inputCopy, outputCopy)
	}

	if alarmValues.batchReadCalls != 1 || alarmValues.bulkCreateCalls != 1 ||
		len(alarmValues.created) != len(alarmValues.source) {
		t.Fatalf(
			"alarm-value batch shape: reads=%d creates=%d values=%d",
			alarmValues.batchReadCalls,
			alarmValues.bulkCreateCalls,
			len(alarmValues.created),
		)
	}
	if !reflect.DeepEqual(uuidSetForCopyTest(alarmValues.readIDs), uuidSetForCopyTest([]uuid.UUID{originalInputID, originalOutputID})) {
		t.Fatalf("alarm source IDs: got %v", alarmValues.readIDs)
	}
	createdAlarmByParent := make(map[uuid.UUID]*domainFacility.BacnetObjectAlarmValue, len(alarmValues.created))
	for _, value := range alarmValues.created {
		createdAlarmByParent[value.BacnetObjectID] = value
		if value.ID == uuid.Nil || value.BacnetObject != nil || value.AlarmTypeField != nil || value.Unit != nil {
			t.Fatalf("alarm clone persistence state: %+v", value)
		}
	}
	inputAlarm := createdAlarmByParent[inputCopy.ID]
	outputAlarm := createdAlarmByParent[outputCopy.ID]
	if inputAlarm == nil || inputAlarm.ValueNumber == nil || *inputAlarm.ValueNumber != valueNumber ||
		inputAlarm.AlarmTypeFieldID != alarmTypeFieldID || inputAlarm.UnitID == nil || *inputAlarm.UnitID != unitID ||
		inputAlarm.Source != domainFacility.AlarmValueSourceImport {
		t.Fatalf("input alarm clone: %+v", inputAlarm)
	}
	if outputAlarm == nil || outputAlarm.ValueString == nil || *outputAlarm.ValueString != valueString ||
		outputAlarm.Source != domainFacility.AlarmValueSourceUser {
		t.Fatalf("output alarm clone: %+v", outputAlarm)
	}
}

func TestHierarchyCopyProcessesFieldDevicesInBoundedKeysetPages(t *testing.T) {
	const fieldDeviceCount = copyFieldDevicePageSize*2 + 201
	originalSystemTypeID := uuid.New()
	newSystemTypeID := uuid.New()
	source := make(map[uuid.UUID]*domainFacility.FieldDevice, fieldDeviceCount)
	for i := 0; i < fieldDeviceCount; i++ {
		id := uuid.New()
		source[id] = &domainFacility.FieldDevice{
			Base:                      domain.Base{ID: id},
			SPSControllerSystemTypeID: originalSystemTypeID,
			ApparatNr:                 i + 1,
		}
	}

	fieldDevices := &hierarchyCopyFieldDeviceStore{source: source}
	copier := projectFacilityCopy{
		fieldDeviceRepo:   fieldDevices,
		specificationRepo: &hierarchyCopySpecificationStore{},
		bacnetObjectRepo:  &hierarchyCopyBacnetObjectStore{},
	}
	if err := copier.copyFieldDevicesForSystemTypes(
		context.Background(),
		map[uuid.UUID]uuid.UUID{originalSystemTypeID: newSystemTypeID},
	); err != nil {
		t.Fatalf("copy paged field devices: %v", err)
	}

	wantPageLengths := []int{copyFieldDevicePageSize, copyFieldDevicePageSize, 201}
	if fieldDevices.pageCalls != len(wantPageLengths) ||
		!reflect.DeepEqual(fieldDevices.pageLengths, wantPageLengths) ||
		!reflect.DeepEqual(fieldDevices.loadedLengths, wantPageLengths) {
		t.Fatalf(
			"page shape: calls=%d page_lengths=%v loaded_lengths=%v want=%v",
			fieldDevices.pageCalls,
			fieldDevices.pageLengths,
			fieldDevices.loadedLengths,
			wantPageLengths,
		)
	}
	if fieldDevices.bulkCreateCalls != len(wantPageLengths) || len(fieldDevices.created) != fieldDeviceCount {
		t.Fatalf(
			"copy batches: calls=%d copies=%d want=%d/%d",
			fieldDevices.bulkCreateCalls,
			len(fieldDevices.created),
			len(wantPageLengths),
			fieldDeviceCount,
		)
	}
	copiedIDs := make(map[uuid.UUID]struct{}, fieldDeviceCount)
	for _, item := range fieldDevices.created {
		if item.SPSControllerSystemTypeID != newSystemTypeID {
			t.Fatalf("copy %s parent: got %s, want %s", item.ID, item.SPSControllerSystemTypeID, newSystemTypeID)
		}
		if _, sourceID := source[item.ID]; sourceID {
			t.Fatalf("copy reused source ID %s", item.ID)
		}
		if _, duplicate := copiedIDs[item.ID]; duplicate {
			t.Fatalf("duplicate copied ID %s", item.ID)
		}
		copiedIDs[item.ID] = struct{}{}
	}
}

func cloneUUIDAssignments(source map[uuid.UUID]uuid.UUID) map[uuid.UUID]uuid.UUID {
	clone := make(map[uuid.UUID]uuid.UUID, len(source))
	for id, targetID := range source {
		clone[id] = targetID
	}
	return clone
}

func uuidSetForCopyTest(ids []uuid.UUID) map[uuid.UUID]struct{} {
	set := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}
