package db

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/auth"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&user.User{},
		&user.BusinessDetails{},
		&auth.RefreshToken{},
		&project.Project{},
		&project.Phase{},
		// Facility Domain
		&facility.Building{},
		&facility.ControlCabinet{},
		&facility.SPSController{},
		&facility.SPSControllerSystemType{},
		&facility.FieldDevice{},
		// &facility.IOPoint{}, // Removed
		&facility.SystemType{},
		&facility.SystemPart{},
		&facility.Specification{},
		&facility.NotificationClass{},
		&facility.AlarmDefinition{},
		&facility.Apparat{},
		&facility.ObjectData{},
		&facility.ObjectDataHistory{},
		&facility.BacnetObject{},
		// &facility.BacnetObjectProperty{}, // Removed
		&facility.StateText{},
	)
}
