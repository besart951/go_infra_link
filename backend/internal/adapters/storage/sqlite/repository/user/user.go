package userrepo

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/adapters/storage"
	"github.com/besart951/go_infra_link/backend/internal/core/domain/user"
	"gorm.io/gorm"
)

type UserStorage struct {
	storage.BaseRepository[user.User]
}

func NewUserStorage(db *gorm.DB) user.UserRepository {
	return &UserStorage{
		BaseRepository: storage.BaseRepository[user.User]{DB: db},
	}
}

// Implement Create, Update, Delete, GetByID, GetAll via embedding BaseRepository
// But GetByUsername is specific to User, so we implement it here.

func (r *UserStorage) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	var u user.User
	err := r.DB.WithContext(ctx).Where("username = ?", username).First(&u).Error
	return &u, err
}
