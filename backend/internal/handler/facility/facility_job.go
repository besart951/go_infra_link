package facility

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	cursorcodec "github.com/besart951/go_infra_link/backend/internal/cursor"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	sharedpresenter "github.com/besart951/go_infra_link/backend/internal/handler/presenter/shared"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func enqueueFacilityCopy(c *gin.Context, jobs *facilityservice.FacilityJobManager, kind facilityservice.FacilityJobKind, sourceID uuid.UUID) bool {
	if jobs == nil || !jobs.SupportsDurableTasks() {
		respondLocalizedError(c, http.StatusServiceUnavailable, "durable_jobs_unavailable", "errors.service_unavailable")
		return true
	}
	operationID, ok := facilityOperationID(c)
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
	job, err := jobs.SubmitTask(c.Request.Context(), facilityservice.FacilityJob{
		ID: operationID, OwnerID: actorID, Kind: kind,
		Class: facilityservice.FacilityJobClassMutation, Type: facilityservice.FacilityJobTypeCopy,
		Task: facilityCopyTaskName(kind), Payload: payload,
	})
	if err != nil {
		respondLocalizedError(c, http.StatusServiceUnavailable, "service_unavailable", "errors.service_unavailable")
		return true
	}
	c.JSON(http.StatusAccepted, sharedpresenter.ToFacilityJobResponse(job))
	return true
}

func submitFieldDeviceBulkJob(c *gin.Context, jobs *facilityservice.FacilityJobManager, task string, payload any, total int) bool {
	if jobs == nil || !jobs.SupportsDurableTasks() {
		respondLocalizedError(c, http.StatusServiceUnavailable, "durable_jobs_unavailable", "errors.service_unavailable")
		return true
	}
	operationID, ok := facilityOperationID(c)
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
	job, err := jobs.SubmitTask(c.Request.Context(), facilityservice.FacilityJob{
		ID: operationID, OwnerID: actorID, Kind: facilityservice.FacilityJobKindFieldDevice,
		Class: facilityservice.FacilityJobClassMutation, Type: facilityservice.FacilityJobTypeBulk,
		Task: task, Payload: encoded, Total: &totalItems,
	})
	if err != nil {
		respondLocalizedError(c, http.StatusServiceUnavailable, "service_unavailable", "errors.service_unavailable")
		return true
	}
	c.JSON(http.StatusAccepted, sharedpresenter.ToFacilityJobResponse(job))
	return true
}

var facilityDeleteTasks = map[facilityservice.FacilityJobKind]string{
	facilityservice.FacilityJobKindControlCabinet:          facilityservice.FacilityJobTaskDeleteControlCabinet,
	facilityservice.FacilityJobKindSPSController:           facilityservice.FacilityJobTaskDeleteSPSController,
	facilityservice.FacilityJobKindSPSControllerSystemType: facilityservice.FacilityJobTaskDeleteSPSControllerSystemType,
}

func startPersistedFacilityDeleteJob(c *gin.Context, jobs *facilityservice.FacilityJobManager, kind facilityservice.FacilityJobKind, sourceID uuid.UUID, baseVersion uint64) {
	if jobs == nil || !jobs.SupportsDurableTasks() {
		respondLocalizedError(c, http.StatusServiceUnavailable, "durable_jobs_unavailable", "errors.service_unavailable")
		return
	}
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		respondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	operationID, ok := facilityOperationID(c)
	if !ok {
		return
	}
	payload, err := json.Marshal(facilityservice.FacilityDeleteTaskPayload{SourceID: sourceID, BaseVersion: baseVersion})
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "facility.deletion_failed")
		return
	}
	job, err := jobs.SubmitTask(c.Request.Context(), facilityservice.FacilityJob{
		ID: operationID, OwnerID: actorID, Kind: kind,
		Class: facilityservice.FacilityJobClassMutation, Type: facilityservice.FacilityJobTypeDelete,
		Task: facilityDeleteTasks[kind], Payload: payload,
		Admission: &facilityservice.FacilityAggregateAdmission{
			ResourceID: sourceID, BaseVersion: baseVersion, State: facilityservice.FacilityAggregateStateDeleting,
		},
	})
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			respondLocalizedError(c, http.StatusConflict, "write_conflict", "errors.write_conflict")
			return
		}
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
	c.JSON(http.StatusAccepted, sharedpresenter.ToFacilityJobResponse(job))
}

var facilityCopyTasks = map[facilityservice.FacilityJobKind]string{
	facilityservice.FacilityJobKindControlCabinet:          facilityservice.FacilityJobTaskCopyControlCabinet,
	facilityservice.FacilityJobKindSPSController:           facilityservice.FacilityJobTaskCopySPSController,
	facilityservice.FacilityJobKindSPSControllerSystemType: facilityservice.FacilityJobTaskCopySPSControllerSystemType,
	facilityservice.FacilityJobKindFieldDevice:             facilityservice.FacilityJobTaskCopyFieldDevice,
	facilityservice.FacilityJobKindObjectData:              facilityservice.FacilityJobTaskCopyObjectData,
}

func facilityCopyTaskName(kind facilityservice.FacilityJobKind) string {
	return facilityCopyTasks[kind]
}

type FacilityJobHandler struct {
	jobs *facilityservice.FacilityJobManager
}

func NewFacilityJobHandler(jobs *facilityservice.FacilityJobManager) *FacilityJobHandler {
	return &FacilityJobHandler{jobs: jobs}
}

// GetJob godoc
// @Summary Get a facility job
// @Tags facility-jobs
// @Produce json
// @Param id path string true "Copy Job ID"
// @Success 200 {object} dto.FacilityJobResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/facility/jobs/{id} [get]
func (h *FacilityJobHandler) GetJob(c *gin.Context) {
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
	c.JSON(http.StatusOK, toFacilityJobResponse(job))
}

// ListJobs godoc
// @Summary List the current user's facility jobs
// @Tags facility-jobs
// @Produce json
// @Success 200 {object} dto.FacilityJobListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/facility/jobs [get]
func (h *FacilityJobHandler) ListJobs(c *gin.Context) {
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
		items[i] = toFacilityJobResponse(page.Items[i])
	}
	c.JSON(http.StatusOK, dto.FacilityJobListResponse{Items: items, NextCursor: page.NextCursor, PreviousCursor: page.PreviousCursor})
}

func (h *FacilityJobHandler) RetryJob(c *gin.Context) {
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
	c.JSON(http.StatusAccepted, toFacilityJobResponse(job))
}

func facilityOperationID(c *gin.Context) (uuid.UUID, bool) {
	operationID, err := handlerutil.ParseIdempotencyKey(c)
	if err != nil {
		respondLocalizedInvalidArgument(c, "facility.invalid_id")
		return operationID, false
	}
	return operationID, true
}

func toFacilityJobResponse(job facilityservice.FacilityJob) dto.FacilityJobResponse {
	return sharedpresenter.ToFacilityJobResponse(job)
}
