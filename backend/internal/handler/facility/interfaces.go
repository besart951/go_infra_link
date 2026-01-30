package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type BuildingService interface {
	Create(building *domainFacility.Building) error
	GetByID(id uuid.UUID) (*domainFacility.Building, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.Building], error)
	Update(building *domainFacility.Building) error
	DeleteByID(id uuid.UUID) error
}

type SystemTypeService interface {
	Create(systemType *domainFacility.SystemType) error
	GetByID(id uuid.UUID) (*domainFacility.SystemType, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SystemType], error)
	Update(systemType *domainFacility.SystemType) error
	DeleteByID(id uuid.UUID) error
}

type SystemPartService interface {
	Create(systemPart *domainFacility.SystemPart) error
	GetByID(id uuid.UUID) (*domainFacility.SystemPart, error)
	GetByIDs(ids []uuid.UUID) ([]*domainFacility.SystemPart, error)
	GetApparatIDs(id uuid.UUID) ([]uuid.UUID, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SystemPart], error)
	Update(systemPart *domainFacility.SystemPart) error
	DeleteByID(id uuid.UUID) error
}

type SpecificationService interface {
	Create(specification *domainFacility.Specification) error
	GetByID(id uuid.UUID) (*domainFacility.Specification, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.Specification], error)
	Update(specification *domainFacility.Specification) error
	DeleteByID(id uuid.UUID) error
}

type ApparatService interface {
	Create(apparat *domainFacility.Apparat) error
	GetByID(id uuid.UUID) (*domainFacility.Apparat, error)
	GetByIDs(ids []uuid.UUID) ([]*domainFacility.Apparat, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.Apparat], error)
	Update(apparat *domainFacility.Apparat) error
	DeleteByID(id uuid.UUID) error
	GetSystemPartIDs(id uuid.UUID) ([]uuid.UUID, error)
}

type BacnetObjectService interface {
	CreateWithParent(bacnetObject *domainFacility.BacnetObject, fieldDeviceID *uuid.UUID, objectDataID *uuid.UUID) error
	GetByID(id uuid.UUID) (*domainFacility.BacnetObject, error)
	GetByIDs(ids []uuid.UUID) ([]*domainFacility.BacnetObject, error)
	Update(bacnetObject *domainFacility.BacnetObject, objectDataID *uuid.UUID) error
	ReplaceForObjectData(objectDataID uuid.UUID, inputs []domainFacility.BacnetObject) error
}

type FieldDeviceService interface {
	Create(fieldDevice *domainFacility.FieldDevice) error
	CreateWithBacnetObjects(fieldDevice *domainFacility.FieldDevice, objectDataID *uuid.UUID, bacnetObjects []domainFacility.BacnetObject) error
	GetByID(id uuid.UUID) (*domainFacility.FieldDevice, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.FieldDevice], error)
	ListAvailableApparatNumbers(spsControllerSystemTypeID uuid.UUID, systemPartID *uuid.UUID, apparatID uuid.UUID) ([]int, error)
	GetFieldDeviceOptions() (*domainFacility.FieldDeviceOptions, error)
	Update(fieldDevice *domainFacility.FieldDevice) error
	UpdateWithBacnetObjects(fieldDevice *domainFacility.FieldDevice, objectDataID *uuid.UUID, bacnetObjects *[]domainFacility.BacnetObject) error
	DeleteByID(id uuid.UUID) error
	ListBacnetObjects(fieldDeviceID uuid.UUID) ([]domainFacility.BacnetObject, error)
	CreateSpecification(fieldDeviceID uuid.UUID, specification *domainFacility.Specification) error
	UpdateSpecification(fieldDeviceID uuid.UUID, patch *domainFacility.Specification) (*domainFacility.Specification, error)
}

type ControlCabinetService interface {
	Create(controlCabinet *domainFacility.ControlCabinet) error
	GetByID(id uuid.UUID) (*domainFacility.ControlCabinet, error)
	GetDeleteImpact(id uuid.UUID) (*domainFacility.ControlCabinetDeleteImpact, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.ControlCabinet], error)
	ListByBuildingID(buildingID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.ControlCabinet], error)
	Update(controlCabinet *domainFacility.ControlCabinet) error
	DeleteByID(id uuid.UUID) error
}

type SPSControllerService interface {
	Create(spsController *domainFacility.SPSController) error
	CreateWithSystemTypes(spsController *domainFacility.SPSController, systemTypes []domainFacility.SPSControllerSystemType) error
	GetByID(id uuid.UUID) (*domainFacility.SPSController, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSController], error)
	ListByControlCabinetID(controlCabinetID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSController], error)
	Update(spsController *domainFacility.SPSController) error
	UpdateWithSystemTypes(spsController *domainFacility.SPSController, systemTypes []domainFacility.SPSControllerSystemType) error
	DeleteByID(id uuid.UUID) error
}

type StateTextService interface {
	Create(stateText *domainFacility.StateText) error
	GetByID(id uuid.UUID) (*domainFacility.StateText, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.StateText], error)
	Update(stateText *domainFacility.StateText) error
	DeleteByID(id uuid.UUID) error
}

type NotificationClassService interface {
	Create(notificationClass *domainFacility.NotificationClass) error
	GetByID(id uuid.UUID) (*domainFacility.NotificationClass, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.NotificationClass], error)
	Update(notificationClass *domainFacility.NotificationClass) error
	DeleteByID(id uuid.UUID) error
}

type AlarmDefinitionService interface {
	Create(alarmDefinition *domainFacility.AlarmDefinition) error
	GetByID(id uuid.UUID) (*domainFacility.AlarmDefinition, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.AlarmDefinition], error)
	Update(alarmDefinition *domainFacility.AlarmDefinition) error
	DeleteByID(id uuid.UUID) error
}

type ObjectDataService interface {
	Create(objectData *domainFacility.ObjectData) error
	GetByID(id uuid.UUID) (*domainFacility.ObjectData, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.ObjectData], error)
	ListByApparatID(page, limit int, search string, apparatID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error)
	ListBySystemPartID(page, limit int, search string, systemPartID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error)
	ListByApparatAndSystemPartID(page, limit int, search string, apparatID, systemPartID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error)
	Update(objectData *domainFacility.ObjectData) error
	DeleteByID(id uuid.UUID) error
	GetBacnetObjectIDs(id uuid.UUID) ([]uuid.UUID, error)
	GetApparatIDs(id uuid.UUID) ([]uuid.UUID, error)
}

type SPSControllerSystemTypeService interface {
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error)
	ListBySPSControllerID(spsControllerID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error)
}
