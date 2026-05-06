package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type sessionTokenValidatorStub struct {
	userID uuid.UUID
	err    error
	token  string
}

func (s *sessionTokenValidatorStub) ValidateToken(token string) (uuid.UUID, error) {
	s.token = token
	return s.userID, s.err
}

type sessionUserServiceStub struct {
	user *domainUser.User
	err  error
	id   uuid.UUID
}

func (s *sessionUserServiceStub) GetByID(_ context.Context, id uuid.UUID) (*domainUser.User, error) {
	s.id = id
	return s.user, s.err
}

func TestSessionReturnsGuestWithoutAccessToken(t *testing.T) {
	handler := NewAuthHandler(nil, nil, nil, nil, 0, 0, CookieSettings{})
	response := runSessionRequest(handler, "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if authenticated := decodeSessionResponse(t, response).Authenticated; authenticated {
		t.Fatal("expected guest session")
	}
}

func TestSessionReturnsAuthenticatedForValidActiveUser(t *testing.T) {
	userID := uuid.New()
	validator := &sessionTokenValidatorStub{userID: userID}
	users := &sessionUserServiceStub{user: &domainUser.User{
		Base: domain.Base{
			ID:        userID,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		IsActive: true,
		Role:     domainUser.RoleSuperAdmin,
	}}
	handler := NewAuthHandler(nil, users, nil, validator, 0, 0, CookieSettings{})

	response := runSessionRequest(handler, "valid-token")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if authenticated := decodeSessionResponse(t, response).Authenticated; !authenticated {
		t.Fatal("expected authenticated session")
	}
	if validator.token != "valid-token" {
		t.Fatalf("expected token validator to receive cookie value, got %q", validator.token)
	}
	if users.id != userID {
		t.Fatalf("expected user lookup for %s, got %s", userID, users.id)
	}
}

func TestSessionReturnsGuestForInvalidOrInactiveSession(t *testing.T) {
	userID := uuid.New()
	disabledAt := time.Now().UTC()
	lockedUntil := time.Now().UTC().Add(time.Hour)

	tests := []struct {
		name      string
		validator *sessionTokenValidatorStub
		users     *sessionUserServiceStub
	}{
		{
			name:      "invalid token",
			validator: &sessionTokenValidatorStub{err: errors.New("invalid token")},
			users:     &sessionUserServiceStub{},
		},
		{
			name:      "missing user",
			validator: &sessionTokenValidatorStub{userID: userID},
			users:     &sessionUserServiceStub{},
		},
		{
			name:      "disabled user",
			validator: &sessionTokenValidatorStub{userID: userID},
			users: &sessionUserServiceStub{user: &domainUser.User{
				Base:       domain.Base{ID: userID},
				IsActive:   false,
				DisabledAt: &disabledAt,
			}},
		},
		{
			name:      "locked user",
			validator: &sessionTokenValidatorStub{userID: userID},
			users: &sessionUserServiceStub{user: &domainUser.User{
				Base:        domain.Base{ID: userID},
				IsActive:    true,
				LockedUntil: &lockedUntil,
			}},
		},
		{
			name:      "lookup error",
			validator: &sessionTokenValidatorStub{userID: userID},
			users:     &sessionUserServiceStub{err: errors.New("database unavailable")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewAuthHandler(nil, tt.users, nil, tt.validator, 0, 0, CookieSettings{})
			response := runSessionRequest(handler, "token")

			if response.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", response.Code)
			}
			if authenticated := decodeSessionResponse(t, response).Authenticated; authenticated {
				t.Fatal("expected guest session")
			}
		})
	}
}

func runSessionRequest(handler *AuthHandler, accessToken string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	if accessToken != "" {
		context.Request.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	}

	handler.Session(context)
	return recorder
}

func decodeSessionResponse(t *testing.T, recorder *httptest.ResponseRecorder) dto.SessionResponse {
	t.Helper()

	var response dto.SessionResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode session response: %v", err)
	}
	return response
}
