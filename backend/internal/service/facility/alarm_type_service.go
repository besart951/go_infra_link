package facility

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type AlarmTypeService struct {
	baseService[domainFacility.AlarmType]
	extRepo     domainFacility.AlarmTypeRepository
	deleteGuard bacnetReferenceDeleteGuard
}

func NewAlarmTypeService(repo domainFacility.AlarmTypeRepository, usageRepos ...domainFacility.BacnetReferenceUsageRepository) *AlarmTypeService {
	return &AlarmTypeService{
		baseService: newBase(repo, 20),
		extRepo:     repo,
		deleteGuard: newBacnetReferenceDeleteGuard(domainFacility.BacnetReferenceResourceAlarmType, usageRepos...),
	}
}

func (s *AlarmTypeService) Create(ctx context.Context, at *domainFacility.AlarmType) error {
	return s.extRepo.Create(ctx, at)
}

func (s *AlarmTypeService) Update(ctx context.Context, at *domainFacility.AlarmType) error {
	return s.extRepo.Update(ctx, at)
}

func (s *AlarmTypeService) DeleteByID(ctx context.Context, id uuid.UUID) error {
	if err := s.deleteGuard.ensureDeleteAllowed(ctx, id); err != nil {
		return err
	}
	return s.baseService.DeleteByID(ctx, id)
}

func (s *AlarmTypeService) GetWithFields(ctx context.Context, id uuid.UUID) (*domainFacility.AlarmType, error) {
	return s.extRepo.GetWithFields(ctx, id)
}
