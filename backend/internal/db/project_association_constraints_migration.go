package db

import (
	"fmt"

	"gorm.io/gorm"
)

type projectAssociationForeignKey struct {
	name     string
	table    string
	onDelete string
}

var projectAssociationForeignKeys = []projectAssociationForeignKey{
	{
		name:     "fk_project_control_cabinets_project",
		table:    "project_control_cabinets",
		onDelete: "RESTRICT",
	},
	{
		name:     "fk_project_sps_controllers_project",
		table:    "project_sps_controllers",
		onDelete: "RESTRICT",
	},
	{
		name:     "fk_project_field_devices_project",
		table:    "project_field_devices",
		onDelete: "RESTRICT",
	},
	{
		name:     "fk_project_users_project",
		table:    "project_users",
		onDelete: "CASCADE",
	},
}

// migrateProjectAssociationForeignKeys closes the concurrency window between
// project-deletion eligibility checks and project-link inserts. It reports
// existing orphan rows and stops; it never repairs or drops user data.
func migrateProjectAssociationForeignKeys(db *gorm.DB) error {
	if db == nil || db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, foreignKey := range projectAssociationForeignKeys {
			var orphanCount int64
			orphanQuery := fmt.Sprintf(`
				SELECT COUNT(*)
				FROM %s AS association
				LEFT JOIN projects AS project ON project.id = association.project_id
				WHERE project.id IS NULL
			`, foreignKey.table)
			if err := tx.Raw(orphanQuery).Scan(&orphanCount).Error; err != nil {
				return fmt.Errorf(
					"inspect %s project associations: %w",
					foreignKey.table,
					err,
				)
			}
			if orphanCount != 0 {
				return fmt.Errorf(
					"cannot install %s: %s contains %d orphan project association rows",
					foreignKey.name,
					foreignKey.table,
					orphanCount,
				)
			}

			var constraint struct {
				Validated bool `gorm:"column:convalidated"`
			}
			result := tx.Raw(`
				SELECT convalidated
				FROM pg_constraint
				WHERE conname = ?
				  AND conrelid = ?::regclass
			`, foreignKey.name, foreignKey.table).Scan(&constraint)
			if result.Error != nil {
				return fmt.Errorf("inspect constraint %s: %w", foreignKey.name, result.Error)
			}

			if result.RowsAffected == 0 {
				addConstraint := fmt.Sprintf(`
					ALTER TABLE %s
					ADD CONSTRAINT %s
					FOREIGN KEY (project_id)
					REFERENCES projects(id)
					ON UPDATE CASCADE
					ON DELETE %s
					NOT VALID
				`, foreignKey.table, foreignKey.name, foreignKey.onDelete)
				if err := tx.Exec(addConstraint).Error; err != nil {
					return fmt.Errorf("add constraint %s: %w", foreignKey.name, err)
				}
			}

			if result.RowsAffected == 0 || !constraint.Validated {
				validateConstraint := fmt.Sprintf(
					"ALTER TABLE %s VALIDATE CONSTRAINT %s",
					foreignKey.table,
					foreignKey.name,
				)
				if err := tx.Exec(validateConstraint).Error; err != nil {
					return fmt.Errorf(
						"validate constraint %s; resolve reported orphan rows before retrying: %w",
						foreignKey.name,
						err,
					)
				}
			}
		}
		return nil
	})
}
