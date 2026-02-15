package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

type CookieSettings struct {
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

type AuthHandler struct {
	service         AuthService
	userService     UserService
	permissionSvc   PermissionQueryService
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	cookieSettings  CookieSettings
	devAuthEnabled  bool
	devAuthEmail    string
	devAuthPassword string
}

func NewAuthHandler(service AuthService, userService UserService, permissionSvc PermissionQueryService, accessTokenTTL, refreshTokenTTL time.Duration, cookieSettings CookieSettings, devAuthEnabled bool, devAuthEmail, devAuthPassword string) *AuthHandler {
	return &AuthHandler{
		service:         service,
		userService:     userService,
		permissionSvc:   permissionSvc,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		cookieSettings:  cookieSettings,
		devAuthEnabled:  devAuthEnabled,
		devAuthEmail:    devAuthEmail,
		devAuthPassword: devAuthPassword,
	}
}

// Login godoc
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param login body dto.LoginRequest true "Login data"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	h.handleLogin(c, req.Email, req.Password)
}

// DevLogin godoc
// @Summary Dev login (no credentials)
// @Tags auth
// @Produce json
// @Success 200 {object} dto.AuthResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/dev-login [post]
func (h *AuthHandler) DevLogin(c *gin.Context) {
	if !h.devAuthEnabled || h.devAuthEmail == "" || h.devAuthPassword == "" {
		handlerutil.RespondLocalizedNotFound(c)
		return
	}

	h.handleLogin(c, h.devAuthEmail, h.devAuthPassword)
}

// Refresh godoc
// @Summary Refresh access token
// @Tags auth
// @Produce json
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || strings.TrimSpace(refreshToken) == "" {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}

	userAgent := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	result, err := h.service.Refresh(refreshToken, &userAgent, &ip)
	if err != nil {
		h.handleRefreshError(c, err)
		return
	}

	h.setAuthCookies(c, result)

	c.JSON(http.StatusOK, h.buildAuthResponse(result))
}

// handleRefreshError centralizes error handling for refresh operations
func (h *AuthHandler) handleRefreshError(c *gin.Context, err error) {
	switch err {
	case domainAuth.ErrInvalidToken, domainAuth.ErrTokenExpired, domainAuth.ErrTokenRevoked:
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
	default:
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "refresh_failed", "auth.refresh_failed")
	}
}

// Logout godoc
// @Summary Logout
// @Tags auth
// @Success 204
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil && strings.TrimSpace(refreshToken) != "" {
		if err := h.service.Logout(refreshToken); err != nil {
			// Logout is intentionally best-effort and idempotent.
			ginErr := c.Error(err)
			if ginErr != nil {
				ginErr.Type = gin.ErrorTypePrivate
			}
		}
	}

	h.clearAuthCookies(c)
	c.Status(http.StatusNoContent)
}

// handleLogin is a helper method to handle common login logic
// This eliminates duplicate code between Login and DevLogin handlers
func (h *AuthHandler) handleLogin(c *gin.Context, email, password string) {
	userAgent := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	result, err := h.service.Login(email, password, &userAgent, &ip)
	if err != nil {
		h.handleLoginError(c, err)
		return
	}

	h.setAuthCookies(c, result)

	c.JSON(http.StatusOK, h.buildAuthResponse(result))
}

// handleLoginError centralizes error handling for login operations
func (h *AuthHandler) handleLoginError(c *gin.Context, err error) {
	switch err {
	case domainAuth.ErrInvalidCredentials:
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "invalid_credentials", "auth.invalid_credentials")
	case domainAuth.ErrAccountDisabled:
		handlerutil.RespondLocalizedError(c, http.StatusForbidden, "account_disabled", "auth.account_disabled")
	case domainAuth.ErrAccountLocked:
		handlerutil.RespondLocalizedError(c, http.StatusLocked, "account_locked", "auth.account_locked")
	default:
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "login_failed", "auth.login_failed")
	}
}

// buildAuthResponse creates a consistent auth response
func (h *AuthHandler) buildAuthResponse(result *domainAuth.LoginResult) dto.AuthResponse {
	permissions := h.getRolePermissions(result.User.Role)
	return dto.AuthResponse{
		User: dto.AuthUserResponse{
			ID:                  result.User.ID,
			FirstName:           result.User.FirstName,
			LastName:            result.User.LastName,
			Email:               result.User.Email,
			IsActive:            result.User.IsActive,
			Role:                string(result.User.Role),
			Permissions:         permissions,
			CreatedAt:           result.User.CreatedAt,
			UpdatedAt:           result.User.UpdatedAt,
			LastLoginAt:         result.User.LastLoginAt,
			DisabledAt:          result.User.DisabledAt,
			LockedUntil:         result.User.LockedUntil,
			FailedLoginAttempts: result.User.FailedLoginAttempts,
		},
		AccessTokenExpiresAt:  result.AccessTokenExpiry,
		RefreshTokenExpiresAt: result.RefreshTokenExpiry,
		CsrfToken:             result.CSRFFriendlyToken,
	}
}

