package usermutationpolicy

import (
	"context"
	"errors"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type RoleProvider interface {
	GetGlobalRole(ctx context.Context, userID uuid.UUID) (domainUser.Role, error)
	HasPermission(ctx context.Context, role domainUser.Role, permission string) (bool, error)
}

type InvitationReader interface {
	GetInvitationByUserID(ctx context.Context, userID uuid.UUID) (*domainUser.UserInvitation, error)
}

type Policy struct {
	roles       RoleProvider
	invitations InvitationReader
}

func New(roles RoleProvider, invitations InvitationReader) *Policy {
	return &Policy{roles: roles, invitations: invitations}
}

func (p *Policy) CanDirectCreateUser(ctx context.Context, actorID uuid.UUID, targetRole domainUser.Role) error {
	actorRole, err := p.actorRole(ctx, actorID)
	if err != nil {
		return err
	}
	if !domainUser.IsValidRole(targetRole) {
		return domain.ErrInvalidArgument
	}
	if actorRole != domainUser.RoleSuperAdmin {
		return domainUser.ErrRoleNotAssignable
	}
	return nil
}

func (p *Policy) CanInviteUser(ctx context.Context, actorID uuid.UUID, targetRole domainUser.Role) error {
	return p.canMutateRole(ctx, actorID, targetRole, domainUser.PermissionUserCreate)
}

func (p *Policy) CanReadRegistrationProcess(ctx context.Context, actorID uuid.UUID, target domainUser.User) error {
	if actorID == target.ID {
		return nil
	}
	return p.canMutateRole(ctx, actorID, target.Role, domainUser.PermissionUserRead)
}

func (p *Policy) CanUpdateProfile(ctx context.Context, actorID uuid.UUID, target domainUser.User) error {
	return p.canMutateRole(ctx, actorID, target.Role, domainUser.PermissionUserUpdate)
}

func (p *Policy) CanDeleteUser(ctx context.Context, actorID uuid.UUID, target domainUser.User) error {
	return p.canMutateRole(ctx, actorID, target.Role, domainUser.PermissionUserDelete)
}

func (p *Policy) CanRestoreUser(ctx context.Context, actorID uuid.UUID, target domainUser.User) error {
	if err := p.canMutateRole(ctx, actorID, target.Role, domainUser.PermissionUserDelete); err != nil {
		return err
	}
	actorRole, err := p.actorRole(ctx, actorID)
	if err != nil {
		return err
	}
	hasPermission, err := p.roles.HasPermission(ctx, actorRole, domainUser.PermissionUserReadDeleted)
	if err != nil {
		return err
	}
	if !hasPermission {
		return domainUser.ErrRoleNotAssignable
	}
	return nil
}

func (p *Policy) CanDisableUser(ctx context.Context, actorID uuid.UUID, target domainUser.User) error {
	return p.canMutateRole(ctx, actorID, target.Role, domainUser.PermissionUserUpdate)
}

func (p *Policy) CanEnableUser(ctx context.Context, actorID uuid.UUID, target domainUser.User) error {
	if err := p.canMutateRole(ctx, actorID, target.Role, domainUser.PermissionUserUpdate); err != nil {
		return err
	}
	return p.ensureRegistrationComplete(ctx, target.ID)
}

func (p *Policy) CanAssignRole(ctx context.Context, actorID uuid.UUID, currentRole, nextRole domainUser.Role) error {
	if err := p.canMutateRole(ctx, actorID, currentRole, domainUser.PermissionUserUpdate); err != nil {
		return err
	}
	return p.canMutateRole(ctx, actorID, nextRole, domainUser.PermissionUserUpdate)
}

func (p *Policy) canMutateRole(ctx context.Context, actorID uuid.UUID, targetRole domainUser.Role, permission string) error {
	actorRole, err := p.actorRole(ctx, actorID)
	if err != nil {
		return err
	}
	if !domainUser.IsValidRole(targetRole) {
		return domain.ErrInvalidArgument
	}
	if p.roles == nil {
		return domainUser.ErrRoleNotAssignable
	}
	hasPermission, err := p.roles.HasPermission(ctx, actorRole, permission)
	if err != nil {
		return err
	}
	if !hasPermission {
		return domainUser.ErrRoleNotAssignable
	}
	if actorRole == domainUser.RoleSuperAdmin {
		return nil
	}
	if domainUser.RoleLevel(targetRole) >= domainUser.RoleLevel(actorRole) {
		return domainUser.ErrRoleNotAssignable
	}
	return nil
}

func (p *Policy) actorRole(ctx context.Context, actorID uuid.UUID) (domainUser.Role, error) {
	if actorID == uuid.Nil {
		return "", domain.ErrInvalidArgument
	}
	if p.roles == nil {
		return "", domainUser.ErrRoleNotAssignable
	}
	role, err := p.roles.GetGlobalRole(ctx, actorID)
	if err != nil {
		return "", err
	}
	if !domainUser.IsValidRole(role) {
		return "", domainUser.ErrRoleNotAssignable
	}
	return role, nil
}

func (p *Policy) ensureRegistrationComplete(ctx context.Context, userID uuid.UUID) error {
	if p.invitations == nil {
		return nil
	}
	invitation, err := p.invitations.GetInvitationByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}
	if invitation.AcceptedAt == nil {
		return domainUser.ErrRegistrationPending
	}
	return nil
}
