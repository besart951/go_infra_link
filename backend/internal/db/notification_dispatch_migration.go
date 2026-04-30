package db

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"gorm.io/gorm"
)

func migrateNotificationDispatch(db *gorm.DB) error {
	return db.AutoMigrate(
		&notification.EmailOutbox{},
		&notification.NotificationRule{},
	)
}
