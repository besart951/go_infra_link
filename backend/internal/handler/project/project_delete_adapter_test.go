package project

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	appproject "github.com/besart951/go_infra_link/backend/internal/application/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type projectDeleteAccessStub struct{}

func (*projectDeleteAccessStub) CanAccessProject(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	*domainUser.Role,
) (bool, error) {
	return true, nil
}

func (*projectDeleteAccessStub) CanUseProjectPermission(
	context.Context,
	uuid.UUID,
	*domainUser.Role,
	string,
) (bool, error) {
	return true, nil
}

func (*projectDeleteAccessStub) CanUseProjectPermissionForProject(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	*domainUser.Role,
	string,
) (bool, error) {
	return true, nil
}

type projectDeleterStub struct {
	commands []appproject.DeleteCommand
	err      error
}

func (stub *projectDeleterStub) Delete(
	_ context.Context,
	command appproject.DeleteCommand,
) error {
	stub.commands = append(stub.commands, command)
	return stub.err
}

func TestDeleteProjectRoutesToApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	projectID := uuid.New()
	actorID := uuid.New()
	deleter := &projectDeleterStub{}
	handler := &ProjectHandler{
		access:        &projectDeleteAccessStub{},
		projectDelete: deleter,
	}
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Params = gin.Params{{Key: "id", Value: projectID.String()}}
	ginContext.Request = httptest.NewRequest(http.MethodDelete, "/", nil)
	ginContext.Set(middleware.ContextUserIDKey, actorID)
	ginContext.Set(middleware.ContextUserRoleKey, domainUser.RoleSuperAdmin)

	handler.DeleteProject(ginContext)

	if ginContext.Writer.Status() != http.StatusNoContent {
		t.Fatalf(
			"status: got %d, want %d",
			ginContext.Writer.Status(),
			http.StatusNoContent,
		)
	}
	if len(deleter.commands) != 1 ||
		deleter.commands[0].ProjectID != projectID {
		t.Fatalf("commands: %+v", deleter.commands)
	}
}

func TestDeleteProjectMapsRemainingHierarchyLinksToTypedConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	projectID := uuid.New()
	deleter := &projectDeleterStub{err: ErrWrappedHierarchyLinksRemain}
	handler := &ProjectHandler{
		access:        &projectDeleteAccessStub{},
		projectDelete: deleter,
	}
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Params = gin.Params{{Key: "id", Value: projectID.String()}}
	ginContext.Request = httptest.NewRequest(http.MethodDelete, "/", nil)
	ginContext.Set(middleware.ContextUserIDKey, uuid.New())
	ginContext.Set(middleware.ContextUserRoleKey, domainUser.RoleAdminFZAG)

	handler.DeleteProject(ginContext)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status: got %d, want %d", recorder.Code, http.StatusConflict)
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != "project_hierarchy_linked" {
		t.Fatalf("error code: got %q", body.Code)
	}
}

var ErrWrappedHierarchyLinksRemain = errors.Join(
	errors.New("delete rejected"),
	appproject.ErrHierarchyLinksRemain,
)
