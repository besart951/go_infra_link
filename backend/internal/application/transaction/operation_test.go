package transaction

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestOperationRunsCurrentServiceWithoutConfiguredBoundary(t *testing.T) {
	ctx := context.Background()
	current := testService{name: "current"}
	op := Bind(NewBoundary[testBundle](nil, nil), current, func(bundle testBundle) testService {
		return bundle.service
	})

	err := op.Run(ctx, func(runCtx context.Context, service testService) error {
		if runCtx != ctx {
			return fmt.Errorf("expected original context")
		}
		if service.name != "current" {
			return fmt.Errorf("expected current service, got %q", service.name)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected operation to succeed, got %v", err)
	}
}

func TestOperationRunsTransactionalServiceAndPropagatesContext(t *testing.T) {
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "expected")
	unit := struct{ name string }{name: "tx-unit"}
	runnerCalls := 0
	factoryCalls := 0

	boundary := NewBoundary[testBundle](
		func(runCtx context.Context, run func(context.Context, UnitOfWork) error) error {
			runnerCalls++
			return run(runCtx, unit)
		},
		func(got UnitOfWork) (testBundle, error) {
			factoryCalls++
			if got != unit {
				return testBundle{}, fmt.Errorf("expected unit %v, got %v", unit, got)
			}
			return testBundle{service: testService{name: "tx"}}, nil
		},
	)
	op := Bind(boundary, testService{name: "current"}, func(bundle testBundle) testService {
		return bundle.service
	})

	got, err := RunResult(ctx, op, func(runCtx context.Context, service testService) (string, error) {
		if got := runCtx.Value(contextKey{}); got != "expected" {
			return "", fmt.Errorf("expected context value to propagate, got %v", got)
		}
		return service.name, nil
	})

	if err != nil {
		t.Fatalf("expected operation to succeed, got %v", err)
	}
	if got != "tx" {
		t.Fatalf("expected transactional service result, got %q", got)
	}
	if runnerCalls != 1 || factoryCalls != 1 {
		t.Fatalf("expected one runner and factory call, got runner=%d factory=%d", runnerCalls, factoryCalls)
	}
}

func TestOperationReturnsZeroValueWhenTransactionalCallbackFails(t *testing.T) {
	stepErr := errors.New("step failed")
	boundary := NewBoundary[testBundle](
		func(ctx context.Context, run func(context.Context, UnitOfWork) error) error {
			return run(ctx, nil)
		},
		func(UnitOfWork) (testBundle, error) {
			return testBundle{service: testService{name: "tx"}}, nil
		},
	)
	op := Bind(boundary, testService{name: "current"}, func(bundle testBundle) testService {
		return bundle.service
	})

	got, err := RunResult(context.Background(), op, func(context.Context, testService) (string, error) {
		return "partial", stepErr
	})

	if !errors.Is(err, stepErr) {
		t.Fatalf("expected step error, got %v", err)
	}
	if got != "" {
		t.Fatalf("expected zero value on failure, got %q", got)
	}
}

func TestOperationReturnsFactoryError(t *testing.T) {
	factoryErr := errors.New("factory failed")
	boundary := NewBoundary[testBundle](
		func(ctx context.Context, run func(context.Context, UnitOfWork) error) error {
			return run(ctx, nil)
		},
		func(UnitOfWork) (testBundle, error) {
			return testBundle{}, factoryErr
		},
	)
	op := Bind(boundary, testService{name: "current"}, func(bundle testBundle) testService {
		return bundle.service
	})

	err := op.Run(context.Background(), func(context.Context, testService) error {
		return nil
	})

	if !errors.Is(err, factoryErr) {
		t.Fatalf("expected factory error, got %v", err)
	}
}

type testBundle struct {
	service testService
}

type testService struct {
	name string
}
