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
}

func NewControlCabinetRepository(db *gorm.DB) domainFacility.ControlCabinetRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.TrimSpace(search) + "%"
		return query.Where("control_cabinet_nr ILIKE ?", pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.ControlCabinet](db, searchCallback)
	return &controlCabinetRepo{BaseRepository: baseRepo}
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
