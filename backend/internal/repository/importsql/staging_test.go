package importsql

import (
	"context"
	"testing"
	"time"

	fielddeviceimport "github.com/besart951/go_infra_link/backend/internal/application/fielddeviceimport"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCompleteRemovesPayloadRowsButKeepsSessionMetadata(t *testing.T) {
	db := newStagingTestDB(t)
	fixture := newStagingFixture(t, db)
	id, session, err := NewStore(db).Start(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	stageFixture(t, session, fixture)

	if err := session.Complete(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertImportRecordCounts(t, db, importRecordExpectation{id: id, sessions: 1})
}

func TestCleanupRemovesOnlyExpiredSessions(t *testing.T) {
	db := newStagingTestDB(t)
	store := NewStore(db)
	oldID, _, _ := store.Start(context.Background(), uuid.New())
	newID, _, _ := store.Start(context.Background(), uuid.New())
	db.Model(&sessionRecord{}).Where("id = ?", oldID).Update("updated_at", time.Now().Add(-91*24*time.Hour))

	if err := store.Cleanup(context.Background(), time.Now().Add(-90*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	assertImportRecordCounts(t, db, importRecordExpectation{id: oldID})
	assertImportRecordCounts(t, db, importRecordExpectation{id: newID, sessions: 1})
}

type importRecordExpectation struct {
	id       uuid.UUID
	sessions int64
	rows     int64
}

func assertImportRecordCounts(t *testing.T, db *gorm.DB, expected importRecordExpectation) {
	t.Helper()
	var sessionCount, rowCount int64
	db.Model(&sessionRecord{}).Where("id = ?", expected.id).Count(&sessionCount)
	db.Model(&rowRecord{}).Where("import_id = ?", expected.id).Count(&rowCount)
	if sessionCount != expected.sessions || rowCount != expected.rows {
		t.Fatalf("sessions=%d rows=%d", sessionCount, rowCount)
	}
}

func TestSessionValidatesAndPagesCompleteAggregate(t *testing.T) {
	db := newStagingTestDB(t)
	fixture := newStagingFixture(t, db)
	_, session, err := NewStore(db).Start(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	stageFixture(t, session, fixture)
	if err := session.Seal(context.Background(), fixture.manifest(1)); err != nil {
		t.Fatal(err)
	}

	issues, err := session.Validate(context.Background())
	if err != nil || len(issues) != 0 {
		t.Fatalf("issues=%+v err=%v", issues, err)
	}
	page, err := session.Aggregates(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || len(page.Items[0].BacnetObjects) != 2 || alarmCount(page.Items[0].BacnetObjects) != 1 {
		t.Fatalf("aggregate was not reconstructed: %+v", page)
	}
}

func alarmCount(objects []domainFacility.BacnetObject) int {
	total := 0
	for index := range objects {
		total += len(objects[index].AlarmValues)
	}
	return total
}

func TestSessionRejectsCrossAggregateSoftwareReference(t *testing.T) {
	db := newStagingTestDB(t)
	fixture := newStagingFixture(t, db)
	otherDeviceID, otherObjectID := uuid.New(), uuid.New()
	fixture.devices = append(fixture.devices, domainFacility.FieldDevice{Base: domain.Base{ID: otherDeviceID, Version: 1}, SPSControllerSystemTypeID: fixture.assignmentID, SystemPartID: fixture.systemPartID, ApparatID: fixture.apparatID, ApparatNr: 8})
	fixture.objects = append(fixture.objects, domainFacility.BacnetObject{Base: domain.Base{ID: otherObjectID, Version: 1}, FieldDeviceID: &otherDeviceID, TextFix: "Other", SoftwareType: domainFacility.BacnetSoftwareTypeAV})
	fixture.objects[0].SoftwareReferenceID = &otherObjectID
	fixture.references[0].TargetObjectID = otherObjectID
	_, session, _ := NewStore(db).Start(context.Background(), uuid.New())
	stageFixture(t, session, fixture)
	if err := session.Seal(context.Background(), fixture.manifest(2)); err != nil {
		t.Fatal(err)
	}

	issues, err := session.Validate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !hasIssue(issues, "invalid_software_reference") {
		t.Fatalf("expected cross-owner issue, got %+v", issues)
	}
}

type stagingFixture struct {
	devices      []domainFacility.FieldDevice
	specs        []domainFacility.Specification
	objects      []domainFacility.BacnetObject
	references   []fielddeviceimport.SoftwareReference
	alarms       []domainFacility.BacnetObjectAlarmValue
	assignmentID uuid.UUID
	systemPartID uuid.UUID
	apparatID    uuid.UUID
}

func (f stagingFixture) manifest(deviceCount int64) fielddeviceimport.Manifest {
	return fielddeviceimport.Manifest{SchemaVersion: 2, DeviceCount: deviceCount, Counts: fielddeviceimport.Counts{
		Specifications: int64(len(f.specs)), BacnetObjects: int64(len(f.objects)),
		SoftwareReferences: int64(len(f.references)), AlarmValues: int64(len(f.alarms)),
	}}
}

func newStagingFixture(t *testing.T, db *gorm.DB) stagingFixture {
	t.Helper()
	deviceID, objectID, targetID := uuid.New(), uuid.New(), uuid.New()
	assignmentID, systemPartID, apparatID := uuid.New(), uuid.New(), uuid.New()
	alarmFieldID := uuid.New()
	insertReference(t, db, "sps_controller_system_types", assignmentID)
	insertReference(t, db, "system_parts", systemPartID)
	insertReference(t, db, "apparats", apparatID)
	insertReference(t, db, "alarm_type_fields", alarmFieldID)
	return stagingFixture{
		devices: []domainFacility.FieldDevice{{Base: domain.Base{ID: deviceID, Version: 1}, SPSControllerSystemTypeID: assignmentID, SystemPartID: systemPartID, ApparatID: apparatID, ApparatNr: 7}},
		specs:   []domainFacility.Specification{{Base: domain.Base{ID: uuid.New(), Version: 1}, FieldDeviceID: &deviceID}},
		objects: []domainFacility.BacnetObject{
			{Base: domain.Base{ID: objectID, Version: 1}, FieldDeviceID: &deviceID, TextFix: "Source", SoftwareType: domainFacility.BacnetSoftwareTypeAI, SoftwareReferenceID: &targetID},
			{Base: domain.Base{ID: targetID, Version: 1}, FieldDeviceID: &deviceID, TextFix: "Target", SoftwareType: domainFacility.BacnetSoftwareTypeAV},
		},
		references:   []fielddeviceimport.SoftwareReference{{SourceObjectID: objectID, TargetObjectID: targetID, FieldDeviceID: deviceID}},
		alarms:       []domainFacility.BacnetObjectAlarmValue{{Base: domain.Base{ID: uuid.New(), Version: 1}, BacnetObjectID: objectID, AlarmTypeFieldID: alarmFieldID, Source: domainFacility.AlarmValueSourceImport}},
		assignmentID: assignmentID, systemPartID: systemPartID, apparatID: apparatID,
	}
}

func stageFixture(t *testing.T, session fielddeviceimport.Session, fixture stagingFixture) {
	t.Helper()
	ctx := context.Background()
	for _, action := range []func() error{
		func() error { return session.FieldDevices(ctx, fixture.devices) },
		func() error { return session.Specifications(ctx, fixture.specs) },
		func() error { return session.BacnetObjects(ctx, fixture.objects) },
		func() error { return session.SoftwareReferences(ctx, fixture.references) },
		func() error { return session.AlarmValues(ctx, fixture.alarms) },
	} {
		if err := action(); err != nil {
			t.Fatal(err)
		}
	}
}

func newStagingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"field_devices", "specifications", "bacnet_objects", "bacnet_object_alarm_values", "sps_controller_system_types", "system_parts", "apparats", "alarm_type_fields", "state_texts", "notification_classes", "alarm_types", "units"} {
		if err := db.Exec("CREATE TABLE " + table + " (id text primary key)").Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func insertReference(t *testing.T, db *gorm.DB, table string, id uuid.UUID) {
	t.Helper()
	if err := db.Exec("INSERT INTO "+table+" (id) VALUES (?)", id).Error; err != nil {
		t.Fatal(err)
	}
}

func hasIssue(issues []fielddeviceimport.Issue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}
