package db

import (
	"github.com/besart951/go_infra_link/backend/internal/repository/searchspec"
	"gorm.io/gorm"
)

func migratePGTrgmSearch(db *gorm.DB) error {
	if db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}

	statements := []string{"CREATE EXTENSION IF NOT EXISTS pg_trgm"}
	for _, spec := range searchspec.All {
		statements = append(statements, spec.IndexStatements()...)
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
