package db

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/repository/historysql"
	"gorm.io/gorm"
)

func migrateHistory(db *gorm.DB) error {
	return historysql.AutoMigrate(db)
}

func migrateHistoryV2(db *gorm.DB) error {
	return historysql.AutoMigrateV2(db, time.Now().UTC())
}
