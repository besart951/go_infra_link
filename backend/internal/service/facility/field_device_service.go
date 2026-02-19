package facility

import (
	"fmt"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type FieldDeviceService struct {
	repo                        domainFacility.FieldDeviceStore
	spsControllerSystemTypeRepo domainFacility.SPSControllerSystemTypeStore
	spsControllerRepo           domainFacility.SPSControllerRepository
	controlCabinetRepo          domainFacility.ControlCabinetRepository
	systemTypeRepo              domainFacility.SystemTypeRepository
	buildingRepo                domainFacility.BuildingRepository
	apparatRepo                 domainFacility.ApparatRepository
	systemPartRepo              domainFacility.SystemPartRepository
	specificationRepo           domainFacility.SpecificationStore
	bacnetObjectRepo            domainFacility.BacnetObjectStore
	objectDataRepo              domainFacility.ObjectDataStore
}

func NewFieldDeviceService(
	repo domainFacility.FieldDeviceStore,
	spsControllerSystemTypeRepo domainFacility.SPSControllerSystemTypeStore,
	spsControllerRepo domainFacility.SPSControllerRepository,
	controlCabinetRepo domainFacility.ControlCabinetRepository,
	systemTypeRepo domainFacility.SystemTypeRepository,
	buildingRepo domainFacility.BuildingRepository,
	apparatRepo domainFacility.ApparatRepository,
	systemPartRepo domainFacility.SystemPartRepository,
	specificationRepo domainFacility.SpecificationStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	objectDataRepo domainFacility.ObjectDataStore,
) *FieldDeviceService {
	return &FieldDeviceService{
		repo:                        repo,
		spsControllerSystemTypeRepo: spsControllerSystemTypeRepo,
		spsControllerRepo:           spsControllerRepo,
		controlCabinetRepo:          controlCabinetRepo,
		systemTypeRepo:              systemTypeRepo,
		buildingRepo:                buildingRepo,
		apparatRepo:                 apparatRepo,
		systemPartRepo:              systemPartRepo,
		specificationRepo:           specificationRepo,
		bacnetObjectRepo:            bacnetObjectRepo,
		objectDataRepo:              objectDataRepo,
	}
}

func (s *FieldDeviceService) Create(fieldDevice *domainFacility.FieldDevice) error {
	return s.CreateWithBacnetObjects(fieldDevice, nil, nil)
}

func (s *FieldDeviceService) CreateWithBacnetObjects(fieldDevice *domainFacility.FieldDevice, objectDataID *uuid.UUID, bacnetObjects []domainFacility.BacnetObject) error {
	if objectDataID != nil && len(bacnetObjects) > 0 {
		return domain.ErrInvalidArgument
	}
	if err := s.validateRequiredFields(fieldDevice); err != nil {
		return err
	}
	if err := s.ensureParentsExist(fieldDevice); err != nil {
		return err
	}
	if err := s.ensureApparatNrAvailable(fieldDevice, nil); err != nil {
		return err
	}

	if err := s.repo.Create(fieldDevice); err != nil {
		return err
	}

	if objectDataID != nil {
		return s.replaceBacnetObjectsFromObjectData(fieldDevice.ID, *objectDataID)
	}
	if len(bacnetObjects) > 0 {
		return s.replaceBacnetObjects(fieldDevice.ID, bacnetObjects)
	}
	return nil
}

func (s *FieldDeviceService) GetByID(id uuid.UUID) (*domainFacility.FieldDevice, error) {
	return domain.GetByID(s.repo, id)
}

func (s *FieldDeviceService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	page, limit = domain.NormalizePagination(page, limit, 300)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *FieldDeviceService) ListWithFilters(params domain.PaginationParams, filters domainFacility.FieldDeviceFilterParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 300)
	params.Page = page
	params.Limit = limit
	return s.repo.GetPaginatedListWithFilters(params, filters)
}
func (s *FieldDeviceService) Update(fieldDevice *domainFacility.FieldDevice) error {
	if err := s.validateRequiredFields(fieldDevice); err != nil {
		return err
	}
	if err := s.ensureParentsExist(fieldDevice); err != nil {
		return err
	}
	if err := s.ensureApparatNrAvailable(fieldDevice, &fieldDevice.ID); err != nil {
		return err
	}
	return s.repo.Update(fieldDevice)
}

