package gormbase

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestBaseRepositoryUsesPostgresRevisionCompareAndSwap(t *testing.T) {
	dsn := os.Getenv("FACILITY_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("FACILITY_TEST_DATABASE_URL is not set")
	}
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	tx := database.Begin()
	if tx.Error != nil {
		t.Fatalf("begin test transaction: %v", tx.Error)
	}
	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})

	buildingID := uuid.New()
	if err := tx.Exec(`
		INSERT INTO buildings (id, iws_code, building_group)
		VALUES (?, ?, 92)
	`, buildingID, "REV-"+buildingID.String()).Error; err != nil {
		t.Fatalf("seed building: %v", err)
	}
	number := "REV-" + uuid.New().String()
	cabinet := &domainFacility.ControlCabinet{
		BuildingID:       buildingID,
		ControlCabinetNr: &number,
	}
	repository := NewBaseRepository[*domainFacility.ControlCabinet](tx, nil)
	if err := repository.Create(context.Background(), cabinet); err != nil {
		t.Fatalf("create cabinet: %v", err)
	}
	if cabinet.Revision != 1 {
		t.Fatalf("created revision: got %d, want 1", cabinet.Revision)
	}

	stale := *cabinet
	firstNumber := "REV-1"
	cabinet.ControlCabinetNr = &firstNumber
	if err := repository.Update(context.Background(), cabinet); err != nil {
		t.Fatalf("first compare-and-swap: %v", err)
	}
	if cabinet.Revision != 2 {
		t.Fatalf("updated revision: got %d, want 2", cabinet.Revision)
	}

	staleNumber := "REV-STALE"
	stale.ControlCabinetNr = &staleNumber
	err = repository.Update(context.Background(), &stale)
	var conflict *domain.RevisionConflict
	if !errors.As(err, &conflict) ||
		conflict.Expected != 1 ||
		conflict.Current != 2 ||
		conflict.EntityID != cabinet.ID {
		t.Fatalf("stale update error: %#v", err)
	}

	// Simulates an old blue-green binary that does not know about revision.
	if err := tx.Exec(`
		UPDATE control_cabinets
		SET control_cabinet_nr = 'REV-OLD-BINARY'
		WHERE id = ?
	`, cabinet.ID).Error; err != nil {
		t.Fatalf("old-binary update: %v", err)
	}
	var revision uint64
	if err := tx.Raw(`
		SELECT revision
		FROM control_cabinets
		WHERE id = ?
	`, cabinet.ID).Scan(&revision).Error; err != nil {
		t.Fatalf("read old-binary revision: %v", err)
	}
	if revision != 3 {
		t.Fatalf("old-binary revision: got %d, want 3", revision)
	}
}
