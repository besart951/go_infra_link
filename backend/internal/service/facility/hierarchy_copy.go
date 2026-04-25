package facility

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// HierarchyCopier keeps existing service wiring stable while projectFacilityCopy
// owns copy planning and execution.
type HierarchyCopier struct {
	controlCabinetRepo      domainFacility.ControlCabinetRepository
	buildingRepo            domainFacility.BuildingRepository
	spsControllerRepo       domainFacility.SPSControllerRepository
	systemTypeRepo          domainFacility.SystemTypeRepository
	spsControllerSystemRepo domainFacility.SPSControllerSystemTypeStore
	fieldDeviceRepo         domainFacility.FieldDeviceStore
	specificationRepo       domainFacility.SpecificationStore
	bacnetObjectRepo        domainFacility.BacnetObjectStore
	tx                      txCoordinator
}

func NewHierarchyCopier(
	controlCabinetRepo domainFacility.ControlCabinetRepository,
	buildingRepo domainFacility.BuildingRepository,
	spsControllerRepo domainFacility.SPSControllerRepository,
	systemTypeRepo domainFacility.SystemTypeRepository,
	spsControllerSystemRepo domainFacility.SPSControllerSystemTypeStore,
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	specificationRepo domainFacility.SpecificationStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
) *HierarchyCopier {
	return &HierarchyCopier{
		controlCabinetRepo:      controlCabinetRepo,
		buildingRepo:            buildingRepo,
		spsControllerRepo:       spsControllerRepo,
		systemTypeRepo:          systemTypeRepo,
		spsControllerSystemRepo: spsControllerSystemRepo,
		fieldDeviceRepo:         fieldDeviceRepo,
		specificationRepo:       specificationRepo,
		bacnetObjectRepo:        bacnetObjectRepo,
	}
}

func (c *HierarchyCopier) bindTransactions(tx txCoordinator) {
	c.tx = tx
}

func (c *HierarchyCopier) transaction() facilityTx[*HierarchyCopier] {
	return newFacilityTx(c.tx, c, func(services *Services) *HierarchyCopier {
		return services.HierarchyCopier
	})
}

func (c *HierarchyCopier) projectFacilityCopy() projectFacilityCopy {
	return projectFacilityCopy{
		controlCabinetRepo:      c.controlCabinetRepo,
		buildingRepo:            c.buildingRepo,
		spsControllerRepo:       c.spsControllerRepo,
		systemTypeRepo:          c.systemTypeRepo,
		spsControllerSystemRepo: c.spsControllerSystemRepo,
		fieldDeviceRepo:         c.fieldDeviceRepo,
		specificationRepo:       c.specificationRepo,
		bacnetObjectRepo:        c.bacnetObjectRepo,
	}
}

func (c *HierarchyCopier) CopyControlCabinetByID(ctx context.Context, id uuid.UUID) (*domainFacility.ControlCabinet, error) {
	return runWithFacilityTxResult(c.transaction(), func(copier *HierarchyCopier) (*domainFacility.ControlCabinet, error) {
		return copier.projectFacilityCopy().copyControlCabinetByID(ctx, id)
	})
}

func (c *HierarchyCopier) CopySPSControllerByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSController, error) {
	return runWithFacilityTxResult(c.transaction(), func(copier *HierarchyCopier) (*domainFacility.SPSController, error) {
		return copier.projectFacilityCopy().copySPSControllerByID(ctx, id)
	})
}

func (c *HierarchyCopier) CopySPSControllerSystemTypeByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	return runWithFacilityTxResult(c.transaction(), func(copier *HierarchyCopier) (*domainFacility.SPSControllerSystemType, error) {
		return copier.projectFacilityCopy().copySPSControllerSystemTypeByID(ctx, id)
	})
}
