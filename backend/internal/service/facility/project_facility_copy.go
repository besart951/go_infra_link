package facility

import (
	"context"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type projectFacilityCopy struct {
	controlCabinetRepo      domainFacility.ControlCabinetRepository
	buildingRepo            domainFacility.BuildingRepository
	spsControllerRepo       domainFacility.SPSControllerRepository
	systemTypeRepo          domainFacility.SystemTypeRepository
	spsControllerSystemRepo domainFacility.SPSControllerSystemTypeStore
	fieldDeviceRepo         domainFacility.FieldDeviceStore
	specificationRepo       domainFacility.SpecificationStore
	bacnetObjectRepo        domainFacility.BacnetObjectStore
	objectDataRepo          domainFacility.ObjectDataStore
	alarmTypeRepo           domainFacility.AlarmTypeRepository
	bacnetAlarmValueRepo    domainFacility.BacnetObjectAlarmValueRepository
}

type spsControllerBulkCreator interface {
	BulkCreate(ctx context.Context, entities []*domainFacility.SPSController, batchSize int) error
}

type spsControllerSystemTypeBulkCreator interface {
	BulkCreate(ctx context.Context, entities []*domainFacility.SPSControllerSystemType, batchSize int) error
}

const (
	copyFieldDeviceSystemTypeChunkSize = 10
	copySPSControllerPageLimit         = 500
)

func (s *FieldDeviceService) projectFacilityCopy() projectFacilityCopy {
	return projectFacilityCopy{
		fieldDeviceRepo:         s.repo,
		spsControllerSystemRepo: s.spsControllerSystemTypeRepo,
		systemTypeRepo:          s.systemTypeRepo,
		specificationRepo:       s.specificationRepo,
		bacnetObjectRepo:        s.bacnetObjectRepo,
		objectDataRepo:          s.objectDataRepo,
		alarmTypeRepo:           s.alarmTypeRepo,
		bacnetAlarmValueRepo:    s.bacnetAlarmValueRepo,
	}
}

// CopyObjectDataTemplatesForProject creates project-scoped ObjectData copies
// and remaps internal BACnet Object software references.
func CopyObjectDataTemplatesForProject(
	ctx context.Context,
	objectDataRepo domainFacility.ObjectDataStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	projectID uuid.UUID,
) error {
	return projectFacilityCopy{
		objectDataRepo:   objectDataRepo,
		bacnetObjectRepo: bacnetObjectRepo,
	}.copyObjectDataTemplatesForProject(ctx, projectID)
}

func (c projectFacilityCopy) copyControlCabinetByID(ctx context.Context, id uuid.UUID) (*domainFacility.ControlCabinet, error) {
	original, err := domain.GetByID(ctx, c.controlCabinetRepo, id)
	if err != nil {
		return nil, err
	}

	baseNr := ""
	if original.ControlCabinetNr != nil {
		baseNr = strings.TrimSpace(*original.ControlCabinetNr)
	}

	nextNr, err := c.nextAvailableControlCabinetNr(ctx, original.BuildingID, baseNr)
	if err != nil {
		return nil, err
	}

	copyEntity := &domainFacility.ControlCabinet{
		BuildingID:       original.BuildingID,
		ControlCabinetNr: &nextNr,
	}
	if err := c.controlCabinetRepo.Create(ctx, copyEntity); err != nil {
		return nil, err
	}

	if err := c.copySPSControllersForControlCabinet(ctx, original.ID, copyEntity.ID); err != nil {
		return nil, err
	}

	return copyEntity, nil
}

func (c projectFacilityCopy) copySPSControllerByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSController, error) {
	original, err := domain.GetByID(ctx, c.spsControllerRepo, id)
	if err != nil {
		return nil, err
	}

	controlCabinet, err := domain.GetByID(ctx, c.controlCabinetRepo, original.ControlCabinetID)
	if err != nil {
		return nil, err
	}

	building, err := domain.GetByID(ctx, c.buildingRepo, controlCabinet.BuildingID)
	if err != nil {
		return nil, err
	}

	nextGADevice, err := c.nextAvailableGADevice(ctx, original.ControlCabinetID)
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
	if err := c.spsControllerRepo.Create(ctx, copyEntity); err != nil {
		return nil, err
	}

	if err := c.copySystemTypesAndFieldDevicesForSPSController(ctx, original.ID, copyEntity.ID); err != nil {
		return nil, err
	}

	return copyEntity, nil
}

func (c projectFacilityCopy) copySPSControllerSystemTypeByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	original, err := domain.GetByID(ctx, c.spsControllerSystemRepo, id)
	if err != nil {
		return nil, err
	}

	systemType, err := domain.GetByID(ctx, c.systemTypeRepo, original.SystemTypeID)
	if err != nil {
		return nil, err
	}

	existing, err := c.spsControllerSystemRepo.ListBySPSControllerID(ctx, original.SPSControllerID)
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
	if err := c.spsControllerSystemRepo.Create(ctx, copyEntity); err != nil {
		return nil, err
	}

	if err := c.copyFieldDevicesForSystemTypes(ctx, map[uuid.UUID]uuid.UUID{original.ID: copyEntity.ID}); err != nil {
		return nil, err
	}

	return domain.GetByID(ctx, c.spsControllerSystemRepo, copyEntity.ID)
}