func (s *FieldDeviceService) UpdateWithBacnetObjects(fieldDevice *domainFacility.FieldDevice, objectDataID *uuid.UUID, bacnetObjects *[]domainFacility.BacnetObject) error {
	if objectDataID != nil && bacnetObjects != nil {
		return domain.ErrInvalidArgument
	}
	if err := s.validateRequiredFields(fieldDevice); err != nil {
		return err
	}
	if err := s.ensureParentsExist(fieldDevice); err != nil {
		return err
	}
	if err := s.ensureApparatNrAvailable(fieldDevice, &fieldDevice.ID); err != nil {
		return err
	}

	if err := s.repo.Update(fieldDevice); err != nil {
		return err
	}

	if objectDataID != nil {
		return s.replaceBacnetObjectsFromObjectData(fieldDevice.ID, *objectDataID)
	}
	if bacnetObjects != nil {
		return s.replaceBacnetObjects(fieldDevice.ID, *bacnetObjects)
	}

	return nil
}

func (s *FieldDeviceService) DeleteByID(id uuid.UUID) error {
	ids := []uuid.UUID{id}
	if err := s.bacnetObjectRepo.DeleteByFieldDeviceIDs(ids); err != nil {
		return err
	}
	if err := s.specificationRepo.DeleteByFieldDeviceIDs(ids); err != nil {
		return err
	}
	return s.repo.DeleteByIds(ids)
}

func (s *FieldDeviceService) CreateSpecification(fieldDeviceID uuid.UUID, specification *domainFacility.Specification) error {
	// Ensure field device exists (and not deleted)
	fieldDevice, err := domain.GetByID(s.repo, fieldDeviceID)
	if err != nil {
		return err
	}

	// Ensure 1:1 uniqueness
	existing, err := s.specificationRepo.GetByFieldDeviceIDs([]uuid.UUID{fieldDeviceID})
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return domain.ErrConflict
	}

	id := fieldDeviceID
	specification.FieldDeviceID = &id
	if err := s.specificationRepo.Create(specification); err != nil {
		return err
	}

	// Update the bidirectional relationship: set field_device.specification_id
	fieldDevice.SpecificationID = &specification.ID
	return s.repo.Update(fieldDevice)
}

