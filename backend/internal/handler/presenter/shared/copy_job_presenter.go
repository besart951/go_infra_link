package shared

import (
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
)

// ToCopyJobResponse is the single HTTP representation of an asynchronous
// hierarchy-copy job, independent of whether it started globally or in a
// project context.
func ToCopyJobResponse(job facilityservice.CopyJob) dto.CopyJobResponse {
	return dto.CopyJobResponse{
		JobID:     job.ID,
		Kind:      string(job.Kind),
		Status:    string(job.Status),
		Progress:  job.Progress,
		Stage:     job.Stage,
		Error:     job.Error,
		CreatedAt: job.CreatedAt,
		UpdatedAt: job.UpdatedAt,
	}
}
