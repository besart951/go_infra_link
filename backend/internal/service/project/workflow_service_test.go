package project

import (
	"context"
	"errors"
	"testing"

	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type workflowLifecycleFake struct {
	called bool
	err    error
}

func (f *workflowLifecycleFake) Create(_ context.Context, _ *domainProject.Project) error {
	f.called = true
	return f.err
}

type workflowMembershipFake struct {
	inviteCalled bool
	removeCalled bool
	listCalled   bool
	err          error
	users        []domainUser.User
}

func (f *workflowMembershipFake) InviteUser(_ context.Context, _, _ uuid.UUID) error {
	f.inviteCalled = true
	return f.err
}

func (f *workflowMembershipFake) RemoveUser(_ context.Context, _, _ uuid.UUID) error {
	f.removeCalled = true
	return f.err
}

func (f *workflowMembershipFake) ListUsers(_ context.Context, _ uuid.UUID) ([]domainUser.User, error) {
	f.listCalled = true
	return f.users, f.err
}

func TestProjectWorkflowService_DelegatesCreateProject(t *testing.T) {
	ctx := context.Background()
	lifecycle := &workflowLifecycleFake{}
	membership := &workflowMembershipFake{}
	svc := newProjectWorkflowService(lifecycle, membership)

	if err := svc.CreateProject(ctx, &domainProject.Project{}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !lifecycle.called {
		t.Fatal("expected lifecycle create to be called")
	}
}

func TestProjectWorkflowService_DelegatesMembershipOps(t *testing.T) {
	ctx := context.Background()
	membership := &workflowMembershipFake{users: []domainUser.User{{Email: "u@example.com"}}}
	svc := newProjectWorkflowService(&workflowLifecycleFake{}, membership)
	id := uuid.New()

	if err := svc.InviteUser(ctx, id, id); err != nil {
		t.Fatalf("expected invite to succeed, got %v", err)
	}
	if !membership.inviteCalled {
		t.Fatal("expected invite to be delegated")
	}

	if _, err := svc.ListUsers(ctx, id); err != nil {
		t.Fatalf("expected list to succeed, got %v", err)
	}
	if !membership.listCalled {
		t.Fatal("expected list users to be delegated")
	}

	if err := svc.RemoveUser(ctx, id, id); err != nil {
		t.Fatalf("expected remove to succeed, got %v", err)
	}
	if !membership.removeCalled {
		t.Fatal("expected remove user to be delegated")
	}
}

func TestProjectWorkflowService_PropagatesErrors(t *testing.T) {
	ctx := context.Background()
	boom := errors.New("boom")
	svc := newProjectWorkflowService(&workflowLifecycleFake{err: boom}, &workflowMembershipFake{err: boom})

	if err := svc.CreateProject(ctx, &domainProject.Project{}); !errors.Is(err, boom) {
		t.Fatalf("expected create error propagation, got %v", err)
	}
	id := uuid.New()
	if err := svc.InviteUser(ctx, id, id); !errors.Is(err, boom) {
		t.Fatalf("expected invite error propagation, got %v", err)
	}
}
