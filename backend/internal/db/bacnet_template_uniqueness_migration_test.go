package db

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBacnetTemplateUniquenessDoesNotConstrainTextFix(t *testing.T) {
	for _, identifier := range []string{
		objectDataBacnetSoftwareIndex,
		objectDataBacnetSyncFunction,
		bacnetObjectTemplateFunction,
	} {
		if strings.Contains(strings.ToLower(identifier), "text") {
			t.Fatalf("template uniqueness must not constrain TextFix: %q", identifier)
		}
	}

	found := false
	for _, migration := range migrations {
		if migration.version == "202607230005" {
			found = true
			if !migration.blueGreenCompatible {
				t.Fatal("template software-key migration must be additive")
			}
		}
	}
	if !found {
		t.Fatal("template software-key migration is not registered")
	}
}

func TestBacnetTemplateUniquenessMigrationIsNoOpOutsidePostgres(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:bacnet_template_uniqueness?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	if err := migrateBacnetTemplateUniqueness(database); err != nil {
		t.Fatalf("SQLite no-op: %v", err)
	}
}
