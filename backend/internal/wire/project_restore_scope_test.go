package wire

import (
	"context"
	"errors"
	"testing"

	appcontrolcabinet "github.com/besart951/go_infra_link/backend/internal/application/facility/controlcabinet"
	apphistory "github.com/besart951/go_infra_link/backend/internal/application/history"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type projectRestoreAccessStub struct {
	hasAccess bool
	err       error
	calls     int
	actorID   uuid.UUID
	projectID uuid.UUID
}

func (s *projectRestoreAccessStub) CanAccessProject(
	_ context.Context,
	actorID uuid.UUID,
	projectID uuid.UUID,
	_ *domainUser.Role,
) (bool, error) {
	s.calls++
	s.actorID = actorID
	s.projectID = projectID
	return s.hasAccess, s.err
}

type projectRestoreCabinetLookupStub struct {
	links []*domainProject.ProjectControlCabinet
	err   error
	calls int
}

func (s *projectRestoreCabinetLookupStub) GetByControlCabinetID(
	context.Context,
	uuid.UUID,
) ([]*domainProject.ProjectControlCabinet, error) {
	s.calls++
	return s.links, s.err
}

type projectRestoreHistoryScopeStub struct {
	hasScope bool
	err      error
	calls    int
}

func (s *projectRestoreHistoryScopeStub) HasHistoricalProjectControlCabinet(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) (bool, error) {
	s.calls++
	return s.hasScope, s.err
}

func TestProjectControlCabinetRestoreScopeRequiresProjectAccessFirst(t *testing.T) {
	actorID := uuid.New()
	projectID := uuid.New()
	access := &projectRestoreAccessStub{}
	links := &projectRestoreCabinetLookupStub{}
	history := &projectRestoreHistoryScopeStub{}
	scope := &projectControlCabinetRestoreScope{
		access:  access,
		links:   links,
		history: history,
	}

	err := scope.RequireControlCabinetRestoreScope(
		context.Background(),
		actorID,
		projectID,
		uuid.New(),
	)
	if !errors.Is(err, appcontrolcabinet.ErrProjectRestoreAccessDenied) {
		t.Fatalf("error: got %v, want access denied", err)
	}
	if access.calls != 1 || access.actorID != actorID || access.projectID != projectID {
		t.Fatalf("access call: %+v", access)
	}
	if links.calls != 0 || history.calls != 0 {
		t.Fatal("cabinet association was queried before project access passed")
	}
}

func TestProjectControlCabinetRestoreScopeAcceptsCurrentLinkWithoutHistoryQuery(t *testing.T) {
	projectID := uuid.New()
	controlCabinetID := uuid.New()
	history := &projectRestoreHistoryScopeStub{}
	scope := &projectControlCabinetRestoreScope{
		access: &projectRestoreAccessStub{hasAccess: true},
		links: &projectRestoreCabinetLookupStub{links: []*domainProject.ProjectControlCabinet{
			{ProjectID: projectID, ControlCabinetID: controlCabinetID},
		}},
		history: history,
	}

	err := scope.RequireControlCabinetRestoreScope(
		context.Background(),
		uuid.New(),
		projectID,
		controlCabinetID,
	)
	if err != nil {
		t.Fatalf("RequireControlCabinetRestoreScope returned error: %v", err)
	}
	if history.calls != 0 {
		t.Fatal("historical scope query should be skipped for a current link")
	}
}

func TestProjectControlCabinetRestoreScopeAcceptsHistoricalDeletedLink(t *testing.T) {
	history := &projectRestoreHistoryScopeStub{hasScope: true}
	scope := &projectControlCabinetRestoreScope{
		access:  &projectRestoreAccessStub{hasAccess: true},
		links:   &projectRestoreCabinetLookupStub{},
		history: history,
	}

	err := scope.RequireControlCabinetRestoreScope(
		context.Background(),
		uuid.New(),
		uuid.New(),
		uuid.New(),
	)
	if err != nil {
		t.Fatalf("RequireControlCabinetRestoreScope returned error: %v", err)
	}
	if history.calls != 1 {
		t.Fatalf("historical scope calls: got %d, want 1", history.calls)
	}
}

func TestProjectControlCabinetRestoreScopeRejectsUnrelatedCabinet(t *testing.T) {
	scope := &projectControlCabinetRestoreScope{
		access:  &projectRestoreAccessStub{hasAccess: true},
		links:   &projectRestoreCabinetLookupStub{},
		history: &projectRestoreHistoryScopeStub{},
	}

	err := scope.RequireControlCabinetRestoreScope(
		context.Background(),
		uuid.New(),
		uuid.New(),
		uuid.New(),
	)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("error: got %v, want not found", err)
	}
}

