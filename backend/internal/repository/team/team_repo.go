package team

import (
	"context"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/besart951/go_infra_link/backend/internal/repository/searchspec"
	"gorm.io/gorm"
)

type teamRepo struct {
	*gormbase.BaseRepository[*domainTeam.Team]
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) domainTeam.TeamRepository {
	baseRepo := gormbase.NewBaseRepository(db,
		gormbase.TrigramSearchCallback[*domainTeam.Team](searchspec.Teams.SearchColumns("")...),
	)
	return &teamRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *teamRepo) Update(ctx context.Context, entity *domainTeam.Team) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.WithContext(ctx).Model(&domainTeam.Team{}).
		Where("id = ?", entity.ID).
		Updates(map[string]any{
			"updated_at":  entity.UpdatedAt,
			"name":        entity.Name,
			"description": entity.Description,
		}).Error
}

func (r *teamRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainTeam.Team], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}
