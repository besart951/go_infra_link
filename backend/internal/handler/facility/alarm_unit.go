package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type UnitHandler struct {
	crud crudHandler[domainFacility.Unit, dto.CreateUnitRequest, dto.UpdateUnitRequest]
}

func NewUnitHandler(svc UnitService) *UnitHandler {
	return &UnitHandler{crud: newCRUD(
		svc,
		toUnitModel,
		applyUnitUpdate,
		respFn(toUnitResponse),
		listRespFn(toUnitListResponse),
		"facility.not_found",
	)}
}

func (h *UnitHandler) CreateUnit(c *gin.Context) { h.crud.handleCreate(c) }
func (h *UnitHandler) GetUnit(c *gin.Context)    { h.crud.handleGetByID(c) }
func (h *UnitHandler) ListUnits(c *gin.Context)  { h.crud.handleList(c) }
func (h *UnitHandler) UpdateUnit(c *gin.Context) { h.crud.handleUpdate(c) }
func (h *UnitHandler) DeleteUnit(c *gin.Context) { h.crud.handleDelete(c) }
