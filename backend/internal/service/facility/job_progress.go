package facility

import "context"

type facilityJobReporterContextKey struct{}

func WithFacilityJobReporter(ctx context.Context, reporter FacilityJobReporter) context.Context {
	if reporter == nil {
		return ctx
	}
	return context.WithValue(ctx, facilityJobReporterContextKey{}, reporter)
}

func reportCopyProgress(ctx context.Context, progress int, stage string) {
	reporter, _ := ctx.Value(facilityJobReporterContextKey{}).(FacilityJobReporter)
	if reporter != nil {
		reporter.Report(FacilityJobProgress{Progress: progress, Stage: stage})
	}
}