// Me godoc
// @Summary Get current user
// @Tags auth
// @Produce json
// @Success 200 {object} dto.AuthUserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}

	usr, err := h.userService.GetByID(userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
			return
		}
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "user.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, dto.AuthUserResponse{
		ID:                  usr.ID,
		FirstName:           usr.FirstName,
		LastName:            usr.LastName,
		Email:               usr.Email,
		IsActive:            usr.IsActive,
		Role:                string(usr.Role),
		Permissions:         h.getRolePermissions(usr.Role),
		CreatedAt:           usr.CreatedAt,
		UpdatedAt:           usr.UpdatedAt,
		LastLoginAt:         usr.LastLoginAt,
		DisabledAt:          usr.DisabledAt,
		LockedUntil:         usr.LockedUntil,
		FailedLoginAttempts: usr.FailedLoginAttempts,
	})
}

func (h *AuthHandler) getRolePermissions(role domainUser.Role) []string {
	if h.permissionSvc == nil {
		return []string{}
	}
	permissions, err := h.permissionSvc.GetRolePermissions(role)
	if err != nil {
		return []string{}
	}
	return permissions
}

// ConfirmPasswordReset godoc
// @Summary Confirm password reset
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body dto.PasswordResetConfirmRequest true "Password reset confirmation"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/password-reset/confirm [post]
func (h *AuthHandler) ConfirmPasswordReset(c *gin.Context) {
	var req dto.PasswordResetConfirmRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	if err := h.service.ConfirmPasswordReset(req.Token, req.NewPassword); err != nil {
		h.handlePasswordResetError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// handlePasswordResetError centralizes error handling for password reset operations
func (h *AuthHandler) handlePasswordResetError(c *gin.Context, err error) {
	switch err {
	case domainAuth.ErrPasswordResetTokenInvalid:
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "password_reset_token_invalid", "auth.password_reset_token_invalid")
	case domainAuth.ErrPasswordResetTokenExpired:
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "password_reset_token_expired", "auth.password_reset_token_expired")
	case domainAuth.ErrPasswordResetTokenUsed:
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "password_reset_token_used", "auth.password_reset_token_used")
	default:
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "reset_failed", "auth.reset_failed")
	}
}

func (h *AuthHandler) setAuthCookies(c *gin.Context, result *domainAuth.LoginResult) {
	setCookie(c, "access_token", result.AccessToken, h.accessTokenTTL, h.cookieSettings)
	setCookie(c, "refresh_token", result.RefreshToken, h.refreshTokenTTL, h.cookieSettings)
	setCSRFCookie(c, result.CSRFFriendlyToken, h.cookieSettings)
}

func (h *AuthHandler) clearAuthCookies(c *gin.Context) {
	clearCookie(c, "access_token", h.cookieSettings)
	clearCookie(c, "refresh_token", h.cookieSettings)
	clearCookie(c, "csrf_token", h.cookieSettings)
}

func setCookie(c *gin.Context, name, value string, ttl time.Duration, settings CookieSettings) {
	maxAge := int(ttl.Seconds())
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   settings.Domain,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   settings.Secure,
		SameSite: settings.SameSite,
	}
	http.SetCookie(c.Writer, cookie)
}

func setCSRFCookie(c *gin.Context, value string, settings CookieSettings) {
	cookie := &http.Cookie{
		Name:     "csrf_token",
		Value:    value,
		Path:     "/",
		Domain:   settings.Domain,
		MaxAge:   int((24 * time.Hour).Seconds()),
		HttpOnly: false,
		Secure:   settings.Secure,
		SameSite: settings.SameSite,
	}
	http.SetCookie(c.Writer, cookie)
}

func clearCookie(c *gin.Context, name string, settings CookieSettings) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Domain:   settings.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   settings.Secure,
		SameSite: settings.SameSite,
	}
	if name == "csrf_token" {
		cookie.HttpOnly = false
	}
	http.SetCookie(c.Writer, cookie)
}
