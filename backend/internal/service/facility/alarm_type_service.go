package facility

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type AlarmTypeService struct {
	baseService[domainFacility.AlarmType]
	extRepo domainFacility.AlarmTypeRepository
}

func NewAlarmTypeService(repo domainFacility.AlarmTypeRepository) *AlarmTypeService {
	return &AlarmTypeService{
		baseService: newBase(repo, 20),
		extRepo:     repo,
	}
}

func (s *AlarmTypeService) Create(ctx context.Context, at *domainFacility.AlarmType) error {
	return s.extRepo.Create(ctx, at)
}

func (s *AlarmTypeService) Update(ctx context.Context, at *domainFacility.AlarmType) error {
	return s.extRepo.Update(ctx, at)
}

func (s *AlarmTypeService) GetWithFields(ctx context.Context, id uuid.UUID) (*domainFacility.AlarmType, error) {
	return s.extRepo.GetWithFields(ctx, id)
}
