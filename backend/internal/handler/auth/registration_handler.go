package auth

import (
	"net/http"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/auth"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	userregistrationservice "github.com/besart951/go_infra_link/backend/internal/service/userregistration"
	"github.com/gin-gonic/gin"
)

type RegistrationHandler struct {
	service        RegistrationService
	errorResponder *handlerutil.ErrorResponder
}

func NewRegistrationHandler(service RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{
		service:        service,
		errorResponder: handlerutil.NewErrorResponder(),
	}
}

func (h *RegistrationHandler) GetRegistration(c *gin.Context) {
	if h.service == nil {
		handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "errors.not_found")
		return
	}
	view, err := h.service.GetPublicRegistration(c.Request.Context(), c.Param("token"))
	if err != nil {
		h.errorResponder.RespondRegistrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.PublicRegistrationResponse{
		Email:           view.Email,
		Role:            string(view.Role),
		RoleDisplayName: domainUser.RoleDisplayName(view.Role),
		ExpiresAt:       view.ExpiresAt,
	})
}

func (h *RegistrationHandler) CompleteRegistration(c *gin.Context) {
	if h.service == nil {
		handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "errors.not_found")
		return
	}
	var req dto.CompleteRegistrationRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}
	_, err := h.service.CompleteRegistration(c.Request.Context(), userregistrationservice.CompleteInput{
		Token:      c.Param("token"),
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Password:   req.Password,
		PrivacyAck: req.PrivacyAck,
	})
	if err != nil {
		h.errorResponder.RespondRegistrationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
