package history

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	apphistory "github.com/besart951/go_infra_link/backend/internal/application/history"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type projectTimelineReaderHandlerStub struct {
	query  apphistory.ListProjectTimelineQuery
	result *domain.PaginatedList[domainHistory.ChangeEvent]
	err    error
	calls  int
}

func (s *projectTimelineReaderHandlerStub) ListProjectTimeline(
	_ context.Context,
	query apphistory.ListProjectTimelineQuery,
) (*domain.PaginatedList[domainHistory.ChangeEvent], error) {
	s.calls++
	s.query = query
	return s.result, s.err
}

func TestListProjectTimelineDelegatesTypedApplicationQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	projectID := uuid.New()
	controlCabinetID := uuid.New()
	eventID := uuid.New()
	reader := &projectTimelineReaderHandlerStub{
		result: &domain.PaginatedList[domainHistory.ChangeEvent]{
			Items: []domainHistory.ChangeEvent{{ID: eventID}},
			Total: 1,
			Page:  2,
		},
	}
	handler := NewHandler(historyServiceStub{}, nil, reader)
	router := gin.New()
	router.GET("/projects/:id/history/timeline", handler.ListProjectTimeline)
	req := httptest.NewRequest(
		http.MethodGet,
		"/projects/"+projectID.String()+"/history/timeline?scope_type=control_cabinet&scope_id="+
			controlCabinetID.String()+"&entity_table=field_devices&page=2&limit=25",
		nil,
	)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%s", res.Code, res.Body.String())
	}
	if reader.calls != 1 || reader.query.ProjectID != projectID ||
		reader.query.Filter.ScopeType != "control_cabinet" ||
		reader.query.Filter.ScopeID != controlCabinetID ||
		reader.query.Filter.EntityTable != "field_devices" ||
		reader.query.Filter.Page != 2 || reader.query.Filter.Limit != 25 {
		t.Fatalf("typed query: %+v", reader.query)
	}
}

func TestListProjectTimelineMapsProjectAccessDenial(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reader := &projectTimelineReaderHandlerStub{err: apphistory.ErrProjectTimelineAccessDenied}
	handler := NewHandler(historyServiceStub{}, nil, reader)
	router := gin.New()
	router.GET("/projects/:id/history/timeline", handler.ListProjectTimeline)
	req := httptest.NewRequest(
		http.MethodGet,
		"/projects/"+uuid.NewString()+"/history/timeline",
		nil,
	)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("status: got %d body=%s", res.Code, res.Body.String())
	}
}

func TestListProjectTimelineRequiresConfiguredApplicationReader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(historyServiceStub{}, nil, nil)
	router := gin.New()
	router.GET("/projects/:id/history/timeline", handler.ListProjectTimeline)
	req := httptest.NewRequest(
		http.MethodGet,
		"/projects/"+uuid.NewString()+"/history/timeline",
		nil,
	)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d body=%s", res.Code, res.Body.String())
	}
}
