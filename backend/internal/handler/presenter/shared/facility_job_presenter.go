package shared

import (
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
)

// ToFacilityJobResponse is the single HTTP representation of an asynchronous
// hierarchy-copy job, independent of whether it started globally or in a
// project context.
func ToFacilityJobResponse(job facilityservice.FacilityJob) dto.FacilityJobResponse {
	return dto.FacilityJobResponse{
		JobID:        job.ID,
		Kind:         string(job.Kind),
		Type:         string(job.Type),
		Class:        string(job.Class),
		Status:       string(job.Status),
		Progress:     job.Progress,
		Stage:        job.Stage,
		Error:        job.Error,
		Attempts:     job.Attempts,
		Processed:    job.Processed,
		Total:        job.Total,
		SuccessCount: job.Succeeded,
		FailureCount: job.Failed,
		Retryable:    job.Retryable,
		Result:       job.Result,
		CreatedAt:    job.CreatedAt,
		UpdatedAt:    job.UpdatedAt,
		CompletedAt:  job.CompletedAt,
	}
}
