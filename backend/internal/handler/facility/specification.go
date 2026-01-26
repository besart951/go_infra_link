package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type SpecificationHandler struct {
	service SpecificationService
}

func NewSpecificationHandler(service SpecificationService) *SpecificationHandler {
	return &SpecificationHandler{service: service}
}

// CreateSpecification godoc
// @Summary Create a new specification
// @Tags facility-specifications
// @Accept json
// @Produce json
// @Param specification body dto.CreateSpecificationRequest true "Specification data"
// @Success 201 {object} dto.SpecificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/specifications [post]
func (h *SpecificationHandler) CreateSpecification(c *gin.Context) {
	var req dto.CreateSpecificationRequest
	if !bindJSON(c, &req) {
		return
	}

	specification := toSpecificationModel(req)

	if err := h.service.Create(specification); err != nil {
		if errors.Is(err, domain.ErrInvalidArgument) {
			respondInvalidArgument(c, "field_device_id is required")
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			respondConflict(c, "Specification already exists for this field device")
			return
		}
		respondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	c.JSON(http.StatusCreated, toSpecificationResponse(*specification))
}

// GetSpecification godoc
// @Summary Get a specification by ID
// @Tags facility-specifications
// @Produce json
// @Param id path string true "Specification ID"
// @Success 200 {object} dto.SpecificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/specifications/{id} [get]
func (h *SpecificationHandler) GetSpecification(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	specification, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Specification not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toSpecificationResponse(*specification))
}

// ListSpecifications godoc
// @Summary List specifications with pagination
// @Tags facility-specifications
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.SpecificationListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/specifications [get]
func (h *SpecificationHandler) ListSpecifications(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toSpecificationListResponse(result))
}

// UpdateSpecification godoc
// @Summary Update a specification
// @Tags facility-specifications
// @Accept json
// @Produce json
// @Param id path string true "Specification ID"
// @Param specification body dto.UpdateSpecificationRequest true "Specification data"
// @Success 200 {object} dto.SpecificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/specifications/{id} [put]
func (h *SpecificationHandler) UpdateSpecification(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateSpecificationRequest
	if !bindJSON(c, &req) {
		return
	}

	specification, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Specification not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	applySpecificationUpdate(specification, req)

	if err := h.service.Update(specification); err != nil {
		respondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toSpecificationResponse(*specification))
}

// DeleteSpecification godoc
// @Summary Delete a specification
// @Tags facility-specifications
// @Produce json
// @Param id path string true "Specification ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/specifications/{id} [delete]
func (h *SpecificationHandler) DeleteSpecification(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByID(id); err != nil {
		respondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
