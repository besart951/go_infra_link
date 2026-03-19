package project

import (
	"errors"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	repo                      domainProject.ProjectRepository
	projectControlCabinetRepo domainProject.ProjectControlCabinetRepository
	projectSPSControllerRepo  domainProject.ProjectSPSControllerRepository
	projectFieldDeviceRepo    domainProject.ProjectFieldDeviceRepository
	userRepo                  domainUser.UserRepository
	objectDataRepo            domainFacility.ObjectDataStore
	bacnetObjectRepo          domainFacility.BacnetObjectStore
	specificationRepo         domainFacility.SpecificationStore
	controlCabinetRepo        domainFacility.ControlCabinetRepository
	spsControllerRepo         domainFacility.SPSControllerRepository
	spsControllerSystemRepo   domainFacility.SPSControllerSystemTypeStore
	fieldDeviceRepo           domainFacility.FieldDeviceStore
	hierarchyCopier           *facilityservice.HierarchyCopier
}

func New(
	repo domainProject.ProjectRepository,
	projectControlCabinetRepo domainProject.ProjectControlCabinetRepository,
	projectSPSControllerRepo domainProject.ProjectSPSControllerRepository,
	projectFieldDeviceRepo domainProject.ProjectFieldDeviceRepository,
	userRepo domainUser.UserRepository,
	objectDataRepo domainFacility.ObjectDataStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	specificationRepo domainFacility.SpecificationStore,
	controlCabinetRepo domainFacility.ControlCabinetRepository,
	spsControllerRepo domainFacility.SPSControllerRepository,
	spsControllerSystemRepo domainFacility.SPSControllerSystemTypeStore,
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	hierarchyCopier *facilityservice.HierarchyCopier,
) *Service {
	return &Service{
		repo:                      repo,
		projectControlCabinetRepo: projectControlCabinetRepo,
		projectSPSControllerRepo:  projectSPSControllerRepo,
		projectFieldDeviceRepo:    projectFieldDeviceRepo,
		userRepo:                  userRepo,
		objectDataRepo:            objectDataRepo,
		bacnetObjectRepo:          bacnetObjectRepo,
		specificationRepo:         specificationRepo,
		controlCabinetRepo:        controlCabinetRepo,
		spsControllerRepo:         spsControllerRepo,
		spsControllerSystemRepo:   spsControllerSystemRepo,
		fieldDeviceRepo:           fieldDeviceRepo,
		hierarchyCopier:           hierarchyCopier,
	}
}
func (s *Service) Create(project *domainProject.Project) error {
	if project.Status == "" {
		project.Status = domainProject.StatusPlanned
	}

	if err := s.repo.Create(project); err != nil {
		return err
	}

	if project.CreatorID != uuid.Nil {
		if err := s.repo.AddUser(project.ID, project.CreatorID); err != nil {
			return err
		}
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
				AlarmTypeID:         bo.AlarmTypeID,
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

	if err := s.linkDescendantsForControlCabinet(projectID, controlCabinetID); err != nil {
		_ = s.cleanupProjectLinksForControlCabinetHierarchy(controlCabinetID)
		return nil, err
	}

	return entity, nil
}

func (s *Service) CopyControlCabinet(projectID, controlCabinetID uuid.UUID) (*domainFacility.ControlCabinet, error) {
	copyEntity, err := s.hierarchyCopier.CopyControlCabinetByID(controlCabinetID)
	if err != nil {
		return nil, err
	}

	if _, err := s.CreateControlCabinet(projectID, copyEntity.ID); err != nil {
		_ = s.rollbackCopiedControlCabinet(copyEntity.ID)
		return nil, err
	}

	return copyEntity, nil
}
func (s *Service) UpdateControlCabinet(linkID, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	entity, err := domain.GetByID(s.projectControlCabinetRepo, linkID)
	if err != nil {
		return nil, err
	}
	if entity.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	previousControlCabinetID := entity.ControlCabinetID
	entity.ControlCabinetID = controlCabinetID
	if err := s.projectControlCabinetRepo.Update(entity); err != nil {
		return nil, err
	}

	if err := s.linkDescendantsForControlCabinet(projectID, controlCabinetID); err != nil {
		_ = s.cleanupProjectLinksForControlCabinetHierarchy(controlCabinetID)
		entity.ControlCabinetID = previousControlCabinetID
		_ = s.projectControlCabinetRepo.Update(entity)
		return nil, err
	}

	return entity, nil
}

func (s *Service) DeleteControlCabinet(linkID, projectID uuid.UUID) error {
	entity, err := domain.GetByID(s.projectControlCabinetRepo, linkID)
	if err != nil {
		return err
	}
	if entity.ProjectID != projectID {
		return domain.ErrNotFound
	}

	controlCabinetID := entity.ControlCabinetID
	spsControllerIDs, _, fieldDeviceIDs, err := s.collectDescendantIDsForControlCabinet(controlCabinetID)
	if err != nil {
		return err
	}

	if err := s.deleteProjectControlCabinetLinksByControlCabinetIDs([]uuid.UUID{controlCabinetID}); err != nil {
		return err
	}
	if err := s.deleteProjectSPSControllerLinksBySPSControllerIDs(spsControllerIDs); err != nil {
		return err
	}
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}

	if err := s.deleteFieldDevicesWithChildren(fieldDeviceIDs); err != nil {
		return err
	}
	if len(spsControllerIDs) > 0 {
		if err := s.spsControllerSystemRepo.DeleteBySPSControllerIDs(spsControllerIDs); err != nil {
			return err
		}
		if err := s.spsControllerRepo.DeleteByIds(spsControllerIDs); err != nil {
			return err
		}
	}

	return s.controlCabinetRepo.DeleteByIds([]uuid.UUID{controlCabinetID})
}

