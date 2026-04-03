package notification

import (
	"errors"
	"net/http"

	domain "github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/notification"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

type NotificationSettingsHandler struct {
	service NotificationSettingsService
}

func NewNotificationSettingsHandler(service NotificationSettingsService) *NotificationSettingsHandler {
	return &NotificationSettingsHandler{service: service}
}

// GetSMTPSettings godoc
// @Summary Get SMTP notification settings
// @Tags notifications
// @Produce json
// @Success 200 {object} dto.SMTPSettingsResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/notifications/smtp [get]
func (h *NotificationSettingsHandler) GetSMTPSettings(c *gin.Context) {
	settings, err := h.service.GetSMTPSettings()
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domainNotification.ErrProviderNotConfigured) {
			handlerutil.RespondError(c, http.StatusNotFound, "not_found", "SMTP settings not configured")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", "Failed to load SMTP settings")
		return
	}

	c.JSON(http.StatusOK, mapSMTPSettingsResponse(settings))
}

// UpsertSMTPSettings godoc
// @Summary Create or update SMTP notification settings
// @Tags notifications
// @Accept json
// @Produce json
// @Param payload body dto.UpsertSMTPSettingsRequest true "SMTP settings"
// @Success 200 {object} dto.SMTPSettingsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/notifications/smtp [put]
func (h *NotificationSettingsHandler) UpsertSMTPSettings(c *gin.Context) {
	var req dto.UpsertSMTPSettingsRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	settings, err := h.service.UpsertSMTPSettings(domainNotification.UpsertSMTPSettingsInput{
		ActorID:          userID,
		Enabled:          *req.Enabled,
		Host:             req.Host,
		Port:             req.Port,
		Username:         req.Username,
		Password:         req.Password,
		FromEmail:        req.FromEmail,
		FromName:         req.FromName,
		ReplyTo:          req.ReplyTo,
		Security:         domainNotification.SecurityMode(req.Security),
		AuthMode:         domainNotification.AuthMode(req.AuthMode),
		AllowInsecureTLS: *req.AllowInsecureTLS,
	})
	if err != nil {
		if ve, ok := domain.AsValidationError(err); ok {
			handlerutil.RespondValidationError(c, ve.Fields)
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "update_failed", "Failed to save SMTP settings")
		return
	}

	c.JSON(http.StatusOK, mapSMTPSettingsResponse(settings))
}

// SendSMTPTestEmail godoc
// @Summary Send an SMTP test email
// @Tags notifications
// @Accept json
// @Param payload body dto.SendSMTPTestEmailRequest true "Test email payload"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/notifications/smtp/test [post]
func (h *NotificationSettingsHandler) SendSMTPTestEmail(c *gin.Context) {
	var req dto.SendSMTPTestEmailRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	err := h.service.SendTestEmail(c.Request.Context(), domainNotification.SendTestEmailInput{
		To:      req.To,
		Subject: req.Subject,
		Body:    req.Body,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainNotification.ErrProviderNotConfigured), errors.Is(err, domain.ErrNotFound):
			handlerutil.RespondError(c, http.StatusNotFound, "not_found", "SMTP settings not configured")
			return
		default:
			if ve, ok := domain.AsValidationError(err); ok {
				handlerutil.RespondValidationError(c, ve.Fields)
				return
			}
			handlerutil.RespondError(c, http.StatusInternalServerError, "smtp_test_failed", err.Error())
			return
		}
	}

	c.Status(http.StatusNoContent)
}

func mapSMTPSettingsResponse(settings *domainNotification.SMTPSettings) dto.SMTPSettingsResponse {
	return dto.SMTPSettingsResponse{
		ID:               settings.ID,
		Provider:         string(settings.Provider),
		Enabled:          settings.Enabled,
		Host:             settings.Host,
		Port:             settings.Port,
		Username:         settings.Username,
		HasPassword:      settings.PasswordEncrypted != "",
		FromEmail:        settings.FromEmail,
		FromName:         settings.FromName,
		ReplyTo:          settings.ReplyTo,
		Security:         string(settings.Security),
		AuthMode:         string(settings.AuthMode),
		AllowInsecureTLS: settings.AllowInsecureTLS,
		UpdatedAt:        settings.UpdatedAt,
		UpdatedByID:      settings.UpdatedByID,
	}
}
