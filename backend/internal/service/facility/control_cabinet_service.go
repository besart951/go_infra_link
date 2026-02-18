package facility

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ControlCabinetService struct {
	repo                    domainFacility.ControlCabinetRepository
	buildingRepo            domainFacility.BuildingRepository
	spsControllerRepo       domainFacility.SPSControllerRepository
	spsControllerSystemRepo domainFacility.SPSControllerSystemTypeStore
	fieldDeviceRepo         domainFacility.FieldDeviceStore
	bacnetObjectRepo        domainFacility.BacnetObjectStore
	specificationRepo       domainFacility.SpecificationStore
}

func NewControlCabinetService(
	repo domainFacility.ControlCabinetRepository,
	buildingRepo domainFacility.BuildingRepository,
	spsControllerRepo domainFacility.SPSControllerRepository,
	spsControllerSystemRepo domainFacility.SPSControllerSystemTypeStore,
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	specificationRepo domainFacility.SpecificationStore,
) *ControlCabinetService {
	return &ControlCabinetService{
		repo:                    repo,
		buildingRepo:            buildingRepo,
		spsControllerRepo:       spsControllerRepo,
		spsControllerSystemRepo: spsControllerSystemRepo,
		fieldDeviceRepo:         fieldDeviceRepo,
		bacnetObjectRepo:        bacnetObjectRepo,
		specificationRepo:       specificationRepo,
	}
}

func (s *ControlCabinetService) Create(controlCabinet *domainFacility.ControlCabinet) error {
	if err := s.Validate(controlCabinet, nil); err != nil {
		return err
	}
	if err := s.ensureBuildingExists(controlCabinet.BuildingID); err != nil {
		return err
	}
	return s.repo.Create(controlCabinet)
}

func (s *ControlCabinetService) GetByID(id uuid.UUID) (*domainFacility.ControlCabinet, error) {
	return domain.GetByID(s.repo, id)
}

func (s *ControlCabinetService) GetByIDs(ids []uuid.UUID) ([]domainFacility.ControlCabinet, error) {
	controlCabinets, err := s.repo.GetByIds(ids)
	if err != nil {
		return nil, err
	}
	items := make([]domainFacility.ControlCabinet, len(controlCabinets))
	for i, item := range controlCabinets {
		items[i] = *item
	}
	return items, nil
}

func (s *ControlCabinetService) CopyByID(id uuid.UUID) (*domainFacility.ControlCabinet, error) {
	original, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	baseNr := ""
	if original.ControlCabinetNr != nil {
		baseNr = strings.TrimSpace(*original.ControlCabinetNr)
	}

	nextNr, err := s.nextAvailableControlCabinetNr(original.BuildingID, baseNr)
	if err != nil {
		return nil, err
	}

	copyEntity := &domainFacility.ControlCabinet{
		BuildingID:       original.BuildingID,
		ControlCabinetNr: &nextNr,
	}

	if err := s.Create(copyEntity); err != nil {
		return nil, err
	}

	// Copy all SPS Controllers with their system types, field devices, specifications, and BACnet objects
	if err := s.copySPSControllersForControlCabinet(id, copyEntity.ID); err != nil {
		return nil, err
	}

	return copyEntity, nil
}

