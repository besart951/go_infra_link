package handler

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	if !BindJSON(c, &req) {
		return
	}

	t := &team.Team{Name: req.Name, Description: req.Description}
	if err := h.service.Create(t); err != nil {
		RespondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	c.JSON(http.StatusCreated, dto.TeamResponse{ID: t.ID, Name: t.Name, Description: t.Description, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt})
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
	if !BindQuery(c, &query) {
		return
	}

	res, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	items := make([]dto.TeamResponse, len(res.Items))
	for i, t := range res.Items {
		items[i] = dto.TeamResponse{ID: t.ID, Name: t.Name, Description: t.Description, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt}
	}

	c.JSON(http.StatusOK, dto.TeamListResponse{Items: items, Total: res.Total, Page: res.Page, TotalPages: res.TotalPages})
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
	id, ok := ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	t, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			RespondNotFound(c, "Team not found")
			return
		}
		RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.TeamResponse{ID: t.ID, Name: t.Name, Description: t.Description, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt})
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
	id, ok := ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateTeamRequest
	if !BindJSON(c, &req) {
		return
	}

	t, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			RespondNotFound(c, "Team not found")
			return
		}
		RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	if req.Name != "" {
		t.Name = req.Name
	}
	if req.Description != nil {
		t.Description = req.Description
	}

	if err := h.service.Update(t); err != nil {
		RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.TeamResponse{ID: t.ID, Name: t.Name, Description: t.Description, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt})
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
	id, ok := ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByIds([]uuid.UUID{id}); err != nil {
		RespondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
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
	teamID, ok := ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.AddTeamMemberRequest
	if !BindJSON(c, &req) {
		return
	}

	if err := h.service.AddMember(teamID, req.UserID, team.MemberRole(req.Role)); err != nil {
		RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
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
	teamID, ok := ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	userID, ok := ParseUUIDParamWithCode(c, "userId", "invalid_user_id")
	if !ok {
		return
	}

	if err := h.service.RemoveMember(teamID, userID); err != nil {
		RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
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
	teamID, ok := ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var query dto.PaginationQuery
	if !BindQuery(c, &query) {
		return
	}

	res, err := h.service.ListMembers(teamID, query.Page, query.Limit)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	items := make([]dto.TeamMemberResponse, len(res.Items))
	for i, m := range res.Items {
		items[i] = dto.TeamMemberResponse{TeamID: m.TeamID, UserID: m.UserID, Role: string(m.Role), JoinedAt: m.JoinedAt}
	}

	c.JSON(http.StatusOK, dto.TeamMemberListResponse{Items: items, Total: res.Total, Page: res.Page, TotalPages: res.TotalPages})
}
