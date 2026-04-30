package phasepermission

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	rules       domainProject.PhasePermissionRepository
	phases      domainProject.PhaseRepository
	permissions domainUser.PermissionRepository
}

func New(rules domainProject.PhasePermissionRepository, phases domainProject.PhaseRepository, permissions domainUser.PermissionRepository) *Service {
	return &Service{
		rules:       rules,
		phases:      phases,
		permissions: permissions,
	}
}

func (s *Service) Create(ctx context.Context, rule *domainProject.PhasePermission) error {
	if rule == nil {
		return errors.New("phase_permission_required")
	}
	if err := s.prepareRule(ctx, rule); err != nil {
		return err
	}

	if _, err := s.rules.GetByPhaseAndRole(ctx, rule.PhaseID, rule.Role); err == nil {
		return domain.ErrConflict
	} else if !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	return s.rules.Create(ctx, rule)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domainProject.PhasePermission, error) {
	return domain.GetByID(ctx, s.rules, id)
}

func (s *Service) List(ctx context.Context, phaseID *uuid.UUID) ([]domainProject.PhasePermission, error) {
	return s.rules.List(ctx, phaseID)
}

func (s *Service) Update(ctx context.Context, rule *domainProject.PhasePermission) error {
	if rule == nil {
		return errors.New("phase_permission_required")
	}
	if err := s.prepareRule(ctx, rule); err != nil {
		return err
	}

	existing, err := s.rules.GetByPhaseAndRole(ctx, rule.PhaseID, rule.Role)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	if existing != nil && existing.ID != rule.ID {
		return domain.ErrConflict
	}

	return s.rules.Update(ctx, rule)
}

func (s *Service) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return s.rules.DeleteByIds(ctx, []uuid.UUID{id})
}

func (s *Service) prepareRule(ctx context.Context, rule *domainProject.PhasePermission) error {
	if rule.PhaseID == uuid.Nil || !domainUser.IsValidRole(rule.Role) || rule.Role == domainUser.RoleSuperAdmin {
		return domain.ErrInvalidArgument
	}
	if s.phases != nil {
		if _, err := domain.GetByID(ctx, s.phases, rule.PhaseID); err != nil {
			return err
		}
	}

	permissions := normalizePermissions(rule.Permissions)
	if err := s.validatePermissions(ctx, permissions); err != nil {
		return err
	}
	rule.Permissions = permissions
	return nil
}

func (s *Service) validatePermissions(ctx context.Context, permissions []string) error {
	for _, permission := range permissions {
		if !isPhaseScopedProjectPermission(permission) {
			return domain.ErrInvalidArgument
		}
	}
	if len(permissions) == 0 || s.permissions == nil {
		return nil
	}

	found, err := s.permissions.ListByNames(ctx, permissions)
	if err != nil {
		return err
	}
	if len(found) != len(permissions) {
		return domain.ErrNotFound
	}
	return nil
}

func normalizePermissions(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		permission := strings.TrimSpace(value)
		if permission == "" {
			continue
		}
		if _, ok := seen[permission]; ok {
			continue
		}
		seen[permission] = struct{}{}
		out = append(out, permission)
	}
	sort.Strings(out)
	return out
}

func isPhaseScopedProjectPermission(permission string) bool {
	if !strings.HasPrefix(permission, "project.") {
		return false
	}
	switch permission {
	case domainUser.PermissionProjectCreate, domainUser.PermissionProjectListAll:
		return false
	default:
		return !strings.HasSuffix(permission, ".edit")
	}
}
