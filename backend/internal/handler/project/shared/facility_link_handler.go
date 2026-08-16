package shared

import (
	"context"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	sharedpresenter "github.com/besart951/go_infra_link/backend/internal/handler/presenter/shared"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProjectCopyJobNotifier publishes a completed project-scoped copy for the
// initiating user. The concrete copied resource remains local to its handler.
type ProjectCopyJobNotifier func(context.Context, *uuid.UUID, uuid.UUID, string, string)

// FacilityLinkHandler owns the HTTP mechanics common to project facility-link
// handlers: project authorization, lifecycle notifications and asynchronous
// copy-job execution. It intentionally does not own domain CRUD operations.
type FacilityLinkHandler struct {
	access     AccessPolicyService
	notify     ProjectChangeNotifier
	copyJobs   *facilityservice.CopyJobManager
	notifyCopy ProjectCopyJobNotifier
}

func NewFacilityLinkHandler(access AccessPolicyService, notify ProjectChangeNotifier) *FacilityLinkHandler {
	return &FacilityLinkHandler{access: access, notify: notify}
}

func (h *FacilityLinkHandler) ConfigureCopyJobs(copyJobs *facilityservice.CopyJobManager, notify ProjectCopyJobNotifier) {
	h.copyJobs = copyJobs
	h.notifyCopy = notify
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

// StartCopy starts a project hierarchy copy when asynchronous jobs are
// configured. It writes the common HTTP response and returns true when the
// caller must stop handling the request; false selects the synchronous
// compatibility path.
func (h *FacilityLinkHandler) StartCopy(
	c *gin.Context,
	projectID uuid.UUID,
	kind facilityservice.CopyJobKind,
	eventType string,
	work func(context.Context) (string, error),
) bool {
	if h.copyJobs == nil {
		return false
	}

	operationID, err := handlerutil.ParseCopyOperationID(c)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_operation_id", "project.creation_failed")
		return true
	}
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return true
	}

	job, err := h.copyJobs.Start(actorID, operationID, kind, func(ctx context.Context) error {
		entityID, err := work(ctx)
		if err == nil && h.notifyCopy != nil {
			h.notifyCopy(ctx, &actorID, projectID, eventType, entityID)
		}
		return err
	})
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusServiceUnavailable, "service_unavailable", "errors.service_unavailable")
		return true
	}

	c.JSON(http.StatusAccepted, sharedpresenter.ToCopyJobResponse(job))
	return true
}
