package historysql

import (
	"context"
	"encoding/json"
	"sort"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type indexedMutation struct {
	index    int
	mutation Mutation
}

func (s *Store) resolveMutationScopes(
	ctx context.Context,
	mutations []Mutation,
) ([][]resolvedScope, error) {
	resolved := make([][]resolvedScope, len(mutations))
	fieldDeviceAssociated := make([]indexedMutation, 0)
	hierarchyAssociated := make([]indexedMutation, 0)
	for i, mutation := range mutations {
		if isFieldDeviceAssociatedTable(mutation.EntityTable) {
			fieldDeviceAssociated = append(
				fieldDeviceAssociated,
				indexedMutation{index: i, mutation: mutation},
			)
			continue
		}
		if isHierarchyAssociatedTable(mutation.EntityTable) {
			hierarchyAssociated = append(
				hierarchyAssociated,
				indexedMutation{index: i, mutation: mutation},
			)
			continue
		}
		scopes, err := s.resolveScopes(
			ctx,
			mutation.EntityTable,
			mutation.EntityID,
			mutationScopeSnapshot(mutation),
		)
		if err != nil {
			return nil, err
		}
		resolved[i] = scopes
	}

	if len(fieldDeviceAssociated) > 0 {
		batched, err := s.resolveFieldDeviceMutationScopesBatch(
			ctx,
			fieldDeviceAssociated,
		)
		if err != nil {
			return nil, err
		}
		for index, scopes := range batched {
			resolved[index] = scopes
		}
	}
	if len(hierarchyAssociated) > 0 {
		batched, err := s.resolveHierarchyMutationScopesBatch(
			ctx,
			hierarchyAssociated,
		)
		if err != nil {
			return nil, err
		}
		for index, scopes := range batched {
			resolved[index] = scopes
		}
	}
	return resolved, nil
}

func isFieldDeviceAssociatedTable(table string) bool {
	switch table {
	case "field_devices",
		"specifications",
		"bacnet_objects",
		"bacnet_object_alarm_values",
		"project_field_devices":
		return true
	default:
		return false
	}
}

func mutationScopeSnapshot(mutation Mutation) domainHistory.JSONB {
	if len(mutation.AfterJSON) > 0 {
		return mutation.AfterJSON
	}
	return mutation.BeforeJSON
}

type fieldDeviceAncestorRow struct {
	SystemTypeID     uuid.UUID `gorm:"column:system_type_id"`
	SPSControllerID  uuid.UUID `gorm:"column:sps_controller_id"`
	ControlCabinetID uuid.UUID `gorm:"column:control_cabinet_id"`
	BuildingID       uuid.UUID `gorm:"column:building_id"`
}

type fieldDeviceSystemTypeRow struct {
	FieldDeviceID uuid.UUID `gorm:"column:field_device_id"`
	SystemTypeID  uuid.UUID `gorm:"column:system_type_id"`
}

type fieldDeviceProjectRow struct {
	FieldDeviceID uuid.UUID `gorm:"column:field_device_id"`
	ProjectID     uuid.UUID `gorm:"column:project_id"`
}

type systemTypeProjectRow struct {
	SystemTypeID uuid.UUID `gorm:"column:system_type_id"`
	ProjectID    uuid.UUID `gorm:"column:project_id"`
}

type bacnetObjectFieldDeviceRow struct {
	BacnetObjectID uuid.UUID  `gorm:"column:bacnet_object_id"`
	FieldDeviceID  *uuid.UUID `gorm:"column:field_device_id"`
}

type bacnetObjectObjectDataRow struct {
	BacnetObjectID uuid.UUID  `gorm:"column:bacnet_object_id"`
	ObjectDataID   uuid.UUID  `gorm:"column:object_data_id"`
	ProjectID      *uuid.UUID `gorm:"column:project_id"`
}

func (s *Store) resolveFieldDeviceMutationScopesBatch(
	ctx context.Context,
	mutations []indexedMutation,
) (map[int][]resolvedScope, error) {
	scopeSets := make(map[int]scopeSet, len(mutations))
	fieldDevicesByOwner := make(map[int]map[uuid.UUID]struct{}, len(mutations))
	systemTypesByOwner := make(map[int]map[uuid.UUID]struct{}, len(mutations))
	bacnetOwners := make(map[uuid.UUID][]int)
	bacnetNeedingFieldDevice := make(map[uuid.UUID]struct{})

	for _, item := range mutations {
		scopes := scopeSet{}
		scopeSets[item.index] = scopes

		var snapshot map[string]any
		_ = json.Unmarshal(mutationScopeSnapshot(item.mutation), &snapshot)

		switch item.mutation.EntityTable {
		case "field_devices":
			scopes.add(scopeFieldDevice, item.mutation.EntityID)
			addOwnerUUID(fieldDevicesByOwner, item.index, item.mutation.EntityID)
			systemTypesByOwner[item.index] = mutationUUIDFieldSet(
				item.mutation,
				"sps_controller_system_type_id",
			)
		case "specifications":
			scopes.add(scopeSpecification, item.mutation.EntityID)
			fieldDeviceID := uuidField(snapshot, "field_device_id")
			scopes.add(scopeFieldDevice, fieldDeviceID)
			addOwnerUUID(fieldDevicesByOwner, item.index, fieldDeviceID)
		case "bacnet_objects":
			scopes.add(scopeBacnetObject, item.mutation.EntityID)
			fieldDeviceIDs := mutationUUIDFieldSet(item.mutation, "field_device_id")
			for fieldDeviceID := range fieldDeviceIDs {
				scopes.add(scopeFieldDevice, fieldDeviceID)
				addOwnerUUID(fieldDevicesByOwner, item.index, fieldDeviceID)
			}
			if len(fieldDeviceIDs) == 0 {
				bacnetNeedingFieldDevice[item.mutation.EntityID] = struct{}{}
			}
			bacnetOwners[item.mutation.EntityID] = append(
				bacnetOwners[item.mutation.EntityID],
				item.index,
			)
		case "bacnet_object_alarm_values":
			bacnetObjectID := uuidField(snapshot, "bacnet_object_id")
			scopes.add(scopeBacnetObject, bacnetObjectID)
			bacnetNeedingFieldDevice[bacnetObjectID] = struct{}{}
			bacnetOwners[bacnetObjectID] = append(
				bacnetOwners[bacnetObjectID],
				item.index,
			)
		case "project_field_devices":
			projectID := uuidField(snapshot, "project_id")
			fieldDeviceID := uuidField(snapshot, "field_device_id")
			scopes.add(scopeProject, projectID)
			scopes.add(scopeFieldDevice, fieldDeviceID)
			addOwnerUUID(fieldDevicesByOwner, item.index, fieldDeviceID)
		}
	}

	if err := s.addBacnetObjectScopesBatch(
		ctx,
		scopeSets,
		fieldDevicesByOwner,
		bacnetOwners,
		bacnetNeedingFieldDevice,
	); err != nil {
		return nil, err
	}
	if err := s.addFieldDeviceScopesBatch(
		ctx,
		scopeSets,
		fieldDevicesByOwner,
		systemTypesByOwner,
	); err != nil {
		return nil, err
	}

	out := make(map[int][]resolvedScope, len(scopeSets))
	for index, scopes := range scopeSets {
		out[index] = scopes.values()
	}
	return out, nil
}

func (s *Store) addBacnetObjectScopesBatch(
	ctx context.Context,
	scopeSets map[int]scopeSet,
	fieldDevicesByOwner map[int]map[uuid.UUID]struct{},
	bacnetOwners map[uuid.UUID][]int,
	bacnetNeedingFieldDevice map[uuid.UUID]struct{},
) error {
	for _, chunk := range uuidChunks(
		uuidSetValues(bacnetNeedingFieldDevice),
		historyWriteBatchSize,
	) {
		var rows []bacnetObjectFieldDeviceRow
		if err := s.db.WithContext(ctx).
			Table("bacnet_objects").
			Select("id AS bacnet_object_id, field_device_id").
			Where("id IN ?", chunk).
			Scan(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			if row.FieldDeviceID == nil {
				continue
			}
			for _, owner := range bacnetOwners[row.BacnetObjectID] {
				scopeSets[owner].add(scopeFieldDevice, *row.FieldDeviceID)
				addOwnerUUID(fieldDevicesByOwner, owner, *row.FieldDeviceID)
			}
		}
	}

	for _, chunk := range uuidChunks(
		uuidMapKeys(bacnetOwners),
		historyWriteBatchSize,
	) {
		var rows []bacnetObjectObjectDataRow
		if err := s.db.WithContext(ctx).
			Table("object_data_bacnet_objects AS odb").
			Select(`
				odb.bacnet_object_id,
				odb.object_data_id,
				od.project_id
			`).
			Joins("JOIN object_data AS od ON od.id = odb.object_data_id").
			Where("odb.bacnet_object_id IN ?", chunk).
			Scan(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			for _, owner := range bacnetOwners[row.BacnetObjectID] {
				scopeSets[owner].add(scopeObjectData, row.ObjectDataID)
				if row.ProjectID != nil {
					scopeSets[owner].add(scopeProject, *row.ProjectID)
				}
			}
		}
	}
	return nil
}

func (s *Store) addFieldDeviceScopesBatch(
	ctx context.Context,
	scopeSets map[int]scopeSet,
	fieldDevicesByOwner map[int]map[uuid.UUID]struct{},
	systemTypesByOwner map[int]map[uuid.UUID]struct{},
) error {
	ownersByFieldDevice := make(map[uuid.UUID][]int)
	missingSystemTypes := make(map[uuid.UUID]struct{})
	for owner, fieldDeviceIDs := range fieldDevicesByOwner {
		for fieldDeviceID := range fieldDeviceIDs {
			if fieldDeviceID == uuid.Nil {
				continue
			}
			ownersByFieldDevice[fieldDeviceID] = append(
				ownersByFieldDevice[fieldDeviceID],
				owner,
			)
			if len(systemTypesByOwner[owner]) == 0 {
				missingSystemTypes[fieldDeviceID] = struct{}{}
			}
		}
	}

	systemTypesByFieldDevice := make(map[uuid.UUID]uuid.UUID)
	for _, chunk := range uuidChunks(
		uuidSetValues(missingSystemTypes),
		historyWriteBatchSize,
	) {
		var rows []fieldDeviceSystemTypeRow
		if err := s.db.WithContext(ctx).
			Table("field_devices").
			Select(`
				id AS field_device_id,
				sps_controller_system_type_id AS system_type_id
			`).
			Where("id IN ?", chunk).
			Scan(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			systemTypesByFieldDevice[row.FieldDeviceID] = row.SystemTypeID
		}
	}

	ownersBySystemType := make(map[uuid.UUID][]int)
	for fieldDeviceID, owners := range ownersByFieldDevice {
		for _, owner := range owners {
			systemTypeIDs := systemTypesByOwner[owner]
			if len(systemTypeIDs) == 0 {
				systemTypeID := systemTypesByFieldDevice[fieldDeviceID]
				if systemTypeID != uuid.Nil {
					systemTypeIDs = map[uuid.UUID]struct{}{
						systemTypeID: {},
					}
				}
			}
			for systemTypeID := range systemTypeIDs {
				if systemTypeID == uuid.Nil {
					continue
				}
				scopeSets[owner].add(scopeSPSControllerSystemType, systemTypeID)
				ownersBySystemType[systemTypeID] = append(
					ownersBySystemType[systemTypeID],
					owner,
				)
			}
		}
	}

	for _, chunk := range uuidChunks(
		uuidMapKeys(ownersBySystemType),
		historyWriteBatchSize,
	) {
		var ancestors []fieldDeviceAncestorRow
		if err := s.db.WithContext(ctx).
			Table("sps_controller_system_types AS st").
			Select(`
				st.id AS system_type_id,
				st.sps_controller_id,
				s.control_cabinet_id,
				c.building_id
			`).
			Joins("LEFT JOIN sps_controllers AS s ON s.id = st.sps_controller_id").
			Joins("LEFT JOIN control_cabinets AS c ON c.id = s.control_cabinet_id").
			Where("st.id IN ?", chunk).
			Scan(&ancestors).Error; err != nil {
			return err
		}
		for _, ancestor := range ancestors {
			for _, owner := range ownersBySystemType[ancestor.SystemTypeID] {
				scopes := scopeSets[owner]
				scopes.add(scopeSPSController, ancestor.SPSControllerID)
				scopes.add(scopeControlCabinet, ancestor.ControlCabinetID)
				scopes.add(scopeBuilding, ancestor.BuildingID)
			}
		}
	}

	for _, chunk := range uuidChunks(
		uuidMapKeys(ownersByFieldDevice),
		historyWriteBatchSize,
	) {
		var projects []fieldDeviceProjectRow
		if err := s.db.WithContext(ctx).
			Table("project_field_devices").
			Select("field_device_id, project_id").
			Where("field_device_id IN ?", chunk).
			Scan(&projects).Error; err != nil {
			return err
		}
		for _, project := range projects {
			for _, owner := range ownersByFieldDevice[project.FieldDeviceID] {
				scopeSets[owner].add(scopeProject, project.ProjectID)
			}
		}
	}

	for _, chunk := range uuidChunks(
		uuidMapKeys(ownersBySystemType),
		historyWriteBatchSize,
	) {
		var projects []systemTypeProjectRow
		if err := s.db.WithContext(ctx).
			Raw(`
				SELECT st.id AS system_type_id, psc.project_id
				FROM sps_controller_system_types st
				JOIN project_sps_controllers psc
					ON psc.sps_controller_id = st.sps_controller_id
				WHERE st.id IN ?
				UNION
				SELECT st.id AS system_type_id, pcc.project_id
				FROM sps_controller_system_types st
				JOIN sps_controllers s ON s.id = st.sps_controller_id
				JOIN project_control_cabinets pcc
					ON pcc.control_cabinet_id = s.control_cabinet_id
				WHERE st.id IN ?
			`, chunk, chunk).
			Scan(&projects).Error; err != nil {
			return err
		}
		for _, project := range projects {
			for _, owner := range ownersBySystemType[project.SystemTypeID] {
				scopeSets[owner].add(scopeProject, project.ProjectID)
			}
		}
	}
	return nil
}

func addOwnerUUID(
	idsByOwner map[int]map[uuid.UUID]struct{},
	owner int,
	id uuid.UUID,
) {
	if id == uuid.Nil {
		return
	}
	if idsByOwner[owner] == nil {
		idsByOwner[owner] = make(map[uuid.UUID]struct{})
	}
	idsByOwner[owner][id] = struct{}{}
}

func mutationUUIDFieldSet(mutation Mutation, key string) map[uuid.UUID]struct{} {
	ids := make(map[uuid.UUID]struct{}, 2)
	for _, snapshot := range []domainHistory.JSONB{
		mutation.BeforeJSON,
		mutation.AfterJSON,
	} {
		if len(snapshot) == 0 {
			continue
		}
		var row map[string]any
		_ = json.Unmarshal(snapshot, &row)
		if id := uuidField(row, key); id != uuid.Nil {
			ids[id] = struct{}{}
		}
	}
	return ids
}

func uuidSetValues(set map[uuid.UUID]struct{}) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(set))
	for id := range set {
		if id != uuid.Nil {
			ids = append(ids, id)
		}
	}
	sortUUIDs(ids)
	return ids
}

func uuidMapKeys[T any](items map[uuid.UUID]T) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for id := range items {
		if id != uuid.Nil {
			ids = append(ids, id)
		}
	}
	sortUUIDs(ids)
	return ids
}

func sortUUIDs(ids []uuid.UUID) {
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].String() < ids[j].String()
	})
}
