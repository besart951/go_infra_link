package history

import (
	"context"
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type projectTimelineAccessStub struct {
	err       error
	actorID   uuid.UUID
	projectID uuid.UUID
	calls     int
	order     *[]string
}

func (s *projectTimelineAccessStub) RequireProjectAccess(
	_ context.Context,
	actorID uuid.UUID,
	projectID uuid.UUID,
) error {
	s.calls++
	s.actorID = actorID
	s.projectID = projectID
	if s.order != nil {
		*s.order = append(*s.order, "access")
	}
	return s.err
}

type projectTimelineReaderStub struct {
	filter domainHistory.TimelineFilter
	result *domain.PaginatedList[domainHistory.ChangeEvent]
	err    error
	calls  int
	order  *[]string
}

func (s *projectTimelineReaderStub) ListTimeline(
	_ context.Context,
	filter domainHistory.TimelineFilter,
) (*domain.PaginatedList[domainHistory.ChangeEvent], error) {
	s.calls++
	s.filter = filter
	if s.order != nil {
		*s.order = append(*s.order, "read")
	}
	return s.result, s.err
}

func TestProjectTimelineRequiresAccessThenAppliesProjectAsPrimaryScope(t *testing.T) {
	actorID := uuid.New()
	projectID := uuid.New()
	controlCabinetID := uuid.New()
	order := []string{}
	access := &projectTimelineAccessStub{order: &order}
	want := &domain.PaginatedList[domainHistory.ChangeEvent]{
		Items: []domainHistory.ChangeEvent{{ID: uuid.New()}},
		Total: 1,
		Page:  1,
	}
	reader := &projectTimelineReaderStub{result: want, order: &order}
	handler := NewProjectTimelineHandler(ProjectTimelineDependencies{
		Access:   access,
		Timeline: reader,
		Actor:    func(context.Context) *uuid.UUID { return &actorID },
	})
	input := domainHistory.TimelineFilter{
		ScopeType:   "control_cabinet",
		ScopeID:     controlCabinetID,
		EntityTable: "field_devices",
		Page:        2,
		Limit:       25,
	}

	got, err := handler.ListProjectTimeline(context.Background(), ListProjectTimelineQuery{
		ProjectID: projectID,
		Filter:    input,
	})
	if err != nil {
		t.Fatalf("ListProjectTimeline returned error: %v", err)
	}
	if got != want {
		t.Fatalf("result: got %p, want %p", got, want)
	}
	if len(order) != 2 || order[0] != "access" || order[1] != "read" {
		t.Fatalf("call order: %v", order)
	}
	if access.actorID != actorID || access.projectID != projectID {
		t.Fatalf("access request: %+v", access)
	}
	if reader.filter.ScopeType != "project" || reader.filter.ScopeID != projectID ||
		reader.filter.SecondaryScopeType != "control_cabinet" ||
		reader.filter.SecondaryScopeID != controlCabinetID ||
		reader.filter.EntityTable != input.EntityTable || reader.filter.Page != input.Page ||
		reader.filter.Limit != input.Limit {
		t.Fatalf("timeline filter: %+v", reader.filter)
	}
	if input.SecondaryScopeType != "" || input.SecondaryScopeID != uuid.Nil {
		t.Fatalf("input filter was mutated: %+v", input)
	}
}

func TestProjectTimelineRejectsMissingActorOrAccessBeforeRead(t *testing.T) {
	projectID := uuid.New()
	for _, test := range []struct {
		name   string
		actor  ActorProvider
		access *projectTimelineAccessStub
		want   error
	}{
		{
			name:   "missing actor",
			access: &projectTimelineAccessStub{},
			want:   ErrProjectTimelineAccessDenied,
		},
		{
			name:  "access denied",
			actor: func(context.Context) *uuid.UUID { id := uuid.New(); return &id },
			access: &projectTimelineAccessStub{
				err: ErrProjectTimelineAccessDenied,
			},
			want: ErrProjectTimelineAccessDenied,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			reader := &projectTimelineReaderStub{}
			handler := NewProjectTimelineHandler(ProjectTimelineDependencies{
				Access:   test.access,
				Timeline: reader,
				Actor:    test.actor,
			})
			_, err := handler.ListProjectTimeline(
				context.Background(),
				ListProjectTimelineQuery{ProjectID: projectID},
			)
			if !errors.Is(err, test.want) {
				t.Fatalf("error: got %v, want %v", err, test.want)
			}
			if reader.calls != 0 {
				t.Fatal("timeline read occurred after access denial")
			}
		})
	}
}

func TestProjectTimelinePropagatesReadErrorAfterAccess(t *testing.T) {
	actorID := uuid.New()
	readErr := errors.New("timeline unavailable")
	reader := &projectTimelineReaderStub{err: readErr}
	handler := NewProjectTimelineHandler(ProjectTimelineDependencies{
		Access:   &projectTimelineAccessStub{},
		Timeline: reader,
		Actor:    func(context.Context) *uuid.UUID { return &actorID },
	})

	_, err := handler.ListProjectTimeline(
		context.Background(),
		ListProjectTimelineQuery{ProjectID: uuid.New()},
	)
	if !errors.Is(err, readErr) {
		t.Fatalf("error: got %v, want %v", err, readErr)
	}
}

func TestProjectTimelineValidatesConfigurationAndProjectID(t *testing.T) {
	if _, err := NewProjectTimelineHandler(ProjectTimelineDependencies{}).
		ListProjectTimeline(context.Background(), ListProjectTimelineQuery{}); !errors.Is(
		err,
		ErrProjectTimelineNotConfigured,
	) {
		t.Fatalf("configuration error: got %v", err)
	}
	actorID := uuid.New()
	handler := NewProjectTimelineHandler(ProjectTimelineDependencies{
		Access:   &projectTimelineAccessStub{},
		Timeline: &projectTimelineReaderStub{},
		Actor:    func(context.Context) *uuid.UUID { return &actorID },
	})
	if _, err := handler.ListProjectTimeline(
		context.Background(),
		ListProjectTimelineQuery{},
	); !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("validation error: got %v", err)
	}
}
