package facility

import (
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

func (c *HierarchyCopier) CopyControlCabinetByID(id uuid.UUID) (*domainFacility.ControlCabinet, error) {
	original, err := domain.GetByID(c.controlCabinetRepo, id)
	if err != nil {
		return nil, err
	}

	baseNr := ""
	if original.ControlCabinetNr != nil {
		baseNr = strings.TrimSpace(*original.ControlCabinetNr)
	}

	nextNr, err := c.nextAvailableControlCabinetNr(original.BuildingID, baseNr)
	if err != nil {
		return nil, err
	}

	copyEntity := &domainFacility.ControlCabinet{
		BuildingID:       original.BuildingID,
		ControlCabinetNr: &nextNr,
	}
	if err := c.controlCabinetRepo.Create(copyEntity); err != nil {
		return nil, err
	}

	if err := c.copySPSControllersForControlCabinet(original.ID, copyEntity.ID); err != nil {
		_ = c.rollbackCopiedControlCabinet(copyEntity.ID)
		return nil, err
	}

	return copyEntity, nil
}

func (c *HierarchyCopier) CopySPSControllerByID(id uuid.UUID) (*domainFacility.SPSController, error) {
	original, err := domain.GetByID(c.spsControllerRepo, id)
	if err != nil {
		return nil, err
	}

	controlCabinet, err := domain.GetByID(c.controlCabinetRepo, original.ControlCabinetID)
	if err != nil {
		return nil, err
	}

	building, err := domain.GetByID(c.buildingRepo, controlCabinet.BuildingID)
	if err != nil {
		return nil, err
	}

	nextGADevice, err := c.nextAvailableGADevice(original.ControlCabinetID)
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
	if err := c.spsControllerRepo.Create(copyEntity); err != nil {
		return nil, err
	}

	if err := c.copySystemTypesAndFieldDevicesForSPSController(original.ID, copyEntity.ID); err != nil {
		_ = c.rollbackCopiedSPSController(copyEntity.ID)
		return nil, err
	}

	return copyEntity, nil
}

func (c *HierarchyCopier) CopySPSControllerSystemTypeByID(id uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	original, err := domain.GetByID(c.spsControllerSystemRepo, id)
	if err != nil {
		return nil, err
	}

	systemType, err := domain.GetByID(c.systemTypeRepo, original.SystemTypeID)
	if err != nil {
		return nil, err
	}

	existing, err := c.spsControllerSystemRepo.ListBySPSControllerID(original.SPSControllerID)
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
	if err := c.spsControllerSystemRepo.Create(copyEntity); err != nil {
		return nil, err
	}

	if err := c.copyFieldDevicesForSystemTypes(map[uuid.UUID]uuid.UUID{original.ID: copyEntity.ID}); err != nil {
		_ = c.rollbackCopiedSPSControllerSystemType(copyEntity.ID)
		return nil, err
	}

	return domain.GetByID(c.spsControllerSystemRepo, copyEntity.ID)
}

func (c *HierarchyCopier) copySPSControllersForControlCabinet(originalControlCabinetID, newControlCabinetID uuid.UUID) error {
	newControlCabinet, err := domain.GetByID(c.controlCabinetRepo, newControlCabinetID)
	if err != nil {
		return err
	}

	building, err := domain.GetByID(c.buildingRepo, newControlCabinet.BuildingID)
	if err != nil {
		return err
	}

	newCabinetNr := ""
	if newControlCabinet.ControlCabinetNr != nil {
		newCabinetNr = strings.TrimSpace(*newControlCabinet.ControlCabinetNr)
	}
	buildingIWSCode := strings.TrimSpace(building.IWSCode)

	originalSPSControllers, err := c.listSPSControllersByControlCabinetID(originalControlCabinetID)
	if err != nil {
		return err
	}

	for _, originalSPS := range originalSPSControllers {
		gaDevice := ""
		if originalSPS.GADevice != nil {
			gaDevice = strings.ToUpper(strings.TrimSpace(*originalSPS.GADevice))
		}
		if gaDevice == "" {
			nextGADevice, err := c.nextAvailableGADevice(newControlCabinetID)
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
		if err := c.spsControllerRepo.Create(spsCopy); err != nil {
			return err
		}

		if err := c.copySystemTypesAndFieldDevicesForSPSController(originalSPS.ID, spsCopy.ID); err != nil {
			return err
		}
	}

	return nil
}

func (c *HierarchyCopier) copySystemTypesAndFieldDevicesForSPSController(originalSPSControllerID, newSPSControllerID uuid.UUID) error {
	originalSystemTypes, err := c.listSystemTypesBySPSControllerID(originalSPSControllerID)
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

		systemTypeMap, err := loadSystemTypeDefinitions(c.systemTypeRepo, systemTypesToCreate)
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
			if err := c.spsControllerSystemRepo.Create(newSystemType); err != nil {
				return err
			}
			newSystemTypeMap[originalSystemTypes[idx].ID] = newSystemType.ID
		}
	}

	return c.copyFieldDevicesForSystemTypes(newSystemTypeMap)
}

