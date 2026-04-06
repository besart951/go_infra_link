package project

import (
	"context"
	"errors"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TxRunner func(func(tx *gorm.DB) error) error

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
	txRunner                  TxRunner
	txFactory                 func(tx *gorm.DB) (*Service, error)
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
	txRunner TxRunner,
	txFactory func(tx *gorm.DB) (*Service, error),
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
		txRunner:                  txRunner,
		txFactory:                 txFactory,
	}
}

func (s *Service) withTx(fn func(*Service) error) error {
	if s.txRunner == nil || s.txFactory == nil {
		return fn(s)
	}

	return s.txRunner(func(tx *gorm.DB) error {
		txService, err := s.txFactory(tx)
		if err != nil {
			return err
		}
		return fn(txService)
	})
}

func withTxResult[T any](s *Service, fn func(*Service) (T, error)) (T, error) {
	var zero T
	if s.txRunner == nil || s.txFactory == nil {
		return fn(s)
	}

	var result T
	err := s.txRunner(func(tx *gorm.DB) error {
		txService, buildErr := s.txFactory(tx)
		if buildErr != nil {
			return buildErr
		}

		value, runErr := fn(txService)
		if runErr != nil {
			return runErr
		}

		result = value
		return nil
	})
	if err != nil {
		return zero, err
	}

	return result, nil
}

func (s *Service) ListProjectIDsByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]uuid.UUID, error) {
	items, err := s.projectControlCabinetRepo.GetByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return nil, err
	}

	projectIDSet := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		projectIDSet[item.ProjectID] = struct{}{}
	}

	projectIDs := make([]uuid.UUID, 0, len(projectIDSet))
	for projectID := range projectIDSet {
		projectIDs = append(projectIDs, projectID)
	}
	return projectIDs, nil
}

func (s *Service) ListProjectIDsBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID) ([]uuid.UUID, error) {
	items, err := s.projectSPSControllerRepo.GetBySPSControllerID(ctx, spsControllerID)
	if err != nil {
		return nil, err
	}

	projectIDSet := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		projectIDSet[item.ProjectID] = struct{}{}
	}

	projectIDs := make([]uuid.UUID, 0, len(projectIDSet))
	for projectID := range projectIDSet {
		projectIDs = append(projectIDs, projectID)
	}
	return projectIDs, nil
}

func (s *Service) Create(ctx context.Context, project *domainProject.Project) error {
	return s.withTx(func(txService *Service) error {
		return txService.createProject(ctx, project)
	})
}

func (s *Service) createProject(ctx context.Context, project *domainProject.Project) error {
	if project.Status == "" {
		project.Status = domainProject.StatusPlanned
	}

	if err := s.repo.Create(ctx, project); err != nil {
		return err
	}

	if project.CreatorID != uuid.Nil {
		if err := s.repo.AddUser(ctx, project.ID, project.CreatorID); err != nil {
			return err
		}
	}

	// Copy ObjectData templates
	templates, err := s.objectDataRepo.GetTemplates(ctx)
	if err != nil {
		return err
	}

	for _, tmpl := range templates {
		copy := *tmpl
		copy.ID = uuid.Nil
		copy.ProjectID = &project.ID
		copy.BacnetObjects = nil // clear for now, we will rebuild them

		if err := s.objectDataRepo.Create(ctx, &copy); err != nil {
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
			if err := s.bacnetObjectRepo.Create(ctx, newBO); err != nil {
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
					if err := s.bacnetObjectRepo.Update(ctx, newBO); err != nil {
						return err
					}
				}
			}
			newBacnetObjects = append(newBacnetObjects, newBO)
		}

		// 3. Associate with ObjectData
		copy.BacnetObjects = newBacnetObjects
		if err := s.objectDataRepo.Update(ctx, &copy); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) CreateControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	return withTxResult(s, func(txService *Service) (*domainProject.ProjectControlCabinet, error) {
		return txService.createControlCabinet(ctx, projectID, controlCabinetID)
	})
}

