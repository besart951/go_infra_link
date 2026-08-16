package facility

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/handler/facility/internal/routing"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type facilityChangeBroadcasterSpy struct {
	resource                  string
	action                    string
	ids                       []uuid.UUID
	called                    int
	referenceCacheInvalidated [][]string
}

func (s *facilityChangeBroadcasterSpy) BroadcastFacilityChange(_ context.Context, resource, action string, ids []uuid.UUID, _ *uuid.UUID) {
	s.resource = resource
	s.action = action
	s.ids = append([]uuid.UUID(nil), ids...)
	s.called++
}

func (s *facilityChangeBroadcasterSpy) BroadcastFacilityReferenceDataChange(_ context.Context, resources ...string) {
	s.referenceCacheInvalidated = append(s.referenceCacheInvalidated, append([]string(nil), resources...))
}

func TestFacilityMutationForRoute(t *testing.T) {
	tests := []struct {
		name  string
		route routing.Definition
		want  facilityMutation
		ok    bool
	}{
		{name: "standard update", route: routing.Put("/apparats/:id", "", nil), want: facilityMutation{resource: "apparats", action: "updated", pathIDIsTarget: true}, ok: true},
		{name: "copy", route: routing.Post("/control-cabinets/:id/copy", "", nil), want: facilityMutation{resource: "control_cabinets", action: "copied", pathIDIsTarget: true}, ok: true},
		{name: "bulk delete", route: routing.Delete("/field-devices/bulk-delete", "", nil), want: facilityMutation{resource: "field_devices", action: "bulk_deleted", pathIDIsTarget: true}, ok: true},
		{name: "multi create", route: routing.Post("/field-devices/multi-create", "", nil), want: facilityMutation{resource: "field_devices", action: "bulk_created", pathIDIsTarget: true}, ok: true},
		{name: "mapping creation does not publish parent alarm type ID", route: routing.Post("/alarm-types/:id/fields", "", nil), want: facilityMutation{resource: "alarm_type_fields", action: "created"}, ok: true},
		{name: "alarm unit route", route: routing.Put("/alarm-units/:id", "", nil), want: facilityMutation{resource: "units", action: "updated", pathIDIsTarget: true}, ok: true},
		{name: "validation is excluded", route: routing.Post("/buildings/validate", "", nil)},
		{name: "cabinet validation is excluded", route: routing.Post("/control-cabinets/validate", "", nil)},
		{name: "controller validation is excluded", route: routing.Post("/sps-controllers/validate", "", nil)},
		{name: "bulk read is excluded", route: routing.Post("/apparats/bulk", "", nil)},
		{name: "building bulk read is excluded", route: routing.Post("/buildings/bulk", "", nil)},
		{name: "cabinet bulk read is excluded", route: routing.Post("/control-cabinets/bulk", "", nil)},
		{name: "controller bulk read is excluded", route: routing.Post("/sps-controllers/bulk", "", nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := facilityMutationForRoute(tt.route)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("facilityMutationForRoute() = %#v, %v; want %#v, %v", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestEveryFacilityMutationRouteHasCatalogResource(t *testing.T) {
	for _, route := range routeDefinitions(&Handlers{}) {
		if !isFacilityMutationRoute(route) {
			continue
		}
		if _, ok := facilityMutationForRoute(route); !ok {
			t.Errorf("mutation route %s %s has no facility resource catalog entry", route.Method, route.Path)
		}
	}
}

func TestFacilityChangeBroadcastsOnlySuccessfulMutations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	spy := &facilityChangeBroadcasterSpy{}
	handler := withFacilityChangeBroadcast(
		func(c *gin.Context) { c.Status(http.StatusNoContent) },
		spy,
		facilityMutation{resource: "apparats", action: "deleted", pathIDIsTarget: true},
	)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/", nil)
	ctx.Params = gin.Params{{Key: "id", Value: id.String()}}
	handler(ctx)
	if spy.called != 1 || spy.resource != "apparats" || spy.action != "deleted" || len(spy.ids) != 1 || spy.ids[0] != id {
		t.Fatalf("broadcast = %#v, want deleted apparat %s", spy, id)
	}
	if len(spy.referenceCacheInvalidated) != 1 || len(spy.referenceCacheInvalidated[0]) != 1 || spy.referenceCacheInvalidated[0][0] != "apparats" {
		t.Fatalf("reference cache invalidation = %#v, want apparats", spy.referenceCacheInvalidated)
	}

	failingSpy := &facilityChangeBroadcasterSpy{}
	failing := withFacilityChangeBroadcast(
		func(c *gin.Context) { c.Status(http.StatusBadRequest) },
		failingSpy,
		facilityMutation{resource: "apparats", action: "updated", pathIDIsTarget: true},
	)
	failingCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	failingCtx.Request = httptest.NewRequest(http.MethodPut, "/", nil)
	failingCtx.Params = gin.Params{{Key: "id", Value: id.String()}}
	failing(failingCtx)
	if failingSpy.called != 0 {
		t.Fatalf("failed mutation broadcast %d times, want 0", failingSpy.called)
	}
	if len(failingSpy.referenceCacheInvalidated) != 0 {
		t.Fatalf("failed mutation invalidated reference cache: %#v", failingSpy.referenceCacheInvalidated)
	}
}

func TestFacilityChangeBroadcastExtractsCreatedAndBulkIDsFromResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	createdID := uuid.New()
	bulkID := uuid.New()
	spy := &facilityChangeBroadcasterSpy{}
	handler := withFacilityChangeBroadcast(
		func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{
				"id": createdID,
				"results": []gin.H{
					{"success": true, "field_device": gin.H{"id": bulkID}},
					{"success": false, "id": uuid.New()},
				},
			})
		},
		spy,
		facilityMutation{resource: "field_devices", action: "bulk_created"},
	)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	handler(ctx)

	if spy.called != 1 || len(spy.ids) != 2 || spy.ids[0] != createdID || spy.ids[1] != bulkID {
		t.Fatalf("broadcast IDs = %#v, want [%s %s]", spy.ids, createdID, bulkID)
	}
}

func TestFacilityChangeIDsFromResponseDoesNotUseRelatedEntityIDs(t *testing.T) {
	targetID := uuid.New()
	relatedID := uuid.New()
	ids := facilityChangeIDsFromResponse([]byte(`{"id":"` + targetID.String() + `","system_parts":[{"id":"` + relatedID.String() + `"}]}`))
	if len(ids) != 1 || ids[0] != targetID {
		t.Fatalf("IDs = %#v, want only target %s", ids, targetID)
	}
}
