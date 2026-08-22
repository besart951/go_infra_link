package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	"github.com/google/uuid"
)

type BacnetAlarmValueService struct {
	valueRepo     domainFacility.BacnetObjectAlarmValueRepository
	alarmTypeRepo domainFacility.AlarmTypeRepository
	bacnetRepo    domainFacility.BacnetObjectRepository
	templateStore domainObjectData.BacnetObjectTemplateStore
	writer        *BacnetObjectService
}

type BacnetAlarmValueDependencies struct {
	Values     domainFacility.BacnetObjectAlarmValueRepository
	AlarmTypes domainFacility.AlarmTypeRepository
	Objects    domainFacility.BacnetObjectRepository
	Templates  domainObjectData.BacnetObjectTemplateStore
	Writer     *BacnetObjectService
}

func NewBacnetAlarmValueService(deps BacnetAlarmValueDependencies) *BacnetAlarmValueService {
	return &BacnetAlarmValueService{
		valueRepo:     deps.Values,
		alarmTypeRepo: deps.AlarmTypes,
		bacnetRepo:    deps.Objects,
		templateStore: deps.Templates,
		writer:        deps.Writer,
	}
}

// GetSchema returns the alarm field schema for a BacnetObject
func (s *BacnetAlarmValueService) GetSchema(ctx context.Context, bacnetObjectID uuid.UUID) (*domainFacility.AlarmType, error) {
	bacnetObjs, err := s.bacnetRepo.GetByIds(ctx, []uuid.UUID{bacnetObjectID})
	if err != nil {
		return nil, err
	}
	alarmTypeID, err := s.alarmTypeID(ctx, bacnetObjectID, bacnetObjs)
	if err != nil || alarmTypeID == nil {
		return nil, err
	}
	return s.alarmTypeRepo.GetWithFields(ctx, *alarmTypeID)
}

func (s *BacnetAlarmValueService) alarmTypeID(ctx context.Context, id uuid.UUID, instances []*domainFacility.BacnetObject) (*uuid.UUID, error) {
	if len(instances) > 0 {
		return instances[0].AlarmTypeID, nil
	}
	template, err := s.templateStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return template.AlarmTypeID, nil
}

// GetValues returns the stored alarm values for a BacnetObject
func (s *BacnetAlarmValueService) GetValues(ctx context.Context, bacnetObjectID uuid.UUID) (*domainFacility.BacnetAlarmValues, error) {
	instances, err := s.bacnetRepo.GetByIds(ctx, []uuid.UUID{bacnetObjectID})
	if err != nil {
		return nil, err
	}
	if len(instances) > 0 {
		values, valuesErr := s.valueRepo.GetByBacnetObjectID(ctx, bacnetObjectID)
		return &domainFacility.BacnetAlarmValues{Version: instances[0].Version, Items: values}, valuesErr
	}
	return s.templateValues(ctx, bacnetObjectID)
}

// PutValues replaces all alarm values for a BacnetObject
func (s *BacnetAlarmValueService) PutValues(ctx context.Context, bacnetObjectID uuid.UUID, version uint64, values []domainFacility.BacnetObjectAlarmValue) (uint64, error) {
	return s.writer.ReplaceAlarmValues(ctx, bacnetObjectID, version, values)
}

func (s *BacnetAlarmValueService) templateValues(ctx context.Context, id uuid.UUID) (*domainFacility.BacnetAlarmValues, error) {
	template, err := s.templateStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	values := make([]domainFacility.BacnetObjectAlarmValue, len(template.AlarmValues))
	for index, value := range template.AlarmValues {
		values[index] = domainFacility.BacnetObjectAlarmValue{
			Base:           domain.Base{ID: value.ID, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, Version: value.Version},
			BacnetObjectID: id, AlarmTypeFieldID: value.AlarmTypeFieldID, ValueNumber: value.ValueNumber,
			ValueInteger: value.ValueInteger, ValueBoolean: value.ValueBoolean, ValueString: value.ValueString,
			ValueJSON: value.ValueJSON, UnitID: value.UnitID, Source: value.Source,
		}
	}
	return &domainFacility.BacnetAlarmValues{Version: template.Version, Items: values}, nil
}
