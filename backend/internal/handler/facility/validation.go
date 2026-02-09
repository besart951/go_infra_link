package facility

import (
	"net/http"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type ValidationHandler struct {
	buildingService       BuildingService
	controlCabinetService ControlCabinetService
	spsControllerService  SPSControllerService
}

func NewValidationHandler(
	buildingService BuildingService,
	controlCabinetService ControlCabinetService,
	spsControllerService SPSControllerService,
) *ValidationHandler {
	return &ValidationHandler{
		buildingService:       buildingService,
		controlCabinetService: controlCabinetService,
		spsControllerService:  spsControllerService,
	}
}

// ValidateBuilding godoc
// @Summary Validate building fields
// @Tags facility-validation
// @Accept json
// @Produce json
// @Param request body dto.ValidateBuildingRequest true "Validation payload"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/buildings/validate [post]
func (h *ValidationHandler) ValidateBuilding(c *gin.Context) {
	var req dto.ValidateBuildingRequest
	if !bindJSON(c, &req) {
		return
	}

	building := &domainFacility.Building{
		IWSCode:       req.IWSCode,
		BuildingGroup: req.BuildingGroup,
	}

	if err := h.buildingService.Validate(building, req.ID); respondValidationOrError(c, err, "validation_failed") {
		return
	}

	c.Status(http.StatusNoContent)
}

// ValidateControlCabinet godoc
// @Summary Validate control cabinet fields
// @Tags facility-validation
// @Accept json
// @Produce json
// @Param request body dto.ValidateControlCabinetRequest true "Validation payload"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/control-cabinets/validate [post]
func (h *ValidationHandler) ValidateControlCabinet(c *gin.Context) {
	var req dto.ValidateControlCabinetRequest
	if !bindJSON(c, &req) {
		return
	}

	cabinet := &domainFacility.ControlCabinet{
		BuildingID:       req.BuildingID,
		ControlCabinetNr: req.ControlCabinetNr,
	}

	if err := h.controlCabinetService.Validate(cabinet, req.ID); respondValidationOrError(c, err, "validation_failed") {
		return
	}

	c.Status(http.StatusNoContent)
}

// ValidateSPSController godoc
// @Summary Validate SPS controller fields
// @Tags facility-validation
// @Accept json
// @Produce json
// @Param request body dto.ValidateSPSControllerRequest true "Validation payload"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/sps-controllers/validate [post]
func (h *ValidationHandler) ValidateSPSController(c *gin.Context) {
	var req dto.ValidateSPSControllerRequest
	if !bindJSON(c, &req) {
		return
	}

	controller := &domainFacility.SPSController{
		ControlCabinetID: req.ControlCabinetID,
		GADevice:         req.GADevice,
		DeviceName:       req.DeviceName,
		IPAddress:        req.IPAddress,
		Subnet:           req.Subnet,
		Gateway:          req.Gateway,
		Vlan:             req.Vlan,
	}

	if err := h.spsControllerService.Validate(controller, req.ID); respondValidationOrError(c, err, "validation_failed") {
		return
	}

	c.Status(http.StatusNoContent)
}