func (s *Service) createControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	entity := &domainProject.ProjectControlCabinet{
		ProjectID:        projectID,
		ControlCabinetID: controlCabinetID,
	}
	if err := s.projectControlCabinetRepo.Create(ctx, entity); err != nil {
		return nil, err
	}

	if err := s.linkDescendantsForControlCabinet(ctx, projectID, controlCabinetID); err != nil {
		_ = s.cleanupProjectLinksForControlCabinetHierarchy(ctx, controlCabinetID)
		return nil, err
	}

	return entity, nil
}

func (s *Service) CopyControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) (*domainFacility.ControlCabinet, error) {
	return withTxResult(s, func(txService *Service) (*domainFacility.ControlCabinet, error) {
		return txService.copyControlCabinet(ctx, projectID, controlCabinetID)
	})
}

func (s *Service) copyControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) (*domainFacility.ControlCabinet, error) {
	copyEntity, err := s.hierarchyCopier.CopyControlCabinetByID(ctx, controlCabinetID)
	if err != nil {
		return nil, err
	}

	if _, err := s.createControlCabinet(ctx, projectID, copyEntity.ID); err != nil {
		_ = s.rollbackCopiedControlCabinet(ctx, copyEntity.ID)
		return nil, err
	}

	return copyEntity, nil
}
func (s *Service) UpdateControlCabinet(ctx context.Context, linkID, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	return withTxResult(s, func(txService *Service) (*domainProject.ProjectControlCabinet, error) {
		return txService.updateControlCabinet(ctx, linkID, projectID, controlCabinetID)
	})
}

func (s *Service) updateControlCabinet(ctx context.Context, linkID, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	entity, err := domain.GetByID(ctx, s.projectControlCabinetRepo, linkID)
	if err != nil {
		return nil, err
	}
	if entity.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	previousControlCabinetID := entity.ControlCabinetID
	entity.ControlCabinetID = controlCabinetID
	if err := s.projectControlCabinetRepo.Update(ctx, entity); err != nil {
		return nil, err
	}

	if err := s.linkDescendantsForControlCabinet(ctx, projectID, controlCabinetID); err != nil {
		_ = s.cleanupProjectLinksForControlCabinetHierarchy(ctx, controlCabinetID)
		entity.ControlCabinetID = previousControlCabinetID
		_ = s.projectControlCabinetRepo.Update(ctx, entity)
		return nil, err
	}

	return entity, nil
}

func (s *Service) DeleteControlCabinet(ctx context.Context, linkID, projectID uuid.UUID) error {
	return s.withTx(func(txService *Service) error {
		return txService.deleteControlCabinet(ctx, linkID, projectID)
	})
}

func (s *Service) deleteControlCabinet(ctx context.Context, linkID, projectID uuid.UUID) error {
	entity, err := domain.GetByID(ctx, s.projectControlCabinetRepo, linkID)
	if err != nil {
		return err
	}
	if entity.ProjectID != projectID {
		return domain.ErrNotFound
	}

	controlCabinetID := entity.ControlCabinetID
	spsControllerIDs, _, fieldDeviceIDs, err := s.collectDescendantIDsForControlCabinet(ctx, controlCabinetID)
	if err != nil {
		return err
	}

	if err := s.deleteProjectControlCabinetLinksByControlCabinetIDs(ctx, []uuid.UUID{controlCabinetID}); err != nil {
		return err
	}
	if err := s.deleteProjectSPSControllerLinksBySPSControllerIDs(ctx, spsControllerIDs); err != nil {
		return err
	}
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(ctx, fieldDeviceIDs); err != nil {
		return err
	}

	if err := s.deleteFieldDevicesWithChildren(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	if len(spsControllerIDs) > 0 {
		if err := s.spsControllerSystemRepo.DeleteBySPSControllerIDs(ctx, spsControllerIDs); err != nil {
			return err
		}
		if err := s.spsControllerRepo.DeleteByIds(ctx, spsControllerIDs); err != nil {
			return err
		}
	}

	return s.controlCabinetRepo.DeleteByIds(ctx, []uuid.UUID{controlCabinetID})
}

func (s *Service) CreateSPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	return withTxResult(s, func(txService *Service) (*domainProject.ProjectSPSController, error) {
		return txService.createSPSController(ctx, projectID, spsControllerID)
	})
}

