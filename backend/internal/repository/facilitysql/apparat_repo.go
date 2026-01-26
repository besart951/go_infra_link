package facilitysql

import (
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type apparatRepo struct {
	db *gorm.DB
}

func NewApparatRepository(db *gorm.DB) domainFacility.ApparatRepository {
	return &apparatRepo{db: db}
}

func (r *apparatRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	if len(ids) == 0 {
		return []*domainFacility.Apparat{}, nil
	}
	var items []*domainFacility.Apparat
	err := r.db.Where("deleted_at IS NULL").Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *apparatRepo) Create(entity *domainFacility.Apparat) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}
	return r.db.Create(entity).Error
}

func (r *apparatRepo) Update(entity *domainFacility.Apparat) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.Save(entity).Error
}

func (r *apparatRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.Model(&domainFacility.Apparat{}).
		Where("id IN ?", ids).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}

func (r *apparatRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainFacility.Apparat{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where("short_name ILIKE ? OR name ILIKE ?", pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.Apparat
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.Apparat]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
