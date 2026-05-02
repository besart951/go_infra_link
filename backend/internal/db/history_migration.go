package db

import (
	"github.com/besart951/go_infra_link/backend/internal/repository/historysql"
	"gorm.io/gorm"
)

func migrateHistory(db *gorm.DB) error {
	return historysql.AutoMigrate(db)
}
