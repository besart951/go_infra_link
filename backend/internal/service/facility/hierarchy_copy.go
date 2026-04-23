package facility

import (
	"context"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// HierarchyCopier encapsulates the reusable deep-copy logic for the
// control-cabinet -> SPS controller -> system type -> field-device hierarchy.
type HierarchyCopier struct {
	controlCabinetRepo      domainFacility.ControlCabinetRepository
	buildingRepo            domainFacility.BuildingRepository
	spsControllerRepo       domainFacility.SPSControllerRepository
	systemTypeRepo          domainFacility.SystemTypeRepository
	spsControllerSystemRepo domainFacility.SPSControllerSystemTypeStore
	fieldDeviceRepo         domainFacility.FieldDeviceStore
	specificationRepo       domainFacility.SpecificationStore
	bacnetObjectRepo        domainFacility.BacnetObjectStore
	tx                      txCoordinator
}

func NewHierarchyCopier(
	controlCabinetRepo domainFacility.ControlCabinetRepository,
	buildingRepo domainFacility.BuildingRepository,
	spsControllerRepo domainFacility.SPSControllerRepository,
	systemTypeRepo domainFacility.SystemTypeRepository,
	spsControllerSystemRepo domainFacility.SPSControllerSystemTypeStore,
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	specificationRepo domainFacility.SpecificationStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
) *HierarchyCopier {
	return &HierarchyCopier{
		controlCabinetRepo:      controlCabinetRepo,
		buildingRepo:            buildingRepo,
		spsControllerRepo:       spsControllerRepo,
		systemTypeRepo:          systemTypeRepo,
		spsControllerSystemRepo: spsControllerSystemRepo,
		fieldDeviceRepo:         fieldDeviceRepo,
		specificationRepo:       specificationRepo,
		bacnetObjectRepo:        bacnetObjectRepo,
	}
}

func (c *HierarchyCopier) bindTransactions(tx txCoordinator) {
	c.tx = tx
}

func (c *HierarchyCopier) CopyControlCabinetByID(ctx context.Context, id uuid.UUID) (*domainFacility.ControlCabinet, error) {
	return runWithFacilityTxResult(c.tx, c, func(services *Services) *HierarchyCopier {
		return services.HierarchyCopier
	}, func(copier *HierarchyCopier) (*domainFacility.ControlCabinet, error) {
		original, err := domain.GetByID(ctx, copier.controlCabinetRepo, id)
		if err != nil {
			return nil, err
		}

		baseNr := ""
		if original.ControlCabinetNr != nil {
			baseNr = strings.TrimSpace(*original.ControlCabinetNr)
		}

		nextNr, err := copier.nextAvailableControlCabinetNr(ctx, original.BuildingID, baseNr)
		if err != nil {
			return nil, err
		}

		copyEntity := &domainFacility.ControlCabinet{
			BuildingID:       original.BuildingID,
			ControlCabinetNr: &nextNr,
		}
		if err := copier.controlCabinetRepo.Create(ctx, copyEntity); err != nil {
			return nil, err
		}

		if err := copier.copySPSControllersForControlCabinet(ctx, original.ID, copyEntity.ID); err != nil {
			return nil, err
		}

		return copyEntity, nil
	})
}

func (c *HierarchyCopier) CopySPSControllerByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSController, error) {
	return runWithFacilityTxResult(c.tx, c, func(services *Services) *HierarchyCopier {
		return services.HierarchyCopier
	}, func(copier *HierarchyCopier) (*domainFacility.SPSController, error) {
		original, err := domain.GetByID(ctx, copier.spsControllerRepo, id)
		if err != nil {
			return nil, err
		}

		controlCabinet, err := domain.GetByID(ctx, copier.controlCabinetRepo, original.ControlCabinetID)
		if err != nil {
			return nil, err
		}

		building, err := domain.GetByID(ctx, copier.buildingRepo, controlCabinet.BuildingID)
		if err != nil {
			return nil, err
		}

		nextGADevice, err := copier.nextAvailableGADevice(ctx, original.ControlCabinetID)
		if err != nil {
			return nil, err
		}
		if nextGADevice == "" {
			return nil, domain.NewValidationError().Add("spscontroller.ga_device", "no available ga_device for control cabinet")
		}

		iwsCode := strings.TrimSpace(building.IWSCode)
		cabinetNr := ""
		if controlCabinet.ControlCabinetNr != nil {
			cabinetNr = strings.TrimSpace(*controlCabinet.ControlCabinetNr)
		}

		deviceName := nextGADevice
		if iwsCode != "" && cabinetNr != "" {
			deviceName = strings.ToUpper(iwsCode + "_" + cabinetNr + "_" + nextGADevice)
		}

		copyEntity := &domainFacility.SPSController{
			ControlCabinetID:  original.ControlCabinetID,
			GADevice:          &nextGADevice,
			DeviceName:        deviceName,
			DeviceDescription: original.DeviceDescription,
			DeviceLocation:    original.DeviceLocation,
			IPAddress:         nil,
			Subnet:            original.Subnet,
			Gateway:           original.Gateway,
			Vlan:              original.Vlan,
		}
		if err := copier.spsControllerRepo.Create(ctx, copyEntity); err != nil {
			return nil, err
		}

		if err := copier.copySystemTypesAndFieldDevicesForSPSController(ctx, original.ID, copyEntity.ID); err != nil {
			return nil, err
		}

		return copyEntity, nil
	})
}

