package handler

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/app/project"
	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	service *project.Service
}

func NewProjectHandler(service *project.Service) *ProjectHandler {
	return &ProjectHandler{service: service}
}

type createProjectRequest struct {
	Name    string `json:"name"`
	OwnerID string `json:"owner_id"`
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if req.Name == "" || req.OwnerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and owner_id are required"})
		return
	}

	projectID, err := h.service.CreateProject(c.Request.Context(), req.Name, req.OwnerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": projectID})
}

func (h *ProjectHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	p, err := h.service.GetProject(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if p == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, p)
}

type addMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func (h *ProjectHandler) AddMember(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project id is required"})
		return
	}

	var req addMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if req.UserID == "" || req.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and role are required"})
		return
	}

	if err := h.service.AddMember(c.Request.Context(), projectID, req.UserID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}
