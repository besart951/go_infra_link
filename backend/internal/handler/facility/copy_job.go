package facility

import (
	"net/http"

	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
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
	return dto.CopyJobResponse{
		JobID: job.ID, Kind: string(job.Kind), Status: string(job.Status), Progress: job.Progress,
		Stage: job.Stage, Error: job.Error, CreatedAt: job.CreatedAt, UpdatedAt: job.UpdatedAt,
	}
}
