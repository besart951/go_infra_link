package handler

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/app/object"
	"github.com/gin-gonic/gin"
)

type ObjectHandler struct {
	service *object.Service
}

func NewObjectHandler(service *object.Service) *ObjectHandler {
	return &ObjectHandler{service: service}
}

type createObjectRequest struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
}

func (h *ObjectHandler) Create(c *gin.Context) {
	var req createObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if req.ProjectID == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id and name are required"})
		return
	}

	objectID, err := h.service.CreateObject(c.Request.Context(), req.ProjectID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": objectID})
}

func (h *ObjectHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	o, err := h.service.GetObject(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if o == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, o)
}

type grantAccessRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func (h *ObjectHandler) GrantAccess(c *gin.Context) {
	objectID := c.Param("id")
	if objectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "object id is required"})
		return
	}

	var req grantAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if req.UserID == "" || req.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and role are required"})
		return
	}

	if err := h.service.GrantAccess(c.Request.Context(), objectID, req.UserID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}
