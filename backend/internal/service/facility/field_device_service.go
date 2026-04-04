package facility

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
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
	}
}

func (s *FieldDeviceService) Create(ctx context.Context, fieldDevice *domainFacility.FieldDevice) error {
	return s.CreateWithBacnetObjects(ctx, fieldDevice, nil, nil)
}

func (s *FieldDeviceService) CreateWithBacnetObjects(ctx context.Context, fieldDevice *domainFacility.FieldDevice, objectDataID *uuid.UUID, bacnetObjects []domainFacility.BacnetObject) error {
	if objectDataID != nil && len(bacnetObjects) > 0 {
		return domain.ErrInvalidArgument
	}
	if err := s.validateRequiredFields(fieldDevice); err != nil {
		return err
	}
	if err := s.ensureParentsExist(ctx, fieldDevice); err != nil {
		return err
	}
	if err := s.ensureApparatNrAvailable(ctx, fieldDevice, nil); err != nil {
		return err
	}

	if err := s.repo.Create(ctx, fieldDevice); err != nil {
		return err
	}

	if objectDataID != nil {
		if err := s.replaceBacnetObjectsFromObjectData(ctx, fieldDevice.ID, *objectDataID); err != nil {
			_ = s.DeleteByID(ctx, fieldDevice.ID)
			return err
		}
		return nil
	}
	if len(bacnetObjects) > 0 {
		if err := s.replaceBacnetObjects(ctx, fieldDevice.ID, bacnetObjects); err != nil {
			_ = s.DeleteByID(ctx, fieldDevice.ID)
			return err
		}
	}
	return nil
}

func (s *FieldDeviceService) GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.FieldDevice, error) {
	return domain.GetByID(ctx, s.repo, id)
}

func (s *FieldDeviceService) List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	page, limit = domain.NormalizePagination(page, limit, 300)
	return s.repo.GetPaginatedList(ctx, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *FieldDeviceService) ListWithFilters(ctx context.Context, params domain.PaginationParams, filters domainFacility.FieldDeviceFilterParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 300)
	params.Page = page
	params.Limit = limit
	return s.repo.GetPaginatedListWithFilters(ctx, params, filters)
}
func (s *FieldDeviceService) Update(ctx context.Context, fieldDevice *domainFacility.FieldDevice) error {
	if err := s.validateRequiredFields(fieldDevice); err != nil {
		return err
	}
	if err := s.ensureParentsExist(ctx, fieldDevice); err != nil {
		return err
	}
	if err := s.ensureApparatNrAvailable(ctx, fieldDevice, &fieldDevice.ID); err != nil {
		return err
	}
	return s.repo.Update(ctx, fieldDevice)
}

func (s *FieldDeviceService) UpdateWithBacnetObjects(ctx context.Context, fieldDevice *domainFacility.FieldDevice, objectDataID *uuid.UUID, bacnetObjects *[]domainFacility.BacnetObject) error {
	if objectDataID != nil && bacnetObjects != nil {
		return domain.ErrInvalidArgument
	}
	if err := s.validateRequiredFields(fieldDevice); err != nil {
		return err
	}
	if err := s.ensureParentsExist(ctx, fieldDevice); err != nil {
		return err
	}
	if err := s.ensureApparatNrAvailable(ctx, fieldDevice, &fieldDevice.ID); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, fieldDevice); err != nil {
		return err
	}

	if objectDataID != nil {
		return s.replaceBacnetObjectsFromObjectData(ctx, fieldDevice.ID, *objectDataID)
	}
	if bacnetObjects != nil {
		return s.replaceBacnetObjects(ctx, fieldDevice.ID, *bacnetObjects)
	}

	return nil
}

func (s *FieldDeviceService) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteByIds(ctx, []uuid.UUID{id})
}

func (s *FieldDeviceService) DeleteByIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return s.repo.DeleteByIds(ctx, ids)
}
func (s *FieldDeviceService) CreateSpecification(ctx context.Context, fieldDeviceID uuid.UUID, specification *domainFacility.Specification) error {
	// Ensure field device exists (and not deleted)
	fieldDevice, err := domain.GetByID(ctx, s.repo, fieldDeviceID)
	if err != nil {
		return err
	}

	// Ensure 1:1 uniqueness
	existing, err := s.specificationRepo.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return domain.ErrConflict
	}

	specification.SpecificationSupplier = normalizeOptionalString(specification.SpecificationSupplier)
	specification.SpecificationBrand = normalizeOptionalString(specification.SpecificationBrand)
	specification.SpecificationType = normalizeOptionalString(specification.SpecificationType)
	specification.AdditionalInfoMotorValve = normalizeOptionalString(specification.AdditionalInfoMotorValve)
	specification.AdditionalInformationInstallationLocation = normalizeOptionalString(specification.AdditionalInformationInstallationLocation)
	specification.ElectricalConnectionACDC = normalizeOptionalString(specification.ElectricalConnectionACDC)
	if err := s.validateSpecification(specification); err != nil {
		return err
	}

	id := fieldDeviceID
	specification.FieldDeviceID = &id
	if err := s.specificationRepo.Create(ctx, specification); err != nil {
		return err
	}

	// Update the bidirectional relationship: set field_device.specification_id
	fieldDevice.SpecificationID = &specification.ID
	return s.repo.Update(ctx, fieldDevice)
}

