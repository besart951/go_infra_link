package facility

import (
	"fmt"

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
	specificationRepo           domainFacility.SpecificationRepository
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
	specificationRepo domainFacility.SpecificationRepository,
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
	fieldDevices, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(fieldDevices) == 0 {
		return nil, nil
	}
	return fieldDevices[0], nil
}

func (s *FieldDeviceService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	page, limit = normalizePagination(page, limit)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *FieldDeviceService) Update(fieldDevice *domainFacility.FieldDevice) error {
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

func (s *FieldDeviceService) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
}

func (s *FieldDeviceService) ListBacnetObjects(fieldDeviceID uuid.UUID) ([]domainFacility.BacnetObject, error) {
	// Ensure field device exists (and not deleted)
	fds, err := s.repo.GetByIds([]uuid.UUID{fieldDeviceID})
	if err != nil {
		return nil, err
	}
	if len(fds) == 0 {
		return nil, domain.ErrNotFound
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
	// sps_controller_system_type must exist and not be deleted
	sts, err := s.spsControllerSystemTypeRepo.GetByIds([]uuid.UUID{fieldDevice.SPSControllerSystemTypeID})
	if err != nil {
		return err
	}
	if len(sts) == 0 {
		return domain.ErrNotFound
	}

	// system_type must exist and not be deleted (prevents new instances on soft-deleted/deprecated types)
	systemTypes, err := s.systemTypeRepo.GetByIds([]uuid.UUID{sts[0].SystemTypeID})
	if err != nil {
		return err
	}
	if len(systemTypes) == 0 {
		return domain.ErrNotFound
	}

	// sps_controller must exist and not be deleted
	controllers, err := s.spsControllerRepo.GetByIds([]uuid.UUID{sts[0].SPSControllerID})
	if err != nil {
		return err
	}
	if len(controllers) == 0 {
		return domain.ErrNotFound
	}

	// control cabinet must exist and not be deleted
	cabs, err := s.controlCabinetRepo.GetByIds([]uuid.UUID{controllers[0].ControlCabinetID})
	if err != nil {
		return err
	}
	if len(cabs) == 0 {
		return domain.ErrNotFound
	}

	// building must exist and not be deleted (avoid linking into a soft-deleted building)
	buildings, err := s.buildingRepo.GetByIds([]uuid.UUID{cabs[0].BuildingID})
	if err != nil {
		return err
	}
	if len(buildings) == 0 {
		return domain.ErrNotFound
	}

	// apparat must exist and not be deleted
	apparats, err := s.apparatRepo.GetByIds([]uuid.UUID{fieldDevice.ApparatID})
	if err != nil {
		return err
	}
	if len(apparats) == 0 {
		return domain.ErrNotFound
	}

	// optional parents
	if fieldDevice.SystemPartID != nil {
		parts, err := s.systemPartRepo.GetByIds([]uuid.UUID{*fieldDevice.SystemPartID})
		if err != nil {
			return err
		}
		if len(parts) == 0 {
			return domain.ErrNotFound
		}
	}
	if fieldDevice.SpecificationID != nil {
		specs, err := s.specificationRepo.GetByIds([]uuid.UUID{*fieldDevice.SpecificationID})
		if err != nil {
			return err
		}
		if len(specs) == 0 {
			return domain.ErrNotFound
		}
	}

	return nil
}

func (s *FieldDeviceService) ensureApparatNrAvailable(fieldDevice *domainFacility.FieldDevice, excludeID *uuid.UUID) error {
	if fieldDevice.ApparatNr == nil {
		return fmt.Errorf("apparat_nr is required")
	}
	exists, err := s.repo.ExistsApparatNrConflict(
		fieldDevice.SPSControllerSystemTypeID,
		fieldDevice.SystemPartID,
		fieldDevice.ApparatID,
		*fieldDevice.ApparatNr,
		excludeID,
	)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrConflict
	}
	return nil
}

func (s *FieldDeviceService) replaceBacnetObjects(fieldDeviceID uuid.UUID, bacnetObjects []domainFacility.BacnetObject) error {
	if err := s.bacnetObjectRepo.SoftDeleteByFieldDeviceIDs([]uuid.UUID{fieldDeviceID}); err != nil {
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

func (s *FieldDeviceService) replaceBacnetObjectsFromObjectData(fieldDeviceID uuid.UUID, objectDataID uuid.UUID) error {
	ods, err := s.objectDataRepo.GetByIds([]uuid.UUID{objectDataID})
	if err != nil {
		return err
	}
	if len(ods) == 0 {
		return domain.ErrNotFound
	}
	if !ods[0].IsActive {
		return domain.ErrNotFound
	}

	if err := s.bacnetObjectRepo.SoftDeleteByFieldDeviceIDs([]uuid.UUID{fieldDeviceID}); err != nil {
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
