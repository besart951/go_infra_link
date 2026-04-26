package requestutil

import (
	"context"
	"errors"
)

func IsCanceled(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func IsRequestCanceled(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	return IsCanceled(ctx.Err())
}

func ShouldSuppressErrorResponse(ctx context.Context, err error) bool {
	return IsCanceled(err) || IsRequestCanceled(ctx)
}
