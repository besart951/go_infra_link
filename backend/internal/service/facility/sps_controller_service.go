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
	if err := s.Validate(spsController, nil); err != nil {
		return err
	}
	if err := s.ensureControlCabinetExists(spsController.ControlCabinetID); err != nil {
		return err
	}
	if err := s.ensureSystemTypesExist(systemTypes); err != nil {
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
	if err := s.ensureSystemTypesExist(systemTypes); err != nil {
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

func (s *SPSControllerService) ensureSystemTypesExist(systemTypes []domainFacility.SPSControllerSystemType) error {
	if len(systemTypes) == 0 {
		return nil
	}

	unique := make(map[uuid.UUID]struct{}, len(systemTypes))
	ids := make([]uuid.UUID, 0, len(systemTypes))
	for _, st := range systemTypes {
		if st.SystemTypeID == uuid.Nil {
			return domain.ErrNotFound
		}
		if _, ok := unique[st.SystemTypeID]; ok {
			continue
		}
		unique[st.SystemTypeID] = struct{}{}
		ids = append(ids, st.SystemTypeID)
	}

	found, err := s.systemTypeRepo.GetByIds(ids)
	if err != nil {
		return err
	}
	if len(found) != len(ids) {
		return domain.ErrNotFound
	}
	return nil
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
