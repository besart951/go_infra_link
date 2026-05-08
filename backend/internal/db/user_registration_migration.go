package db

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"gorm.io/gorm"
)

func migrateUserRegistrations(db *gorm.DB) error {
	return db.AutoMigrate(&user.UserInvitation{})
}