func (s *Service) CreateSPSController(projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	entity := &domainProject.ProjectSPSController{
		ProjectID:       projectID,
		SPSControllerID: spsControllerID,
	}
	if err := s.projectSPSControllerRepo.Create(entity); err != nil {
		return nil, err
	}

	if err := s.linkDescendantsForSPSControllers(projectID, []uuid.UUID{spsControllerID}); err != nil {
		_ = s.cleanupProjectLinksForSPSControllers([]uuid.UUID{spsControllerID})
		return nil, err
	}

	return entity, nil
}

func (s *Service) CopySPSController(projectID, spsControllerID uuid.UUID) (*domainFacility.SPSController, error) {
	copyEntity, err := s.hierarchyCopier.CopySPSControllerByID(spsControllerID)
	if err != nil {
		return nil, err
	}

	if _, err := s.CreateSPSController(projectID, copyEntity.ID); err != nil {
		_ = s.rollbackCopiedSPSController(copyEntity.ID)
		return nil, err
	}

	return copyEntity, nil
}

func (s *Service) CopySPSControllerSystemType(projectID, systemTypeID uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	copyEntity, err := s.hierarchyCopier.CopySPSControllerSystemTypeByID(systemTypeID)
	if err != nil {
		return nil, err
	}

	if err := s.linkFieldDevicesForSystemTypes(projectID, []uuid.UUID{copyEntity.ID}); err != nil {
		_ = s.cleanupProjectLinksForSystemTypes([]uuid.UUID{copyEntity.ID})
		_ = s.rollbackCopiedSPSControllerSystemType(copyEntity.ID)
		return nil, err
	}

	return copyEntity, nil
}

func (s *Service) UpdateSPSController(linkID, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	entity, err := domain.GetByID(s.projectSPSControllerRepo, linkID)
	if err != nil {
		return nil, err
	}
	if entity.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	previousSPSControllerID := entity.SPSControllerID
	entity.SPSControllerID = spsControllerID
	if err := s.projectSPSControllerRepo.Update(entity); err != nil {
		return nil, err
	}

	if err := s.linkDescendantsForSPSControllers(projectID, []uuid.UUID{spsControllerID}); err != nil {
		_ = s.cleanupProjectLinksForSPSControllers([]uuid.UUID{spsControllerID})
		entity.SPSControllerID = previousSPSControllerID
		_ = s.projectSPSControllerRepo.Update(entity)
		return nil, err
	}

	return entity, nil
}

func (s *Service) DeleteSPSController(linkID, projectID uuid.UUID) error {
	entity, err := domain.GetByID(s.projectSPSControllerRepo, linkID)
	if err != nil {
		return err
	}
	if entity.ProjectID != projectID {
		return domain.ErrNotFound
	}

	spsControllerID := entity.SPSControllerID
	_, fieldDeviceIDs, err := s.collectDescendantIDsForSPSControllers([]uuid.UUID{spsControllerID})
	if err != nil {
		return err
	}

	if err := s.deleteProjectSPSControllerLinksBySPSControllerIDs([]uuid.UUID{spsControllerID}); err != nil {
		return err
	}
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}

	if err := s.deleteFieldDevicesWithChildren(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.spsControllerSystemRepo.DeleteBySPSControllerIDs([]uuid.UUID{spsControllerID}); err != nil {
		return err
	}
	return s.spsControllerRepo.DeleteByIds([]uuid.UUID{spsControllerID})
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
	if _, err := domain.GetByID(s.repo, projectID); err != nil {
		return err
	}
	if _, err := domain.GetByID(s.userRepo, userID); err != nil {
		return err
	}
	return s.repo.AddUser(projectID, userID)
}

