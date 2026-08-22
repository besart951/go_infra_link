package facility

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	cursorcodec "github.com/besart951/go_infra_link/backend/internal/cursor"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	sharedpresenter "github.com/besart951/go_infra_link/backend/internal/handler/presenter/shared"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func startPersistedFacilityCopyJob(c *gin.Context, jobs *facilityservice.CopyJobManager, kind facilityservice.CopyJobKind, sourceID uuid.UUID) bool {
	if jobs == nil || !jobs.SupportsDurableTasks() {
		respondLocalizedError(c, http.StatusServiceUnavailable, "durable_jobs_unavailable", "errors.service_unavailable")
		return true
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
	payload, err := json.Marshal(facilityservice.FacilityCopyTaskPayload{SourceID: sourceID})
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "copy_failed", "facility.copy_failed")
		return true
	}
	job, err := jobs.SubmitTask(c.Request.Context(), facilityservice.CopyJob{
		ID: operationID, OwnerID: actorID, Kind: kind,
		Class: facilityservice.FacilityJobClassMutation, Type: facilityservice.FacilityJobTypeCopy,
		Task: facilityCopyTaskName(kind), Payload: payload,
	})
	if err != nil {
		respondLocalizedError(c, http.StatusServiceUnavailable, "service_unavailable", "errors.service_unavailable")
		return true
	}
	c.JSON(http.StatusAccepted, sharedpresenter.ToCopyJobResponse(job))
	return true
}

func submitFieldDeviceBulkJob(c *gin.Context, jobs *facilityservice.CopyJobManager, task string, payload any, total int) bool {
	if jobs == nil || !jobs.SupportsDurableTasks() {
		respondLocalizedError(c, http.StatusServiceUnavailable, "durable_jobs_unavailable", "errors.service_unavailable")
		return true
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
	encoded, err := json.Marshal(payload)
	if err != nil {
		respondLocalizedError(c, http.StatusBadRequest, "invalid_bulk_payload", "validation.invalid_request")
		return true
	}
	totalItems := int64(total)
	job, err := jobs.SubmitTask(c.Request.Context(), facilityservice.CopyJob{
		ID: operationID, OwnerID: actorID, Kind: facilityservice.CopyJobKindFieldDevice,
		Class: facilityservice.FacilityJobClassMutation, Type: facilityservice.FacilityJobTypeBulk,
		Task: task, Payload: encoded, Total: &totalItems,
	})
	if err != nil {
		respondLocalizedError(c, http.StatusServiceUnavailable, "service_unavailable", "errors.service_unavailable")
		return true
	}
	c.JSON(http.StatusAccepted, sharedpresenter.ToCopyJobResponse(job))
	return true
}

var facilityDeleteTasks = map[facilityservice.CopyJobKind]string{
	facilityservice.CopyJobKindControlCabinet:          facilityservice.FacilityJobTaskDeleteControlCabinet,
	facilityservice.CopyJobKindSPSController:           facilityservice.FacilityJobTaskDeleteSPSController,
	facilityservice.CopyJobKindSPSControllerSystemType: facilityservice.FacilityJobTaskDeleteSPSControllerSystemType,
}

func startPersistedFacilityDeleteJob(c *gin.Context, jobs *facilityservice.CopyJobManager, kind facilityservice.CopyJobKind, sourceID uuid.UUID) {
	if jobs == nil || !jobs.SupportsDurableTasks() {
		respondLocalizedError(c, http.StatusServiceUnavailable, "durable_jobs_unavailable", "errors.service_unavailable")
		return
	}
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		respondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	operationID, ok := copyOperationID(c)
	if !ok {
		return
	}
	payload, err := json.Marshal(facilityservice.FacilityDeleteTaskPayload{SourceID: sourceID})
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "facility.deletion_failed")
		return
	}
	job, err := jobs.SubmitTask(c.Request.Context(), facilityservice.CopyJob{
		ID: operationID, OwnerID: actorID, Kind: kind,
		Class: facilityservice.FacilityJobClassMutation, Type: facilityservice.FacilityJobTypeDelete,
		Task: facilityDeleteTasks[kind], Payload: payload,
		Admission: &facilityservice.FacilityAggregateAdmission{
			ResourceID: sourceID, State: facilityservice.FacilityAggregateStateDeleting,
		},
	})
	if err != nil {
		if errors.Is(err, facilityservice.ErrAggregateLocked) || errors.Is(err, facilityservice.ErrFacilityJobLimit) {
			respondLocalizedError(c, http.StatusConflict, "aggregate_locked", "facility.aggregate_locked")
			return
		}
		if errors.Is(err, facilityservice.ErrAggregateNotFound) {
			respondLocalizedError(c, http.StatusNotFound, "not_found", "facility.fetch_failed")
			return
		}
		respondLocalizedError(c, http.StatusServiceUnavailable, "service_unavailable", "errors.service_unavailable")
		return
	}
	c.JSON(http.StatusAccepted, sharedpresenter.ToCopyJobResponse(job))
}

var facilityCopyTasks = map[facilityservice.CopyJobKind]string{
	facilityservice.CopyJobKindControlCabinet:          facilityservice.FacilityJobTaskCopyControlCabinet,
	facilityservice.CopyJobKindSPSController:           facilityservice.FacilityJobTaskCopySPSController,
	facilityservice.CopyJobKindSPSControllerSystemType: facilityservice.FacilityJobTaskCopySPSControllerSystemType,
	facilityservice.CopyJobKindFieldDevice:             facilityservice.FacilityJobTaskCopyFieldDevice,
	facilityservice.CopyJobKindObjectData:              facilityservice.FacilityJobTaskCopyObjectData,
}

func facilityCopyTaskName(kind facilityservice.CopyJobKind) string {
	return facilityCopyTasks[kind]
}

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

// ListJobs godoc
// @Summary List the current user's facility jobs
// @Tags facility-jobs
// @Produce json
// @Success 200 {object} dto.FacilityJobListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/facility/jobs [get]
func (h *CopyJobHandler) ListJobs(c *gin.Context) {
	if h == nil || h.jobs == nil {
		c.JSON(http.StatusOK, dto.FacilityJobListResponse{Items: []dto.FacilityJobResponse{}})
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	limit := 50
	if raw := c.Query("limit"); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed < 1 || parsed > 100 {
			respondLocalizedInvalidArgument(c, "facility.invalid_pagination")
			return
		}
		limit = parsed
	}
	page, err := h.jobs.ListPage(userID, limit, c.Query("cursor"))
	if err != nil {
		if errors.Is(err, cursorcodec.ErrInvalid) {
			respondLocalizedInvalidArgument(c, "facility.invalid_cursor")
			return
		}
		respondLocalizedError(c, http.StatusServiceUnavailable, "service_unavailable", "errors.service_unavailable")
		return
	}
	items := make([]dto.FacilityJobResponse, len(page.Items))
	for i := range page.Items {
		items[i] = toCopyJobResponse(page.Items[i])
	}
	c.JSON(http.StatusOK, dto.FacilityJobListResponse{Items: items, NextCursor: page.NextCursor, PreviousCursor: page.PreviousCursor})
}

func (h *CopyJobHandler) RetryJob(c *gin.Context) {
	jobID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	job, err := h.jobs.Retry(c.Request.Context(), userID, jobID)
	if err != nil {
		respondLocalizedDomainError(c, err, "job_retry_failed", "facility.job_retry_failed", localizedConflict("facility.job_not_retryable"))
		return
	}
	c.JSON(http.StatusAccepted, toCopyJobResponse(job))
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
