package user

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/besart951/go_infra_link/backend/internal/repository/searchspec"
	"gorm.io/gorm"
)

type permissionRepo struct {
	*gormbase.BaseRepository[*domainUser.Permission]
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) domainUser.PermissionRepository {
	baseRepo := gormbase.NewBaseRepository(db,
		gormbase.TrigramSearchCallback[*domainUser.Permission](searchspec.Permissions.SearchColumns("")...),
	)
	return &permissionRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *permissionRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainUser.Permission], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 50)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *permissionRepo) GetByName(ctx context.Context, name string) (*domainUser.Permission, error) {
	var perm domainUser.Permission
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&perm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &perm, nil
}

func (r *permissionRepo) ListAll(ctx context.Context) ([]domainUser.Permission, error) {
	var perms []domainUser.Permission
	err := r.db.WithContext(ctx).Order("name ASC").Find(&perms).Error
	return perms, err
}

func (r *permissionRepo) ListByNames(ctx context.Context, names []string) ([]domainUser.Permission, error) {
	if len(names) == 0 {
		return []domainUser.Permission{}, nil
	}
	var perms []domainUser.Permission
	err := r.db.WithContext(ctx).Where("name IN ?", names).Find(&perms).Error
	return perms, err
}
