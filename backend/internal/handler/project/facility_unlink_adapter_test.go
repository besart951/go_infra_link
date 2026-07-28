package project

import (
	"context"
	"reflect"
	"testing"

	appprojectlink "github.com/besart951/go_infra_link/backend/internal/application/facility/projectlink"
	"github.com/google/uuid"
)

type projectFacilityUnlinkerStub struct {
	commands []appprojectlink.Command
}

func (stub *projectFacilityUnlinkerStub) Unlink(
	_ context.Context,
	command appprojectlink.Command,
) error {
	stub.commands = append(stub.commands, command)
	return nil
}

func TestApplicationProjectFacilityLinkRoutesDeletesToExactUnlinkKinds(t *testing.T) {
	projectID := uuid.New()
	cabinetLinkID := uuid.New()
	spsLinkID := uuid.New()
	fieldDeviceLinkID := uuid.New()
	unlinker := &projectFacilityUnlinkerStub{}
	service := &applicationProjectFacilityLink{unlinker: unlinker}

	if err := service.DeleteControlCabinet(context.Background(), cabinetLinkID, projectID); err != nil {
		t.Fatalf("unlink cabinet: %v", err)
	}
	if err := service.DeleteSPSController(context.Background(), spsLinkID, projectID); err != nil {
		t.Fatalf("unlink sps controller: %v", err)
	}
	if err := service.DeleteFieldDevice(context.Background(), fieldDeviceLinkID, projectID); err != nil {
		t.Fatalf("unlink field device: %v", err)
	}

	want := []appprojectlink.Command{
		{Kind: appprojectlink.KindControlCabinet, ProjectID: projectID, LinkID: cabinetLinkID},
		{Kind: appprojectlink.KindSPSController, ProjectID: projectID, LinkID: spsLinkID},
		{Kind: appprojectlink.KindFieldDevice, ProjectID: projectID, LinkID: fieldDeviceLinkID},
	}
	if !reflect.DeepEqual(unlinker.commands, want) {
		t.Fatalf("commands: got %+v, want %+v", unlinker.commands, want)
	}
}
