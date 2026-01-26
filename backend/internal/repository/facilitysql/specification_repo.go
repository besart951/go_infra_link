package facilitysql

import (
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type specificationRepo struct {
	db *gorm.DB
}

func NewSpecificationRepository(db *gorm.DB) domainFacility.SpecificationStore {
	return &specificationRepo{db: db}
}

func (r *specificationRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.Specification, error) {
	if len(ids) == 0 {
		return []*domainFacility.Specification{}, nil
	}
	var items []*domainFacility.Specification
	err := r.db.Where("deleted_at IS NULL").Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *specificationRepo) Create(entity *domainFacility.Specification) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}
	return r.db.Create(entity).Error
}

func (r *specificationRepo) Update(entity *domainFacility.Specification) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.Save(entity).Error
}

func (r *specificationRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.Model(&domainFacility.Specification{}).
		Where("id IN ?", ids).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}

func (r *specificationRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Specification], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainFacility.Specification{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where("specification_supplier ILIKE ? OR specification_brand ILIKE ? OR specification_type ILIKE ?", pattern, pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.Specification
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.Specification]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *specificationRepo) GetByFieldDeviceIDs(fieldDeviceIDs []uuid.UUID) ([]*domainFacility.Specification, error) {
	if len(fieldDeviceIDs) == 0 {
		return []*domainFacility.Specification{}, nil
	}
	var items []*domainFacility.Specification
	err := r.db.Where("deleted_at IS NULL").Where("field_device_id IN ?", fieldDeviceIDs).Find(&items).Error
	return items, err
}

func (r *specificationRepo) SoftDeleteByFieldDeviceIDs(fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.Model(&domainFacility.Specification{}).
		Where("field_device_id IN ?", fieldDeviceIDs).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}
