package db

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestFacilityContractAllowsDeferredFieldDeviceNumberSwapPostgres(t *testing.T) {
	database := benchmarkContractDatabase(t)
	tx := database.Begin()
	defer tx.Rollback()
	devices := contractSwapDevices(t, tx)
	if err := tx.Exec(`SET CONSTRAINTS uq_field_devices_number_scope DEFERRED`).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Exec(`UPDATE field_devices SET apparat_nr=? WHERE id=?`, devices[1].Number, devices[0].ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Exec(`UPDATE field_devices SET apparat_nr=? WHERE id=?`, devices[0].Number, devices[1].ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Exec(`SET CONSTRAINTS uq_field_devices_number_scope IMMEDIATE`).Error; err != nil {
		t.Fatalf("deferred swap did not satisfy the final invariant: %v", err)
	}
}

type contractSwapDevice struct {
	ID     uuid.UUID
	Number int `gorm:"column:apparat_nr"`
}

func contractSwapDevices(t *testing.T, database *gorm.DB) []contractSwapDevice {
	t.Helper()
	var devices []contractSwapDevice
	err := database.Raw(`SELECT id,apparat_nr FROM field_devices WHERE (sps_controller_system_type_id,system_part_id,apparat_id)=(SELECT sps_controller_system_type_id,system_part_id,apparat_id FROM field_devices LIMIT 1) ORDER BY apparat_nr LIMIT 2`).Scan(&devices).Error
	if err != nil || len(devices) != 2 {
		t.Fatalf("load swap fixtures: rows=%d err=%v", len(devices), err)
	}
	return devices
}

func benchmarkContractDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("FACILITY_BENCHMARK_DSN")
	if dsn == "" {
		t.Skip("FACILITY_BENCHMARK_DSN is not configured")
	}
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return database
}