func (s *Service) ListUsers(projectID uuid.UUID) ([]domainUser.User, error) {
	if _, err := domain.GetByID(s.repo, projectID); err != nil {
		return nil, err
	}
	return s.repo.ListUsers(projectID)
}

func (s *Service) RemoveUser(projectID, userID uuid.UUID) error {
	if _, err := domain.GetByID(s.repo, projectID); err != nil {
		return err
	}
	if _, err := domain.GetByID(s.userRepo, userID); err != nil {
		return err
	}
	return s.repo.RemoveUser(projectID, userID)
}

func (s *Service) UpdateFieldDevice(linkID, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error) {
	entity, err := domain.GetByID(s.projectFieldDeviceRepo, linkID)
	if err != nil {
		return nil, err
	}
	if entity.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	entity.FieldDeviceID = fieldDeviceID
	if err := s.projectFieldDeviceRepo.Update(entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) DeleteFieldDevice(linkID, projectID uuid.UUID) error {
	entity, err := domain.GetByID(s.projectFieldDeviceRepo, linkID)
	if err != nil {
		return err
	}
	if entity.ProjectID != projectID {
		return domain.ErrNotFound
	}

	fieldDeviceID := entity.FieldDeviceID
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs([]uuid.UUID{fieldDeviceID}); err != nil {
		return err
	}
	return s.deleteFieldDevicesWithChildren([]uuid.UUID{fieldDeviceID})
}

func (s *Service) AddObjectData(projectID, objectDataID uuid.UUID) (*domainFacility.ObjectData, error) {
	if _, err := domain.GetByID(s.repo, projectID); err != nil {
		return nil, err
	}
	obj, err := domain.GetByID(s.objectDataRepo, objectDataID)
	if err != nil {
		return nil, err
	}
	if obj.ProjectID != nil && *obj.ProjectID != projectID {
		return nil, domain.ErrConflict
	}
	if obj.ProjectID == nil {
		obj.ProjectID = &projectID
	}
	obj.IsActive = true
	if err := s.objectDataRepo.Update(obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *Service) RemoveObjectData(projectID, objectDataID uuid.UUID) (*domainFacility.ObjectData, error) {
	if _, err := domain.GetByID(s.repo, projectID); err != nil {
		return nil, err
	}
	obj, err := domain.GetByID(s.objectDataRepo, objectDataID)
	if err != nil {
		return nil, err
	}
	if obj.ProjectID == nil || *obj.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	obj.IsActive = false
	if err := s.objectDataRepo.Update(obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *Service) CanAccessProject(requesterID, projectID uuid.UUID) (bool, error) {
	project, err := domain.GetByID(s.repo, projectID)
	if err != nil {
		return false, err
	}

	if project.CreatorID == requesterID {
		return true, nil
	}

	users, err := s.userRepo.GetByIds([]uuid.UUID{requesterID})
	if err != nil {
		return false, err
	}

	if len(users) > 0 && domainUser.IsAdmin(users[0].Role) {
		return true, nil
	}

	return s.repo.HasUser(projectID, requesterID)
}

func (s *Service) GetByIds(ids []uuid.UUID) ([]*domainProject.Project, error) {
	return s.repo.GetByIds(ids)
}

func (s *Service) GetByID(id uuid.UUID) (*domainProject.Project, error) {
	return domain.GetByID(s.repo, id)
}

func (s *Service) Update(project *domainProject.Project) error {
	return s.repo.Update(project)
}

func (s *Service) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}

func (s *Service) List(requesterID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainProject.Project], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)

	users, err := s.userRepo.GetByIds([]uuid.UUID{requesterID})
	if err != nil {
		return nil, err
	}

	params := domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	}

	if len(users) > 0 && domainUser.IsAdmin(users[0].Role) {
		return s.repo.GetPaginatedList(params)
	}

	return s.repo.GetPaginatedListForUser(params, requesterID)
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

