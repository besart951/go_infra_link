package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type AlarmTypeService struct {
	baseService[domainFacility.AlarmType]
	extRepo domainFacility.AlarmTypeRepository
}

func NewAlarmTypeService(repo domainFacility.AlarmTypeRepository) *AlarmTypeService {
	return &AlarmTypeService{
		baseService: newBase[domainFacility.AlarmType](repo, 20),
		extRepo:     repo,
	}
}

func (s *AlarmTypeService) Create(at *domainFacility.AlarmType) error {
	return s.extRepo.Create(at)
}

func (s *AlarmTypeService) Update(at *domainFacility.AlarmType) error {
	return s.extRepo.Update(at)
}

func (s *AlarmTypeService) GetWithFields(id uuid.UUID) (*domainFacility.AlarmType, error) {
	return s.extRepo.GetWithFields(id)
}
