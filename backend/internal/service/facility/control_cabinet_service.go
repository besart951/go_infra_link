package facility

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type ControlCabinetService struct {
	repo                      domainFacility.ControlCabinetRepository
	buildingRepo              domainFacility.BuildingRepository
	spsControllerRepo         domainFacility.SPSControllerRepository
	spsControllerSystemRepo   domainFacility.SPSControllerSystemTypeStore
	fieldDeviceRepo           domainFacility.FieldDeviceStore
	bacnetObjectRepo          domainFacility.BacnetObjectStore
	specificationRepo         domainFacility.SpecificationStore
	projectControlCabinetRepo domainProject.ProjectControlCabinetRepository
	projectSPSControllerRepo  domainProject.ProjectSPSControllerRepository
	projectFieldDeviceRepo    domainProject.ProjectFieldDeviceRepository
	hierarchyCopier           *HierarchyCopier
}

func NewControlCabinetService(
	repo domainFacility.ControlCabinetRepository,
	buildingRepo domainFacility.BuildingRepository,
	spsControllerRepo domainFacility.SPSControllerRepository,
	spsControllerSystemRepo domainFacility.SPSControllerSystemTypeStore,
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	specificationRepo domainFacility.SpecificationStore,
	projectControlCabinetRepo domainProject.ProjectControlCabinetRepository,
	projectSPSControllerRepo domainProject.ProjectSPSControllerRepository,
	projectFieldDeviceRepo domainProject.ProjectFieldDeviceRepository,
	hierarchyCopier *HierarchyCopier,
) *ControlCabinetService {
	return &ControlCabinetService{
		repo:                      repo,
		buildingRepo:              buildingRepo,
		spsControllerRepo:         spsControllerRepo,
		spsControllerSystemRepo:   spsControllerSystemRepo,
		fieldDeviceRepo:           fieldDeviceRepo,
		bacnetObjectRepo:          bacnetObjectRepo,
		specificationRepo:         specificationRepo,
		projectControlCabinetRepo: projectControlCabinetRepo,
		projectSPSControllerRepo:  projectSPSControllerRepo,
		projectFieldDeviceRepo:    projectFieldDeviceRepo,
		hierarchyCopier:           hierarchyCopier,
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
	return derefSlice(controlCabinets), nil
}

func (s *ControlCabinetService) CopyByID(id uuid.UUID) (*domainFacility.ControlCabinet, error) {
	return s.hierarchyCopier.CopyControlCabinetByID(id)
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

	if err := s.deleteProjectControlCabinetLinksByControlCabinetIDs([]uuid.UUID{id}); err != nil {
		return err
	}
	if err := s.deleteProjectSPSControllerLinksBySPSControllerIDs(spsControllerIDs); err != nil {
		return err
	}
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}

	// Nothing references this cabinet => delete directly
	if impact.SPSControllersCount == 0 {
		return s.repo.DeleteByIds([]uuid.UUID{id})
	}

	// Delete dependents in safe order (children before parents)
	if err := s.bacnetObjectRepo.DeleteByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.specificationRepo.DeleteByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.fieldDeviceRepo.DeleteByIds(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.spsControllerSystemRepo.DeleteBySPSControllerIDs(spsControllerIDs); err != nil {
		return err
	}

	if err := s.spsControllerRepo.DeleteByIds(spsControllerIDs); err != nil {
		return err
	}

	return s.repo.DeleteByIds([]uuid.UUID{id})
}

func (s *ControlCabinetService) deleteProjectControlCabinetLinksByControlCabinetIDs(controlCabinetIDs []uuid.UUID) error {
	if s.projectControlCabinetRepo == nil {
		return nil
	}

	linkIDs, err := collectProjectControlCabinetLinkIDsByControlCabinetIDs(s.projectControlCabinetRepo, controlCabinetIDs)
	if err != nil {
		return err
	}
	if len(linkIDs) == 0 {
		return nil
	}
	return s.projectControlCabinetRepo.DeleteByIds(linkIDs)
}

func (s *ControlCabinetService) deleteProjectSPSControllerLinksBySPSControllerIDs(spsControllerIDs []uuid.UUID) error {
	if s.projectSPSControllerRepo == nil {
		return nil
	}

	linkIDs, err := collectProjectSPSControllerLinkIDsBySPSControllerIDs(s.projectSPSControllerRepo, spsControllerIDs)
	if err != nil {
		return err
	}
	if len(linkIDs) == 0 {
		return nil
	}
	return s.projectSPSControllerRepo.DeleteByIds(linkIDs)
}

func (s *ControlCabinetService) deleteProjectFieldDeviceLinksByFieldDeviceIDs(fieldDeviceIDs []uuid.UUID) error {
	if s.projectFieldDeviceRepo == nil {
		return nil
	}

	linkIDs, err := collectProjectFieldDeviceLinkIDsByFieldDeviceIDs(s.projectFieldDeviceRepo, fieldDeviceIDs)
	if err != nil {
		return err
	}
	if len(linkIDs) == 0 {
		return nil
	}
	return s.projectFieldDeviceRepo.DeleteByIds(linkIDs)
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
	newControlCabinet, err := domain.GetByID(s.repo, newControlCabinetID)
	if err != nil {
		return err
	}

	building, err := domain.GetByID(s.buildingRepo, newControlCabinet.BuildingID)
	if err != nil {
		return err
	}

	newCabinetNr := ""
	if newControlCabinet.ControlCabinetNr != nil {
		newCabinetNr = strings.TrimSpace(*newControlCabinet.ControlCabinetNr)
	}
	buildingIWSCode := strings.TrimSpace(building.IWSCode)

	// Get all SPS controllers for the original control cabinet
	originalSPSControllers, err := s.listSPSControllersByControlCabinetID(originalControlCabinetID)
	if err != nil {
		return err
	}

	// Copy each SPS controller
	for _, originalSPS := range originalSPSControllers {
		gaDevice := ""
		if originalSPS.GADevice != nil {
			gaDevice = strings.ToUpper(strings.TrimSpace(*originalSPS.GADevice))
		}
		if gaDevice == "" {
			nextGADevice, err := s.nextAvailableGADevice(newControlCabinetID)
			if err != nil {
				return err
			}
			gaDevice = nextGADevice
		}
		var gaDevicePtr *string
		if gaDevice != "" {
			gaDevicePtr = &gaDevice
		}

		deviceName := strings.TrimSpace(originalSPS.DeviceName)
		if buildingIWSCode != "" && newCabinetNr != "" && gaDevice != "" {
			deviceName = strings.ToUpper(buildingIWSCode + "_" + newCabinetNr + "_" + gaDevice)
		}

		// Create copy of SPS controller
		spsCopy := &domainFacility.SPSController{
			ControlCabinetID:  newControlCabinetID,
			GADevice:          gaDevicePtr,
			DeviceName:        deviceName,
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

func (s *ControlCabinetService) nextAvailableGADevice(controlCabinetID uuid.UUID) (string, error) {
	devices, err := s.spsControllerRepo.ListGADevicesByControlCabinetID(controlCabinetID)
	if err != nil {
		return "", err
	}

	used := make(map[string]struct{}, len(devices))
	for _, device := range devices {
		normalized := strings.ToUpper(strings.TrimSpace(device))
		if isValidGADevice(normalized) {
			used[normalized] = struct{}{}
		}
	}

	if next, ok := findLowestAvailableGADevice(used); ok {
		return next, nil
	}
	return "", domain.ErrConflict
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

	return copyFieldDevicesWithChildren(s.fieldDeviceRepo, s.specificationRepo, s.bacnetObjectRepo, originalFieldDevices, newSystemTypeMap)
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
