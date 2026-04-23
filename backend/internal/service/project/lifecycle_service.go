package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectLifecycleService struct {
	repo               domainProject.ProjectRepository
	userRepo           domainUser.UserRepository
	rolePermissionRepo domainUser.RolePermissionRepository
	objectDataRepo     domainFacility.ObjectDataStore
	bacnetObjectRepo   domainFacility.BacnetObjectStore
	txRunner           TxRunner
	txFactory          func(tx *gorm.DB) (*Services, error)
}

func (s *ProjectLifecycleService) withTx(fn func(*ProjectLifecycleService) error) error {
	if s.txRunner == nil || s.txFactory == nil {
		return fn(s)
	}

	return s.txRunner(func(tx *gorm.DB) error {
		txServices, err := s.txFactory(tx)
		if err != nil {
			return err
		}
		return fn(txServices.Lifecycle)
	})
}

func (s *ProjectLifecycleService) Create(ctx context.Context, project *domainProject.Project) error {
	return s.withTx(func(txService *ProjectLifecycleService) error {
		return txService.createProject(ctx, project)
	})
}

func (s *ProjectLifecycleService) createProject(ctx context.Context, project *domainProject.Project) error {
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

	templates, err := s.objectDataRepo.GetTemplates(ctx)
	if err != nil {
		return err
	}

	for _, tmpl := range templates {
		copy := *tmpl
		copy.ID = uuid.Nil
		copy.ProjectID = &project.ID
		copy.BacnetObjects = nil

		if err := s.objectDataRepo.Create(ctx, &copy); err != nil {
			return err
		}

		if len(tmpl.BacnetObjects) == 0 {
			continue
		}

		oldToNew := make(map[uuid.UUID]*domainFacility.BacnetObject)
		oldRefs := make(map[uuid.UUID]*uuid.UUID)

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
			}
			if err := s.bacnetObjectRepo.Create(ctx, newBO); err != nil {
				return err
			}
			oldToNew[bo.ID] = newBO
			oldRefs[bo.ID] = bo.SoftwareReferenceID
		}

		newBacnetObjects := make([]*domainFacility.BacnetObject, 0, len(tmpl.BacnetObjects))
		for oldID, newBO := range oldToNew {
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

		copy.BacnetObjects = newBacnetObjects
		if err := s.objectDataRepo.Update(ctx, &copy); err != nil {
			return err
		}
	}

	return nil
}

func (s *ProjectLifecycleService) GetByID(ctx context.Context, id uuid.UUID) (*domainProject.Project, error) {
	return domain.GetByID(ctx, s.repo, id)
}

func (s *ProjectLifecycleService) Update(ctx context.Context, project *domainProject.Project) error {
	return s.repo.Update(ctx, project)
}

func (s *ProjectLifecycleService) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteByIds(ctx, []uuid.UUID{id})
}

func (s *ProjectLifecycleService) List(ctx context.Context, requesterID uuid.UUID, page, limit int, search string, status *domainProject.ProjectStatus) (*domain.PaginatedList[domainProject.Project], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)

	params := domain.PaginationParams{Page: page, Limit: limit, Search: search}

	canReadAllProjects, err := s.canReadAllProjects(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	if canReadAllProjects {
		return s.repo.GetPaginatedListWithStatus(ctx, params, status)
	}

	return s.repo.GetPaginatedListForUserWithStatus(ctx, params, requesterID, status)
}

func (s *ProjectLifecycleService) canReadAllProjects(ctx context.Context, requesterID uuid.UUID) (bool, error) {
	if s.rolePermissionRepo == nil || s.userRepo == nil {
		return false, nil
	}

	users, err := s.userRepo.GetByIds(ctx, []uuid.UUID{requesterID})
	if err != nil {
		return false, err
	}
	if len(users) == 0 {
		return false, nil
	}

	rolePermissions, err := s.rolePermissionRepo.ListByRole(ctx, users[0].Role)
	if err != nil {
		return false, err
	}

	for _, rolePermission := range rolePermissions {
		if rolePermission.Permission == "project.read" {
			return true, nil
		}
	}

	return false, nil
}
