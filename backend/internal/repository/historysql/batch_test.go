package historysql

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/besart951/go_infra_link/backend/internal/service/auditctx"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type historyQueryCounter struct {
	logger.Interface
	count atomic.Int64
}

func (c *historyQueryCounter) Trace(
	context.Context,
	time.Time,
	func() (string, int64),
	error,
) {
	c.count.Add(1)
}

func (c *historyQueryCounter) reset() {
	c.count.Store(0)
}

func TestRecordMutationsUsesBoundedQueriesForFieldDeviceBatch(t *testing.T) {
	singleQueries := recordFieldDeviceHistoryBatch(t, 1)
	manyQueries := recordFieldDeviceHistoryBatch(t, 20)

	if manyQueries != singleQueries {
		t.Fatalf(
			"query count must not grow per FieldDevice: one=%d, twenty=%d",
			singleQueries,
			manyQueries,
		)
	}
	if manyQueries > 7 {
		t.Fatalf("expected at most 7 statements for one history batch, got %d", manyQueries)
	}
}

func TestRecordMutationsChunksLargeFieldDeviceBatch(t *testing.T) {
	const itemCount = 501
	const expectedQueries = 17

	if got := recordFieldDeviceHistoryBatch(t, itemCount); got != expectedQueries {
		t.Fatalf(
			"queries for %d FieldDevices: got %d, want %d",
			itemCount,
			got,
			expectedQueries,
		)
	}
}

func TestRecordMutationsUsesBoundedQueriesForSpecificationBatch(t *testing.T) {
	singleQueries := recordSpecificationHistoryBatch(t, 1)
	manyQueries := recordSpecificationHistoryBatch(t, 20)

	if manyQueries != singleQueries {
		t.Fatalf(
			"query count must not grow per Specification: one=%d, twenty=%d",
			singleQueries,
			manyQueries,
		)
	}
	if manyQueries > 7 {
		t.Fatalf("expected at most 7 statements for one history batch, got %d", manyQueries)
	}
}

func TestRecordMutationsUsesBoundedQueriesForBacnetChildBatches(t *testing.T) {
	for _, table := range []string{
		"bacnet_objects",
		"bacnet_object_alarm_values",
	} {
		t.Run(table, func(t *testing.T) {
			singleQueries := recordBacnetChildHistoryBatch(t, table, 1)
			manyQueries := recordBacnetChildHistoryBatch(t, table, 20)

			if manyQueries != singleQueries {
				t.Fatalf(
					"query count must not grow per %s row: one=%d, twenty=%d",
					table,
					singleQueries,
					manyQueries,
				)
			}
			if manyQueries > 9 {
				t.Fatalf(
					"expected at most 9 statements for one %s history batch, got %d",
					table,
					manyQueries,
				)
			}
		})
	}
}

func TestRecordMutationsUsesBoundedQueriesForBacnetObjectMoveBatch(t *testing.T) {
	singleQueries := recordBacnetObjectMoveHistoryBatch(t, 1)
	manyQueries := recordBacnetObjectMoveHistoryBatch(t, 20)

	if manyQueries != singleQueries {
		t.Fatalf(
			"query count must not grow per BACnet move: one=%d, twenty=%d",
			singleQueries,
			manyQueries,
		)
	}
	if manyQueries > 9 {
		t.Fatalf("expected at most 9 statements for one BACnet move batch, got %d", manyQueries)
	}
}

func TestRecordMutationsUsesBoundedQueriesForSystemTypeBatch(t *testing.T) {
	singleQueries := recordSystemTypeHistoryBatch(t, 1)
	manyQueries := recordSystemTypeHistoryBatch(t, 20)

	if manyQueries != singleQueries {
		t.Fatalf(
			"query count must not grow per SPS system type: one=%d, twenty=%d",
			singleQueries,
			manyQueries,
		)
	}
	if manyQueries > 5 {
		t.Fatalf("expected at most 5 statements for one history batch, got %d", manyQueries)
	}
}

func TestRecordMutationsUsesBoundedQueriesForSPSControllerMoveBatch(t *testing.T) {
	singleQueries := recordSPSControllerMoveHistoryBatch(t, 1)
	manyQueries := recordSPSControllerMoveHistoryBatch(t, 20)

	if manyQueries != singleQueries {
		t.Fatalf(
			"query count must not grow per SPS controller move: one=%d, twenty=%d",
			singleQueries,
			manyQueries,
		)
	}
	if manyQueries > 9 {
		t.Fatalf("expected at most 9 statements for one SPS move batch, got %d", manyQueries)
	}
}

func TestRecordMutationsUsesBoundedQueriesForControlCabinetMoveBatch(t *testing.T) {
	singleQueries := recordControlCabinetMoveHistoryBatch(t, 1)
	manyQueries := recordControlCabinetMoveHistoryBatch(t, 20)

	if manyQueries != singleQueries {
		t.Fatalf(
			"query count must not grow per control cabinet move: one=%d, twenty=%d",
			singleQueries,
			manyQueries,
		)
	}
	if manyQueries > 5 {
		t.Fatalf("expected at most 5 statements for one cabinet move batch, got %d", manyQueries)
	}
}

