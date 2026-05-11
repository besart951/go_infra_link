package handlerutil

import (
	"net/http"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/gin-gonic/gin"
)

// ErrorResponder provides a single source of truth for mapping domain errors to HTTP responses.
type ErrorResponder struct {
	// Reusable error mappings for user domain
	userMappings []ErrorMapping
	// Reusable error mappings for registration domain
	registrationMappings []ErrorMapping
}

// NewErrorResponder creates a new error responder with standard mappings.
func NewErrorResponder() *ErrorResponder {
	return &ErrorResponder{
		userMappings: []ErrorMapping{
			MapError(domainUser.ErrRoleNotAssignable, LocalizedError(http.StatusForbidden, "role_not_assignable", "user.role_not_assignable")),
			MapError(domainUser.ErrDeletedUserRestorable, LocalizedError(http.StatusConflict, "deleted_user_restorable", "user.deleted_user_restorable")),
			MapError(domainUser.ErrRestoreWindowExpired, LocalizedError(http.StatusConflict, "restore_window_expired", "user.restore_window_expired")),
			MapError(domainUser.ErrUserAlreadyAnonymized, LocalizedError(http.StatusConflict, "user_already_anonymized", "user.user_already_anonymized")),
			MapError(domainUser.ErrPasswordHashingFailed, LocalizedError(http.StatusInternalServerError, "password_hashing_failed", "user.password_hashing_failed")),
			MapError(domainUser.ErrForbiddenUserDirectory, LocalizedError(http.StatusForbidden, "forbidden_user_directory", "user.forbidden_user_directory")),
		},
		registrationMappings: []ErrorMapping{
			MapError(domainUser.ErrRegistrationTokenInvalid, LocalizedError(http.StatusNotFound, "registration_invalid", "auth.registration_invalid")),
			MapError(domainUser.ErrRegistrationTokenExpired, LocalizedError(http.StatusGone, "registration_expired", "auth.registration_expired")),
			MapError(domainUser.ErrRegistrationAlreadyAccepted, LocalizedError(http.StatusConflict, "registration_already_accepted", "auth.registration_already_accepted")),
			MapError(domainUser.ErrRegistrationUserDeleted, LocalizedError(http.StatusConflict, "registration_user_deleted", "auth.registration_user_deleted")),
			MapError(domainUser.ErrRegistrationPending, LocalizedError(http.StatusConflict, "registration_pending", "auth.registration_pending")),
			MapError(domainUser.ErrRegistrationResendTooSoon, LocalizedError(http.StatusTooManyRequests, "registration_resend_too_soon", "auth.registration_resend_too_soon")),
		},
	}
}

// RespondUserError responds to a user domain error with appropriate HTTP status and localized message.
func (er *ErrorResponder) RespondUserError(c *gin.Context, err error) {
	fallback := LocalizedError(http.StatusInternalServerError, "user_error", "user.user_error")
	RespondDomainError(c, err, fallback, er.userMappings...)
}

// RespondRegistrationError responds to a registration domain error with appropriate HTTP status and localized message.
func (er *ErrorResponder) RespondRegistrationError(c *gin.Context, err error) {
	fallback := LocalizedError(http.StatusInternalServerError, "registration_failed", "auth.registration_failed")
	RespondDomainError(c, err, fallback, er.registrationMappings...)
}
