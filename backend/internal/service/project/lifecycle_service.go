package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

type ProjectLifecycleService struct {
	repo               domainProject.ProjectRepository
	userRepo           domainUser.UserRepository
	rolePermissionRepo domainUser.RolePermissionRepository
	objectDataRepo     domainObjectData.ObjectDataStore
	bacnetObjectRepo   domainObjectData.BacnetObjectStore
	tx                 txCoordinator
}

var autoAssignedProjectRoles = []domainUser.Role{
	domainUser.RoleSuperAdmin,
	domainUser.RoleAdminFZAG,
}

func (s *ProjectLifecycleService) bindTransactions(tx txCoordinator) {
	s.tx = tx
}

func (s *ProjectLifecycleService) transaction() projectTx[*ProjectLifecycleService] {
	return newProjectTx(s.tx, s, func(services *Services) *ProjectLifecycleService {
		return services.Lifecycle
	})
}

func (s *ProjectLifecycleService) Create(ctx context.Context, project *domainProject.Project) error {
	return s.transaction().run(ctx, func(txCtx context.Context, txService *ProjectLifecycleService) error {
		return txService.createProject(txCtx, project)
	})
}

func (s *ProjectLifecycleService) createProject(ctx context.Context, project *domainProject.Project) error {
	if project.Status == "" {
		project.Status = domainProject.StatusPlanned
	}

	if err := s.repo.Create(ctx, project); err != nil {
		return err
	}

	userIDs, err := s.autoAssignedProjectUserIDs(ctx, project.CreatorID)
	if err != nil {
		return err
	}
	for _, userID := range userIDs {
		if err := s.repo.AddUser(ctx, project.ID, userID); err != nil {
			return err
		}
	}

	return facilityservice.CopyObjectDataTemplatesForProject(ctx, s.objectDataRepo, s.bacnetObjectRepo, project.ID)
}

func (s *ProjectLifecycleService) autoAssignedProjectUserIDs(ctx context.Context, creatorID uuid.UUID) ([]uuid.UUID, error) {
	userIDs := make([]uuid.UUID, 0, 3)
	seen := make(map[uuid.UUID]struct{}, 3)
	add := func(userID uuid.UUID) {
		if userID == uuid.Nil {
			return
		}
		if _, ok := seen[userID]; ok {
			return
		}
		seen[userID] = struct{}{}
		userIDs = append(userIDs, userID)
	}

	add(creatorID)

	roleReader, ok := s.userRepo.(domainUser.UserRoleRepository)
	if !ok || roleReader == nil {
		return userIDs, nil
	}

	users, err := roleReader.ListByRoles(ctx, autoAssignedProjectRoles)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		add(user.ID)
	}

	return userIDs, nil
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

func (s *ProjectLifecycleService) List(ctx context.Context, requesterID uuid.UUID, page, limit int, search string, status *domainProject.ProjectStatus, phaseID *uuid.UUID) (*domain.PaginatedList[domainProject.Project], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)

	params := domain.PaginationParams{Page: page, Limit: limit, Search: search}

	canReadAllProjects, err := s.canReadAllProjects(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	if canReadAllProjects {
		return s.repo.GetPaginatedListWithStatus(ctx, params, status, phaseID)
	}

	return s.repo.GetPaginatedListForUserWithStatus(ctx, params, requesterID, status, phaseID)
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
	if roleCanAccessAllProjects(users[0].Role) {
		return true, nil
	}

	rolePermissions, err := s.rolePermissionRepo.ListByRole(ctx, users[0].Role)
	if err != nil {
		return false, err
	}

	for _, rolePermission := range rolePermissions {
		if rolePermission.Permission == domainUser.PermissionProjectListAll {
			return true, nil
		}
	}

	return false, nil
}
