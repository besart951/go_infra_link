package facilitysql

import (
	"context"
	"sort"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type deleteImpactRepo struct {
	db *gorm.DB
}

type deleteImpactRow struct {
	ID       uuid.UUID `gorm:"column:id"`
	Resource string    `gorm:"column:resource"`
	Count    int64     `gorm:"column:count"`
}

func NewDeleteImpactRepository(db *gorm.DB) domainFacility.DeleteImpactRepository {
	return &deleteImpactRepo{db: db}
}

func (r *deleteImpactRepo) DeleteImpacts(ctx context.Context, resource domainFacility.DeleteImpactResource, ids []uuid.UUID) ([]domainFacility.DeleteImpact, error) {
	if len(ids) == 0 {
		return []domainFacility.DeleteImpact{}, nil
	}

	rows, err := r.rows(ctx, resource, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID][]domainFacility.DeleteImpactBlocker, len(ids))
	for _, row := range rows {
		if row.Count <= 0 {
			continue
		}
		byID[row.ID] = append(byID[row.ID], domainFacility.DeleteImpactBlocker{Resource: row.Resource, Count: row.Count})
	}

	impacts := make([]domainFacility.DeleteImpact, 0, len(ids))
	for _, id := range ids {
		blockers := byID[id]
		sort.Slice(blockers, func(i, j int) bool { return blockers[i].Resource < blockers[j].Resource })
		impacts = append(impacts, domainFacility.DeleteImpact{Resource: resource, ID: id, Blockers: blockers})
	}
	return impacts, nil
}

func (r *deleteImpactRepo) rows(ctx context.Context, resource domainFacility.DeleteImpactResource, ids []uuid.UUID) ([]deleteImpactRow, error) {
	query := ""
	switch resource {
	case domainFacility.DeleteImpactResourceApparat:
		query = `
			SELECT apparat_id AS id, 'field_devices' AS resource, COUNT(*) AS count
			FROM field_devices WHERE apparat_id IN ? GROUP BY apparat_id
			UNION ALL
			SELECT apparat_id AS id, 'system_parts' AS resource, COUNT(*) AS count
			FROM system_part_apparats WHERE apparat_id IN ? GROUP BY apparat_id
			UNION ALL
			SELECT apparat_id AS id, 'object_data' AS resource, COUNT(*) AS count
			FROM object_data_apparats WHERE apparat_id IN ? GROUP BY apparat_id
		`
	case domainFacility.DeleteImpactResourceSystemPart:
		query = `
			SELECT system_part_id AS id, 'field_devices' AS resource, COUNT(*) AS count
			FROM field_devices WHERE system_part_id IN ? GROUP BY system_part_id
			UNION ALL
			SELECT system_part_id AS id, 'apparats' AS resource, COUNT(*) AS count
			FROM system_part_apparats WHERE system_part_id IN ? GROUP BY system_part_id
		`
	default:
		return []deleteImpactRow{}, nil
	}

	var rows []deleteImpactRow
	args := []any{ids, ids}
	if resource == domainFacility.DeleteImpactResourceApparat {
		args = append(args, ids)
	}
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
