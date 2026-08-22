package facilitysql

import (
	"context"
	"fmt"
	"sort"

	hierarchydelete "github.com/besart951/go_infra_link/backend/internal/application/hierarchydelete"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/besart951/go_infra_link/backend/internal/repository/historysql"
	"github.com/besart951/go_infra_link/backend/internal/service/auditctx"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const hierarchyDeleteDefaultLimit = 500

type HierarchyDeleteStore struct {
	db      *gorm.DB
	history *historysql.Store
}

type hierarchyDeleteRows struct {
	table   string
	ids     []uuid.UUID
	command hierarchydelete.Command
}

type hierarchyDeleteStrategy func(context.Context, *HierarchyDeleteStore, hierarchydelete.Command) (hierarchydelete.Result, error)

var hierarchyDeleteStrategies = map[hierarchydelete.Stage]hierarchyDeleteStrategy{
	hierarchydelete.StageFieldDevices: deleteFieldDeviceChunk,
	hierarchydelete.StageSystemTypes:  deleteSystemTypeChunk,
	hierarchydelete.StageControllers:  deleteControllerChunk,
	hierarchydelete.StageRoot:         deleteHierarchyRoot,
}

func NewHierarchyDeleteStore(db *gorm.DB) *HierarchyDeleteStore {
	return &HierarchyDeleteStore{db: db, history: historysql.NewStore(db)}
}

func (s *HierarchyDeleteStore) DeleteChunk(ctx context.Context, command hierarchydelete.Command) (hierarchydelete.Result, error) {
	strategy := hierarchyDeleteStrategies[command.Stage]
	if s == nil || s.db == nil || strategy == nil || command.RootID == uuid.Nil {
		return hierarchydelete.Result{}, fmt.Errorf("invalid hierarchy delete command")
	}
	if fieldDeviceRootScopes[command.RootKind] == nil {
		return hierarchydelete.Result{}, fmt.Errorf("unsupported hierarchy root %q", command.RootKind)
	}
	if command.Limit <= 0 || command.Limit > hierarchyDeleteDefaultLimit {
		command.Limit = hierarchyDeleteDefaultLimit
	}
	return strategy(ctx, s, command)
}

func deleteFieldDeviceChunk(ctx context.Context, store *HierarchyDeleteStore, command hierarchydelete.Command) (hierarchydelete.Result, error) {
	ids, err := store.fieldDeviceIDs(command)
	if err != nil || len(ids) == 0 {
		return hierarchydelete.Result{Done: len(ids) == 0}, err
	}
	ownedRows, err := store.fieldDeviceOwnedRows(command, ids)
	if err != nil {
		return hierarchydelete.Result{}, err
	}
	for _, rows := range ownedRows {
		if err := store.deleteRows(ctx, rows); err != nil {
			return hierarchydelete.Result{}, err
		}
	}
	return hierarchydelete.Result{Deleted: len(ids), Done: len(ids) < command.Limit}, nil
}

func (s *HierarchyDeleteStore) fieldDeviceOwnedRows(command hierarchydelete.Command, ids []uuid.UUID) ([]hierarchyDeleteRows, error) {
	alarmIDs, err := s.alarmValueIDs(ids)
	if err != nil {
		return nil, err
	}
	bacnetIDs, err := s.bacnetObjectIDs(ids)
	if err != nil {
		return nil, err
	}
	specificationIDs, err := s.specificationIDs(ids)
	if err != nil {
		return nil, err
	}
	linkIDs, err := s.projectLinkIDs("project_field_devices", "field_device_id", ids)
	if err != nil {
		return nil, err
	}
	return []hierarchyDeleteRows{
		{table: "bacnet_object_alarm_values", ids: alarmIDs, command: command},
		{table: "bacnet_objects", ids: bacnetIDs, command: command},
		{table: "specifications", ids: specificationIDs, command: command},
		{table: "project_field_devices", ids: linkIDs, command: command},
		{table: "field_devices", ids: ids, command: command},
	}, nil
}

func deleteSystemTypeChunk(ctx context.Context, store *HierarchyDeleteStore, command hierarchydelete.Command) (hierarchydelete.Result, error) {
	ids, err := store.systemTypeIDs(command)
	if err != nil || len(ids) == 0 {
		return hierarchydelete.Result{Done: len(ids) == 0}, err
	}
	rows := hierarchyDeleteRows{table: "sps_controller_system_types", ids: ids, command: command}
	if err := store.deleteRows(ctx, rows); err != nil {
		return hierarchydelete.Result{}, err
	}
	return hierarchydelete.Result{Deleted: len(ids), Done: len(ids) < command.Limit}, nil
}

func deleteControllerChunk(ctx context.Context, store *HierarchyDeleteStore, command hierarchydelete.Command) (hierarchydelete.Result, error) {
	ids, err := store.controllerIDs(command)
	if err != nil || len(ids) == 0 {
		return hierarchydelete.Result{Done: len(ids) == 0}, err
	}
	links, err := store.projectLinkIDs("project_sps_controllers", "sps_controller_id", ids)
	if err != nil {
		return hierarchydelete.Result{}, err
	}
	if err := store.deleteRows(ctx, hierarchyDeleteRows{table: "project_sps_controllers", ids: links, command: command}); err != nil {
		return hierarchydelete.Result{}, err
	}
	if err := store.deleteRows(ctx, hierarchyDeleteRows{table: "sps_controllers", ids: ids, command: command}); err != nil {
		return hierarchydelete.Result{}, err
	}
	return hierarchydelete.Result{Deleted: len(ids), Done: len(ids) < command.Limit}, nil
}

var hierarchyRootDeleteStrategies = map[hierarchydelete.RootKind]hierarchyDeleteStrategy{
	hierarchydelete.RootControlCabinet:          deleteCabinetRoot,
	hierarchydelete.RootSPSController:           deleteControllerRoot,
	hierarchydelete.RootSPSControllerSystemType: deleteSystemTypeRoot,
}

func deleteHierarchyRoot(ctx context.Context, store *HierarchyDeleteStore, command hierarchydelete.Command) (hierarchydelete.Result, error) {
	strategy := hierarchyRootDeleteStrategies[command.RootKind]
	if strategy == nil {
		return hierarchydelete.Result{}, fmt.Errorf("unsupported hierarchy root %q", command.RootKind)
	}
	return strategy(ctx, store, command)
}

func deleteCabinetRoot(ctx context.Context, store *HierarchyDeleteStore, command hierarchydelete.Command) (hierarchydelete.Result, error) {
	links, err := store.projectLinkIDs("project_control_cabinets", "control_cabinet_id", []uuid.UUID{command.RootID})
	if err != nil {
		return hierarchydelete.Result{}, err
	}
	if err := store.deleteRows(ctx, hierarchyDeleteRows{table: "project_control_cabinets", ids: links, command: command}); err != nil {
		return hierarchydelete.Result{}, err
	}
	return store.deleteFinalRoot(ctx, command, "control_cabinets")
}

func deleteControllerRoot(ctx context.Context, store *HierarchyDeleteStore, command hierarchydelete.Command) (hierarchydelete.Result, error) {
	links, err := store.projectLinkIDs("project_sps_controllers", "sps_controller_id", []uuid.UUID{command.RootID})
	if err != nil {
		return hierarchydelete.Result{}, err
	}
	if err := store.deleteRows(ctx, hierarchyDeleteRows{table: "project_sps_controllers", ids: links, command: command}); err != nil {
		return hierarchydelete.Result{}, err
	}
	return store.deleteFinalRoot(ctx, command, "sps_controllers")
}

func deleteSystemTypeRoot(ctx context.Context, store *HierarchyDeleteStore, command hierarchydelete.Command) (hierarchydelete.Result, error) {
	return store.deleteFinalRoot(ctx, command, "sps_controller_system_types")
}

func (s *HierarchyDeleteStore) deleteFinalRoot(ctx context.Context, command hierarchydelete.Command, table string) (hierarchydelete.Result, error) {
	rows := hierarchyDeleteRows{table: table, ids: []uuid.UUID{command.RootID}, command: command}
	if err := s.deleteRows(ctx, rows); err != nil {
		return hierarchydelete.Result{}, err
	}
	result := s.db.WithContext(ctx).Exec(
		"DELETE FROM facility_aggregate_lifecycle WHERE kind = ? AND resource_id = ?",
		command.RootKind, command.RootID,
	)
	return hierarchydelete.Result{Deleted: 1, Done: true}, result.Error
}

func (s *HierarchyDeleteStore) deleteRows(ctx context.Context, request hierarchyDeleteRows) error {
	if len(request.ids) == 0 {
		return nil
	}
	snapshots, err := s.history.LoadRows(ctx, request.table, request.ids)
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Exec("DELETE FROM "+request.table+" WHERE id IN ?", request.ids).Error; err != nil {
		return err
	}
	return s.recordDeletes(ctx, request, snapshots)
}

func (s *HierarchyDeleteStore) recordDeletes(ctx context.Context, request hierarchyDeleteRows, snapshots map[uuid.UUID]domainHistory.JSONB) error {
	ids := append([]uuid.UUID(nil), request.ids...)
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
	ctx = auditctx.WithActorID(ctx, request.command.ActorID)
	for _, id := range ids {
		if err := s.history.RecordMutation(ctx, historysql.Mutation{
			Action: domainHistory.ActionDelete, EntityTable: request.table, EntityID: id,
			BeforeJSON: snapshots[id], BatchID: &request.command.BatchID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *HierarchyDeleteStore) fieldDeviceIDs(command hierarchydelete.Command) ([]uuid.UUID, error) {
	query := fieldDeviceRootScopes[command.RootKind](s.db, command.RootID)
	var ids []uuid.UUID
	err := query.Order("field_devices.id ASC").Limit(command.Limit).Pluck("field_devices.id", &ids).Error
	return ids, err
}

func (s *HierarchyDeleteStore) systemTypeIDs(command hierarchydelete.Command) ([]uuid.UUID, error) {
	query := systemTypeRootScopes[command.RootKind](s.db, command.RootID)
	var ids []uuid.UUID
	err := query.Order("sps_controller_system_types.id ASC").Limit(command.Limit).Pluck("sps_controller_system_types.id", &ids).Error
	return ids, err
}

func (s *HierarchyDeleteStore) controllerIDs(command hierarchydelete.Command) ([]uuid.UUID, error) {
	query := controllerRootScopes[command.RootKind](s.db, command.RootID)
	var ids []uuid.UUID
	err := query.Order("sps_controllers.id ASC").Limit(command.Limit).Pluck("sps_controllers.id", &ids).Error
	return ids, err
}

type rootScope func(*gorm.DB, uuid.UUID) *gorm.DB

var fieldDeviceRootScopes = map[hierarchydelete.RootKind]rootScope{
	hierarchydelete.RootControlCabinet: func(db *gorm.DB, id uuid.UUID) *gorm.DB {
		return db.Table("field_devices").Joins("JOIN sps_controller_system_types st ON st.id = field_devices.sps_controller_system_type_id").
			Joins("JOIN sps_controllers sc ON sc.id = st.sps_controller_id").Where("sc.control_cabinet_id = ?", id)
	},
	hierarchydelete.RootSPSController: func(db *gorm.DB, id uuid.UUID) *gorm.DB {
		return db.Table("field_devices").Joins("JOIN sps_controller_system_types st ON st.id = field_devices.sps_controller_system_type_id").
			Where("st.sps_controller_id = ?", id)
	},
	hierarchydelete.RootSPSControllerSystemType: func(db *gorm.DB, id uuid.UUID) *gorm.DB {
		return db.Table("field_devices").Where("sps_controller_system_type_id = ?", id)
	},
}

var systemTypeRootScopes = map[hierarchydelete.RootKind]rootScope{
	hierarchydelete.RootControlCabinet: func(db *gorm.DB, id uuid.UUID) *gorm.DB {
		return db.Table("sps_controller_system_types").Joins("JOIN sps_controllers sc ON sc.id = sps_controller_system_types.sps_controller_id").
			Where("sc.control_cabinet_id = ?", id)
	},
	hierarchydelete.RootSPSController: func(db *gorm.DB, id uuid.UUID) *gorm.DB {
		return db.Table("sps_controller_system_types").Where("sps_controller_id = ?", id)
	},
}

var controllerRootScopes = map[hierarchydelete.RootKind]rootScope{
	hierarchydelete.RootControlCabinet: func(db *gorm.DB, id uuid.UUID) *gorm.DB {
		return db.Table("sps_controllers").Where("control_cabinet_id = ?", id)
	},
}

func (s *HierarchyDeleteStore) alarmValueIDs(fieldDeviceIDs []uuid.UUID) ([]uuid.UUID, error) {
	return s.relatedIDs("bacnet_object_alarm_values", "bacnet_object_alarm_values.id", func(db *gorm.DB) *gorm.DB {
		return db.Joins("JOIN bacnet_objects bo ON bo.id = bacnet_object_alarm_values.bacnet_object_id").Where("bo.field_device_id IN ?", fieldDeviceIDs)
	})
}

func (s *HierarchyDeleteStore) bacnetObjectIDs(fieldDeviceIDs []uuid.UUID) ([]uuid.UUID, error) {
	return s.relatedIDs("bacnet_objects", "bacnet_objects.id", func(db *gorm.DB) *gorm.DB {
		return db.Where("field_device_id IN ?", fieldDeviceIDs)
	})
}

func (s *HierarchyDeleteStore) specificationIDs(fieldDeviceIDs []uuid.UUID) ([]uuid.UUID, error) {
	return s.relatedIDs("specifications", "specifications.id", func(db *gorm.DB) *gorm.DB {
		return db.Where("field_device_id IN ?", fieldDeviceIDs)
	})
}

func (s *HierarchyDeleteStore) projectLinkIDs(table, ownerColumn string, ownerIDs []uuid.UUID) ([]uuid.UUID, error) {
	return s.relatedIDs(table, table+".id", func(db *gorm.DB) *gorm.DB {
		return db.Where(ownerColumn+" IN ?", ownerIDs)
	})
}

func (s *HierarchyDeleteStore) relatedIDs(table, column string, scope func(*gorm.DB) *gorm.DB) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := scope(s.db.Table(table)).Order(column+" ASC").Pluck(column, &ids).Error
	return ids, err
}
