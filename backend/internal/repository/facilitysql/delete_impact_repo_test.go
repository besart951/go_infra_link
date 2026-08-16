package facilitysql

import (
	"context"
	"testing"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func TestDeleteImpactRepositoryCountsDirectFacilityReferences(t *testing.T) {
	db := newBacnetReferenceUsageRepoTestDB(t)
	repo := NewDeleteImpactRepository(db)

	apparatID := uuid.New()
	systemPartID := uuid.New()
	if err := db.Exec("INSERT INTO system_parts (id, short_name, name) VALUES (?, ?, ?)", systemPartID, "SYS", "System").Error; err != nil {
		t.Fatalf("seed system part: %v", err)
	}
	if err := db.Exec("INSERT INTO apparats (id, short_name, name) VALUES (?, ?, ?)", apparatID, "APP", "Apparat").Error; err != nil {
		t.Fatalf("seed apparat: %v", err)
	}
	if err := db.Exec("INSERT INTO system_part_apparats (system_part_id, apparat_id) VALUES (?, ?)", systemPartID, apparatID).Error; err != nil {
		t.Fatalf("seed relation: %v", err)
	}
	if err := db.Exec("INSERT INTO field_devices (id, sps_controller_system_type_id, system_part_id, apparat_id, apparat_nr) VALUES (?, ?, ?, ?, ?)", uuid.New(), uuid.New(), systemPartID, apparatID, 1).Error; err != nil {
		t.Fatalf("seed field device: %v", err)
	}
	objectDataID := uuid.New()
	if err := db.Exec("INSERT INTO object_data (id, description, obj_version, is_active) VALUES (?, ?, ?, ?)", objectDataID, "Object data", "1", true).Error; err != nil {
		t.Fatalf("seed object data relation owner: %v", err)
	}
	if err := db.Exec("INSERT INTO object_data_apparats (object_data_id, apparat_id) VALUES (?, ?)", objectDataID, apparatID).Error; err != nil {
		t.Fatalf("seed object data relation: %v", err)
	}

	assertDeleteImpactBlockers(t, repo, domainFacility.DeleteImpactResourceApparat, apparatID, map[string]int64{
		"field_devices": 1, "system_parts": 1, "object_data": 1,
	})
	assertDeleteImpactBlockers(t, repo, domainFacility.DeleteImpactResourceSystemPart, systemPartID, map[string]int64{
		"field_devices": 1, "apparats": 1,
	})
}

func assertDeleteImpactBlockers(t *testing.T, repo domainFacility.DeleteImpactRepository, resource domainFacility.DeleteImpactResource, id uuid.UUID, want map[string]int64) {
	t.Helper()
	impacts, err := repo.DeleteImpacts(context.Background(), resource, []uuid.UUID{id})
	if err != nil {
		t.Fatalf("DeleteImpacts() error = %v", err)
	}
	if len(impacts) != 1 {
		t.Fatalf("DeleteImpacts() returned %d impacts, want 1", len(impacts))
	}
	got := make(map[string]int64, len(impacts[0].Blockers))
	for _, blocker := range impacts[0].Blockers {
		got[blocker.Resource] = blocker.Count
	}
	for resource, count := range want {
		if got[resource] != count {
			t.Fatalf("blocker %q = %d, want %d; all = %#v", resource, got[resource], count, got)
		}
	}
}
