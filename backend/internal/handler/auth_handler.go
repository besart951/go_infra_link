package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
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
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	cookieSettings  CookieSettings
	devAuthEnabled  bool
	devAuthEmail    string
	devAuthPassword string
}

func NewAuthHandler(service AuthService, userService UserService, accessTokenTTL, refreshTokenTTL time.Duration, cookieSettings CookieSettings, devAuthEnabled bool, devAuthEmail, devAuthPassword string) *AuthHandler {
	return &AuthHandler{
		service:         service,
		userService:     userService,
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
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found"})
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
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
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
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "refresh_failed", Message: err.Error()})
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
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid_credentials"})
	case domainAuth.ErrAccountDisabled:
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "account_disabled"})
	case domainAuth.ErrAccountLocked:
		c.JSON(http.StatusLocked, dto.ErrorResponse{Error: "account_locked"})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "login_failed", Message: err.Error()})
	}
}

// buildAuthResponse creates a consistent auth response
func (h *AuthHandler) buildAuthResponse(result *domainAuth.LoginResult) dto.AuthResponse {
	return dto.AuthResponse{
		User: dto.AuthUserResponse{
			ID:        result.User.ID,
			FirstName: result.User.FirstName,
			LastName:  result.User.LastName,
			Email:     result.User.Email,
			IsActive:  result.User.IsActive,
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
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	usr, err := h.userService.GetByID(userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "fetch_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.AuthUserResponse{
		ID:        usr.ID,
		FirstName: usr.FirstName,
		LastName:  usr.LastName,
		Email:     usr.Email,
		IsActive:  usr.IsActive,
	})
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
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "password_reset_token_invalid"})
	case domainAuth.ErrPasswordResetTokenExpired:
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "password_reset_token_expired"})
	case domainAuth.ErrPasswordResetTokenUsed:
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "password_reset_token_used"})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "reset_failed", Message: err.Error()})
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