func (c projectFacilityCopy) copySPSControllersForControlCabinet(ctx context.Context, originalControlCabinetID, newControlCabinetID uuid.UUID) error {
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

	spsCopies := make([]*domainFacility.SPSController, 0, len(originalSPSControllers))
	originalToCopy := make(map[uuid.UUID]*domainFacility.SPSController, len(originalSPSControllers))

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
		spsCopies = append(spsCopies, spsCopy)
		originalToCopy[originalSPS.ID] = spsCopy
	}

	if err := c.createSPSControllerCopies(ctx, spsCopies); err != nil {
		return err
	}

	spsIDMap := make(map[uuid.UUID]uuid.UUID, len(originalToCopy))
	for originalID, copyEntity := range originalToCopy {
		spsIDMap[originalID] = copyEntity.ID
	}

	return c.copySystemTypesAndFieldDevicesForSPSControllers(ctx, spsIDMap)
}

func (c projectFacilityCopy) copySystemTypesAndFieldDevicesForSPSController(ctx context.Context, originalSPSControllerID, newSPSControllerID uuid.UUID) error {
	return c.copySystemTypesAndFieldDevicesForSPSControllers(ctx, map[uuid.UUID]uuid.UUID{originalSPSControllerID: newSPSControllerID})
}

func (c projectFacilityCopy) copySystemTypesAndFieldDevicesForSPSControllers(ctx context.Context, spsIDMap map[uuid.UUID]uuid.UUID) error {
	if len(spsIDMap) == 0 {
		return nil
	}

	originalSPSControllerIDs := make([]uuid.UUID, 0, len(spsIDMap))
	for originalID := range spsIDMap {
		originalSPSControllerIDs = append(originalSPSControllerIDs, originalID)
	}

	systemTypeIDs, err := c.spsControllerSystemRepo.GetIDsBySPSControllerIDs(ctx, originalSPSControllerIDs)
	if err != nil {
		return err
	}
	if len(systemTypeIDs) == 0 {
		return nil
	}

	originalSystemTypes, err := c.spsControllerSystemRepo.GetByIds(ctx, systemTypeIDs)
	if err != nil {
		return err
	}

	newSystemTypeMap := make(map[uuid.UUID]uuid.UUID, len(originalSystemTypes))
	newSystemTypes := make([]*domainFacility.SPSControllerSystemType, 0, len(originalSystemTypes))
	originalToCopy := make(map[uuid.UUID]*domainFacility.SPSControllerSystemType, len(originalSystemTypes))

	for _, item := range originalSystemTypes {
		if item == nil {
			continue
		}
		newSPSControllerID, ok := spsIDMap[item.SPSControllerID]
		if !ok {
			continue
		}

		newSystemType := &domainFacility.SPSControllerSystemType{
			Number:          cloneIntPointer(item.Number),
			DocumentName:    item.DocumentName,
			SPSControllerID: newSPSControllerID,
			SystemTypeID:    item.SystemTypeID,
		}
		newSystemTypes = append(newSystemTypes, newSystemType)
		originalToCopy[item.ID] = newSystemType
	}

	if err := c.createSPSControllerSystemTypeCopies(ctx, newSystemTypes); err != nil {
		return err
	}
	for originalID, copyEntity := range originalToCopy {
		newSystemTypeMap[originalID] = copyEntity.ID
	}

	return c.copyFieldDevicesForSystemTypes(ctx, newSystemTypeMap)
}

func (c projectFacilityCopy) createSPSControllerCopies(ctx context.Context, copies []*domainFacility.SPSController) error {
	if len(copies) == 0 {
		return nil
	}
	if repo, ok := c.spsControllerRepo.(spsControllerBulkCreator); ok {
		return repo.BulkCreate(ctx, copies, 100)
	}
	for _, copyEntity := range copies {
		if err := c.spsControllerRepo.Create(ctx, copyEntity); err != nil {
			return err
		}
	}
	return nil
}

func (c projectFacilityCopy) createSPSControllerSystemTypeCopies(ctx context.Context, copies []*domainFacility.SPSControllerSystemType) error {
	if len(copies) == 0 {
		return nil
	}
	if repo, ok := c.spsControllerSystemRepo.(spsControllerSystemTypeBulkCreator); ok {
		return repo.BulkCreate(ctx, copies, 100)
	}
	for _, copyEntity := range copies {
		if err := c.spsControllerSystemRepo.Create(ctx, copyEntity); err != nil {
			return err
		}
	}
	return nil
}

func cloneIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func (c projectFacilityCopy) listSPSControllersByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]domainFacility.SPSController, error) {
	items := make([]domainFacility.SPSController, 0)
	for page := 1; ; page++ {
		result, err := c.spsControllerRepo.GetPaginatedListByControlCabinetID(ctx, controlCabinetID, domain.PaginationParams{
			Page:  page,
			Limit: copySPSControllerPageLimit,
		})
		if err != nil {
			return nil, err
		}

		items = append(items, result.Items...)
		if page >= result.TotalPages || len(result.Items) == 0 {
			break
		}
	}
	return items, nil
}

func (c projectFacilityCopy) nextAvailableControlCabinetNr(ctx context.Context, buildingID uuid.UUID, base string) (string, error) {
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

func (c projectFacilityCopy) nextAvailableGADevice(ctx context.Context, controlCabinetID uuid.UUID) (string, error) {
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

func (c projectFacilityCopy) loadSystemTypeDefinitions(
	ctx context.Context,
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

	found, err := c.systemTypeRepo.GetByIds(ctx, ids)
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

func loadSystemTypeDefinitions(
	ctx context.Context,
	systemTypeRepo domainFacility.SystemTypeRepository,
	systemTypes []domainFacility.SPSControllerSystemType,
) (map[uuid.UUID]domainFacility.SystemType, error) {
	return projectFacilityCopy{systemTypeRepo: systemTypeRepo}.loadSystemTypeDefinitions(ctx, systemTypes)
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
