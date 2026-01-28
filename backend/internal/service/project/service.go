package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	repo                      domainProject.ProjectRepository
	projectControlCabinetRepo domainProject.ProjectControlCabinetRepository
	projectSPSControllerRepo  domainProject.ProjectSPSControllerRepository
	projectFieldDeviceRepo    domainProject.ProjectFieldDeviceRepository
	userRepo                  domainUser.UserRepository
	objectDataRepo            domainFacility.ObjectDataStore
	bacnetObjectRepo          domainFacility.BacnetObjectStore
}

func New(
	repo domainProject.ProjectRepository,
	projectControlCabinetRepo domainProject.ProjectControlCabinetRepository,
	projectSPSControllerRepo domainProject.ProjectSPSControllerRepository,
	projectFieldDeviceRepo domainProject.ProjectFieldDeviceRepository,
	userRepo domainUser.UserRepository,
	objectDataRepo domainFacility.ObjectDataStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
) *Service {
	return &Service{
		repo:                      repo,
		projectControlCabinetRepo: projectControlCabinetRepo,
		projectSPSControllerRepo:  projectSPSControllerRepo,
		projectFieldDeviceRepo:    projectFieldDeviceRepo,
		userRepo:                  userRepo,
		objectDataRepo:            objectDataRepo,
		bacnetObjectRepo:          bacnetObjectRepo,
	}
}

