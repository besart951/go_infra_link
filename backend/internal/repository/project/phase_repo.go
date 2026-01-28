package project

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"gorm.io/gorm"
)

type phaseRepo struct {
	*gormbase.BaseRepository[*domainProject.Phase]
	db *gorm.DB
}

func NewPhaseRepository(db *gorm.DB) domainProject.PhaseRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(name) LIKE ?", pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainProject.Phase](db, searchCallback)
	return &phaseRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *phaseRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainProject.Phase], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*Phase to []Phase for the interface
	items := make([]domainProject.Phase, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[domainProject.Phase]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}