func (c *HierarchyCopier) CopySPSControllerSystemTypeByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	return runWithFacilityTxResult(c.tx, c, func(services *Services) *HierarchyCopier {
		return services.HierarchyCopier
	}, func(copier *HierarchyCopier) (*domainFacility.SPSControllerSystemType, error) {
		original, err := domain.GetByID(ctx, copier.spsControllerSystemRepo, id)
		if err != nil {
			return nil, err
		}

		systemType, err := domain.GetByID(ctx, copier.systemTypeRepo, original.SystemTypeID)
		if err != nil {
			return nil, err
		}

		existing, err := copier.spsControllerSystemRepo.ListBySPSControllerID(ctx, original.SPSControllerID)
		if err != nil {
			return nil, err
		}

		usedNumbers := make(map[int]struct{}, len(existing))
		for _, item := range existing {
			if item.SystemTypeID != original.SystemTypeID || item.Number == nil {
				continue
			}
			usedNumbers[*item.Number] = struct{}{}
		}

		nextNumber, ok := findLowestAvailableNumber(systemType.NumberMin, systemType.NumberMax, usedNumbers)
		if !ok {
			return nil, domain.NewValidationError().Add("spscontroller.system_types", "no available number in the system type range")
		}

		copyNumber := nextNumber
		copyEntity := &domainFacility.SPSControllerSystemType{
			Number:          &copyNumber,
			DocumentName:    original.DocumentName,
			SPSControllerID: original.SPSControllerID,
			SystemTypeID:    original.SystemTypeID,
		}
		if err := copier.spsControllerSystemRepo.Create(ctx, copyEntity); err != nil {
			return nil, err
		}

		if err := copier.copyFieldDevicesForSystemTypes(ctx, map[uuid.UUID]uuid.UUID{original.ID: copyEntity.ID}); err != nil {
			return nil, err
		}

		return domain.GetByID(ctx, copier.spsControllerSystemRepo, copyEntity.ID)
	})
}

func (c *HierarchyCopier) copySPSControllersForControlCabinet(ctx context.Context, originalControlCabinetID, newControlCabinetID uuid.UUID) error {
	newControlCabinet, err := domain.GetByID(ctx, c.controlCabinetRepo, newControlCabinetID)
	if err != nil {
		return err
	}

	building, err := domain.GetByID(ctx, c.buildingRepo, newControlCabinet.BuildingID)
	if err != nil {
		return err
	}

	newCabinetNr := ""
	if newControlCabinet.ControlCabinetNr != nil {
		newCabinetNr = strings.TrimSpace(*newControlCabinet.ControlCabinetNr)
	}
	buildingIWSCode := strings.TrimSpace(building.IWSCode)

	originalSPSControllers, err := c.listSPSControllersByControlCabinetID(ctx, originalControlCabinetID)
	if err != nil {
		return err
	}

	for _, originalSPS := range originalSPSControllers {
		gaDevice := ""
		if originalSPS.GADevice != nil {
			gaDevice = strings.ToUpper(strings.TrimSpace(*originalSPS.GADevice))
		}
		if gaDevice == "" {
			nextGADevice, err := c.nextAvailableGADevice(ctx, newControlCabinetID)
			if err != nil {
				return err
			}
			gaDevice = nextGADevice
		}

		var gaDevicePtr *string
		if gaDevice != "" {
			gaDevicePtr = &gaDevice
		}

		deviceName := strings.TrimSpace(originalSPS.DeviceName)
		if buildingIWSCode != "" && newCabinetNr != "" && gaDevice != "" {
			deviceName = strings.ToUpper(buildingIWSCode + "_" + newCabinetNr + "_" + gaDevice)
		}

		spsCopy := &domainFacility.SPSController{
			ControlCabinetID:  newControlCabinetID,
			GADevice:          gaDevicePtr,
			DeviceName:        deviceName,
			DeviceDescription: originalSPS.DeviceDescription,
			DeviceLocation:    originalSPS.DeviceLocation,
			IPAddress:         nil,
			Subnet:            originalSPS.Subnet,
			Gateway:           originalSPS.Gateway,
			Vlan:              originalSPS.Vlan,
		}
		if err := c.spsControllerRepo.Create(ctx, spsCopy); err != nil {
			return err
		}

		if err := c.copySystemTypesAndFieldDevicesForSPSController(ctx, originalSPS.ID, spsCopy.ID); err != nil {
			return err
		}
	}

	return nil
}