func TestRecordMutationsIncludesBacnetObjectDataScopes(t *testing.T) {
	counter := &historyQueryCounter{Interface: logger.Discard}
	db := newHistoryBatchTestDB(t, counter, 1)
	bacnetObjectID := historyBatchUUID(3001)
	objectDataID := historyBatchUUID(4001)
	objectDataProjectID := historyBatchUUID(4002)
	if err := db.Exec(
		`INSERT INTO object_data (id, project_id) VALUES (?, ?)`,
		objectDataID,
		objectDataProjectID,
	).Error; err != nil {
		t.Fatalf("insert ObjectData: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO object_data_bacnet_objects (bacnet_object_id, object_data_id) VALUES (?, ?)`,
		bacnetObjectID,
		objectDataID,
	).Error; err != nil {
		t.Fatalf("insert ObjectData BACnet link: %v", err)
	}

	fieldDeviceID := historyBatchUUID(1)
	if err := NewStore(db).RecordMutations(
		context.Background(),
		[]Mutation{{
			Action:      domainHistory.ActionCreate,
			EntityTable: "bacnet_objects",
			EntityID:    bacnetObjectID,
			AfterJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","field_device_id":"%s"}`,
				bacnetObjectID,
				fieldDeviceID,
			)),
		}},
	); err != nil {
		t.Fatalf("record BACnet mutation: %v", err)
	}

	var scopes []domainHistory.ChangeEventScope
	if err := db.Find(&scopes).Error; err != nil {
		t.Fatalf("load scopes: %v", err)
	}
	got := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		got[scope.ScopeType+":"+scope.ScopeID.String()] = struct{}{}
	}
	for _, want := range []string{
		scopeObjectData + ":" + objectDataID.String(),
		scopeProject + ":" + objectDataProjectID.String(),
	} {
		if _, ok := got[want]; !ok {
			t.Fatalf("missing scope %s in %v", want, got)
		}
	}
}

func TestRecordMutationsBacnetMoveIncludesOldAndNewFieldDeviceScopes(t *testing.T) {
	counter := &historyQueryCounter{Interface: logger.Discard}
	db := newHistoryBatchTestDB(t, counter, 1)
	bacnetObjectID := historyBatchUUID(3001)
	oldFieldDeviceID := historyBatchUUID(1)
	newFieldDeviceID := historyBatchUUID(9001)
	newSystemTypeID := historyBatchUUID(110)
	newSPSControllerID := historyBatchUUID(111)
	newControlCabinetID := historyBatchUUID(112)
	newBuildingID := historyBatchUUID(113)
	newFieldDeviceProjectID := historyBatchUUID(210)
	newSPSProjectID := historyBatchUUID(211)
	newCabinetProjectID := historyBatchUUID(212)

	insertBacnetMoveDestination(
		t,
		db,
		newFieldDeviceID,
		newSystemTypeID,
		newSPSControllerID,
		newControlCabinetID,
		newBuildingID,
		newFieldDeviceProjectID,
		newSPSProjectID,
		newCabinetProjectID,
	)

	if err := NewStore(db).RecordMutations(
		context.Background(),
		[]Mutation{{
			Action:      domainHistory.ActionUpdate,
			EntityTable: "bacnet_objects",
			EntityID:    bacnetObjectID,
			BeforeJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","field_device_id":"%s"}`,
				bacnetObjectID,
				oldFieldDeviceID,
			)),
			AfterJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","field_device_id":"%s"}`,
				bacnetObjectID,
				newFieldDeviceID,
			)),
		}},
	); err != nil {
		t.Fatalf("record BACnet move: %v", err)
	}

	var scopes []domainHistory.ChangeEventScope
	if err := db.Find(&scopes).Error; err != nil {
		t.Fatalf("load BACnet move scopes: %v", err)
	}
	got := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		got[scope.ScopeType+":"+scope.ScopeID.String()] = struct{}{}
	}
	for _, want := range []string{
		scopeBacnetObject + ":" + bacnetObjectID.String(),
		scopeFieldDevice + ":" + oldFieldDeviceID.String(),
		scopeFieldDevice + ":" + newFieldDeviceID.String(),
		scopeSPSControllerSystemType + ":" + historyBatchUUID(100).String(),
		scopeSPSControllerSystemType + ":" + newSystemTypeID.String(),
		scopeSPSController + ":" + historyBatchUUID(101).String(),
		scopeSPSController + ":" + newSPSControllerID.String(),
		scopeControlCabinet + ":" + historyBatchUUID(102).String(),
		scopeControlCabinet + ":" + newControlCabinetID.String(),
		scopeBuilding + ":" + historyBatchUUID(103).String(),
		scopeBuilding + ":" + newBuildingID.String(),
		scopeProject + ":" + historyBatchUUID(200).String(),
		scopeProject + ":" + historyBatchUUID(201).String(),
		scopeProject + ":" + historyBatchUUID(202).String(),
		scopeProject + ":" + newFieldDeviceProjectID.String(),
		scopeProject + ":" + newSPSProjectID.String(),
		scopeProject + ":" + newCabinetProjectID.String(),
	} {
		if _, ok := got[want]; !ok {
			t.Fatalf("missing BACnet move scope %s in %v", want, got)
		}
	}
	if len(got) != 17 {
		t.Fatalf("BACnet move scopes: got %d, want 17: %v", len(got), got)
	}
}

