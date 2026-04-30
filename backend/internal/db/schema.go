package db

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/auth"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	facilityrepo "github.com/besart951/go_infra_link/backend/internal/repository/facilitysql"
	projectrepo "github.com/besart951/go_infra_link/backend/internal/repository/project"
	projectsql "github.com/besart951/go_infra_link/backend/internal/repository/projectsql"
	"gorm.io/gorm"
)

// autoMigrateCurrentSchema creates the current schema baseline.
// It first creates tables without relationship constraints so GORM cannot trip
// over relationship cycles, then runs AutoMigrate normally to add associations.
func autoMigrateCurrentSchema(db *gorm.DB) error {
	models := currentSchemaModels()
	if err := autoMigrateTablesOnly(db, models...); err != nil {
		return err
	}
	return db.AutoMigrate(models...)
}

func currentSchemaModels() []any {
	models := []any{
		&user.User{},
		&user.BusinessDetails{},
		&user.Permission{},
		&user.RolePermission{},

		&auth.RefreshToken{},
		&notification.SMTPSettings{},
		&notification.UserPreference{},
		&notification.SystemNotification{},
		&notification.EmailOutbox{},
		&notification.NotificationRule{},

		&team.Team{},
		&team.TeamMember{},
		&user.UserTeam{},

		&project.Phase{},
		&project.PhasePermission{},
		&facility.Building{},
		&facility.ControlCabinet{},
		&facility.SPSController{},
		&facility.SystemType{},
		&facility.SPSControllerSystemType{},
		&facility.SystemPart{},
		&facility.Apparat{},
		&facility.Specification{},
		&facility.StateText{},
		&facility.NotificationClass{},
		&facility.AlarmDefinition{},
		&facility.BacnetObject{},
		&facility.ObjectData{},
		&facility.Unit{},
		&facility.AlarmField{},
		&facility.AlarmType{},
		&facility.AlarmTypeField{},
		&facility.AlarmDefinitionFieldOverride{},
		&facility.BacnetObjectAlarmValue{},

		&projectrepo.ProjectRecord{},
		&projectrepo.ProjectUserRecord{},
		&projectsql.ProjectControlCabinetRecord{},
		&projectsql.ProjectSPSControllerRecord{},
		&projectsql.ProjectFieldDeviceRecord{},
		&facilityrepo.FieldDeviceRecord{},
	}

	return models
}

func autoMigrateTablesOnly(db *gorm.DB, models ...any) error {
	previous := db.Config.IgnoreRelationshipsWhenMigrating
	db.Config.IgnoreRelationshipsWhenMigrating = true
	defer func() {
		db.Config.IgnoreRelationshipsWhenMigrating = previous
	}()

	return db.AutoMigrate(models...)
}
