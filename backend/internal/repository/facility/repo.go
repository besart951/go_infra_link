package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"gorm.io/gorm"
)

type facilityRepo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) facility.Repository {
	return &facilityRepo{db: db}
}