func (s *Service) ListObjectData(projectID uuid.UUID, page, limit int, search string, apparatID, systemPartID *uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	params := domain.PaginationParams{Page: page, Limit: limit, Search: search}

	switch {
	case apparatID != nil && systemPartID != nil:
		return s.objectDataRepo.GetPaginatedListForProjectByApparatAndSystemPartID(projectID, *apparatID, *systemPartID, params)
	case apparatID != nil:
		return s.objectDataRepo.GetPaginatedListForProjectByApparatID(projectID, *apparatID, params)
	case systemPartID != nil:
		return s.objectDataRepo.GetPaginatedListForProjectBySystemPartID(projectID, *systemPartID, params)
	default:
		return s.objectDataRepo.GetPaginatedListForProject(projectID, params)
	}
}

// MultiCreateFieldDevices creates multiple field devices and links them to a project in one operation.
// For each successfully created field device, it creates a ProjectFieldDevice link.
// Returns the IDs of the created field devices and any association errors.
func (s *Service) MultiCreateFieldDevices(projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) ([]uuid.UUID, []string) {
	if _, err := domain.GetByID(s.repo, projectID); err != nil {
		return nil, []string{"project not found"}
	}

	successIDs := make([]uuid.UUID, 0, len(fieldDeviceIDs))
	errors := make([]string, 0)

	for i, fdID := range fieldDeviceIDs {
		entity := &domainProject.ProjectFieldDevice{
			ProjectID:     projectID,
			FieldDeviceID: fdID,
		}
		if err := s.projectFieldDeviceRepo.Create(entity); err != nil {
			errors = append(errors, err.Error())
		} else {
			successIDs = append(successIDs, fdID)
		}
		// Continue even if one fails
		_ = i // Use index if needed for error reporting
	}

	return successIDs, errors
}

func (s *Service) collectDescendantIDsForControlCabinet(controlCabinetID uuid.UUID) ([]uuid.UUID, []uuid.UUID, []uuid.UUID, error) {
	spsControllerIDs, err := s.spsControllerRepo.GetIDsByControlCabinetID(controlCabinetID)
	if err != nil {
		return nil, nil, nil, err
	}

	systemTypeIDs, fieldDeviceIDs, err := s.collectDescendantIDsForSPSControllers(spsControllerIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	return spsControllerIDs, systemTypeIDs, fieldDeviceIDs, nil
}

func (s *Service) collectDescendantIDsForSPSControllers(spsControllerIDs []uuid.UUID) ([]uuid.UUID, []uuid.UUID, error) {
	if len(spsControllerIDs) == 0 {
		return nil, nil, nil
	}

	systemTypeIDs, err := s.spsControllerSystemRepo.GetIDsBySPSControllerIDs(spsControllerIDs)
	if err != nil {
		return nil, nil, err
	}
	if len(systemTypeIDs) == 0 {
		return nil, nil, nil
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(systemTypeIDs)
	if err != nil {
		return nil, nil, err
	}

	return systemTypeIDs, fieldDeviceIDs, nil
}

func (s *Service) deleteFieldDevicesWithChildren(fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	if err := s.bacnetObjectRepo.DeleteByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.specificationRepo.DeleteByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	return s.fieldDeviceRepo.DeleteByIds(fieldDeviceIDs)
}

func (s *Service) deleteProjectControlCabinetLinksByControlCabinetIDs(controlCabinetIDs []uuid.UUID) error {
	if len(controlCabinetIDs) == 0 {
		return nil
	}

	idSet := toUUIDSet(controlCabinetIDs)
	linkIDs, err := s.collectProjectControlCabinetLinkIDs(idSet)
	if err != nil {
		return err
	}
	if len(linkIDs) == 0 {
		return nil
	}
	return s.projectControlCabinetRepo.DeleteByIds(linkIDs)
}

func (s *Service) deleteProjectSPSControllerLinksBySPSControllerIDs(spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}

	idSet := toUUIDSet(spsControllerIDs)
	linkIDs, err := s.collectProjectSPSControllerLinkIDs(idSet)
	if err != nil {
		return err
	}
	if len(linkIDs) == 0 {
		return nil
	}
	return s.projectSPSControllerRepo.DeleteByIds(linkIDs)
}

func (s *Service) deleteProjectFieldDeviceLinksByFieldDeviceIDs(fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	idSet := toUUIDSet(fieldDeviceIDs)
	linkIDs, err := s.collectProjectFieldDeviceLinkIDs(idSet)
	if err != nil {
		return err
	}
	if len(linkIDs) == 0 {
		return nil
	}
	return s.projectFieldDeviceRepo.DeleteByIds(linkIDs)
}

func (s *Service) collectProjectControlCabinetLinkIDs(controlCabinetIDSet map[uuid.UUID]struct{}) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0)
	page := 1

	for {
		items, err := s.projectControlCabinetRepo.GetPaginatedList(domain.PaginationParams{
			Page:  page,
			Limit: 500,
		})
		if err != nil {
			return nil, err
		}

		for _, item := range items.Items {
			if _, ok := controlCabinetIDSet[item.ControlCabinetID]; ok {
				result = append(result, item.ID)
			}
		}

		if page >= items.TotalPages || len(items.Items) == 0 {
			break
		}
		page++
	}

	return result, nil
}

