package db

import (
	"fmt"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	projectrepo "github.com/besart951/go_infra_link/backend/internal/repository/project"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var defaultProjectPhaseID = uuid.MustParse("019c780c-f7eb-709a-93dc-5e7458cf4466")

var seedProjectPhases = []struct {
	id   uuid.UUID
	name string
}{
	{uuid.MustParse("019c780c-f7eb-709a-93dc-5e7458cf4466"), "SIA 21"},
	{uuid.MustParse("019c780c-f7eb-709b-99f8-b9c44a64353c"), "SIA 31"},
	{uuid.MustParse("019c780c-f7eb-709c-b571-e26f5de5fa72"), "SIA 41"},
	{uuid.MustParse("019c780c-f7eb-709d-b887-4acad7e198fd"), "SIA 51"},
}

func migrateProjectPhases(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&project.Phase{}); err != nil {
			return err
		}
		if err := seedProjectPhaseRows(tx); err != nil {
			return err
		}
		if err := ensureProjectsPhaseIDColumn(tx); err != nil {
			return err
		}
		if err := backfillProjectsPhaseID(tx); err != nil {
			return err
		}
		return enforceProjectsPhaseIDDefault(tx)
	})
}

func seedProjectPhaseRows(tx *gorm.DB) error {
	now := time.Now().UTC()
	for _, seed := range seedProjectPhases {
		phase := project.Phase{
			Base: domain.Base{
				ID:        seed.id,
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name: seed.name,
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&phase).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureProjectsPhaseIDColumn(tx *gorm.DB) error {
	if tx.Migrator().HasColumn(&projectrepo.ProjectRecord{}, "phase_id") {
		return nil
	}

	if tx.Dialector != nil && tx.Dialector.Name() == "postgres" {
		return tx.Exec("ALTER TABLE projects ADD COLUMN IF NOT EXISTS phase_id uuid").Error
	}
	return tx.Exec("ALTER TABLE projects ADD COLUMN phase_id uuid").Error
}

func backfillProjectsPhaseID(tx *gorm.DB) error {
	return tx.Exec(
		"UPDATE projects SET phase_id = ? WHERE phase_id IS NULL OR phase_id = ?",
		defaultProjectPhaseID,
		uuid.Nil,
	).Error
}

func enforceProjectsPhaseIDDefault(tx *gorm.DB) error {
	if tx.Dialector == nil || tx.Dialector.Name() != "postgres" {
		return nil
	}

	defaultValue := fmt.Sprintf("'%s'::uuid", defaultProjectPhaseID.String())
	if err := tx.Exec("ALTER TABLE projects ALTER COLUMN phase_id SET DEFAULT " + defaultValue).Error; err != nil {
		return err
	}
	return tx.Exec("ALTER TABLE projects ALTER COLUMN phase_id SET NOT NULL").Error
}
