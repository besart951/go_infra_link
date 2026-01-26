package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type systemTypeRepo struct {
	*gormbase.BaseRepository[*domainFacility.SystemType]
}

func NewSystemTypeRepository(db *gorm.DB) domainFacility.SystemTypeRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.TrimSpace(search) + "%"
		return query.Where("name ILIKE ?", pattern)
	}
	
	baseRepo := gormbase.NewBaseRepository[*domainFacility.SystemType](db, searchCallback)
	return &systemTypeRepo{BaseRepository: baseRepo}
}

func (r *systemTypeRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.SystemType, error) {
	return r.BaseRepository.GetByIds(ids)
}

func (r *systemTypeRepo) Create(entity *domainFacility.SystemType) error {
	return r.BaseRepository.Create(entity)
}

func (r *systemTypeRepo) Update(entity *domainFacility.SystemType) error {
	return r.BaseRepository.Update(entity)
}

func (r *systemTypeRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *systemTypeRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemType], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}
	
	// Convert []*SystemType to []SystemType for the interface
	items := make([]domainFacility.SystemType, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}
	
	return &domain.PaginatedList[domainFacility.SystemType]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}
