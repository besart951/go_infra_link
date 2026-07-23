package historysql

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

func isHierarchyAssociatedTable(table string) bool {
	switch table {
	case "control_cabinets",
		"sps_controllers",
		"sps_controller_system_types",
		"project_control_cabinets",
		"project_sps_controllers":
		return true
	default:
		return false
	}
}

type hierarchySPSAncestorRow struct {
	SPSControllerID  uuid.UUID `gorm:"column:sps_controller_id"`
	ControlCabinetID uuid.UUID `gorm:"column:control_cabinet_id"`
	BuildingID       uuid.UUID `gorm:"column:building_id"`
}

type hierarchyCabinetAncestorRow struct {
	ControlCabinetID uuid.UUID `gorm:"column:control_cabinet_id"`
	BuildingID       uuid.UUID `gorm:"column:building_id"`
}

type hierarchyCabinetProjectRow struct {
	ControlCabinetID uuid.UUID `gorm:"column:control_cabinet_id"`
	ProjectID        uuid.UUID `gorm:"column:project_id"`
}

type hierarchySPSProjectRow struct {
	SPSControllerID uuid.UUID `gorm:"column:sps_controller_id"`
	ProjectID       uuid.UUID `gorm:"column:project_id"`
}

func (s *Store) resolveHierarchyMutationScopesBatch(
	ctx context.Context,
	mutations []indexedMutation,
) (map[int][]resolvedScope, error) {
	scopeSets := make(map[int]scopeSet, len(mutations))
	cabinetByOwner := make(map[int]uuid.UUID, len(mutations))
	buildingByOwner := make(map[int]uuid.UUID, len(mutations))
	spsByOwner := make(map[int]uuid.UUID, len(mutations))
	cabinetProjectOwners := make(map[uuid.UUID][]int)
	spsProjectOwners := make(map[uuid.UUID][]int)
	movedSPSCabinetsByOwner := make(map[int]map[uuid.UUID]struct{})

	for _, item := range mutations {
		scopes := scopeSet{}
		scopeSets[item.index] = scopes

		var snapshot map[string]any
		_ = json.Unmarshal(mutationScopeSnapshot(item.mutation), &snapshot)

		switch item.mutation.EntityTable {
		case "control_cabinets":
			controlCabinetID := item.mutation.EntityID
			buildingID := uuidField(snapshot, "building_id")
			buildingIDs := mutationUUIDFieldSet(item.mutation, "building_id")
			scopes.add(scopeControlCabinet, controlCabinetID)
			scopes.add(scopeBuilding, buildingID)
			for scopedBuildingID := range buildingIDs {
				scopes.add(scopeBuilding, scopedBuildingID)
			}
			cabinetByOwner[item.index] = controlCabinetID
			buildingByOwner[item.index] = buildingID
			cabinetProjectOwners[controlCabinetID] = append(
				cabinetProjectOwners[controlCabinetID],
				item.index,
			)
		case "sps_controllers":
			spsControllerID := item.mutation.EntityID
			controlCabinetID := uuidField(snapshot, "control_cabinet_id")
			moveCabinetIDs := mutationUUIDFieldSet(
				item.mutation,
				"control_cabinet_id",
			)
			scopes.add(scopeSPSController, spsControllerID)
			scopes.add(scopeControlCabinet, controlCabinetID)
			for cabinetID := range moveCabinetIDs {
				scopes.add(scopeControlCabinet, cabinetID)
			}
			if len(moveCabinetIDs) > 1 {
				movedSPSCabinetsByOwner[item.index] = moveCabinetIDs
			}
			spsByOwner[item.index] = spsControllerID
			cabinetByOwner[item.index] = controlCabinetID
			spsProjectOwners[spsControllerID] = append(
				spsProjectOwners[spsControllerID],
				item.index,
			)
		case "sps_controller_system_types":
			spsControllerID := uuidField(snapshot, "sps_controller_id")
			scopes.add(scopeSPSControllerSystemType, item.mutation.EntityID)
			scopes.add(scopeSPSController, spsControllerID)
			spsByOwner[item.index] = spsControllerID
			spsProjectOwners[spsControllerID] = append(
				spsProjectOwners[spsControllerID],
				item.index,
			)
		case "project_control_cabinets":
			projectID := uuidField(snapshot, "project_id")
			controlCabinetID := uuidField(snapshot, "control_cabinet_id")
			scopes.add(scopeProject, projectID)
			scopes.add(scopeControlCabinet, controlCabinetID)
			cabinetByOwner[item.index] = controlCabinetID
		case "project_sps_controllers":
			projectID := uuidField(snapshot, "project_id")
			spsControllerID := uuidField(snapshot, "sps_controller_id")
			scopes.add(scopeProject, projectID)
			scopes.add(scopeSPSController, spsControllerID)
			spsByOwner[item.index] = spsControllerID
		}
	}

	if err := s.addHierarchyAncestorScopesBatch(
		ctx,
		scopeSets,
		cabinetByOwner,
		buildingByOwner,
		spsByOwner,
	); err != nil {
		return nil, err
	}
	if err := s.addHierarchyProjectScopesBatch(
		ctx,
		scopeSets,
		cabinetProjectOwners,
		spsProjectOwners,
	); err != nil {
		return nil, err
	}
	if err := s.addSPSMoveScopesBatch(
		ctx,
		scopeSets,
		movedSPSCabinetsByOwner,
	); err != nil {
		return nil, err
	}

	out := make(map[int][]resolvedScope, len(scopeSets))
	for index, scopes := range scopeSets {
		out[index] = scopes.values()
	}
	return out, nil
}

