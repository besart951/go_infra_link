package facility

import (
	"context"
	"errors"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ApparatService struct {
	baseService[domainFacility.Apparat]
	extRepo domainFacility.ApparatRepository
}

func NewApparatService(repo domainFacility.ApparatRepository) *ApparatService {
	return &ApparatService{
		baseService: newBase[domainFacility.Apparat](repo, 10),
		extRepo:     repo,
	}
}

func (s *ApparatService) Create(ctx context.Context, apparat *domainFacility.Apparat) error {
	if err := s.Validate(ctx, apparat, nil); err != nil {
		return err
	}
	if err := s.repo.Create(ctx, apparat); err != nil {
		return s.mapWriteConflict(ctx, apparat, nil, err)
	}
	return nil
}

func (s *ApparatService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	return s.extRepo.GetByIds(ctx, ids)
}

func (s *ApparatService) Update(ctx context.Context, apparat *domainFacility.Apparat) error {
	if err := s.Validate(ctx, apparat, &apparat.ID); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, apparat); err != nil {
		return s.mapWriteConflict(ctx, apparat, &apparat.ID, err)
	}
	return nil
}

func (s *ApparatService) GetSystemPartIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	apparat, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return extractIDs(apparat.SystemParts, func(sp *domainFacility.SystemPart) uuid.UUID { return sp.ID }), nil
}

func (s *ApparatService) Validate(ctx context.Context, apparat *domainFacility.Apparat, excludeID *uuid.UUID) error {
	if err := s.validateRequiredFields(apparat); err != nil {
		return err
	}
	return s.ensureUnique(ctx, apparat, excludeID)
}

func (s *ApparatService) validateRequiredFields(apparat *domainFacility.Apparat) error {
	builder := domain.NewValidationBuilder()
	shortName := apparatShortNameField.RequireTrimmed(builder, apparat.ShortName)
	apparatShortNameField.ExactLength(builder, shortName, 3)
	apparatNameField.RequireTrimmed(builder, apparat.Name)
	return builder.Err()
}

func (s *ApparatService) ensureUnique(ctx context.Context, apparat *domainFacility.Apparat, excludeID *uuid.UUID) error {
	ve, err := s.uniqueValidationError(ctx, apparat, excludeID)
	if err != nil {
		return err
	}
	if ve != nil {
		return ve
	}
	return nil
}

func (s *ApparatService) uniqueValidationError(ctx context.Context, apparat *domainFacility.Apparat, excludeID *uuid.UUID) (*domain.ValidationError, error) {
	builder := domain.NewValidationBuilder()
	if strings.TrimSpace(apparat.ShortName) != "" {
		exists, err := s.extRepo.ExistsShortName(ctx, apparat.ShortName, excludeID)
		if err != nil {
			return nil, err
		}
		if exists {
			apparatShortNameField.Unique(builder)
		}
	}
	if strings.TrimSpace(apparat.Name) != "" {
		exists, err := s.extRepo.ExistsName(ctx, apparat.Name, excludeID)
		if err != nil {
			return nil, err
		}
		if exists {
			apparatNameField.Unique(builder)
		}
	}
	return builder.ValidationError(), nil
}

func (s *ApparatService) mapWriteConflict(ctx context.Context, apparat *domainFacility.Apparat, excludeID *uuid.UUID, err error) error {
	if !errors.Is(err, domain.ErrConflict) {
		return err
	}

	ve, checkErr := s.uniqueValidationError(ctx, apparat, excludeID)
	if checkErr != nil {
		return checkErr
	}
	if ve != nil {
		return ve
	}

	return err
}
