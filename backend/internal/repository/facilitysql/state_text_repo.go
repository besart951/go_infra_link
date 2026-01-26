package facilitysql

import (
	"strconv"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type stateTextRepo struct {
	*gormbase.BaseRepository[*domainFacility.StateText]
}

func NewStateTextRepository(db *gorm.DB) domainFacility.StateTextRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		// Try to parse search as int for RefNumber
		if refNum, err := strconv.Atoi(search); err == nil {
			return query.Where("ref_number = ?", refNum)
		}
		// Search in text fields if not a number
		return query.Where("state_text1 ILIKE ?", "%"+search+"%")
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.StateText](db, searchCallback)
	return &stateTextRepo{BaseRepository: baseRepo}
}

func (r *stateTextRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.StateText, error) {
	return r.BaseRepository.GetByIds(ids)
}

func (r *stateTextRepo) Create(entity *domainFacility.StateText) error {
	return r.BaseRepository.Create(entity)
}

func (r *stateTextRepo) Update(entity *domainFacility.StateText) error {
	return r.BaseRepository.Update(entity)
}

func (r *stateTextRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *stateTextRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.StateText], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*StateText to []StateText for the interface
	items := make([]domainFacility.StateText, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[domainFacility.StateText]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}
