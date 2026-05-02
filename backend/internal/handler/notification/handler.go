package notification

import (
	"net/http"

	domain "github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/notification"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationSettingsHandler struct {
	service  NotificationSettingsService
	streamer SystemNotificationStreamer
}

func NewNotificationSettingsHandler(service NotificationSettingsService, streamer SystemNotificationStreamer) *NotificationSettingsHandler {
	return &NotificationSettingsHandler{service: service, streamer: streamer}
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
	settings, err := h.service.GetSMTPSettings(c.Request.Context())
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "fetch_failed", "Failed to load SMTP settings"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.PlainError(http.StatusNotFound, "not_found", "SMTP settings not configured")),
			handlerutil.MapError(domainNotification.ErrProviderNotConfigured, handlerutil.PlainError(http.StatusNotFound, "not_found", "SMTP settings not configured")),
		)
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

	settings, err := h.service.UpsertSMTPSettings(c.Request.Context(), domainNotification.UpsertSMTPSettingsInput{
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
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "update_failed", "Failed to save SMTP settings"),
		)
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
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "smtp_test_failed", err.Error()),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.PlainError(http.StatusNotFound, "not_found", "SMTP settings not configured")),
			handlerutil.MapError(domainNotification.ErrProviderNotConfigured, handlerutil.PlainError(http.StatusNotFound, "not_found", "SMTP settings not configured")),
		)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUserPreference godoc
// @Summary Get current user's notification preference
// @Tags notifications
// @Produce json
// @Success 200 {object} dto.UserNotificationPreferenceResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/account/notifications/preferences [get]
func (h *NotificationSettingsHandler) GetUserPreference(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	preference, err := h.service.GetUserPreference(c.Request.Context(), userID)
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "fetch_failed", "Failed to load notification preference"),
		)
		return
	}

	c.JSON(http.StatusOK, mapUserPreferenceResponse(preference))
}

// UpsertUserPreference godoc
// @Summary Create or update current user's notification preference
// @Tags notifications
// @Accept json
// @Produce json
// @Param payload body dto.UpsertUserNotificationPreferenceRequest true "Notification preference"
// @Success 200 {object} dto.UserNotificationPreferenceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/account/notifications/preferences [put]
func (h *NotificationSettingsHandler) UpsertUserPreference(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	var req dto.UpsertUserNotificationPreferenceRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	preference, err := h.service.UpsertUserPreference(c.Request.Context(), domainNotification.UpsertUserPreferenceInput{
		UserID:            userID,
		NotificationEmail: req.NotificationEmail,
		Channel:           domainNotification.DeliveryChannel(req.Channel),
		Frequency:         domainNotification.DeliveryFrequency(req.Frequency),
	})
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "update_failed", "Failed to save notification preference"),
		)
		return
	}

	c.JSON(http.StatusOK, mapUserPreferenceResponse(preference))
}

// SendUserPreferenceVerificationCode godoc
// @Summary Send an email verification code for current user's notification email
// @Tags notifications
// @Produce json
// @Success 200 {object} dto.UserNotificationPreferenceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/account/notifications/preferences/email-verification [post]
func (h *NotificationSettingsHandler) SendUserPreferenceVerificationCode(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	preference, err := h.service.SendUserPreferenceVerificationCode(c.Request.Context(), domainNotification.SendUserPreferenceVerificationCodeInput{
		UserID: userID,
	})
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "email_verification_failed", err.Error()),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.PlainError(http.StatusNotFound, "not_found", "SMTP settings not configured")),
			handlerutil.MapError(domainNotification.ErrProviderNotConfigured, handlerutil.PlainError(http.StatusNotFound, "not_found", "SMTP settings not configured")),
		)
		return
	}

	c.JSON(http.StatusOK, mapUserPreferenceResponse(preference))
}

// VerifyUserPreferenceEmail godoc
// @Summary Verify current user's notification email with a code
// @Tags notifications
// @Accept json
// @Produce json
// @Param payload body dto.VerifyUserNotificationEmailRequest true "Verification code"
// @Success 200 {object} dto.UserNotificationPreferenceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/account/notifications/preferences/email-verification/verify [post]
func (h *NotificationSettingsHandler) VerifyUserPreferenceEmail(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	var req dto.VerifyUserNotificationEmailRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	preference, err := h.service.VerifyUserPreferenceEmail(c.Request.Context(), domainNotification.VerifyUserPreferenceEmailInput{
		UserID: userID,
		Code:   req.Code,
	})
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "email_verification_failed", "Failed to verify notification email"),
		)
		return
	}

	c.JSON(http.StatusOK, mapUserPreferenceResponse(preference))
}

