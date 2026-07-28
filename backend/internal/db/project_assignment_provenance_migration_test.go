package db

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestProjectAssignmentProvenanceMigrationIsRegisteredAsAdditive(t *testing.T) {
	for _, migration := range migrations {
		if migration.version != "202607230007" {
			continue
		}
		if !migration.blueGreenCompatible {
			t.Fatal("project assignment provenance expand migration must be blue-green compatible")
		}
		return
	}
	t.Fatal("project assignment provenance migration is not registered")
}

func TestProjectAssignmentProvenanceMigrationIsNoOpOutsidePostgres(t *testing.T) {
	database, err := gorm.Open(
		sqlite.Open("file:project_assignment_provenance?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	if err := migrateProjectAssignmentProvenance(database); err != nil {
		t.Fatalf("SQLite no-op: %v", err)
	}
}
