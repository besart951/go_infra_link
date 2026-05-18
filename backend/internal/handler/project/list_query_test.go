package project

import (
	"net/http"
	"net/http/httptest"
	"testing"

	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestListProjectsQueryAcceptsPhaseIDQueryString(t *testing.T) {
	gin.SetMode(gin.TestMode)
	phaseID := uuid.New()
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/projects?phase_id="+phaseID.String(), nil)

	var query dto.ListProjectsQuery
	if !handlerutil.BindQuery(context, &query) {
		t.Fatalf("expected phase_id query binding to succeed, got status %d body %s", recorder.Code, recorder.Body.String())
	}
	if query.PhaseID != phaseID.String() {
		t.Fatalf("expected phase id %s, got %q", phaseID, query.PhaseID)
	}
}
