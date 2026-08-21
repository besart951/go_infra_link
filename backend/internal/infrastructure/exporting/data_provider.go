package exporting

import (
	"context"
	"errors"
	"fmt"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	"math"
	"strings"
	"unicode"

	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type DataProvider struct {
	fieldDevices    domainFieldDevice.FieldDeviceStore
	exportReader    fieldDeviceExportReader
	specifications  domainFieldDevice.SpecificationStore
	bacnetObjects   domainObjectData.BacnetObjectStore
	spsControllers  domainFacility.SPSControllerRepository
	controlCabinets domainFacility.ControlCabinetRepository
	alarmValues     alarmValueExportReader
	snapshotRunner  func(context.Context, func(domainExport.DataProvider) error) error
}

type fieldDeviceExportReader interface {
	GetExportPage(context.Context, domainFacility.FieldDeviceFilterParams, uuid.UUID, int) ([]domainFacility.FieldDevice, error)
	GetExportControllerIDs(context.Context, domainFacility.FieldDeviceFilterParams, string) ([]uuid.UUID, error)
}

type alarmValueExportReader interface {
	GetByBacnetObjectIDs(context.Context, []uuid.UUID) ([]domainFacility.BacnetObjectAlarmValue, error)
}

func (p *DataProvider) SetSnapshotRunner(runner func(context.Context, func(domainExport.DataProvider) error) error) {
	p.snapshotRunner = runner
}

func (p *DataProvider) WithinSnapshot(ctx context.Context, consume func(domainExport.DataProvider) error) error {
	if p.snapshotRunner == nil {
		return consume(p)
	}
	return p.snapshotRunner(ctx, consume)
}

func NewDataProvider(
	fieldDevices domainFieldDevice.FieldDeviceStore,
	specifications domainFieldDevice.SpecificationStore,
	bacnetObjects domainObjectData.BacnetObjectStore,
	spsControllers domainFacility.SPSControllerRepository,
	controlCabinets domainFacility.ControlCabinetRepository,
	alarmValues ...domainFacility.BacnetObjectAlarmValueRepository,
) *DataProvider {
	provider := &DataProvider{
		fieldDevices:    fieldDevices,
		specifications:  specifications,
		bacnetObjects:   bacnetObjects,
		spsControllers:  spsControllers,
		controlCabinets: controlCabinets,
	}
	provider.exportReader, _ = fieldDevices.(fieldDeviceExportReader)
	if len(alarmValues) > 0 {
		provider.alarmValues, _ = alarmValues[0].(alarmValueExportReader)
	}
	return provider
}

func (p *DataProvider) ResolveControllers(ctx context.Context, req domainExport.Request) ([]domainExport.Controller, error) {
	filters := domainFacility.FieldDeviceFilterParams{
		BuildingIDs: req.BuildingIDs, ControlCabinetIDs: req.ControlCabinetIDs,
		SPSControllerIDs: req.SPSControllerIDs, SPSControllerSystemTypeIDs: req.SPSControllerSystemTypeIDs,
		ProjectIDs: req.ProjectIDs,
	}
	if p.exportReader == nil {
		return nil, errors.New("field-device export reader is unavailable")
	}
	ids, err := p.exportReader.GetExportControllerIDs(ctx, filters, req.Search)
	if err != nil {
		return nil, err
	}

	controllers, err := p.spsControllers.GetByIdsForExport(ctx, ids)
	if err != nil {
		return nil, err
	}

	out := make([]domainExport.Controller, 0, len(controllers))
	for _, c := range controllers {
		out = append(out, buildExportController(c))
	}

	return out, nil
}

func (p *DataProvider) ListFieldDevicesByControllerAfter(ctx context.Context, controllerID uuid.UUID, req domainExport.Request, afterID uuid.UUID, limit int) ([]domainFacility.FieldDevice, error) {
	filters := domainFacility.FieldDeviceFilterParams{
		Search: req.Search, SPSControllerID: &controllerID,
		BuildingIDs: req.BuildingIDs, ControlCabinetIDs: req.ControlCabinetIDs,
		SPSControllerSystemTypeIDs: req.SPSControllerSystemTypeIDs,
	}

	if len(req.ProjectIDs) > 0 {
		filters.ProjectIDs = req.ProjectIDs
	}

	if p.exportReader == nil {
		return nil, errors.New("field-device export reader is unavailable")
	}
	items, err := p.exportReader.GetExportPage(ctx, filters, afterID, limit)
	if err != nil {
		return nil, err
	}

	items, err = p.hydrateFieldDevicesForExport(ctx, items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (p *DataProvider) hydrateFieldDevicesForExport(ctx context.Context, items []domainFacility.FieldDevice) ([]domainFacility.FieldDevice, error) {
	if len(items) == 0 {
		return items, nil
	}

	fieldDeviceIDs := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		fieldDeviceIDs = append(fieldDeviceIDs, item.ID)
	}

	specifications, err := p.specifications.GetByFieldDeviceIDs(ctx, fieldDeviceIDs)
	if err != nil {
		return nil, err
	}
	specificationsByFieldDeviceID := make(map[uuid.UUID]*domainFacility.Specification, len(specifications))
	for _, specification := range specifications {
		if specification == nil || specification.FieldDeviceID == nil {
			continue
		}
		specificationsByFieldDeviceID[*specification.FieldDeviceID] = specification
	}

	bacnetObjects, err := p.bacnetObjects.GetByFieldDeviceIDs(ctx, fieldDeviceIDs)
	if err != nil {
		return nil, err
	}
	bacnetObjectsByFieldDeviceID := make(map[uuid.UUID][]domainFacility.BacnetObject, len(items))
	bacnetObjectIDs := make([]uuid.UUID, 0, len(bacnetObjects))
	for _, bacnetObject := range bacnetObjects {
		if bacnetObject == nil || bacnetObject.FieldDeviceID == nil {
			continue
		}
		bacnetObjectIDs = append(bacnetObjectIDs, bacnetObject.ID)
		bacnetObjectsByFieldDeviceID[*bacnetObject.FieldDeviceID] = append(
			bacnetObjectsByFieldDeviceID[*bacnetObject.FieldDeviceID],
			*bacnetObject,
		)
	}
	if p.alarmValues != nil && len(bacnetObjectIDs) > 0 {
		values, err := p.alarmValues.GetByBacnetObjectIDs(ctx, bacnetObjectIDs)
		if err != nil {
			return nil, err
		}
		valuesByObject := make(map[uuid.UUID][]domainFacility.BacnetObjectAlarmValue)
		for _, value := range values {
			valuesByObject[value.BacnetObjectID] = append(valuesByObject[value.BacnetObjectID], value)
		}
		for fieldDeviceID, objects := range bacnetObjectsByFieldDeviceID {
			for i := range objects {
				objects[i].AlarmValues = valuesByObject[objects[i].ID]
			}
			bacnetObjectsByFieldDeviceID[fieldDeviceID] = objects
		}
	}

	for i := range items {
		item := &items[i]
		if specification, ok := specificationsByFieldDeviceID[item.ID]; ok {
			item.Specification = specification
			item.SpecificationID = &specification.ID
		}
		item.BacnetObjects = bacnetObjectsByFieldDeviceID[item.ID]
	}

	return items, nil
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
