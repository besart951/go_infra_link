package facilityservice

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/core/domain/facility"
	"github.com/google/uuid"
)

type CabinetService struct {
	repo facility.CabinetRepository
}

func NewCabinetService(repo facility.CabinetRepository) *CabinetService {
	return &CabinetService{repo: repo}
}

func (s *CabinetService) AddCabinetToBuilding(ctx context.Context, buildingID uuid.UUID, nr string) error {
	cab := &facility.ControlCabinet{
		BuildingID:       buildingID,
		ControlCabinetNr: &nr,
	}

	return s.repo.Create(ctx, cab)
}

func (s *CabinetService) ListCabinets(ctx context.Context, buildingID uuid.UUID) ([]facility.ControlCabinet, error) {
	return s.repo.FindAllByBuildingID(ctx, buildingID)
}
