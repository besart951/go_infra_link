package facility

import (
	"context"
	"fmt"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/service/changecapture"
	"github.com/google/uuid"
)

const (
	fieldDeviceListDefaultLimit = 300
	fieldDeviceListMaxLimit     = 300
)

type FieldDeviceService struct {
	repo                        domainFacility.FieldDeviceStore
	spsControllerSystemTypeRepo domainFacility.SPSControllerSystemTypeStore
	systemTypeRepo              domainFacility.SystemTypeRepository
	apparatRepo                 domainFacility.ApparatRepository
	systemPartRepo              domainFacility.SystemPartRepository
	specificationRepo           domainFacility.SpecificationStore
	bacnetObjectRepo            domainFacility.BacnetObjectStore
	objectDataRepo              domainFacility.ObjectDataStore
	alarmTypeRepo               domainFacility.AlarmTypeRepository
	bacnetAlarmValueRepo        domainFacility.BacnetObjectAlarmValueRepository
	fieldDeviceOptionsCache     *fieldDeviceOptionsCache
	changeRecorder              changecapture.Recorder
	tx                          txCoordinator
}

func normalizeFieldDeviceListPagination(page, limit int) (int, int) {
	page, limit = domain.NormalizePagination(page, limit, fieldDeviceListDefaultLimit)
	if limit > fieldDeviceListMaxLimit {
		limit = fieldDeviceListMaxLimit
	}
	return page, limit
}

func NewFieldDeviceService(
	repo domainFacility.FieldDeviceStore,
	spsControllerSystemTypeRepo domainFacility.SPSControllerSystemTypeStore,
	systemTypeRepo domainFacility.SystemTypeRepository,
	apparatRepo domainFacility.ApparatRepository,
	systemPartRepo domainFacility.SystemPartRepository,
	specificationRepo domainFacility.SpecificationStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	objectDataRepo domainFacility.ObjectDataStore,
	alarmTypeRepo domainFacility.AlarmTypeRepository,
	bacnetAlarmValueRepo domainFacility.BacnetObjectAlarmValueRepository,
) *FieldDeviceService {
	return &FieldDeviceService{
		repo:                        repo,
		spsControllerSystemTypeRepo: spsControllerSystemTypeRepo,
		systemTypeRepo:              systemTypeRepo,
		apparatRepo:                 apparatRepo,
		systemPartRepo:              systemPartRepo,
		specificationRepo:           specificationRepo,
		bacnetObjectRepo:            bacnetObjectRepo,
		objectDataRepo:              objectDataRepo,
		alarmTypeRepo:               alarmTypeRepo,
		bacnetAlarmValueRepo:        bacnetAlarmValueRepo,
		fieldDeviceOptionsCache:     newFieldDeviceOptionsCache(),
		changeRecorder:              changecapture.NoopRecorder{},
	}
}

func (s *FieldDeviceService) bindTransactions(tx txCoordinator) {
	s.tx = tx
}

func (s *FieldDeviceService) bindChangeRecorder(recorder changecapture.Recorder) {
	s.changeRecorder = changecapture.DefaultRecorder(recorder)
}

func (s *FieldDeviceService) transaction() facilityTx[*FieldDeviceService] {
	return newFacilityTx(s.tx, s, func(services *Services) *FieldDeviceService {
		return services.FieldDevice
	})
}

func (s *FieldDeviceService) writer() fieldDeviceWriter {
	return fieldDeviceWriter{service: s}
}

func (s *FieldDeviceService) recordFieldDeviceChange(ctx context.Context, action changecapture.Action, id uuid.UUID) error {
	return changecapture.DefaultRecorder(s.changeRecorder).Record(ctx, changecapture.Change{
		Action: action,
		Entity: changecapture.EntityRef{
			Domain: "facility",
			Type:   "field_device",
			ID:     id,
		},
	})
}

func (s *FieldDeviceService) Create(ctx context.Context, fieldDevice *domainFacility.FieldDevice) error {
	return s.CreateWithBacnetObjects(ctx, fieldDevice, nil, nil)
}

