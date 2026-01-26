package team

import (
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type teamRepo struct {
	*gormbase.BaseRepository[*domainTeam.Team]
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) domainTeam.TeamRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.TrimSpace(search) + "%"
		return query.Where("name ILIKE ? OR description ILIKE ?", pattern, pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainTeam.Team](db, searchCallback)
	return &teamRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *teamRepo) GetByIds(ids []uuid.UUID) ([]*domainTeam.Team, error) {
	return r.BaseRepository.GetByIds(ids)
}

func (r *teamRepo) Create(entity *domainTeam.Team) error {
	return r.BaseRepository.Create(entity)
}

func (r *teamRepo) Update(entity *domainTeam.Team) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.Model(&domainTeam.Team{}).
		Where("deleted_at IS NULL AND id = ?", entity.ID).
		Updates(map[string]any{
			"updated_at":  entity.UpdatedAt,
			"name":        entity.Name,
			"description": entity.Description,
		}).Error
}

func (r *teamRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *teamRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainTeam.Team], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*Team to []Team for the interface
	items := make([]domainTeam.Team, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[domainTeam.Team]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}
