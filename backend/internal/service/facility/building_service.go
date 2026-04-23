package facility

import (
	"context"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type BuildingService struct {
	repo domainFacility.BuildingRepository
}

func NewBuildingService(repo domainFacility.BuildingRepository) *BuildingService {
	return &BuildingService{
		repo: repo,
	}
}

func (s *BuildingService) Create(ctx context.Context, building *domainFacility.Building) error {
	if err := s.Validate(ctx, building, nil); err != nil {
		return err
	}
	return s.repo.Create(ctx, building)
}

func (s *BuildingService) GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.Building, error) {
	return domain.GetByID(ctx, s.repo, id)
}

func (s *BuildingService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domainFacility.Building, error) {
	buildings, err := s.repo.GetByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	return derefSlice(buildings), nil
}

func (s *BuildingService) List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.Building], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(ctx, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *BuildingService) Update(ctx context.Context, building *domainFacility.Building) error {
	if err := s.Validate(ctx, building, &building.ID); err != nil {
		return err
	}
	return s.repo.Update(ctx, building)
}

func (s *BuildingService) Validate(ctx context.Context, building *domainFacility.Building, excludeID *uuid.UUID) error {
	if err := s.validateRequiredFields(building); err != nil {
		return err
	}
	if err := s.ensureUnique(ctx, building, excludeID); err != nil {
		return err
	}
	return nil
}

func (s *BuildingService) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteByIds(ctx, []uuid.UUID{id})
}

func (s *BuildingService) validateRequiredFields(building *domainFacility.Building) error {
	builder := domain.NewValidationBuilder()
	iwsCode := buildingIWSCodeField.RequireTrimmed(builder, building.IWSCode)
	buildingIWSCodeField.ExactLength(builder, iwsCode, 4)
	buildingGroupField.RequireNonZero(builder, building.BuildingGroup)
	return builder.Err()
}

func (s *BuildingService) ensureUnique(ctx context.Context, building *domainFacility.Building, excludeID *uuid.UUID) error {
	if strings.TrimSpace(building.IWSCode) == "" || building.BuildingGroup == 0 {
		return nil
	}
	exists, err := s.repo.ExistsIWSCodeGroup(ctx, building.IWSCode, building.BuildingGroup, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return buildingIWSCodeField.UniqueWithinError(buildingGroupScope)
	}
	return nil
}