func (s *Service) createSPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	entity := &domainProject.ProjectSPSController{
		ProjectID:       projectID,
		SPSControllerID: spsControllerID,
	}
	if err := s.projectSPSControllerRepo.Create(ctx, entity); err != nil {
		return nil, err
	}

	if err := s.linkDescendantsForSPSControllers(ctx, projectID, []uuid.UUID{spsControllerID}); err != nil {
		_ = s.cleanupProjectLinksForSPSControllers(ctx, []uuid.UUID{spsControllerID})
		return nil, err
	}

	return entity, nil
}

func (s *Service) CopySPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) (*domainFacility.SPSController, error) {
	return withTxResult(s, func(txService *Service) (*domainFacility.SPSController, error) {
		return txService.copySPSController(ctx, projectID, spsControllerID)
	})
}

func (s *Service) copySPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) (*domainFacility.SPSController, error) {
	copyEntity, err := s.hierarchyCopier.CopySPSControllerByID(ctx, spsControllerID)
	if err != nil {
		return nil, err
	}

	if _, err := s.createSPSController(ctx, projectID, copyEntity.ID); err != nil {
		_ = s.rollbackCopiedSPSController(ctx, copyEntity.ID)
		return nil, err
	}

	return copyEntity, nil
}

func (s *Service) CopySPSControllerSystemType(ctx context.Context, projectID, systemTypeID uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	return withTxResult(s, func(txService *Service) (*domainFacility.SPSControllerSystemType, error) {
		return txService.copySPSControllerSystemType(ctx, projectID, systemTypeID)
	})
}

func (s *Service) copySPSControllerSystemType(ctx context.Context, projectID, systemTypeID uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	copyEntity, err := s.hierarchyCopier.CopySPSControllerSystemTypeByID(ctx, systemTypeID)
	if err != nil {
		return nil, err
	}

	if err := s.linkFieldDevicesForSystemTypes(ctx, projectID, []uuid.UUID{copyEntity.ID}); err != nil {
		_ = s.cleanupProjectLinksForSystemTypes(ctx, []uuid.UUID{copyEntity.ID})
		_ = s.rollbackCopiedSPSControllerSystemType(ctx, copyEntity.ID)
		return nil, err
	}

	return copyEntity, nil
}

func (s *Service) UpdateSPSController(ctx context.Context, linkID, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	return withTxResult(s, func(txService *Service) (*domainProject.ProjectSPSController, error) {
		return txService.updateSPSController(ctx, linkID, projectID, spsControllerID)
	})
}

func (s *Service) updateSPSController(ctx context.Context, linkID, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	entity, err := domain.GetByID(ctx, s.projectSPSControllerRepo, linkID)
	if err != nil {
		return nil, err
	}
	if entity.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	previousSPSControllerID := entity.SPSControllerID
	entity.SPSControllerID = spsControllerID
	if err := s.projectSPSControllerRepo.Update(ctx, entity); err != nil {
		return nil, err
	}

	if err := s.linkDescendantsForSPSControllers(ctx, projectID, []uuid.UUID{spsControllerID}); err != nil {
		_ = s.cleanupProjectLinksForSPSControllers(ctx, []uuid.UUID{spsControllerID})
		entity.SPSControllerID = previousSPSControllerID
		_ = s.projectSPSControllerRepo.Update(ctx, entity)
		return nil, err
	}

	return entity, nil
}