func (c *HierarchyCopier) copySystemTypesAndFieldDevicesForSPSController(ctx context.Context, originalSPSControllerID, newSPSControllerID uuid.UUID) error {
	originalSystemTypes, err := c.listSystemTypesBySPSControllerID(ctx, originalSPSControllerID)
	if err != nil {
		return err
	}

	newSystemTypeMap := make(map[uuid.UUID]uuid.UUID, len(originalSystemTypes))
	if len(originalSystemTypes) > 0 {
		systemTypesToCreate := make([]domainFacility.SPSControllerSystemType, 0, len(originalSystemTypes))
		for _, item := range originalSystemTypes {
			systemTypesToCreate = append(systemTypesToCreate, domainFacility.SPSControllerSystemType{
				Number:       item.Number,
				DocumentName: item.DocumentName,
				SystemTypeID: item.SystemTypeID,
			})
		}

		systemTypeMap, err := loadSystemTypeDefinitions(ctx, c.systemTypeRepo, systemTypesToCreate)
		if err != nil {
			return err
		}
		if err := assignSystemTypeNumbers(systemTypesToCreate, systemTypeMap); err != nil {
			return err
		}

		for idx, item := range systemTypesToCreate {
			newSystemType := &domainFacility.SPSControllerSystemType{
				Number:          item.Number,
				DocumentName:    item.DocumentName,
				SPSControllerID: newSPSControllerID,
				SystemTypeID:    item.SystemTypeID,
			}
			if err := c.spsControllerSystemRepo.Create(ctx, newSystemType); err != nil {
				return err
			}
			newSystemTypeMap[originalSystemTypes[idx].ID] = newSystemType.ID
		}
	}

	return c.copyFieldDevicesForSystemTypes(ctx, newSystemTypeMap)
}

func (c *HierarchyCopier) copyFieldDevicesForSystemTypes(ctx context.Context, newSystemTypeMap map[uuid.UUID]uuid.UUID) error {
	if len(newSystemTypeMap) == 0 {
		return nil
	}

	originalSystemTypeIDs := make([]uuid.UUID, 0, len(newSystemTypeMap))
	for originalSystemTypeID := range newSystemTypeMap {
		originalSystemTypeIDs = append(originalSystemTypeIDs, originalSystemTypeID)
	}

	fieldDeviceIDs, err := c.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, originalSystemTypeIDs)
	if err != nil {
		return err
	}
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	originalFieldDevices, err := c.fieldDeviceRepo.GetByIds(ctx, fieldDeviceIDs)
	if err != nil {
		return err
	}

	return copyFieldDevicesWithChildren(
		ctx,
		c.fieldDeviceRepo,
		c.specificationRepo,
		c.bacnetObjectRepo,
		derefSlice(originalFieldDevices),
		newSystemTypeMap,
	)
}