func (s *FieldDeviceService) UpdateSpecification(ctx context.Context, fieldDeviceID uuid.UUID, patch *domainFacility.Specification) (*domainFacility.Specification, error) {
	// Ensure field device exists (and not deleted)
	if _, err := domain.GetByID(ctx, s.repo, fieldDeviceID); err != nil {
		return nil, err
	}

	specs, err := s.specificationRepo.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	if err != nil {
		return nil, err
	}
	if len(specs) == 0 {
		return nil, domain.ErrNotFound
	}
	spec := specs[0]

	if patch.SpecificationSupplier != nil {
		spec.SpecificationSupplier = normalizeOptionalString(patch.SpecificationSupplier)
	}
	if patch.SpecificationBrand != nil {
		spec.SpecificationBrand = normalizeOptionalString(patch.SpecificationBrand)
	}
	if patch.SpecificationType != nil {
		spec.SpecificationType = normalizeOptionalString(patch.SpecificationType)
	}
	if patch.AdditionalInfoMotorValve != nil {
		spec.AdditionalInfoMotorValve = normalizeOptionalString(patch.AdditionalInfoMotorValve)
	}
	if patch.AdditionalInfoSize != nil {
		spec.AdditionalInfoSize = patch.AdditionalInfoSize
	}
	if patch.AdditionalInformationInstallationLocation != nil {
		spec.AdditionalInformationInstallationLocation = normalizeOptionalString(patch.AdditionalInformationInstallationLocation)
	}
	if patch.ElectricalConnectionPH != nil {
		spec.ElectricalConnectionPH = patch.ElectricalConnectionPH
	}
	if patch.ElectricalConnectionACDC != nil {
		spec.ElectricalConnectionACDC = normalizeOptionalString(patch.ElectricalConnectionACDC)
	}
	if patch.ElectricalConnectionAmperage != nil {
		spec.ElectricalConnectionAmperage = patch.ElectricalConnectionAmperage
	}
	if patch.ElectricalConnectionPower != nil {
		spec.ElectricalConnectionPower = patch.ElectricalConnectionPower
	}
	if patch.ElectricalConnectionRotation != nil {
		spec.ElectricalConnectionRotation = patch.ElectricalConnectionRotation
	}
	if err := s.validateSpecification(spec); err != nil {
		return nil, err
	}

	if err := s.specificationRepo.Update(ctx, spec); err != nil {
		return nil, err
	}
	return spec, nil
}