func TestRecordMutationsFieldDeviceMoveIncludesOldAndNewHierarchyScopes(t *testing.T) {
	counter := &historyQueryCounter{Interface: logger.Discard}
	db := newHistoryBatchTestDB(t, counter, 1)
	newSystemTypeID := historyBatchUUID(110)
	newSPSControllerID := historyBatchUUID(111)
	newControlCabinetID := historyBatchUUID(112)
	newBuildingID := historyBatchUUID(113)
	newSPSProjectID := historyBatchUUID(211)
	newCabinetProjectID := historyBatchUUID(212)
	for statement, args := range map[string][]any{
		`INSERT INTO sps_controller_system_types (id, sps_controller_id) VALUES (?, ?)`: {
			newSystemTypeID,
			newSPSControllerID,
		},
		`INSERT INTO sps_controllers (id, control_cabinet_id) VALUES (?, ?)`: {
			newSPSControllerID,
			newControlCabinetID,
		},
		`INSERT INTO control_cabinets (id, building_id) VALUES (?, ?)`: {
			newControlCabinetID,
			newBuildingID,
		},
		`INSERT INTO project_sps_controllers (sps_controller_id, project_id) VALUES (?, ?)`: {
			newSPSControllerID,
			newSPSProjectID,
		},
		`INSERT INTO project_control_cabinets (control_cabinet_id, project_id) VALUES (?, ?)`: {
			newControlCabinetID,
			newCabinetProjectID,
		},
	} {
		if err := db.Exec(statement, args...).Error; err != nil {
			t.Fatalf("insert move scope fixture: %v", err)
		}
	}

	fieldDeviceID := historyBatchUUID(1)
	oldSystemTypeID := historyBatchUUID(100)
	if err := NewStore(db).RecordMutations(
		context.Background(),
		[]Mutation{{
			Action:      domainHistory.ActionUpdate,
			EntityTable: "field_devices",
			EntityID:    fieldDeviceID,
			BeforeJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","sps_controller_system_type_id":"%s"}`,
				fieldDeviceID,
				oldSystemTypeID,
			)),
			AfterJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","sps_controller_system_type_id":"%s"}`,
				fieldDeviceID,
				newSystemTypeID,
			)),
		}},
	); err != nil {
		t.Fatalf("record FieldDevice move: %v", err)
	}

	var scopes []domainHistory.ChangeEventScope
	if err := db.Find(&scopes).Error; err != nil {
		t.Fatalf("load move scopes: %v", err)
	}
	got := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		got[scope.ScopeType+":"+scope.ScopeID.String()] = struct{}{}
	}
	for _, want := range []string{
		scopeFieldDevice + ":" + fieldDeviceID.String(),
		scopeSPSControllerSystemType + ":" + oldSystemTypeID.String(),
		scopeSPSControllerSystemType + ":" + newSystemTypeID.String(),
		scopeSPSController + ":" + historyBatchUUID(101).String(),
		scopeSPSController + ":" + newSPSControllerID.String(),
		scopeControlCabinet + ":" + historyBatchUUID(102).String(),
		scopeControlCabinet + ":" + newControlCabinetID.String(),
		scopeBuilding + ":" + historyBatchUUID(103).String(),
		scopeBuilding + ":" + newBuildingID.String(),
		scopeProject + ":" + historyBatchUUID(200).String(),
		scopeProject + ":" + historyBatchUUID(201).String(),
		scopeProject + ":" + historyBatchUUID(202).String(),
		scopeProject + ":" + newSPSProjectID.String(),
		scopeProject + ":" + newCabinetProjectID.String(),
	} {
		if _, ok := got[want]; !ok {
			t.Fatalf("missing move scope %s in %v", want, got)
		}
	}
	if len(got) != 14 {
		t.Fatalf("move scopes: got %d, want 14: %v", len(got), got)
	}
}

func TestRecordMutationsSPSControllerMoveIncludesOldAndNewCabinetScopes(t *testing.T) {
	counter := &historyQueryCounter{Interface: logger.Discard}
	db := newHistoryBatchTestDB(t, counter, 1)
	spsControllerID := historyBatchUUID(101)
	oldCabinetID := historyBatchUUID(102)
	oldBuildingID := historyBatchUUID(103)
	newCabinetID := historyBatchUUID(112)
	newBuildingID := historyBatchUUID(113)
	newCabinetProjectID := historyBatchUUID(212)
	unrelatedProjectID := historyBatchUUID(299)

	if err := db.Exec(
		`INSERT INTO control_cabinets (id, building_id) VALUES (?, ?)`,
		newCabinetID,
		newBuildingID,
	).Error; err != nil {
		t.Fatalf("insert destination cabinet: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO project_control_cabinets (control_cabinet_id, project_id) VALUES (?, ?)`,
		newCabinetID,
		newCabinetProjectID,
	).Error; err != nil {
		t.Fatalf("insert destination cabinet link: %v", err)
	}
	if err := db.Exec(
		`UPDATE sps_controllers SET control_cabinet_id = ? WHERE id = ?`,
		newCabinetID,
		spsControllerID,
	).Error; err != nil {
		t.Fatalf("move SPS fixture: %v", err)
	}
	unrelatedSPSID := historyBatchUUID(190)
	if err := db.Exec(
		`INSERT INTO sps_controllers (id, control_cabinet_id) VALUES (?, ?)`,
		unrelatedSPSID,
		oldCabinetID,
	).Error; err != nil {
		t.Fatalf("insert unrelated SPS fixture: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO project_sps_controllers (sps_controller_id, project_id) VALUES (?, ?)`,
		unrelatedSPSID,
		unrelatedProjectID,
	).Error; err != nil {
		t.Fatalf("insert unrelated SPS project link: %v", err)
	}

	if err := NewStore(db).RecordMutations(
		context.Background(),
		[]Mutation{{
			Action:      domainHistory.ActionUpdate,
			EntityTable: "sps_controllers",
			EntityID:    spsControllerID,
			BeforeJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","control_cabinet_id":"%s"}`,
				spsControllerID,
				oldCabinetID,
			)),
			AfterJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","control_cabinet_id":"%s"}`,
				spsControllerID,
				newCabinetID,
			)),
		}},
	); err != nil {
		t.Fatalf("record SPS move: %v", err)
	}

	var scopes []domainHistory.ChangeEventScope
	if err := db.Find(&scopes).Error; err != nil {
		t.Fatalf("load SPS move scopes: %v", err)
	}
	got := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		got[scope.ScopeType+":"+scope.ScopeID.String()] = struct{}{}
	}
	for _, want := range []string{
		scopeSPSController + ":" + spsControllerID.String(),
		scopeControlCabinet + ":" + oldCabinetID.String(),
		scopeControlCabinet + ":" + newCabinetID.String(),
		scopeBuilding + ":" + oldBuildingID.String(),
		scopeBuilding + ":" + newBuildingID.String(),
		scopeProject + ":" + historyBatchUUID(200).String(),
		scopeProject + ":" + historyBatchUUID(201).String(),
		scopeProject + ":" + historyBatchUUID(202).String(),
		scopeProject + ":" + newCabinetProjectID.String(),
	} {
		if _, ok := got[want]; !ok {
			t.Fatalf("missing SPS move scope %s in %v", want, got)
		}
	}
	if _, leaked := got[scopeProject+":"+unrelatedProjectID.String()]; leaked {
		t.Fatalf("SPS move leaked unrelated old-cabinet project %s: %v", unrelatedProjectID, got)
	}
	if len(got) != 9 {
		t.Fatalf("SPS move scopes: got %d, want 9: %v", len(got), got)
	}
}

