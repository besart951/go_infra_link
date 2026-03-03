package facility

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type SystemTypeService struct {
	baseService[domainFacility.SystemType]
	extRepo domainFacility.SystemTypeRepository
}

func NewSystemTypeService(repo domainFacility.SystemTypeRepository) *SystemTypeService {
	return &SystemTypeService{
		baseService: newBase[domainFacility.SystemType](repo, 10),
		extRepo:     repo,
	}
}

func (s *SystemTypeService) Create(systemType *domainFacility.SystemType) error {
	if err := s.Validate(systemType, nil); err != nil {
		return err
	}
	return s.extRepo.Create(systemType)
}

func (s *SystemTypeService) Update(systemType *domainFacility.SystemType) error {
	if err := s.Validate(systemType, &systemType.ID); err != nil {
		return err
	}
	return s.extRepo.Update(systemType)
}

func (s *SystemTypeService) Validate(systemType *domainFacility.SystemType, excludeID *uuid.UUID) error {
	if err := s.validateRequiredFields(systemType); err != nil {
		return err
	}
	return s.ensureUnique(systemType, excludeID)
}

func (s *SystemTypeService) validateRequiredFields(systemType *domainFacility.SystemType) error {
	ve := domain.NewValidationError()
	name := strings.TrimSpace(systemType.Name)
	if name == "" {
		ve = ve.Add("systemtype.name", "name is required")
	} else if len(name) > 150 {
		ve = ve.Add("systemtype.name", "name must be 150 characters or less")
	}
	if systemType.NumberMin > systemType.NumberMax {
		ve = ve.Add("systemtype.number_max", "number_max must be greater than or equal to number_min")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func (s *SystemTypeService) ensureUnique(systemType *domainFacility.SystemType, excludeID *uuid.UUID) error {
	name := strings.TrimSpace(systemType.Name)
	if name != "" {
		exists, err := s.extRepo.ExistsName(name, excludeID)
		if err != nil {
			return err
		}
		if exists {
			return domain.NewValidationError().Add("systemtype.name", "name must be unique")
		}
	}
	exists, err := s.extRepo.ExistsOverlappingRange(systemType.NumberMin, systemType.NumberMax, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return domain.NewValidationError().Add(
			"systemtype.number_min",
			"number_min and number_max range must not overlap existing ranges",
		)
	}
	return nil
}