func (s *Service) collectProjectSPSControllerLinkIDs(spsControllerIDSet map[uuid.UUID]struct{}) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0)
	page := 1

	for {
		items, err := s.projectSPSControllerRepo.GetPaginatedList(domain.PaginationParams{
			Page:  page,
			Limit: 500,
		})
		if err != nil {
			return nil, err
		}

		for _, item := range items.Items {
			if _, ok := spsControllerIDSet[item.SPSControllerID]; ok {
				result = append(result, item.ID)
			}
		}

		if page >= items.TotalPages || len(items.Items) == 0 {
			break
		}
		page++
	}

	return result, nil
}

func (s *Service) collectProjectFieldDeviceLinkIDs(fieldDeviceIDSet map[uuid.UUID]struct{}) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0)
	page := 1

	for {
		items, err := s.projectFieldDeviceRepo.GetPaginatedList(domain.PaginationParams{
			Page:  page,
			Limit: 500,
		})
		if err != nil {
			return nil, err
		}

		for _, item := range items.Items {
			if _, ok := fieldDeviceIDSet[item.FieldDeviceID]; ok {
				result = append(result, item.ID)
			}
		}

		if page >= items.TotalPages || len(items.Items) == 0 {
			break
		}
		page++
	}

	return result, nil
}

func toUUIDSet(ids []uuid.UUID) map[uuid.UUID]struct{} {
	result := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		result[id] = struct{}{}
	}
	return result
}

func (s *Service) linkDescendantsForControlCabinet(projectID, controlCabinetID uuid.UUID) error {
	spsControllerIDs, err := s.spsControllerRepo.GetIDsByControlCabinetID(controlCabinetID)
	if err != nil {
		return err
	}
	return s.linkDescendantsForSPSControllers(projectID, spsControllerIDs)
}

func (s *Service) linkDescendantsForSPSControllers(projectID uuid.UUID, spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}

	existingSPS, err := s.listProjectSPSControllerIDSet(projectID)
	if err != nil {
		return err
	}
	for _, spsID := range spsControllerIDs {
		if _, ok := existingSPS[spsID]; ok {
			continue
		}
		if err := s.createProjectSPSControllerLink(projectID, spsID); err != nil {
			return err
		}
		existingSPS[spsID] = struct{}{}
	}

	systemTypeIDs, err := s.spsControllerSystemRepo.GetIDsBySPSControllerIDs(spsControllerIDs)
	if err != nil {
		return err
	}
	return s.linkFieldDevicesForSystemTypes(projectID, systemTypeIDs)
}

func (s *Service) linkFieldDevicesForSystemTypes(projectID uuid.UUID, systemTypeIDs []uuid.UUID) error {
	if len(systemTypeIDs) == 0 {
		return nil
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(systemTypeIDs)
	if err != nil {
		return err
	}
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	existingFieldDevices, err := s.listProjectFieldDeviceIDSet(projectID)
	if err != nil {
		return err
	}
	for _, fieldDeviceID := range fieldDeviceIDs {
		if _, ok := existingFieldDevices[fieldDeviceID]; ok {
			continue
		}
		if err := s.createProjectFieldDeviceLink(projectID, fieldDeviceID); err != nil {
			return err
		}
		existingFieldDevices[fieldDeviceID] = struct{}{}
	}

	return nil
}

func (s *Service) cleanupProjectLinksForControlCabinetHierarchy(controlCabinetID uuid.UUID) error {
	spsControllerIDs, _, fieldDeviceIDs, err := s.collectDescendantIDsForControlCabinet(controlCabinetID)
	if err != nil {
		return err
	}
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.deleteProjectSPSControllerLinksBySPSControllerIDs(spsControllerIDs); err != nil {
		return err
	}
	return s.deleteProjectControlCabinetLinksByControlCabinetIDs([]uuid.UUID{controlCabinetID})
}

func (s *Service) cleanupProjectLinksForSPSControllers(spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}

	_, fieldDeviceIDs, err := s.collectDescendantIDsForSPSControllers(spsControllerIDs)
	if err != nil {
		return err
	}
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	return s.deleteProjectSPSControllerLinksBySPSControllerIDs(spsControllerIDs)
}

