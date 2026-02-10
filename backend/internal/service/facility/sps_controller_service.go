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
	systemTypeRepo           domainFacility.SystemTypeRepository
	spsControllerSystemTyper domainFacility.SPSControllerSystemTypeStore
}

func NewSPSControllerService(
	repo domainFacility.SPSControllerRepository,
	controlCabinetRepo domainFacility.ControlCabinetRepository,
	systemTypeRepo domainFacility.SystemTypeRepository,
	spsControllerSystemTypeStore domainFacility.SPSControllerSystemTypeStore,
) *SPSControllerService {
	return &SPSControllerService{
		repo:                     repo,
		controlCabinetRepo:       controlCabinetRepo,
		systemTypeRepo:           systemTypeRepo,
		spsControllerSystemTyper: spsControllerSystemTypeStore,
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
	spsControllers, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(spsControllers) == 0 {
		return nil, domain.ErrNotFound
	}
	return spsControllers[0], nil
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

	if err := s.spsControllerSystemTyper.SoftDeleteBySPSControllerIDs([]uuid.UUID{spsController.ID}); err != nil {
		return err
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

func (s *SPSControllerService) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}

func (s *SPSControllerService) ensureControlCabinetExists(controlCabinetID uuid.UUID) error {
	controlCabinets, err := s.controlCabinetRepo.GetByIds([]uuid.UUID{controlCabinetID})
	if err != nil {
		return err
	}
	if len(controlCabinets) == 0 {
		return domain.ErrNotFound
	}
	return nil
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

	if excludeID != nil {
		if err := s.ensureUnique(spsController, excludeID); err != nil {
			return err
		}
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

func gaDeviceFromIndex(index int) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	first := index % 26
	second := (index / 26) % 26
	third := (index / (26 * 26)) % 26
	return string([]byte{alphabet[first], alphabet[second], alphabet[third]})
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
		controllers, err := s.repo.GetByIds([]uuid.UUID{*excludeID})
		if err != nil {
			return "", err
		}
		if len(controllers) == 0 {
			return "", domain.ErrNotFound
		}
		if controllers[0].GADevice != nil {
			current := strings.ToUpper(strings.TrimSpace(*controllers[0].GADevice))
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