func (s *Store) addSPSMoveScopesBatch(
	ctx context.Context,
	scopeSets map[int]scopeSet,
	cabinetIDsByOwner map[int]map[uuid.UUID]struct{},
) error {
	ownersByCabinet := make(map[uuid.UUID][]int)
	for owner, cabinetIDs := range cabinetIDsByOwner {
		for cabinetID := range cabinetIDs {
			if cabinetID == uuid.Nil {
				continue
			}
			ownersByCabinet[cabinetID] = append(ownersByCabinet[cabinetID], owner)
		}
	}

	for _, chunk := range uuidChunks(
		uuidMapKeys(ownersByCabinet),
		historyWriteBatchSize,
	) {
		var ancestors []hierarchyCabinetAncestorRow
		if err := s.db.WithContext(ctx).
			Table("control_cabinets").
			Select("id AS control_cabinet_id, building_id").
			Where("id IN ?", chunk).
			Scan(&ancestors).Error; err != nil {
			return err
		}
		for _, ancestor := range ancestors {
			for _, owner := range ownersByCabinet[ancestor.ControlCabinetID] {
				scopeSets[owner].add(scopeBuilding, ancestor.BuildingID)
			}
		}

		// Only direct cabinet links are added here. Descendant SPS/FieldDevice
		// links remain attached to the moved controller and are resolved by the
		// normal SPS query. Including every descendant of the old cabinet would
		// incorrectly widen this event to unrelated projects.
		var projects []hierarchyCabinetProjectRow
		if err := s.db.WithContext(ctx).
			Table("project_control_cabinets").
			Select("control_cabinet_id, project_id").
			Where("control_cabinet_id IN ?", chunk).
			Scan(&projects).Error; err != nil {
			return err
		}
		for _, project := range projects {
			for _, owner := range ownersByCabinet[project.ControlCabinetID] {
				scopeSets[owner].add(scopeProject, project.ProjectID)
			}
		}
	}
	return nil
}

func (s *Store) addHierarchyAncestorScopesBatch(
	ctx context.Context,
	scopeSets map[int]scopeSet,
	cabinetByOwner map[int]uuid.UUID,
	buildingByOwner map[int]uuid.UUID,
	spsByOwner map[int]uuid.UUID,
) error {
	ownersBySPS := reverseOwnerIDs(spsByOwner)
	for _, chunk := range uuidChunks(
		uuidMapKeys(ownersBySPS),
		historyWriteBatchSize,
	) {
		var ancestors []hierarchySPSAncestorRow
		if err := s.db.WithContext(ctx).
			Table("sps_controllers AS s").
			Select(`
				s.id AS sps_controller_id,
				s.control_cabinet_id,
				c.building_id
			`).
			Joins("LEFT JOIN control_cabinets AS c ON c.id = s.control_cabinet_id").
			Where("s.id IN ?", chunk).
			Scan(&ancestors).Error; err != nil {
			return err
		}
		for _, ancestor := range ancestors {
			for _, owner := range ownersBySPS[ancestor.SPSControllerID] {
				scopeSets[owner].add(
					scopeControlCabinet,
					ancestor.ControlCabinetID,
				)
				scopeSets[owner].add(scopeBuilding, ancestor.BuildingID)
				cabinetByOwner[owner] = ancestor.ControlCabinetID
				buildingByOwner[owner] = ancestor.BuildingID
			}
		}
	}

	cabinetOwnersNeedingBuilding := make(map[uuid.UUID][]int)
	for owner, controlCabinetID := range cabinetByOwner {
		if controlCabinetID == uuid.Nil || buildingByOwner[owner] != uuid.Nil {
			continue
		}
		cabinetOwnersNeedingBuilding[controlCabinetID] = append(
			cabinetOwnersNeedingBuilding[controlCabinetID],
			owner,
		)
	}
	for _, chunk := range uuidChunks(
		uuidMapKeys(cabinetOwnersNeedingBuilding),
		historyWriteBatchSize,
	) {
		var ancestors []hierarchyCabinetAncestorRow
		if err := s.db.WithContext(ctx).
			Table("control_cabinets").
			Select("id AS control_cabinet_id, building_id").
			Where("id IN ?", chunk).
			Scan(&ancestors).Error; err != nil {
			return err
		}
		for _, ancestor := range ancestors {
			for _, owner := range cabinetOwnersNeedingBuilding[ancestor.ControlCabinetID] {
				scopeSets[owner].add(scopeBuilding, ancestor.BuildingID)
				buildingByOwner[owner] = ancestor.BuildingID
			}
		}
	}
	return nil
}