func TestRecordMutationsControlCabinetMoveIncludesOldAndNewBuildingScopes(t *testing.T) {
	counter := &historyQueryCounter{Interface: logger.Discard}
	db := newHistoryBatchTestDB(t, counter, 1)
	controlCabinetID := historyBatchUUID(102)
	oldBuildingID := historyBatchUUID(103)
	newBuildingID := historyBatchUUID(113)
	if err := db.Exec(
		`UPDATE control_cabinets SET building_id = ? WHERE id = ?`,
		newBuildingID,
		controlCabinetID,
	).Error; err != nil {
		t.Fatalf("move cabinet fixture: %v", err)
	}

	if err := NewStore(db).RecordMutations(
		context.Background(),
		[]Mutation{{
			Action:      domainHistory.ActionUpdate,
			EntityTable: "control_cabinets",
			EntityID:    controlCabinetID,
			BeforeJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","building_id":"%s"}`,
				controlCabinetID,
				oldBuildingID,
			)),
			AfterJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","building_id":"%s"}`,
				controlCabinetID,
				newBuildingID,
			)),
		}},
	); err != nil {
		t.Fatalf("record cabinet move: %v", err)
	}

	var scopes []domainHistory.ChangeEventScope
	if err := db.Find(&scopes).Error; err != nil {
		t.Fatalf("load cabinet move scopes: %v", err)
	}
	got := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		got[scope.ScopeType+":"+scope.ScopeID.String()] = struct{}{}
	}
	for _, want := range []string{
		scopeControlCabinet + ":" + controlCabinetID.String(),
		scopeBuilding + ":" + oldBuildingID.String(),
		scopeBuilding + ":" + newBuildingID.String(),
		scopeProject + ":" + historyBatchUUID(200).String(),
		scopeProject + ":" + historyBatchUUID(201).String(),
		scopeProject + ":" + historyBatchUUID(202).String(),
	} {
		if _, ok := got[want]; !ok {
			t.Fatalf("missing cabinet move scope %s in %v", want, got)
		}
	}
	if len(got) != 6 {
		t.Fatalf("cabinet move scopes: got %d, want 6: %v", len(got), got)
	}
}

func TestRecordDeletesPreservesSnapshotsAndWritesDeleteVersions(t *testing.T) {
	counter := &historyQueryCounter{Interface: logger.Discard}
	db := newHistoryBatchTestDB(t, counter, 2)
	store := NewStore(db)
	systemTypeID := historyBatchUUID(100)
	rows := map[uuid.UUID]domainHistory.JSONB{}
	for i := 1; i <= 2; i++ {
		fieldDeviceID := historyBatchUUID(i)
		rows[fieldDeviceID] = domainHistory.JSONB(fmt.Sprintf(
			`{"id":"%s","sps_controller_system_type_id":"%s","apparat_nr":%d}`,
			fieldDeviceID,
			systemTypeID,
			i,
		))
	}

	if err := store.RecordDeletes(
		context.Background(),
		"field_devices",
		rows,
	); err != nil {
		t.Fatalf("record deletes: %v", err)
	}

	var events []domainHistory.ChangeEvent
	if err := db.Order("entity_id").Find(&events).Error; err != nil {
		t.Fatalf("load delete events: %v", err)
	}
	if len(events) != len(rows) {
		t.Fatalf("events: got %d, want %d", len(events), len(rows))
	}
	for _, event := range events {
		if event.Action != domainHistory.ActionDelete {
			t.Fatalf("event %s action: got %q, want delete", event.ID, event.Action)
		}
		if got, want := string(event.BeforeJSON), string(rows[event.EntityID]); got != want {
			t.Fatalf("event %s before: got %s, want %s", event.ID, got, want)
		}
		if len(event.AfterJSON) != 0 {
			t.Fatalf("event %s after: got %s, want empty", event.ID, event.AfterJSON)
		}
	}

	var versions []domainHistory.EntityVersion
	if err := db.Find(&versions).Error; err != nil {
		t.Fatalf("load delete versions: %v", err)
	}
	if len(versions) != len(rows) {
		t.Fatalf("versions: got %d, want %d", len(versions), len(rows))
	}
	for _, version := range versions {
		if version.Action != domainHistory.ActionDelete {
			t.Fatalf(
				"version %s action: got %q, want delete",
				version.ID,
				version.Action,
			)
		}
		if len(version.SnapshotJSON) != 0 {
			t.Fatalf(
				"version %s snapshot: got %s, want empty",
				version.ID,
				version.SnapshotJSON,
			)
		}
	}
}

