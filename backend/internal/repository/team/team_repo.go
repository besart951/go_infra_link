package team

import (
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type teamRepo struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) domainTeam.TeamRepository {
	return &teamRepo{db: db}
}

func (r *teamRepo) GetByIds(ids []uuid.UUID) ([]*domainTeam.Team, error) {
	if len(ids) == 0 {
		return []*domainTeam.Team{}, nil
	}
	var items []*domainTeam.Team
	err := r.db.Where("deleted_at IS NULL").Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *teamRepo) Create(entity *domainTeam.Team) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}
	return r.db.Create(entity).Error
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
	if len(ids) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.Model(&domainTeam.Team{}).
		Where("id IN ?", ids).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}

func (r *teamRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainTeam.Team], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainTeam.Team{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainTeam.Team
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainTeam.Team]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
