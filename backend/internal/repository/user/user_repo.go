package user

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domainUser.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) GetByIds(ids []uuid.UUID) ([]*domainUser.User, error) {
	var users []*domainUser.User
	err := r.db.Preload("CreatedBy").Preload("BusinessDetails").Where("id IN ?", ids).Find(&users).Error
	return users, err
}

func (r *userRepo) Create(entity *domainUser.User) error {
	return r.db.Create(entity).Error
}

func (r *userRepo) Update(entity *domainUser.User) error {
	return r.db.Save(entity).Error
}

func (r *userRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&domainUser.User{}).Error
}

func (r *userRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainUser.User], error) {
	searchFields := []string{"first_name", "last_name", "email"}
	return repository.Paginate[domainUser.User](r.db, params, searchFields)
}
