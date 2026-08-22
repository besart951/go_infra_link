package facilityjobsql

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	facilityjobs "github.com/besart951/go_infra_link/backend/internal/application/facilityjobs"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFieldDeviceUpdatePlanStorePersistsStableGroupOrder(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&fieldDeviceUpdatePlanRecord{}); err != nil {
		t.Fatal(err)
	}
	store := NewFieldDeviceUpdatePlanStore(database)
	ownerID, jobID, groupID := uuid.New(), uuid.New(), uuid.New()
	fixture := planFixture{ownerID: ownerID, jobID: jobID, groupID: groupID}
	items := []facilityjobs.FieldDeviceUpdatePlanItem{
		fixture.item(2),
		fixture.item(1),
	}
	if err := store.Save(t.Context(), items); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.List(t.Context(), ownerID, jobID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 2 || loaded[0].Ordinal != 1 || loaded[1].Ordinal != 2 {
		t.Fatalf("unexpected plan order: %+v", loaded)
	}
}

func TestFieldDeviceUpdatePlanStorePlansDependenciesInPostgres(t *testing.T) {
	dsn := os.Getenv("FACILITY_BENCHMARK_DSN")
	if dsn == "" {
		t.Skip("FACILITY_BENCHMARK_DSN is not configured")
	}
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	tx := database.Begin()
	defer tx.Rollback()
	devices := postgresPlanDevices(t, tx)
	ownerID, jobID := uuid.New(), uuid.New()
	insertPlanJob(t, tx, ownerID, jobID)
	store := NewFieldDeviceUpdatePlanStore(tx)
	items := postgresPlanItems(t, ownerID, jobID, devices)
	if err := store.Save(t.Context(), items); err != nil {
		t.Fatal(err)
	}
	if err := store.Plan(t.Context(), ownerID, jobID); err != nil {
		t.Fatal(err)
	}
	planned, err := store.List(t.Context(), ownerID, jobID)
	if err != nil {
		t.Fatal(err)
	}
	if len(planned) != 3 || planned[0].DependencyGroupID != planned[1].DependencyGroupID || planned[1].DependencyGroupID == planned[2].DependencyGroupID {
		t.Fatalf("unexpected relational groups: %+v", planned)
	}
}

type postgresPlanDevice struct {
	ID        uuid.UUID
	ApparatNr int
}

func postgresPlanDevices(t *testing.T, database *gorm.DB) []postgresPlanDevice {
	t.Helper()
	var devices []postgresPlanDevice
	err := database.Raw(`SELECT id,apparat_nr FROM field_devices WHERE (sps_controller_system_type_id,system_part_id,apparat_id)=(SELECT sps_controller_system_type_id,system_part_id,apparat_id FROM field_devices LIMIT 1) ORDER BY apparat_nr LIMIT 3`).Scan(&devices).Error
	if err != nil || len(devices) != 3 {
		t.Fatalf("load PostgreSQL plan fixtures: rows=%d err=%v", len(devices), err)
	}
	return devices
}

func insertPlanJob(t *testing.T, database *gorm.DB, ownerID, jobID uuid.UUID) {
	t.Helper()
	now := time.Now().UTC()
	err := database.Exec(`INSERT INTO facility_jobs(owner_id,id,kind,class,job_type,status,progress,stage,attempts,processed,succeeded,failed,retryable,created_at,updated_at) VALUES (?,?,?,'mutation','bulk','running',0,'planning',0,0,0,0,true,?,?)`, ownerID, jobID, "field_device", now, now).Error
	if err != nil {
		t.Fatal(err)
	}
}

func postgresPlanItems(t *testing.T, ownerID, jobID uuid.UUID, devices []postgresPlanDevice) []facilityjobs.FieldDeviceUpdatePlanItem {
	t.Helper()
	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: devices[0].ID, BaseVersion: domain.AggregateVersion(1), ApparatNr: &devices[1].ApparatNr},
		{ID: devices[1].ID, BaseVersion: domain.AggregateVersion(1), ApparatNr: &devices[0].ApparatNr},
		{ID: devices[2].ID, BaseVersion: domain.AggregateVersion(1), BMK: stringPlanPointer("independent"), HasBMK: true},
	}
	items := make([]facilityjobs.FieldDeviceUpdatePlanItem, len(updates))
	for index, update := range updates {
		command, err := json.Marshal(update)
		if err != nil {
			t.Fatal(err)
		}
		items[index] = facilityjobs.FieldDeviceUpdatePlanItem{
			OwnerID: ownerID, JobID: jobID, Ordinal: int64(index), GroupOrdinal: int64(index),
			DependencyGroupID: uuid.New(), FieldDeviceID: update.ID, Command: command,
		}
	}
	return items
}

func stringPlanPointer(value string) *string { return &value }

type planFixture struct {
	ownerID uuid.UUID
	jobID   uuid.UUID
	groupID uuid.UUID
}

func (fixture planFixture) item(ordinal int64) facilityjobs.FieldDeviceUpdatePlanItem {
	return facilityjobs.FieldDeviceUpdatePlanItem{
		OwnerID: fixture.ownerID, JobID: fixture.jobID, Ordinal: ordinal, GroupOrdinal: 0,
		DependencyGroupID: fixture.groupID, FieldDeviceID: uuid.New(), Command: json.RawMessage(`{"id":"item"}`),
	}
}
