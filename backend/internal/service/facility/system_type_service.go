package facility

import (
	"context"

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
		baseService: newBase(repo, 10),
		extRepo:     repo,
	}
}

func (s *SystemTypeService) Create(ctx context.Context, systemType *domainFacility.SystemType) error {
	if err := s.Validate(ctx, systemType, nil); err != nil {
		return err
	}
	return s.extRepo.Create(ctx, systemType)
}

func (s *SystemTypeService) Update(ctx context.Context, systemType *domainFacility.SystemType) error {
	if err := s.Validate(ctx, systemType, &systemType.ID); err != nil {
		return err
	}
	return s.extRepo.Update(ctx, systemType)
}

func (s *SystemTypeService) Validate(ctx context.Context, systemType *domainFacility.SystemType, excludeID *uuid.UUID) error {
	if err := s.validateRequiredFields(systemType); err != nil {
		return err
	}
	return s.ensureUnique(ctx, systemType, excludeID)
}

func (s *SystemTypeService) validateRequiredFields(systemType *domainFacility.SystemType) error {
	return validateRules(
		requiredTrimmedMax(systemTypeNameField, systemType.Name, 150),
		addIf(systemTypeNumberMaxField, systemType.NumberMin > systemType.NumberMax, "number_max must be greater than or equal to number_min"),
	)
}

func (s *SystemTypeService) ensureUnique(ctx context.Context, systemType *domainFacility.SystemType, excludeID *uuid.UUID) error {
	return validateChecks(
		uniqueIfPresent(systemTypeNameField, systemType.Name, func() (bool, error) {
			return s.extRepo.ExistsName(ctx, systemType.Name, excludeID)
		}),
		func(builder *domain.ValidationBuilder) error {
			exists, err := s.extRepo.ExistsOverlappingRange(ctx, systemType.NumberMin, systemType.NumberMax, excludeID)
			if err != nil {
				return err
			}
			if exists {
				systemTypeNumberMinField.Add(builder, "number_min and number_max range must not overlap existing ranges")
			}
			return nil
		},
	)
}
