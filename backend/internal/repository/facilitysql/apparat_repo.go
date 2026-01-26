package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type apparatRepo struct {
	*gormbase.BaseRepository[*domainFacility.Apparat]
}

func NewApparatRepository(db *gorm.DB) domainFacility.ApparatRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.TrimSpace(search) + "%"
		return query.Where("short_name ILIKE ? OR name ILIKE ?", pattern, pattern)
	}
	
	baseRepo := gormbase.NewBaseRepository[*domainFacility.Apparat](db, searchCallback)
	return &apparatRepo{BaseRepository: baseRepo}
}

func (r *apparatRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	return r.BaseRepository.GetByIds(ids)
}

func (r *apparatRepo) Create(entity *domainFacility.Apparat) error {
	return r.BaseRepository.Create(entity)
}

func (r *apparatRepo) Update(entity *domainFacility.Apparat) error {
	return r.BaseRepository.Update(entity)
}

func (r *apparatRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *apparatRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}
	
	// Convert []*Apparat to []Apparat for the interface
	items := make([]domainFacility.Apparat, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}
	
	return &domain.PaginatedList[domainFacility.Apparat]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}
