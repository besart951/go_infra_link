package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/auth"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service         AuthService
	userService     UserService
	permissionSvc   PermissionQueryService
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	cookieSettings  CookieSettings
}

func NewAuthHandler(service AuthService, userService UserService, permissionSvc PermissionQueryService, accessTokenTTL, refreshTokenTTL time.Duration, cookieSettings CookieSettings) *AuthHandler {
	return &AuthHandler{
		service:         service,
		userService:     userService,
		permissionSvc:   permissionSvc,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		cookieSettings:  cookieSettings,
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

	result, err := h.service.Refresh(c.Request.Context(), refreshToken, &userAgent, &ip)
	if err != nil {
		h.handleRefreshError(c, err)
		return
	}

	h.setAuthCookies(c, result)

	c.JSON(http.StatusOK, h.buildAuthResponse(c.Request.Context(), result))
}

// handleRefreshError centralizes error handling for refresh operations
func (h *AuthHandler) handleRefreshError(c *gin.Context, err error) {
	handlerutil.RespondDomainError(
		c,
		err,
		handlerutil.LocalizedError(http.StatusInternalServerError, "refresh_failed", "auth.refresh_failed"),
		handlerutil.MapError(domainAuth.ErrInvalidToken, handlerutil.LocalizedError(http.StatusUnauthorized, "unauthorized", "errors.unauthorized")),
		handlerutil.MapError(domainAuth.ErrTokenExpired, handlerutil.LocalizedError(http.StatusUnauthorized, "unauthorized", "errors.unauthorized")),
		handlerutil.MapError(domainAuth.ErrTokenRevoked, handlerutil.LocalizedError(http.StatusUnauthorized, "unauthorized", "errors.unauthorized")),
	)
}

// Logout godoc
// @Summary Logout
// @Tags auth
// @Success 204
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil && strings.TrimSpace(refreshToken) != "" {
		if err := h.service.Logout(c.Request.Context(), refreshToken); err != nil {
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

// handleLogin is a helper method to handle login logic.
func (h *AuthHandler) handleLogin(c *gin.Context, email, password string) {
	ctx := c.Request.Context()
	userAgent := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	result, err := h.service.Login(ctx, email, password, &userAgent, &ip)
	if err != nil {
		h.handleLoginError(c, err)
		return
	}

	h.setAuthCookies(c, result)

	c.JSON(http.StatusOK, h.buildAuthResponse(ctx, result))
}

// handleLoginError centralizes error handling for login operations
func (h *AuthHandler) handleLoginError(c *gin.Context, err error) {
	handlerutil.RespondDomainError(
		c,
		err,
		handlerutil.LocalizedError(http.StatusInternalServerError, "login_failed", "auth.login_failed"),
		handlerutil.MapError(domainAuth.ErrInvalidCredentials, handlerutil.LocalizedError(http.StatusUnauthorized, "invalid_credentials", "auth.invalid_credentials")),
		handlerutil.MapError(domainAuth.ErrAccountDisabled, handlerutil.LocalizedError(http.StatusForbidden, "account_disabled", "auth.account_disabled")),
		handlerutil.MapError(domainAuth.ErrAccountLocked, handlerutil.LocalizedError(http.StatusLocked, "account_locked", "auth.account_locked")),
	)
}

// buildAuthResponse creates a consistent auth response
func (h *AuthHandler) buildAuthResponse(ctx context.Context, result *domainAuth.LoginResult) dto.AuthResponse {
	permissions := h.getRolePermissions(ctx, result.User.Role)
	return dto.AuthResponse{
		User: dto.AuthUserResponse{
			ID:                  result.User.ID,
			FirstName:           result.User.FirstName,
			LastName:            result.User.LastName,
			Email:               result.User.Email,
			IsActive:            result.User.IsActive,
			Role:                string(result.User.Role),
			Permissions:         permissions,
			CanAccessUserDirectory: h.canAccessUserDirectory(result.User.Role),
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

	usr, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "fetch_failed", "user.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusUnauthorized, "unauthorized", "errors.unauthorized")),
		)
		return
	}

	c.JSON(http.StatusOK, dto.AuthUserResponse{
		ID:                  usr.ID,
		FirstName:           usr.FirstName,
		LastName:            usr.LastName,
		Email:               usr.Email,
		IsActive:            usr.IsActive,
		Role:                string(usr.Role),
		Permissions:         h.getRolePermissions(c.Request.Context(), usr.Role),
		CanAccessUserDirectory: h.canAccessUserDirectory(usr.Role),
		CreatedAt:           usr.CreatedAt,
		UpdatedAt:           usr.UpdatedAt,
		LastLoginAt:         usr.LastLoginAt,
		DisabledAt:          usr.DisabledAt,
		LockedUntil:         usr.LockedUntil,
		FailedLoginAttempts: usr.FailedLoginAttempts,
	})
}

func (h *AuthHandler) getRolePermissions(ctx context.Context, role domainUser.Role) []string {
	if h.permissionSvc == nil {
		return []string{}
	}
	permissions, err := h.permissionSvc.GetRolePermissions(ctx, role)
	if err != nil {
		return []string{}
	}
	return permissions
}

func (h *AuthHandler) canAccessUserDirectory(role domainUser.Role) bool {
	if h.permissionSvc == nil {
		return false
	}
	return h.permissionSvc.CanAccessUserDirectory(role)
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