// ListSystemNotifications godoc
// @Summary List current user's system notifications
// @Tags notifications
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param unread_only query bool false "Only unread notifications"
// @Success 200 {object} dto.SystemNotificationListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/account/notifications [get]
func (h *NotificationSettingsHandler) ListSystemNotifications(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	var query struct {
		Page       int  `form:"page"`
		Limit      int  `form:"limit"`
		UnreadOnly bool `form:"unread_only"`
	}
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	result, err := h.service.ListSystemNotifications(c.Request.Context(), userID, query.Page, query.Limit, query.UnreadOnly)
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "fetch_failed", "Failed to load notifications"),
		)
		return
	}

	unreadCount, err := h.service.CountUnreadSystemNotifications(c.Request.Context(), userID)
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "fetch_failed", "Failed to load notification count"),
		)
		return
	}

	c.JSON(http.StatusOK, dto.SystemNotificationListResponse{
		Items:       mapSystemNotificationResponses(result.Items),
		Total:       result.Total,
		Page:        result.Page,
		TotalPages:  result.TotalPages,
		UnreadCount: unreadCount,
	})
}

func (h *NotificationSettingsHandler) StreamSystemNotifications(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}
	if h.streamer == nil {
		handlerutil.RespondError(c, http.StatusNotFound, "not_found", "Notification stream not configured")
		return
	}

	h.streamer.Stream(c.Writer, c.Request, userID)
}

// MarkSystemNotificationRead godoc
// @Summary Mark one system notification as read
// @Tags notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} dto.SystemNotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/account/notifications/{id}/read [post]
func (h *NotificationSettingsHandler) MarkSystemNotificationRead(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	notificationID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	notification, err := h.service.MarkSystemNotificationRead(c.Request.Context(), notificationID, userID)
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "update_failed", "Failed to mark notification read"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.PlainError(http.StatusNotFound, "not_found", "Notification not found")),
		)
		return
	}

	c.JSON(http.StatusOK, mapSystemNotificationResponse(notification))
}

// MarkAllSystemNotificationsRead godoc
// @Summary Mark all current user's system notifications as read
// @Tags notifications
// @Success 204
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/account/notifications/read-all [post]
func (h *NotificationSettingsHandler) MarkAllSystemNotificationsRead(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	if err := h.service.MarkAllSystemNotificationsRead(c.Request.Context(), userID); err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "update_failed", "Failed to mark notifications read"),
		)
		return
	}

	c.Status(http.StatusNoContent)
}

// ToggleSystemNotificationRead godoc
// @Summary Toggle read state for one system notification
// @Tags notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} dto.SystemNotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/account/notifications/{id}/read-toggle [post]
func (h *NotificationSettingsHandler) ToggleSystemNotificationRead(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	notificationID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	notification, err := h.service.ToggleSystemNotificationRead(c.Request.Context(), notificationID, userID)
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "update_failed", "Failed to update notification"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.PlainError(http.StatusNotFound, "not_found", "Notification not found")),
		)
		return
	}

	c.JSON(http.StatusOK, mapSystemNotificationResponse(notification))
}

// ToggleSystemNotificationImportant godoc
// @Summary Toggle important state for one system notification
// @Tags notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} dto.SystemNotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/account/notifications/{id}/important [post]
func (h *NotificationSettingsHandler) ToggleSystemNotificationImportant(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	notificationID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	notification, err := h.service.ToggleSystemNotificationImportant(c.Request.Context(), notificationID, userID)
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "update_failed", "Failed to update notification"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.PlainError(http.StatusNotFound, "not_found", "Notification not found")),
		)
		return
	}

	c.JSON(http.StatusOK, mapSystemNotificationResponse(notification))
}

// DeleteSystemNotification godoc
// @Summary Delete one system notification
// @Tags notifications
// @Param id path string true "Notification ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/account/notifications/{id} [delete]
func (h *NotificationSettingsHandler) DeleteSystemNotification(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	notificationID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteSystemNotification(c.Request.Context(), notificationID, userID); err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "delete_failed", "Failed to delete notification"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.PlainError(http.StatusNotFound, "not_found", "Notification not found")),
		)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *NotificationSettingsHandler) ListNotificationRules(c *gin.Context) {
	var query struct {
		EventKey string `form:"event_key"`
		Enabled  *bool  `form:"enabled"`
	}
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	rules, err := h.service.ListNotificationRules(c.Request.Context(), domainNotification.NotificationRuleFilter{
		EventKey: query.EventKey,
		Enabled:  query.Enabled,
	})
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "fetch_failed", "Failed to load notification rules"),
		)
		return
	}

	c.JSON(http.StatusOK, dto.NotificationRuleListResponse{Items: mapNotificationRuleResponses(rules)})
}

