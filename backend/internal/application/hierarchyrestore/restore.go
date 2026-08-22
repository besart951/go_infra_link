package hierarchyrestore

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Phase string

const (
	PhaseDelete  Phase = "delete"
	PhaseRestore Phase = "restore"
)

type Position struct {
	PhaseIndex int       `json:"phase_index"`
	TableIndex int       `json:"table_index"`
	AfterID    uuid.UUID `json:"after_id,omitempty"`
	Ordinal    int64     `json:"ordinal"`
	Processed  int64     `json:"processed"`
	Restored   int64     `json:"restored"`
	Deleted    int64     `json:"deleted"`
	Skipped    int64     `json:"skipped"`
}

type Command struct {
	ControlCabinetID uuid.UUID
	ProjectID        *uuid.UUID
	AsOf             time.Time
	Phase            Phase
	Table            string
	AfterID          uuid.UUID
	Limit            int
	ActorID          uuid.UUID
	BatchID          uuid.UUID
}

type Result struct {
	NextID    uuid.UUID `json:"next_id,omitempty"`
	Done      bool      `json:"done"`
	Processed int       `json:"processed"`
	Restored  int       `json:"restored"`
	Deleted   int       `json:"deleted"`
	Skipped   int       `json:"skipped"`
}

type Store interface {
	RestoreChunk(context.Context, Command) (Result, error)
}

var phaseTables = map[Phase][]string{
	PhaseDelete: {
		"project_field_devices", "project_sps_controllers", "project_control_cabinets",
		"bacnet_object_alarm_values", "bacnet_objects", "specifications", "field_devices",
		"sps_controller_system_types", "sps_controllers", "control_cabinets",
	},
	PhaseRestore: {
		"control_cabinets", "sps_controllers", "sps_controller_system_types", "field_devices",
		"specifications", "bacnet_objects", "bacnet_object_alarm_values",
		"project_control_cabinets", "project_sps_controllers", "project_field_devices",
	},
}

func Phases() []Phase {
	return []Phase{PhaseDelete, PhaseRestore}
}

func Tables(phase Phase) []string {
	return append([]string(nil), phaseTables[phase]...)
}
