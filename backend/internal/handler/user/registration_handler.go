package user

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	userregistration "github.com/besart951/go_infra_link/backend/internal/service/userregistration"
	"github.com/gin-gonic/gin"
)

type RegistrationHandler struct {
	service UserRegistrationService
}

func NewRegistrationHandler(service UserRegistrationService) *RegistrationHandler {
	return &RegistrationHandler{service: service}
}

func (h *RegistrationHandler) CreateInvitation(c *gin.Context) {
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	var req dto.CreateUserInvitationRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}
	usr, process, err := h.service.CreateInvitation(c.Request.Context(), userregistration.InviteInput{
		ActorID: actorID,
		Email:   req.Email,
		Role:    domainUser.Role(req.Role),
	})
	if err != nil {
		respondRegistrationAdminError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.CreateUserInvitationResponse{
		User:                ToUserResponse(usr),
		RegistrationProcess: *ToRegistrationProcessResponse(process),
	})
}

func (h *RegistrationHandler) GetProcess(c *gin.Context) {
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	userID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	process, err := h.service.GetProcess(c.Request.Context(), actorID, userID)
	if err != nil {
		respondRegistrationAdminError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToRegistrationProcessResponse(process))
}

func (h *RegistrationHandler) ResendInvitation(c *gin.Context) {
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	userID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	process, err := h.service.ResendInvitation(c.Request.Context(), actorID, userID)
	if err != nil {
		respondRegistrationAdminError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToRegistrationProcessResponse(process))
}

func respondRegistrationAdminError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainUser.ErrRoleNotAssignable):
		handlerutil.RespondLocalizedError(c, http.StatusForbidden, "role_not_assignable", "user.role_not_assignable")
	case errors.Is(err, domainUser.ErrRegistrationAlreadyAccepted):
		handlerutil.RespondLocalizedError(c, http.StatusConflict, "registration_already_accepted", "auth.registration_already_accepted")
	case errors.Is(err, domainUser.ErrRegistrationResendTooSoon):
		handlerutil.RespondLocalizedError(c, http.StatusTooManyRequests, "registration_resend_too_soon", "auth.registration_resend_too_soon")
	case errors.Is(err, domain.ErrConflict):
		handlerutil.RespondLocalizedError(c, http.StatusConflict, "email_already_registered", "auth.email_already_registered")
	default:
		handlerutil.RespondDomainError(c, err, handlerutil.LocalizedError(http.StatusInternalServerError, "registration_failed", "auth.registration_failed"))
	}
}
