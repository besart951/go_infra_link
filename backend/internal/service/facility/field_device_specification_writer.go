package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/service/changecapture"
	"github.com/google/uuid"
)

type fieldDeviceSpecificationWriter struct {
	writer fieldDeviceWriter
}

func newFieldDeviceSpecificationWriter(writer fieldDeviceWriter) fieldDeviceSpecificationWriter {
	return fieldDeviceSpecificationWriter{writer: writer}
}

func (w fieldDeviceSpecificationWriter) createInTx(ctx context.Context, fieldDeviceID uuid.UUID, specification *domainFacility.Specification) error {
	fieldDevice, err := domain.GetByID(ctx, w.writer.service.repo, fieldDeviceID)
	if err != nil {
		return err
	}
	existing, err := w.writer.service.specificationRepo.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return domain.ErrConflict
	}
	normalizeSpecification(specification)
	if err := w.writer.service.validateSpecification(specification); err != nil {
		return err
	}
	return w.persistNewSpecification(ctx, fieldDevice, specification)
}

func (w fieldDeviceSpecificationWriter) persistNewSpecification(ctx context.Context, fieldDevice *domainFacility.FieldDevice, specification *domainFacility.Specification) error {
	fieldDeviceID := fieldDevice.ID
	specification.FieldDeviceID = &fieldDeviceID
	if err := w.writer.service.specificationRepo.Create(ctx, specification); err != nil {
		return err
	}
	fieldDevice.SpecificationID = &specification.ID
	if err := w.writer.service.repo.Update(ctx, fieldDevice); err != nil {
		return err
	}
	return w.writer.service.recordFieldDeviceChange(ctx, changecapture.ActionUpdated, fieldDevice.ID)
}

func (w fieldDeviceSpecificationWriter) updatePatchInTx(ctx context.Context, fieldDeviceID uuid.UUID, patch *domainFacility.SpecificationPatch) (*domainFacility.Specification, error) {
	fieldDevice, err := domain.GetByID(ctx, w.writer.service.repo, fieldDeviceID)
	if err != nil {
		return nil, err
	}
	spec, err := w.existingSpecification(ctx, fieldDeviceID)
	if err != nil {
		return nil, err
	}
	spec.Version = patch.BaseVersion
	applySpecificationPatch(spec, patch)
	if err := w.writer.service.validateSpecification(spec); err != nil {
		return nil, err
	}
	if err := w.persistSpecificationUpdate(ctx, fieldDevice, spec); err != nil {
		return nil, err
	}
	return spec, nil
}

func (w fieldDeviceSpecificationWriter) existingSpecification(ctx context.Context, fieldDeviceID uuid.UUID) (*domainFacility.Specification, error) {
	specifications, err := w.writer.service.specificationRepo.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	if err != nil {
		return nil, err
	}
	if len(specifications) == 0 {
		return nil, domain.ErrNotFound
	}
	return specifications[0], nil
}

func (w fieldDeviceSpecificationWriter) persistSpecificationUpdate(ctx context.Context, fieldDevice *domainFacility.FieldDevice, specification *domainFacility.Specification) error {
	if err := w.writer.service.specificationRepo.Update(ctx, specification); err != nil {
		return err
	}
	if err := w.writer.service.repo.Update(ctx, fieldDevice); err != nil {
		return err
	}
	return w.writer.service.recordFieldDeviceChange(ctx, changecapture.ActionUpdated, fieldDevice.ID)
}

func (w fieldDeviceSpecificationWriter) applyPatch(ctx context.Context, fieldDeviceID uuid.UUID, patch *domainFacility.SpecificationPatch) error {
	if patch == nil || !patch.HasChanges() {
		return nil
	}
	if _, err := domain.GetByID(ctx, w.writer.service.repo, fieldDeviceID); err != nil {
		return err
	}
	specifications, err := w.writer.service.specificationRepo.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	if err != nil {
		return err
	}
	if len(specifications) == 0 {
		return w.createFromPatch(ctx, fieldDeviceID, patch)
	}
	specification := specifications[0]
	applySpecificationPatch(specification, patch)
	if err := w.writer.service.validateSpecification(specification); err != nil {
		return err
	}
	return w.writer.service.specificationRepo.Update(ctx, specification)
}

