package db

import (
	"errors"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"gorm.io/gorm"
)

func ensureNotificationSMTPManagePermission(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var permission user.Permission
		err := tx.Where("name = ?", user.PermissionNotificationSMTPManage).First(&permission).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			now := time.Now().UTC()
			permission = user.Permission{
				Name:        user.PermissionNotificationSMTPManage,
				Resource:    "notification.smtp",
				Action:      "manage",
				Description: "Manage SMTP notification settings",
			}
			if err := permission.InitForCreate(now); err != nil {
				return err
			}
			if err := tx.Create(&permission).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		var rolePermission user.RolePermission
		err = tx.Where("role = ? AND permission = ?", user.RoleSuperAdmin, user.PermissionNotificationSMTPManage).First(&rolePermission).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		now := time.Now().UTC()
		rolePermission = user.RolePermission{Role: user.RoleSuperAdmin, Permission: user.PermissionNotificationSMTPManage}
		if err := rolePermission.InitForCreate(now); err != nil {
			return err
		}
		return tx.Create(&rolePermission).Error
	})
}