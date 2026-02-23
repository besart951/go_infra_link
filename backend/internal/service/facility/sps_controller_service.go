package facility

import (
	"net"
	"strconv"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type SPSControllerService struct {
	repo                     domainFacility.SPSControllerRepository
	controlCabinetRepo       domainFacility.ControlCabinetRepository
	buildingRepo             domainFacility.BuildingRepository
	systemTypeRepo           domainFacility.SystemTypeRepository
	spsControllerSystemTyper domainFacility.SPSControllerSystemTypeStore
	fieldDeviceRepo          domainFacility.FieldDeviceStore
	specificationRepo        domainFacility.SpecificationStore
	bacnetObjectRepo         domainFacility.BacnetObjectStore
}

func NewSPSControllerService(
	repo domainFacility.SPSControllerRepository,
	controlCabinetRepo domainFacility.ControlCabinetRepository,
	buildingRepo domainFacility.BuildingRepository,
	systemTypeRepo domainFacility.SystemTypeRepository,
	spsControllerSystemTypeStore domainFacility.SPSControllerSystemTypeStore,
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	specificationRepo domainFacility.SpecificationStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
) *SPSControllerService {
	return &SPSControllerService{
		repo:                     repo,
		controlCabinetRepo:       controlCabinetRepo,
		buildingRepo:             buildingRepo,
		systemTypeRepo:           systemTypeRepo,
		spsControllerSystemTyper: spsControllerSystemTypeStore,
		fieldDeviceRepo:          fieldDeviceRepo,
		specificationRepo:        specificationRepo,
		bacnetObjectRepo:         bacnetObjectRepo,
	}
}

func (s *SPSControllerService) Create(spsController *domainFacility.SPSController) error {
	return s.CreateWithSystemTypes(spsController, nil)
}

func (s *SPSControllerService) CreateWithSystemTypes(spsController *domainFacility.SPSController, systemTypes []domainFacility.SPSControllerSystemType) error {
	if err := s.ensureControlCabinetExists(spsController.ControlCabinetID); err != nil {
		return err
	}
	if err := s.ensureGADeviceAssigned(spsController, nil); err != nil {
		return err
	}
	if err := s.Validate(spsController, nil); err != nil {
		return err
	}
	systemTypeMap, err := s.loadSystemTypes(systemTypes)
	if err != nil {
		return err
	}
	if err := s.assignSystemTypeNumbers(systemTypes, systemTypeMap); err != nil {
		return err
	}

	if err := s.repo.Create(spsController); err != nil {
		return err
	}
	if len(systemTypes) == 0 {
		return nil
	}

	for _, st := range systemTypes {
		entity := &domainFacility.SPSControllerSystemType{
			Number:          st.Number,
			DocumentName:    st.DocumentName,
			SPSControllerID: spsController.ID,
			SystemTypeID:    st.SystemTypeID,
		}
		if err := s.spsControllerSystemTyper.Create(entity); err != nil {
			return err
		}
	}
	return nil
}

func (s *SPSControllerService) GetByID(id uuid.UUID) (*domainFacility.SPSController, error) {
	return domain.GetByID(s.repo, id)
}

func (s *SPSControllerService) GetByIDs(ids []uuid.UUID) ([]domainFacility.SPSController, error) {
	spsControllers, err := s.repo.GetByIds(ids)
	if err != nil {
		return nil, err
	}
	items := make([]domainFacility.SPSController, len(spsControllers))
	for i, item := range spsControllers {
		items[i] = *item
	}
	return items, nil
}

func (s *SPSControllerService) CopyByID(id uuid.UUID) (*domainFacility.SPSController, error) {
	original, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Load control cabinet to get building ID
	controlCabinet, err := domain.GetByID(s.controlCabinetRepo, original.ControlCabinetID)
	if err != nil {
		return nil, err
	}

	// Load building
	building, err := domain.GetByID(s.buildingRepo, controlCabinet.BuildingID)
	if err != nil {
		return nil, err
	}

	// Get next available GA Device
	nextGADevice, err := s.nextAvailableGADevice(original.ControlCabinetID)
	if err != nil {
		return nil, err
	}
	if nextGADevice == "" {
		return nil, domain.NewValidationError().Add("spscontroller.ga_device", "no available ga_device for control cabinet")
	}

	// Generate device name using the same logic as frontend: {iwsCode}_{cabinetNr}_{gaDevice}
	iwsCode := strings.TrimSpace(building.IWSCode)
	cabinetNr := ""
	if controlCabinet.ControlCabinetNr != nil {
		cabinetNr = strings.TrimSpace(*controlCabinet.ControlCabinetNr)
	}

	var deviceName string
	if iwsCode != "" && cabinetNr != "" {
		deviceName = strings.ToUpper(iwsCode + "_" + cabinetNr + "_" + nextGADevice)
	} else {
		// Fallback if building or cabinet number is missing
		deviceName = nextGADevice
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

	if err := s.Create(copyEntity); err != nil {
		return nil, err
	}

	originalSystemTypes, err := s.listSystemTypesBySPSControllerID(id)
	if err != nil {
		return nil, err
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

		systemTypeMap, err := s.loadSystemTypes(systemTypesToCreate)
		if err != nil {
			return nil, err
		}
		if err := s.assignSystemTypeNumbers(systemTypesToCreate, systemTypeMap); err != nil {
			return nil, err
		}

		for idx, item := range systemTypesToCreate {
			newSystemType := &domainFacility.SPSControllerSystemType{
				Number:          item.Number,
				DocumentName:    item.DocumentName,
				SPSControllerID: copyEntity.ID,
				SystemTypeID:    item.SystemTypeID,
			}
			if err := s.spsControllerSystemTyper.Create(newSystemType); err != nil {
				return nil, err
			}
			newSystemTypeMap[originalSystemTypes[idx].ID] = newSystemType.ID
		}
	}

	originalFieldDevices, err := s.listFieldDevicesBySPSControllerID(id)
	if err != nil {
		return nil, err
	}

	// Copy field devices along with their specifications and BACnet objects
	for _, originalFieldDevice := range originalFieldDevices {
		newSystemTypeID, ok := newSystemTypeMap[originalFieldDevice.SPSControllerSystemTypeID]
		if !ok {
			continue
		}

		fieldDeviceCopy := &domainFacility.FieldDevice{
			BMK:                       originalFieldDevice.BMK,
			Description:               originalFieldDevice.Description,
			ApparatNr:                 originalFieldDevice.ApparatNr,
			SPSControllerSystemTypeID: newSystemTypeID,
			SystemPartID:              originalFieldDevice.SystemPartID,
			ApparatID:                 originalFieldDevice.ApparatID,
		}
		if err := s.fieldDeviceRepo.Create(fieldDeviceCopy); err != nil {
			return nil, err
		}

		// Copy specification if exists
		if err := copySpecificationForFieldDevice(s.specificationRepo, originalFieldDevice.ID, fieldDeviceCopy.ID); err != nil {
			return nil, err
		}

		// Copy BACnet objects if exist
		if err := copyBacnetObjectsForFieldDevice(s.bacnetObjectRepo, originalFieldDevice.ID, fieldDeviceCopy.ID); err != nil {
			return nil, err
		}
	}

	return copyEntity, nil
}

func (s *SPSControllerService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSController], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SPSControllerService) ListByControlCabinetID(controlCabinetID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSController], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedListByControlCabinetID(controlCabinetID, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SPSControllerService) Update(spsController *domainFacility.SPSController) error {
	if err := s.Validate(spsController, &spsController.ID); err != nil {
		return err
	}
	if err := s.ensureControlCabinetExists(spsController.ControlCabinetID); err != nil {
		return err
	}
	return s.repo.Update(spsController)
}

func (s *SPSControllerService) UpdateWithSystemTypes(spsController *domainFacility.SPSController, systemTypes []domainFacility.SPSControllerSystemType) error {
	if err := s.Validate(spsController, &spsController.ID); err != nil {
		return err
	}
	if err := s.ensureControlCabinetExists(spsController.ControlCabinetID); err != nil {
		return err
	}
	systemTypeMap, err := s.loadSystemTypes(systemTypes)
	if err != nil {
		return err
	}
	if err := s.assignSystemTypeNumbers(systemTypes, systemTypeMap); err != nil {
		return err
	}

	if err := s.repo.Update(spsController); err != nil {
		return err
	}

	existing, err := s.spsControllerSystemTyper.ListBySPSControllerID(spsController.ID)
	if err != nil {
		return err
	}

	existingBySystemType := make(map[uuid.UUID]*domainFacility.SPSControllerSystemType, len(existing))
	for _, item := range existing {
		existingBySystemType[item.SystemTypeID] = item
	}

	incomingSystemTypeIDs := make(map[uuid.UUID]struct{}, len(systemTypes))
	for _, st := range systemTypes {
		incomingSystemTypeIDs[st.SystemTypeID] = struct{}{}
		if existingItem, ok := existingBySystemType[st.SystemTypeID]; ok {
			existingItem.Number = st.Number
			existingItem.DocumentName = st.DocumentName
			if err := s.spsControllerSystemTyper.Update(existingItem); err != nil {
				return err
			}
			continue
		}

		entity := &domainFacility.SPSControllerSystemType{
			Number:          st.Number,
			DocumentName:    st.DocumentName,
			SPSControllerID: spsController.ID,
			SystemTypeID:    st.SystemTypeID,
		}
		if err := s.spsControllerSystemTyper.Create(entity); err != nil {
			return err
		}
	}

	var deleteIDs []uuid.UUID
	for _, item := range existing {
		if _, ok := incomingSystemTypeIDs[item.SystemTypeID]; !ok {
			deleteIDs = append(deleteIDs, item.ID)
		}
	}

	if len(deleteIDs) > 0 {
		fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(deleteIDs)
		if err != nil {
			return err
		}
		if len(fieldDeviceIDs) > 0 {
			return domain.NewValidationError().Add("spscontroller.system_types", "referenced_entity_in_use")
		}
		if err := s.spsControllerSystemTyper.DeleteByIds(deleteIDs); err != nil {
			return err
		}
	}

	return nil
}

func (s *SPSControllerService) DeleteByID(id uuid.UUID) error {
	// Cascade delete: SPSController → SPSControllerSystemTypes → FieldDevices
	spsControllerSystemTypeIDs, err := s.spsControllerSystemTyper.GetIDsBySPSControllerIDs([]uuid.UUID{id})
	if err != nil {
		return err
	}

	if len(spsControllerSystemTypeIDs) == 0 {
		// No children, delete directly
		return s.repo.DeleteByIds([]uuid.UUID{id})
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(spsControllerSystemTypeIDs)
	if err != nil {
		return err
	}

	// Delete in correct order (bottom-up)
	if err := s.bacnetObjectRepo.DeleteByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.specificationRepo.DeleteByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.fieldDeviceRepo.DeleteByIds(fieldDeviceIDs); err != nil {
		return err
	}

	if err := s.spsControllerSystemTyper.DeleteBySPSControllerIDs([]uuid.UUID{id}); err != nil {
		return err
	}

	return s.repo.DeleteByIds([]uuid.UUID{id})
}

func (s *SPSControllerService) ensureControlCabinetExists(controlCabinetID uuid.UUID) error {
	_, err := domain.GetByID(s.controlCabinetRepo, controlCabinetID)
	return err
}

func (s *SPSControllerService) loadSystemTypes(systemTypes []domainFacility.SPSControllerSystemType) (map[uuid.UUID]domainFacility.SystemType, error) {
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

	found, err := s.systemTypeRepo.GetByIds(ids)
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

func (s *SPSControllerService) assignSystemTypeNumbers(systemTypes []domainFacility.SPSControllerSystemType, systemTypeMap map[uuid.UUID]domainFacility.SystemType) error {
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

func findLowestAvailableNumber(min, max int, used map[int]struct{}) (int, bool) {
	for i := min; i <= max; i++ {
		if _, exists := used[i]; !exists {
			return i, true
		}
	}
	return 0, false
}

func (s *SPSControllerService) GetNextGADevice(controlCabinetID uuid.UUID) (string, error) {
	if controlCabinetID == uuid.Nil {
		return "", domain.ErrInvalidArgument
	}
	if err := s.ensureControlCabinetExists(controlCabinetID); err != nil {
		return "", err
	}
	return s.nextAvailableGADevice(controlCabinetID)
}

func (s *SPSControllerService) ensureGADeviceAssigned(spsController *domainFacility.SPSController, excludeID *uuid.UUID) error {
	if spsController.GADevice != nil && strings.TrimSpace(*spsController.GADevice) != "" {
		return nil
	}

	next, err := s.nextAvailableGADevice(spsController.ControlCabinetID)
	if err != nil {
		return err
	}
	if next == "" {
		return domain.NewValidationError().Add("spscontroller.ga_device", "no available ga_device for control cabinet")
	}
	spsController.GADevice = &next

	// Always ensure uniqueness, not just when updating
	if err := s.ensureUnique(spsController, excludeID); err != nil {
		return err
	}
	return nil
}

func (s *SPSControllerService) nextAvailableGADevice(controlCabinetID uuid.UUID) (string, error) {
	devices, err := s.repo.ListGADevicesByControlCabinetID(controlCabinetID)
	if err != nil {
		return "", err
	}

	used := make(map[string]struct{}, len(devices))
	for _, device := range devices {
		normalized := strings.TrimSpace(strings.ToUpper(device))
		if len(normalized) == 3 {
			used[normalized] = struct{}{}
		}
	}

	const max = 26 * 26 * 26
	for i := 0; i < max; i++ {
		candidate := gaDeviceFromIndex(i)
		if _, exists := used[candidate]; !exists {
			return candidate, nil
		}
	}

	return "", nil
}

func (s *SPSControllerService) validateRequiredFields(spsController *domainFacility.SPSController) error {
	ve := domain.NewValidationError()
	if spsController.ControlCabinetID == uuid.Nil {
		ve = ve.Add("spscontroller.control_cabinet_id", "control_cabinet_id is required")
	}
	if strings.TrimSpace(spsController.DeviceName) == "" {
		ve = ve.Add("spscontroller.device_name", "device_name is required")
	}
	if spsController.GADevice == nil || strings.TrimSpace(*spsController.GADevice) == "" {
		ve = ve.Add("spscontroller.ga_device", "ga_device is required")
	} else if !isValidGADevice(*spsController.GADevice) {
		ve = ve.Add("spscontroller.ga_device", "ga_device must be exactly 3 uppercase letters (A-Z)")
	}

	if err := validateNetworkFields(spsController); err != nil {
		if veNet, ok := domain.AsValidationError(err); ok {
			for field, msg := range veNet.Fields {
				ve = ve.Add(field, msg)
			}
		}
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func (s *SPSControllerService) Validate(spsController *domainFacility.SPSController, excludeID *uuid.UUID) error {
	if err := s.validateRequiredFields(spsController); err != nil {
		return err
	}
	if err := s.ensureUnique(spsController, excludeID); err != nil {
		return err
	}
	return nil
}

func (s *SPSControllerService) NextAvailableGADevice(controlCabinetID uuid.UUID, excludeID *uuid.UUID) (string, error) {
	if err := s.ensureControlCabinetExists(controlCabinetID); err != nil {
		return "", err
	}

	items, err := s.repo.ListGADevicesByControlCabinetID(controlCabinetID)
	if err != nil {
		return "", err
	}

	used := make(map[string]struct{}, len(items))
	for _, item := range items {
		normalized := strings.ToUpper(strings.TrimSpace(item))
		if isValidGADevice(normalized) {
			used[normalized] = struct{}{}
		}
	}

	if excludeID != nil {
		controller, err := domain.GetByID(s.repo, *excludeID)
		if err != nil {
			return "", err
		}
		if controller.GADevice != nil {
			current := strings.ToUpper(strings.TrimSpace(*controller.GADevice))
			if current != "" {
				delete(used, current)
			}
		}
	}

	if next, ok := findLowestAvailableGADevice(used); ok {
		return next, nil
	}
	return "", domain.ErrConflict
}

func (s *SPSControllerService) ensureUnique(spsController *domainFacility.SPSController, excludeID *uuid.UUID) error {
	ve := domain.NewValidationError()

	if strings.TrimSpace(spsController.DeviceName) != "" {
		items, err := s.repo.GetPaginatedList(domain.PaginationParams{Page: 1, Limit: 1000, Search: spsController.DeviceName})
		if err != nil {
			return err
		}
		for i := range items.Items {
			item := items.Items[i]
			if excludeID != nil && item.ID == *excludeID {
				continue
			}
			if item.ControlCabinetID == spsController.ControlCabinetID && strings.EqualFold(item.DeviceName, spsController.DeviceName) {
				ve = ve.Add("spscontroller.device_name", "device_name must be unique within the control cabinet")
				break
			}
		}
	}

	if spsController.GADevice != nil && strings.TrimSpace(*spsController.GADevice) != "" {
		exists, err := s.repo.ExistsGADevice(spsController.ControlCabinetID, *spsController.GADevice, excludeID)
		if err != nil {
			return err
		}
		if exists {
			ve = ve.Add("spscontroller.ga_device", "ga_device must be unique within the control cabinet")
		}
	}

	if spsController.IPAddress != nil && spsController.Vlan != nil {
		ip := strings.TrimSpace(*spsController.IPAddress)
		vlan := strings.TrimSpace(*spsController.Vlan)
		if ip != "" && vlan != "" {
			exists, err := s.repo.ExistsIPAddressVlan(ip, vlan, excludeID)
			if err != nil {
				return err
			}
			if exists {
				ve = ve.Add("spscontroller.ip_address", "ip_address must be unique per vlan")
				ve = ve.Add("spscontroller.vlan", "vlan must be unique per ip_address")
			}
		}
	}

	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func (s *SPSControllerService) listSystemTypesBySPSControllerID(spsControllerID uuid.UUID) ([]domainFacility.SPSControllerSystemType, error) {
	items := make([]domainFacility.SPSControllerSystemType, 0)
	page := 1

	for {
		result, err := s.spsControllerSystemTyper.GetPaginatedListBySPSControllerID(spsControllerID, domain.PaginationParams{Page: page, Limit: 500})
		if err != nil {
			return nil, err
		}

		items = append(items, result.Items...)
		if page >= result.TotalPages || len(result.Items) == 0 {
			break
		}
		page++
	}

	return items, nil
}

func (s *SPSControllerService) listFieldDevicesBySPSControllerID(spsControllerID uuid.UUID) ([]domainFacility.FieldDevice, error) {
	items := make([]domainFacility.FieldDevice, 0)
	page := 1
	filters := domainFacility.FieldDeviceFilterParams{SPSControllerID: &spsControllerID}

	for {
		result, err := s.fieldDeviceRepo.GetPaginatedListWithFilters(domain.PaginationParams{Page: page, Limit: 500}, filters)
		if err != nil {
			return nil, err
		}

		items = append(items, result.Items...)
		if page >= result.TotalPages || len(result.Items) == 0 {
			break
		}
		page++
	}

	return items, nil
}

func isValidGADevice(value string) bool {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) != 3 {
		return false
	}
	for _, r := range trimmed {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func findLowestAvailableGADevice(used map[string]struct{}) (string, bool) {
	const max = 26 * 26 * 26
	for i := 0; i < max; i++ {
		candidate := gaDeviceFromIndex(i)
		if _, exists := used[candidate]; !exists {
			return candidate, true
		}
	}
	return "", false
}

func gaDeviceFromIndex(index int) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	first := index % 26
	second := (index / 26) % 26
	third := (index / (26 * 26)) % 26
	return string([]byte{alphabet[first], alphabet[second], alphabet[third]})
}

func validateNetworkFields(spsController *domainFacility.SPSController) error {
	ve := domain.NewValidationError()

	if spsController.IPAddress != nil && strings.TrimSpace(*spsController.IPAddress) != "" {
		if !isValidIPv4(*spsController.IPAddress) {
			ve = ve.Add("spscontroller.ip_address", "ip_address must be a valid IPv4 address")
		}
	}
	if spsController.Gateway != nil && strings.TrimSpace(*spsController.Gateway) != "" {
		if !isValidIPv4(*spsController.Gateway) {
			ve = ve.Add("spscontroller.gateway", "gateway must be a valid IPv4 address")
		}
	}
	if spsController.Subnet != nil && strings.TrimSpace(*spsController.Subnet) != "" {
		if !isValidSubnetMask(*spsController.Subnet) {
			ve = ve.Add("spscontroller.subnet", "subnet must be a valid IPv4 subnet mask")
		}
	}
	if spsController.Vlan != nil && strings.TrimSpace(*spsController.Vlan) != "" {
		vlanValue := strings.TrimSpace(*spsController.Vlan)
		vlan, err := strconv.Atoi(vlanValue)
		if err != nil || vlan < 1 || vlan > 4094 {
			ve = ve.Add("spscontroller.vlan", "vlan must be a number between 1 and 4094")
		}
	}

	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func isValidIPv4(value string) bool {
	ip := net.ParseIP(strings.TrimSpace(value))
	return ip != nil && ip.To4() != nil
}

func isValidSubnetMask(value string) bool {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return false
	}
	mask := net.IPMask(ip.To4())
	if mask == nil {
		return false
	}
	ones, bits := mask.Size()
	return bits == 32 && ones > 0
}