func (s *FieldDeviceService) CreateWithBacnetObjects(ctx context.Context, fieldDevice *domainFacility.FieldDevice, objectDataID *uuid.UUID, bacnetObjects []domainFacility.BacnetObject) error {
	return s.writer().create(ctx, fieldDevice, fieldDeviceBacnetSelection{
		objectDataID: objectDataID,
		objects:      bacnetObjects,
		objectsSet:   len(bacnetObjects) > 0,
	})
}

func (s *FieldDeviceService) GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.FieldDevice, error) {
	return domain.GetByID(ctx, s.repo, id)
}

func (s *FieldDeviceService) List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	page, limit = normalizeFieldDeviceListPagination(page, limit)
	return s.repo.GetPaginatedList(ctx, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *FieldDeviceService) ListWithFilters(ctx context.Context, params domain.PaginationParams, filters domainFacility.FieldDeviceFilterParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	page, limit := normalizeFieldDeviceListPagination(params.Page, params.Limit)
	params.Page = page
	params.Limit = limit
	return s.repo.GetPaginatedListWithFilters(ctx, params, filters)
}
func (s *FieldDeviceService) Update(ctx context.Context, fieldDevice *domainFacility.FieldDevice) error {
	return s.writer().updateBase(ctx, fieldDevice)
}

func (s *FieldDeviceService) UpdateWithBacnetObjects(ctx context.Context, fieldDevice *domainFacility.FieldDevice, objectDataID *uuid.UUID, bacnetObjects *[]domainFacility.BacnetObject) error {
	selection := fieldDeviceBacnetSelection{objectDataID: objectDataID}
	if bacnetObjects != nil {
		selection.objects = *bacnetObjects
		selection.objectsSet = true
	}
	return s.writer().update(ctx, fieldDevice, selection)
}

func (s *FieldDeviceService) Validate(ctx context.Context, fieldDevice *domainFacility.FieldDevice, excludeID *uuid.UUID) error {
	if err := s.validateRequiredFields(fieldDevice); err != nil {
		return err
	}
	if err := s.ensureParentsExist(ctx, fieldDevice); err != nil {
		return err
	}
	return s.ensureApparatNrAvailable(ctx, fieldDevice, excludeID)
}

func (s *FieldDeviceService) DeleteByID(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteByIds(ctx, []uuid.UUID{id}); err != nil {
		return err
	}
	return s.recordFieldDeviceChange(ctx, changecapture.ActionDeleted, id)
}

func (s *FieldDeviceService) DeleteByIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	if err := s.repo.DeleteByIds(ctx, ids); err != nil {
		return err
	}
	for _, id := range ids {
		if err := s.recordFieldDeviceChange(ctx, changecapture.ActionDeleted, id); err != nil {
			return err
		}
	}
	return nil
}
func (s *FieldDeviceService) CreateSpecification(ctx context.Context, fieldDeviceID uuid.UUID, specification *domainFacility.Specification) error {
	return s.writer().createSpecification(ctx, fieldDeviceID, specification)
}

func (s *FieldDeviceService) UpdateSpecificationPatch(ctx context.Context, fieldDeviceID uuid.UUID, patch *domainFacility.SpecificationPatch) (*domainFacility.Specification, error) {
	return s.writer().updateSpecificationPatch(ctx, fieldDeviceID, patch)
}

func (s *FieldDeviceService) ApplySpecificationPatch(ctx context.Context, fieldDeviceID uuid.UUID, patch *domainFacility.SpecificationPatch) error {
	return s.writer().applySpecificationPatch(ctx, fieldDeviceID, patch)
}

