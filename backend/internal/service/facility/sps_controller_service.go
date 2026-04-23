package facility

import (
	"context"
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
	fieldDeviceRepo          domainFacility.FieldDeviceStore
	hierarchyCopier          *HierarchyCopier
}

func NewSPSControllerService(
	repo domainFacility.SPSControllerRepository,
	controlCabinetRepo domainFacility.ControlCabinetRepository,
	systemTypeRepo domainFacility.SystemTypeRepository,
	spsControllerSystemTypeStore domainFacility.SPSControllerSystemTypeStore,
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	hierarchyCopier *HierarchyCopier,
) *SPSControllerService {
	return &SPSControllerService{
		repo:                     repo,
		controlCabinetRepo:       controlCabinetRepo,
		systemTypeRepo:           systemTypeRepo,
		spsControllerSystemTyper: spsControllerSystemTypeStore,
		fieldDeviceRepo:          fieldDeviceRepo,
		hierarchyCopier:          hierarchyCopier,
	}
}

func (s *SPSControllerService) Create(ctx context.Context, spsController *domainFacility.SPSController) error {
	return s.CreateWithSystemTypes(ctx, spsController, nil)
}

func (s *SPSControllerService) CreateWithSystemTypes(ctx context.Context, spsController *domainFacility.SPSController, systemTypes []domainFacility.SPSControllerSystemType) error {
	if err := s.ensureGADeviceAssigned(ctx, spsController, nil); err != nil {
		return err
	}
	if err := s.Validate(ctx, spsController, nil); err != nil {
		return err
	}
	systemTypeMap, err := s.loadSystemTypes(ctx, systemTypes)
	if err != nil {
		return err
	}
	if err := s.assignSystemTypeNumbers(systemTypes, systemTypeMap); err != nil {
		return err
	}

	if err := s.repo.Create(ctx, spsController); err != nil {
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
		if err := s.spsControllerSystemTyper.Create(ctx, entity); err != nil {
			return err
		}
	}
	return nil
}

func (s *SPSControllerService) GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSController, error) {
	return domain.GetByID(ctx, s.repo, id)
}

func (s *SPSControllerService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domainFacility.SPSController, error) {
	spsControllers, err := s.repo.GetByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	return derefSlice(spsControllers), nil
}

func (s *SPSControllerService) CopyByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSController, error) {
	return s.hierarchyCopier.CopySPSControllerByID(ctx, id)
}

func (s *SPSControllerService) List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSController], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(ctx, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SPSControllerService) ListByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSController], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedListByControlCabinetID(ctx, controlCabinetID, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SPSControllerService) Update(ctx context.Context, spsController *domainFacility.SPSController) error {
	if err := s.Validate(ctx, spsController, &spsController.ID); err != nil {
		return err
	}
	return s.repo.Update(ctx, spsController)
}

