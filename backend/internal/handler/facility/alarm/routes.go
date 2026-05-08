package alarm

import (
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/internal/routing"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	ListAlarmDefinitions  gin.HandlerFunc
	GetAlarmDefinition    gin.HandlerFunc
	CreateAlarmDefinition gin.HandlerFunc
	UpdateAlarmDefinition gin.HandlerFunc
	DeleteAlarmDefinition gin.HandlerFunc
	ListAlarmTypes        gin.HandlerFunc
	CreateAlarmType       gin.HandlerFunc
	GetAlarmType          gin.HandlerFunc
	UpdateAlarmType       gin.HandlerFunc
	DeleteAlarmType       gin.HandlerFunc
	GetAlarmTypeFields    gin.HandlerFunc
	CreateAlarmTypeField  gin.HandlerFunc
	UpdateAlarmTypeField  gin.HandlerFunc
	DeleteAlarmTypeField  gin.HandlerFunc
	ListUnits             gin.HandlerFunc
	GetUnit               gin.HandlerFunc
	CreateUnit            gin.HandlerFunc
	UpdateUnit            gin.HandlerFunc
	DeleteUnit            gin.HandlerFunc
	ListAlarmFields       gin.HandlerFunc
	GetAlarmField         gin.HandlerFunc
	CreateAlarmField      gin.HandlerFunc
	UpdateAlarmField      gin.HandlerFunc
	DeleteAlarmField      gin.HandlerFunc
	GetAlarmSchema        gin.HandlerFunc
	GetAlarmValues        gin.HandlerFunc
	PutAlarmValues        gin.HandlerFunc
}

func Routes(handlers Handlers) []routing.Definition {
	return []routing.Definition{
		routing.Get("/alarm-definitions", domainUser.PermissionAlarmDefinitionRead, handlers.ListAlarmDefinitions),
		routing.Get("/alarm-definitions/:id", domainUser.PermissionAlarmDefinitionRead, handlers.GetAlarmDefinition),
		routing.Post("/alarm-definitions", domainUser.PermissionAlarmDefinitionCreate, handlers.CreateAlarmDefinition),
		routing.Put("/alarm-definitions/:id", domainUser.PermissionAlarmDefinitionUpdate, handlers.UpdateAlarmDefinition),
		routing.Delete("/alarm-definitions/:id", domainUser.PermissionAlarmDefinitionDelete, handlers.DeleteAlarmDefinition),
		routing.Get("/alarm-types", domainUser.PermissionAlarmTypeRead, handlers.ListAlarmTypes),
		routing.Post("/alarm-types", domainUser.PermissionAlarmTypeCreate, handlers.CreateAlarmType),
		routing.Get("/alarm-types/:id", domainUser.PermissionAlarmTypeRead, handlers.GetAlarmType),
		routing.Put("/alarm-types/:id", domainUser.PermissionAlarmTypeUpdate, handlers.UpdateAlarmType),
		routing.Delete("/alarm-types/:id", domainUser.PermissionAlarmTypeDelete, handlers.DeleteAlarmType),
		routing.Get("/alarm-types/:id/fields", domainUser.PermissionAlarmFieldRead, handlers.GetAlarmTypeFields),
		routing.Post("/alarm-types/:id/fields", domainUser.PermissionAlarmFieldCreate, handlers.CreateAlarmTypeField),
		routing.Put("/alarm-type-fields/:id", domainUser.PermissionAlarmFieldUpdate, handlers.UpdateAlarmTypeField),
		routing.Delete("/alarm-type-fields/:id", domainUser.PermissionAlarmFieldDelete, handlers.DeleteAlarmTypeField),
		routing.Get("/alarm-units", domainUser.PermissionUnitRead, handlers.ListUnits),
		routing.Get("/alarm-units/:id", domainUser.PermissionUnitRead, handlers.GetUnit),
		routing.Post("/alarm-units", domainUser.PermissionUnitCreate, handlers.CreateUnit),
		routing.Put("/alarm-units/:id", domainUser.PermissionUnitUpdate, handlers.UpdateUnit),
		routing.Delete("/alarm-units/:id", domainUser.PermissionUnitDelete, handlers.DeleteUnit),
		routing.Get("/alarm-fields", domainUser.PermissionAlarmFieldRead, handlers.ListAlarmFields),
		routing.Get("/alarm-fields/:id", domainUser.PermissionAlarmFieldRead, handlers.GetAlarmField),
		routing.Post("/alarm-fields", domainUser.PermissionAlarmFieldCreate, handlers.CreateAlarmField),
		routing.Put("/alarm-fields/:id", domainUser.PermissionAlarmFieldUpdate, handlers.UpdateAlarmField),
		routing.Delete("/alarm-fields/:id", domainUser.PermissionAlarmFieldDelete, handlers.DeleteAlarmField),
		routing.Get("/bacnet-objects/:id/alarm-schema", domainUser.PermissionBacnetObjectRead, handlers.GetAlarmSchema),
		routing.Get("/bacnet-objects/:id/alarm-values", domainUser.PermissionBacnetObjectRead, handlers.GetAlarmValues),
		routing.Put("/bacnet-objects/:id/alarm-values", domainUser.PermissionBacnetObjectUpdate, handlers.PutAlarmValues),
	}
}
