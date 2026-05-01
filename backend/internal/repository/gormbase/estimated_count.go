package gormbase

import (
	"context"

	"gorm.io/gorm"
)

func EstimatedTableCount(ctx context.Context, db *gorm.DB, table string) (int64, bool, error) {
	if db == nil || db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return 0, false, nil
	}

	var total int64
	err := db.WithContext(ctx).
		Raw("SELECT COALESCE(reltuples, 0)::bigint FROM pg_class WHERE oid = to_regclass(?)", table).
		Scan(&total).Error
	if err != nil {
		return 0, false, err
	}
	return total, total > 0, nil
}
