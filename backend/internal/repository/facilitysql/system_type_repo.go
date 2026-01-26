package facilitysql

import (
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type systemTypeRepo struct {
	db *gorm.DB
}

func NewSystemTypeRepository(db *gorm.DB) domainFacility.SystemTypeRepository {
	return &systemTypeRepo{db: db}
}

func (r *systemTypeRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.SystemType, error) {
	if len(ids) == 0 {
		return []*domainFacility.SystemType{}, nil
	}
	var items []*domainFacility.SystemType
	err := r.db.Where("deleted_at IS NULL").Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *systemTypeRepo) Create(entity *domainFacility.SystemType) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}
	return r.db.Create(entity).Error
}

func (r *systemTypeRepo) Update(entity *domainFacility.SystemType) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.Save(entity).Error
}

func (r *systemTypeRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.Model(&domainFacility.SystemType{}).
		Where("id IN ?", ids).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}

func (r *systemTypeRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemType], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainFacility.SystemType{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where("name ILIKE ?", pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.SystemType
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.SystemType]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
