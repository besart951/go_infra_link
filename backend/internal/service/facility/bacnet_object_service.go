package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type BacnetObjectService struct {
	repo                  domainFacility.BacnetObjectStore
	fieldDeviceRepo       domainFacility.FieldDeviceStore
	objectDataRepo        domainFacility.ObjectDataStore
	objectDataBacnetStore domainFacility.ObjectDataBacnetObjectStore
}

func (s *BacnetObjectService) ensureTextFixUniqueForFieldDevice(fieldDeviceID uuid.UUID, textFix string, excludeID *uuid.UUID) error {
	items, err := s.repo.GetByFieldDeviceIDs([]uuid.UUID{fieldDeviceID})
	if err != nil {
		return err
	}
	for _, it := range items {
		if excludeID != nil && it.ID == *excludeID {
			continue
		}
		if it.TextFix == textFix {
			return domain.NewValidationError().Add("fielddevice.bacnetobject.textfix", "textfix must be unique within the field device")
		}
	}
	return nil
}

func (s *BacnetObjectService) ensureTextFixUniqueForObjectData(objectDataID uuid.UUID, textFix string, excludeID *uuid.UUID) error {
	ids, err := s.objectDataRepo.GetBacnetObjectIDs(objectDataID)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	items, err := s.repo.GetByIds(ids)
	if err != nil {
		return err
	}
	for _, it := range items {
		if excludeID != nil && it.ID == *excludeID {
			continue
		}
		if it.TextFix == textFix {
			return domain.NewValidationError().Add("objectdata.bacnetobject.textfix", "textfix must be unique within the object data")
		}
	}
	return nil
}

func NewBacnetObjectService(
	repo domainFacility.BacnetObjectStore,
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	objectDataRepo domainFacility.ObjectDataStore,
	objectDataBacnetStore domainFacility.ObjectDataBacnetObjectStore,
) *BacnetObjectService {
	return &BacnetObjectService{
		repo:                  repo,
		fieldDeviceRepo:       fieldDeviceRepo,
		objectDataRepo:        objectDataRepo,
		objectDataBacnetStore: objectDataBacnetStore,
	}
}

func (s *BacnetObjectService) GetByID(id uuid.UUID) (*domainFacility.BacnetObject, error) {
	items, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, domain.ErrNotFound
	}
	return items[0], nil
}

// CreateWithParent creates a bacnet object either for a field device (fieldDeviceID)
// or for an object data template (objectDataID). Exactly one must be provided.
func (s *BacnetObjectService) CreateWithParent(bacnetObject *domainFacility.BacnetObject, fieldDeviceID *uuid.UUID, objectDataID *uuid.UUID) error {
	if (fieldDeviceID == nil && objectDataID == nil) || (fieldDeviceID != nil && objectDataID != nil) {
		return domain.ErrInvalidArgument
	}

	if fieldDeviceID != nil {
		fds, err := s.fieldDeviceRepo.GetByIds([]uuid.UUID{*fieldDeviceID})
		if err != nil {
			return err
		}
		if len(fds) == 0 {
			return domain.ErrNotFound
		}
		if err := s.ensureTextFixUniqueForFieldDevice(*fieldDeviceID, bacnetObject.TextFix, nil); err != nil {
			return err
		}
		bacnetObject.FieldDeviceID = fieldDeviceID
		return s.repo.Create(bacnetObject)
	}

	ods, err := s.objectDataRepo.GetByIds([]uuid.UUID{*objectDataID})
	if err != nil {
		return err
	}
	if len(ods) == 0 || !ods[0].IsActive {
		return domain.ErrNotFound
	}

	if err := s.ensureTextFixUniqueForObjectData(*objectDataID, bacnetObject.TextFix, nil); err != nil {
		return err
	}

	bacnetObject.FieldDeviceID = nil
	if err := s.repo.Create(bacnetObject); err != nil {
		return err
	}
	return s.objectDataBacnetStore.Add(*objectDataID, bacnetObject.ID)
}

// Update updates a bacnet object. If objectDataID is provided, it will also attach
// the bacnet object to that object data (template) after validating the object data.
func (s *BacnetObjectService) Update(bacnetObject *domainFacility.BacnetObject, objectDataID *uuid.UUID) error {
	items, err := s.repo.GetByIds([]uuid.UUID{bacnetObject.ID})
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return domain.ErrNotFound
	}
	if bacnetObject.FieldDeviceID != nil {
		if err := s.ensureTextFixUniqueForFieldDevice(*bacnetObject.FieldDeviceID, bacnetObject.TextFix, &bacnetObject.ID); err != nil {
			return err
		}
	}

	if objectDataID != nil {
		if err := s.ensureTextFixUniqueForObjectData(*objectDataID, bacnetObject.TextFix, &bacnetObject.ID); err != nil {
			return err
		}
	}

	if err := s.repo.Update(bacnetObject); err != nil {
		return err
	}

	if objectDataID != nil {
		ods, err := s.objectDataRepo.GetByIds([]uuid.UUID{*objectDataID})
		if err != nil {
			return err
		}
		if len(ods) == 0 || !ods[0].IsActive {
			return domain.ErrNotFound
		}
		return s.objectDataBacnetStore.Add(*objectDataID, bacnetObject.ID)
	}

	return nil
}