func (s *FieldDeviceService) ApplySpecificationPatch(ctx context.Context, fieldDeviceID uuid.UUID, patch *domainFacility.SpecificationPatch) error {
	if patch == nil || !patch.HasChanges() {
		return nil
	}

	if _, err := domain.GetByID(ctx, s.repo, fieldDeviceID); err != nil {
		return err
	}

	specs, err := s.specificationRepo.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	if err != nil {
		return err
	}

	if len(specs) == 0 {
		if !patch.HasNonNilValues() {
			return nil
		}

		spec := &domainFacility.Specification{
			SpecificationSupplier:                     normalizeOptionalString(patch.SpecificationSupplier),
			SpecificationBrand:                        normalizeOptionalString(patch.SpecificationBrand),
			SpecificationType:                         normalizeOptionalString(patch.SpecificationType),
			AdditionalInfoMotorValve:                  normalizeOptionalString(patch.AdditionalInfoMotorValve),
			AdditionalInfoSize:                        patch.AdditionalInfoSize,
			AdditionalInformationInstallationLocation: normalizeOptionalString(patch.AdditionalInformationInstallationLocation),
			ElectricalConnectionPH:                    patch.ElectricalConnectionPH,
			ElectricalConnectionACDC:                  normalizeOptionalString(patch.ElectricalConnectionACDC),
			ElectricalConnectionAmperage:              patch.ElectricalConnectionAmperage,
			ElectricalConnectionPower:                 patch.ElectricalConnectionPower,
			ElectricalConnectionRotation:              patch.ElectricalConnectionRotation,
		}
		return s.CreateSpecification(ctx, fieldDeviceID, spec)
	}

	spec := specs[0]
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

	if err := s.validateSpecification(spec); err != nil {
		return err
	}
	return s.specificationRepo.Update(ctx, spec)
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
	if _, err := domain.GetByID(ctx, s.systemTypeRepo, sts.SystemTypeID); err != nil {
		return err
	}

	// 3. apparat must exist and not be deleted
	if _, err := domain.GetByID(ctx, s.apparatRepo, fieldDevice.ApparatID); err != nil {
		return err
	}

	// 4. system_part (optional) must exist if provided
	if fieldDevice.SystemPartID != uuid.Nil {
		if _, err := domain.GetByID(ctx, s.systemPartRepo, fieldDevice.SystemPartID); err != nil {
			return err
		}
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
	if fieldDevice.ApparatNr == 0 {
		return domain.NewValidationError().Add("fielddevice.apparat_nr", "apparat_nr is required")
	}
	if fieldDevice.ApparatNr < 1 || fieldDevice.ApparatNr > 99 {
		return domain.NewValidationError().Add("fielddevice.apparat_nr", "apparat_nr must be between 1 and 99")
	}

	var systemPartID *uuid.UUID
	if fieldDevice.SystemPartID != uuid.Nil {
		systemPartID = &fieldDevice.SystemPartID
	}

	exists, err := s.repo.ExistsApparatNrConflict(ctx,
		fieldDevice.SPSControllerSystemTypeID,
		systemPartID,
		fieldDevice.ApparatID,
		fieldDevice.ApparatNr,
		excludeIDs,
	)
	if err != nil {
		return err
	}
	if exists {
		return domain.NewValidationError().Add("fielddevice.apparat_nr", "apparatnummer ist bereits vergeben")
	}
	return nil
}

func (s *FieldDeviceService) ListAvailableApparatNumbers(ctx context.Context, spsControllerSystemTypeID uuid.UUID, systemPartID *uuid.UUID, apparatID uuid.UUID) ([]int, error) {
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
	// Fetch all active object datas (templates) with their apparats only
	objectDatas, err := s.objectDataRepo.GetTemplatesLite(ctx)
	if err != nil {
		return nil, err
	}

	return s.buildFieldDeviceOptions(ctx, objectDatas)
}

// GetFieldDeviceOptionsForProject returns all metadata needed for creating/editing field devices within a project.
// This fetches object data that belongs to the specified project (project_id = projectID AND is_active = true).
func (s *FieldDeviceService) GetFieldDeviceOptionsForProject(ctx context.Context, projectID uuid.UUID) (*domainFacility.FieldDeviceOptions, error) {
	// Fetch object datas for the project with their apparats only
	objectDatas, err := s.objectDataRepo.GetForProjectLite(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return s.buildFieldDeviceOptions(ctx, objectDatas)
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
	ve := domain.NewValidationError()
	if fieldDevice.SPSControllerSystemTypeID == uuid.Nil {
		ve = ve.Add("fielddevice.sps_controller_system_type_id", "sps_controller_system_type_id is required")
	}
	if fieldDevice.ApparatID == uuid.Nil {
		ve = ve.Add("fielddevice.apparat_id", "apparat_id is required")
	}
	if fieldDevice.SystemPartID == uuid.Nil {
		ve = ve.Add("fielddevice.system_part_id", "system_part_id is required")
	}
	if fieldDevice.BMK != nil && len(*fieldDevice.BMK) > 10 {
		ve = ve.Add("fielddevice.bmk", "bmk must be at most 10 characters")
	}
	if fieldDevice.Description != nil && len(*fieldDevice.Description) > 250 {
		ve = ve.Add("fielddevice.description", "description must be at most 250 characters")
	}
	if fieldDevice.TextIndividuell != nil && len(*fieldDevice.TextIndividuell) > 250 {
		ve = ve.Add("fielddevice.text_fix", "text_fix must be at most 250 characters")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
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

	ve := domain.NewValidationError()
	checkMax := func(field string, value *string) {
		if value != nil && len(*value) > 250 {
			ve = ve.Add(field, "must be at most 250 characters")
		}
	}

	checkMax("specification.specification_supplier", spec.SpecificationSupplier)
	checkMax("specification.specification_brand", spec.SpecificationBrand)
	checkMax("specification.specification_type", spec.SpecificationType)
	checkMax("specification.additional_info_motor_valve", spec.AdditionalInfoMotorValve)
	checkMax(
		"specification.additional_information_installation_location",
		spec.AdditionalInformationInstallationLocation,
	)

	if spec.ElectricalConnectionACDC != nil && *spec.ElectricalConnectionACDC != "" && len(*spec.ElectricalConnectionACDC) != 2 {
		ve = ve.Add("specification.electrical_connection_acdc", "electrical_connection_acdc must be exactly 2 characters")
	}

	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func (s *FieldDeviceService) buildAlarmValuesForBacnetObjects(ctx context.Context, bacnetObjects []*domainFacility.BacnetObject) ([]*domainFacility.BacnetObjectAlarmValue, error) {
	if len(bacnetObjects) == 0 {
		return nil, nil
	}

	alarmTypeCache := make(map[uuid.UUID]*domainFacility.AlarmType)
	values := make([]*domainFacility.BacnetObjectAlarmValue, 0)

	for _, obj := range bacnetObjects {
		if obj == nil || obj.AlarmTypeID == nil {
			continue
		}

		alarmType, ok := alarmTypeCache[*obj.AlarmTypeID]
		if !ok {
			loaded, err := s.alarmTypeRepo.GetWithFields(ctx, *obj.AlarmTypeID)
			if err != nil {
				return nil, err
			}
			if loaded == nil {
				return nil, domain.ErrNotFound
			}
			alarmType = loaded
			alarmTypeCache[*obj.AlarmTypeID] = loaded
		}

		for _, field := range alarmType.Fields {
			value := &domainFacility.BacnetObjectAlarmValue{
				BacnetObjectID:   obj.ID,
				AlarmTypeFieldID: field.ID,
				UnitID:           field.DefaultUnitID,
				Source:           domainFacility.AlarmValueSourceDefault,
			}

			if field.DefaultValueJSON != nil && field.AlarmField != nil {
				applyAlarmDefaultValue(value, field.AlarmField.DataType, *field.DefaultValueJSON)
			}

			values = append(values, value)
		}
	}

	return values, nil
}

func (s *FieldDeviceService) createAlarmValuesForBacnetObjects(ctx context.Context, bacnetObjects []*domainFacility.BacnetObject) error {
	values, err := s.buildAlarmValuesForBacnetObjects(ctx, bacnetObjects)
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return nil
	}

	return s.bacnetAlarmValueRepo.BulkCreate(ctx, values, 500)
}

func applyAlarmDefaultValue(value *domainFacility.BacnetObjectAlarmValue, dataType string, defaultValueJSON string) {
	if value == nil {
		return
	}

	var decoded any
	if err := json.Unmarshal([]byte(defaultValueJSON), &decoded); err != nil {
		value.ValueString = &defaultValueJSON
		return
	}

	switch strings.ToLower(strings.TrimSpace(dataType)) {
	case "number", "duration":
		if n, ok := toFloat64(decoded); ok {
			value.ValueNumber = &n
		}
	case "integer":
		if n, ok := toInt64(decoded); ok {
			value.ValueInteger = &n
		}
	case "boolean":
		if b, ok := decoded.(bool); ok {
			value.ValueBoolean = &b
		}
	case "string", "enum":
		if s, ok := decoded.(string); ok {
			value.ValueString = &s
		}
	case "state_map", "json":
		if b, err := json.Marshal(decoded); err == nil {
			raw := string(b)
			value.ValueJSON = &raw
		}
	default:
		if b, err := json.Marshal(decoded); err == nil {
			raw := string(b)
			value.ValueJSON = &raw
		}
	}
}

func toFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

func toInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	default:
		return 0, false
	}
}

func (s *FieldDeviceService) replaceBacnetObjects(ctx context.Context, fieldDeviceID uuid.UUID, bacnetObjects []domainFacility.BacnetObject) error {
	ve := domain.NewValidationError()
	for i, obj := range bacnetObjects {
		obj.TextFix = normalizeBacnetTextFix(obj.TextFix)
		bacnetObjects[i].TextFix = obj.TextFix
		if obj.TextFix == "" {
			ve = ve.Add(fmt.Sprintf("bacnet_objects.%d.text_fix", i), "text_fix is required")
			continue
		}
	}
	if len(ve.Fields) > 0 {
		return ve
	}

	if err := s.bacnetObjectRepo.DeleteByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID}); err != nil {
		return err
	}
	if len(bacnetObjects) == 0 {
		return nil
	}

	objects := make([]*domainFacility.BacnetObject, 0, len(bacnetObjects))
	for i := range bacnetObjects {
		id := fieldDeviceID
		bacnetObjects[i].FieldDeviceID = &id
		objects = append(objects, &bacnetObjects[i])
	}

	if err := s.bacnetObjectRepo.BulkCreate(ctx, objects, 200); err != nil {
		return err
	}

	return s.createAlarmValuesForBacnetObjects(ctx, objects)
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
	od, err := domain.GetByID(ctx, s.objectDataRepo, objectDataID)
	if err != nil {
		return err
	}
	if !od.IsActive {
		return domain.ErrNotFound
	}

	ids, err := s.objectDataRepo.GetBacnetObjectIDs(ctx, objectDataID)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return s.bacnetObjectRepo.DeleteByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	}

	templates, err := s.bacnetObjectRepo.GetByIds(ctx, ids)
	if err != nil {
		return err
	}
	if len(templates) != len(ids) {
		return domain.ErrNotFound
	}

	templateToClone := make(map[uuid.UUID]*domainFacility.BacnetObject, len(templates))
	templateRef := make(map[uuid.UUID]*uuid.UUID, len(templates))
	clones := make([]*domainFacility.BacnetObject, 0, len(templates))

	for _, t := range templates {
		textFix := normalizeBacnetTextFix(t.TextFix)
		if textFix == "" {
			return domain.NewValidationError().Add("bacnet_objects.text_fix", "text_fix is required")
		}

		clone := &domainFacility.BacnetObject{
			TextFix:             textFix,
			Description:         t.Description,
			GMSVisible:          t.GMSVisible,
			Optional:            t.Optional,
			TextIndividual:      t.TextIndividual,
			SoftwareType:        t.SoftwareType,
			SoftwareNumber:      t.SoftwareNumber,
			HardwareType:        t.HardwareType,
			HardwareQuantity:    t.HardwareQuantity,
			FieldDeviceID:       &fieldDeviceID,
			SoftwareReferenceID: nil,
			StateTextID:         t.StateTextID,
			NotificationClassID: t.NotificationClassID,
			AlarmTypeID:         t.AlarmTypeID,
		}
		templateToClone[t.ID] = clone
		templateRef[t.ID] = t.SoftwareReferenceID
		clones = append(clones, clone)
	}

	if err := s.bacnetObjectRepo.DeleteByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID}); err != nil {
		return err
	}

	// First pass: create clones without software references.
	if err := s.bacnetObjectRepo.BulkCreate(ctx, clones, 200); err != nil {
		return err
	}
	if err := s.createAlarmValuesForBacnetObjects(ctx, clones); err != nil {
		return err
	}

	oldToNew := make(map[uuid.UUID]uuid.UUID, len(templates))
	for tid, clone := range templateToClone {
		oldToNew[tid] = clone.ID
	}

	// Second pass: update internal software references.
	for tid, ref := range templateRef {
		if ref == nil {
			continue
		}
		mapped, ok := oldToNew[*ref]
		if !ok {
			continue
		}
		clone := templateToClone[tid]
		clone.SoftwareReferenceID = &mapped
		if err := s.bacnetObjectRepo.Update(ctx, clone); err != nil {
			return err
		}
	}

	return nil
}