func applySpecificationPatch(spec *domainFacility.Specification, patch *domainFacility.SpecificationPatch) {
	if spec == nil || patch == nil {
		return
	}
	if patch.HasSpecificationSupplier {
		spec.SpecificationSupplier = normalizeOptionalString(patch.SpecificationSupplier)
	}
	if patch.HasSpecificationBrand {
		spec.SpecificationBrand = normalizeOptionalString(patch.SpecificationBrand)
	}
	if patch.HasSpecificationType {
		spec.SpecificationType = normalizeOptionalString(patch.SpecificationType)
	}
	if patch.HasAdditionalInfoMotorValve {
		spec.AdditionalInfoMotorValve = normalizeOptionalString(patch.AdditionalInfoMotorValve)
	}
	if patch.HasAdditionalInfoSize {
		spec.AdditionalInfoSize = patch.AdditionalInfoSize
	}
	if patch.HasAdditionalInformationInstallationLocation {
		spec.AdditionalInformationInstallationLocation = normalizeOptionalString(patch.AdditionalInformationInstallationLocation)
	}
	if patch.HasElectricalConnectionPH {
		spec.ElectricalConnectionPH = patch.ElectricalConnectionPH
	}
	if patch.HasElectricalConnectionACDC {
		spec.ElectricalConnectionACDC = normalizeOptionalString(patch.ElectricalConnectionACDC)
	}
	if patch.HasElectricalConnectionAmperage {
		spec.ElectricalConnectionAmperage = patch.ElectricalConnectionAmperage
	}
	if patch.HasElectricalConnectionPower {
		spec.ElectricalConnectionPower = patch.ElectricalConnectionPower
	}
	if patch.HasElectricalConnectionRotation {
		spec.ElectricalConnectionRotation = patch.ElectricalConnectionRotation
	}
}

func (s *FieldDeviceService) ListBacnetObjects(ctx context.Context, fieldDeviceID uuid.UUID) ([]domainFacility.BacnetObject, error) {
	// Ensure field device exists (and not deleted)
	if _, err := domain.GetByID(ctx, s.repo, fieldDeviceID); err != nil {
		return nil, err
	}

	objs, err := s.bacnetObjectRepo.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	if err != nil {
		return nil, err
	}
	out := make([]domainFacility.BacnetObject, 0, len(objs))
	for _, o := range objs {
		out = append(out, *o)
	}
	return out, nil
}

func (s *FieldDeviceService) ensureParentsExist(ctx context.Context, fieldDevice *domainFacility.FieldDevice) error {
	// With cascaded deletes, we only need to check direct parents.
	// If SPSControllerSystemType exists, all ancestors (SPSController, ControlCabinet, Building) are guaranteed to exist.

	// 1. sps_controller_system_type must exist and not be deleted
	sts, err := domain.GetByID(ctx, s.spsControllerSystemTypeRepo, fieldDevice.SPSControllerSystemTypeID)
	if err != nil {
		return err
	}

	// 2. system_type must exist (business rule: prevent creating instances with deprecated types)
	if err := validateChecks(referenceExists(ctx, s.systemTypeRepo, sts.SystemTypeID)); err != nil {
		return err
	}

	// 3. apparat must exist and not be deleted
	if err := validateChecks(referenceExists(ctx, s.apparatRepo, fieldDevice.ApparatID)); err != nil {
		return err
	}

	// 4. system_part must exist and not be deleted
	if err := validateChecks(referenceExists(ctx, s.systemPartRepo, fieldDevice.SystemPartID)); err != nil {
		return err
	}

	return nil
}

func (s *FieldDeviceService) ensureApparatNrAvailable(ctx context.Context, fieldDevice *domainFacility.FieldDevice, excludeID *uuid.UUID) error {
	var excludeIDs []uuid.UUID
	if excludeID != nil {
		excludeIDs = []uuid.UUID{*excludeID}
	}
	return s.ensureApparatNrAvailableWithExclusions(ctx, fieldDevice, excludeIDs)
}

func (s *FieldDeviceService) ensureApparatNrAvailableWithExclusions(ctx context.Context, fieldDevice *domainFacility.FieldDevice, excludeIDs []uuid.UUID) error {
	if err := validateRules(
		requiredNonZero(fieldDeviceApparatNrField, fieldDevice.ApparatNr),
		func(builder *domain.ValidationBuilder) {
			fieldDeviceApparatNrField.Between(builder, fieldDevice.ApparatNr, 1, 99)
		},
	); err != nil {
		return err
	}

	return validateChecks(
		func(builder *domain.ValidationBuilder) error {
			exists, err := s.repo.ExistsApparatNrConflict(ctx,
				fieldDevice.SPSControllerSystemTypeID,
				fieldDevice.SystemPartID,
				fieldDevice.ApparatID,
				fieldDevice.ApparatNr,
				excludeIDs,
			)
			if err != nil {
				return err
			}
			if exists {
				fieldDeviceApparatNrField.Add(builder, "apparatnummer ist bereits vergeben")
			}
			return nil
		},
	)
}

