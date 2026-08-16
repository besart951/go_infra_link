package shared

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestFacilityLinkHandlerStartCopyRespondsAndNotifiesInitiator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := facilityservice.NewCopyJobManager(nil)
	t.Cleanup(manager.Close)

	projectID := uuid.New()
	actorID := uuid.New()
	operationID := uuid.New()
	entityID := uuid.New()
	notified := make(chan struct {
		actorID  uuid.UUID
		project  uuid.UUID
		event    string
		entityID string
	}, 1)

	handler := NewFacilityLinkHandler(nil, nil)
	handler.ConfigureCopyJobs(manager, func(_ context.Context, actor *uuid.UUID, project uuid.UUID, event, entity string) {
		notified <- struct {
			actorID  uuid.UUID
			project  uuid.UUID
			event    string
			entityID string
		}{actorID: *actor, project: project, event: event, entityID: entity}
	})

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodPost, "/projects/"+projectID.String()+"/copy", nil)
	request.Header.Set("X-Copy-Operation-ID", operationID.String())
	c.Request = request
	c.Set(middleware.ContextUserIDKey, actorID)

	if !handler.StartCopy(c, projectID, facilityservice.CopyJobKindControlCabinet, "project.control_cabinet.copied", func(context.Context) (string, error) {
		return entityID.String(), nil
	}) {
		t.Fatal("expected configured copy jobs to handle the request")
	}
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d: %s", recorder.Code, recorder.Body.String())
	}

	select {
	case event := <-notified:
		if event.actorID != actorID || event.project != projectID || event.event != "project.control_cabinet.copied" || event.entityID != entityID.String() {
			t.Fatalf("unexpected notification %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("copy completion notification was not sent")
	}
}