// MultiCreate creates multiple field devices in a single operation.
// It validates each device independently and continues on failures.
// Returns detailed results for each device creation attempt.
func (s *FieldDeviceService) MultiCreate(ctx context.Context, items []domainFacility.FieldDeviceCreateItem) *domainFacility.FieldDeviceMultiCreateResult {
	result := &domainFacility.FieldDeviceMultiCreateResult{
		Results:       make([]domainFacility.FieldDeviceCreateResult, len(items)),
		TotalRequests: len(items),
		SuccessCount:  0,
		FailureCount:  0,
	}

	// Cache lookups to reduce DB round-trips during multi-create.
	stsCache := make(map[uuid.UUID]*domainFacility.SPSControllerSystemType)
	systemTypeCache := make(map[uuid.UUID]bool)
	apparatCache := make(map[uuid.UUID]bool)
	systemPartCache := make(map[uuid.UUID]bool)
	usedApparatNumbersCache := make(map[string]map[int]struct{})

	getApparatNumbersKey := func(spsControllerSystemTypeID uuid.UUID, systemPartID *uuid.UUID, apparatID uuid.UUID) string {
		key := spsControllerSystemTypeID.String() + "|" + apparatID.String() + "|"
		if systemPartID != nil {
			key += systemPartID.String()
		} else {
			key += "-"
		}
		return key
	}

	ensureParentsExistCached := func(fieldDevice *domainFacility.FieldDevice) error {
		sts, ok := stsCache[fieldDevice.SPSControllerSystemTypeID]
		if !ok {
			loaded, err := domain.GetByID(ctx, s.spsControllerSystemTypeRepo, fieldDevice.SPSControllerSystemTypeID)
			if err != nil {
				stsCache[fieldDevice.SPSControllerSystemTypeID] = nil
				return err
			}
			sts = loaded
			stsCache[fieldDevice.SPSControllerSystemTypeID] = loaded
		}
		if sts == nil {
			return domain.ErrNotFound
		}

		if _, ok := systemTypeCache[sts.SystemTypeID]; !ok {
			if _, err := domain.GetByID(ctx, s.systemTypeRepo, sts.SystemTypeID); err != nil {
				return err
			}
			systemTypeCache[sts.SystemTypeID] = true
		}

		if _, ok := apparatCache[fieldDevice.ApparatID]; !ok {
			if _, err := domain.GetByID(ctx, s.apparatRepo, fieldDevice.ApparatID); err != nil {
				return err
			}
			apparatCache[fieldDevice.ApparatID] = true
		}

		if fieldDevice.SystemPartID != uuid.Nil {
			if _, ok := systemPartCache[fieldDevice.SystemPartID]; !ok {
				if _, err := domain.GetByID(ctx, s.systemPartRepo, fieldDevice.SystemPartID); err != nil {
					return err
				}
				systemPartCache[fieldDevice.SystemPartID] = true
			}
		}

		return nil
	}

	setCreateError := func(resultItem *domainFacility.FieldDeviceCreateResult, err error, defaultField string) {
		if ve, ok := domain.AsValidationError(err); ok {
			for field, msg := range ve.Fields {
				resultItem.Error = msg
				resultItem.ErrorField = field
				return
			}
		}
		resultItem.Error = err.Error()
		resultItem.ErrorField = defaultField
	}

	type createWorkItem struct {
		index int
		item  domainFacility.FieldDeviceCreateItem
	}
	createWork := make([]createWorkItem, 0, len(items))

	for i, item := range items {
		createResult := &result.Results[i]
		createResult.Index = i
		createResult.Success = false

		// Validate the item
		if item.FieldDevice == nil {
			createResult.Error = "field device is required"
			createResult.ErrorField = "fielddevice"
			result.FailureCount++
			continue
		}

		// Check for mutually exclusive options
		if item.ObjectDataID != nil && len(item.BacnetObjects) > 0 {
			createResult.Error = "object_data_id and bacnet_objects are mutually exclusive"
			createResult.ErrorField = "fielddevice"
			result.FailureCount++
			continue
		}

		// Validate required fields
		if err := s.validateRequiredFields(item.FieldDevice); err != nil {
			setCreateError(createResult, err, "fielddevice")
			result.FailureCount++
			continue
		}

		// Ensure parent entities exist
		if err := ensureParentsExistCached(item.FieldDevice); err != nil {
			if err == domain.ErrNotFound {
				createResult.Error = "one or more parent entities (SPS controller, apparat, system part) not found"
				createResult.ErrorField = "fielddevice"
			} else {
				createResult.Error = err.Error()
				createResult.ErrorField = "fielddevice"
			}
			result.FailureCount++
			continue
		}

		// Validate apparat_nr uniqueness
		if item.FieldDevice.ApparatNr == 0 {
			createResult.Error = "apparat_nr is required"
			createResult.ErrorField = "fielddevice.apparat_nr"
			result.FailureCount++
			continue
		}
		if item.FieldDevice.ApparatNr < 1 || item.FieldDevice.ApparatNr > 99 {
			createResult.Error = "apparat_nr must be between 1 and 99"
			createResult.ErrorField = "fielddevice.apparat_nr"
			result.FailureCount++
			continue
		}

		var systemPartID *uuid.UUID
		if item.FieldDevice.SystemPartID != uuid.Nil {
			systemPartID = &item.FieldDevice.SystemPartID
		}
		key := getApparatNumbersKey(
			item.FieldDevice.SPSControllerSystemTypeID,
			systemPartID,
			item.FieldDevice.ApparatID,
		)
		usedSet, ok := usedApparatNumbersCache[key]
		if !ok {
			usedNumbers, err := s.repo.GetUsedApparatNumbers(ctx,
				item.FieldDevice.SPSControllerSystemTypeID,
				systemPartID,
				item.FieldDevice.ApparatID,
			)
			if err != nil {
				createResult.Error = err.Error()
				createResult.ErrorField = "fielddevice.apparat_nr"
				result.FailureCount++
				continue
			}
			usedSet = make(map[int]struct{}, len(usedNumbers))
			for _, n := range usedNumbers {
				usedSet[n] = struct{}{}
			}
			usedApparatNumbersCache[key] = usedSet
		}
		if _, exists := usedSet[item.FieldDevice.ApparatNr]; exists {
			createResult.Error = "apparatnummer ist bereits vergeben"
			createResult.ErrorField = "fielddevice.apparat_nr"
			result.FailureCount++
			continue
		}
		usedSet[item.FieldDevice.ApparatNr] = struct{}{}
		createWork = append(createWork, createWorkItem{index: i, item: item})
	}

	if len(createWork) == 0 {
		return result
	}

	fieldDevices := make([]*domainFacility.FieldDevice, 0, len(createWork))
	for _, work := range createWork {
		fieldDevices = append(fieldDevices, work.item.FieldDevice)
	}

	if err := s.repo.BulkCreate(ctx, fieldDevices, 100); err != nil {
		for _, work := range createWork {
			createResult := &result.Results[work.index]
			setCreateError(createResult, err, "fielddevice")
			result.FailureCount++
		}
		return result
	}

	for _, work := range createWork {
		createResult := &result.Results[work.index]
		if work.item.ObjectDataID != nil {
			if err := s.replaceBacnetObjectsFromObjectData(ctx, work.item.FieldDevice.ID, *work.item.ObjectDataID); err != nil {
				if cleanupErr := s.DeleteByID(ctx, work.item.FieldDevice.ID); cleanupErr != nil {
					err = fmt.Errorf("%w; cleanup failed: %v", err, cleanupErr)
				}
				setCreateError(createResult, err, "fielddevice")
				result.FailureCount++
				continue
			}
		} else if len(work.item.BacnetObjects) > 0 {
			if err := s.replaceBacnetObjects(ctx, work.item.FieldDevice.ID, work.item.BacnetObjects); err != nil {
				if cleanupErr := s.DeleteByID(ctx, work.item.FieldDevice.ID); cleanupErr != nil {
					err = fmt.Errorf("%w; cleanup failed: %v", err, cleanupErr)
				}
				setCreateError(createResult, err, "fielddevice")
				result.FailureCount++
				continue
			}
		}

		createResult.Success = true
		createResult.FieldDevice = work.item.FieldDevice
		result.SuccessCount++
	}

	return result
}