func (s *SPSControllerService) UpdateWithSystemTypes(ctx context.Context, spsController *domainFacility.SPSController, systemTypes []domainFacility.SPSControllerSystemType) error {
	if err := s.Validate(ctx, spsController, &spsController.ID); err != nil {
		return err
	}
	systemTypeMap, err := s.loadSystemTypes(ctx, systemTypes)
	if err != nil {
		return err
	}
	if err := s.assignSystemTypeNumbers(systemTypes, systemTypeMap); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, spsController); err != nil {
		return err
	}

	existing, err := s.spsControllerSystemTyper.ListBySPSControllerID(ctx, spsController.ID)
	if err != nil {
		return err
	}

	existingByID := make(map[uuid.UUID]*domainFacility.SPSControllerSystemType, len(existing))
	existingBySystemType := make(map[uuid.UUID]*domainFacility.SPSControllerSystemType, len(existing))
	for _, item := range existing {
		existingByID[item.ID] = item
		if _, ok := existingBySystemType[item.SystemTypeID]; !ok {
			existingBySystemType[item.SystemTypeID] = item
		}
	}

	incomingIDs := make(map[uuid.UUID]struct{}, len(systemTypes))
	incomingSystemTypeIDs := make(map[uuid.UUID]struct{}, len(systemTypes))
	hasIncomingIDs := false
	for _, st := range systemTypes {
		if st.ID != uuid.Nil {
			hasIncomingIDs = true
			break
		}
	}

	for _, st := range systemTypes {
		if st.ID != uuid.Nil {
			incomingIDs[st.ID] = struct{}{}
		}
		incomingSystemTypeIDs[st.SystemTypeID] = struct{}{}

		if st.ID != uuid.Nil {
			if existingItem, ok := existingByID[st.ID]; ok {
				existingItem.SystemTypeID = st.SystemTypeID
				existingItem.Number = st.Number
				existingItem.DocumentName = st.DocumentName
				if err := s.spsControllerSystemTyper.Update(ctx, existingItem); err != nil {
					return err
				}
				continue
			}
		}

		if !hasIncomingIDs {
			if existingItem, ok := existingBySystemType[st.SystemTypeID]; ok {
				existingItem.Number = st.Number
				existingItem.DocumentName = st.DocumentName
				if err := s.spsControllerSystemTyper.Update(ctx, existingItem); err != nil {
					return err
				}
				continue
			}
		}

		entity := &domainFacility.SPSControllerSystemType{
			Number:          st.Number,
			DocumentName:    st.DocumentName,
			SPSControllerID: spsController.ID,
			SystemTypeID:    st.SystemTypeID,
		}
		if err := s.spsControllerSystemTyper.Create(ctx, entity); err != nil {
			return err
		}
	}

	var deleteIDs []uuid.UUID
	if hasIncomingIDs {
		for _, item := range existing {
			if _, ok := incomingIDs[item.ID]; !ok {
				deleteIDs = append(deleteIDs, item.ID)
			}
		}
	} else {
		for _, item := range existing {
			if _, ok := incomingSystemTypeIDs[item.SystemTypeID]; !ok {
				deleteIDs = append(deleteIDs, item.ID)
			}
		}
	}

	if len(deleteIDs) > 0 {
		fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, deleteIDs)
		if err != nil {
			return err
		}
		if len(fieldDeviceIDs) > 0 {
			return domain.NewValidationError().Add("spscontroller.system_types", "referenced_entity_in_use")
		}
		if err := s.spsControllerSystemTyper.DeleteByIds(ctx, deleteIDs); err != nil {
			return err
		}
	}

	return nil
}

func (s *SPSControllerService) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteByIds(ctx, []uuid.UUID{id})
}
func (s *SPSControllerService) ensureControlCabinetExists(ctx context.Context, controlCabinetID uuid.UUID) error {
	return domain.EnsureReferenceExists(ctx, s.controlCabinetRepo, controlCabinetID)
}

func (s *SPSControllerService) loadSystemTypes(ctx context.Context, systemTypes []domainFacility.SPSControllerSystemType) (map[uuid.UUID]domainFacility.SystemType, error) {
	return loadSystemTypeDefinitions(ctx, s.systemTypeRepo, systemTypes)
}

func (s *SPSControllerService) assignSystemTypeNumbers(systemTypes []domainFacility.SPSControllerSystemType, systemTypeMap map[uuid.UUID]domainFacility.SystemType) error {
	return assignSystemTypeNumbers(systemTypes, systemTypeMap)
}

func findLowestAvailableNumber(min, max int, used map[int]struct{}) (int, bool) {
	for i := min; i <= max; i++ {
		if _, exists := used[i]; !exists {
			return i, true
		}
	}
	return 0, false
}

func (s *SPSControllerService) GetNextGADevice(ctx context.Context, controlCabinetID uuid.UUID) (string, error) {
	if controlCabinetID == uuid.Nil {
		return "", domain.ErrInvalidArgument
	}
	if err := s.ensureControlCabinetExists(ctx, controlCabinetID); err != nil {
		return "", err
	}
	return s.nextAvailableGADevice(ctx, controlCabinetID)
}