func TestRecordMutationsScopeFailureDoesNotInsertPartialHistory(t *testing.T) {
	counter := &historyQueryCounter{Interface: logger.Discard}
	db := newHistoryBatchTestDB(t, counter, 1)
	if err := db.Exec("DROP TABLE project_field_devices").Error; err != nil {
		t.Fatalf("drop project link table: %v", err)
	}

	fieldDeviceID := historyBatchUUID(1)
	systemTypeID := historyBatchUUID(100)
	err := NewStore(db).RecordMutations(context.Background(), []Mutation{{
		Action:      domainHistory.ActionCreate,
		EntityTable: "field_devices",
		EntityID:    fieldDeviceID,
		AfterJSON: domainHistory.JSONB(fmt.Sprintf(
			`{"id":"%s","sps_controller_system_type_id":"%s"}`,
			fieldDeviceID,
			systemTypeID,
		)),
	}})
	if err == nil {
		t.Fatal("expected scope resolution failure")
	}

	for name, model := range map[string]any{
		"events":   &domainHistory.ChangeEvent{},
		"scopes":   &domainHistory.ChangeEventScope{},
		"versions": &domainHistory.EntityVersion{},
	} {
		var count int64
		if countErr := db.Model(model).Count(&count).Error; countErr != nil {
			t.Fatalf("count %s: %v", name, countErr)
		}
		if count != 0 {
			t.Fatalf("%s: got %d rows after preparation failure, want 0", name, count)
		}
	}
}

func recordFieldDeviceHistoryBatch(t *testing.T, count int) int64 {
	t.Helper()

	counter := &historyQueryCounter{Interface: logger.Discard}
	db := newHistoryBatchTestDB(t, counter, count)
	store := NewStore(db)
	batchID := uuid.New()
	actorID := uuid.New()
	systemTypeID := historyBatchUUID(100)
	mutations := make([]Mutation, 0, count)
	for i := 1; i <= count; i++ {
		fieldDeviceID := historyBatchUUID(i)
		mutations = append(mutations, Mutation{
			Action:      domainHistory.ActionCreate,
			EntityTable: "field_devices",
			EntityID:    fieldDeviceID,
			AfterJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","sps_controller_system_type_id":"%s"}`,
				fieldDeviceID,
				systemTypeID,
			)),
		})
	}

	counter.reset()
	ctx := auditctx.WithBatchID(
		auditctx.WithActorID(context.Background(), actorID),
		batchID,
	)
	if err := store.RecordMutations(ctx, mutations); err != nil {
		t.Fatalf("record %d FieldDevice mutations: %v", count, err)
	}
	queryCount := counter.count.Load()

	var eventCount int64
	if err := db.Model(&domainHistory.ChangeEvent{}).Count(&eventCount).Error; err != nil {
		t.Fatalf("count events: %v", err)
	}
	if eventCount != int64(count) {
		t.Fatalf("events: got %d, want %d", eventCount, count)
	}
	var versionCount int64
	if err := db.Model(&domainHistory.EntityVersion{}).Count(&versionCount).Error; err != nil {
		t.Fatalf("count versions: %v", err)
	}
	if versionCount != int64(count) {
		t.Fatalf("versions: got %d, want %d", versionCount, count)
	}
	var scopeCount int64
	if err := db.Model(&domainHistory.ChangeEventScope{}).Count(&scopeCount).Error; err != nil {
		t.Fatalf("count scopes: %v", err)
	}
	const scopesPerFieldDevice = 8
	if scopeCount != int64(count*scopesPerFieldDevice) {
		t.Fatalf("scopes: got %d, want %d", scopeCount, count*scopesPerFieldDevice)
	}

	var events []domainHistory.ChangeEvent
	if err := db.Find(&events).Error; err != nil {
		t.Fatalf("load events: %v", err)
	}
	for _, event := range events {
		if event.BatchID == nil || *event.BatchID != batchID {
			t.Fatalf("event %s batch: got %v, want %s", event.ID, event.BatchID, batchID)
		}
		if event.ActorID == nil || *event.ActorID != actorID {
			t.Fatalf("event %s actor: got %v, want %s", event.ID, event.ActorID, actorID)
		}
	}
	return queryCount
}

func recordSpecificationHistoryBatch(t *testing.T, count int) int64 {
	t.Helper()

	counter := &historyQueryCounter{Interface: logger.Discard}
	db := newHistoryBatchTestDB(t, counter, count)
	fieldDeviceID := historyBatchUUID(1)
	mutations := make([]Mutation, 0, count)
	for i := 1; i <= count; i++ {
		specificationID := historyBatchUUID(1000 + i)
		mutations = append(mutations, Mutation{
			Action:      domainHistory.ActionCreate,
			EntityTable: "specifications",
			EntityID:    specificationID,
			AfterJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","field_device_id":"%s"}`,
				specificationID,
				fieldDeviceID,
			)),
		})
	}

	counter.reset()
	if err := NewStore(db).RecordMutations(
		context.Background(),
		mutations,
	); err != nil {
		t.Fatalf("record %d Specification mutations: %v", count, err)
	}
	queryCount := counter.count.Load()

	var scopeCount int64
	if err := db.Model(&domainHistory.ChangeEventScope{}).Count(&scopeCount).Error; err != nil {
		t.Fatalf("count scopes: %v", err)
	}
	const scopesPerSpecification = 9
	if scopeCount != int64(count*scopesPerSpecification) {
		t.Fatalf("scopes: got %d, want %d", scopeCount, count*scopesPerSpecification)
	}
	return queryCount
}