func (s *FieldDeviceService) ListAvailableApparatNumbers(ctx context.Context, spsControllerSystemTypeID uuid.UUID, systemPartID uuid.UUID, apparatID uuid.UUID) ([]int, error) {
	used, err := s.repo.GetUsedApparatNumbers(ctx, spsControllerSystemTypeID, systemPartID, apparatID)
	if err != nil {
		return nil, err
	}

	usedSet := make(map[int]struct{}, len(used))
	for _, n := range used {
		if n >= 1 && n <= 99 {
			usedSet[n] = struct{}{}
		}
	}

	available := make([]int, 0, 99-len(usedSet))
	for n := 1; n <= 99; n++ {
		if _, ok := usedSet[n]; !ok {
			available = append(available, n)
		}
	}
	return available, nil
}

// GetFieldDeviceOptions returns all metadata needed for creating/editing field devices.
// This implements the "Single-Fetch Metadata Strategy" to avoid multiple API calls.
func (s *FieldDeviceService) GetFieldDeviceOptions(ctx context.Context) (*domainFacility.FieldDeviceOptions, error) {
	return s.getFieldDeviceOptions(ctx, nil, func(ctx context.Context) ([]*domainFacility.ObjectData, error) {
		return s.objectDataRepo.GetTemplatesLite(ctx)
	})
}

// GetFieldDeviceOptionsForProject returns all metadata needed for creating/editing field devices within a project.
// This fetches object data that belongs to the specified project (project_id = projectID AND is_active = true).
func (s *FieldDeviceService) GetFieldDeviceOptionsForProject(ctx context.Context, projectID uuid.UUID) (*domainFacility.FieldDeviceOptions, error) {
	return s.getFieldDeviceOptions(ctx, &projectID, func(ctx context.Context) ([]*domainFacility.ObjectData, error) {
		return s.objectDataRepo.GetForProjectLite(ctx, projectID)
	})
}

func (s *FieldDeviceService) getFieldDeviceOptions(ctx context.Context, projectID *uuid.UUID, loadObjectDatas func(context.Context) ([]*domainFacility.ObjectData, error)) (*domainFacility.FieldDeviceOptions, error) {
	cacheKey := "templates"
	if projectID != nil {
		cacheKey = "project:" + projectID.String()
	}

	revision := ""
	if revisioner, ok := s.objectDataRepo.(fieldDeviceOptionsRevisioner); ok {
		var err error
		revision, err = revisioner.GetFieldDeviceOptionsRevision(ctx, projectID)
		if err != nil {
			return nil, err
		}
		if options, ok := s.fieldDeviceOptionsCache.get(cacheKey, revision); ok {
			return options, nil
		}
	}

	objectDatas, err := loadObjectDatas(ctx)
	if err != nil {
		return nil, err
	}

	options, err := s.buildFieldDeviceOptions(ctx, objectDatas)
	if err != nil {
		return nil, err
	}
	s.fieldDeviceOptionsCache.set(cacheKey, revision, options)
	return options, nil
}

func (s *FieldDeviceService) buildFieldDeviceOptions(ctx context.Context, objectDatas []*domainFacility.ObjectData) (*domainFacility.FieldDeviceOptions, error) {
	apparatIDs := collectUniqueApparatIDs(objectDatas)

	fullApparats := make([]*domainFacility.Apparat, 0, len(apparatIDs))
	if len(apparatIDs) > 0 {
		var err error
		fullApparats, err = s.apparatRepo.GetByIds(ctx, apparatIDs)
		if err != nil {
			return nil, err
		}
	}

	apparatToSystemPart := make(map[uuid.UUID][]uuid.UUID, len(fullApparats))
	apparats := make([]domainFacility.Apparat, 0, len(fullApparats))
	systemParts := make([]domainFacility.SystemPart, 0)
	seenSystemParts := make(map[uuid.UUID]struct{})

	for _, app := range fullApparats {
		if app == nil {
			continue
		}

		apparats = append(apparats, *app)
		apparatToSystemPart[app.ID] = extractIDs(app.SystemParts, func(sp *domainFacility.SystemPart) uuid.UUID { return sp.ID })

		for _, sp := range app.SystemParts {
			if sp == nil {
				continue
			}
			if _, exists := seenSystemParts[sp.ID]; exists {
				continue
			}
			seenSystemParts[sp.ID] = struct{}{}
			systemParts = append(systemParts, *sp)
		}
	}

	objectDataToApparat := make(map[uuid.UUID][]uuid.UUID, len(objectDatas))
	objectDataValues := make([]domainFacility.ObjectData, 0, len(objectDatas))
	for _, od := range objectDatas {
		if od == nil {
			continue
		}

		objectDataToApparat[od.ID] = extractIDs(od.Apparats, func(a *domainFacility.Apparat) uuid.UUID { return a.ID })
		objectDataValues = append(objectDataValues, *od)
	}

	return &domainFacility.FieldDeviceOptions{
		Apparats:            apparats,
		SystemParts:         systemParts,
		ObjectDatas:         objectDataValues,
		ApparatToSystemPart: apparatToSystemPart,
		ObjectDataToApparat: objectDataToApparat,
	}, nil
}

