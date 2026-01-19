package projectrepo

import (
	"github.com/besart951/go_infra_link/backend/internal/adapters/storage"
	"github.com/besart951/go_infra_link/backend/internal/core/domain/project"
	"gorm.io/gorm"
)

type ProjectStorage struct {
	storage.BaseRepository[project.Project]
}

func NewProjectStorage(db *gorm.DB) project.ProjectRepository {
	return &ProjectStorage{
		BaseRepository: storage.BaseRepository[project.Project]{DB: db},
	}
}
