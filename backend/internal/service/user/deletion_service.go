package user

import (
	"context"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

// DeletionService owns the complete soft-delete lifecycle: Delete, Restore, Anonymize.
// It consolidates guards, cascade logic, and audit tracking in one place.
type DeletionService struct {
	repo   Store
	policy MutationPolicy
}

// NewDeletionService creates a new DeletionService.
func NewDeletionService(repo Store, policy MutationPolicy) *DeletionService {
	return &DeletionService{
		repo:   repo,
		policy: policy,
	}
}

// DeleteByID soft-deletes a user with no actor ID (system deletion).
// Returns ErrUserAlreadyAnonymized if user is anonymized.
// Returns nil if user is already deleted (idempotent).
func (d *DeletionService) DeleteByID(ctx context.Context, id uuid.UUID) error {
	usr, err := d.getUser(ctx, id)
	if err != nil {
		return err
	}

	// Guard: cannot delete anonymized users
	if usr.IsAnonymized() {
		return user.ErrUserAlreadyAnonymized
	}

	// Idempotent: already deleted
	if usr.IsDeleted() {
		return nil
	}

	// Mark for deletion
	markUserDeleted(usr, uuid.Nil, time.Now().UTC())

	// Persist
	return d.repo.Update(ctx, usr)
}

// DeleteByIDForActor soft-deletes a user, checking actor's permissions.
// Returns ErrRoleNotAssignable if actor cannot delete.
// Returns ErrUserAlreadyAnonymized if user is anonymized.
// Returns nil if user is already deleted (idempotent).
func (d *DeletionService) DeleteByIDForActor(ctx context.Context, actorID, userID uuid.UUID) error {
	if d.policy == nil {
		return user.ErrRoleNotAssignable
	}

	usr, err := d.getUser(ctx, userID)
	if err != nil {
		return err
	}

	// Guard: cannot delete anonymized users
	if usr.IsAnonymized() {
		return user.ErrUserAlreadyAnonymized
	}

	// Idempotent: already deleted
	if usr.IsDeleted() {
		return nil
	}

	// Guard: check permission
	if err := d.policy.CanDeleteUser(ctx, actorID, *usr); err != nil {
		return err
	}

	// Mark for deletion with audit trail
	markUserDeleted(usr, actorID, time.Now().UTC())

	// Persist
	return d.repo.Update(ctx, usr)
}

// RestoreByIDForActor restores a soft-deleted user, checking:
// - Restore window hasn't expired
// - Actor has permission to restore
// Returns ErrRestoreWindowExpired if restore window has passed.
// Returns ErrUserAlreadyAnonymized if user is anonymized (cannot restore).
// Returns nil if user is not deleted (idempotent).
func (d *DeletionService) RestoreByIDForActor(ctx context.Context, actorID, userID uuid.UUID) error {
	if d.policy == nil {
		return user.ErrRoleNotAssignable
	}

	usr, err := d.getUser(ctx, userID)
	if err != nil {
		return err
	}

	// Guard: cannot restore anonymized users
	if usr.IsAnonymized() {
		return user.ErrUserAlreadyAnonymized
	}

	// Idempotent: not deleted
	if !usr.IsDeleted() {
		return nil
	}

	// Guard: check restore window
	if usr.RestoreUntil == nil || !time.Now().UTC().Before(*usr.RestoreUntil) {
		return user.ErrRestoreWindowExpired
	}

	// Guard: check permission
	if err := d.policy.CanRestoreUser(ctx, actorID, *usr); err != nil {
		return err
	}

	// Clear deletion markers
	clearDeleteMarkers(usr)
	usr.IsActive = true
	usr.DisabledAt = nil

	// Persist
	return d.repo.Update(ctx, usr)
}

// Anonymize completes the soft-delete lifecycle by anonymizing user identity.
// Called by background worker after ScheduledPurgeAt.
// Does not check permissions (automated process).
func (d *DeletionService) Anonymize(ctx context.Context, userID uuid.UUID) error {
	usr, err := d.getUser(ctx, userID)
	if err != nil {
		return err
	}

	// Guard: cannot anonymize non-deleted users
	if !usr.IsDeleted() {
		return nil
	}

	// Guard: already anonymized
	if usr.IsAnonymized() {
		return nil
	}

	// Anonymize identity
	anonymizeUser(usr, time.Now().UTC())

	// Persist
	return d.repo.Update(ctx, usr)
}

// getUser is a helper that fetches a user using the domain helper.
func (d *DeletionService) getUser(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return domain.GetByID(ctx, d.repo, id)
}
