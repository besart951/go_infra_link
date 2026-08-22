package shared

import (
	"context"
	"encoding/json"
	"net/http"

	domainproject "github.com/besart951/go_infra_link/backend/internal/domain/project"
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
	Kind      facilityservice.CopyJobKind
	Task      string
}

type projectCopyIdentity struct {
	operationID uuid.UUID
	actorID     uuid.UUID
}

// FacilityLinkHandler owns the HTTP mechanics common to project facility-link
// handlers: project authorization, lifecycle notifications and asynchronous
// copy-job execution. It intentionally does not own domain CRUD operations.
type FacilityLinkHandler struct {
	access   AccessPolicyService
	notify   ProjectChangeNotifier
	copyJobs *facilityservice.CopyJobManager
}

func NewFacilityLinkHandler(access AccessPolicyService, notify ProjectChangeNotifier) *FacilityLinkHandler {
	return &FacilityLinkHandler{access: access, notify: notify}
}

func (h *FacilityLinkHandler) ConfigureCopyJobs(copyJobs *facilityservice.CopyJobManager) {
	h.copyJobs = copyJobs
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
	if h.copyJobs == nil || !h.copyJobs.SupportsDurableTasks() {
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

	c.JSON(http.StatusAccepted, sharedpresenter.ToCopyJobResponse(job))
}

func projectCopyIdentityFromContext(c *gin.Context) (projectCopyIdentity, bool) {
	operationID, err := handlerutil.ParseCopyOperationID(c)
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

func (h *FacilityLinkHandler) submitProjectCopy(ctx context.Context, identity projectCopyIdentity, command ProjectCopyCommand) (facilityservice.CopyJob, error) {
	payload, err := json.Marshal(domainproject.FacilityCopyJobPayload{ProjectID: command.ProjectID, SourceID: command.SourceID})
	if err != nil {
		return facilityservice.CopyJob{}, err
	}
	return h.copyJobs.SubmitTask(ctx, facilityservice.CopyJob{
		ID: identity.operationID, OwnerID: identity.actorID, Kind: command.Kind,
		Class: facilityservice.FacilityJobClassMutation, Type: facilityservice.FacilityJobTypeCopy,
		Task: command.Task, Payload: payload,
	})
}
