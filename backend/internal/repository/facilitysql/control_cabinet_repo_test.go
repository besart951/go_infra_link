package facilitysql

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestControlCabinetRepo_GetPaginatedListSearchesBuildingFields(t *testing.T) {
	tests := []struct {
		name   string
		search string
	}{
		{name: "iws code", search: "plant-a"},
		{name: "building group", search: "42"},
		{name: "building label", search: "plant-a-42"},
		{name: "visible combobox label", search: "plant-a-42 cab-alpha"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := newControlCabinetRepoTestDB(t)
			repo := NewControlCabinetRepository(db)

			targetBuilding := seedControlCabinetRepoRecord(t, db, &domainFacility.Building{
				IWSCode:       "PLANT-A",
				BuildingGroup: 42,
			})
			otherBuilding := seedControlCabinetRepoRecord(t, db, &domainFacility.Building{
				IWSCode:       "PLANT-B",
				BuildingGroup: 99,
			})

			targetNr := "CAB-ALPHA"
			targetCabinet := seedControlCabinetRepoRecord(t, db, &domainFacility.ControlCabinet{
				BuildingID:       targetBuilding.ID,
				ControlCabinetNr: &targetNr,
			})
			otherNr := "CAB-BETA"
			seedControlCabinetRepoRecord(t, db, &domainFacility.ControlCabinet{
				BuildingID:       otherBuilding.ID,
				ControlCabinetNr: &otherNr,
			})

			list, err := repo.GetPaginatedList(ctx, domain.PaginationParams{
				Page:   1,
				Limit:  10,
				Search: tt.search,
			})
			if err != nil {
				t.Fatalf("expected control cabinet list search to succeed, got %v", err)
			}
			if len(list.Items) != 1 {
				t.Fatalf("expected one control cabinet for search %q, got %+v", tt.search, list.Items)
			}
			if list.Items[0].ID != targetCabinet.ID {
				t.Fatalf("expected target control cabinet %s, got %s", targetCabinet.ID, list.Items[0].ID)
			}
		})
	}
}

func newControlCabinetRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.NewReplacer("/", "_", " ", "_", "#", "_").Replace(t.Name()))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("expected sqlite db to open, got %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("expected sql db handle, got %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(&domainFacility.Building{}, &domainFacility.ControlCabinet{}); err != nil {
		t.Fatalf("expected control cabinet repo tables to migrate, got %v", err)
	}

	return db
}

func seedControlCabinetRepoRecord[T interface{ GetBase() *domain.Base }](t *testing.T, db *gorm.DB, entity T) T {
	t.Helper()

	if err := entity.GetBase().InitForCreate(time.Now().UTC()); err != nil {
		t.Fatalf("expected base init to succeed, got %v", err)
	}
	if err := db.Create(entity).Error; err != nil {
		t.Fatalf("expected record seed to succeed, got %v", err)
	}
	return entity
}