func recordSystemTypeHistoryBatch(t *testing.T, count int) int64 {
	t.Helper()

	counter := &historyQueryCounter{Interface: logger.Discard}
	db := newHistoryBatchTestDB(t, counter, count)
	spsControllerID := historyBatchUUID(101)
	mutations := make([]Mutation, 0, count)
	for i := 1; i <= count; i++ {
		systemTypeID := historyBatchUUID(5000 + i)
		mutations = append(mutations, Mutation{
			Action:      domainHistory.ActionCreate,
			EntityTable: "sps_controller_system_types",
			EntityID:    systemTypeID,
			AfterJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","sps_controller_id":"%s"}`,
				systemTypeID,
				spsControllerID,
			)),
		})
	}

	counter.reset()
	if err := NewStore(db).RecordMutations(
		context.Background(),
		mutations,
	); err != nil {
		t.Fatalf("record %d SPS system type mutations: %v", count, err)
	}
	queryCount := counter.count.Load()

	var scopeCount int64
	if err := db.Model(&domainHistory.ChangeEventScope{}).Count(&scopeCount).Error; err != nil {
		t.Fatalf("count scopes: %v", err)
	}
	const scopesPerSystemType = 7
	if scopeCount != int64(count*scopesPerSystemType) {
		t.Fatalf("scopes: got %d, want %d", scopeCount, count*scopesPerSystemType)
	}
	return queryCount
}

func recordSPSControllerMoveHistoryBatch(t *testing.T, count int) int64 {
	t.Helper()

	counter := &historyQueryCounter{Interface: logger.Discard}
	db := newHistoryBatchTestDB(t, counter, count)
	oldCabinetID := historyBatchUUID(102)
	newCabinetID := historyBatchUUID(712)
	newBuildingID := historyBatchUUID(713)
	newCabinetProjectID := historyBatchUUID(714)
	if err := db.Exec(
		`INSERT INTO control_cabinets (id, building_id) VALUES (?, ?)`,
		newCabinetID,
		newBuildingID,
	).Error; err != nil {
		t.Fatalf("insert destination cabinet: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO project_control_cabinets (control_cabinet_id, project_id) VALUES (?, ?)`,
		newCabinetID,
		newCabinetProjectID,
	).Error; err != nil {
		t.Fatalf("insert destination cabinet link: %v", err)
	}

	mutations := make([]Mutation, 0, count)
	for i := 1; i <= count; i++ {
		spsControllerID := historyBatchUUID(7000 + i)
		if err := db.Exec(
			`INSERT INTO sps_controllers (id, control_cabinet_id) VALUES (?, ?)`,
			spsControllerID,
			newCabinetID,
		).Error; err != nil {
			t.Fatalf("insert SPS controller %d: %v", i, err)
		}
		if err := db.Exec(
			`INSERT INTO project_sps_controllers (sps_controller_id, project_id) VALUES (?, ?)`,
			spsControllerID,
			historyBatchUUID(201),
		).Error; err != nil {
			t.Fatalf("insert SPS project link %d: %v", i, err)
		}
		mutations = append(mutations, Mutation{
			Action:      domainHistory.ActionUpdate,
			EntityTable: "sps_controllers",
			EntityID:    spsControllerID,
			BeforeJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","control_cabinet_id":"%s"}`,
				spsControllerID,
				oldCabinetID,
			)),
			AfterJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","control_cabinet_id":"%s"}`,
				spsControllerID,
				newCabinetID,
			)),
		})
	}

	counter.reset()
	if err := NewStore(db).RecordMutations(context.Background(), mutations); err != nil {
		t.Fatalf("record %d SPS controller moves: %v", count, err)
	}
	queryCount := counter.count.Load()

	var scopeCount int64
	if err := db.Model(&domainHistory.ChangeEventScope{}).Count(&scopeCount).Error; err != nil {
		t.Fatalf("count scopes: %v", err)
	}
	const scopesPerMove = 8
	if scopeCount != int64(count*scopesPerMove) {
		t.Fatalf("scopes: got %d, want %d", scopeCount, count*scopesPerMove)
	}
	return queryCount
}

func recordControlCabinetMoveHistoryBatch(t *testing.T, count int) int64 {
	t.Helper()

	counter := &historyQueryCounter{Interface: logger.Discard}
	db := newHistoryBatchTestDB(t, counter, count)
	oldBuildingID := historyBatchUUID(803)
	newBuildingID := historyBatchUUID(813)
	projectID := historyBatchUUID(814)
	mutations := make([]Mutation, 0, count)
	for i := 1; i <= count; i++ {
		controlCabinetID := historyBatchUUID(8000 + i)
		if err := db.Exec(
			`INSERT INTO control_cabinets (id, building_id) VALUES (?, ?)`,
			controlCabinetID,
			newBuildingID,
		).Error; err != nil {
			t.Fatalf("insert control cabinet %d: %v", i, err)
		}
		if err := db.Exec(
			`INSERT INTO project_control_cabinets (control_cabinet_id, project_id) VALUES (?, ?)`,
			controlCabinetID,
			projectID,
		).Error; err != nil {
			t.Fatalf("insert cabinet project link %d: %v", i, err)
		}
		mutations = append(mutations, Mutation{
			Action:      domainHistory.ActionUpdate,
			EntityTable: "control_cabinets",
			EntityID:    controlCabinetID,
			BeforeJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","building_id":"%s"}`,
				controlCabinetID,
				oldBuildingID,
			)),
			AfterJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","building_id":"%s"}`,
				controlCabinetID,
				newBuildingID,
			)),
		})
	}

	counter.reset()
	if err := NewStore(db).RecordMutations(context.Background(), mutations); err != nil {
		t.Fatalf("record %d control cabinet moves: %v", count, err)
	}
	queryCount := counter.count.Load()

	var scopeCount int64
	if err := db.Model(&domainHistory.ChangeEventScope{}).Count(&scopeCount).Error; err != nil {
		t.Fatalf("count scopes: %v", err)
	}
	const scopesPerMove = 4
	if scopeCount != int64(count*scopesPerMove) {
		t.Fatalf("scopes: got %d, want %d", scopeCount, count*scopesPerMove)
	}
	return queryCount
}