func (s *Service) DeleteSPSController(ctx context.Context, linkID, projectID uuid.UUID) error {
	return s.withTx(func(txService *Service) error {
		return txService.deleteSPSController(ctx, linkID, projectID)
	})
}

func (s *Service) deleteSPSController(ctx context.Context, linkID, projectID uuid.UUID) error {
	entity, err := domain.GetByID(ctx, s.projectSPSControllerRepo, linkID)
	if err != nil {
		return err
	}
	if entity.ProjectID != projectID {
		return domain.ErrNotFound
	}

	spsControllerID := entity.SPSControllerID
	_, fieldDeviceIDs, err := s.collectDescendantIDsForSPSControllers(ctx, []uuid.UUID{spsControllerID})
	if err != nil {
		return err
	}

	if err := s.deleteProjectSPSControllerLinksBySPSControllerIDs(ctx, []uuid.UUID{spsControllerID}); err != nil {
		return err
	}
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(ctx, fieldDeviceIDs); err != nil {
		return err
	}

	if err := s.deleteFieldDevicesWithChildren(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.spsControllerSystemRepo.DeleteBySPSControllerIDs(ctx, []uuid.UUID{spsControllerID}); err != nil {
		return err
	}
	return s.spsControllerRepo.DeleteByIds(ctx, []uuid.UUID{spsControllerID})
}

func (s *Service) CreateFieldDevice(ctx context.Context, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error) {
	return withTxResult(s, func(txService *Service) (*domainProject.ProjectFieldDevice, error) {
		return txService.createFieldDevice(ctx, projectID, fieldDeviceID)
	})
}

func (s *Service) createFieldDevice(ctx context.Context, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error) {
	entity := &domainProject.ProjectFieldDevice{
		ProjectID:     projectID,
		FieldDeviceID: fieldDeviceID,
	}
	if err := s.projectFieldDeviceRepo.Create(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) InviteUser(ctx context.Context, projectID, userID uuid.UUID) error {
	if _, err := domain.GetByID(ctx, s.repo, projectID); err != nil {
		return err
	}
	if _, err := domain.GetByID(ctx, s.userRepo, userID); err != nil {
		return err
	}
	return s.repo.AddUser(ctx, projectID, userID)
}

func (s *Service) ListUsers(ctx context.Context, projectID uuid.UUID) ([]domainUser.User, error) {
	if _, err := domain.GetByID(ctx, s.repo, projectID); err != nil {
		return nil, err
	}
	return s.repo.ListUsers(ctx, projectID)
}

func (s *Service) RemoveUser(ctx context.Context, projectID, userID uuid.UUID) error {
	if _, err := domain.GetByID(ctx, s.repo, projectID); err != nil {
		return err
	}
	if _, err := domain.GetByID(ctx, s.userRepo, userID); err != nil {
		return err
	}
	return s.repo.RemoveUser(ctx, projectID, userID)
}

func (s *Service) UpdateFieldDevice(ctx context.Context, linkID, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error) {
	entity, err := domain.GetByID(ctx, s.projectFieldDeviceRepo, linkID)
	if err != nil {
		return nil, err
	}
	if entity.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	entity.FieldDeviceID = fieldDeviceID
	if err := s.projectFieldDeviceRepo.Update(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) DeleteFieldDevice(ctx context.Context, linkID, projectID uuid.UUID) error {
	return s.withTx(func(txService *Service) error {
		return txService.deleteFieldDevice(ctx, linkID, projectID)
	})
}

func (s *Service) deleteFieldDevice(ctx context.Context, linkID, projectID uuid.UUID) error {
	entity, err := domain.GetByID(ctx, s.projectFieldDeviceRepo, linkID)
	if err != nil {
		return err
	}
	if entity.ProjectID != projectID {
		return domain.ErrNotFound
	}

	fieldDeviceID := entity.FieldDeviceID
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID}); err != nil {
		return err
	}
	return s.deleteFieldDevicesWithChildren(ctx, []uuid.UUID{fieldDeviceID})
}

func (s *Service) AddObjectData(ctx context.Context, projectID, objectDataID uuid.UUID) (*domainFacility.ObjectData, error) {
	if _, err := domain.GetByID(ctx, s.repo, projectID); err != nil {
		return nil, err
	}
	obj, err := domain.GetByID(ctx, s.objectDataRepo, objectDataID)
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
	if err := s.objectDataRepo.Update(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *Service) RemoveObjectData(ctx context.Context, projectID, objectDataID uuid.UUID) (*domainFacility.ObjectData, error) {
	if _, err := domain.GetByID(ctx, s.repo, projectID); err != nil {
		return nil, err
	}
	obj, err := domain.GetByID(ctx, s.objectDataRepo, objectDataID)
	if err != nil {
		return nil, err
	}
	if obj.ProjectID == nil || *obj.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	obj.IsActive = false
	if err := s.objectDataRepo.Update(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *Service) CanAccessProject(ctx context.Context, requesterID, projectID uuid.UUID) (bool, error) {
	project, err := domain.GetByID(ctx, s.repo, projectID)
	if err != nil {
		return false, err
	}

	if project.CreatorID == requesterID {
		return true, nil
	}

	users, err := s.userRepo.GetByIds(ctx, []uuid.UUID{requesterID})
	if err != nil {
		return false, err
	}

	if len(users) > 0 && domainUser.IsAdmin(users[0].Role) {
		return true, nil
	}

	return s.repo.HasUser(ctx, projectID, requesterID)
}

func (s *Service) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainProject.Project, error) {
	return s.repo.GetByIds(ctx, ids)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domainProject.Project, error) {
	return domain.GetByID(ctx, s.repo, id)
}

func (s *Service) Update(ctx context.Context, project *domainProject.Project) error {
	return s.repo.Update(ctx, project)
}

func (s *Service) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteByIds(ctx, []uuid.UUID{id})
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, page, limit int, search string, status *domainProject.ProjectStatus) (*domain.PaginatedList[domainProject.Project], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)

	users, err := s.userRepo.GetByIds(ctx, []uuid.UUID{requesterID})
	if err != nil {
		return nil, err
	}

	params := domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	}

	if len(users) > 0 && domainUser.IsAdmin(users[0].Role) {
		return s.repo.GetPaginatedListWithStatus(ctx, params, status)
	}

	return s.repo.GetPaginatedListForUserWithStatus(ctx, params, requesterID, status)
}

