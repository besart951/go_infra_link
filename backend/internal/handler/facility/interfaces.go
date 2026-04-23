package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type BuildingService interface {
	Create(ctx context.Context, building *domainFacility.Building) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.Building, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domainFacility.Building, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.Building], error)
	Update(ctx context.Context, building *domainFacility.Building) error
	Validate(ctx context.Context, building *domainFacility.Building, excludeID *uuid.UUID) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type SystemTypeService interface {
	Create(ctx context.Context, systemType *domainFacility.SystemType) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.SystemType, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.SystemType], error)
	Update(ctx context.Context, systemType *domainFacility.SystemType) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type SystemPartService interface {
	Create(ctx context.Context, systemPart *domainFacility.SystemPart) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.SystemPart, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.SystemPart, error)
	GetApparatIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.SystemPart], error)
	Update(ctx context.Context, systemPart *domainFacility.SystemPart) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type ApparatService interface {
	Create(ctx context.Context, apparat *domainFacility.Apparat) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.Apparat, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.Apparat, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.Apparat], error)
	Update(ctx context.Context, apparat *domainFacility.Apparat) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	GetSystemPartIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error)
}

type BacnetObjectService interface {
	CreateWithParent(ctx context.Context, bacnetObject *domainFacility.BacnetObject, fieldDeviceID *uuid.UUID, objectDataID *uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.BacnetObject, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.BacnetObject, error)
	Update(ctx context.Context, bacnetObject *domainFacility.BacnetObject, objectDataID *uuid.UUID) error
	ReplaceForObjectData(ctx context.Context, objectDataID uuid.UUID, inputs []domainFacility.BacnetObject) error
}

type FieldDeviceService interface {
	Create(ctx context.Context, fieldDevice *domainFacility.FieldDevice) error
	CreateWithBacnetObjects(ctx context.Context, fieldDevice *domainFacility.FieldDevice, objectDataID *uuid.UUID, bacnetObjects []domainFacility.BacnetObject) error
	MultiCreate(ctx context.Context, items []domainFacility.FieldDeviceCreateItem) *domainFacility.FieldDeviceMultiCreateResult
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.FieldDevice, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.FieldDevice], error)
	ListWithFilters(ctx context.Context, params domain.PaginationParams, filters domainFacility.FieldDeviceFilterParams) (*domain.PaginatedList[domainFacility.FieldDevice], error)
	ListAvailableApparatNumbers(ctx context.Context, spsControllerSystemTypeID uuid.UUID, systemPartID uuid.UUID, apparatID uuid.UUID) ([]int, error)
	GetFieldDeviceOptions(ctx context.Context) (*domainFacility.FieldDeviceOptions, error)
	Update(ctx context.Context, fieldDevice *domainFacility.FieldDevice) error
	UpdateWithBacnetObjects(ctx context.Context, fieldDevice *domainFacility.FieldDevice, objectDataID *uuid.UUID, bacnetObjects *[]domainFacility.BacnetObject) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	ListBacnetObjects(ctx context.Context, fieldDeviceID uuid.UUID) ([]domainFacility.BacnetObject, error)
	CreateSpecification(ctx context.Context, fieldDeviceID uuid.UUID, specification *domainFacility.Specification) error
	UpdateSpecification(ctx context.Context, fieldDeviceID uuid.UUID, patch *domainFacility.Specification) (*domainFacility.Specification, error)
	BulkUpdate(ctx context.Context, updates []domainFacility.BulkFieldDeviceUpdate) *domainFacility.BulkOperationResult
	BulkDelete(ctx context.Context, ids []uuid.UUID) *domainFacility.BulkOperationResult
}