func recordBacnetChildHistoryBatch(
	t *testing.T,
	table string,
	count int,
) int64 {
	t.Helper()

	counter := &historyQueryCounter{Interface: logger.Discard}
	db := newHistoryBatchTestDB(t, counter, count)
	fieldDeviceID := historyBatchUUID(1)
	bacnetObjectID := historyBatchUUID(3000)
	if table == "bacnet_object_alarm_values" {
		if err := db.Exec(
			`INSERT INTO bacnet_objects (id, field_device_id) VALUES (?, ?)`,
			bacnetObjectID,
			fieldDeviceID,
		).Error; err != nil {
			t.Fatalf("insert BACnet object: %v", err)
		}
	}

	mutations := make([]Mutation, 0, count)
	for i := 1; i <= count; i++ {
		entityID := historyBatchUUID(3000 + i)
		var snapshot string
		switch table {
		case "bacnet_objects":
			snapshot = fmt.Sprintf(
				`{"id":"%s","field_device_id":"%s"}`,
				entityID,
				fieldDeviceID,
			)
		case "bacnet_object_alarm_values":
			snapshot = fmt.Sprintf(
				`{"id":"%s","bacnet_object_id":"%s"}`,
				entityID,
				bacnetObjectID,
			)
		default:
			t.Fatalf("unsupported BACnet child table %q", table)
		}
		mutations = append(mutations, Mutation{
			Action:      domainHistory.ActionCreate,
			EntityTable: table,
			EntityID:    entityID,
			AfterJSON:   domainHistory.JSONB(snapshot),
		})
	}

	counter.reset()
	if err := NewStore(db).RecordMutations(
		context.Background(),
		mutations,
	); err != nil {
		t.Fatalf("record %d %s mutations: %v", count, table, err)
	}
	queryCount := counter.count.Load()

	var scopeCount int64
	if err := db.Model(&domainHistory.ChangeEventScope{}).Count(&scopeCount).Error; err != nil {
		t.Fatalf("count scopes: %v", err)
	}
	const scopesPerBacnetChild = 9
	if scopeCount != int64(count*scopesPerBacnetChild) {
		t.Fatalf(
			"scopes: got %d, want %d",
			scopeCount,
			count*scopesPerBacnetChild,
		)
	}
	return queryCount
}

func recordBacnetObjectMoveHistoryBatch(t *testing.T, count int) int64 {
	t.Helper()

	counter := &historyQueryCounter{Interface: logger.Discard}
	db := newHistoryBatchTestDB(t, counter, count)
	newSystemTypeID := historyBatchUUID(9100)
	newSPSControllerID := historyBatchUUID(9101)
	newControlCabinetID := historyBatchUUID(9102)
	newBuildingID := historyBatchUUID(9103)
	newFieldDeviceProjectID := historyBatchUUID(9200)
	newSPSProjectID := historyBatchUUID(9201)
	newCabinetProjectID := historyBatchUUID(9202)

	mutations := make([]Mutation, 0, count)
	for i := 1; i <= count; i++ {
		newFieldDeviceID := historyBatchUUID(9000 + i)
		if i == 1 {
			insertBacnetMoveDestination(
				t,
				db,
				newFieldDeviceID,
				newSystemTypeID,
				newSPSControllerID,
				newControlCabinetID,
				newBuildingID,
				newFieldDeviceProjectID,
				newSPSProjectID,
				newCabinetProjectID,
			)
		} else {
			if err := db.Exec(
				`INSERT INTO field_devices (id, sps_controller_system_type_id) VALUES (?, ?)`,
				newFieldDeviceID,
				newSystemTypeID,
			).Error; err != nil {
				t.Fatalf("insert destination FieldDevice %d: %v", i, err)
			}
			if err := db.Exec(
				`INSERT INTO project_field_devices (field_device_id, project_id) VALUES (?, ?)`,
				newFieldDeviceID,
				newFieldDeviceProjectID,
			).Error; err != nil {
				t.Fatalf("insert destination FieldDevice project link %d: %v", i, err)
			}
		}

		bacnetObjectID := historyBatchUUID(3000 + i)
		mutations = append(mutations, Mutation{
			Action:      domainHistory.ActionUpdate,
			EntityTable: "bacnet_objects",
			EntityID:    bacnetObjectID,
			BeforeJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","field_device_id":"%s"}`,
				bacnetObjectID,
				historyBatchUUID(i),
			)),
			AfterJSON: domainHistory.JSONB(fmt.Sprintf(
				`{"id":"%s","field_device_id":"%s"}`,
				bacnetObjectID,
				newFieldDeviceID,
			)),
		})
	}

	counter.reset()
	if err := NewStore(db).RecordMutations(context.Background(), mutations); err != nil {
		t.Fatalf("record %d BACnet object moves: %v", count, err)
	}
	queryCount := counter.count.Load()

	var scopeCount int64
	if err := db.Model(&domainHistory.ChangeEventScope{}).Count(&scopeCount).Error; err != nil {
		t.Fatalf("count scopes: %v", err)
	}
	const scopesPerMove = 17
	if scopeCount != int64(count*scopesPerMove) {
		t.Fatalf("scopes: got %d, want %d", scopeCount, count*scopesPerMove)
	}
	return queryCount
}