func TestProjectControlCabinetRestoreScopePropagatesLookupErrors(t *testing.T) {
	accessErr := errors.New("access lookup failed")
	linkErr := errors.New("link lookup failed")
	historyErr := errors.New("history lookup failed")

	for _, test := range []struct {
		name    string
		access  *projectRestoreAccessStub
		links   *projectRestoreCabinetLookupStub
		history *projectRestoreHistoryScopeStub
		want    error
	}{
		{
			name:    "access",
			access:  &projectRestoreAccessStub{err: accessErr},
			links:   &projectRestoreCabinetLookupStub{},
			history: &projectRestoreHistoryScopeStub{},
			want:    accessErr,
		},
		{
			name:    "current link",
			access:  &projectRestoreAccessStub{hasAccess: true},
			links:   &projectRestoreCabinetLookupStub{err: linkErr},
			history: &projectRestoreHistoryScopeStub{},
			want:    linkErr,
		},
		{
			name:    "historical scope",
			access:  &projectRestoreAccessStub{hasAccess: true},
			links:   &projectRestoreCabinetLookupStub{},
			history: &projectRestoreHistoryScopeStub{err: historyErr},
			want:    historyErr,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			scope := &projectControlCabinetRestoreScope{
				access:  test.access,
				links:   test.links,
				history: test.history,
			}
			err := scope.RequireControlCabinetRestoreScope(
				context.Background(),
				uuid.New(),
				uuid.New(),
				uuid.New(),
			)
			if !errors.Is(err, test.want) {
				t.Fatalf("error: got %v, want %v", err, test.want)
			}
		})
	}
}

func TestProjectHistoryAccessScopeUsesProjectAccessPolicy(t *testing.T) {
	actorID := uuid.New()
	projectID := uuid.New()
	access := &projectRestoreAccessStub{hasAccess: true}
	scope := &projectHistoryAccessScope{access: access}

	if err := scope.RequireProjectAccess(
		context.Background(),
		actorID,
		projectID,
	); err != nil {
		t.Fatalf("RequireProjectAccess returned error: %v", err)
	}
	if access.calls != 1 || access.actorID != actorID || access.projectID != projectID {
		t.Fatalf("access call: %+v", access)
	}
}

func TestProjectHistoryAccessScopeRejectsDeniedOrMissingPolicy(t *testing.T) {
	for _, test := range []struct {
		name  string
		scope *projectHistoryAccessScope
		want  error
	}{
		{
			name:  "missing policy",
			scope: &projectHistoryAccessScope{},
			want:  apphistory.ErrProjectTimelineNotConfigured,
		},
		{
			name: "denied",
			scope: &projectHistoryAccessScope{
				access: &projectRestoreAccessStub{},
			},
			want: apphistory.ErrProjectTimelineAccessDenied,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := test.scope.RequireProjectAccess(
				context.Background(),
				uuid.New(),
				uuid.New(),
			)
			if !errors.Is(err, test.want) {
				t.Fatalf("error: got %v, want %v", err, test.want)
			}
		})
	}
}
