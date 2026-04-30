package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/besart951/go_infra_link/backend/internal/repository/searchspec"
	"gorm.io/gorm"
)

type phaseRepo struct {
	*gormbase.BaseRepository[*domainProject.Phase]
	db *gorm.DB
}

func NewPhaseRepository(db *gorm.DB) domainProject.PhaseRepository {
	baseRepo := gormbase.NewBaseRepository(db,
		gormbase.TrigramSearchCallback[*domainProject.Phase](searchspec.Phases.SearchColumns("")...),
	)
	return &phaseRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *phaseRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainProject.Phase], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}