func (h *NotificationSettingsHandler) CreateNotificationRule(c *gin.Context) {
	h.upsertNotificationRule(c, false)
}

func (h *NotificationSettingsHandler) UpdateNotificationRule(c *gin.Context) {
	h.upsertNotificationRule(c, true)
}

func (h *NotificationSettingsHandler) upsertNotificationRule(c *gin.Context, existing bool) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	var req dto.UpsertNotificationRuleRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	var id uuid.UUID
	if existing {
		parsed, ok := handlerutil.ParseUUIDParam(c, "id")
		if !ok {
			return
		}
		id = parsed
	}

	rule, err := h.service.UpsertNotificationRule(c.Request.Context(), domainNotification.UpsertNotificationRuleInput{
		ID:               id,
		ActorID:          userID,
		Name:             req.Name,
		Enabled:          *req.Enabled,
		EventKey:         req.EventKey,
		ProjectID:        req.ProjectID,
		ResourceType:     req.ResourceType,
		ResourceID:       req.ResourceID,
		RecipientType:    domainNotification.RuleRecipientType(req.RecipientType),
		RecipientUserIDs: req.RecipientUserIDs,
		RecipientTeamID:  req.RecipientTeamID,
		RecipientRole:    domainUser.Role(req.RecipientRole),
	})
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "update_failed", "Failed to save notification rule"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.PlainError(http.StatusNotFound, "not_found", "Notification rule not found")),
		)
		return
	}

	c.JSON(http.StatusOK, mapNotificationRuleResponse(rule))
}

func (h *NotificationSettingsHandler) DeleteNotificationRule(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteNotificationRule(c.Request.Context(), id); err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.PlainError(http.StatusInternalServerError, "deletion_failed", "Failed to delete notification rule"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.PlainError(http.StatusNotFound, "not_found", "Notification rule not found")),
		)
		return
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

func mapUserPreferenceResponse(preference *domainNotification.UserPreference) dto.UserNotificationPreferenceResponse {
	return dto.UserNotificationPreferenceResponse{
		ID:                          preference.ID,
		UserID:                      preference.UserID,
		NotificationEmail:           preference.NotificationEmail,
		NotificationEmailVerifiedAt: preference.NotificationEmailVerifiedAt,
		EmailVerificationSentAt:     preference.EmailVerificationSentAt,
		EmailVerificationExpiresAt:  preference.EmailVerificationExpiresAt,
		Channel:                     string(preference.Channel),
		Frequency:                   string(preference.Frequency),
		CreatedAt:                   preference.CreatedAt,
		UpdatedAt:                   preference.UpdatedAt,
	}
}

func mapSystemNotificationResponses(notifications []domainNotification.SystemNotification) []dto.SystemNotificationResponse {
	result := make([]dto.SystemNotificationResponse, len(notifications))
	for i := range notifications {
		result[i] = mapSystemNotificationResponse(&notifications[i])
	}
	return result
}

func mapSystemNotificationResponse(notification *domainNotification.SystemNotification) dto.SystemNotificationResponse {
	return dto.SystemNotificationResponse{
		ID:           notification.ID,
		RecipientID:  notification.RecipientID,
		ActorID:      notification.ActorID,
		EventKey:     notification.EventKey,
		Title:        notification.Title,
		Body:         notification.Body,
		ResourceType: notification.ResourceType,
		ResourceID:   notification.ResourceID,
		Metadata:     notification.Metadata,
		ReadAt:       notification.ReadAt,
		IsImportant:  notification.IsImportant,
		CreatedAt:    notification.CreatedAt,
		UpdatedAt:    notification.UpdatedAt,
	}
}

func mapNotificationRuleResponses(rules []domainNotification.NotificationRule) []dto.NotificationRuleResponse {
	result := make([]dto.NotificationRuleResponse, len(rules))
	for i := range rules {
		result[i] = mapNotificationRuleResponse(&rules[i])
	}
	return result
}

func mapNotificationRuleResponse(rule *domainNotification.NotificationRule) dto.NotificationRuleResponse {
	return dto.NotificationRuleResponse{
		ID:               rule.ID,
		Name:             rule.Name,
		Enabled:          rule.Enabled,
		EventKey:         rule.EventKey,
		ProjectID:        rule.ProjectID,
		ResourceType:     rule.ResourceType,
		ResourceID:       rule.ResourceID,
		RecipientType:    string(rule.RecipientType),
		RecipientUserIDs: rule.RecipientUserIDs,
		RecipientTeamID:  rule.RecipientTeamID,
		RecipientRole:    string(rule.RecipientRole),
		CreatedByID:      rule.CreatedByID,
		CreatedAt:        rule.CreatedAt,
		UpdatedAt:        rule.UpdatedAt,
	}
}
