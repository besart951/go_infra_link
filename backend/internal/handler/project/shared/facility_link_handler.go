package shared

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainproject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	projectdto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	sharedpresenter "github.com/besart951/go_infra_link/backend/internal/handler/presenter/shared"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProjectCopyCommand contains the durable identity and payload of one
// project-scoped hierarchy copy.
type ProjectCopyCommand struct {
	ProjectID uuid.UUID
	SourceID  uuid.UUID
	Kind      facilityservice.FacilityJobKind
	Task      string
}

func RequiredBaseVersion(c *gin.Context) (domain.AggregateVersion, bool) {
	var query projectdto.RequiredBaseVersionQuery
	if !handlerutil.BindQuery(c, &query) {
		return 0, false
	}
	version, err := domain.NewAggregateVersion(query.BaseVersion)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "validation_error", "errors.validation_error")
		return 0, false
	}
	return version, true
}

type projectCopyIdentity struct {
	operationID uuid.UUID
	actorID     uuid.UUID
}

// FacilityLinkHandler owns the HTTP mechanics common to project facility-link
// handlers: project authorization, lifecycle notifications and asynchronous
// facility-job execution. It intentionally does not own domain CRUD operations.
type FacilityLinkHandler struct {
	access       AccessPolicyService
	notify       ProjectChangeNotifier
	facilityJobs *facilityservice.FacilityJobManager
}

func NewFacilityLinkHandler(access AccessPolicyService, notify ProjectChangeNotifier) *FacilityLinkHandler {
	return &FacilityLinkHandler{access: access, notify: notify}
}

func (h *FacilityLinkHandler) ConfigureFacilityJobs(facilityJobs *facilityservice.FacilityJobManager) {
	h.facilityJobs = facilityJobs
}

// ProjectIDWithPermission parses the route parameter and verifies both project
// membership and the requested effective project capability.
func (h *FacilityLinkHandler) ProjectIDWithPermission(c *gin.Context, permissions ...string) (uuid.UUID, bool) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return uuid.Nil, false
	}
	if !EnsureProjectAccessAndAnyPermission(c, h.access, projectID, permissions...) {
		return uuid.Nil, false
	}
	return projectID, true
}

func (h *FacilityLinkHandler) Notify(c *gin.Context, projectID uuid.UUID, eventType string, entityIDs ...string) {
	if h.notify != nil {
		h.notify(c, projectID, eventType, entityIDs...)
	}
}

// StartCopy persists a project hierarchy command and writes its HTTP response.
func (h *FacilityLinkHandler) StartCopy(
	c *gin.Context,
	command ProjectCopyCommand,
) {
	if h.facilityJobs == nil || !h.facilityJobs.SupportsDurableTasks() {
		handlerutil.RespondLocalizedError(c, http.StatusServiceUnavailable, "durable_jobs_unavailable", "errors.service_unavailable")
		return
	}
	identity, ok := projectCopyIdentityFromContext(c)
	if !ok {
		return
	}
	job, err := h.submitProjectCopy(c.Request.Context(), identity, command)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusServiceUnavailable, "service_unavailable", "errors.service_unavailable")
		return
	}

	c.JSON(http.StatusAccepted, sharedpresenter.ToFacilityJobResponse(job))
}

func projectCopyIdentityFromContext(c *gin.Context) (projectCopyIdentity, bool) {
	operationID, err := handlerutil.ParseIdempotencyKey(c)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_operation_id", "project.creation_failed")
		return projectCopyIdentity{}, false
	}
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return projectCopyIdentity{}, false
	}
	return projectCopyIdentity{operationID: operationID, actorID: actorID}, true
}

func (h *FacilityLinkHandler) submitProjectCopy(ctx context.Context, identity projectCopyIdentity, command ProjectCopyCommand) (facilityservice.FacilityJob, error) {
	payload, err := json.Marshal(domainproject.ProjectFacilityCopyCommand{ProjectID: command.ProjectID, SourceID: command.SourceID})
	if err != nil {
		return facilityservice.FacilityJob{}, err
	}
	return h.facilityJobs.SubmitTask(ctx, facilityservice.FacilityJob{
		ID: identity.operationID, OwnerID: identity.actorID, Kind: command.Kind,
		Class: facilityservice.FacilityJobClassMutation, Type: facilityservice.FacilityJobTypeCopy,
		Task: command.Task, Payload: payload,
	})
}
