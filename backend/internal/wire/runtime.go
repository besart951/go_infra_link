package wire

import (
	projecthandler "github.com/besart951/go_infra_link/backend/internal/handler/project"
	"github.com/besart951/go_infra_link/backend/internal/infrastructure/realtime"
)

type RuntimeAdapters struct {
	ProjectCollaboration     *projecthandler.ProjectCollaborationHub
	SystemNotificationStream *realtime.SystemNotificationHub
}

func NewRuntimeAdapters() *RuntimeAdapters {
	return &RuntimeAdapters{
		ProjectCollaboration:     projecthandler.NewProjectCollaborationHub(),
		SystemNotificationStream: realtime.NewSystemNotificationHub(),
	}
}
