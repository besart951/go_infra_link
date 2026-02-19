package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type BacnetAlarmValueService struct {
	valueRepo     domainFacility.BacnetObjectAlarmValueRepository
	alarmTypeRepo domainFacility.AlarmTypeRepository
	bacnetRepo    domainFacility.BacnetObjectRepository
	alarmDefRepo  domainFacility.AlarmDefinitionRepository
}

func NewBacnetAlarmValueService(
	valueRepo domainFacility.BacnetObjectAlarmValueRepository,
	alarmTypeRepo domainFacility.AlarmTypeRepository,
	bacnetRepo domainFacility.BacnetObjectRepository,
	alarmDefRepo domainFacility.AlarmDefinitionRepository,
) *BacnetAlarmValueService {
	return &BacnetAlarmValueService{
		valueRepo:     valueRepo,
		alarmTypeRepo: alarmTypeRepo,
		bacnetRepo:    bacnetRepo,
		alarmDefRepo:  alarmDefRepo,
	}
}

// GetSchema returns the alarm field schema for a BacnetObject
func (s *BacnetAlarmValueService) GetSchema(bacnetObjectID uuid.UUID) (*domainFacility.AlarmType, error) {
	bacnetObjs, err := s.bacnetRepo.GetByIds([]uuid.UUID{bacnetObjectID})
	if err != nil || len(bacnetObjs) == 0 {
		return nil, err
	}
	bo := bacnetObjs[0]
	if bo.AlarmDefinitionID == nil {
		return nil, nil
	}

	alarmDefSlice, err := s.alarmDefRepo.GetByIds([]uuid.UUID{*bo.AlarmDefinitionID})
	if err != nil || len(alarmDefSlice) == 0 {
		return nil, err
	}
	alarmDef := alarmDefSlice[0]
	if alarmDef.AlarmTypeID == nil {
		return nil, nil
	}

	return s.alarmTypeRepo.GetWithFields(*alarmDef.AlarmTypeID)
}

// GetValues returns the stored alarm values for a BacnetObject
func (s *BacnetAlarmValueService) GetValues(bacnetObjectID uuid.UUID) ([]domainFacility.BacnetObjectAlarmValue, error) {
	return s.valueRepo.GetByBacnetObjectID(bacnetObjectID)
}

// PutValues replaces all alarm values for a BacnetObject
func (s *BacnetAlarmValueService) PutValues(bacnetObjectID uuid.UUID, values []domainFacility.BacnetObjectAlarmValue) error {
	return s.valueRepo.ReplaceForBacnetObject(bacnetObjectID, values)
}
