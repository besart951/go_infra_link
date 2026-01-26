package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type systemPartRepo struct {
	*gormbase.BaseRepository[*domainFacility.SystemPart]
}

func NewSystemPartRepository(db *gorm.DB) domainFacility.SystemPartRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.TrimSpace(search) + "%"
		return query.Where("short_name ILIKE ? OR name ILIKE ?", pattern, pattern)
	}
	
	baseRepo := gormbase.NewBaseRepository[*domainFacility.SystemPart](db, searchCallback)
	return &systemPartRepo{BaseRepository: baseRepo}
}

func (r *systemPartRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	return r.BaseRepository.GetByIds(ids)
}

func (r *systemPartRepo) Create(entity *domainFacility.SystemPart) error {
	return r.BaseRepository.Create(entity)
}

func (r *systemPartRepo) Update(entity *domainFacility.SystemPart) error {
	return r.BaseRepository.Update(entity)
}

func (r *systemPartRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *systemPartRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemPart], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}
	
	// Convert []*SystemPart to []SystemPart for the interface
	items := make([]domainFacility.SystemPart, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}
	
	return &domain.PaginatedList[domainFacility.SystemPart]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}
