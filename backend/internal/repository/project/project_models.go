package project

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type ProjectRecord struct {
	domain.Base
	Name        string `gorm:"not null"`
	Description string
	Status      domainProject.ProjectStatus `gorm:"type:varchar(50);not null"`
	StartDate   *time.Time
	PhaseID     uuid.UUID `gorm:"type:uuid;not null"`
	CreatorID   uuid.UUID `gorm:"type:uuid;not null"`
}

func (ProjectRecord) TableName() string {
	return "projects"
}

type ProjectUserRecord struct {
	ProjectID uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
}

func (ProjectUserRecord) TableName() string {
	return "project_users"
}

func toProjectRecord(entity *domainProject.Project) *ProjectRecord {
	if entity == nil {
		return nil
	}

	record := &ProjectRecord{
		Base:        entity.Base,
		Name:        entity.Name,
		Description: entity.Description,
		Status:      entity.Status,
		StartDate:   entity.StartDate,
		PhaseID:     entity.PhaseID,
		CreatorID:   entity.CreatorID,
	}

	return record
}

func toProjectDomain(record *ProjectRecord) *domainProject.Project {
	if record == nil {
		return nil
	}

	entity := &domainProject.Project{
		Base:        record.Base,
		Name:        record.Name,
		Description: record.Description,
		Status:      record.Status,
		StartDate:   record.StartDate,
		PhaseID:     record.PhaseID,
		CreatorID:   record.CreatorID,
	}

	return entity
}

func toProjectDomains(records []*ProjectRecord) []*domainProject.Project {
	items := make([]*domainProject.Project, len(records))
	for i, record := range records {
		items[i] = toProjectDomain(record)
	}
	return items
}

func projectDomainValues(records []ProjectRecord) []domainProject.Project {
	items := make([]domainProject.Project, len(records))
	for i := range records {
		items[i] = *toProjectDomain(&records[i])
	}
	return items
}
