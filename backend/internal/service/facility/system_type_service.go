package facility

import (
	"context"
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
	builder := domain.NewValidationBuilder()
	name := strings.TrimSpace(systemType.Name)
	if name == "" {
		systemTypeNameField.Add(builder, "name is required")
	} else {
		systemTypeNameField.MaxLength(builder, name, 150)
	}
	if systemType.NumberMin > systemType.NumberMax {
		systemTypeNumberMaxField.Add(builder, "number_max must be greater than or equal to number_min")
	}
	return builder.Err()
}

func (s *SystemTypeService) ensureUnique(ctx context.Context, systemType *domainFacility.SystemType, excludeID *uuid.UUID) error {
	name := strings.TrimSpace(systemType.Name)
	if name != "" {
		exists, err := s.extRepo.ExistsName(ctx, name, excludeID)
		if err != nil {
			return err
		}
		if exists {
			return systemTypeNameField.UniqueError()
		}
	}
	exists, err := s.extRepo.ExistsOverlappingRange(ctx, systemType.NumberMin, systemType.NumberMax, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return domain.NewValidationError().Add(
			systemTypeNumberMinField.Key,
			"number_min and number_max range must not overlap existing ranges",
		)
	}
	return nil
}