func (c *HierarchyCopier) listSPSControllersByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]domainFacility.SPSController, error) {
	result, err := c.spsControllerRepo.GetPaginatedListByControlCabinetID(ctx, controlCabinetID, domain.PaginationParams{
		Page:  1,
		Limit: 500,
	})
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (c *HierarchyCopier) listSystemTypesBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID) ([]domainFacility.SPSControllerSystemType, error) {
	result, err := c.spsControllerSystemRepo.GetPaginatedListBySPSControllerID(ctx, spsControllerID, domain.PaginationParams{
		Page:  1,
		Limit: 500,
	})
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (c *HierarchyCopier) nextAvailableControlCabinetNr(ctx context.Context, buildingID uuid.UUID, base string) (string, error) {
	for i := 1; i <= 9999; i++ {
		candidate := nextIncrementedValue(base, i, 11)
		exists, err := c.controlCabinetRepo.ExistsControlCabinetNr(ctx, buildingID, candidate, nil)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}

	return "", domain.ErrConflict
}

func (c *HierarchyCopier) nextAvailableGADevice(ctx context.Context, controlCabinetID uuid.UUID) (string, error) {
	devices, err := c.spsControllerRepo.ListGADevicesByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return "", err
	}

	used := make(map[string]struct{}, len(devices))
	for _, device := range devices {
		normalized := strings.ToUpper(strings.TrimSpace(device))
		if isValidGADevice(normalized) {
			used[normalized] = struct{}{}
		}
	}

	if next, ok := findLowestAvailableGADevice(used); ok {
		return next, nil
	}
	return "", domain.ErrConflict
}

func loadSystemTypeDefinitions(
	ctx context.Context,
	systemTypeRepo domainFacility.SystemTypeRepository,
	systemTypes []domainFacility.SPSControllerSystemType,
) (map[uuid.UUID]domainFacility.SystemType, error) {
	if len(systemTypes) == 0 {
		return map[uuid.UUID]domainFacility.SystemType{}, nil
	}

	unique := make(map[uuid.UUID]struct{}, len(systemTypes))
	ids := make([]uuid.UUID, 0, len(systemTypes))
	for _, st := range systemTypes {
		if st.SystemTypeID == uuid.Nil {
			return nil, domain.ErrNotFound
		}
		if _, ok := unique[st.SystemTypeID]; ok {
			continue
		}
		unique[st.SystemTypeID] = struct{}{}
		ids = append(ids, st.SystemTypeID)
	}

	found, err := systemTypeRepo.GetByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	if len(found) != len(ids) {
		return nil, domain.ErrNotFound
	}

	mapOut := make(map[uuid.UUID]domainFacility.SystemType, len(found))
	for _, item := range found {
		mapOut[item.ID] = *item
	}
	return mapOut, nil
}

func assignSystemTypeNumbers(
	systemTypes []domainFacility.SPSControllerSystemType,
	systemTypeMap map[uuid.UUID]domainFacility.SystemType,
) error {
	if len(systemTypes) == 0 {
		return nil
	}

	ve := domain.NewValidationError()
	usedNumbers := make(map[uuid.UUID]map[int]struct{}, len(systemTypes))

	for _, st := range systemTypes {
		systemType, ok := systemTypeMap[st.SystemTypeID]
		if !ok {
			return domain.ErrNotFound
		}
		if st.Number == nil {
			continue
		}
		number := *st.Number
		if number < systemType.NumberMin || number > systemType.NumberMax {
			ve = ve.Add("spscontroller.system_types", "number must be within the system type range")
			continue
		}
		if usedNumbers[st.SystemTypeID] == nil {
			usedNumbers[st.SystemTypeID] = map[int]struct{}{}
		}
		if _, exists := usedNumbers[st.SystemTypeID][number]; exists {
			ve = ve.Add("spscontroller.system_types", "number must be unique per system type")
			continue
		}
		usedNumbers[st.SystemTypeID][number] = struct{}{}
	}

	if len(ve.Fields) > 0 {
		return ve
	}

	for i := range systemTypes {
		if systemTypes[i].Number != nil {
			continue
		}
		systemType, ok := systemTypeMap[systemTypes[i].SystemTypeID]
		if !ok {
			return domain.ErrNotFound
		}
		if usedNumbers[systemTypes[i].SystemTypeID] == nil {
			usedNumbers[systemTypes[i].SystemTypeID] = map[int]struct{}{}
		}
		next, ok := findLowestAvailableNumber(systemType.NumberMin, systemType.NumberMax, usedNumbers[systemTypes[i].SystemTypeID])
		if !ok {
			return domain.NewValidationError().Add("spscontroller.system_types", "no available number in the system type range")
		}
		systemTypes[i].Number = &next
		usedNumbers[systemTypes[i].SystemTypeID][next] = struct{}{}
	}

	return nil
}
