package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/gin-gonic/gin"
)

// CreateSPSControllerSystemType godoc
// @Summary Create an SPS controller system type assignment
// @Tags facility-sps-controller-system-types
// @Accept json
// @Produce json
// @Param system_type body dto.CreateSPSControllerSystemTypeRequest true "SPS controller system type data"
// @Success 201 {object} SPSControllerSystemTypeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/sps-controller-system-types [post]
func (h *SPSControllerSystemTypeHandler) CreateSPSControllerSystemType(c *gin.Context) {
	var req dto.CreateSPSControllerSystemTypeRequest
	if !bindJSON(c, &req) {
		return
	}
	item := &domainFacility.SPSControllerSystemType{
		SPSControllerID: req.SPSControllerID,
		SystemTypeID:    req.SystemTypeID,
		Number:          req.Number,
		DocumentName:    req.DocumentName,
	}
	if err := h.service.Create(c.Request.Context(), item); err != nil {
		respondLocalizedDomainError(c, err, "creation_failed", "facility.creation_failed")
		return
	}
	if h.collaboration != nil {
		h.collaboration.BroadcastSPSControllerSystemTypeChange(c.Request.Context(), currentActorID(c), item.SPSControllerID, item.ID, "created")
	}
	c.JSON(http.StatusCreated, toSPSControllerSystemTypeResponse(*item))
}

// UpdateSPSControllerSystemType godoc
// @Summary Update an SPS controller system type
// @Tags facility-sps-controller-system-types
// @Accept json
// @Produce json
// @Param id path string true "SPS Controller System Type ID"
// @Param system_type body dto.UpdateSPSControllerSystemTypeRequest true "SPS controller system type data"
// @Success 200 {object} SPSControllerSystemTypeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/sps-controller-system-types/{id} [put]
// @Router /api/v1/facility/sps-controller-system-types/{id} [patch]
func (h *SPSControllerSystemTypeHandler) UpdateSPSControllerSystemType(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateSPSControllerSystemTypeRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		respondLocalizedDomainError(c, err, "fetch_failed", "facility.fetch_failed", localizedNotFound("facility.sps_controller_system_type_not_found"))
		return
	}
	baseVersion := item.Version
	if req.BaseVersion != nil {
		baseVersion = *req.BaseVersion
		item.Version = *req.BaseVersion
	}
	if req.Number != nil {
		item.Number = req.Number
	}
	if req.DocumentName != nil {
		item.DocumentName = req.DocumentName
	}
	if err := h.service.Update(c.Request.Context(), item); err != nil {
		if current, getErr := h.service.GetByID(c.Request.Context(), id); getErr == nil && respondWriteConflict(c, err, "sps_controller", current.SPSControllerID, baseVersion, []string{"system_types." + id.String()}, current.Version, toSPSControllerSystemTypeResponse(*current)) {
			return
		}
		respondLocalizedDomainError(c, err, "update_failed", "facility.update_failed")
		return
	}
	if h.collaboration != nil {
		h.collaboration.BroadcastSPSControllerSystemTypeChange(c.Request.Context(), currentActorID(c), item.SPSControllerID, item.ID, "updated", req.ChangedFields()...)
	}
	c.JSON(http.StatusOK, toSPSControllerSystemTypeResponse(*item))
}

type SPSControllerSystemTypeHandler struct {
	service       SPSControllerSystemTypeService
	collaboration ProjectRefreshBroadcaster
	copyJobs      *facilityservice.CopyJobManager
}

func NewSPSControllerSystemTypeHandler(service SPSControllerSystemTypeService, collaboration ...ProjectRefreshBroadcaster) *SPSControllerSystemTypeHandler {
	h := &SPSControllerSystemTypeHandler{service: service}
	if len(collaboration) > 0 {
		h.collaboration = collaboration[0]
	}
	return h
}

func NewSPSControllerSystemTypeHandlerWithCopyJobs(service SPSControllerSystemTypeService, collaboration ProjectRefreshBroadcaster, copyJobs *facilityservice.CopyJobManager) *SPSControllerSystemTypeHandler {
	handler := NewSPSControllerSystemTypeHandler(service, collaboration)
	handler.copyJobs = copyJobs
	return handler
}

// ListSPSControllerSystemTypes godoc
// @Summary List SPS controller system types with pagination
// @Tags facility-sps-controller-system-types
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Param sps_controller_id query string false "SPS Controller ID(s), accepts a single UUID or a | separated list"
// @Param project_id query string false "Project ID (filter by project)"
// @Success 200 {object} SPSControllerSystemTypeListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/sps-controller-system-types [get]
func (h *SPSControllerSystemTypeHandler) ListSPSControllerSystemTypes(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	spsControllerIDs, ok := parseUUIDListQueryParam(c, "sps_controller_id")
	if !ok {
		return
	}

	projectID, ok := parseUUIDQueryParam(c, "project_id")
	if !ok {
		return
	}

	ctx := c.Request.Context()

	var result *domain.PaginatedList[domainFacility.SPSControllerSystemType]
	var err error
	if len(spsControllerIDs) == 1 {
		result, err = h.service.ListBySPSControllerID(ctx, spsControllerIDs[0], query.Page, query.Limit, query.Search)
	} else if len(spsControllerIDs) > 1 {
		result, err = h.service.ListBySPSControllerIDs(ctx, spsControllerIDs, query.Page, query.Limit, query.Search)
	} else if projectID != nil {
		result, err = h.service.ListByProjectID(ctx, *projectID, query.Page, query.Limit, query.Search)
	} else {
		result, err = h.service.List(ctx, query.Page, query.Limit, query.Search)
	}
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toSPSControllerSystemTypeListResponse(result))
}

// GetSPSControllerSystemType godoc
// @Summary Get an SPS controller system type by ID
// @Tags facility-sps-controller-system-types
// @Produce json
// @Param id path string true "SPS Controller System Type ID"
// @Success 200 {object} SPSControllerSystemTypeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/sps-controller-system-types/{id} [get]
func (h *SPSControllerSystemTypeHandler) GetSPSControllerSystemType(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.sps_controller_system_type_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toSPSControllerSystemTypeResponse(*item))
}

// CopySPSControllerSystemType godoc
// @Summary Copy an SPS controller system type
// @Tags facility-sps-controller-system-types
// @Produce json
// @Param id path string true "SPS Controller System Type ID"
// @Param X-Copy-Operation-ID header string false "Client-generated copy operation UUID"
// @Success 202 {object} dto.CopyJobResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/v1/facility/sps-controller-system-types/{id}/copy [post]
func (h *SPSControllerSystemTypeHandler) CopySPSControllerSystemType(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	startPersistedFacilityCopyJob(c, h.copyJobs, facilityservice.CopyJobKindSPSControllerSystemType, id)
}

// DeleteSPSControllerSystemType godoc
// @Summary Delete an SPS controller system type
// @Tags facility-sps-controller-system-types
// @Produce json
// @Param id path string true "SPS Controller System Type ID"
// @Success 202 {object} dto.FacilityJobResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/sps-controller-system-types/{id} [delete]
func (h *SPSControllerSystemTypeHandler) DeleteSPSControllerSystemType(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	startPersistedFacilityDeleteJob(c, h.copyJobs, facilityservice.CopyJobKindSPSControllerSystemType, id)
}
