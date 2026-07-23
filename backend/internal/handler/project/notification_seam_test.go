package project

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type notificationEventDispatcherChannel struct {
	inputs chan domainNotification.DispatchEventInput
}

func (s *notificationEventDispatcherChannel) DispatchEvent(
	_ context.Context,
	input domainNotification.DispatchEventInput,
) error {
	s.inputs <- input
	return nil
}

func TestNotifyProjectEventPreservesSystemTypeCopyNotificationWithoutRealtimeCallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	actorID := uuid.New()
	projectID := uuid.New()
	spsControllerID := uuid.New()
	dispatcher := &notificationEventDispatcherChannel{
		inputs: make(chan domainNotification.DispatchEventInput, 1),
	}
	handler := &ProjectHandler{notifications: dispatcher}
	recorder := httptest.NewRecorder()
	contextValue, _ := gin.CreateTestContext(recorder)
	contextValue.Request = httptest.NewRequest("POST", "/", nil)
	contextValue.Set(middleware.ContextUserIDKey, actorID)

	handler.notifyProjectEvent(
		contextValue,
		projectID,
		"project.sps_controller_system_type.copied",
		spsControllerID.String(),
	)

	select {
	case input := <-dispatcher.inputs:
		if input.ActorID == nil || *input.ActorID != actorID ||
			input.ProjectID == nil || *input.ProjectID != projectID ||
			input.EventKey != "project.sps_controller_system_type.copied" ||
			input.Metadata["project_id"] != projectID.String() ||
			input.Metadata["count"] != "1" ||
			input.Metadata["entity_ids"] != spsControllerID.String() {
			t.Fatalf("notification input: %+v", input)
		}
		// Preserve the current notification contract: this legacy event has no
		// resource type, but still points at the owning SPSController ID.
		if input.ResourceType != "" || input.ResourceID == nil ||
			*input.ResourceID != spsControllerID {
			t.Fatalf("legacy notification resource mapping changed: %+v", input)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for project notification dispatch")
	}
}