type ControlCabinetService interface {
	Create(ctx context.Context, controlCabinet *domainFacility.ControlCabinet) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.ControlCabinet, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domainFacility.ControlCabinet, error)
	CopyByID(ctx context.Context, id uuid.UUID) (*domainFacility.ControlCabinet, error)
	GetDeleteImpact(ctx context.Context, id uuid.UUID) (*domainFacility.ControlCabinetDeleteImpact, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.ControlCabinet], error)
	ListByBuildingID(ctx context.Context, buildingID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.ControlCabinet], error)
	Update(ctx context.Context, controlCabinet *domainFacility.ControlCabinet) error
	Validate(ctx context.Context, controlCabinet *domainFacility.ControlCabinet, excludeID *uuid.UUID) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type SPSControllerService interface {
	Create(ctx context.Context, spsController *domainFacility.SPSController) error
	CreateWithSystemTypes(ctx context.Context, spsController *domainFacility.SPSController, systemTypes []domainFacility.SPSControllerSystemType) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSController, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domainFacility.SPSController, error)
	CopyByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSController, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSController], error)
	ListByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSController], error)
	Update(ctx context.Context, spsController *domainFacility.SPSController) error
	UpdateWithSystemTypes(ctx context.Context, spsController *domainFacility.SPSController, systemTypes []domainFacility.SPSControllerSystemType) error
	Validate(ctx context.Context, spsController *domainFacility.SPSController, excludeID *uuid.UUID) error
	NextAvailableGADevice(ctx context.Context, controlCabinetID uuid.UUID, excludeID *uuid.UUID) (string, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type StateTextService interface {
	Create(ctx context.Context, stateText *domainFacility.StateText) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.StateText, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.StateText], error)
	Update(ctx context.Context, stateText *domainFacility.StateText) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type NotificationClassService interface {
	Create(ctx context.Context, notificationClass *domainFacility.NotificationClass) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.NotificationClass, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.NotificationClass], error)
	Update(ctx context.Context, notificationClass *domainFacility.NotificationClass) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type AlarmDefinitionService interface {
	Create(ctx context.Context, alarmDefinition *domainFacility.AlarmDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.AlarmDefinition, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.AlarmDefinition], error)
	Update(ctx context.Context, alarmDefinition *domainFacility.AlarmDefinition) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type ObjectDataService interface {
	Create(ctx context.Context, objectData *domainFacility.ObjectData) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.ObjectData, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.ObjectData], error)
	ListByApparatID(ctx context.Context, page, limit int, search string, apparatID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error)
	ListBySystemPartID(ctx context.Context, page, limit int, search string, systemPartID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error)
	ListByApparatAndSystemPartID(ctx context.Context, page, limit int, search string, apparatID, systemPartID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error)
	Update(ctx context.Context, objectData *domainFacility.ObjectData) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	GetBacnetObjectIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error)
	GetApparatIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error)
	ExistsByDescription(ctx context.Context, projectID *uuid.UUID, description string, excludeID *uuid.UUID) (bool, error)
}

type SPSControllerSystemTypeService interface {
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error)
	ListBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error)
	ListByProjectID(ctx context.Context, projectID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error)
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSControllerSystemType, error)
	CopyByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSControllerSystemType, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type ExportService interface {
	Create(ctx context.Context, req domainExport.Request) (domainExport.Job, error)
	Get(ctx context.Context, id uuid.UUID) (domainExport.Job, error)
}

type AlarmTypeService interface {
	Create(ctx context.Context, alarmType *domainFacility.AlarmType) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.AlarmType, error)
	GetWithFields(ctx context.Context, id uuid.UUID) (*domainFacility.AlarmType, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.AlarmType], error)
	Update(ctx context.Context, alarmType *domainFacility.AlarmType) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type UnitService interface {
	Create(ctx context.Context, unit *domainFacility.Unit) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.Unit, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.Unit], error)
	Update(ctx context.Context, unit *domainFacility.Unit) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type AlarmFieldService interface {
	Create(ctx context.Context, field *domainFacility.AlarmField) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.AlarmField, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.AlarmField], error)
	Update(ctx context.Context, field *domainFacility.AlarmField) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type AlarmTypeFieldService interface {
	Create(ctx context.Context, item *domainFacility.AlarmTypeField) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.AlarmTypeField, error)
	Update(ctx context.Context, item *domainFacility.AlarmTypeField) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type BacnetAlarmValueService interface {
	GetSchema(ctx context.Context, bacnetObjectID uuid.UUID) (*domainFacility.AlarmType, error)
	GetValues(ctx context.Context, bacnetObjectID uuid.UUID) ([]domainFacility.BacnetObjectAlarmValue, error)
	PutValues(ctx context.Context, bacnetObjectID uuid.UUID, values []domainFacility.BacnetObjectAlarmValue) error
}
