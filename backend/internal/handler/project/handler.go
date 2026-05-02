package project

import (
	"context"
	"strconv"
	"strings"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
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
	collaboration *ProjectCollaborationHub
	notifications NotificationEventDispatcher
}

func NewProjectHandler(lifecycle ProjectLifecycleService, access ProjectAccessPolicyService, membership ProjectMembershipService, facilityLink ProjectFacilityLinkService) *ProjectHandler {
	return newProjectHandler(lifecycle, access, membership, newWorkflowFromServices(lifecycle, membership), facilityLink, NewProjectCollaborationHub(), nil)
}

func newProjectHandler(
	lifecycle ProjectLifecycleService,
	access ProjectAccessPolicyService,
	membership ProjectMembershipService,
	workflow ProjectWorkflowService,
	facilityLink ProjectFacilityLinkService,
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

func (h *ProjectHandler) notifyProjectFieldDeviceDelta(c *gin.Context, projectID uuid.UUID, fieldDevices []domainFacility.FieldDevice) {
	if h.collaboration == nil || len(fieldDevices) == 0 {
		return
	}

	var actorID *uuid.UUID
	if userID, ok := middleware.GetUserID(c); ok {
		actorID = &userID
	}

	h.collaboration.BroadcastFieldDeviceDelta(projectID, actorID, projectFieldDeviceDeltaPayload(fieldDevices))
}

func (h *ProjectHandler) notifyProjectControlCabinetDelta(c *gin.Context, projectID uuid.UUID, controlCabinet domainFacility.ControlCabinet) {
	if h.collaboration == nil {
		return
	}

	var actorID *uuid.UUID
	if userID, ok := middleware.GetUserID(c); ok {
		actorID = &userID
	}

	h.collaboration.BroadcastControlCabinetDelta(projectID, actorID, controlCabinet)
}

func (h *ProjectHandler) notifyProjectSPSControllerDelta(c *gin.Context, projectID uuid.UUID, spsController domainFacility.SPSController) {
	if h.collaboration == nil {
		return
	}

	var actorID *uuid.UUID
	if userID, ok := middleware.GetUserID(c); ok {
		actorID = &userID
	}

	h.collaboration.BroadcastSPSControllerDelta(projectID, actorID, spsController)
}

func (h *ProjectHandler) ensureProjectAccess(c *gin.Context, projectID uuid.UUID) bool {
	return projectshared.EnsureProjectAccess(c, h.access, projectID)
}

func projectFieldDeviceDeltaPayload(fieldDevices []domainFacility.FieldDevice) []map[string]any {
	items := make([]map[string]any, 0, len(fieldDevices))
	for _, item := range fieldDevices {
		apparatNr := item.ApparatNr
		systemPartID := item.SystemPartID
		items = append(items, map[string]any{
			"id":                            item.ID,
			"bmk":                           item.BMK,
			"description":                   item.Description,
			"text_fix":                      item.TextIndividuell,
			"apparat_nr":                    &apparatNr,
			"sps_controller_system_type_id": item.SPSControllerSystemTypeID,
			"system_part_id":                &systemPartID,
			"specification_id":              item.SpecificationID,
			"apparat_id":                    item.ApparatID,
			"created_at":                    item.CreatedAt,
			"updated_at":                    item.UpdatedAt,
		})
	}
	return items
}
