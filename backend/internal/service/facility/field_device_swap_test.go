package facility

import (
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestSwapPlannerGroupsCyclesAndLeavesIndependentUpdatesAlone(t *testing.T) {
	systemTypeID, systemPartID, apparatID := uuid.New(), uuid.New(), uuid.New()
	firstID, secondID, independentID := uuid.New(), uuid.New(), uuid.New()
	existing := map[uuid.UUID]*domainFacility.FieldDevice{
		firstID:       deviceAt(firstID, systemTypeID, systemPartID, apparatID, 1),
		secondID:      deviceAt(secondID, systemTypeID, systemPartID, apparatID, 2),
		independentID: deviceAt(independentID, systemTypeID, systemPartID, apparatID, 3),
	}
	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: firstID, ApparatNr: intPointer(2)},
		{ID: secondID, ApparatNr: intPointer(1)},
		{ID: independentID, BMK: stringPointer("new")},
	}
	proposed := proposedDevices(existing, updates)

	groups := planFieldDeviceUpdateGroups(updates, existing, proposed)

	if len(groups) != 2 || len(groups[0].Indexes) != 2 || len(groups[1].Indexes) != 1 {
		t.Fatalf("unexpected groups: %+v", groups)
	}
	if groups[0].ID != dependencyGroupID(updates, []int{1, 0}) {
		t.Fatalf("group ID must be deterministic, got %s", groups[0].ID)
	}
}

func TestSwapPlannerGroupsDuplicateTargetsForAtomicFailure(t *testing.T) {
	systemTypeID, systemPartID, apparatID := uuid.New(), uuid.New(), uuid.New()
	firstID, secondID := uuid.New(), uuid.New()
	existing := map[uuid.UUID]*domainFacility.FieldDevice{
		firstID:  deviceAt(firstID, systemTypeID, systemPartID, apparatID, 1),
		secondID: deviceAt(secondID, systemTypeID, systemPartID, apparatID, 2),
	}
	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: firstID, ApparatNr: intPointer(3)},
		{ID: secondID, ApparatNr: intPointer(3)},
	}

	groups := planFieldDeviceUpdateGroups(updates, existing, proposedDevices(existing, updates))

	if len(groups) != 1 || len(groups[0].Indexes) != 2 {
		t.Fatalf("duplicate targets must share one group: %+v", groups)
	}
}

func TestSwapPlannerGroupsTransitiveNumberChain(t *testing.T) {
	scope, part, apparat := uuid.New(), uuid.New(), uuid.New()
	first, second, third := uuid.New(), uuid.New(), uuid.New()
	existing := map[uuid.UUID]*domainFacility.FieldDevice{
		first: deviceAt(first, scope, part, apparat, 1), second: deviceAt(second, scope, part, apparat, 2),
		third: deviceAt(third, scope, part, apparat, 3),
	}
	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: first, ApparatNr: intPointer(2)}, {ID: second, ApparatNr: intPointer(3)}, {ID: third, ApparatNr: intPointer(1)},
	}
	groups := planFieldDeviceUpdateGroups(updates, existing, proposedDevices(existing, updates))
	if len(groups) != 1 || len(groups[0].Indexes) != 3 {
		t.Fatalf("transitive cycle must be one group: %+v", groups)
	}
}

func TestSwapPlannerGroupsDuplicateCommandsForSameDevice(t *testing.T) {
	scope, part, apparat, id := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	existing := map[uuid.UUID]*domainFacility.FieldDevice{id: deviceAt(id, scope, part, apparat, 1)}
	updates := []domainFacility.BulkFieldDeviceUpdate{{ID: id, ApparatNr: intPointer(2)}, {ID: id, ApparatNr: intPointer(3)}}
	groups := planFieldDeviceUpdateGroups(updates, existing, proposedDevices(existing, updates))
	if len(groups) != 1 || len(groups[0].Indexes) != 2 {
		t.Fatalf("duplicate commands must be atomic: %+v", groups)
	}
}

func TestSwapPlannerConnectsScopeChangingCycle(t *testing.T) {
	scope, firstPart, secondPart, apparat := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	first, second := uuid.New(), uuid.New()
	existing := map[uuid.UUID]*domainFacility.FieldDevice{
		first:  deviceAt(first, scope, firstPart, apparat, 1),
		second: deviceAt(second, scope, secondPart, apparat, 1),
	}
	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: first, SystemPartID: &secondPart},
		{ID: second, SystemPartID: &firstPart},
	}
	groups := planFieldDeviceUpdateGroups(updates, existing, proposedDevices(existing, updates))
	if len(groups) != 1 || len(groups[0].Indexes) != 2 {
		t.Fatalf("scope-changing cycle must be atomic: %+v", groups)
	}
}

func TestFieldDeviceNumberConstraintMapsToWriteConflict(t *testing.T) {
	cause := &pgconn.PgError{Code: "23505", ConstraintName: "uq_field_devices_number_scope"}
	if err := mapFieldDeviceNumberConflict(cause); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("mapped error = %v", err)
	}
}

func proposedDevices(existing map[uuid.UUID]*domainFacility.FieldDevice, updates []domainFacility.BulkFieldDeviceUpdate) map[uuid.UUID]*domainFacility.FieldDevice {
	result := make(map[uuid.UUID]*domainFacility.FieldDevice, len(updates))
	for _, update := range updates {
		result[update.ID] = buildProposedFieldDevice(existing[update.ID], update)
	}
	return result
}

func deviceAt(id, systemTypeID, systemPartID, apparatID uuid.UUID, number int) *domainFacility.FieldDevice {
	return &domainFacility.FieldDevice{Base: domain.Base{ID: id, Version: 1}, SPSControllerSystemTypeID: systemTypeID, SystemPartID: systemPartID, ApparatID: apparatID, ApparatNr: number}
}

func intPointer(value int) *int { return &value }

func stringPointer(value string) *string { return &value }
