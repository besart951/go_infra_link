package facility

import (
	"context"
	"net/http"

	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	sharedpresenter "github.com/besart951/go_infra_link/backend/internal/handler/presenter/shared"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CopyJobHandler struct {
	jobs *facilityservice.CopyJobManager
}

func NewCopyJobHandler(jobs *facilityservice.CopyJobManager) *CopyJobHandler {
	return &CopyJobHandler{jobs: jobs}
}

// GetCopyJob godoc
// @Summary Get a facility copy job
// @Tags facility-copy-jobs
// @Produce json
// @Param id path string true "Copy Job ID"
// @Success 200 {object} dto.CopyJobResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/facility/copy-jobs/{id} [get]
func (h *CopyJobHandler) GetCopyJob(c *gin.Context) {
	if h == nil || h.jobs == nil {
		respondLocalizedError(c, http.StatusNotFound, "not_found", "facility.fetch_failed")
		return
	}
	jobID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	job, err := h.jobs.Get(userID, jobID)
	if err != nil {
		respondLocalizedError(c, http.StatusNotFound, "not_found", "facility.fetch_failed")
		return
	}
	c.JSON(http.StatusOK, toCopyJobResponse(job))
}

func copyOperationID(c *gin.Context) (uuid.UUID, bool) {
	operationID, err := handlerutil.ParseCopyOperationID(c)
	if err != nil {
		respondLocalizedInvalidArgument(c, "facility.invalid_id")
		return operationID, false
	}
	return operationID, true
}

func toCopyJobResponse(job facilityservice.CopyJob) dto.CopyJobResponse {
	return sharedpresenter.ToCopyJobResponse(job)
}

// startFacilityCopyJob owns the transport mechanics shared by the global
// hierarchy-copy endpoints. The callback keeps each domain's copy and
// realtime side effects local to its handler.
func startFacilityCopyJob(
	c *gin.Context,
	jobs *facilityservice.CopyJobManager,
	kind facilityservice.CopyJobKind,
	work func(context.Context, uuid.UUID) error,
) bool {
	if jobs == nil {
		return false
	}

	operationID, ok := copyOperationID(c)
	if !ok {
		return true
	}
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		respondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return true
	}
	job, err := jobs.Start(actorID, operationID, kind, func(ctx context.Context) error {
		return work(ctx, actorID)
	})
	if err != nil {
		respondLocalizedError(c, http.StatusServiceUnavailable, "service_unavailable", "errors.service_unavailable")
		return true
	}

	c.JSON(http.StatusAccepted, sharedpresenter.ToCopyJobResponse(job))
	return true
}
