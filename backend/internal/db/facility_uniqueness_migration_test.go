package db

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFacilityUniquenessMigrationsAreNamedAndRegistered(t *testing.T) {
	if fieldDevicePlacementConstraint != "uq_field_devices_placement_apparat_nr" {
		t.Fatalf("placement constraint: %q", fieldDevicePlacementConstraint)
	}
	if !strings.Contains(spsDeviceNameNormalizedIndex, "device_name_normalized") ||
		!strings.Contains(spsGADeviceNormalizedIndex, "ga_device_normalized") {
		t.Fatalf(
			"normalized SPS indexes: device=%q ga=%q",
			spsDeviceNameNormalizedIndex,
			spsGADeviceNormalizedIndex,
		)
	}

	want := map[string]bool{
		"202607230003": false,
		"202607230004": false,
	}
	for _, migration := range migrations {
		if _, ok := want[migration.version]; ok {
			want[migration.version] = true
			if !migration.blueGreenCompatible {
				t.Fatalf("migration %s must remain additive for blue-green rollout", migration.version)
			}
		}
	}
	for version, found := range want {
		if !found {
			t.Fatalf("migration %s is not registered", version)
		}
	}
}

func TestFacilityUniquenessMigrationsAreNoOpsOutsidePostgres(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:facility_uniqueness?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	if err := migrateFieldDevicePlacementUniqueness(database); err != nil {
		t.Fatalf("placement migration SQLite no-op: %v", err)
	}
	if err := migrateSPSControllerNormalizedUniqueness(database); err != nil {
		t.Fatalf("SPS migration SQLite no-op: %v", err)
	}
}
