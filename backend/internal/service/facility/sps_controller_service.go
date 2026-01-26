package facility

import (
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
	if err := s.validateRequiredFields(spsController); err != nil {
		return err
	}
	if err := s.ensureUnique(spsController, nil); err != nil {
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

func (s *SPSControllerService) Update(spsController *domainFacility.SPSController) error {
	if err := s.validateRequiredFields(spsController); err != nil {
		return err
	}
	if err := s.ensureUnique(spsController, &spsController.ID); err != nil {
		return err
	}
	if err := s.ensureControlCabinetExists(spsController.ControlCabinetID); err != nil {
		return err
	}
	return s.repo.Update(spsController)
}

func (s *SPSControllerService) UpdateWithSystemTypes(spsController *domainFacility.SPSController, systemTypes []domainFacility.SPSControllerSystemType) error {
	if err := s.validateRequiredFields(spsController); err != nil {
		return err
	}
	if err := s.ensureUnique(spsController, &spsController.ID); err != nil {
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

func (s *SPSControllerService) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
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
		ve.Add("spscontroller.control_cabinet_id", "control_cabinet_id is required")
	}
	if strings.TrimSpace(spsController.DeviceName) == "" {
		ve.Add("spscontroller.device_name", "device_name is required")
	}
	if len(ve.Fields) > 0 {
		return ve
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
				ve.Add("spscontroller.device_name", "device_name must be unique within the control cabinet")
				break
			}
		}
	}

	if spsController.IPAddress != nil && spsController.Vlan != nil {
		ip := strings.TrimSpace(*spsController.IPAddress)
		vlan := strings.TrimSpace(*spsController.Vlan)
		if ip != "" && vlan != "" {
			items, err := s.repo.GetPaginatedList(domain.PaginationParams{Page: 1, Limit: 1000, Search: ip})
			if err != nil {
				return err
			}
			for i := range items.Items {
				item := items.Items[i]
				if excludeID != nil && item.ID == *excludeID {
					continue
				}
				if item.IPAddress != nil && item.Vlan != nil && *item.IPAddress == ip && *item.Vlan == vlan {
					ve.Add("spscontroller.ip_address", "ip_address must be unique per vlan")
					ve.Add("spscontroller.vlan", "vlan must be unique per ip_address")
					break
				}
			}
		}
	}

	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}
