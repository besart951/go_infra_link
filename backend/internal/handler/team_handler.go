package handler

import (
	"net/http"

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

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req dto.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	t := &team.Team{Name: req.Name, Description: req.Description}
	if err := h.service.Create(t); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "creation_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.TeamResponse{ID: t.ID, Name: t.Name, Description: t.Description, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt})
}

func (h *TeamHandler) ListTeams(c *gin.Context) {
	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	res, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "fetch_failed", Message: err.Error()})
		return
	}

	items := make([]dto.TeamResponse, len(res.Items))
	for i, t := range res.Items {
		items[i] = dto.TeamResponse{ID: t.ID, Name: t.Name, Description: t.Description, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt}
	}

	c.JSON(http.StatusOK, dto.TeamListResponse{Items: items, Total: res.Total, Page: res.Page, TotalPages: res.TotalPages})
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	t, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "fetch_failed", Message: err.Error()})
		return
	}
	if t == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "Team not found"})
		return
	}

	c.JSON(http.StatusOK, dto.TeamResponse{ID: t.ID, Name: t.Name, Description: t.Description, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt})
}

func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	var req dto.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	t, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "fetch_failed", Message: err.Error()})
		return
	}
	if t == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "Team not found"})
		return
	}

	if req.Name != "" {
		t.Name = req.Name
	}
	if req.Description != nil {
		t.Description = req.Description
	}

	if err := h.service.Update(t); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "update_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.TeamResponse{ID: t.ID, Name: t.Name, Description: t.Description, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt})
}

func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	if err := h.service.DeleteByIds([]uuid.UUID{id}); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "deletion_failed", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TeamHandler) AddMember(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	var req dto.AddTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	if err := h.service.AddMember(teamID, req.UserID, team.MemberRole(req.Role)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "update_failed", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_user_id", Message: "Invalid UUID format"})
		return
	}

	if err := h.service.RemoveMember(teamID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "update_failed", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TeamHandler) ListMembers(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	res, err := h.service.ListMembers(teamID, query.Page, query.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "fetch_failed", Message: err.Error()})
		return
	}

	items := make([]dto.TeamMemberResponse, len(res.Items))
	for i, m := range res.Items {
		items[i] = dto.TeamMemberResponse{TeamID: m.TeamID, UserID: m.UserID, Role: string(m.Role), JoinedAt: m.JoinedAt}
	}

	c.JSON(http.StatusOK, dto.TeamMemberListResponse{Items: items, Total: res.Total, Page: res.Page, TotalPages: res.TotalPages})
}
