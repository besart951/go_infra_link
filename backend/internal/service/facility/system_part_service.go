package facility

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type SystemPartService struct {
	baseService[domainFacility.SystemPart]
	extRepo domainFacility.SystemPartRepository
}

func NewSystemPartService(repo domainFacility.SystemPartRepository) *SystemPartService {
	return &SystemPartService{
		baseService: newBase(repo, 10),
		extRepo:     repo,
	}
}

func (s *SystemPartService) Create(ctx context.Context, systemPart *domainFacility.SystemPart) error {
	if err := s.Validate(ctx, systemPart, nil); err != nil {
		return err
	}
	return s.repo.Create(ctx, systemPart)
}

func (s *SystemPartService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	return s.extRepo.GetByIds(ctx, ids)
}

func (s *SystemPartService) GetApparatIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	systemPart, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return extractIDs(systemPart.Apparats, func(a *domainFacility.Apparat) uuid.UUID { return a.ID }), nil
}

func (s *SystemPartService) Update(ctx context.Context, systemPart *domainFacility.SystemPart) error {
	if err := s.Validate(ctx, systemPart, &systemPart.ID); err != nil {
		return err
	}
	return s.repo.Update(ctx, systemPart)
}

func (s *SystemPartService) Validate(ctx context.Context, systemPart *domainFacility.SystemPart, excludeID *uuid.UUID) error {
	if err := s.validateRequiredFields(systemPart); err != nil {
		return err
	}
	return s.ensureUnique(ctx, systemPart, excludeID)
}

func (s *SystemPartService) validateRequiredFields(systemPart *domainFacility.SystemPart) error {
	return validateRules(
		requiredTrimmedExact(systemPartShortNameField, systemPart.ShortName, 3),
		requiredTrimmed(systemPartNameField, systemPart.Name),
	)
}

func (s *SystemPartService) ensureUnique(ctx context.Context, systemPart *domainFacility.SystemPart, excludeID *uuid.UUID) error {
	return validateChecks(
		uniqueIfPresent(systemPartShortNameField, systemPart.ShortName, func() (bool, error) {
			return s.extRepo.ExistsShortName(ctx, systemPart.ShortName, excludeID)
		}),
		uniqueIfPresent(systemPartNameField, systemPart.Name, func() (bool, error) {
			return s.extRepo.ExistsName(ctx, systemPart.Name, excludeID)
		}),
	)
}