func (s *Store) addHierarchyProjectScopesBatch(
	ctx context.Context,
	scopeSets map[int]scopeSet,
	cabinetProjectOwners map[uuid.UUID][]int,
	spsProjectOwners map[uuid.UUID][]int,
) error {
	for _, chunk := range uuidChunks(
		uuidMapKeys(cabinetProjectOwners),
		historyWriteBatchSize,
	) {
		var projects []hierarchyCabinetProjectRow
		if err := s.db.WithContext(ctx).
			Raw(`
				SELECT pcc.control_cabinet_id, pcc.project_id
				FROM project_control_cabinets pcc
				WHERE pcc.control_cabinet_id IN ?
				UNION
				SELECT s.control_cabinet_id, psc.project_id
				FROM project_sps_controllers psc
				JOIN sps_controllers s ON s.id = psc.sps_controller_id
				WHERE s.control_cabinet_id IN ?
				UNION
				SELECT s.control_cabinet_id, pfd.project_id
				FROM project_field_devices pfd
				JOIN field_devices fd ON fd.id = pfd.field_device_id
				JOIN sps_controller_system_types st
					ON st.id = fd.sps_controller_system_type_id
				JOIN sps_controllers s ON s.id = st.sps_controller_id
				WHERE s.control_cabinet_id IN ?
			`, chunk, chunk, chunk).
			Scan(&projects).Error; err != nil {
			return err
		}
		for _, project := range projects {
			for _, owner := range cabinetProjectOwners[project.ControlCabinetID] {
				scopeSets[owner].add(scopeProject, project.ProjectID)
			}
		}
	}

	for _, chunk := range uuidChunks(
		uuidMapKeys(spsProjectOwners),
		historyWriteBatchSize,
	) {
		var projects []hierarchySPSProjectRow
		if err := s.db.WithContext(ctx).
			Raw(`
				SELECT psc.sps_controller_id, psc.project_id
				FROM project_sps_controllers psc
				WHERE psc.sps_controller_id IN ?
				UNION
				SELECT s.id AS sps_controller_id, pcc.project_id
				FROM project_control_cabinets pcc
				JOIN sps_controllers s
					ON s.control_cabinet_id = pcc.control_cabinet_id
				WHERE s.id IN ?
				UNION
				SELECT st.sps_controller_id, pfd.project_id
				FROM project_field_devices pfd
				JOIN field_devices fd ON fd.id = pfd.field_device_id
				JOIN sps_controller_system_types st
					ON st.id = fd.sps_controller_system_type_id
				WHERE st.sps_controller_id IN ?
			`, chunk, chunk, chunk).
			Scan(&projects).Error; err != nil {
			return err
		}
		for _, project := range projects {
			for _, owner := range spsProjectOwners[project.SPSControllerID] {
				scopeSets[owner].add(scopeProject, project.ProjectID)
			}
		}
	}
	return nil
}

func reverseOwnerIDs(idsByOwner map[int]uuid.UUID) map[uuid.UUID][]int {
	ownersByID := make(map[uuid.UUID][]int)
	for owner, id := range idsByOwner {
		if id == uuid.Nil {
			continue
		}
		ownersByID[id] = append(ownersByID[id], owner)
	}
	return ownersByID
}
