package user

import (
	"context"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"gorm.io/gorm"
)

type rolePermissionRepo struct {
	*gormbase.BaseRepository[*domainUser.RolePermission]
	db *gorm.DB
}

func NewRolePermissionRepository(db *gorm.DB) domainUser.RolePermissionRepository {
	baseRepo := gormbase.NewBaseRepository[*domainUser.RolePermission](db, nil)
	return &rolePermissionRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *rolePermissionRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainUser.RolePermission], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 50)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *rolePermissionRepo) ListByRole(ctx context.Context, role domainUser.Role) ([]domainUser.RolePermission, error) {
	var perms []domainUser.RolePermission
	err := r.db.WithContext(ctx).Where("role = ?", role).Find(&perms).Error
	return perms, err
}

func (r *rolePermissionRepo) ListByRoles(ctx context.Context, roles []domainUser.Role) ([]domainUser.RolePermission, error) {
	if len(roles) == 0 {
		return []domainUser.RolePermission{}, nil
	}
	var perms []domainUser.RolePermission
	err := r.db.WithContext(ctx).Where("role IN ?", roles).Find(&perms).Error
	return perms, err
}

func (r *rolePermissionRepo) ReplaceRolePermissions(ctx context.Context, role domainUser.Role, permissions []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role = ?", role).Delete(&domainUser.RolePermission{}).Error; err != nil {
			return err
		}

		if len(permissions) == 0 {
			return nil
		}

		now := time.Now().UTC()
		items := make([]domainUser.RolePermission, len(permissions))
		for i, perm := range permissions {
			items[i] = domainUser.RolePermission{
				Role:       role,
				Permission: perm,
			}
			if err := items[i].Base.InitForCreate(now); err != nil {
				return err
			}
		}

		return tx.Create(&items).Error
	})
}

func (r *rolePermissionRepo) AddPermissionToRole(ctx context.Context, role domainUser.Role, permission string) (*domainUser.RolePermission, error) {
	if err := r.db.WithContext(ctx).
		Where("role = ? AND permission = ?", role, permission).
		Delete(&domainUser.RolePermission{}).Error; err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	item := domainUser.RolePermission{
		Role:       role,
		Permission: permission,
	}
	if err := item.Base.InitForCreate(now); err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *rolePermissionRepo) RemovePermissionFromRole(ctx context.Context, role domainUser.Role, permission string) error {
	return r.db.WithContext(ctx).
		Where("role = ? AND permission = ?", role, permission).
		Delete(&domainUser.RolePermission{}).Error
}

func (r *rolePermissionRepo) DeleteByPermissionName(ctx context.Context, permission string) error {
	return r.db.WithContext(ctx).
		Where("permission = ?", permission).
		Delete(&domainUser.RolePermission{}).Error
}