// BulkUpdate updates multiple field devices in a single operation.
// It processes updates ensuring that swaps/permutations within the batch are handled correctly.
// Uniqueness constraints are checked against the database (excluding batch items) AND internally within the batch.
func (s *FieldDeviceService) BulkUpdate(ctx context.Context, updates []domainFacility.BulkFieldDeviceUpdate) *domainFacility.BulkOperationResult {
	result := &domainFacility.BulkOperationResult{
		Results:      make([]domainFacility.BulkOperationResultItem, len(updates)),
		TotalCount:   len(updates),
		SuccessCount: 0,
		FailureCount: 0,
	}

	// 1. Collect all IDs involved in the batch
	ids := make([]uuid.UUID, 0, len(updates))
	updateMap := make(map[uuid.UUID]domainFacility.BulkFieldDeviceUpdate)
	for _, u := range updates {
		ids = append(ids, u.ID)
		updateMap[u.ID] = u
	}

	// 2. Fetch existing items from DB to build context
	existingItems, err := s.repo.GetByIds(ctx, ids)
	if err != nil {
		// If we can't fetch state, we can't safely proceed
		for i := range result.Results {
			result.Results[i].Error = "failed to fetch existing items: " + err.Error()
			result.FailureCount++
		}
		return result
	}

	existingMap := make(map[uuid.UUID]*domainFacility.FieldDevice)
	for _, item := range existingItems {
		existingMap[item.ID] = item
	}

	// 3. Build "Proposed State" for all items
	// This represents what the items WILL look like if updates succeed.
	proposedMap := make(map[uuid.UUID]*domainFacility.FieldDevice)
	for _, update := range updates {
		existing, ok := existingMap[update.ID]
		if !ok {
			continue // Handled in the processing loop
		}

		// Shallow copy existing item to apply updates
		clone := *existing
		if update.HasBMKUpdate() {
			clone.BMK = normalizeOptionalString(update.BMK)
		}
		if update.HasDescriptionUpdate() {
			clone.Description = normalizeOptionalString(update.Description)
		}
		if update.HasTextIndividuellUpdate() {
			clone.TextIndividuell = normalizeOptionalString(update.TextIndividuell)
		}
		if update.ApparatNr != nil {
			clone.ApparatNr = *update.ApparatNr
		}
		if update.ApparatID != nil {
			clone.ApparatID = *update.ApparatID
		}
		if update.SystemPartID != nil {
			clone.SystemPartID = *update.SystemPartID
		}
		proposedMap[update.ID] = &clone
	}

	// Helper to check conflict between two proposed states
	isApparatNrConflict := func(a, b *domainFacility.FieldDevice) bool {
		if a.SPSControllerSystemTypeID != b.SPSControllerSystemTypeID {
			return false
		}
		if a.ApparatID != b.ApparatID {
			return false
		}
		if a.ApparatNr != b.ApparatNr {
			return false
		}
		return a.SystemPartID == b.SystemPartID
	}

	// 4. Process Updates with Granular Field-Level Success Tracking
	// Each device update is now split into 3 independent phases:
	// - Phase 1: Base FieldDevice fields (bmk, description, apparat_nr, etc.)
	// - Phase 2: Specification update/create
	// - Phase 3: BACnet objects patch
	// If any phase fails, we still attempt the other phases and track field-level errors.
	for i, update := range updates {
		resultItem := &result.Results[i]
		resultItem.ID = update.ID
		resultItem.Success = false
		resultItem.Fields = make(map[string]string)

		_, ok := existingMap[update.ID]
		if !ok {
			resultItem.Error = "field device not found"
			result.FailureCount++
			continue
		}

		proposed, ok := proposedMap[update.ID]
		if !ok {
			resultItem.Error = "field device not found"
			result.FailureCount++
			continue
		}

		hasBaseFieldUpdates := update.HasBMKUpdate() || update.HasDescriptionUpdate() || update.HasTextIndividuellUpdate() || update.ApparatNr != nil || update.ApparatID != nil || update.SystemPartID != nil
		phaseErrors := make(map[string]string)
		phaseSuccesses := 0
		totalPhases := 0

		// PHASE 1: Update base FieldDevice fields
		if hasBaseFieldUpdates {
			totalPhases++
			baseUpdateSuccess := true

			// Basic Validation
			if err := s.validateRequiredFields(proposed); err != nil {
				baseUpdateSuccess = false
				if ve, ok := domain.AsValidationError(err); ok {
					for field, msg := range ve.Fields {
						phaseErrors[field] = msg
					}
				} else {
					phaseErrors["fielddevice"] = err.Error()
				}
			}

			// Check if parents exist
			if baseUpdateSuccess {
				if err := s.ensureParentsExist(ctx, proposed); err != nil {
					baseUpdateSuccess = false
					if err == domain.ErrNotFound {
						phaseErrors["fielddevice"] = "one or more parent entities not found"
					} else {
						phaseErrors["fielddevice"] = err.Error()
					}
				}
			}

			// ApparatNr Validation (The "Swap" Logic)
			// Validate if ANY of the uniqueness-constraint fields changed: apparat_nr, apparat_id, system_part_id
			hasApparatNrConstraintChanges := update.ApparatNr != nil || update.ApparatID != nil || update.SystemPartID != nil
			if baseUpdateSuccess && hasApparatNrConstraintChanges {
				for otherID, otherProposed := range proposedMap {
					if otherID == update.ID {
						continue
					}
					if isApparatNrConflict(proposed, otherProposed) {
						baseUpdateSuccess = false
						phaseErrors["fielddevice.apparat_nr"] = "apparatnummer ist bereits vergeben"
						break
					}
				}

				// B. Check for conflicts in DB (Excluding ALL batch items)
				if baseUpdateSuccess {
					if err := s.ensureApparatNrAvailableWithExclusions(ctx, proposed, ids); err != nil {
						baseUpdateSuccess = false
						if ve, ok := domain.AsValidationError(err); ok {
							for field, msg := range ve.Fields {
								phaseErrors[field] = msg
							}
						} else {
							phaseErrors["fielddevice.apparat_nr"] = err.Error()
						}
					}
				}
			}

			// Attempt to save base fields if validation passed
			if baseUpdateSuccess {
				if err := s.repo.Update(ctx, proposed); err != nil {
					phaseErrors["fielddevice"] = err.Error()
				} else {
					phaseSuccesses++
				}
			}
		}

		// PHASE 2: Handle specification update/create (independent of Phase 1)
		if update.Specification != nil && update.Specification.HasChanges() {
			totalPhases++
			if err := s.ApplySpecificationPatch(ctx, proposed.ID, update.Specification); err != nil {
				phaseErrors["specification"] = "failed to update specification: " + err.Error()
			} else {
				phaseSuccesses++
			}
		}

		// PHASE 3: Handle BACnet objects patch updates (independent of Phase 1 and 2)
		if update.BacnetObjects != nil {
			totalPhases++
			if err := s.patchBacnetObjects(ctx, proposed.ID, *update.BacnetObjects); err != nil {
				if ve, ok := domain.AsValidationError(err); ok {
					for field, msg := range ve.Fields {
						phaseErrors[field] = msg
					}
				} else {
					phaseErrors["bacnet_objects"] = "failed to update BACnet objects: " + err.Error()
				}
			} else {
				phaseSuccesses++
			}
		}

		// Determine overall success
		if len(phaseErrors) == 0 && totalPhases > 0 {
			resultItem.Success = true
			result.SuccessCount++
		} else {
			resultItem.Fields = phaseErrors
			// Set a summary error message
			if len(phaseErrors) > 0 {
				// Pick the first error as the main error message
				for _, msg := range phaseErrors {
					resultItem.Error = msg
					break
				}
			}
			result.FailureCount++
		}
	}

	return result
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
