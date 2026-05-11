package user

import "errors"

var ErrPasswordHashingFailed = errors.New("password_hashing_failed")
var ErrForbiddenUserDirectory = errors.New("forbidden_user_directory")
var ErrRegistrationTokenInvalid = errors.New("registration_token_invalid")
var ErrRegistrationTokenExpired = errors.New("registration_token_expired")
var ErrRegistrationAlreadyAccepted = errors.New("registration_already_accepted")
var ErrRegistrationPending = errors.New("registration_pending")
var ErrRegistrationResendTooSoon = errors.New("registration_resend_too_soon")
var ErrRegistrationUserDeleted = errors.New("registration_user_deleted")
var ErrRoleNotAssignable = errors.New("role_not_assignable")
var ErrRestoreWindowExpired = errors.New("restore_window_expired")
var ErrUserAlreadyAnonymized = errors.New("user_already_anonymized")
var ErrDeletedUserRestorable = errors.New("deleted_user_restorable")