func (s *Service) ListControlCabinets(ctx context.Context, projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectControlCabinet], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.projectControlCabinetRepo.GetPaginatedListByProjectID(ctx, projectID, domain.PaginationParams{Page: page, Limit: limit})
}

func (s *Service) ListSPSControllers(ctx context.Context, projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectSPSController], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.projectSPSControllerRepo.GetPaginatedListByProjectID(ctx, projectID, domain.PaginationParams{Page: page, Limit: limit})
}

func (s *Service) ListFieldDevices(ctx context.Context, projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectFieldDevice], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.projectFieldDeviceRepo.GetPaginatedListByProjectID(ctx, projectID, domain.PaginationParams{Page: page, Limit: limit})
}

func (s *Service) ListObjectData(ctx context.Context, projectID uuid.UUID, page, limit int, search string, apparatID, systemPartID *uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	params := domain.PaginationParams{Page: page, Limit: limit, Search: search}

	switch {
	case apparatID != nil && systemPartID != nil:
		return s.objectDataRepo.GetPaginatedListForProjectByApparatAndSystemPartID(ctx, projectID, *apparatID, *systemPartID, params)
	case apparatID != nil:
		return s.objectDataRepo.GetPaginatedListForProjectByApparatID(ctx, projectID, *apparatID, params)
	case systemPartID != nil:
		return s.objectDataRepo.GetPaginatedListForProjectBySystemPartID(ctx, projectID, *systemPartID, params)
	default:
		return s.objectDataRepo.GetPaginatedListForProject(ctx, projectID, params)
	}
}

