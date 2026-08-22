package wire

import (
	"encoding/json"
	"testing"

	facilityjobs "github.com/besart951/go_infra_link/backend/internal/application/facilityjobs"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

func TestFieldDeviceBulkPlanRoundTripPreservesAtomicGroups(t *testing.T) {
	job := facilityservice.FacilityJob{ID: uuid.New(), OwnerID: uuid.New()}
	groupID := uuid.New()
	updates := []domainFacility.BulkFieldDeviceUpdate{{ID: uuid.New()}, {ID: uuid.New()}}
	groups := []facilityservice.FieldDeviceBulkUpdateGroup{
		{ID: groupID, Indexes: []int{7, 8}, Updates: updates},
	}
	items, err := persistedGroupItems(job, groups[0])
	if err != nil {
		t.Fatal(err)
	}
	restored, err := updateGroupsFromPlan(items)
	if err != nil {
		t.Fatal(err)
	}
	if len(restored) != 1 || restored[0].ID != groupID || len(restored[0].Updates) != 2 {
		t.Fatalf("restored groups = %+v", restored)
	}
	if restored[0].Indexes[0] != 7 || restored[0].Indexes[1] != 8 {
		t.Fatalf("restored indexes = %v", restored[0].Indexes)
	}
}

func persistedGroupItems(job facilityservice.FacilityJob, group facilityservice.FieldDeviceBulkUpdateGroup) ([]facilityjobs.FieldDeviceUpdatePlanItem, error) {
	items := make([]facilityjobs.FieldDeviceUpdatePlanItem, len(group.Updates))
	for index, update := range group.Updates {
		command, err := json.Marshal(update)
		if err != nil {
			return nil, err
		}
		items[index] = facilityjobs.FieldDeviceUpdatePlanItem{
			OwnerID: job.OwnerID, JobID: job.ID, Ordinal: int64(group.Indexes[index]), GroupOrdinal: 0,
			DependencyGroupID: group.ID, FieldDeviceID: update.ID, Command: command,
		}
	}
	return items, nil
}
