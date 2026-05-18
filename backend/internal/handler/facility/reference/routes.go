package reference

import (
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/internal/routing"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	CreateSystemType         gin.HandlerFunc
	ListSystemTypes          gin.HandlerFunc
	GetSystemType            gin.HandlerFunc
	UpdateSystemType         gin.HandlerFunc
	DeleteSystemType         gin.HandlerFunc
	CreateSystemPart         gin.HandlerFunc
	ListSystemParts          gin.HandlerFunc
	GetSystemPart            gin.HandlerFunc
	UpdateSystemPart         gin.HandlerFunc
	DeleteSystemPart         gin.HandlerFunc
	CreateApparat            gin.HandlerFunc
	GetApparatsByIDs         gin.HandlerFunc
	ListApparats             gin.HandlerFunc
	GetApparat               gin.HandlerFunc
	UpdateApparat            gin.HandlerFunc
	DeleteApparat            gin.HandlerFunc
	ListStateTexts           gin.HandlerFunc
	GetStateText             gin.HandlerFunc
	CreateStateText          gin.HandlerFunc
	UpdateStateText          gin.HandlerFunc
	DeleteStateText          gin.HandlerFunc
	ListNotificationClasses  gin.HandlerFunc
	GetNotificationClass     gin.HandlerFunc
	CreateNotificationClass  gin.HandlerFunc
	UpdateNotificationClass  gin.HandlerFunc
	DeleteNotificationClass  gin.HandlerFunc
	GetBacnetReferenceUsages gin.HandlerFunc
}

func Routes(handlers Handlers) []routing.Definition {
	return []routing.Definition{
		routing.Post("/system-types", domainUser.PermissionSystemTypeCreate, handlers.CreateSystemType),
		routing.Get("/system-types", domainUser.PermissionSystemTypeRead, handlers.ListSystemTypes),
		routing.Get("/system-types/:id", domainUser.PermissionSystemTypeRead, handlers.GetSystemType),
		routing.Put("/system-types/:id", domainUser.PermissionSystemTypeUpdate, handlers.UpdateSystemType),
		routing.Delete("/system-types/:id", domainUser.PermissionSystemTypeDelete, handlers.DeleteSystemType),
		routing.Post("/system-parts", domainUser.PermissionSystemPartCreate, handlers.CreateSystemPart),
		routing.Get("/system-parts", domainUser.PermissionSystemPartRead, handlers.ListSystemParts),
		routing.Get("/system-parts/:id", domainUser.PermissionSystemPartRead, handlers.GetSystemPart),
		routing.Put("/system-parts/:id", domainUser.PermissionSystemPartUpdate, handlers.UpdateSystemPart),
		routing.Delete("/system-parts/:id", domainUser.PermissionSystemPartDelete, handlers.DeleteSystemPart),
		routing.Post("/apparats", domainUser.PermissionApparatCreate, handlers.CreateApparat),
		routing.Post("/apparats/bulk", domainUser.PermissionApparatRead, handlers.GetApparatsByIDs),
		routing.Get("/apparats", domainUser.PermissionApparatRead, handlers.ListApparats),
		routing.Get("/apparats/:id", domainUser.PermissionApparatRead, handlers.GetApparat),
		routing.Put("/apparats/:id", domainUser.PermissionApparatUpdate, handlers.UpdateApparat),
		routing.Delete("/apparats/:id", domainUser.PermissionApparatDelete, handlers.DeleteApparat),
		routing.Get("/state-texts", domainUser.PermissionStateTextRead, handlers.ListStateTexts),
		routing.Get("/state-texts/:id", domainUser.PermissionStateTextRead, handlers.GetStateText),
		routing.Post("/state-texts", domainUser.PermissionStateTextCreate, handlers.CreateStateText),
		routing.Put("/state-texts/:id", domainUser.PermissionStateTextUpdate, handlers.UpdateStateText),
		routing.Delete("/state-texts/:id", domainUser.PermissionStateTextDelete, handlers.DeleteStateText),
		routing.Get("/notification-classes", domainUser.PermissionNotificationClassRead, handlers.ListNotificationClasses),
		routing.Get("/notification-classes/:id", domainUser.PermissionNotificationClassRead, handlers.GetNotificationClass),
		routing.Post("/notification-classes", domainUser.PermissionNotificationClassCreate, handlers.CreateNotificationClass),
		routing.Put("/notification-classes/:id", domainUser.PermissionNotificationClassUpdate, handlers.UpdateNotificationClass),
		routing.Delete("/notification-classes/:id", domainUser.PermissionNotificationClassDelete, handlers.DeleteNotificationClass),
		routing.Get("/bacnet-reference-usages", domainUser.PermissionBacnetObjectRead, handlers.GetBacnetReferenceUsages),
	}
}
