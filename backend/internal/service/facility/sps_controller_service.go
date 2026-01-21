package facility

import (
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
	page, limit = normalizePagination(page, limit)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SPSControllerService) Update(spsController *domainFacility.SPSController) error {
	if err := s.ensureControlCabinetExists(spsController.ControlCabinetID); err != nil {
		return err
	}
	return s.repo.Update(spsController)
}

func (s *SPSControllerService) UpdateWithSystemTypes(spsController *domainFacility.SPSController, systemTypes []domainFacility.SPSControllerSystemType) error {
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
