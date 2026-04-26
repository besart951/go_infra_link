package project

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
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
}

func NewProjectHandler(lifecycle ProjectLifecycleService, access ProjectAccessPolicyService, membership ProjectMembershipService, facilityLink ProjectFacilityLinkService) *ProjectHandler {
	return newProjectHandler(lifecycle, access, membership, newWorkflowFromServices(lifecycle, membership), facilityLink, NewProjectCollaborationHub())
}

func newProjectHandler(
	lifecycle ProjectLifecycleService,
	access ProjectAccessPolicyService,
	membership ProjectMembershipService,
	workflow ProjectWorkflowService,
	facilityLink ProjectFacilityLinkService,
	collaboration *ProjectCollaborationHub,
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

	h.collaboration.BroadcastControlCabinetDelta(projectID, actorID, toProjectCollaborationControlCabinet(controlCabinet))
}

func (h *ProjectHandler) notifyProjectSPSControllerDelta(c *gin.Context, projectID uuid.UUID, spsController domainFacility.SPSController) {
	if h.collaboration == nil {
		return
	}

	var actorID *uuid.UUID
	if userID, ok := middleware.GetUserID(c); ok {
		actorID = &userID
	}

	h.collaboration.BroadcastSPSControllerDelta(projectID, actorID, toProjectCollaborationSPSController(spsController))
}

func (h *ProjectHandler) ensureProjectAccess(c *gin.Context, projectID uuid.UUID) bool {
	return projectshared.EnsureProjectAccess(c, h.access, projectID)
}

func projectFieldDeviceDeltaPayload(fieldDevices []domainFacility.FieldDevice) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(fieldDevices))
	for _, item := range fieldDevices {
		apparatNr := item.ApparatNr
		systemPartID := item.SystemPartID
		items = append(items, map[string]interface{}{
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