func insertBacnetMoveDestination(
	t *testing.T,
	db *gorm.DB,
	fieldDeviceID uuid.UUID,
	systemTypeID uuid.UUID,
	spsControllerID uuid.UUID,
	controlCabinetID uuid.UUID,
	buildingID uuid.UUID,
	fieldDeviceProjectID uuid.UUID,
	spsProjectID uuid.UUID,
	cabinetProjectID uuid.UUID,
) {
	t.Helper()
	for statement, args := range map[string][]any{
		`INSERT INTO sps_controller_system_types (id, sps_controller_id) VALUES (?, ?)`: {
			systemTypeID,
			spsControllerID,
		},
		`INSERT INTO sps_controllers (id, control_cabinet_id) VALUES (?, ?)`: {
			spsControllerID,
			controlCabinetID,
		},
		`INSERT INTO control_cabinets (id, building_id) VALUES (?, ?)`: {
			controlCabinetID,
			buildingID,
		},
		`INSERT INTO field_devices (id, sps_controller_system_type_id) VALUES (?, ?)`: {
			fieldDeviceID,
			systemTypeID,
		},
		`INSERT INTO project_field_devices (field_device_id, project_id) VALUES (?, ?)`: {
			fieldDeviceID,
			fieldDeviceProjectID,
		},
		`INSERT INTO project_sps_controllers (sps_controller_id, project_id) VALUES (?, ?)`: {
			spsControllerID,
			spsProjectID,
		},
		`INSERT INTO project_control_cabinets (control_cabinet_id, project_id) VALUES (?, ?)`: {
			controlCabinetID,
			cabinetProjectID,
		},
	} {
		if err := db.Exec(statement, args...).Error; err != nil {
			t.Fatalf("insert BACnet move destination fixture: %v", err)
		}
	}
}

func newHistoryBatchTestDB(
	t *testing.T,
	counter *historyQueryCounter,
	fieldDeviceCount int,
) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf(
		"file:%s?mode=memory&cache=shared",
		strings.NewReplacer("/", "_", " ", "_", "#", "_").Replace(t.Name()+fmt.Sprint(fieldDeviceCount)),
	)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   counter,
	})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql database: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(
		&domainHistory.ChangeEvent{},
		&domainHistory.ChangeEventScope{},
		&domainHistory.EntityVersion{},
	); err != nil {
		t.Fatalf("migrate history tables: %v", err)
	}
	for _, statement := range []string{
		`CREATE TABLE sps_controller_system_types (id TEXT PRIMARY KEY, sps_controller_id TEXT)`,
		`CREATE TABLE sps_controllers (id TEXT PRIMARY KEY, control_cabinet_id TEXT)`,
		`CREATE TABLE control_cabinets (id TEXT PRIMARY KEY, building_id TEXT)`,
		`CREATE TABLE field_devices (id TEXT PRIMARY KEY, sps_controller_system_type_id TEXT)`,
		`CREATE TABLE bacnet_objects (id TEXT PRIMARY KEY, field_device_id TEXT)`,
		`CREATE TABLE object_data (id TEXT PRIMARY KEY, project_id TEXT)`,
		`CREATE TABLE object_data_bacnet_objects (bacnet_object_id TEXT, object_data_id TEXT)`,
		`CREATE TABLE project_field_devices (field_device_id TEXT, project_id TEXT)`,
		`CREATE TABLE project_sps_controllers (sps_controller_id TEXT, project_id TEXT)`,
		`CREATE TABLE project_control_cabinets (control_cabinet_id TEXT, project_id TEXT)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("create scope fixture table: %v", err)
		}
	}

	systemTypeID := historyBatchUUID(100)
	spsControllerID := historyBatchUUID(101)
	controlCabinetID := historyBatchUUID(102)
	buildingID := historyBatchUUID(103)
	if err := db.Exec(
		`INSERT INTO sps_controller_system_types (id, sps_controller_id) VALUES (?, ?)`,
		systemTypeID,
		spsControllerID,
	).Error; err != nil {
		t.Fatalf("insert system type: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO sps_controllers (id, control_cabinet_id) VALUES (?, ?)`,
		spsControllerID,
		controlCabinetID,
	).Error; err != nil {
		t.Fatalf("insert SPS controller: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO control_cabinets (id, building_id) VALUES (?, ?)`,
		controlCabinetID,
		buildingID,
	).Error; err != nil {
		t.Fatalf("insert control cabinet: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO project_sps_controllers (sps_controller_id, project_id) VALUES (?, ?)`,
		spsControllerID,
		historyBatchUUID(201),
	).Error; err != nil {
		t.Fatalf("insert SPS project link: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO project_control_cabinets (control_cabinet_id, project_id) VALUES (?, ?)`,
		controlCabinetID,
		historyBatchUUID(202),
	).Error; err != nil {
		t.Fatalf("insert cabinet project link: %v", err)
	}
	for i := 1; i <= fieldDeviceCount; i++ {
		if err := db.Exec(
			`INSERT INTO field_devices (id, sps_controller_system_type_id) VALUES (?, ?)`,
			historyBatchUUID(i),
			systemTypeID,
		).Error; err != nil {
			t.Fatalf("insert FieldDevice: %v", err)
		}
		if err := db.Exec(
			`INSERT INTO project_field_devices (field_device_id, project_id) VALUES (?, ?)`,
			historyBatchUUID(i),
			historyBatchUUID(200),
		).Error; err != nil {
			t.Fatalf("insert FieldDevice project link: %v", err)
		}
	}
	return db
}

func historyBatchUUID(value int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", value))
}
