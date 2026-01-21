package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type FieldDeviceService struct {
	repo domainFacility.FieldDeviceRepository
}

func NewFieldDeviceService(repo domainFacility.FieldDeviceRepository) *FieldDeviceService {
	return &FieldDeviceService{repo: repo}
}

func (s *FieldDeviceService) Create(fieldDevice *domainFacility.FieldDevice) error {
	return s.repo.Create(fieldDevice)
}

func (s *FieldDeviceService) GetByID(id uuid.UUID) (*domainFacility.FieldDevice, error) {
	fieldDevices, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(fieldDevices) == 0 {
		return nil, nil
	}
	return fieldDevices[0], nil
}

func (s *FieldDeviceService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	page, limit = normalizePagination(page, limit)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *FieldDeviceService) Update(fieldDevice *domainFacility.FieldDevice) error {
	return s.repo.Update(fieldDevice)
}

func (s *FieldDeviceService) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
}
