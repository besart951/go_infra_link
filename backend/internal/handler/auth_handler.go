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
	authsvc "github.com/besart951/go_infra_link/backend/internal/service/auth"
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
	if !BindJSON(c, &req) {
		return
	}

	userAgent := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	result, err := h.service.Login(req.Email, req.Password, &userAgent, &ip)
	if err != nil {
		if err == domainAuth.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid_credentials"})
			return
		}
		if err == domainAuth.ErrAccountDisabled {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "account_disabled"})
			return
		}
		if err == domainAuth.ErrAccountLocked {
			c.JSON(http.StatusLocked, dto.ErrorResponse{Error: "account_locked"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "login_failed", Message: err.Error()})
		return
	}

	h.setAuthCookies(c, result)

	c.JSON(http.StatusOK, dto.AuthResponse{
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
	})
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

	userAgent := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	result, err := h.service.Login(h.devAuthEmail, h.devAuthPassword, &userAgent, &ip)
	if err != nil {
		if err == domainAuth.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid_credentials"})
			return
		}
		if err == domainAuth.ErrAccountDisabled {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "account_disabled"})
			return
		}
		if err == domainAuth.ErrAccountLocked {
			c.JSON(http.StatusLocked, dto.ErrorResponse{Error: "account_locked"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "login_failed", Message: err.Error()})
		return
	}

	h.setAuthCookies(c, result)

	c.JSON(http.StatusOK, dto.AuthResponse{
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
	})
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
		switch err {
		case domainAuth.ErrInvalidToken, domainAuth.ErrTokenExpired, domainAuth.ErrTokenRevoked:
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "refresh_failed", Message: err.Error()})
			return
		}
	}

	h.setAuthCookies(c, result)

	c.JSON(http.StatusOK, dto.AuthResponse{
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
	})
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
	if !BindJSON(c, &req) {
		return
	}

	if err := h.service.ConfirmPasswordReset(req.Token, req.NewPassword); err != nil {
		switch err {
		case domainAuth.ErrPasswordResetTokenInvalid:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "password_reset_token_invalid"})
			return
		case domainAuth.ErrPasswordResetTokenExpired:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "password_reset_token_expired"})
			return
		case domainAuth.ErrPasswordResetTokenUsed:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "password_reset_token_used"})
			return
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "reset_failed", Message: err.Error()})
			return
		}
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) setAuthCookies(c *gin.Context, result *authsvc.LoginResult) {
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