func (s *FieldDeviceService) UpdateSpecification(fieldDeviceID uuid.UUID, patch *domainFacility.Specification) (*domainFacility.Specification, error) {
	// Ensure field device exists (and not deleted)
	if _, err := domain.GetByID(s.repo, fieldDeviceID); err != nil {
		return nil, err
	}

	specs, err := s.specificationRepo.GetByFieldDeviceIDs([]uuid.UUID{fieldDeviceID})
	if err != nil {
		return nil, err
	}
	if len(specs) == 0 {
		return nil, domain.ErrNotFound
	}
	spec := specs[0]

	if patch.SpecificationSupplier != nil {
		spec.SpecificationSupplier = patch.SpecificationSupplier
	}
	if patch.SpecificationBrand != nil {
		spec.SpecificationBrand = patch.SpecificationBrand
	}
	if patch.SpecificationType != nil {
		spec.SpecificationType = patch.SpecificationType
	}
	if patch.AdditionalInfoMotorValve != nil {
		spec.AdditionalInfoMotorValve = patch.AdditionalInfoMotorValve
	}
	if patch.AdditionalInfoSize != nil {
		spec.AdditionalInfoSize = patch.AdditionalInfoSize
	}
	if patch.AdditionalInformationInstallationLocation != nil {
		spec.AdditionalInformationInstallationLocation = patch.AdditionalInformationInstallationLocation
	}
	if patch.ElectricalConnectionPH != nil {
		spec.ElectricalConnectionPH = patch.ElectricalConnectionPH
	}
	if patch.ElectricalConnectionACDC != nil {
		spec.ElectricalConnectionACDC = patch.ElectricalConnectionACDC
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

	if err := s.specificationRepo.Update(spec); err != nil {
		return nil, err
	}
	return spec, nil
}

func (s *FieldDeviceService) ListBacnetObjects(fieldDeviceID uuid.UUID) ([]domainFacility.BacnetObject, error) {
	// Ensure field device exists (and not deleted)
	if _, err := domain.GetByID(s.repo, fieldDeviceID); err != nil {
		return nil, err
	}

	objs, err := s.bacnetObjectRepo.GetByFieldDeviceIDs([]uuid.UUID{fieldDeviceID})
	if err != nil {
		return nil, err
	}
	out := make([]domainFacility.BacnetObject, 0, len(objs))
	for _, o := range objs {
		out = append(out, *o)
	}
	return out, nil
}

func (s *FieldDeviceService) ensureParentsExist(fieldDevice *domainFacility.FieldDevice) error {
	// With cascaded deletes, we only need to check direct parents.
	// If SPSControllerSystemType exists, all ancestors (SPSController, ControlCabinet, Building) are guaranteed to exist.

	// 1. sps_controller_system_type must exist and not be deleted
	sts, err := domain.GetByID(s.spsControllerSystemTypeRepo, fieldDevice.SPSControllerSystemTypeID)
	if err != nil {
		return err
	}

	// 2. system_type must exist (business rule: prevent creating instances with deprecated types)
	if _, err := domain.GetByID(s.systemTypeRepo, sts.SystemTypeID); err != nil {
		return err
	}

	// 3. apparat must exist and not be deleted
	if _, err := domain.GetByID(s.apparatRepo, fieldDevice.ApparatID); err != nil {
		return err
	}

	// 4. system_part (optional) must exist if provided
	if fieldDevice.SystemPartID != uuid.Nil {
		if _, err := domain.GetByID(s.systemPartRepo, fieldDevice.SystemPartID); err != nil {
			return err
		}
	}

	return nil
}

func (s *FieldDeviceService) ensureApparatNrAvailable(fieldDevice *domainFacility.FieldDevice, excludeID *uuid.UUID) error {
	var excludeIDs []uuid.UUID
	if excludeID != nil {
		excludeIDs = []uuid.UUID{*excludeID}
	}
	return s.ensureApparatNrAvailableWithExclusions(fieldDevice, excludeIDs)
}

func (s *FieldDeviceService) ensureApparatNrAvailableWithExclusions(fieldDevice *domainFacility.FieldDevice, excludeIDs []uuid.UUID) error {
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

	exists, err := s.repo.ExistsApparatNrConflict(
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

func (s *FieldDeviceService) ListAvailableApparatNumbers(spsControllerSystemTypeID uuid.UUID, systemPartID *uuid.UUID, apparatID uuid.UUID) ([]int, error) {
	used, err := s.repo.GetUsedApparatNumbers(spsControllerSystemTypeID, systemPartID, apparatID)
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
func (s *FieldDeviceService) GetFieldDeviceOptions() (*domainFacility.FieldDeviceOptions, error) {
	// Fetch all active object datas (templates) with their apparats
	objectDatas, err := s.objectDataRepo.GetTemplates()
	if err != nil {
		return nil, err
	}

	// Build sets of unique apparats and system parts from the object datas
	apparatSet := make(map[uuid.UUID]*domainFacility.Apparat)
	systemPartSet := make(map[uuid.UUID]*domainFacility.SystemPart)

	// Collect all apparats from object datas
	for _, od := range objectDatas {
		if od == nil {
			continue
		}
		for _, app := range od.Apparats {
			if app != nil {
				apparatSet[app.ID] = app
			}
		}
	}

	// Fetch full apparat data with system parts for the collected apparats
	apparatIDs := make([]uuid.UUID, 0, len(apparatSet))
	for id := range apparatSet {
		apparatIDs = append(apparatIDs, id)
	}

	if len(apparatIDs) > 0 {
		fullApparats, err := s.apparatRepo.GetByIds(apparatIDs)
		if err != nil {
			return nil, err
		}
		for _, app := range fullApparats {
			if app != nil {
				apparatSet[app.ID] = app
				// Collect system parts from each apparat
				for _, sp := range app.SystemParts {
					if sp != nil {
						systemPartSet[sp.ID] = sp
					}
				}
			}
		}
	}

	// Build relationship maps
	apparatToSystemPart := make(map[uuid.UUID][]uuid.UUID)
	for id, apparat := range apparatSet {
		systemPartIDs := make([]uuid.UUID, 0, len(apparat.SystemParts))
		for _, sp := range apparat.SystemParts {
			if sp != nil {
				systemPartIDs = append(systemPartIDs, sp.ID)
			}
		}
		apparatToSystemPart[id] = systemPartIDs
	}

	objectDataToApparat := make(map[uuid.UUID][]uuid.UUID)
	for _, od := range objectDatas {
		if od == nil {
			continue
		}
		apparatIDs := make([]uuid.UUID, 0, len(od.Apparats))
		for _, app := range od.Apparats {
			if app != nil {
				apparatIDs = append(apparatIDs, app.ID)
			}
		}
		objectDataToApparat[od.ID] = apparatIDs
	}

	// Convert maps to slices
	apparats := make([]domainFacility.Apparat, 0, len(apparatSet))
	for _, app := range apparatSet {
		apparats = append(apparats, *app)
	}

	systemParts := make([]domainFacility.SystemPart, 0, len(systemPartSet))
	for _, sp := range systemPartSet {
		systemParts = append(systemParts, *sp)
	}

	objectDataValues := make([]domainFacility.ObjectData, 0, len(objectDatas))
	for _, od := range objectDatas {
		if od != nil {
			objectDataValues = append(objectDataValues, *od)
		}
	}

	return &domainFacility.FieldDeviceOptions{
		Apparats:            apparats,
		SystemParts:         systemParts,
		ObjectDatas:         objectDataValues,
		ApparatToSystemPart: apparatToSystemPart,
		ObjectDataToApparat: objectDataToApparat,
	}, nil
}

// GetFieldDeviceOptionsForProject returns all metadata needed for creating/editing field devices within a project.
// This fetches object data that belongs to the specified project (project_id = projectID AND is_active = true).
func (s *FieldDeviceService) GetFieldDeviceOptionsForProject(projectID uuid.UUID) (*domainFacility.FieldDeviceOptions, error) {
	// Fetch object datas for the project with their apparats
	objectDatas, err := s.objectDataRepo.GetForProject(projectID)
	if err != nil {
		return nil, err
	}

	// Build sets of unique apparats and system parts from the object datas
	apparatSet := make(map[uuid.UUID]*domainFacility.Apparat)
	systemPartSet := make(map[uuid.UUID]*domainFacility.SystemPart)

	// Collect all apparats from object datas
	for _, od := range objectDatas {
		if od == nil {
			continue
		}
		for _, app := range od.Apparats {
			if app != nil {
				apparatSet[app.ID] = app
			}
		}
	}

	// Fetch full apparat data with system parts for the collected apparats
	apparatIDs := make([]uuid.UUID, 0, len(apparatSet))
	for id := range apparatSet {
		apparatIDs = append(apparatIDs, id)
	}

	if len(apparatIDs) > 0 {
		fullApparats, err := s.apparatRepo.GetByIds(apparatIDs)
		if err != nil {
			return nil, err
		}
		for _, app := range fullApparats {
			if app != nil {
				apparatSet[app.ID] = app
				// Collect system parts from each apparat
				for _, sp := range app.SystemParts {
					if sp != nil {
						systemPartSet[sp.ID] = sp
					}
				}
			}
		}
	}

	// Build relationship maps
	apparatToSystemPart := make(map[uuid.UUID][]uuid.UUID)
	for id, apparat := range apparatSet {
		systemPartIDs := make([]uuid.UUID, 0, len(apparat.SystemParts))
		for _, sp := range apparat.SystemParts {
			if sp != nil {
				systemPartIDs = append(systemPartIDs, sp.ID)
			}
		}
		apparatToSystemPart[id] = systemPartIDs
	}

	objectDataToApparat := make(map[uuid.UUID][]uuid.UUID)
	for _, od := range objectDatas {
		if od == nil {
			continue
		}
		apparatIDs := make([]uuid.UUID, 0, len(od.Apparats))
		for _, app := range od.Apparats {
			if app != nil {
				apparatIDs = append(apparatIDs, app.ID)
			}
		}
		objectDataToApparat[od.ID] = apparatIDs
	}

	// Convert maps to slices
	apparats := make([]domainFacility.Apparat, 0, len(apparatSet))
	for _, app := range apparatSet {
		apparats = append(apparats, *app)
	}

	systemParts := make([]domainFacility.SystemPart, 0, len(systemPartSet))
	for _, sp := range systemPartSet {
		systemParts = append(systemParts, *sp)
	}

	objectDataValues := make([]domainFacility.ObjectData, 0, len(objectDatas))
	for _, od := range objectDatas {
		if od != nil {
			objectDataValues = append(objectDataValues, *od)
		}
	}

	return &domainFacility.FieldDeviceOptions{
		Apparats:            apparats,
		SystemParts:         systemParts,
		ObjectDatas:         objectDataValues,
		ApparatToSystemPart: apparatToSystemPart,
		ObjectDataToApparat: objectDataToApparat,
	}, nil
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
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func (s *FieldDeviceService) replaceBacnetObjects(fieldDeviceID uuid.UUID, bacnetObjects []domainFacility.BacnetObject) error {
	ve := domain.NewValidationError()
	seen := make(map[string]int, len(bacnetObjects))
	for i, obj := range bacnetObjects {
		obj.TextFix = normalizeBacnetTextFix(obj.TextFix)
		bacnetObjects[i].TextFix = obj.TextFix
		if obj.TextFix == "" {
			ve = ve.Add(fmt.Sprintf("bacnet_objects.%d.text_fix", i), "text_fix is required")
			continue
		}
		if prevIdx, ok := seen[obj.TextFix]; ok {
			ve = ve.Add(fmt.Sprintf("bacnet_objects.%d.text_fix", i), fmt.Sprintf("text_fix must be unique within the field device (duplicate of row %d)", prevIdx))
		}
		seen[obj.TextFix] = i
	}
	if len(ve.Fields) > 0 {
		return ve
	}

	if err := s.bacnetObjectRepo.DeleteByFieldDeviceIDs([]uuid.UUID{fieldDeviceID}); err != nil {
		return err
	}

	for i := range bacnetObjects {
		obj := bacnetObjects[i]
		id := fieldDeviceID
		obj.FieldDeviceID = &id
		if err := s.bacnetObjectRepo.Create(&obj); err != nil {
			return err
		}
	}
	return nil
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
	if patch.AlarmDefinitionID != nil {
		target.AlarmDefinitionID = patch.AlarmDefinitionID
	}
}

func (s *FieldDeviceService) patchBacnetObjects(fieldDeviceID uuid.UUID, patches []domainFacility.BacnetObjectPatch) error {
	if len(patches) == 0 {
		return nil
	}

	existingItems, err := s.bacnetObjectRepo.GetByFieldDeviceIDs([]uuid.UUID{fieldDeviceID})
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

	seen := make(map[string]uuid.UUID, len(existingItems))
	for _, item := range existingItems {
		candidate := item
		if updated, ok := updatedMap[item.ID]; ok {
			candidate = updated
		}
		if candidate.TextFix == "" {
			continue
		}
		if prevID, ok := seen[candidate.TextFix]; ok {
			if _, ok := patchedIDs[candidate.ID]; ok {
				ve = ve.Add(
					fmt.Sprintf("bacnet_objects.%s.text_fix", candidate.ID),
					fmt.Sprintf("text_fix must be unique within the field device (duplicate of %s)", prevID),
				)
			}
			if _, ok := patchedIDs[prevID]; ok {
				ve = ve.Add(
					fmt.Sprintf("bacnet_objects.%s.text_fix", prevID),
					fmt.Sprintf("text_fix must be unique within the field device (duplicate of %s)", candidate.ID),
				)
			}
			continue
		}
		seen[candidate.TextFix] = candidate.ID
	}

	if len(ve.Fields) > 0 {
		return ve
	}

	for _, updated := range updatedMap {
		if err := s.bacnetObjectRepo.Update(updated); err != nil {
			return err
		}
	}

	return nil
}

func (s *FieldDeviceService) replaceBacnetObjectsFromObjectData(fieldDeviceID uuid.UUID, objectDataID uuid.UUID) error {
	od, err := domain.GetByID(s.objectDataRepo, objectDataID)
	if err != nil {
		return err
	}
	if !od.IsActive {
		return domain.ErrNotFound
	}

	if err := s.bacnetObjectRepo.DeleteByFieldDeviceIDs([]uuid.UUID{fieldDeviceID}); err != nil {
		return err
	}

	ids, err := s.objectDataRepo.GetBacnetObjectIDs(objectDataID)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	templates, err := s.bacnetObjectRepo.GetByIds(ids)
	if err != nil {
		return err
	}
	if len(templates) != len(ids) {
		return domain.ErrNotFound
	}

	templateToClone := make(map[uuid.UUID]*domainFacility.BacnetObject, len(templates))
	templateRef := make(map[uuid.UUID]*uuid.UUID, len(templates))
	for _, t := range templates {
		clone := &domainFacility.BacnetObject{
			TextFix:             t.TextFix,
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
			AlarmDefinitionID:   t.AlarmDefinitionID,
		}
		templateToClone[t.ID] = clone
		templateRef[t.ID] = t.SoftwareReferenceID
	}

	// First pass: create clones without software references.
	oldToNew := make(map[uuid.UUID]uuid.UUID, len(templates))
	for tid, clone := range templateToClone {
		if err := s.bacnetObjectRepo.Create(clone); err != nil {
			return err
		}
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
		if err := s.bacnetObjectRepo.Update(clone); err != nil {
			return err
		}
	}

	return nil
}

// MultiCreate creates multiple field devices in a single operation.
// It validates each device independently and continues on failures.
// Returns detailed results for each device creation attempt.
func (s *FieldDeviceService) MultiCreate(items []domainFacility.FieldDeviceCreateItem) *domainFacility.FieldDeviceMultiCreateResult {
	result := &domainFacility.FieldDeviceMultiCreateResult{
		Results:       make([]domainFacility.FieldDeviceCreateResult, len(items)),
		TotalRequests: len(items),
		SuccessCount:  0,
		FailureCount:  0,
	}

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
			if ve, ok := domain.AsValidationError(err); ok {
				// Get first validation error
				for field, msg := range ve.Fields {
					createResult.Error = msg
					createResult.ErrorField = field
					break
				}
			} else {
				createResult.Error = err.Error()
				createResult.ErrorField = "fielddevice"
			}
			result.FailureCount++
			continue
		}

		// Ensure parent entities exist
		if err := s.ensureParentsExist(item.FieldDevice); err != nil {
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
		if err := s.ensureApparatNrAvailable(item.FieldDevice, nil); err != nil {
			if ve, ok := domain.AsValidationError(err); ok {
				// Get first validation error
				for field, msg := range ve.Fields {
					createResult.Error = msg
					createResult.ErrorField = field
					break
				}
			} else {
				createResult.Error = err.Error()
				createResult.ErrorField = "fielddevice.apparat_nr"
			}
			result.FailureCount++
			continue
		}

		// Create the field device
		if err := s.CreateWithBacnetObjects(item.FieldDevice, item.ObjectDataID, item.BacnetObjects); err != nil {
			if ve, ok := domain.AsValidationError(err); ok {
				// Get first validation error
				for field, msg := range ve.Fields {
					createResult.Error = msg
					createResult.ErrorField = field
					break
				}
			} else {
				createResult.Error = err.Error()
				createResult.ErrorField = "fielddevice"
			}
			result.FailureCount++
			continue
		}

		// Success
		createResult.Success = true
		createResult.FieldDevice = item.FieldDevice
		result.SuccessCount++
	}

	return result
}

// BulkUpdate updates multiple field devices in a single operation.
// It processes updates ensuring that swaps/permutations within the batch are handled correctly.
// Uniqueness constraints are checked against the database (excluding batch items) AND internally within the batch.
func (s *FieldDeviceService) BulkUpdate(updates []domainFacility.BulkFieldDeviceUpdate) *domainFacility.BulkOperationResult {
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
	existingItems, err := s.repo.GetByIds(ids)
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
		if update.BMK != nil {
			clone.BMK = update.BMK
		}
		if update.Description != nil {
			clone.Description = update.Description
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

		hasBaseFieldUpdates := update.BMK != nil || update.Description != nil || update.ApparatNr != nil || update.ApparatID != nil || update.SystemPartID != nil
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
				if err := s.ensureParentsExist(proposed); err != nil {
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
					if err := s.ensureApparatNrAvailableWithExclusions(proposed, ids); err != nil {
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
				if err := s.repo.Update(proposed); err != nil {
					phaseErrors["fielddevice"] = err.Error()
				} else {
					phaseSuccesses++
				}
			}
		}

		// PHASE 2: Handle specification update/create (independent of Phase 1)
		if update.Specification != nil {
			totalPhases++
			specs, err := s.specificationRepo.GetByFieldDeviceIDs([]uuid.UUID{proposed.ID})
			if err != nil {
				phaseErrors["specification"] = "failed to fetch specification: " + err.Error()
			} else {
				if len(specs) > 0 {
					if _, err := s.UpdateSpecification(proposed.ID, update.Specification); err != nil {
						phaseErrors["specification"] = "failed to update specification: " + err.Error()
					} else {
						phaseSuccesses++
					}
				} else {
					if err := s.CreateSpecification(proposed.ID, update.Specification); err != nil {
						phaseErrors["specification"] = "failed to create specification: " + err.Error()
					} else {
						phaseSuccesses++
					}
				}
			}
		}

		// PHASE 3: Handle BACnet objects patch updates (independent of Phase 1 and 2)
		if update.BacnetObjects != nil {
			totalPhases++
			if err := s.patchBacnetObjects(proposed.ID, *update.BacnetObjects); err != nil {
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
func (s *FieldDeviceService) BulkDelete(ids []uuid.UUID) *domainFacility.BulkOperationResult {
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

		if err := s.DeleteByID(id); err != nil {
			resultItem.Error = err.Error()
			result.FailureCount++
			continue
		}

		resultItem.Success = true
		result.SuccessCount++
	}

	return result
}
