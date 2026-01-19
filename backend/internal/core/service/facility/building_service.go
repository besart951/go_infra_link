package facilityservice

import (
	"context"

	"github.com/google/uuid"
)

type BuildingService struct {
	repo obj	repo facility.BuildingRepository
}

func NewBuildingService(repo facility.BuildingRepository) *BuildingService {
	return &BuildingService{repo: repo}
}

func (s *BuildingService) CreateBuilding(ctx context.Context, iws_code string, building_group int) error {
	newBuilding := &facility.Building{
		ID:            uuid.NewV7(),
		IwsCode:       iws_code,
		BuildingGroup: building_group,
	}
	return s.repo.Create(ctx, newBuilding)
}

func (s *BuildingService) GetBuildig(ctx context.Context, id uuid.UUID) (*facility.Building, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BuildingService) GetPaginatedBuilding(ctx context.Context) ([]facility.Building, error) {
x)
}