// MultiCreateFieldDevices creates multiple field devices and links them to a project in one operation.
// For each successfully created field device, it creates a ProjectFieldDevice link.
// Returns the IDs of the created field devices and any association errors.
func (s *Service) MultiCreateFieldDevices(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) ([]uuid.UUID, []string) {
	return s.multiCreateFieldDevices(ctx, projectID, fieldDeviceIDs)
}

func (s *Service) multiCreateFieldDevices(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) ([]uuid.UUID, []string) {
	if _, err := domain.GetByID(ctx, s.repo, projectID); err != nil {
		return nil, []string{"project not found"}
	}

	successIDs := make([]uuid.UUID, 0, len(fieldDeviceIDs))
	errors := make([]string, 0)

	for i, fdID := range fieldDeviceIDs {
		entity := &domainProject.ProjectFieldDevice{
			ProjectID:     projectID,
			FieldDeviceID: fdID,
		}
		if err := s.projectFieldDeviceRepo.Create(ctx, entity); err != nil {
			errors = append(errors, err.Error())
		} else {
			successIDs = append(successIDs, fdID)
		}
		// Continue even if one fails
		_ = i // Use index if needed for error reporting
	}

	return successIDs, errors
}

func (s *Service) collectDescendantIDsForControlCabinet(ctx context.Context, controlCabinetID uuid.UUID) ([]uuid.UUID, []uuid.UUID, []uuid.UUID, error) {
	spsControllerIDs, err := s.spsControllerRepo.GetIDsByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return nil, nil, nil, err
	}

	systemTypeIDs, fieldDeviceIDs, err := s.collectDescendantIDsForSPSControllers(ctx, spsControllerIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	return spsControllerIDs, systemTypeIDs, fieldDeviceIDs, nil
}

func (s *Service) collectDescendantIDsForSPSControllers(ctx context.Context, spsControllerIDs []uuid.UUID) ([]uuid.UUID, []uuid.UUID, error) {
	if len(spsControllerIDs) == 0 {
		return nil, nil, nil
	}

	systemTypeIDs, err := s.spsControllerSystemRepo.GetIDsBySPSControllerIDs(ctx, spsControllerIDs)
	if err != nil {
		return nil, nil, err
	}
	if len(systemTypeIDs) == 0 {
		return nil, nil, nil
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, systemTypeIDs)
	if err != nil {
		return nil, nil, err
	}

	return systemTypeIDs, fieldDeviceIDs, nil
}

