package auth

import (
	"errors"
	"net/http"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/auth"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	userregistrationservice "github.com/besart951/go_infra_link/backend/internal/service/userregistration"
	"github.com/gin-gonic/gin"
)

type RegistrationHandler struct {
	service RegistrationService
}

func NewRegistrationHandler(service RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{service: service}
}

func (h *RegistrationHandler) GetRegistration(c *gin.Context) {
	if h.service == nil {
		handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "errors.not_found")
		return
	}
	view, err := h.service.GetPublicRegistration(c.Request.Context(), c.Param("token"))
	if err != nil {
		handleRegistrationError(c, err)
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
		handleRegistrationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func handleRegistrationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainUser.ErrRegistrationTokenInvalid):
		handlerutil.RespondLocalizedError(c, http.StatusNotFound, "registration_invalid", "auth.registration_invalid")
	case errors.Is(err, domainUser.ErrRegistrationTokenExpired):
		handlerutil.RespondLocalizedError(c, http.StatusGone, "registration_expired", "auth.registration_expired")
	case errors.Is(err, domainUser.ErrRegistrationAlreadyAccepted):
		handlerutil.RespondLocalizedError(c, http.StatusConflict, "registration_already_accepted", "auth.registration_already_accepted")
	default:
		handlerutil.RespondDomainError(c, err, handlerutil.LocalizedError(http.StatusInternalServerError, "registration_failed", "auth.registration_failed"))
	}
}
