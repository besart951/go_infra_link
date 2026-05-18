package facilitysql

import (
	"context"
	"fmt"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bacnetReferenceUsageRepo struct {
	db *gorm.DB
}

type bacnetReferenceUsageRow struct {
	ID    uuid.UUID `gorm:"column:id"`
	Count int64     `gorm:"column:count"`
}

func NewBacnetReferenceUsageRepository(db *gorm.DB) domainFacility.BacnetReferenceUsageRepository {
	return &bacnetReferenceUsageRepo{db: db}
}

func (r *bacnetReferenceUsageRepo) CountByResource(ctx context.Context, resource domainFacility.BacnetReferenceResource, ids []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]int64{}, nil
	}

	switch resource {
	case domainFacility.BacnetReferenceResourceStateText:
		return r.countByBacnetObjectColumn(ctx, "state_text_id", ids)
	case domainFacility.BacnetReferenceResourceNotificationClass:
		return r.countByBacnetObjectColumn(ctx, "notification_class_id", ids)
	case domainFacility.BacnetReferenceResourceAlarmType:
		return r.countByBacnetObjectColumn(ctx, "alarm_type_id", ids)
	case domainFacility.BacnetReferenceResourceAlarmDefinition:
		return r.countAlarmDefinitions(ctx, ids)
	case domainFacility.BacnetReferenceResourceObjectData:
		return r.countObjectData(ctx, ids)
	case domainFacility.BacnetReferenceResourceApparat:
		return r.countApparats(ctx, ids)
	case domainFacility.BacnetReferenceResourceSystemPart:
		return r.countSystemParts(ctx, ids)
	case domainFacility.BacnetReferenceResourceSystemType:
		return r.countSystemTypes(ctx, ids)
	default:
		return nil, fmt.Errorf("unsupported bacnet reference resource %q", resource)
	}
}

func (r *bacnetReferenceUsageRepo) countByBacnetObjectColumn(ctx context.Context, column string, ids []uuid.UUID) (map[uuid.UUID]int64, error) {
	var rows []bacnetReferenceUsageRow
	err := r.db.WithContext(ctx).
		Table("bacnet_objects").
		Select(column+" AS id, COUNT(*) AS count").
		Where(column+" IN ?", ids).
		Group(column).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return usageRowsToMap(rows), nil
}

func (r *bacnetReferenceUsageRepo) countAlarmDefinitions(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int64, error) {
	const query = `
		SELECT ad.id AS id, COUNT(bo.id) AS count
		FROM alarm_definitions ad
		JOIN bacnet_objects bo ON bo.alarm_type_id = ad.alarm_type_id
		WHERE ad.id IN ?
		GROUP BY ad.id
	`
	return r.scanUsageRows(ctx, query, ids)
}

func (r *bacnetReferenceUsageRepo) countObjectData(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int64, error) {
	const query = `
		SELECT object_data_id AS id, COUNT(DISTINCT bacnet_object_id) AS count
		FROM object_data_bacnet_objects
		WHERE object_data_id IN ?
		GROUP BY object_data_id
	`
	return r.scanUsageRows(ctx, query, ids)
}

func (r *bacnetReferenceUsageRepo) countApparats(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int64, error) {
	const query = `
		SELECT id, COUNT(DISTINCT bacnet_object_id) AS count
		FROM (
			SELECT fd.apparat_id AS id, bo.id AS bacnet_object_id
			FROM field_devices fd
			JOIN bacnet_objects bo ON bo.field_device_id = fd.id
			WHERE fd.apparat_id IN ?
			UNION
			SELECT oda.apparat_id AS id, odb.bacnet_object_id AS bacnet_object_id
			FROM object_data_apparats oda
			JOIN object_data_bacnet_objects odb ON odb.object_data_id = oda.object_data_id
			WHERE oda.apparat_id IN ?
		) AS usage
		GROUP BY id
	`
	return r.scanUsageRows(ctx, query, ids, ids)
}

func (r *bacnetReferenceUsageRepo) countSystemParts(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int64, error) {
	const query = `
		SELECT id, COUNT(DISTINCT bacnet_object_id) AS count
		FROM (
			SELECT fd.system_part_id AS id, bo.id AS bacnet_object_id
			FROM field_devices fd
			JOIN bacnet_objects bo ON bo.field_device_id = fd.id
			WHERE fd.system_part_id IN ?
			UNION
			SELECT spa.system_part_id AS id, odb.bacnet_object_id AS bacnet_object_id
			FROM system_part_apparats spa
			JOIN object_data_apparats oda ON oda.apparat_id = spa.apparat_id
			JOIN object_data_bacnet_objects odb ON odb.object_data_id = oda.object_data_id
			WHERE spa.system_part_id IN ?
		) AS usage
		GROUP BY id
	`
	return r.scanUsageRows(ctx, query, ids, ids)
}

func (r *bacnetReferenceUsageRepo) countSystemTypes(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int64, error) {
	const query = `
		SELECT scst.system_type_id AS id, COUNT(DISTINCT bo.id) AS count
		FROM sps_controller_system_types scst
		JOIN field_devices fd ON fd.sps_controller_system_type_id = scst.id
		JOIN bacnet_objects bo ON bo.field_device_id = fd.id
		WHERE scst.system_type_id IN ?
		GROUP BY scst.system_type_id
	`
	return r.scanUsageRows(ctx, query, ids)
}

func (r *bacnetReferenceUsageRepo) scanUsageRows(ctx context.Context, query string, args ...any) (map[uuid.UUID]int64, error) {
	var rows []bacnetReferenceUsageRow
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return usageRowsToMap(rows), nil
}

func usageRowsToMap(rows []bacnetReferenceUsageRow) map[uuid.UUID]int64 {
	result := make(map[uuid.UUID]int64, len(rows))
	for _, row := range rows {
		result[row.ID] = row.Count
	}
	return result
}
