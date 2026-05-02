package wire

import "github.com/besart951/go_infra_link/backend/internal/infrastructure/realtime"

type RuntimeAdapters struct {
	ProjectCollaboration     *realtime.ProjectCollaborationHub
	SystemNotificationStream *realtime.SystemNotificationHub
}

func NewRuntimeAdapters() *RuntimeAdapters {
	return &RuntimeAdapters{
		ProjectCollaboration:     realtime.NewProjectCollaborationHub(),
		SystemNotificationStream: realtime.NewSystemNotificationHub(),
	}
}
