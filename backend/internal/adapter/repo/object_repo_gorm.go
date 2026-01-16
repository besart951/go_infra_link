package repo

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"gorm.io/gorm"
)

type ObjectRepository struct {
	db *gorm.DB
}

func NewObjectRepository(db *gorm.DB) *ObjectRepository {
	return &ObjectRepository{db: db}
}

func (r *ObjectRepository) Create(ctx context.Context, o *domain.Object) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *ObjectRepository) FindByID(ctx context.Context, id string) (*domain.Object, error) {
	var o domain.Object
	res := r.db.WithContext(ctx).First(&o, "id = ?", id)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}
	return &o, nil
}

func (r *ObjectRepository) GrantAccess(ctx context.Context, p *domain.ObjectPermission) error {
	return r.db.WithContext(ctx).Create(p).Error
}