func (c *HierarchyCopier) copyFieldDevicesForSystemTypes(newSystemTypeMap map[uuid.UUID]uuid.UUID) error {
	if len(newSystemTypeMap) == 0 {
		return nil
	}

	originalSystemTypeIDs := make([]uuid.UUID, 0, len(newSystemTypeMap))
	for originalSystemTypeID := range newSystemTypeMap {
		originalSystemTypeIDs = append(originalSystemTypeIDs, originalSystemTypeID)
	}

	fieldDeviceIDs, err := c.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(originalSystemTypeIDs)
	if err != nil {
		return err
	}
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	originalFieldDevices, err := c.fieldDeviceRepo.GetByIds(fieldDeviceIDs)
	if err != nil {
		return err
	}

	return copyFieldDevicesWithChildren(
		c.fieldDeviceRepo,
		c.specificationRepo,
		c.bacnetObjectRepo,
		derefSlice(originalFieldDevices),
		newSystemTypeMap,
	)
}

func (c *HierarchyCopier) listSPSControllersByControlCabinetID(controlCabinetID uuid.UUID) ([]domainFacility.SPSController, error) {
	result, err := c.spsControllerRepo.GetPaginatedListByControlCabinetID(controlCabinetID, domain.PaginationParams{
		Page:  1,
		Limit: 500,
	})
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (c *HierarchyCopier) listSystemTypesBySPSControllerID(spsControllerID uuid.UUID) ([]domainFacility.SPSControllerSystemType, error) {
	result, err := c.spsControllerSystemRepo.GetPaginatedListBySPSControllerID(spsControllerID, domain.PaginationParams{
		Page:  1,
		Limit: 500,
	})
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (c *HierarchyCopier) nextAvailableControlCabinetNr(buildingID uuid.UUID, base string) (string, error) {
	for i := 1; i <= 9999; i++ {
		candidate := nextIncrementedValue(base, i, 11)
		exists, err := c.controlCabinetRepo.ExistsControlCabinetNr(buildingID, candidate, nil)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}

	return "", domain.ErrConflict
}

func (c *HierarchyCopier) nextAvailableGADevice(controlCabinetID uuid.UUID) (string, error) {
	devices, err := c.spsControllerRepo.ListGADevicesByControlCabinetID(controlCabinetID)
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

func (c *HierarchyCopier) rollbackCopiedControlCabinet(controlCabinetID uuid.UUID) error {
	spsControllerIDs, err := c.spsControllerRepo.GetIDsByControlCabinetID(controlCabinetID)
	if err != nil {
		return err
	}
	if err := c.rollbackCopiedSPSControllers(spsControllerIDs); err != nil {
		return err
	}
	if len(spsControllerIDs) > 0 {
		if err := c.spsControllerRepo.DeleteByIds(spsControllerIDs); err != nil {
			return err
		}
	}
	return c.controlCabinetRepo.DeleteByIds([]uuid.UUID{controlCabinetID})
}

func (c *HierarchyCopier) rollbackCopiedSPSController(spsControllerID uuid.UUID) error {
	if err := c.rollbackCopiedSPSControllers([]uuid.UUID{spsControllerID}); err != nil {
		return err
	}
	return c.spsControllerRepo.DeleteByIds([]uuid.UUID{spsControllerID})
}

func (c *HierarchyCopier) rollbackCopiedSPSControllerSystemType(systemTypeID uuid.UUID) error {
	fieldDeviceIDs, err := c.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs([]uuid.UUID{systemTypeID})
	if err != nil {
		return err
	}
	if err := c.deleteFieldDevicesWithChildren(fieldDeviceIDs); err != nil {
		return err
	}
	return c.spsControllerSystemRepo.DeleteByIds([]uuid.UUID{systemTypeID})
}

func (c *HierarchyCopier) rollbackCopiedSPSControllers(spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}

	systemTypeIDs, err := c.spsControllerSystemRepo.GetIDsBySPSControllerIDs(spsControllerIDs)
	if err != nil {
		return err
	}
	fieldDeviceIDs, err := c.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(systemTypeIDs)
	if err != nil {
		return err
	}
	if err := c.deleteFieldDevicesWithChildren(fieldDeviceIDs); err != nil {
		return err
	}
	return c.spsControllerSystemRepo.DeleteBySPSControllerIDs(spsControllerIDs)
}

func (c *HierarchyCopier) deleteFieldDevicesWithChildren(fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}
	if err := c.bacnetObjectRepo.DeleteByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	if err := c.specificationRepo.DeleteByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	return c.fieldDeviceRepo.DeleteByIds(fieldDeviceIDs)
}

func loadSystemTypeDefinitions(
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

	found, err := systemTypeRepo.GetByIds(ids)
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
