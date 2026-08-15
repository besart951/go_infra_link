package facility

import "context"

type copyProgressReporter func(progress int, stage string)

type copyProgressReporterContextKey struct{}

func withCopyProgressReporter(ctx context.Context, report copyProgressReporter) context.Context {
	if report == nil {
		return ctx
	}
	return context.WithValue(ctx, copyProgressReporterContextKey{}, report)
}

func reportCopyProgress(ctx context.Context, progress int, stage string) {
	report, _ := ctx.Value(copyProgressReporterContextKey{}).(copyProgressReporter)
	if report == nil {
		return
	}
	report(progress, stage)
}