func collectUniqueApparatIDs(objectDatas []*domainFacility.ObjectData) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{})
	ids := make([]uuid.UUID, 0)
	for _, od := range objectDatas {
		if od == nil {
			continue
		}
		for _, app := range od.Apparats {
			if app == nil {
				continue
			}
			if _, exists := seen[app.ID]; exists {
				continue
			}
			seen[app.ID] = struct{}{}
			ids = append(ids, app.ID)
		}
	}
	return ids
}

func (s *FieldDeviceService) validateRequiredFields(fieldDevice *domainFacility.FieldDevice) error {
	return validateRules(
		requiredUUID(fieldDeviceSystemTypeIDField, fieldDevice.SPSControllerSystemTypeID),
		requiredUUID(fieldDeviceApparatIDField, fieldDevice.ApparatID),
		requiredUUID(fieldDeviceSystemPartIDField, fieldDevice.SystemPartID),
		optionalMaxLength(fieldDeviceBMKField, fieldDevice.BMK, 10),
		optionalMaxLength(fieldDeviceDescriptionField, fieldDevice.Description, 250),
		optionalMaxLength(fieldDeviceTextFixField, fieldDevice.TextIndividuell, 250),
	)
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	if *value == "" {
		return nil
	}
	return value
}

func (s *FieldDeviceService) validateSpecification(spec *domainFacility.Specification) error {
	if spec == nil {
		return nil
	}

	return validateRules(
		optionalMaxLength(specificationSupplierField, spec.SpecificationSupplier, 250),
		optionalMaxLength(specificationBrandField, spec.SpecificationBrand, 250),
		optionalMaxLength(specificationTypeField, spec.SpecificationType, 250),
		optionalMaxLength(specificationMotorValveInfoField, spec.AdditionalInfoMotorValve, 250),
		optionalMaxLength(specificationInstallLocationField, spec.AdditionalInformationInstallationLocation, 250),
		optionalExactLength(specificationElectricalACDCField, spec.ElectricalConnectionACDC, 2),
	)
}

func (s *FieldDeviceService) replaceBacnetObjects(ctx context.Context, fieldDeviceID uuid.UUID, bacnetObjects []domainFacility.BacnetObject) error {
	return s.projectFacilityCopy().replaceFieldDeviceBacnetObjects(ctx, fieldDeviceID, bacnetObjects)
}

func applyBacnetObjectPatch(target *domainFacility.BacnetObject, patch domainFacility.BacnetObjectPatch) {
	if patch.TextFix != nil {
		target.TextFix = normalizeBacnetTextFix(*patch.TextFix)
	}
	if patch.Description != nil {
		target.Description = patch.Description
	}
	if patch.GMSVisible != nil {
		target.GMSVisible = *patch.GMSVisible
	}
	if patch.Optional != nil {
		target.Optional = *patch.Optional
	}
	if patch.TextIndividual != nil {
		target.TextIndividual = patch.TextIndividual
	}
	if patch.SoftwareType != nil {
		target.SoftwareType = *patch.SoftwareType
	}
	if patch.SoftwareNumber != nil {
		target.SoftwareNumber = *patch.SoftwareNumber
	}
	if patch.HardwareType != nil {
		target.HardwareType = *patch.HardwareType
	}
	if patch.HardwareQuantity != nil {
		target.HardwareQuantity = *patch.HardwareQuantity
	}
	if patch.SoftwareReferenceID != nil {
		target.SoftwareReferenceID = patch.SoftwareReferenceID
	}
	if patch.StateTextID != nil {
		target.StateTextID = patch.StateTextID
	}
	if patch.NotificationClassID != nil {
		target.NotificationClassID = patch.NotificationClassID
	}
	if patch.AlarmTypeID != nil {
		target.AlarmTypeID = patch.AlarmTypeID
	}
}

