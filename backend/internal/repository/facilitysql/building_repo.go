package facilitysql

import (
	"strconv"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type buildingRepo struct {
	db *gorm.DB
}

func NewBuildingRepository(db *gorm.DB) domainFacility.BuildingRepository {
	return &buildingRepo{db: db}
}

func (r *buildingRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.Building, error) {
	if len(ids) == 0 {
		return []*domainFacility.Building{}, nil
	}
	var items []*domainFacility.Building
	err := r.db.Where("deleted_at IS NULL").Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *buildingRepo) Create(entity *domainFacility.Building) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}
	return r.db.Create(entity).Error
}

func (r *buildingRepo) Update(entity *domainFacility.Building) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.Save(entity).Error
}

func (r *buildingRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.Model(&domainFacility.Building{}).
		Where("id IN ?", ids).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}

func (r *buildingRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Building], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainFacility.Building{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.TrimSpace(params.Search) + "%"
		if num, err := strconv.Atoi(strings.TrimSpace(params.Search)); err == nil {
			query = query.Where("iws_code ILIKE ? OR building_group = ?", pattern, num)
		} else {
			query = query.Where("iws_code ILIKE ?", pattern)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.Building
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.Building]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