func (s *Service) cleanupProjectLinksForSystemTypes(systemTypeIDs []uuid.UUID) error {
	if len(systemTypeIDs) == 0 {
		return nil
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(systemTypeIDs)
	if err != nil {
		return err
	}
	return s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(fieldDeviceIDs)
}

func (s *Service) rollbackCopiedControlCabinet(controlCabinetID uuid.UUID) error {
	spsControllerIDs, _, fieldDeviceIDs, err := s.collectDescendantIDsForControlCabinet(controlCabinetID)
	if err != nil {
		return err
	}
	if err := s.deleteFieldDevicesWithChildren(fieldDeviceIDs); err != nil {
		return err
	}
	if len(spsControllerIDs) > 0 {
		if err := s.spsControllerSystemRepo.DeleteBySPSControllerIDs(spsControllerIDs); err != nil {
			return err
		}
		if err := s.spsControllerRepo.DeleteByIds(spsControllerIDs); err != nil {
			return err
		}
	}
	return s.controlCabinetRepo.DeleteByIds([]uuid.UUID{controlCabinetID})
}

func (s *Service) rollbackCopiedSPSController(spsControllerID uuid.UUID) error {
	_, fieldDeviceIDs, err := s.collectDescendantIDsForSPSControllers([]uuid.UUID{spsControllerID})
	if err != nil {
		return err
	}
	if err := s.deleteFieldDevicesWithChildren(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.spsControllerSystemRepo.DeleteBySPSControllerIDs([]uuid.UUID{spsControllerID}); err != nil {
		return err
	}
	return s.spsControllerRepo.DeleteByIds([]uuid.UUID{spsControllerID})
}

func (s *Service) rollbackCopiedSPSControllerSystemType(systemTypeID uuid.UUID) error {
	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs([]uuid.UUID{systemTypeID})
	if err != nil {
		return err
	}
	if err := s.deleteFieldDevicesWithChildren(fieldDeviceIDs); err != nil {
		return err
	}
	return s.spsControllerSystemRepo.DeleteByIds([]uuid.UUID{systemTypeID})
}

func (s *Service) listProjectSPSControllerIDSet(projectID uuid.UUID) (map[uuid.UUID]struct{}, error) {
	result := make(map[uuid.UUID]struct{})
	page := 1

	for {
		items, err := s.projectSPSControllerRepo.GetPaginatedListByProjectID(projectID, domain.PaginationParams{
			Page:  page,
			Limit: 500,
		})
		if err != nil {
			return nil, err
		}

		for _, item := range items.Items {
			result[item.SPSControllerID] = struct{}{}
		}

		if page >= items.TotalPages || len(items.Items) == 0 {
			break
		}
		page++
	}

	return result, nil
}

func (s *Service) listProjectFieldDeviceIDSet(projectID uuid.UUID) (map[uuid.UUID]struct{}, error) {
	result := make(map[uuid.UUID]struct{})
	page := 1

	for {
		items, err := s.projectFieldDeviceRepo.GetPaginatedListByProjectID(projectID, domain.PaginationParams{
			Page:  page,
			Limit: 500,
		})
		if err != nil {
			return nil, err
		}

		for _, item := range items.Items {
			result[item.FieldDeviceID] = struct{}{}
		}

		if page >= items.TotalPages || len(items.Items) == 0 {
			break
		}
		page++
	}

	return result, nil
}

func (s *Service) createProjectSPSControllerLink(projectID, spsControllerID uuid.UUID) error {
	entity := &domainProject.ProjectSPSController{
		ProjectID:       projectID,
		SPSControllerID: spsControllerID,
	}
	if err := s.projectSPSControllerRepo.Create(entity); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
	return nil
}

func (s *Service) createProjectFieldDeviceLink(projectID, fieldDeviceID uuid.UUID) error {
	entity := &domainProject.ProjectFieldDevice{
		ProjectID:     projectID,
		FieldDeviceID: fieldDeviceID,
	}
	if err := s.projectFieldDeviceRepo.Create(entity); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
	return nil
}
