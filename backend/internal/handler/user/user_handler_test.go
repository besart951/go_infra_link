package user

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	userregistration "github.com/besart951/go_infra_link/backend/internal/service/userregistration"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestCreateUserRejectsNonSuperadminDirectCreate(t *testing.T) {
	handler, fakeService := newTestUserHandler()
	fakeService.createErr = domainUser.ErrRoleNotAssignable
	router := newTestUserRouter(handler, uuid.New())
	router.POST("/users", handler.CreateUser)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{
		"first_name":"Ada",
		"last_name":"Lovelace",
		"email":"ada@example.com",
		"password":"CorrectHorse1",
		"role":"planer"
	}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", res.Code, res.Body.String())
	}
}

func TestCreateUserAllowsServiceApprovedDirectCreate(t *testing.T) {
	handler, _ := newTestUserHandler()
	router := newTestUserRouter(handler, uuid.New())
	router.POST("/users", handler.CreateUser)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{
		"first_name":"Grace",
		"last_name":"Hopper",
		"email":"grace@example.com",
		"password":"CorrectHorse1",
		"role":"superadmin"
	}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", res.Code, res.Body.String())
	}
}

func TestUpdateUserRejectsRestrictedFields(t *testing.T) {
	handler, fakeService := newTestUserHandler()
	router := newTestUserRouter(handler, uuid.New())
	router.PUT("/users/:id", handler.UpdateUser)
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewBufferString(`{
		"first_name":"Ada",
		"role":"superadmin",
		"is_active":true,
		"password":"CorrectHorse1"
	}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", res.Code, res.Body.String())
	}
	if fakeService.getByIDCalled {
		t.Fatalf("restricted update must not load or mutate user")
	}
}

func TestUpdateOwnPasswordRejectsWrongCurrentPassword(t *testing.T) {
	handler, fakeService := newTestUserHandler()
	fakeService.passwordErr = domainAuth.ErrInvalidCredentials
	router := newTestUserRouter(handler, uuid.New())
	router.PUT("/users/me/password", handler.UpdateOwnPassword)

	req := httptest.NewRequest(http.MethodPut, "/users/me/password", bytes.NewBufferString(`{
		"current_password":"wrong-password",
		"new_password":"CorrectHorse1"
	}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", res.Code, res.Body.String())
	}
}

func TestResendInvitationTooSoonReturns429(t *testing.T) {
	registrationService := &fakeRegistrationService{resendErr: domainUser.ErrRegistrationResendTooSoon}
	handler := NewRegistrationHandler(registrationService)
	router := newTestUserRouter(nil, uuid.New())
	router.POST("/users/:id/registration/resend", handler.ResendInvitation)

	req := httptest.NewRequest(http.MethodPost, "/users/"+uuid.NewString()+"/registration/resend", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d: %s", res.Code, res.Body.String())
	}
}

func TestGetRegistrationProcessOutsideScopeReturns403(t *testing.T) {
	actorID := uuid.New()
	targetID := uuid.New()
	registrationService := &fakeRegistrationService{getErr: domainUser.ErrRoleNotAssignable}
	handler := NewRegistrationHandler(registrationService)
	router := newTestUserRouter(nil, actorID)
	router.GET("/users/:id/registration", handler.GetProcess)

	req := httptest.NewRequest(http.MethodGet, "/users/"+targetID.String()+"/registration", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", res.Code, res.Body.String())
	}
	if registrationService.getActorID != actorID || registrationService.getUserID != targetID {
		t.Fatalf("expected actor and target IDs to be forwarded")
	}
}

func newTestUserHandler() (*UserHandler, *fakeUserService) {
	gin.SetMode(gin.TestMode)
	service := &fakeUserService{}
	return NewUserHandler(service, nil, nil, nil), service
}

func newTestUserRouter(_ *UserHandler, actorID uuid.UUID) *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, actorID)
		c.Next()
	})
	return router
}

type fakeUserService struct {
	createErr     error
	passwordErr   error
	getByIDCalled bool
}

func (s *fakeUserService) CreateWithPassword(context.Context, *domainUser.User, string) error {
	return nil
}

func (s *fakeUserService) CreateWithPasswordForActor(_ context.Context, _ uuid.UUID, user *domainUser.User, _ string) error {
	if s.createErr != nil {
		return s.createErr
	}
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	return nil
}

func (s *fakeUserService) UpdateWithPassword(context.Context, *domainUser.User, *string) error {
	return nil
}

func (s *fakeUserService) UpdateProfileForActor(context.Context, uuid.UUID, *domainUser.User) error {
	return nil
}

func (s *fakeUserService) UpdatePassword(_ context.Context, userID uuid.UUID, _, newPassword string) (*domainUser.User, error) {
	if s.passwordErr != nil {
		return nil, s.passwordErr
	}
	return &domainUser.User{Base: domain.Base{ID: userID}, Email: domainUser.EmailPtr("ada@example.com"), Password: newPassword, Role: domainUser.RolePlaner}, nil
}

func (s *fakeUserService) GetByID(context.Context, uuid.UUID) (*domainUser.User, error) {
	s.getByIDCalled = true
	return &domainUser.User{Role: domainUser.RolePlaner}, nil
}

func (s *fakeUserService) List(context.Context, int, int, string, string, string, bool) (*domain.PaginatedList[domainUser.User], error) {
	return &domain.PaginatedList[domainUser.User]{}, nil
}

func (s *fakeUserService) DeleteByID(context.Context, uuid.UUID) error {
	return nil
}

func (s *fakeUserService) DeleteByIDForActor(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (s *fakeUserService) RestoreByIDForActor(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

type fakeRegistrationService struct {
	getErr     error
	resendErr  error
	getActorID uuid.UUID
	getUserID  uuid.UUID
}

func (s *fakeRegistrationService) CreateInvitation(context.Context, userregistration.InviteInput) (*domainUser.User, *userregistration.Process, error) {
	return nil, nil, nil
}

func (s *fakeRegistrationService) GetProcess(_ context.Context, actorID, userID uuid.UUID) (*userregistration.Process, error) {
	s.getActorID = actorID
	s.getUserID = userID
	if s.getErr != nil {
		return nil, s.getErr
	}
	return &userregistration.Process{}, nil
}

func (s *fakeRegistrationService) ListProcessesForUsers(context.Context, []domainUser.User) (map[uuid.UUID]*userregistration.Process, error) {
	return map[uuid.UUID]*userregistration.Process{}, nil
}

func (s *fakeRegistrationService) ResendInvitation(context.Context, uuid.UUID, uuid.UUID) (*userregistration.Process, error) {
	if s.resendErr != nil {
		return nil, s.resendErr
	}
	return &userregistration.Process{}, nil
}
