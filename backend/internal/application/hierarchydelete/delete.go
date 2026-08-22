package hierarchydelete

import (
	"context"

	"github.com/google/uuid"
)

type RootKind string

const (
	RootControlCabinet          RootKind = "control_cabinet"
	RootSPSController           RootKind = "sps_controller"
	RootSPSControllerSystemType RootKind = "sps_controller_system_type"
)

type Stage string

const (
	StageFieldDevices Stage = "field_devices"
	StageSystemTypes  Stage = "system_types"
	StageControllers  Stage = "controllers"
	StageRoot         Stage = "root"
)

type Command struct {
	RootKind RootKind
	RootID   uuid.UUID
	Stage    Stage
	Limit    int
	ActorID  uuid.UUID
	BatchID  uuid.UUID
}

type Result struct {
	Deleted int
	Done    bool
}

type Store interface {
	DeleteChunk(context.Context, Command) (Result, error)
}

var stagePlans = map[RootKind][]Stage{
	RootControlCabinet:          {StageFieldDevices, StageSystemTypes, StageControllers, StageRoot},
	RootSPSController:           {StageFieldDevices, StageSystemTypes, StageRoot},
	RootSPSControllerSystemType: {StageFieldDevices, StageRoot},
}

func Stages(kind RootKind) []Stage {
	stages := stagePlans[kind]
	return append([]Stage(nil), stages...)
}
