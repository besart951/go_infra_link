package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type spsControllerRepo struct {
	*gormbase.BaseRepository[*domainFacility.SPSController]
	db *gorm.DB
}

func NewSPSControllerRepository(db *gorm.DB) domainFacility.SPSControllerRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(device_name) LIKE ? OR LOWER(ip_address) LIKE ?", pattern, pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.SPSController](db, searchCallback)
	return &spsControllerRepo{BaseRepository: baseRepo, db: db}
}

func (r *spsControllerRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.SPSController, error) {
	return r.BaseRepository.GetByIds(ids)
}

func (r *spsControllerRepo) Create(entity *domainFacility.SPSController) error {
	return r.BaseRepository.Create(entity)
}

func (r *spsControllerRepo) Update(entity *domainFacility.SPSController) error {
	return r.BaseRepository.Update(entity)
}

func (r *spsControllerRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *spsControllerRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSController], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*SPSController to []SPSController for the interface
	items := make([]domainFacility.SPSController, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[domainFacility.SPSController]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}

func (r *spsControllerRepo) GetPaginatedListByControlCabinetID(controlCabinetID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSController], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainFacility.SPSController{}).
		Where("deleted_at IS NULL").
		Where("control_cabinet_id = ?", controlCabinetID)

	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Where("LOWER(device_name) LIKE ? OR LOWER(ip_address) LIKE ?", pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.SPSController
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.SPSController]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *spsControllerRepo) GetIDsByControlCabinetID(controlCabinetID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&domainFacility.SPSController{}).
		Where("deleted_at IS NULL").
		Where("control_cabinet_id = ?", controlCabinetID).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}
