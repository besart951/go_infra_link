package historysql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var hierarchyRestoreTables = []string{
	"control_cabinets",
	"sps_controllers",
	"sps_controller_system_types",
	"field_devices",
	"specifications",
	"bacnet_objects",
	"bacnet_object_alarm_values",
	"project_control_cabinets",
	"project_sps_controllers",
	"project_field_devices",
}

var hierarchyDeleteTables = []string{
	"project_field_devices",
	"project_sps_controllers",
	"project_control_cabinets",
	"bacnet_object_alarm_values",
	"bacnet_objects",
	"specifications",
	"field_devices",
	"sps_controller_system_types",
	"sps_controllers",
	"control_cabinets",
}

func (s *Store) RestoreEntityToEvent(ctx context.Context, eventID uuid.UUID, mode domainHistory.RestoreMode) (*domainHistory.RestoreResult, error) {
	event, err := s.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	target := event.AfterJSON
	if mode == domainHistory.RestoreModeBefore {
		target = event.BeforeJSON
	}
	batchID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	var result domainHistory.RestoreResult
	result.BatchID = batchID

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txStore := s.WithDB(tx)
		before, _, err := txStore.LoadRow(ctx, event.EntityTable, event.EntityID)
		if err != nil {
			return err
		}

		if len(target) == 0 {
			if err := deleteRow(ctx, tx, event.EntityTable, event.EntityID); err != nil {
				return err
			}
			result.DeletedCount = 1
		} else {
			if err := upsertRow(ctx, tx, event.EntityTable, target); err != nil {
				return err
			}
			result.RestoredCount = 1
		}

		return txStore.RecordMutation(ctx, Mutation{
			Action:      domainHistory.ActionRestore,
			EntityTable: event.EntityTable,
			EntityID:    event.EntityID,
			BeforeJSON:  before,
			AfterJSON:   target,
			BatchID:     &batchID,
			Summary:     "entity restored from history",
			Metadata: map[string]any{
				"source_event_id": eventID.String(),
				"restore_mode":    string(mode),
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Store) RestoreControlCabinet(ctx context.Context, controlCabinetID uuid.UUID, req domainHistory.RestoreControlCabinetRequest) (*domainHistory.RestoreResult, error) {
	asOf := time.Now().UTC()
	if req.AsOf != nil {
		asOf = req.AsOf.UTC()
	}
	if req.EventID != nil && *req.EventID != uuid.Nil {
		event, err := s.GetEvent(ctx, *req.EventID)
		if err != nil {
			return nil, err
		}
		asOf = event.OccurredAt
	}

	batchID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	result := &domainHistory.RestoreResult{BatchID: batchID}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txStore := s.WithDB(tx)
		targets, err := txStore.collectControlCabinetRestoreTargets(ctx, controlCabinetID, req.ProjectID)
		if err != nil {
			return err
		}

		versions, err := txStore.latestVersionsForTargets(ctx, targets, asOf)
		if err != nil {
			return err
		}

		for _, table := range hierarchyDeleteTables {
			for id := range targets[table] {
				version, ok := versions[targetKey(table, id)]
				if ok && len(version.SnapshotJSON) > 0 {
					continue
				}
				before, exists, err := txStore.LoadRow(ctx, table, id)
				if err != nil {
					return err
				}
				if !exists {
					result.SkippedCount++
					continue
				}
				if err := deleteRow(ctx, tx, table, id); err != nil {
					return err
				}
				result.DeletedCount++
				if err := txStore.RecordMutation(ctx, Mutation{
					Action:      domainHistory.ActionRestore,
					EntityTable: table,
					EntityID:    id,
					BeforeJSON:  before,
					BatchID:     &batchID,
					Summary:     "hierarchy restored from history",
					Metadata: map[string]any{
						"control_cabinet_id": controlCabinetID.String(),
						"restore_as_of":      asOf.Format(time.RFC3339Nano),
						"restore_effect":     "delete",
					},
				}); err != nil {
					return err
				}
			}
		}

		for _, table := range hierarchyRestoreTables {
			ids := sortedUUIDs(targets[table])
			for _, id := range ids {
				version, ok := versions[targetKey(table, id)]
				if !ok || len(version.SnapshotJSON) == 0 {
					continue
				}
				before, _, err := txStore.LoadRow(ctx, table, id)
				if err != nil {
					return err
				}
				if err := upsertRow(ctx, tx, table, version.SnapshotJSON); err != nil {
					return err
				}
				result.RestoredCount++
				if err := txStore.RecordMutation(ctx, Mutation{
					Action:      domainHistory.ActionRestore,
					EntityTable: table,
					EntityID:    id,
					BeforeJSON:  before,
					AfterJSON:   version.SnapshotJSON,
					BatchID:     &batchID,
					Summary:     "hierarchy restored from history",
					Metadata: map[string]any{
						"control_cabinet_id": controlCabinetID.String(),
						"restore_as_of":      asOf.Format(time.RFC3339Nano),
						"restore_effect":     "upsert",
					},
				}); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) collectControlCabinetRestoreTargets(ctx context.Context, controlCabinetID uuid.UUID, projectID *uuid.UUID) (map[string]map[uuid.UUID]struct{}, error) {
	targets := map[string]map[uuid.UUID]struct{}{}
	for _, table := range hierarchyRestoreTables {
		targets[table] = map[uuid.UUID]struct{}{}
	}
	add := func(table string, id uuid.UUID) {
		if id == uuid.Nil {
			return
		}
		if targets[table] == nil {
			targets[table] = map[uuid.UUID]struct{}{}
		}
		targets[table][id] = struct{}{}
	}

	add("control_cabinets", controlCabinetID)

	var rows []struct {
		TableName string    `gorm:"column:table_name"`
		ID        uuid.UUID `gorm:"column:id"`
	}
	query := `
		SELECT 'sps_controllers' AS table_name, s.id
		FROM sps_controllers s
		WHERE s.control_cabinet_id = @control_cabinet_id
		UNION
		SELECT 'sps_controller_system_types', st.id
		FROM sps_controller_system_types st
		JOIN sps_controllers s ON s.id = st.sps_controller_id
		WHERE s.control_cabinet_id = @control_cabinet_id
		UNION
		SELECT 'field_devices', fd.id
		FROM field_devices fd
		JOIN sps_controller_system_types st ON st.id = fd.sps_controller_system_type_id
		JOIN sps_controllers s ON s.id = st.sps_controller_id
		WHERE s.control_cabinet_id = @control_cabinet_id
		UNION
		SELECT 'specifications', sp.id
		FROM specifications sp
		JOIN field_devices fd ON fd.id = sp.field_device_id
		JOIN sps_controller_system_types st ON st.id = fd.sps_controller_system_type_id
		JOIN sps_controllers s ON s.id = st.sps_controller_id
		WHERE s.control_cabinet_id = @control_cabinet_id
		UNION
		SELECT 'bacnet_objects', bo.id
		FROM bacnet_objects bo
		JOIN field_devices fd ON fd.id = bo.field_device_id
		JOIN sps_controller_system_types st ON st.id = fd.sps_controller_system_type_id
		JOIN sps_controllers s ON s.id = st.sps_controller_id
		WHERE s.control_cabinet_id = @control_cabinet_id
		UNION
		SELECT 'bacnet_object_alarm_values', av.id
		FROM bacnet_object_alarm_values av
		JOIN bacnet_objects bo ON bo.id = av.bacnet_object_id
		JOIN field_devices fd ON fd.id = bo.field_device_id
		JOIN sps_controller_system_types st ON st.id = fd.sps_controller_system_type_id
		JOIN sps_controllers s ON s.id = st.sps_controller_id
		WHERE s.control_cabinet_id = @control_cabinet_id
		UNION
		SELECT 'project_control_cabinets', pcc.id
		FROM project_control_cabinets pcc
		WHERE pcc.control_cabinet_id = @control_cabinet_id
			AND (CAST(@project_id AS uuid) IS NULL OR pcc.project_id = CAST(@project_id AS uuid))
		UNION
		SELECT 'project_sps_controllers', psc.id
		FROM project_sps_controllers psc
		JOIN sps_controllers s ON s.id = psc.sps_controller_id
		WHERE s.control_cabinet_id = @control_cabinet_id
			AND (CAST(@project_id AS uuid) IS NULL OR psc.project_id = CAST(@project_id AS uuid))
		UNION
		SELECT 'project_field_devices', pfd.id
		FROM project_field_devices pfd
		JOIN field_devices fd ON fd.id = pfd.field_device_id
		JOIN sps_controller_system_types st ON st.id = fd.sps_controller_system_type_id
		JOIN sps_controllers s ON s.id = st.sps_controller_id
		WHERE s.control_cabinet_id = @control_cabinet_id
			AND (CAST(@project_id AS uuid) IS NULL OR pfd.project_id = CAST(@project_id AS uuid))
	`
	var projectArg any
	if projectID != nil {
		projectArg = *projectID
	}
	if err := s.db.WithContext(ctx).Raw(query,
		map[string]any{"control_cabinet_id": controlCabinetID, "project_id": projectArg},
	).Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		add(row.TableName, row.ID)
	}

	historyRows, err := s.historicalScopedTargets(ctx, controlCabinetID, projectID)
	if err != nil {
		return nil, err
	}
	for _, row := range historyRows {
		add(row.TableName, row.ID)
	}

	return targets, nil
}

type targetRow struct {
	TableName string    `gorm:"column:entity_table"`
	ID        uuid.UUID `gorm:"column:entity_id"`
}

func (s *Store) historicalScopedTargets(ctx context.Context, controlCabinetID uuid.UUID, projectID *uuid.UUID) ([]targetRow, error) {
	query := s.db.WithContext(ctx).
		Model(&domainHistory.ChangeEvent{}).
		Select("DISTINCT change_events.entity_table, change_events.entity_id").
		Joins("JOIN change_event_scopes ccs ON ccs.change_event_id = change_events.id AND ccs.scope_type = ? AND ccs.scope_id = ?", scopeControlCabinet, controlCabinetID)
	if projectID != nil {
		query = query.Joins("JOIN change_event_scopes pcs ON pcs.change_event_id = change_events.id AND pcs.scope_type = ? AND pcs.scope_id = ?", scopeProject, *projectID)
	}
	query = query.Where("change_events.entity_table IN ?", hierarchyRestoreTables)

	var rows []targetRow
	err := query.Scan(&rows).Error
	return rows, err
}

func (s *Store) latestVersionsForTargets(ctx context.Context, targets map[string]map[uuid.UUID]struct{}, asOf time.Time) (map[string]domainHistory.EntityVersion, error) {
	out := map[string]domainHistory.EntityVersion{}
	for table, ids := range targets {
		idList := sortedUUIDs(ids)
		if len(idList) == 0 {
			continue
		}
		for _, chunk := range uuidChunks(idList, 500) {
			var rows []domainHistory.EntityVersion
			err := s.db.WithContext(ctx).
				Raw(`
					SELECT DISTINCT ON (entity_table, entity_id) *
					FROM entity_versions
					WHERE entity_table = ?
						AND entity_id IN ?
						AND version_at <= ?
					ORDER BY entity_table, entity_id, version_at DESC, id DESC
				`, table, chunk, asOf).
				Scan(&rows).Error
			if err != nil {
				return nil, err
			}
			for _, row := range rows {
				out[targetKey(row.EntityTable, row.EntityID)] = row
			}
		}
	}
	return out, nil
}

func targetKey(table string, id uuid.UUID) string {
	return table + ":" + id.String()
}

func deleteRow(ctx context.Context, db *gorm.DB, table string, id uuid.UUID) error {
	if !allowedTable(table) {
		return fmt.Errorf("history table not allowed: %s", table)
	}
	if id == uuid.Nil {
		return nil
	}
	return db.WithContext(ctx).Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", quoteIdent(table)), id).Error
}

func upsertRow(ctx context.Context, db *gorm.DB, table string, snapshot domainHistory.JSONB) error {
	if !allowedTable(table) {
		return fmt.Errorf("history table not allowed: %s", table)
	}
	if len(snapshot) == 0 {
		return nil
	}
	var row map[string]any
	if err := json.Unmarshal(snapshot, &row); err != nil {
		return err
	}
	rawID, ok := row["id"]
	if !ok {
		return fmt.Errorf("history snapshot missing id for %s", table)
	}
	id, err := uuid.Parse(fmt.Sprint(rawID))
	if err != nil {
		return err
	}

	var count int64
	if err := db.WithContext(ctx).Table(table).Where("id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return db.WithContext(ctx).Table(table).Create(row).Error
	}
	delete(row, "id")
	return db.WithContext(ctx).Table(table).Where("id = ?", id).Updates(row).Error
}
