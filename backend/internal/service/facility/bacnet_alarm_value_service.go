package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type BacnetAlarmValueService struct {
	valueRepo     domainFacility.BacnetObjectAlarmValueRepository
	alarmTypeRepo domainFacility.AlarmTypeRepository
	bacnetRepo    domainFacility.BacnetObjectRepository
}

func NewBacnetAlarmValueService(
	valueRepo domainFacility.BacnetObjectAlarmValueRepository,
	alarmTypeRepo domainFacility.AlarmTypeRepository,
	bacnetRepo domainFacility.BacnetObjectRepository,
) *BacnetAlarmValueService {
	return &BacnetAlarmValueService{
		valueRepo:     valueRepo,
		alarmTypeRepo: alarmTypeRepo,
		bacnetRepo:    bacnetRepo,
	}
}

// GetSchema returns the alarm field schema for a BacnetObject
func (s *BacnetAlarmValueService) GetSchema(bacnetObjectID uuid.UUID) (*domainFacility.AlarmType, error) {
	bacnetObjs, err := s.bacnetRepo.GetByIds([]uuid.UUID{bacnetObjectID})
	if err != nil || len(bacnetObjs) == 0 {
		return nil, err
	}
	bo := bacnetObjs[0]
	if bo.AlarmTypeID == nil {
		return nil, nil
	}

	return s.alarmTypeRepo.GetWithFields(*bo.AlarmTypeID)
}

// GetValues returns the stored alarm values for a BacnetObject
func (s *BacnetAlarmValueService) GetValues(bacnetObjectID uuid.UUID) ([]domainFacility.BacnetObjectAlarmValue, error) {
	return s.valueRepo.GetByBacnetObjectID(bacnetObjectID)
}

// PutValues replaces all alarm values for a BacnetObject
func (s *BacnetAlarmValueService) PutValues(bacnetObjectID uuid.UUID, values []domainFacility.BacnetObjectAlarmValue) error {
	return s.valueRepo.ReplaceForBacnetObject(bacnetObjectID, values)
}
