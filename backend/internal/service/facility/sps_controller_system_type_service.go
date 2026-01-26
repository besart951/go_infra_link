package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

type SPSControllerSystemTypeService struct {
	repo domainFacility.SPSControllerSystemTypeStore
}

func NewSPSControllerSystemTypeService(repo domainFacility.SPSControllerSystemTypeStore) *SPSControllerSystemTypeService {
	return &SPSControllerSystemTypeService{repo: repo}
}

func (s *SPSControllerSystemTypeService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}
