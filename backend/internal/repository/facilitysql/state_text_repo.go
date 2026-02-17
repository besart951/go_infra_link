package facilitysql

import (
	"strconv"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
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
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(state_text1) LIKE ?", pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.StateText](db, searchCallback)
	return &stateTextRepo{BaseRepository: baseRepo}
}

func (r *stateTextRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.StateText], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}
