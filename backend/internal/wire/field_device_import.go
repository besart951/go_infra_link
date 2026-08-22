package wire

import (
	"context"

	fielddeviceimport "github.com/besart951/go_infra_link/backend/internal/application/fielddeviceimport"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	importing "github.com/besart951/go_infra_link/backend/internal/infrastructure/importing"
	importsql "github.com/besart951/go_infra_link/backend/internal/repository/importsql"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

func newFieldDeviceImportService(runtime *RuntimeAdapters, fieldDevices *facilityservice.FieldDeviceService) *fielddeviceimport.Service {
	if runtime == nil || runtime.DB == nil || fieldDevices == nil {
		return nil
	}
	return fielddeviceimport.NewService(
		importing.NewArchiveReader(importing.NewExcelizeReader()),
		importsql.NewStore(runtime.DB),
		fieldDeviceAggregateWriter{service: fieldDevices},
	)
}

type fieldDeviceAggregateWriter struct {
	service *facilityservice.FieldDeviceService
}

func (w fieldDeviceAggregateWriter) Import(ctx context.Context, aggregate fielddeviceimport.Aggregate) error {
	references := make(map[uuid.UUID]uuid.UUID, len(aggregate.SoftwareReferences))
	for _, reference := range aggregate.SoftwareReferences {
		references[reference.SourceObjectID] = reference.TargetObjectID
	}
	for index := range aggregate.BacnetObjects {
		target, ok := references[aggregate.BacnetObjects[index].ID]
		if ok {
			aggregate.BacnetObjects[index].SoftwareReferenceID = &target
		}
	}
	return w.service.ImportAggregate(ctx, domainFacility.FieldDeviceImportAggregate{
		FieldDevice: aggregate.FieldDevice, Specification: aggregate.Specification, BacnetObjects: aggregate.BacnetObjects,
	})
}
