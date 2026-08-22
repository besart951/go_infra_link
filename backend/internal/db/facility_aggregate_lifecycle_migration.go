package db

import (
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"gorm.io/gorm"
)

func migrateFacilityAggregateLifecycle(db *gorm.DB) error {
	return facilityservice.MigrateFacilityAggregateLifecycle(db)
}
