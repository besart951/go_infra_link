package fielddevice

import (
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/internal/routing"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	MultiCreateFieldDevices        gin.HandlerFunc
	GetFieldDeviceOptions          gin.HandlerFunc
	ListAvailableApparatNumbers    gin.HandlerFunc
	ListFieldDevices               gin.HandlerFunc
	GetFieldDevice                 gin.HandlerFunc
	ListFieldDeviceBacnetObjects   gin.HandlerFunc
	CreateFieldDeviceSpecification gin.HandlerFunc
	UpdateFieldDeviceSpecification gin.HandlerFunc
	UpdateFieldDevice              gin.HandlerFunc
	DeleteFieldDevice              gin.HandlerFunc
	BulkUpdateFieldDevices         gin.HandlerFunc
	BulkDeleteFieldDevices         gin.HandlerFunc
	CreateFieldDeviceExport        gin.HandlerFunc
	GetExportStatus                gin.HandlerFunc
	DownloadExport                 gin.HandlerFunc
}

func Routes(handlers Handlers) []routing.Definition {
	return []routing.Definition{
		routing.Post("/field-devices/multi-create", domainUser.PermissionFieldDeviceCreate, handlers.MultiCreateFieldDevices),
		routing.Get("/field-devices/options", domainUser.PermissionFieldDeviceRead, handlers.GetFieldDeviceOptions),
		routing.Get("/field-devices/available-apparat-nr", domainUser.PermissionFieldDeviceRead, handlers.ListAvailableApparatNumbers),
		routing.Get("/field-devices", domainUser.PermissionFieldDeviceRead, handlers.ListFieldDevices),
		routing.Get("/field-devices/:id", domainUser.PermissionFieldDeviceRead, handlers.GetFieldDevice),
		routing.Get("/field-devices/:id/bacnet-objects", domainUser.PermissionFieldDeviceRead, handlers.ListFieldDeviceBacnetObjects),
		routing.Post("/field-devices/:id/specification", domainUser.PermissionSpecificationCreate, handlers.CreateFieldDeviceSpecification),
		routing.Put("/field-devices/:id/specification", domainUser.PermissionSpecificationUpdate, handlers.UpdateFieldDeviceSpecification),
		routing.Put("/field-devices/:id", domainUser.PermissionFieldDeviceUpdate, handlers.UpdateFieldDevice),
		routing.Delete("/field-devices/:id", domainUser.PermissionFieldDeviceDelete, handlers.DeleteFieldDevice),
		routing.Patch("/field-devices/bulk-update", domainUser.PermissionFieldDeviceUpdate, handlers.BulkUpdateFieldDevices),
		routing.Delete("/field-devices/bulk-delete", domainUser.PermissionFieldDeviceDelete, handlers.BulkDeleteFieldDevices),
		routing.Post("/exports/field-devices", domainUser.PermissionFieldDeviceRead, handlers.CreateFieldDeviceExport),
		routing.Get("/exports/jobs/:jobId", domainUser.PermissionFieldDeviceRead, handlers.GetExportStatus),
		routing.Get("/exports/jobs/:jobId/download", domainUser.PermissionFieldDeviceRead, handlers.DownloadExport),
	}
}
