package exporting

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type DataProvider struct {
	fieldDevices    domainFacility.FieldDeviceStore
	spsControllers  domainFacility.SPSControllerRepository
	controlCabinets domainFacility.ControlCabinetRepository
}

func NewDataProvider(
	fieldDevices domainFacility.FieldDeviceStore,
	spsControllers domainFacility.SPSControllerRepository,
	controlCabinets domainFacility.ControlCabinetRepository,
) *DataProvider {
	return &DataProvider{
		fieldDevices:    fieldDevices,
		spsControllers:  spsControllers,
		controlCabinets: controlCabinets,
	}
}

func (p *DataProvider) ResolveControllers(ctx context.Context, req domainExport.Request) ([]domainExport.Controller, error) {
	_ = ctx

	controllerSet := map[uuid.UUID]struct{}{}

	for _, id := range req.SPSControllerIDs {
		controllerSet[id] = struct{}{}
	}

	cabinetIDs := append([]uuid.UUID{}, req.ControlCabinetIDs...)
	for _, buildingID := range req.BuildingIDs {
		ids, err := p.controlCabinets.GetIDsByBuildingID(buildingID)
		if err != nil {
			return nil, err
		}
		cabinetIDs = append(cabinetIDs, ids...)
	}

	if len(cabinetIDs) > 0 {
		ids, err := p.spsControllers.GetIDsByControlCabinetIDs(uniqueUUIDs(cabinetIDs))
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			controllerSet[id] = struct{}{}
		}
	}

	if len(controllerSet) == 0 {
		page := 1
		for {
			list, err := p.spsControllers.GetPaginatedList(domain.PaginationParams{Page: page, Limit: 1000})
			if err != nil {
				return nil, err
			}
			for _, item := range list.Items {
				controllerSet[item.ID] = struct{}{}
			}
			if page >= list.TotalPages {
				break
			}
			page++
		}
	}

	ids := make([]uuid.UUID, 0, len(controllerSet))
	for id := range controllerSet {
		ids = append(ids, id)
	}

	controllers, err := p.spsControllers.GetByIds(ids)
	if err != nil {
		return nil, err
	}

	out := make([]domainExport.Controller, 0, len(controllers))
	for _, c := range controllers {
		ga := ""
		if c.GADevice != nil {
			ga = *c.GADevice
		}
		out = append(out, domainExport.Controller{
			ID:               c.ID,
			ControlCabinetID: c.ControlCabinetID,
			GADevice:         ga,
		})
	}

	return out, nil
}

func (p *DataProvider) ListFieldDevicesByController(ctx context.Context, controllerID uuid.UUID, req domainExport.Request, page, limit int) ([]domainFacility.FieldDevice, int64, error) {
	_ = ctx

	params := domain.PaginationParams{Page: page, Limit: limit}
	filters := domainFacility.FieldDeviceFilterParams{SPSControllerID: &controllerID}

	if len(req.ProjectIDs) > 0 {
		filters.ProjectIDs = req.ProjectIDs
	}

	result, err := p.fieldDevices.GetPaginatedListWithFilters(params, filters)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, result.Total, nil
}

func uniqueUUIDs(ids []uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	out := make([]uuid.UUID, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	return out
}