func (s *SPSControllerService) ensureGADeviceAssigned(ctx context.Context, spsController *domainFacility.SPSController, excludeID *uuid.UUID) error {
	if spsController.GADevice != nil && strings.TrimSpace(*spsController.GADevice) != "" {
		return nil
	}
	if err := s.ensureControlCabinetExists(ctx, spsController.ControlCabinetID); err != nil {
		return err
	}

	next, err := s.nextAvailableGADevice(ctx, spsController.ControlCabinetID)
	if err != nil {
		return err
	}
	if next == "" {
		return domain.NewValidationError().Add("spscontroller.ga_device", "no available ga_device for control cabinet")
	}
	spsController.GADevice = &next

	// Always ensure uniqueness, not just when updating
	if err := s.ensureUnique(ctx, spsController, excludeID); err != nil {
		return err
	}
	return nil
}

func (s *SPSControllerService) nextAvailableGADevice(ctx context.Context, controlCabinetID uuid.UUID) (string, error) {
	devices, err := s.repo.ListGADevicesByControlCabinetID(ctx, controlCabinetID)
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
	builder := domain.NewValidationBuilder()
	spsControllerControlCabinetField.RequireUUID(builder, spsController.ControlCabinetID)
	spsControllerDeviceNameField.RequireTrimmed(builder, spsController.DeviceName)
	if spsController.GADevice == nil || strings.TrimSpace(*spsController.GADevice) == "" {
		spsControllerGADeviceField.Add(builder, "ga_device is required")
	} else if !isValidGADevice(*spsController.GADevice) {
		spsControllerGADeviceField.Add(builder, "ga_device must be exactly 3 uppercase letters (A-Z)")
	}

	if err := validateNetworkFields(spsController); err != nil {
		if mergeErr := builder.Merge(err); mergeErr != nil {
			return mergeErr
		}
	}
	return builder.Err()
}

func (s *SPSControllerService) Validate(ctx context.Context, spsController *domainFacility.SPSController, excludeID *uuid.UUID) error {
	if err := s.validateRequiredFields(spsController); err != nil {
		return err
	}
	if err := s.ensureControlCabinetExists(ctx, spsController.ControlCabinetID); err != nil {
		return err
	}
	if err := s.ensureUnique(ctx, spsController, excludeID); err != nil {
		return err
	}
	return nil
}

func (s *SPSControllerService) NextAvailableGADevice(ctx context.Context, controlCabinetID uuid.UUID, excludeID *uuid.UUID) (string, error) {
	if err := s.ensureControlCabinetExists(ctx, controlCabinetID); err != nil {
		return "", err
	}

	items, err := s.repo.ListGADevicesByControlCabinetID(ctx, controlCabinetID)
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
		controller, err := domain.GetByID(ctx, s.repo, *excludeID)
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

func (s *SPSControllerService) ensureUnique(ctx context.Context, spsController *domainFacility.SPSController, excludeID *uuid.UUID) error {
	builder := domain.NewValidationBuilder()

	if strings.TrimSpace(spsController.DeviceName) != "" && spsController.ControlCabinetID != uuid.Nil {
		exists, err := s.repo.ExistsDeviceName(ctx, spsController.ControlCabinetID, spsController.DeviceName, excludeID)
		if err != nil {
			return err
		}
		if exists {
			spsControllerDeviceNameField.UniqueWithin(builder, controlCabinetScope)
		}
	}

	if spsController.GADevice != nil && strings.TrimSpace(*spsController.GADevice) != "" && spsController.ControlCabinetID != uuid.Nil {
		exists, err := s.repo.ExistsGADevice(ctx, spsController.ControlCabinetID, *spsController.GADevice, excludeID)
		if err != nil {
			return err
		}
		if exists {
			spsControllerGADeviceField.UniqueWithin(builder, controlCabinetScope)
		}
	}

	if spsController.IPAddress != nil && spsController.Vlan != nil {
		ip := strings.TrimSpace(*spsController.IPAddress)
		vlan := strings.TrimSpace(*spsController.Vlan)
		if ip != "" && vlan != "" {
			exists, err := s.repo.ExistsIPAddressVlan(ctx, ip, vlan, excludeID)
			if err != nil {
				return err
			}
			if exists {
				spsControllerIPAddressField.Add(builder, "ip_address must be unique per vlan")
				spsControllerVlanField.Add(builder, "vlan must be unique per ip_address")
			}
		}
	}

	return builder.Err()
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
