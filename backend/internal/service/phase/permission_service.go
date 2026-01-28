package phase

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type PermissionService struct {
	repo project.PhasePermissionRepository
}

func NewPhasePermissionService(repo project.PhasePermissionRepository) *PermissionService {
	return &PermissionService{repo: repo}
}

func (s *PermissionService) Create(perm *project.PhasePermission) error {
	if err := perm.InitForCreate(time.Now().UTC()); err != nil {
		return err
	}
	return s.repo.Create(perm)
}

func (s *PermissionService) GetByID(id uuid.UUID) (*project.PhasePermission, error) {
	perms, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(perms) == 0 {
		return nil, domain.ErrNotFound
	}
	return perms[0], nil
}

func (s *PermissionService) GetByPhaseAndRole(phaseID uuid.UUID, role user.Role) (*project.PhasePermission, error) {
	return s.repo.GetByPhaseAndRole(phaseID, role)
}

func (s *PermissionService) ListByPhase(phaseID uuid.UUID) ([]project.PhasePermission, error) {
	return s.repo.ListByPhase(phaseID)
}

func (s *PermissionService) Update(perm *project.PhasePermission) error {
	return s.repo.Update(perm)
}

func (s *PermissionService) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}

func (s *PermissionService) DeleteByPhaseAndRole(phaseID uuid.UUID, role user.Role) error {
	return s.repo.DeleteByPhaseAndRole(phaseID, role)
}
