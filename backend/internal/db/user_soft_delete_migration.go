package db

import (
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"gorm.io/gorm"
)

func migrateUserSoftDelete(db *gorm.DB) error {
	if db.Dialector != nil && db.Dialector.Name() == "postgres" {
		if err := db.Exec("ALTER TABLE users ALTER COLUMN email DROP NOT NULL").Error; err != nil {
			return err
		}
	}
	if err := db.AutoMigrate(&domainUser.User{}); err != nil {
		return err
	}
	return db.Model(&domainUser.User{}).
		Where("anonymized_at IS NOT NULL AND email IS NOT NULL").
		Update("email", nil).Error
}
