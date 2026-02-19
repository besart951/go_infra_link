package facilitysql

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bacnetObjectAlarmValueRepo struct {
	*gormbase.BaseRepository[*domainFacility.BacnetObjectAlarmValue]
	db *gorm.DB
}

func NewBacnetObjectAlarmValueRepository(db *gorm.DB) domainFacility.BacnetObjectAlarmValueRepository {
	return &bacnetObjectAlarmValueRepo{
		BaseRepository: gormbase.NewBaseRepository[*domainFacility.BacnetObjectAlarmValue](db, nil),
		db:             db,
	}
}

func (r *bacnetObjectAlarmValueRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.BacnetObjectAlarmValue], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 50)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *bacnetObjectAlarmValueRepo) GetByBacnetObjectID(bacnetObjectID uuid.UUID) ([]domainFacility.BacnetObjectAlarmValue, error) {
	var values []domainFacility.BacnetObjectAlarmValue
	err := r.db.
		Preload("AlarmTypeField").
		Preload("AlarmTypeField.AlarmField").
		Preload("Unit").
		Where("bacnet_object_id = ?", bacnetObjectID).
		Find(&values).Error
	return values, err
}

func (r *bacnetObjectAlarmValueRepo) ReplaceForBacnetObject(bacnetObjectID uuid.UUID, values []domainFacility.BacnetObjectAlarmValue) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("bacnet_object_id = ?", bacnetObjectID).Delete(&domainFacility.BacnetObjectAlarmValue{}).Error; err != nil {
			return err
		}

		if len(values) == 0 {
			return nil
		}

		now := time.Now().UTC()
		for i := range values {
			values[i].BacnetObjectID = bacnetObjectID
			if err := values[i].GetBase().InitForCreate(now); err != nil {
				return err
			}
		}

		return tx.Create(&values).Error
	})
}