func (s *Service) deleteFieldDevicesWithChildren(ctx context.Context, fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	if err := s.bacnetObjectRepo.DeleteByFieldDeviceIDs(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.specificationRepo.DeleteByFieldDeviceIDs(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	return s.fieldDeviceRepo.DeleteByIds(ctx, fieldDeviceIDs)
}

func (s *Service) deleteProjectControlCabinetLinksByControlCabinetIDs(ctx context.Context, controlCabinetIDs []uuid.UUID) error {
	if len(controlCabinetIDs) == 0 {
		return nil
	}

	idSet := toUUIDSet(controlCabinetIDs)
	linkIDs, err := s.collectProjectControlCabinetLinkIDs(ctx, idSet)
	if err != nil {
		return err
	}
	if len(linkIDs) == 0 {
		return nil
	}
	return s.projectControlCabinetRepo.DeleteByIds(ctx, linkIDs)
}

func (s *Service) deleteProjectSPSControllerLinksBySPSControllerIDs(ctx context.Context, spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}

	idSet := toUUIDSet(spsControllerIDs)
	linkIDs, err := s.collectProjectSPSControllerLinkIDs(ctx, idSet)
	if err != nil {
		return err
	}
	if len(linkIDs) == 0 {
		return nil
	}
	return s.projectSPSControllerRepo.DeleteByIds(ctx, linkIDs)
}

func (s *Service) deleteProjectFieldDeviceLinksByFieldDeviceIDs(ctx context.Context, fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	idSet := toUUIDSet(fieldDeviceIDs)
	linkIDs, err := s.collectProjectFieldDeviceLinkIDs(ctx, idSet)
	if err != nil {
		return err
	}
	if len(linkIDs) == 0 {
		return nil
	}
	return s.projectFieldDeviceRepo.DeleteByIds(ctx, linkIDs)
}

func (s *Service) collectProjectControlCabinetLinkIDs(ctx context.Context, controlCabinetIDSet map[uuid.UUID]struct{}) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0)
	page := 1

	for {
		items, err := s.projectControlCabinetRepo.GetPaginatedList(ctx, domain.PaginationParams{
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

func (s *Service) collectProjectSPSControllerLinkIDs(ctx context.Context, spsControllerIDSet map[uuid.UUID]struct{}) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0)
	page := 1

	for {
		items, err := s.projectSPSControllerRepo.GetPaginatedList(ctx, domain.PaginationParams{
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

func (s *Service) collectProjectFieldDeviceLinkIDs(ctx context.Context, fieldDeviceIDSet map[uuid.UUID]struct{}) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0)
	page := 1

	for {
		items, err := s.projectFieldDeviceRepo.GetPaginatedList(ctx, domain.PaginationParams{
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

func (s *Service) linkDescendantsForControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) error {
	spsControllerIDs, err := s.spsControllerRepo.GetIDsByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return err
	}
	return s.linkDescendantsForSPSControllers(ctx, projectID, spsControllerIDs)
}

func (s *Service) linkDescendantsForSPSControllers(ctx context.Context, projectID uuid.UUID, spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}

	existingSPS, err := s.listProjectSPSControllerIDSet(ctx, projectID)
	if err != nil {
		return err
	}
	for _, spsID := range spsControllerIDs {
		if _, ok := existingSPS[spsID]; ok {
			continue
		}
		if err := s.createProjectSPSControllerLink(ctx, projectID, spsID); err != nil {
			return err
		}
		existingSPS[spsID] = struct{}{}
	}

	systemTypeIDs, err := s.spsControllerSystemRepo.GetIDsBySPSControllerIDs(ctx, spsControllerIDs)
	if err != nil {
		return err
	}
	return s.linkFieldDevicesForSystemTypes(ctx, projectID, systemTypeIDs)
}

func (s *Service) linkFieldDevicesForSystemTypes(ctx context.Context, projectID uuid.UUID, systemTypeIDs []uuid.UUID) error {
	if len(systemTypeIDs) == 0 {
		return nil
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, systemTypeIDs)
	if err != nil {
		return err
	}
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	existingFieldDevices, err := s.listProjectFieldDeviceIDSet(ctx, projectID)
	if err != nil {
		return err
	}
	for _, fieldDeviceID := range fieldDeviceIDs {
		if _, ok := existingFieldDevices[fieldDeviceID]; ok {
			continue
		}
		if err := s.createProjectFieldDeviceLink(ctx, projectID, fieldDeviceID); err != nil {
			return err
		}
		existingFieldDevices[fieldDeviceID] = struct{}{}
	}

	return nil
}

func (s *Service) cleanupProjectLinksForControlCabinetHierarchy(ctx context.Context, controlCabinetID uuid.UUID) error {
	spsControllerIDs, _, fieldDeviceIDs, err := s.collectDescendantIDsForControlCabinet(ctx, controlCabinetID)
	if err != nil {
		return err
	}
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.deleteProjectSPSControllerLinksBySPSControllerIDs(ctx, spsControllerIDs); err != nil {
		return err
	}
	return s.deleteProjectControlCabinetLinksByControlCabinetIDs(ctx, []uuid.UUID{controlCabinetID})
}

func (s *Service) cleanupProjectLinksForSPSControllers(ctx context.Context, spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}

	_, fieldDeviceIDs, err := s.collectDescendantIDsForSPSControllers(ctx, spsControllerIDs)
	if err != nil {
		return err
	}
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	return s.deleteProjectSPSControllerLinksBySPSControllerIDs(ctx, spsControllerIDs)
}

func (s *Service) cleanupProjectLinksForSystemTypes(ctx context.Context, systemTypeIDs []uuid.UUID) error {
	if len(systemTypeIDs) == 0 {
		return nil
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, systemTypeIDs)
	if err != nil {
		return err
	}
	return s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(ctx, fieldDeviceIDs)
}

func (s *Service) rollbackCopiedControlCabinet(ctx context.Context, controlCabinetID uuid.UUID) error {
	spsControllerIDs, _, fieldDeviceIDs, err := s.collectDescendantIDsForControlCabinet(ctx, controlCabinetID)
	if err != nil {
		return err
	}
	if err := s.deleteFieldDevicesWithChildren(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	if len(spsControllerIDs) > 0 {
		if err := s.spsControllerSystemRepo.DeleteBySPSControllerIDs(ctx, spsControllerIDs); err != nil {
			return err
		}
		if err := s.spsControllerRepo.DeleteByIds(ctx, spsControllerIDs); err != nil {
			return err
		}
	}
	return s.controlCabinetRepo.DeleteByIds(ctx, []uuid.UUID{controlCabinetID})
}

func (s *Service) rollbackCopiedSPSController(ctx context.Context, spsControllerID uuid.UUID) error {
	_, fieldDeviceIDs, err := s.collectDescendantIDsForSPSControllers(ctx, []uuid.UUID{spsControllerID})
	if err != nil {
		return err
	}
	if err := s.deleteFieldDevicesWithChildren(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.spsControllerSystemRepo.DeleteBySPSControllerIDs(ctx, []uuid.UUID{spsControllerID}); err != nil {
		return err
	}
	return s.spsControllerRepo.DeleteByIds(ctx, []uuid.UUID{spsControllerID})
}

func (s *Service) rollbackCopiedSPSControllerSystemType(ctx context.Context, systemTypeID uuid.UUID) error {
	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, []uuid.UUID{systemTypeID})
	if err != nil {
		return err
	}
	if err := s.deleteFieldDevicesWithChildren(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	return s.spsControllerSystemRepo.DeleteByIds(ctx, []uuid.UUID{systemTypeID})
}

func (s *Service) listProjectSPSControllerIDSet(ctx context.Context, projectID uuid.UUID) (map[uuid.UUID]struct{}, error) {
	result := make(map[uuid.UUID]struct{})
	page := 1

	for {
		items, err := s.projectSPSControllerRepo.GetPaginatedListByProjectID(ctx, projectID, domain.PaginationParams{
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

func (s *Service) listProjectFieldDeviceIDSet(ctx context.Context, projectID uuid.UUID) (map[uuid.UUID]struct{}, error) {
	result := make(map[uuid.UUID]struct{})
	page := 1

	for {
		items, err := s.projectFieldDeviceRepo.GetPaginatedListByProjectID(ctx, projectID, domain.PaginationParams{
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

func (s *Service) createProjectSPSControllerLink(ctx context.Context, projectID, spsControllerID uuid.UUID) error {
	entity := &domainProject.ProjectSPSController{
		ProjectID:       projectID,
		SPSControllerID: spsControllerID,
	}
	if err := s.projectSPSControllerRepo.Create(ctx, entity); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
	return nil
}

func (s *Service) createProjectFieldDeviceLink(ctx context.Context, projectID, fieldDeviceID uuid.UUID) error {
	entity := &domainProject.ProjectFieldDevice{
		ProjectID:     projectID,
		FieldDeviceID: fieldDeviceID,
	}
	if err := s.projectFieldDeviceRepo.Create(ctx, entity); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
	return nil
}
