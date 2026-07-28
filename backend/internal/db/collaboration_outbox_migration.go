package db

import (
	"github.com/besart951/go_infra_link/backend/internal/repository/collaborationoutbox"
	"gorm.io/gorm"
)

func migrateCollaborationOutbox(db *gorm.DB) error {
	return collaborationoutbox.AutoMigrate(db)
}
