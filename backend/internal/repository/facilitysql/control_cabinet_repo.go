package facilitysql

import (
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type controlCabinetRepo struct {
	db *gorm.DB
}

func NewControlCabinetRepository(db *gorm.DB) domainFacility.ControlCabinetRepository {
	return &controlCabinetRepo{db: db}
}

func (r *controlCabinetRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.ControlCabinet, error) {
	if len(ids) == 0 {
		return []*domainFacility.ControlCabinet{}, nil
	}
	var items []*domainFacility.ControlCabinet
	err := r.db.Where("deleted_at IS NULL").Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *controlCabinetRepo) Create(entity *domainFacility.ControlCabinet) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}
	return r.db.Create(entity).Error
}

func (r *controlCabinetRepo) Update(entity *domainFacility.ControlCabinet) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.Save(entity).Error
}

func (r *controlCabinetRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.Model(&domainFacility.ControlCabinet{}).
		Where("id IN ?", ids).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}

func (r *controlCabinetRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainFacility.ControlCabinet{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where("control_cabinet_nr ILIKE ?", pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.ControlCabinet
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.ControlCabinet]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
