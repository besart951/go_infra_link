package facility

import (
	"context"
	"fmt"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/service/changecapture"
	"github.com/google/uuid"
)

func (s *FieldDeviceService) ImportAggregate(ctx context.Context, aggregate domainFacility.FieldDeviceImportAggregate) error {
	return s.transaction().run(ctx, func(txCtx context.Context, txService *FieldDeviceService) error {
		return txService.importAggregateInTx(txCtx, aggregate)
	})
}

func (s *FieldDeviceService) importAggregateInTx(ctx context.Context, aggregate domainFacility.FieldDeviceImportAggregate) error {
	if err := s.validateImportAggregate(ctx, &aggregate); err != nil {
		return err
	}
	if err := s.repo.Create(ctx, &aggregate.FieldDevice); err != nil {
		return err
	}
	if err := s.createImportedSpecification(ctx, aggregate); err != nil {
		return err
	}
	if err := s.createImportedBacnetGraph(ctx, aggregate); err != nil {
		return err
	}
	return s.recordFieldDeviceChange(ctx, changecapture.ActionCreated, aggregate.FieldDevice.ID)
}

func (s *FieldDeviceService) validateImportAggregate(ctx context.Context, aggregate *domainFacility.FieldDeviceImportAggregate) error {
	if aggregate == nil || aggregate.FieldDevice.ID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	if err := s.Validate(ctx, &aggregate.FieldDevice, nil); err != nil {
		return err
	}
	if aggregate.Specification != nil {
		if err := s.validateSpecification(aggregate.Specification); err != nil {
			return err
		}
	}
	return validateImportedBacnetObjects(aggregate.FieldDevice.ID, aggregate.BacnetObjects)
}

func validateImportedBacnetObjects(fieldDeviceID uuid.UUID, objects []domainFacility.BacnetObject) error {
	ids := make(map[uuid.UUID]struct{}, len(objects))
	for index := range objects {
		if objects[index].ID == uuid.Nil || normalizeBacnetTextFix(objects[index].TextFix) == "" {
			return domain.NewValidationError().Add(fmt.Sprintf("bacnet_objects.%d", index), "source_id and text_fix are required")
		}
		ids[objects[index].ID] = struct{}{}
	}
	for index := range objects {
		if err := validateImportedBacnetObject(fieldDeviceID, &objects[index], ids); err != nil {
			return err
		}
	}
	return nil
}

func validateImportedBacnetObject(fieldDeviceID uuid.UUID, object *domainFacility.BacnetObject, ids map[uuid.UUID]struct{}) error {
	if object.FieldDeviceID == nil || *object.FieldDeviceID != fieldDeviceID {
		return domain.NewValidationError().Add("bacnet_objects.field_device_id", "BACnet object must belong to the imported field device")
	}
	if object.SoftwareReferenceID != nil {
		if _, ok := ids[*object.SoftwareReferenceID]; !ok {
			return domain.NewValidationError().Add("bacnet_objects.software_reference_id", "software reference must target the same field device")
		}
	}
	for index := range object.AlarmValues {
		if object.AlarmValues[index].BacnetObjectID != object.ID {
			return domain.NewValidationError().Add("bacnet_objects.alarm_values", "alarm value owner does not match BACnet object")
		}
	}
	return nil
}

func (s *FieldDeviceService) createImportedSpecification(ctx context.Context, aggregate domainFacility.FieldDeviceImportAggregate) error {
	if aggregate.Specification == nil {
		return nil
	}
	deviceID := aggregate.FieldDevice.ID
	aggregate.Specification.FieldDeviceID = &deviceID
	return s.specificationRepo.Create(ctx, aggregate.Specification)
}

func (s *FieldDeviceService) createImportedBacnetGraph(ctx context.Context, aggregate domainFacility.FieldDeviceImportAggregate) error {
	objects := make([]*domainFacility.BacnetObject, len(aggregate.BacnetObjects))
	values := make([]*domainFacility.BacnetObjectAlarmValue, 0)
	for index := range aggregate.BacnetObjects {
		objects[index] = &aggregate.BacnetObjects[index]
		for valueIndex := range aggregate.BacnetObjects[index].AlarmValues {
			value := &aggregate.BacnetObjects[index].AlarmValues[valueIndex]
			value.Source = domainFacility.AlarmValueSourceImport
			values = append(values, value)
		}
	}
	if len(objects) > 0 {
		if err := s.bacnetObjectRepo.BulkCreate(ctx, objects, 200); err != nil {
			return err
		}
	}
	if len(values) == 0 {
		return nil
	}
	return s.bacnetAlarmValueRepo.BulkCreate(ctx, values, 500)
}