func (s *Service) Create(project *domainProject.Project) error {
	if project.Status == "" {
		project.Status = domainProject.StatusPlanned
	}

	if err := s.repo.Create(project); err != nil {
		return err
	}

	// Copy ObjectData templates
	templates, err := s.objectDataRepo.GetTemplates()
	if err != nil {
		return err
	}

	for _, tmpl := range templates {
		copy := *tmpl
		copy.ID = uuid.Nil
		copy.ProjectID = &project.ID
		copy.BacnetObjects = nil // clear for now, we will rebuild them

		if err := s.objectDataRepo.Create(&copy); err != nil {
			return err
		}

		// Deep copy BacnetObjects
		if len(tmpl.BacnetObjects) == 0 {
			continue
		}

		// Map old ID -> new Instance
		oldToNew := make(map[uuid.UUID]*domainFacility.BacnetObject)
		// Map old ID -> old SoftwareReferenceID (for second pass)
		oldRefs := make(map[uuid.UUID]*uuid.UUID)

		// 1. Create clones
		for _, bo := range tmpl.BacnetObjects {
			newBO := &domainFacility.BacnetObject{
				TextFix:             bo.TextFix,
				Description:         bo.Description,
				GMSVisible:          bo.GMSVisible,
				Optional:            bo.Optional,
				TextIndividual:      bo.TextIndividual,
				SoftwareType:        bo.SoftwareType,
				SoftwareNumber:      bo.SoftwareNumber,
				HardwareType:        bo.HardwareType,
				HardwareQuantity:    bo.HardwareQuantity,
				StateTextID:         bo.StateTextID,
				NotificationClassID: bo.NotificationClassID,
				AlarmDefinitionID:   bo.AlarmDefinitionID,
				// FieldDeviceID is NULL for ObjectData templates
			}
			if err := s.bacnetObjectRepo.Create(newBO); err != nil {
				return err
			}
			oldToNew[bo.ID] = newBO
			oldRefs[bo.ID] = bo.SoftwareReferenceID
		}

		// 2. Fix references and link to new ObjectData
		newBacnetObjects := make([]*domainFacility.BacnetObject, 0, len(tmpl.BacnetObjects))
		for oldID, newBO := range oldToNew {
			// Fix reference
			if refID := oldRefs[oldID]; refID != nil {
				if target, ok := oldToNew[*refID]; ok {
					id := target.ID
					newBO.SoftwareReferenceID = &id
					if err := s.bacnetObjectRepo.Update(newBO); err != nil {
						return err
					}
				}
			}
			newBacnetObjects = append(newBacnetObjects, newBO)
		}

		// 3. Associate with ObjectData
		copy.BacnetObjects = newBacnetObjects
		if err := s.objectDataRepo.Update(&copy); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) CreateControlCabinet(projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	entity := &domainProject.ProjectControlCabinet{
		ProjectID:        projectID,
		ControlCabinetID: controlCabinetID,
	}
	if err := s.projectControlCabinetRepo.Create(entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) UpdateControlCabinet(linkID, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	items, err := s.projectControlCabinetRepo.GetByIds([]uuid.UUID{linkID})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 || items[0].ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	entity := items[0]
	entity.ControlCabinetID = controlCabinetID
	if err := s.projectControlCabinetRepo.Update(entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) DeleteControlCabinet(linkID, projectID uuid.UUID) error {
	items, err := s.projectControlCabinetRepo.GetByIds([]uuid.UUID{linkID})
	if err != nil {
		return err
	}
	if len(items) == 0 || items[0].ProjectID != projectID {
		return domain.ErrNotFound
	}
	return s.projectControlCabinetRepo.DeleteByIds([]uuid.UUID{linkID})
}

func (s *Service) CreateSPSController(projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	entity := &domainProject.ProjectSPSController{
		ProjectID:       projectID,
		SPSControllerID: spsControllerID,
	}
	if err := s.projectSPSControllerRepo.Create(entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) UpdateSPSController(linkID, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	items, err := s.projectSPSControllerRepo.GetByIds([]uuid.UUID{linkID})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 || items[0].ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	entity := items[0]
	entity.SPSControllerID = spsControllerID
	if err := s.projectSPSControllerRepo.Update(entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) DeleteSPSController(linkID, projectID uuid.UUID) error {
	items, err := s.projectSPSControllerRepo.GetByIds([]uuid.UUID{linkID})
	if err != nil {
		return err
	}
	if len(items) == 0 || items[0].ProjectID != projectID {
		return domain.ErrNotFound
	}
	return s.projectSPSControllerRepo.DeleteByIds([]uuid.UUID{linkID})
}

func (s *Service) CreateFieldDevice(projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error) {
	entity := &domainProject.ProjectFieldDevice{
		ProjectID:     projectID,
		FieldDeviceID: fieldDeviceID,
	}
	if err := s.projectFieldDeviceRepo.Create(entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) InviteUser(projectID, userID uuid.UUID) error {
	projects, err := s.repo.GetByIds([]uuid.UUID{projectID})
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		return domain.ErrNotFound
	}
	users, err := s.userRepo.GetByIds([]uuid.UUID{userID})
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return domain.ErrNotFound
	}
	return s.repo.AddUser(projectID, userID)
}

func (s *Service) ListUsers(projectID uuid.UUID) ([]domainUser.User, error) {
	projects, err := s.repo.GetByIds([]uuid.UUID{projectID})
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return nil, domain.ErrNotFound
	}
	return s.repo.ListUsers(projectID)
}

func (s *Service) RemoveUser(projectID, userID uuid.UUID) error {
	projects, err := s.repo.GetByIds([]uuid.UUID{projectID})
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		return domain.ErrNotFound
	}
	users, err := s.userRepo.GetByIds([]uuid.UUID{userID})
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return domain.ErrNotFound
	}
	return s.repo.RemoveUser(projectID, userID)
}

func (s *Service) UpdateFieldDevice(linkID, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error) {
	items, err := s.projectFieldDeviceRepo.GetByIds([]uuid.UUID{linkID})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 || items[0].ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	entity := items[0]
	entity.FieldDeviceID = fieldDeviceID
	if err := s.projectFieldDeviceRepo.Update(entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) DeleteFieldDevice(linkID, projectID uuid.UUID) error {
	items, err := s.projectFieldDeviceRepo.GetByIds([]uuid.UUID{linkID})
	if err != nil {
		return err
	}
	if len(items) == 0 || items[0].ProjectID != projectID {
		return domain.ErrNotFound
	}
	return s.projectFieldDeviceRepo.DeleteByIds([]uuid.UUID{linkID})
}

func (s *Service) AddObjectData(projectID, objectDataID uuid.UUID) (*domainFacility.ObjectData, error) {
	projects, err := s.repo.GetByIds([]uuid.UUID{projectID})
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return nil, domain.ErrNotFound
	}
	objects, err := s.objectDataRepo.GetByIds([]uuid.UUID{objectDataID})
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, domain.ErrNotFound
	}
	obj := objects[0]
	if obj.ProjectID != nil && *obj.ProjectID != projectID {
		return nil, domain.ErrConflict
	}
	obj.ProjectID = &projectID
	if err := s.objectDataRepo.Update(obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *Service) RemoveObjectData(projectID, objectDataID uuid.UUID) (*domainFacility.ObjectData, error) {
	projects, err := s.repo.GetByIds([]uuid.UUID{projectID})
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return nil, domain.ErrNotFound
	}
	objects, err := s.objectDataRepo.GetByIds([]uuid.UUID{objectDataID})
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, domain.ErrNotFound
	}
	obj := objects[0]
	if obj.ProjectID == nil || *obj.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	obj.ProjectID = nil
	if err := s.objectDataRepo.Update(obj); err != nil {
		return nil, err
	}
	return obj, nil
}
func (s *Service) GetByIds(ids []uuid.UUID) ([]*domainProject.Project, error) {
	return s.repo.GetByIds(ids)
}

func (s *Service) GetByID(id uuid.UUID) (*domainProject.Project, error) {
	projects, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return nil, domain.ErrNotFound
	}
	return projects[0], nil
}

func (s *Service) Update(project *domainProject.Project) error {
	return s.repo.Update(project)
}

func (s *Service) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}

func (s *Service) List(page, limit int, search string) (*domain.PaginatedList[domainProject.Project], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *Service) ListControlCabinets(projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectControlCabinet], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.projectControlCabinetRepo.GetPaginatedListByProjectID(projectID, domain.PaginationParams{Page: page, Limit: limit})
}

func (s *Service) ListSPSControllers(projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectSPSController], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.projectSPSControllerRepo.GetPaginatedListByProjectID(projectID, domain.PaginationParams{Page: page, Limit: limit})
}

func (s *Service) ListFieldDevices(projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectFieldDevice], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.projectFieldDeviceRepo.GetPaginatedListByProjectID(projectID, domain.PaginationParams{Page: page, Limit: limit})
}

func (s *Service) ListObjectData(projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.objectDataRepo.GetPaginatedListForProject(projectID, domain.PaginationParams{Page: page, Limit: limit})
}
