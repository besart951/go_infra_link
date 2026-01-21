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
	DeleteByIds(ids []uuid.UUID) error
}

type SystemTypeService interface {
	Create(systemType *domainFacility.SystemType) error
	GetByID(id uuid.UUID) (*domainFacility.SystemType, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SystemType], error)
	Update(systemType *domainFacility.SystemType) error
	DeleteByIds(ids []uuid.UUID) error
}

type SystemPartService interface {
	Create(systemPart *domainFacility.SystemPart) error
	GetByID(id uuid.UUID) (*domainFacility.SystemPart, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SystemPart], error)
	Update(systemPart *domainFacility.SystemPart) error
	DeleteByIds(ids []uuid.UUID) error
}

type SpecificationService interface {
	Create(specification *domainFacility.Specification) error
	GetByID(id uuid.UUID) (*domainFacility.Specification, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.Specification], error)
	Update(specification *domainFacility.Specification) error
	DeleteByIds(ids []uuid.UUID) error
}

type ApparatService interface {
	Create(apparat *domainFacility.Apparat) error
	GetByID(id uuid.UUID) (*domainFacility.Apparat, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.Apparat], error)
	Update(apparat *domainFacility.Apparat) error
	DeleteByIds(ids []uuid.UUID) error
}

type BacnetObjectService interface {
	CreateWithParent(bacnetObject *domainFacility.BacnetObject, fieldDeviceID *uuid.UUID, objectDataID *uuid.UUID) error
	GetByID(id uuid.UUID) (*domainFacility.BacnetObject, error)
	Update(bacnetObject *domainFacility.BacnetObject, objectDataID *uuid.UUID) error
}

type FieldDeviceService interface {
	Create(fieldDevice *domainFacility.FieldDevice) error
	CreateWithBacnetObjects(fieldDevice *domainFacility.FieldDevice, objectDataID *uuid.UUID, bacnetObjects []domainFacility.BacnetObject) error
	GetByID(id uuid.UUID) (*domainFacility.FieldDevice, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.FieldDevice], error)
	Update(fieldDevice *domainFacility.FieldDevice) error
	UpdateWithBacnetObjects(fieldDevice *domainFacility.FieldDevice, objectDataID *uuid.UUID, bacnetObjects *[]domainFacility.BacnetObject) error
	DeleteByIds(ids []uuid.UUID) error
	ListBacnetObjects(fieldDeviceID uuid.UUID) ([]domainFacility.BacnetObject, error)
}

type ControlCabinetService interface {
	Create(controlCabinet *domainFacility.ControlCabinet) error
	GetByID(id uuid.UUID) (*domainFacility.ControlCabinet, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.ControlCabinet], error)
	Update(controlCabinet *domainFacility.ControlCabinet) error
	DeleteByIds(ids []uuid.UUID) error
}

type SPSControllerService interface {
	Create(spsController *domainFacility.SPSController) error
	CreateWithSystemTypes(spsController *domainFacility.SPSController, systemTypes []domainFacility.SPSControllerSystemType) error
	GetByID(id uuid.UUID) (*domainFacility.SPSController, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSController], error)
	Update(spsController *domainFacility.SPSController) error
	UpdateWithSystemTypes(spsController *domainFacility.SPSController, systemTypes []domainFacility.SPSControllerSystemType) error
	DeleteByIds(ids []uuid.UUID) error
}