func (s *ControlCabinetService) GetDeleteImpact(id uuid.UUID) (*domainFacility.ControlCabinetDeleteImpact, error) {
	// Ensure cabinet exists
	if _, err := s.GetByID(id); err != nil {
		return nil, err
	}

	spsControllerIDs, err := s.spsControllerRepo.GetIDsByControlCabinetID(id)
	if err != nil {
		return nil, err
	}

	spsControllerSystemTypeIDs, err := s.spsControllerSystemRepo.GetIDsBySPSControllerIDs(spsControllerIDs)
	if err != nil {
		return nil, err
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(spsControllerSystemTypeIDs)
	if err != nil {
		return nil, err
	}

	bos, err := s.bacnetObjectRepo.GetByFieldDeviceIDs(fieldDeviceIDs)
	if err != nil {
		return nil, err
	}

	specs, err := s.specificationRepo.GetByFieldDeviceIDs(fieldDeviceIDs)
	if err != nil {
		return nil, err
	}

	return &domainFacility.ControlCabinetDeleteImpact{
		ControlCabinetID:              id,
		SPSControllersCount:           len(spsControllerIDs),
		SPSControllerSystemTypesCount: len(spsControllerSystemTypeIDs),
		FieldDevicesCount:             len(fieldDeviceIDs),
		BacnetObjectsCount:            len(bos),
		SpecificationsCount:           len(specs),
	}, nil
}

func (s *ControlCabinetService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *ControlCabinetService) ListByBuildingID(buildingID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedListByBuildingID(buildingID, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *ControlCabinetService) Update(controlCabinet *domainFacility.ControlCabinet) error {
	if err := s.Validate(controlCabinet, &controlCabinet.ID); err != nil {
		return err
	}
	if err := s.ensureBuildingExists(controlCabinet.BuildingID); err != nil {
		return err
	}
	return s.repo.Update(controlCabinet)
}

func (s *ControlCabinetService) Validate(controlCabinet *domainFacility.ControlCabinet, excludeID *uuid.UUID) error {
	if err := s.validateRequiredFields(controlCabinet); err != nil {
		return err
	}
	if err := s.ensureUnique(controlCabinet, excludeID); err != nil {
		return err
	}
	return nil
}

func (s *ControlCabinetService) DeleteByID(id uuid.UUID) error {
	impact, err := s.GetDeleteImpact(id)
	if err != nil {
		return err
	}

	// Nothing references this cabinet => delete directly
	if impact.SPSControllersCount == 0 {
		return s.repo.DeleteByIds([]uuid.UUID{id})
	}

	spsControllerIDs, err := s.spsControllerRepo.GetIDsByControlCabinetID(id)
	if err != nil {
		return err
	}

	spsControllerSystemTypeIDs, err := s.spsControllerSystemRepo.GetIDsBySPSControllerIDs(spsControllerIDs)
	if err != nil {
		return err
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(spsControllerSystemTypeIDs)
	if err != nil {
		return err
	}

	// Delete dependents in safe order
	if err := s.spsControllerSystemRepo.DeleteBySPSControllerIDs(spsControllerIDs); err != nil {
		return err
	}

	if err := s.bacnetObjectRepo.DeleteByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.specificationRepo.DeleteByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.fieldDeviceRepo.DeleteByIds(fieldDeviceIDs); err != nil {
		return err
	}

	if err := s.spsControllerRepo.DeleteByIds(spsControllerIDs); err != nil {
		return err
	}

	return s.repo.DeleteByIds([]uuid.UUID{id})
}

func (s *ControlCabinetService) ensureBuildingExists(buildingID uuid.UUID) error {
	_, err := domain.GetByID(s.buildingRepo, buildingID)
	return err
}

func (s *ControlCabinetService) validateRequiredFields(controlCabinet *domainFacility.ControlCabinet) error {
	ve := domain.NewValidationError()
	if controlCabinet.BuildingID == uuid.Nil {
		ve = ve.Add("controlcabinet.building_id", "building_id is required")
	}
	if controlCabinet.ControlCabinetNr == nil || strings.TrimSpace(*controlCabinet.ControlCabinetNr) == "" {
		ve = ve.Add("controlcabinet.control_cabinet_nr", "control_cabinet_nr is required")
	} else if len(strings.TrimSpace(*controlCabinet.ControlCabinetNr)) > 11 {
		ve = ve.Add("controlcabinet.control_cabinet_nr", "control_cabinet_nr must be 11 characters or less")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func (s *ControlCabinetService) ensureUnique(controlCabinet *domainFacility.ControlCabinet, excludeID *uuid.UUID) error {
	if controlCabinet.ControlCabinetNr == nil || strings.TrimSpace(*controlCabinet.ControlCabinetNr) == "" || controlCabinet.BuildingID == uuid.Nil {
		return nil
	}
	exists, err := s.repo.ExistsControlCabinetNr(controlCabinet.BuildingID, *controlCabinet.ControlCabinetNr, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return domain.NewValidationError().Add("controlcabinet.control_cabinet_nr", "control_cabinet_nr must be unique within the building")
	}
	return nil
}

func (s *ControlCabinetService) nextAvailableControlCabinetNr(buildingID uuid.UUID, base string) (string, error) {
	for i := 1; i <= 9999; i++ {
		candidate := nextIncrementedValue(base, i, 11)
		exists, err := s.repo.ExistsControlCabinetNr(buildingID, candidate, nil)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}

	return "", domain.ErrConflict
}

// copySPSControllersForControlCabinet copies all SPS controllers from the original control cabinet to the new one
// including all system types, field devices, specifications, and BACnet objects
func (s *ControlCabinetService) copySPSControllersForControlCabinet(originalControlCabinetID, newControlCabinetID uuid.UUID) error {
	// Get all SPS controllers for the original control cabinet
	originalSPSControllers, err := s.listSPSControllersByControlCabinetID(originalControlCabinetID)
	if err != nil {
		return err
	}

	// Copy each SPS controller
	for _, originalSPS := range originalSPSControllers {
		// Determine next device name
		nextDeviceName, err := s.nextAvailableDeviceName(newControlCabinetID, originalSPS.DeviceName)
		if err != nil {
			return err
		}

		// Create copy of SPS controller
		spsCopy := &domainFacility.SPSController{
			ControlCabinetID:  newControlCabinetID,
			GADevice:          nil, // Will be assigned automatically
			DeviceName:        nextDeviceName,
			DeviceDescription: originalSPS.DeviceDescription,
			DeviceLocation:    originalSPS.DeviceLocation,
			IPAddress:         nil, // Clear IP address to avoid conflicts
			Subnet:            originalSPS.Subnet,
			Gateway:           originalSPS.Gateway,
			Vlan:              originalSPS.Vlan,
		}

		if err := s.spsControllerRepo.Create(spsCopy); err != nil {
			return err
		}

		// Copy system types
		newSystemTypeMap, err := s.copySPSControllerSystemTypes(originalSPS.ID, spsCopy.ID)
		if err != nil {
			return err
		}

		// Copy field devices with specifications and BACnet objects
		if err := s.copyFieldDevicesForSPSController(originalSPS.ID, newSystemTypeMap); err != nil {
			return err
		}
	}

	return nil
}

// listSPSControllersByControlCabinetID lists all SPS controllers for a control cabinet
func (s *ControlCabinetService) listSPSControllersByControlCabinetID(controlCabinetID uuid.UUID) ([]domainFacility.SPSController, error) {
	items := make([]domainFacility.SPSController, 0)
	page := 1

	for {
		result, err := s.spsControllerRepo.GetPaginatedListByControlCabinetID(controlCabinetID, domain.PaginationParams{Page: page, Limit: 500})
		if err != nil {
			return nil, err
		}

		items = append(items, result.Items...)
		if page >= result.TotalPages || len(result.Items) == 0 {
			break
		}
		page++
	}

	return items, nil
}

// nextAvailableDeviceName finds the next available device name based on the original name
func (s *ControlCabinetService) nextAvailableDeviceName(controlCabinetID uuid.UUID, baseName string) (string, error) {
	base := strings.TrimSpace(baseName)
	for i := 1; i <= 9999; i++ {
		candidate := nextIncrementedValue(base, i, 100)
		taken, err := s.deviceNameExistsInControlCabinet(controlCabinetID, candidate, nil)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
	}

	return "", domain.ErrConflict
}

// deviceNameExistsInControlCabinet checks if a device name exists in a control cabinet
func (s *ControlCabinetService) deviceNameExistsInControlCabinet(controlCabinetID uuid.UUID, deviceName string, excludeID *uuid.UUID) (bool, error) {
	page := 1
	for {
		result, err := s.spsControllerRepo.GetPaginatedListByControlCabinetID(controlCabinetID, domain.PaginationParams{
			Page:  page,
			Limit: 500,
		})
		if err != nil {
			return false, err
		}

		for i := range result.Items {
			item := result.Items[i]
			if excludeID != nil && item.ID == *excludeID {
				continue
			}
			if strings.EqualFold(strings.TrimSpace(item.DeviceName), strings.TrimSpace(deviceName)) {
				return true, nil
			}
		}

		if page >= result.TotalPages || len(result.Items) == 0 {
			break
		}
		page++
	}

	return false, nil
}

// copySPSControllerSystemTypes copies all system types from the original SPS controller to the new one
// Returns a map of original system type ID to new system type ID
func (s *ControlCabinetService) copySPSControllerSystemTypes(originalSPSControllerID, newSPSControllerID uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	// Get all system types for the original SPS controller
	originalSystemTypes, err := s.listSystemTypesBySPSControllerID(originalSPSControllerID)
	if err != nil {
		return nil, err
	}

	newSystemTypeMap := make(map[uuid.UUID]uuid.UUID, len(originalSystemTypes))
	if len(originalSystemTypes) == 0 {
		return newSystemTypeMap, nil
	}

	// Copy each system type
	for _, originalST := range originalSystemTypes {
		newST := &domainFacility.SPSControllerSystemType{
			Number:          originalST.Number,
			DocumentName:    originalST.DocumentName,
			SPSControllerID: newSPSControllerID,
			SystemTypeID:    originalST.SystemTypeID,
		}

		if err := s.spsControllerSystemRepo.Create(newST); err != nil {
			return nil, err
		}

		newSystemTypeMap[originalST.ID] = newST.ID
	}

	return newSystemTypeMap, nil
}

// listSystemTypesBySPSControllerID lists all system types for an SPS controller
func (s *ControlCabinetService) listSystemTypesBySPSControllerID(spsControllerID uuid.UUID) ([]domainFacility.SPSControllerSystemType, error) {
	items := make([]domainFacility.SPSControllerSystemType, 0)
	page := 1

	for {
		result, err := s.spsControllerSystemRepo.GetPaginatedListBySPSControllerID(spsControllerID, domain.PaginationParams{Page: page, Limit: 500})
		if err != nil {
			return nil, err
		}

		items = append(items, result.Items...)
		if page >= result.TotalPages || len(result.Items) == 0 {
			break
		}
		page++
	}

	return items, nil
}

// copyFieldDevicesForSPSController copies all field devices from the original SPS controller to the new one
// using the system type map to link field devices to the correct new system types
func (s *ControlCabinetService) copyFieldDevicesForSPSController(originalSPSControllerID uuid.UUID, newSystemTypeMap map[uuid.UUID]uuid.UUID) error {
	// Get all field devices for the original SPS controller
	originalFieldDevices, err := s.listFieldDevicesBySPSControllerID(originalSPSControllerID)
	if err != nil {
		return err
	}

	// Copy each field device
	for _, originalFD := range originalFieldDevices {
		newSystemTypeID, ok := newSystemTypeMap[originalFD.SPSControllerSystemTypeID]
		if !ok {
			continue
		}

		fieldDeviceCopy := &domainFacility.FieldDevice{
			BMK:                       originalFD.BMK,
			Description:               originalFD.Description,
			ApparatNr:                 originalFD.ApparatNr,
			SPSControllerSystemTypeID: newSystemTypeID,
			SystemPartID:              originalFD.SystemPartID,
			ApparatID:                 originalFD.ApparatID,
		}

		if err := s.fieldDeviceRepo.Create(fieldDeviceCopy); err != nil {
			return err
		}

		// Copy specification if exists
		if err := copySpecificationForFieldDevice(s.specificationRepo, originalFD.ID, fieldDeviceCopy.ID); err != nil {
			return err
		}

		// Copy BACnet objects if exist
		if err := copyBacnetObjectsForFieldDevice(s.bacnetObjectRepo, originalFD.ID, fieldDeviceCopy.ID); err != nil {
			return err
		}
	}

	return nil
}

// listFieldDevicesBySPSControllerID lists all field devices for an SPS controller
func (s *ControlCabinetService) listFieldDevicesBySPSControllerID(spsControllerID uuid.UUID) ([]domainFacility.FieldDevice, error) {
	items := make([]domainFacility.FieldDevice, 0)
	page := 1
	filters := domainFacility.FieldDeviceFilterParams{SPSControllerID: &spsControllerID}

	for {
		result, err := s.fieldDeviceRepo.GetPaginatedListWithFilters(domain.PaginationParams{Page: page, Limit: 500}, filters)
		if err != nil {
			return nil, err
		}

		items = append(items, result.Items...)
		if page >= result.TotalPages || len(result.Items) == 0 {
			break
		}
		page++
	}

	return items, nil
}
