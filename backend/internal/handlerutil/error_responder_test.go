package handlerutil

import (
	"errors"
	"testing"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
)

func TestErrorResponder_RoleNotAssignableError(t *testing.T) {
	testError := domainUser.ErrRoleNotAssignable

	if !errors.Is(testError, domainUser.ErrRoleNotAssignable) {
		t.Fatalf("test error should match ErrRoleNotAssignable")
	}
}

func TestErrorResponder_DeletedUserRestorableError(t *testing.T) {
	testError := domainUser.ErrDeletedUserRestorable

	if !errors.Is(testError, domainUser.ErrDeletedUserRestorable) {
		t.Fatalf("responder should match ErrDeletedUserRestorable")
	}
}

func TestErrorResponder_RegistrationTokenExpiredError(t *testing.T) {
	testError := domainUser.ErrRegistrationTokenExpired

	if !errors.Is(testError, domainUser.ErrRegistrationTokenExpired) {
		t.Fatalf("responder should match ErrRegistrationTokenExpired")
	}
}

func TestErrorResponder_RegistrationUserDeleted(t *testing.T) {
	testError := domainUser.ErrRegistrationUserDeleted

	if !errors.Is(testError, domainUser.ErrRegistrationUserDeleted) {
		t.Fatalf("responder should match ErrRegistrationUserDeleted")
	}
}

func TestErrorResponder_HasUserMappings(t *testing.T) {
	responder := NewErrorResponder()

	if len(responder.userMappings) == 0 {
		t.Fatalf("responder should have user mappings")
	}

	if len(responder.registrationMappings) == 0 {
		t.Fatalf("responder should have registration mappings")
	}
}
