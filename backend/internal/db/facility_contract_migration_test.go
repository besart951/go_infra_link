package db

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFacilityContractRejectsNonPostgresDatabase(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = ApplyFacilityContractMigration(database, validFacilityContractOptions())
	if err == nil || !strings.Contains(err.Error(), "requires PostgreSQL") {
		t.Fatalf("error = %v", err)
	}
}

func TestFacilityContractRequiresReleaseBackupAndIdleWindow(t *testing.T) {
	options := FacilityContractOptions{}
	if err := validateContractOptions(postgresDialectorStub(t), options); err == nil {
		t.Fatal("expected release gate failure")
	}
}

func TestFacilityContractDuplicatePreflightCastsUUIDBeforeMin(t *testing.T) {
	for _, check := range facilityContractChecks {
		if check.name != "duplicate FieldDevice number keys" {
			continue
		}
		if !strings.Contains(check.sql, "min(id::text)") {
			t.Fatalf("duplicate preflight is not PostgreSQL UUID-safe: %s", check.sql)
		}
		return
	}
	t.Fatal("duplicate FieldDevice preflight is missing")
}

func validFacilityContractOptions() FacilityContractOptions {
	return FacilityContractOptions{
		CompatibleReleaseDelivered: true,
		ApplicationsStopped:        true,
		LegacyIdleSince:            time.Now().Add(-15 * 24 * time.Hour),
		BackupVerified:             true,
	}
}

func postgresDialectorStub(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	database.Config.Dialector = namedDialector{Dialector: database.Dialector, name: "postgres"}
	return database
}

type namedDialector struct {
	gorm.Dialector
	name string
}

func (d namedDialector) Name() string { return d.name }