func (w fieldDeviceSpecificationWriter) applyBulkPatch(ctx context.Context, fieldDevice *domainFacility.FieldDevice, patch *domainFacility.SpecificationPatch) error {
	if patch == nil || !patch.HasChanges() {
		return nil
	}
	specifications, err := w.writer.service.specificationRepo.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDevice.ID})
	if err != nil {
		return err
	}
	if len(specifications) == 0 {
		return w.createBulkSpecification(ctx, fieldDevice, patch)
	}
	specification := specifications[0]
	applySpecificationPatch(specification, patch)
	if err := w.writer.service.validateSpecification(specification); err != nil {
		return err
	}
	if err := w.writer.service.specificationRepo.Update(ctx, specification); err != nil {
		return err
	}
	fieldDevice.SpecificationID = &specification.ID
	fieldDevice.Specification = specification
	return nil
}

func (w fieldDeviceSpecificationWriter) createBulkSpecification(ctx context.Context, fieldDevice *domainFacility.FieldDevice, patch *domainFacility.SpecificationPatch) error {
	if !patch.HasNonNilValues() {
		return nil
	}
	specification := specificationFromPatch(patch)
	fieldDeviceID := fieldDevice.ID
	specification.FieldDeviceID = &fieldDeviceID
	if err := w.writer.service.validateSpecification(specification); err != nil {
		return err
	}
	if err := w.writer.service.specificationRepo.Create(ctx, specification); err != nil {
		return err
	}
	fieldDevice.SpecificationID = &specification.ID
	fieldDevice.Specification = specification
	return nil
}

func (w fieldDeviceSpecificationWriter) createFromPatch(ctx context.Context, fieldDeviceID uuid.UUID, patch *domainFacility.SpecificationPatch) error {
	if !patch.HasNonNilValues() {
		return nil
	}
	specification := specificationFromPatch(patch)
	return w.writer.createSpecification(ctx, fieldDeviceID, specification)
}

func specificationFromPatch(patch *domainFacility.SpecificationPatch) *domainFacility.Specification {
	return &domainFacility.Specification{
		SpecificationSupplier: normalizeOptionalString(patch.SpecificationSupplier),
		SpecificationBrand:    normalizeOptionalString(patch.SpecificationBrand), SpecificationType: normalizeOptionalString(patch.SpecificationType),
		AdditionalInfoMotorValve: normalizeOptionalString(patch.AdditionalInfoMotorValve), AdditionalInfoSize: patch.AdditionalInfoSize,
		AdditionalInformationInstallationLocation: normalizeOptionalString(patch.AdditionalInformationInstallationLocation),
		ElectricalConnectionPH:                    patch.ElectricalConnectionPH, ElectricalConnectionACDC: normalizeOptionalString(patch.ElectricalConnectionACDC),
		ElectricalConnectionAmperage: patch.ElectricalConnectionAmperage, ElectricalConnectionPower: patch.ElectricalConnectionPower,
		ElectricalConnectionRotation: patch.ElectricalConnectionRotation,
	}
}

func normalizeSpecification(specification *domainFacility.Specification) {
	specification.SpecificationSupplier = normalizeOptionalString(specification.SpecificationSupplier)
	specification.SpecificationBrand = normalizeOptionalString(specification.SpecificationBrand)
	specification.SpecificationType = normalizeOptionalString(specification.SpecificationType)
	specification.AdditionalInfoMotorValve = normalizeOptionalString(specification.AdditionalInfoMotorValve)
	specification.AdditionalInformationInstallationLocation = normalizeOptionalString(specification.AdditionalInformationInstallationLocation)
	specification.ElectricalConnectionACDC = normalizeOptionalString(specification.ElectricalConnectionACDC)
}
