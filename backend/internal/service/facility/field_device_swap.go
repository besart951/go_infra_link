package facility

import (
	"slices"
	"strings"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type FieldDeviceNumberKey struct {
	SPSControllerSystemTypeID uuid.UUID
	SystemPartID              uuid.UUID
	ApparatID                 uuid.UUID
	ApparatNr                 int
}

type fieldDeviceUpdateGroup struct {
	ID      uuid.UUID
	Indexes []int
}

type FieldDeviceBulkUpdateGroup struct {
	ID      uuid.UUID                              `json:"dependency_group_id"`
	Indexes []int                                  `json:"indexes"`
	Updates []domainFacility.BulkFieldDeviceUpdate `json:"updates"`
}

type swapPlanner struct {
	updates  []domainFacility.BulkFieldDeviceUpdate
	existing map[uuid.UUID]*domainFacility.FieldDevice
	proposed map[uuid.UUID]*domainFacility.FieldDevice
	parents  []int
}

func planFieldDeviceUpdateGroups(
	updates []domainFacility.BulkFieldDeviceUpdate,
	existing map[uuid.UUID]*domainFacility.FieldDevice,
	proposed map[uuid.UUID]*domainFacility.FieldDevice,
) []fieldDeviceUpdateGroup {
	planner := newSwapPlanner(updates, existing, proposed)
	planner.connectDependencies()
	return planner.groups()
}

func newSwapPlanner(
	updates []domainFacility.BulkFieldDeviceUpdate,
	existing map[uuid.UUID]*domainFacility.FieldDevice,
	proposed map[uuid.UUID]*domainFacility.FieldDevice,
) *swapPlanner {
	parents := make([]int, len(updates))
	for index := range parents {
		parents[index] = index
	}
	return &swapPlanner{updates: updates, existing: existing, proposed: proposed, parents: parents}
}

func (p *swapPlanner) connectDependencies() {
	currentOwners := make(map[FieldDeviceNumberKey]int, len(p.updates))
	targetOwners := make(map[FieldDeviceNumberKey]int, len(p.updates))
	commandOwners := make(map[uuid.UUID]int, len(p.updates))
	for index, update := range p.updates {
		if owner, ok := commandOwners[update.ID]; ok {
			p.union(index, owner)
		}
		commandOwners[update.ID] = index
		if current := p.existing[update.ID]; current != nil {
			currentOwners[numberKey(current)] = index
		}
	}
	for index, update := range p.updates {
		p.connectTarget(index, update.ID, currentOwners, targetOwners)
	}
}

func (p *swapPlanner) connectTarget(index int, id uuid.UUID, currentOwners, targetOwners map[FieldDeviceNumberKey]int) {
	target := p.proposed[id]
	if target == nil {
		return
	}
	key := numberKey(target)
	if owner, ok := currentOwners[key]; ok && owner != index {
		p.union(index, owner)
	}
	if owner, ok := targetOwners[key]; ok {
		p.union(index, owner)
	}
	targetOwners[key] = index
}

func (p *swapPlanner) groups() []fieldDeviceUpdateGroup {
	byRoot := make(map[int][]int, len(p.updates))
	for index := range p.updates {
		root := p.find(index)
		byRoot[root] = append(byRoot[root], index)
	}
	groups := make([]fieldDeviceUpdateGroup, 0, len(byRoot))
	for _, indexes := range byRoot {
		groups = append(groups, fieldDeviceUpdateGroup{ID: dependencyGroupID(p.updates, indexes), Indexes: indexes})
	}
	slices.SortFunc(groups, func(a, b fieldDeviceUpdateGroup) int { return a.Indexes[0] - b.Indexes[0] })
	return groups
}

func (p *swapPlanner) find(index int) int {
	for p.parents[index] != index {
		p.parents[index] = p.parents[p.parents[index]]
		index = p.parents[index]
	}
	return index
}

func (p *swapPlanner) union(left, right int) {
	leftRoot, rightRoot := p.find(left), p.find(right)
	if leftRoot != rightRoot {
		p.parents[rightRoot] = leftRoot
	}
}

func numberKey(device *domainFacility.FieldDevice) FieldDeviceNumberKey {
	return FieldDeviceNumberKey{
		SPSControllerSystemTypeID: device.SPSControllerSystemTypeID,
		SystemPartID:              device.SystemPartID,
		ApparatID:                 device.ApparatID,
		ApparatNr:                 device.ApparatNr,
	}
}

func dependencyGroupID(updates []domainFacility.BulkFieldDeviceUpdate, indexes []int) uuid.UUID {
	ids := make([]string, len(indexes))
	for position, index := range indexes {
		ids[position] = updates[index].ID.String()
	}
	slices.Sort(ids)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(strings.Join(ids, ",")))
}
