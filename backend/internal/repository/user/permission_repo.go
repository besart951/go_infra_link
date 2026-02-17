package user

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"gorm.io/gorm"
)

type permissionRepo struct {
	*gormbase.BaseRepository[*domainUser.Permission]
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) domainUser.PermissionRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where(
			"LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(resource) LIKE ? OR LOWER(action) LIKE ?",
			pattern,
			pattern,
			pattern,
			pattern,
		)
	}

	baseRepo := gormbase.NewBaseRepository[*domainUser.Permission](db, searchCallback)
	return &permissionRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *permissionRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainUser.Permission], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 50)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *permissionRepo) GetByName(name string) (*domainUser.Permission, error) {
	var perm domainUser.Permission
	err := r.db.Where("deleted_at IS NULL").Where("name = ?", name).First(&perm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &perm, nil
}

func (r *permissionRepo) ListAll() ([]domainUser.Permission, error) {
	var perms []domainUser.Permission
	err := r.db.Where("deleted_at IS NULL").Order("name ASC").Find(&perms).Error
	return perms, err
}

func (r *permissionRepo) ListByNames(names []string) ([]domainUser.Permission, error) {
	if len(names) == 0 {
		return []domainUser.Permission{}, nil
	}
	var perms []domainUser.Permission
	err := r.db.Where("deleted_at IS NULL").Where("name IN ?", names).Find(&perms).Error
	return perms, err
}
