package project

import (
	"context"
	"strconv"
	"strings"

	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	projectshared "github.com/besart951/go_infra_link/backend/internal/handler/project/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	lifecycle     ProjectLifecycleService
	access        ProjectAccessPolicyService
	workflow      ProjectWorkflowService
	facilityLink  ProjectFacilityLinkService
	projectDelete ProjectDeleter
	collaboration *ProjectCollaborationHub
	notifications NotificationEventDispatcher
}

func NewProjectHandler(lifecycle ProjectLifecycleService, access ProjectAccessPolicyService, membership ProjectMembershipService, facilityLink ProjectFacilityLinkService) *ProjectHandler {
	return newProjectHandler(
		lifecycle,
		access,
		membership,
		newWorkflowFromServices(lifecycle, membership),
		facilityLink,
		nil,
		NewProjectCollaborationHub(),
		nil,
	)
}

func newProjectHandler(
	lifecycle ProjectLifecycleService,
	access ProjectAccessPolicyService,
	membership ProjectMembershipService,
	workflow ProjectWorkflowService,
	facilityLink ProjectFacilityLinkService,
	projectDelete ProjectDeleter,
	collaboration *ProjectCollaborationHub,
	notifications NotificationEventDispatcher,
) *ProjectHandler {
	if workflow == nil {
		workflow = newWorkflowFromServices(lifecycle, membership)
	}
	return &ProjectHandler{
		lifecycle:     lifecycle,
		access:        access,
		workflow:      workflow,
		facilityLink:  facilityLink,
		projectDelete: projectDelete,
		collaboration: collaboration,
		notifications: notifications,
	}
}

func (h *ProjectHandler) notifyProjectChange(c *gin.Context, projectID uuid.UUID, eventType string, entityIDs ...string) {
	var actorID *uuid.UUID
	if userID, ok := middleware.GetUserID(c); ok {
		actorID = &userID
	}

	if h.collaboration != nil {
		if scope, ok := refreshScopeForProjectEvent(eventType); ok {
			h.collaboration.BroadcastRefreshRequest(projectID, actorID, scope, entityIDs)
		}
	}

	h.notifyProjectEvent(c, projectID, eventType, entityIDs...)
}

func (h *ProjectHandler) notifyProjectEvent(c *gin.Context, projectID uuid.UUID, eventType string, entityIDs ...string) {
	var actorID *uuid.UUID
	if userID, ok := middleware.GetUserID(c); ok {
		actorID = &userID
	}

	if h.notifications != nil {
		metadata := map[string]string{
			"project_id": projectID.String(),
			"count":      strconv.Itoa(len(entityIDs)),
		}
		if len(entityIDs) > 0 {
			metadata["entity_ids"] = strings.Join(entityIDs, ",")
		}
		resourceType := resourceTypeForProjectNotificationEvent(eventType)
		if resourceType != "" {
			metadata["resource_type"] = resourceType
		}
		resourceID := resourceIDForProjectNotificationEvent(resourceType, projectID, entityIDs)
		dispatchCtx := context.WithoutCancel(c.Request.Context())
		go func() {
			_ = h.notifications.DispatchEvent(dispatchCtx, domainNotification.DispatchEventInput{
				ActorID:      actorID,
				EventKey:     eventType,
				ProjectID:    &projectID,
				ResourceType: resourceType,
				ResourceID:   resourceID,
				Metadata:     metadata,
			})
		}()
	}
}

func resourceTypeForProjectNotificationEvent(eventType string) string {
	switch {
	case eventType == "project.updated" || eventType == "project.deleted" || eventType == "project.phase.changed":
		return "project"
	case strings.HasPrefix(eventType, "project.user."):
		return "project_user"
	case strings.HasPrefix(eventType, "project.control_cabinet."):
		return "control_cabinet"
	case strings.HasPrefix(eventType, "project.sps_controller."):
		return "sps_controller"
	case strings.HasPrefix(eventType, "project.field_device."):
		return "field_device"
	case strings.HasPrefix(eventType, "project.object_data."):
		return "object_data"
	default:
		return ""
	}
}

func resourceIDForProjectNotificationEvent(resourceType string, projectID uuid.UUID, entityIDs []string) *uuid.UUID {
	if resourceType == "project" {
		return &projectID
	}
	if len(entityIDs) != 1 {
		return nil
	}
	id, err := uuid.Parse(entityIDs[0])
	if err != nil {
		return nil
	}
	return &id
}

func (h *ProjectHandler) ensureProjectAccess(c *gin.Context, projectID uuid.UUID) bool {
	return projectshared.EnsureProjectAccess(c, h.access, projectID)
}
