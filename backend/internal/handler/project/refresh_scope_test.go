package project

import "testing"

func TestRefreshScopeForProjectEvent(t *testing.T) {
	testCases := []struct {
		name      string
		eventType string
		wantScope string
		wantOK    bool
	}{
		{name: "control cabinet event", eventType: "project.control_cabinet.created", wantScope: projectRefreshScopeControlCabinet, wantOK: true},
		{name: "sps controller event", eventType: "project.sps_controller.deleted", wantScope: projectRefreshScopeSPSController, wantOK: true},
		{name: "sps system type copied", eventType: "project.sps_controller_system_type.copied", wantScope: projectRefreshScopeSPSController, wantOK: true},
		{name: "field device event", eventType: "project.field_device.updated", wantScope: projectRefreshScopeFieldDevice, wantOK: true},
		{name: "object data event", eventType: "project.object_data.deleted", wantScope: projectRefreshScopeProject, wantOK: true},
		{name: "project user invited", eventType: "project.user.invited", wantScope: projectRefreshScopeProjectUsers, wantOK: true},
		{name: "project user removed", eventType: "project.user.removed", wantScope: projectRefreshScopeProjectUsers, wantOK: true},
		{name: "project updated", eventType: "project.updated", wantScope: projectRefreshScopeProject, wantOK: true},
		{name: "unmapped event", eventType: "project.phase.created", wantScope: "", wantOK: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotScope, gotOK := refreshScopeForProjectEvent(tc.eventType)
			if gotOK != tc.wantOK {
				t.Fatalf("expected ok=%t, got %t", tc.wantOK, gotOK)
			}
			if gotScope != tc.wantScope {
				t.Fatalf("expected scope %q, got %q", tc.wantScope, gotScope)
			}
		})
	}
}
