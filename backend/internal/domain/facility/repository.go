package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	// Building
	GetBuildingByIds(ids []uuid.UUID) ([]*Building, error)
	CreateBuilding(entity *Building) error
	UpdateBuilding(entity *Building) error
	DeleteBuildingByIds(ids []uuid.UUID) error
	GetPaginatedBuildings(params domain.PaginationParams) (*domain.PaginatedList[Building], error)

	// SystemType
	GetSystemTypeByIds(ids []uuid.UUID) ([]*SystemType, error)
	CreateSystemType(entity *SystemType) error
	UpdateSystemType(entity *SystemType) error
	DeleteSystemTypeByIds(ids []uuid.UUID) error
	GetPaginatedSystemTypes(params domain.PaginationParams) (*domain.PaginatedList[SystemType], error)

	// SystemPart
	GetSystemPartByIds(ids []uuid.UUID) ([]*SystemPart, error)
	CreateSystemPart(entity *SystemPart) error
	UpdateSystemPart(entity *SystemPart) error
	DeleteSystemPartByIds(ids []uuid.UUID) error
	GetPaginatedSystemParts(params domain.PaginationParams) (*domain.PaginatedList[SystemPart], error)

	// Specification
	GetSpecificationByIds(ids []uuid.UUID) ([]*Specification, error)
	CreateSpecification(entity *Specification) error
	UpdateSpecification(entity *Specification) error
	DeleteSpecificationByIds(ids []uuid.UUID) error
	GetPaginatedSpecifications(params domain.PaginationParams) (*domain.PaginatedList[Specification], error)

	// StateText
	GetStateTextByIds(ids []uuid.UUID) ([]*StateText, error)
	CreateStateText(entity *StateText) error
	UpdateStateText(entity *StateText) error
	DeleteStateTextByIds(ids []uuid.UUID) error
	GetPaginatedStateTexts(params domain.PaginationParams) (*domain.PaginatedList[StateText], error)

	// NotificationClass
	GetNotificationClassByIds(ids []uuid.UUID) ([]*NotificationClass, error)
	CreateNotificationClass(entity *NotificationClass) error
	UpdateNotificationClass(entity *NotificationClass) error
	DeleteNotificationClassByIds(ids []uuid.UUID) error
	GetPaginatedNotificationClasses(params domain.PaginationParams) (*domain.PaginatedList[NotificationClass], error)

	// AlarmDefinition
	GetAlarmDefinitionByIds(ids []uuid.UUID) ([]*AlarmDefinition, error)
	CreateAlarmDefinition(entity *AlarmDefinition) error
	UpdateAlarmDefinition(entity *AlarmDefinition) error
	DeleteAlarmDefinitionByIds(ids []uuid.UUID) error
	GetPaginatedAlarmDefinitions(params domain.PaginationParams) (*domain.PaginatedList[AlarmDefinition], error)

	// Apparat
	GetApparatByIds(ids []uuid.UUID) ([]*Apparat, error)
	CreateApparat(entity *Apparat) error
	UpdateApparat(entity *Apparat) error
	DeleteApparatByIds(ids []uuid.UUID) error
	GetPaginatedApparats(params domain.PaginationParams) (*domain.PaginatedList[Apparat], error)

	// ObjectData
	GetObjectDataByIds(ids []uuid.UUID) ([]*ObjectData, error)
	CreateObjectData(entity *ObjectData) error
	UpdateObjectData(entity *ObjectData) error
	DeleteObjectDataByIds(ids []uuid.UUID) error
	GetPaginatedObjectData(params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)

	// ControlCabinet
	GetControlCabinetByIds(ids []uuid.UUID) ([]*ControlCabinet, error)
	CreateControlCabinet(entity *ControlCabinet) error
	UpdateControlCabinet(entity *ControlCabinet) error
	DeleteControlCabinetByIds(ids []uuid.UUID) error
	GetPaginatedControlCabinets(params domain.PaginationParams) (*domain.PaginatedList[ControlCabinet], error)

	// SPSController
	GetSPSControllerByIds(ids []uuid.UUID) ([]*SPSController, error)
	CreateSPSController(entity *SPSController) error
	UpdateSPSController(entity *SPSController) error
	DeleteSPSControllerByIds(ids []uuid.UUID) error
	GetPaginatedSPSControllers(params domain.PaginationParams) (*domain.PaginatedList[SPSController], error)

	// SPSControllerSystemType
	GetSPSControllerSystemTypeByIds(ids []uuid.UUID) ([]*SPSControllerSystemType, error)
	CreateSPSControllerSystemType(entity *SPSControllerSystemType) error
	UpdateSPSControllerSystemType(entity *SPSControllerSystemType) error
	DeleteSPSControllerSystemTypeByIds(ids []uuid.UUID) error
	GetPaginatedSPSControllerSystemTypes(params domain.PaginationParams) (*domain.PaginatedList[SPSControllerSystemType], error)

	// FieldDevice
	GetFieldDeviceByIds(ids []uuid.UUID) ([]*FieldDevice, error)
	CreateFieldDevice(entity *FieldDevice) error
	UpdateFieldDevice(entity *FieldDevice) error
	DeleteFieldDeviceByIds(ids []uuid.UUID) error
	GetPaginatedFieldDevices(params domain.PaginationParams) (*domain.PaginatedList[FieldDevice], error)

	// BacnetObject
	GetBacnetObjectByIds(ids []uuid.UUID) ([]*BacnetObject, error)
	CreateBacnetObject(entity *BacnetObject) error
	UpdateBacnetObject(entity *BacnetObject) error
	DeleteBacnetObjectByIds(ids []uuid.UUID) error
	GetPaginatedBacnetObjects(params domain.PaginationParams) (*domain.PaginatedList[BacnetObject], error)

	// ObjectDataHistory
	GetObjectDataHistoryByIds(ids []uuid.UUID) ([]*ObjectDataHistory, error)
	CreateObjectDataHistory(entity *ObjectDataHistory) error
	GetPaginatedObjectDataHistory(params domain.PaginationParams) (*domain.PaginatedList[ObjectDataHistory], error)
}
