package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

// SPSControllerNameSynchronizer keeps generated controller names in sync when
// a building or control-cabinet input changes. Direct controller writes derive
// their name in SPSControllerService; this type handles the parent cascades.
type SPSControllerNameSynchronizer struct {
	buildings   domainFacility.BuildingRepository
	cabinets    domainFacility.ControlCabinetRepository
	controllers domainFacility.SPSControllerRepository
}

const spsControllerNameSyncPageLimit = 500

func NewSPSControllerNameSynchronizer(
	buildings domainFacility.BuildingRepository,
	cabinets domainFacility.ControlCabinetRepository,
	controllers domainFacility.SPSControllerRepository,
) *SPSControllerNameSynchronizer {
	return &SPSControllerNameSynchronizer{
		buildings: buildings, cabinets: cabinets, controllers: controllers,
	}
}

func (s *SPSControllerNameSynchronizer) RefreshForBuilding(
	ctx context.Context,
	building *domainFacility.Building,
) error {
	if s == nil || building == nil {
		return nil
	}

	for page := 1; ; page++ {
		result, err := s.cabinets.GetPaginatedListByBuildingID(ctx, building.ID, domain.PaginationParams{
			Page: page, Limit: spsControllerNameSyncPageLimit,
		})
		if err != nil {
			return err
		}

		for i := range result.Items {
			if err := s.refreshForCabinetWithBuilding(ctx, &result.Items[i], building); err != nil {
				return err
			}
		}

		if page >= result.TotalPages || len(result.Items) == 0 {
			return nil
		}
	}
}

func (s *SPSControllerNameSynchronizer) RefreshForControlCabinet(
	ctx context.Context,
	cabinet *domainFacility.ControlCabinet,
) error {
	if s == nil || cabinet == nil {
		return nil
	}

	building, err := domain.GetByID(ctx, s.buildings, cabinet.BuildingID)
	if err != nil {
		return err
	}
	return s.refreshForCabinetWithBuilding(ctx, cabinet, building)
}

func (s *SPSControllerNameSynchronizer) refreshForCabinetWithBuilding(
	ctx context.Context,
	cabinet *domainFacility.ControlCabinet,
	building *domainFacility.Building,
) error {
	for page := 1; ; page++ {
		result, err := s.controllers.GetPaginatedListByControlCabinetID(ctx, cabinet.ID, domain.PaginationParams{
			Page: page, Limit: spsControllerNameSyncPageLimit,
		})
		if err != nil {
			return err
		}

		for i := range result.Items {
			controller := &result.Items[i]
			name, ok := generatedSPSControllerDeviceName(cabinet, building, controller.GADevice)
			if !ok || controller.DeviceName == name {
				continue
			}
			controller.DeviceName = name
			if err := s.controllers.Update(ctx, controller); err != nil {
				return err
			}
		}

		if page >= result.TotalPages || len(result.Items) == 0 {
			return nil
		}
	}
}
