package exporting

import (
	"context"
	"fmt"
	"math"
	"strings"
	"unicode"

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

	controllers, err := p.spsControllers.GetByIdsForExport(ids)
	if err != nil {
		return nil, err
	}

	out := make([]domainExport.Controller, 0, len(controllers))
	for _, c := range controllers {
		out = append(out, buildExportController(c))
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

func buildExportController(c domainFacility.SPSController) domainExport.Controller {
	ga := derefStr(c.GADevice)
	building := c.ControlCabinet.Building
	minSysPart := minSystemPartNumber(c.SPSControllerSystemTypes)
	bgStr := fmt.Sprintf("%d", building.BuildingGroup)

	return domainExport.Controller{
		ID:               c.ID,
		ControlCabinetID: c.ControlCabinetID,
		GADevice:         ga,

		IWSCode:             building.IWSCode,
		BuildingGroup:       building.BuildingGroup,
		ControlCabinetNr:    derefStr(c.ControlCabinet.ControlCabinetNr),
		MinSystemPartNumber: minSysPart,
		DeviceName:          strings.Join(filterEmpty([]string{building.IWSCode, bgStr, minSysPart, ga}), "_"),
		DeviceInstance:      lastTwoIWSCode(building.IWSCode) + convertGADeviceToIndex(ga) + bgStr,
		DeviceDescription:   derefStr(c.DeviceDescription),
		DeviceLocation:      derefStr(c.DeviceLocation),
		IPAddress:           derefStr(c.IPAddress),
		Subnet:              derefStr(c.Subnet),
		Gateway:             derefStr(c.Gateway),
		VLAN:                derefStr(c.Vlan),
	}
}

func minSystemPartNumber(systemTypes []domainFacility.SPSControllerSystemType) string {
	lowest := math.MaxInt
	for _, st := range systemTypes {
		if st.Number != nil && *st.Number < lowest {
			lowest = *st.Number
		}
	}
	if lowest == math.MaxInt {
		lowest = 0
	}
	return fmt.Sprintf("%04d", lowest)
}

func convertGADeviceToIndex(gaDevice string) string {
	if gaDevice == "" {
		return "00"
	}
	ch := unicode.ToUpper(rune(gaDevice[0]))
	if ch < 'A' || ch > 'Z' {
		return "00"
	}
	return fmt.Sprintf("%02d", ch-'A')
}

func lastTwoIWSCode(iwsCode string) string {
	if len(iwsCode) < 2 {
		return iwsCode
	}
	return iwsCode[len(iwsCode)-2:]
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func filterEmpty(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
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
