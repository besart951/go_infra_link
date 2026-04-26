package project

import "strings"

const (
	projectRefreshScopeProject        = "project"
	projectRefreshScopeProjectUsers   = "project_users"
	projectRefreshScopeControlCabinet = "control_cabinet"
	projectRefreshScopeSPSController  = "sps_controller"
	projectRefreshScopeFieldDevice    = "field_device"
)

func refreshScopeForProjectEvent(eventType string) (string, bool) {
	switch {
	case strings.HasPrefix(eventType, "project.control_cabinet."):
		return projectRefreshScopeControlCabinet, true
	case strings.HasPrefix(eventType, "project.sps_controller."):
		return projectRefreshScopeSPSController, true
	case eventType == "project.sps_controller_system_type.copied":
		return projectRefreshScopeSPSController, true
	case strings.HasPrefix(eventType, "project.field_device."):
		return projectRefreshScopeFieldDevice, true
	case strings.HasPrefix(eventType, "project.object_data."):
		return projectRefreshScopeProject, true
	case strings.HasPrefix(eventType, "project.user."):
		return projectRefreshScopeProjectUsers, true
	case eventType == "project.updated" || eventType == "project.deleted":
		return projectRefreshScopeProject, true
	default:
		return "", false
	}
}