func (s *FieldDeviceService) patchBacnetObjects(ctx context.Context, fieldDeviceID uuid.UUID, patches []domainFacility.BacnetObjectPatch) error {
	if len(patches) == 0 {
		return nil
	}

	existingItems, err := s.bacnetObjectRepo.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	if err != nil {
		return err
	}

	existingMap := make(map[uuid.UUID]*domainFacility.BacnetObject, len(existingItems))
	for _, item := range existingItems {
		existingMap[item.ID] = item
	}

	updatedMap := make(map[uuid.UUID]*domainFacility.BacnetObject, len(patches))
	patchedIDs := make(map[uuid.UUID]struct{}, len(patches))
	ve := domain.NewValidationError()

	for _, patch := range patches {
		patchedIDs[patch.ID] = struct{}{}
		existing, ok := existingMap[patch.ID]
		if !ok {
			ve = ve.Add(fmt.Sprintf("bacnet_objects.%s.id", patch.ID), "bacnet object not found for field device")
			continue
		}

		clone := *existing
		applyBacnetObjectPatch(&clone, patch)

		if strings.TrimSpace(clone.TextFix) == "" {
			ve = ve.Add(fmt.Sprintf("bacnet_objects.%s.text_fix", patch.ID), "text_fix is required")
		}
		if strings.TrimSpace(string(clone.SoftwareType)) == "" {
			ve = ve.Add(fmt.Sprintf("bacnet_objects.%s.software_type", patch.ID), "software_type is required")
		}

		updatedMap[patch.ID] = &clone
	}

	if len(ve.Fields) > 0 {
		return ve
	}

	for _, updated := range updatedMap {
		if err := s.bacnetObjectRepo.Update(ctx, updated); err != nil {
			return err
		}
	}

	return nil
}

func (s *FieldDeviceService) replaceBacnetObjectsFromObjectData(ctx context.Context, fieldDeviceID uuid.UUID, objectDataID uuid.UUID) error {
	return s.projectFacilityCopy().replaceFieldDeviceBacnetObjectsFromObjectData(ctx, fieldDeviceID, objectDataID)
}

// MultiCreate creates multiple field devices in a single operation.
// It validates each device independently and continues on failures.
// Returns detailed results for each device creation attempt.
func (s *FieldDeviceService) MultiCreate(ctx context.Context, items []domainFacility.FieldDeviceCreateItem) *domainFacility.FieldDeviceMultiCreateResult {
	return s.writer().multiCreate(ctx, items)
}

// BulkUpdate updates multiple field devices in a single operation.
// It processes updates ensuring that swaps/permutations within the batch are handled correctly.
// Uniqueness constraints are checked against the database (excluding batch items) AND internally within the batch.
func (s *FieldDeviceService) BulkUpdate(ctx context.Context, updates []domainFacility.BulkFieldDeviceUpdate) *domainFacility.BulkOperationResult {
	return s.writer().bulkUpdate(ctx, updates)
}

// BulkDelete deletes multiple field devices in a single operation.
// It processes each deletion independently and returns detailed results.
func (s *FieldDeviceService) BulkDelete(ctx context.Context, ids []uuid.UUID) *domainFacility.BulkOperationResult {
	result := &domainFacility.BulkOperationResult{
		Results:      make([]domainFacility.BulkOperationResultItem, len(ids)),
		TotalCount:   len(ids),
		SuccessCount: 0,
		FailureCount: 0,
	}

	for i, id := range ids {
		resultItem := &result.Results[i]
		resultItem.ID = id
		resultItem.Success = false

		if err := s.DeleteByID(ctx, id); err != nil {
			resultItem.Error = err.Error()
			result.FailureCount++
			continue
		}

		resultItem.Success = true
		result.SuccessCount++
	}

	return result
}
