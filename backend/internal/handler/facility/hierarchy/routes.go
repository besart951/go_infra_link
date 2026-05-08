package hierarchy

import (
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/internal/routing"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	ValidateBuilding              gin.HandlerFunc
	CreateBuilding                gin.HandlerFunc
	GetBuildingsByIDs             gin.HandlerFunc
	ListBuildings                 gin.HandlerFunc
	GetBuilding                   gin.HandlerFunc
	UpdateBuilding                gin.HandlerFunc
	DeleteBuilding                gin.HandlerFunc
	ValidateControlCabinet        gin.HandlerFunc
	CreateControlCabinet          gin.HandlerFunc
	GetControlCabinetsByIDs       gin.HandlerFunc
	CopyControlCabinet            gin.HandlerFunc
	ListControlCabinets           gin.HandlerFunc
	GetControlCabinet             gin.HandlerFunc
	GetControlCabinetDeleteImpact gin.HandlerFunc
	UpdateControlCabinet          gin.HandlerFunc
	DeleteControlCabinet          gin.HandlerFunc
	ValidateSPSController         gin.HandlerFunc
	CreateSPSController           gin.HandlerFunc
	GetSPSControllersByIDs        gin.HandlerFunc
	CopySPSController             gin.HandlerFunc
	ListSPSControllers            gin.HandlerFunc
	GetNextAvailableGADevice      gin.HandlerFunc
	GetSPSController              gin.HandlerFunc
	UpdateSPSController           gin.HandlerFunc
	DeleteSPSController           gin.HandlerFunc
	ListSPSControllerSystemTypes  gin.HandlerFunc
	GetSPSControllerSystemType    gin.HandlerFunc
	CopySPSControllerSystemType   gin.HandlerFunc
	DeleteSPSControllerSystemType gin.HandlerFunc
}

func Routes(handlers Handlers) []routing.Definition {
	return []routing.Definition{
		routing.Post("/buildings/validate", domainUser.PermissionBuildingCreate, handlers.ValidateBuilding),
		routing.Post("/buildings", domainUser.PermissionBuildingCreate, handlers.CreateBuilding),
		routing.Post("/buildings/bulk", domainUser.PermissionBuildingRead, handlers.GetBuildingsByIDs),
		routing.Get("/buildings", domainUser.PermissionBuildingRead, handlers.ListBuildings),
		routing.Get("/buildings/:id", domainUser.PermissionBuildingRead, handlers.GetBuilding),
		routing.Put("/buildings/:id", domainUser.PermissionBuildingUpdate, handlers.UpdateBuilding),
		routing.Delete("/buildings/:id", domainUser.PermissionBuildingDelete, handlers.DeleteBuilding),
		routing.Post("/control-cabinets/validate", domainUser.PermissionControlCabinetCreate, handlers.ValidateControlCabinet),
		routing.Post("/control-cabinets", domainUser.PermissionControlCabinetCreate, handlers.CreateControlCabinet),
		routing.Post("/control-cabinets/bulk", domainUser.PermissionControlCabinetRead, handlers.GetControlCabinetsByIDs),
		routing.Post("/control-cabinets/:id/copy", domainUser.PermissionControlCabinetCreate, handlers.CopyControlCabinet),
		routing.Get("/control-cabinets", domainUser.PermissionControlCabinetRead, handlers.ListControlCabinets),
		routing.Get("/control-cabinets/:id", domainUser.PermissionControlCabinetRead, handlers.GetControlCabinet),
		routing.Get("/control-cabinets/:id/delete-impact", domainUser.PermissionControlCabinetRead, handlers.GetControlCabinetDeleteImpact),
		routing.Put("/control-cabinets/:id", domainUser.PermissionControlCabinetUpdate, handlers.UpdateControlCabinet),
		routing.Delete("/control-cabinets/:id", domainUser.PermissionControlCabinetDelete, handlers.DeleteControlCabinet),
		routing.Post("/sps-controllers/validate", domainUser.PermissionSPSControllerCreate, handlers.ValidateSPSController),
		routing.Post("/sps-controllers", domainUser.PermissionSPSControllerCreate, handlers.CreateSPSController),
		routing.Post("/sps-controllers/bulk", domainUser.PermissionSPSControllerRead, handlers.GetSPSControllersByIDs),
		routing.Post("/sps-controllers/:id/copy", domainUser.PermissionSPSControllerCreate, handlers.CopySPSController),
		routing.Get("/sps-controllers", domainUser.PermissionSPSControllerRead, handlers.ListSPSControllers),
		routing.Get("/sps-controllers/next-ga-device", domainUser.PermissionSPSControllerRead, handlers.GetNextAvailableGADevice),
		routing.Get("/sps-controllers/:id", domainUser.PermissionSPSControllerRead, handlers.GetSPSController),
		routing.Put("/sps-controllers/:id", domainUser.PermissionSPSControllerUpdate, handlers.UpdateSPSController),
		routing.Delete("/sps-controllers/:id", domainUser.PermissionSPSControllerDelete, handlers.DeleteSPSController),
		routing.Get("/sps-controller-system-types", domainUser.PermissionSPSControllerSystemTypeRead, handlers.ListSPSControllerSystemTypes),
		routing.Get("/sps-controller-system-types/:id", domainUser.PermissionSPSControllerSystemTypeRead, handlers.GetSPSControllerSystemType),
		routing.Post("/sps-controller-system-types/:id/copy", domainUser.PermissionSPSControllerSystemTypeCreate, handlers.CopySPSControllerSystemType),
		routing.Delete("/sps-controller-system-types/:id", domainUser.PermissionSPSControllerSystemTypeDelete, handlers.DeleteSPSControllerSystemType),
	}
}
