package wire

import (
	"context"
	"encoding/json"
	"fmt"

	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

func registerFacilityCopyTasks(jobs *facilityservice.CopyJobManager, services *Services, runtime *RuntimeAdapters) {
	if jobs == nil || services == nil || services.Facility == nil {
		return
	}
	register := func(task, resource string, copy func(context.Context, uuid.UUID) (uuid.UUID, error)) {
		jobs.RegisterTask(task, func(ctx context.Context, job facilityservice.CopyJob, report func(facilityservice.FacilityJobProgress)) (facilityservice.FacilityJobTaskResult, error) {
			var payload facilityservice.FacilityCopyTaskPayload
			if err := json.Unmarshal(job.Payload, &payload); err != nil {
				return facilityservice.FacilityJobTaskResult{}, fmt.Errorf("decode copy task payload: %w", err)
			}
			if payload.SourceID == uuid.Nil {
				return facilityservice.FacilityJobTaskResult{}, fmt.Errorf("decode copy task payload: source_id is required")
			}
			report(facilityservice.FacilityJobProgress{Progress: 10, Stage: "copying_root"})
			resultID, err := copy(ctx, payload.SourceID)
			if err != nil {
				return facilityservice.FacilityJobTaskResult{}, err
			}
			report(facilityservice.FacilityJobProgress{Progress: 95, Stage: "finalizing", Processed: 1, Succeeded: 1})
			if runtime != nil && runtime.FacilityReferenceData != nil {
				actorID := job.OwnerID
				runtime.FacilityReferenceData.BroadcastFacilityChange(ctx, resource, "copied", []uuid.UUID{resultID}, &actorID)
			}
			result, err := json.Marshal(map[string]string{"resource_id": resultID.String()})
			return facilityservice.FacilityJobTaskResult{Result: result}, err
		})
	}

	register(facilityservice.FacilityJobTaskCopyControlCabinet, "control_cabinets", func(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
		item, err := services.Facility.ControlCabinet.CopyByID(ctx, id)
		if err != nil {
			return uuid.Nil, err
		}
		return item.ID, nil
	})
	register(facilityservice.FacilityJobTaskCopySPSController, "sps_controllers", func(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
		item, err := services.Facility.SPSController.CopyByID(ctx, id)
		if err != nil {
			return uuid.Nil, err
		}
		return item.ID, nil
	})
	register(facilityservice.FacilityJobTaskCopySPSControllerSystemType, "sps_controller_system_types", func(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
		item, err := services.Facility.SPSControllerSystemType.CopyByID(ctx, id)
		if err != nil {
			return uuid.Nil, err
		}
		return item.ID, nil
	})
	register(facilityservice.FacilityJobTaskCopyFieldDevice, "field_devices", func(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
		item, err := services.Facility.FieldDevice.CopyByID(ctx, id)
		if err != nil {
			return uuid.Nil, err
		}
		return item.ID, nil
	})
	register(facilityservice.FacilityJobTaskCopyObjectData, "object_data", func(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
		item, err := services.Facility.ObjectData.CopyByID(ctx, id)
		if err != nil {
			return uuid.Nil, err
		}
		return item.ID, nil
	})
}
