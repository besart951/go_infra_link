package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type controlCabinetRepo struct {
	*gormbase.BaseRepository[*domainFacility.ControlCabinet]
	db *gorm.DB
}

func NewControlCabinetRepository(db *gorm.DB) domainFacility.ControlCabinetRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(control_cabinet_nr) LIKE ?", pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.ControlCabinet](db, searchCallback)
	return &controlCabinetRepo{BaseRepository: baseRepo, db: db}
}

func (r *controlCabinetRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.ControlCabinet, error) {
	return r.BaseRepository.GetByIds(ids)
}

func (r *controlCabinetRepo) Create(entity *domainFacility.ControlCabinet) error {
	return r.BaseRepository.Create(entity)
}

func (r *controlCabinetRepo) Update(entity *domainFacility.ControlCabinet) error {
	return r.BaseRepository.Update(entity)
}

func (r *controlCabinetRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *controlCabinetRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*ControlCabinet to []ControlCabinet for the interface
	items := make([]domainFacility.ControlCabinet, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[domainFacility.ControlCabinet]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}

func (r *controlCabinetRepo) GetPaginatedListByBuildingID(buildingID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainFacility.ControlCabinet{}).
		Where("deleted_at IS NULL").
		Where("building_id = ?", buildingID)

	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Where("LOWER(control_cabinet_nr) LIKE ?", pattern)
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

func (r *controlCabinetRepo) ExistsControlCabinetNr(buildingID uuid.UUID, controlCabinetNr string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.Model(&domainFacility.ControlCabinet{}).
		Where("deleted_at IS NULL").
		Where("building_id = ?", buildingID).
		Where("LOWER(control_cabinet_nr) = ?", strings.ToLower(strings.TrimSpace(controlCabinetNr)))

	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
