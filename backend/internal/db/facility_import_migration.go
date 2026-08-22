package db

import (
	importsql "github.com/besart951/go_infra_link/backend/internal/repository/importsql"
	"gorm.io/gorm"
)

func migrateFacilityImportStaging(db *gorm.DB) error { return importsql.Migrate(db) }
