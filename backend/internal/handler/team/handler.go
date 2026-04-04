package team

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/team"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/team"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	service TeamService
}

func NewTeamHandler(service TeamService) *TeamHandler {
	return &TeamHandler{service: service}
}

// CreateTeam godoc
// @Summary Create a new team
// @Tags teams
// @Accept json
// @Produce json
// @Param team body dto.CreateTeamRequest true "Team data"
// @Success 201 {object} dto.TeamResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/teams [post]
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req dto.CreateTeamRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	t := ToTeamModel(req)
	if err := h.service.Create(c.Request.Context(), t); err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "team.creation_failed")
		return
	}

	c.JSON(http.StatusCreated, ToTeamResponse(t))
}

// ListTeams godoc
// @Summary List teams with pagination
// @Tags teams
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.TeamListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/teams [get]
func (h *TeamHandler) ListTeams(c *gin.Context) {
	var query dto.PaginationQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	res, err := h.service.List(c.Request.Context(), query.Page, query.Limit, query.Search)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "team.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, dto.TeamListResponse{
		Items:      ToTeamListResponse(res.Items),
		Total:      res.Total,
		Page:       res.Page,
		TotalPages: res.TotalPages,
	})
}

// GetTeam godoc
// @Summary Get a team by ID
// @Tags teams
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} dto.TeamResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/teams/{id} [get]
func (h *TeamHandler) GetTeam(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	t, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "fetch_failed", "team.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "team.team_not_found")),
		)
		return
	}

	c.JSON(http.StatusOK, ToTeamResponse(t))
}

// UpdateTeam godoc
// @Summary Update a team
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param team body dto.UpdateTeamRequest true "Team data"
// @Success 200 {object} dto.TeamResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/teams/{id} [put]
func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateTeamRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()
	t, err := h.service.GetByID(ctx, id)
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "fetch_failed", "team.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "team.team_not_found")),
		)
		return
	}

	ApplyTeamUpdate(t, req)

	if err := h.service.Update(ctx, t); err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "update_failed", "team.update_failed")
		return
	}

	c.JSON(http.StatusOK, ToTeamResponse(t))
}

// DeleteTeam godoc
// @Summary Delete a team
// @Tags teams
// @Param id path string true "Team ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/teams/{id} [delete]
func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByID(c.Request.Context(), id); err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "team.deletion_failed")
		return
	}

	c.Status(http.StatusNoContent)
}

// AddMember godoc
// @Summary Add a member to a team
// @Tags teams
// @Accept json
// @Param id path string true "Team ID"
// @Param payload body dto.AddTeamMemberRequest true "Member data"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/teams/{id}/members [post]
func (h *TeamHandler) AddMember(c *gin.Context) {
	teamID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.AddTeamMemberRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	if err := h.service.AddMember(c.Request.Context(), teamID, req.UserID, team.MemberRole(req.Role)); err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "update_failed", "team.update_failed")
		return
	}

	c.Status(http.StatusNoContent)
}

// RemoveMember godoc
// @Summary Remove a member from a team
// @Tags teams
// @Param id path string true "Team ID"
// @Param userId path string true "User ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/teams/{id}/members/{userId} [delete]
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	userID, ok := handlerutil.ParseUUIDParamWithCode(c, "userId", "invalid_user_id")
	if !ok {
		return
	}

	if err := h.service.RemoveMember(c.Request.Context(), teamID, userID); err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "update_failed", "team.update_failed")
		return
	}

	c.Status(http.StatusNoContent)
}

// ListMembers godoc
// @Summary List team members
// @Tags teams
// @Produce json
// @Param id path string true "Team ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.TeamMemberListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/teams/{id}/members [get]
func (h *TeamHandler) ListMembers(c *gin.Context) {
	teamID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var query dto.PaginationQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	res, err := h.service.ListMembers(c.Request.Context(), teamID, query.Page, query.Limit)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "team.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, dto.TeamMemberListResponse{
		Items:      ToTeamMemberListResponse(res.Items),
		Total:      res.Total,
		Page:       res.Page,
		TotalPages: res.TotalPages,
	})
}
