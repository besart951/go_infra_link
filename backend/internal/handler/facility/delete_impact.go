package facility

import (
	"net/http"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

type DeleteImpactHandler struct {
	service DeleteImpactService
	authz   middleware.AuthorizationChecker
}

func NewDeleteImpactHandler(service DeleteImpactService) *DeleteImpactHandler {
	return &DeleteImpactHandler{service: service}
}

func (h *DeleteImpactHandler) SetAuthorizationChecker(authz middleware.AuthorizationChecker) {
	h.authz = authz
}

// GetDeleteImpacts godoc
// @Summary Preview blocking facility references before deletion
// @Tags facility-delete-impacts
// @Produce json
// @Param resource query string true "Reference resource (apparat or system_part)"
// @Param ids query []string true "Reference IDs"
// @Success 200 {object} dto.DeleteImpactListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/delete-impacts [get]
func (h *DeleteImpactHandler) GetDeleteImpacts(c *gin.Context) {
	resource, ok := domainFacility.ParseDeleteImpactResource(c.Query("resource"))
	if !ok {
		respondLocalizedInvalidArgument(c, "facility.invalid_delete_impact_resource")
		return
	}
	if !h.canRead(c, resource) {
		return
	}
	ids, ok := parseUUIDListQueryParam(c, "ids")
	if !ok {
		return
	}

	impacts, err := h.service.List(c.Request.Context(), resource, ids)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}
	items := make([]dto.DeleteImpactResponse, 0, len(impacts))
	for _, impact := range impacts {
		blockers := make([]dto.DeleteImpactBlockerResponse, 0, len(impact.Blockers))
		for _, blocker := range impact.Blockers {
			blockers = append(blockers, dto.DeleteImpactBlockerResponse{Resource: blocker.Resource, Count: blocker.Count})
		}
		items = append(items, dto.DeleteImpactResponse{Resource: string(impact.Resource), ID: impact.ID, Blockers: blockers})
	}
	c.JSON(http.StatusOK, dto.DeleteImpactListResponse{Items: items})
}

func (h *DeleteImpactHandler) canRead(c *gin.Context, resource domainFacility.DeleteImpactResource) bool {
	if h.authz == nil {
		return true
	}
	role, ok := middleware.GetUserRole(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return false
	}
	permission := domainUser.PermissionApparatRead
	if resource == domainFacility.DeleteImpactResourceSystemPart {
		permission = domainUser.PermissionSystemPartRead
	}
	allowed, err := h.authz.HasPermission(c.Request.Context(), role, permission)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
		return false
	}
	if !allowed {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return false
	}
	return true
}
