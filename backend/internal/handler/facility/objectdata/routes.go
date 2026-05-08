package objectdata

import (
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/internal/routing"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	CreateBacnetObject         gin.HandlerFunc
	UpdateBacnetObject         gin.HandlerFunc
	ListObjectData             gin.HandlerFunc
	GetObjectData              gin.HandlerFunc
	GetObjectDataBacnetObjects gin.HandlerFunc
	CreateObjectData           gin.HandlerFunc
	UpdateObjectData           gin.HandlerFunc
	DeleteObjectData           gin.HandlerFunc
}

func Routes(handlers Handlers) []routing.Definition {
	return []routing.Definition{
		routing.Post("/bacnet-objects", domainUser.PermissionBacnetObjectCreate, handlers.CreateBacnetObject),
		routing.Put("/bacnet-objects/:id", domainUser.PermissionBacnetObjectUpdate, handlers.UpdateBacnetObject),
		routing.Get("/object-data", domainUser.PermissionObjectDataRead, handlers.ListObjectData),
		routing.Get("/object-data/:id", domainUser.PermissionObjectDataRead, handlers.GetObjectData),
		routing.Get("/object-data/:id/bacnet-objects", domainUser.PermissionObjectDataRead, handlers.GetObjectDataBacnetObjects),
		routing.Post("/object-data", domainUser.PermissionObjectDataCreate, handlers.CreateObjectData),
		routing.Put("/object-data/:id", domainUser.PermissionObjectDataUpdate, handlers.UpdateObjectData),
		routing.Delete("/object-data/:id", domainUser.PermissionObjectDataDelete, handlers.DeleteObjectData),
	}
}
