package transaction

import "context"

// UnitOfWork is intentionally opaque to application services.
// Infrastructure adapters decide which concrete handle backs a transaction.
type UnitOfWork interface{}

type Runner func(context.Context, func(context.Context, UnitOfWork) error) error

type Factory[TBundle any] func(UnitOfWork) (TBundle, error)

type Boundary[TBundle any] struct {
	runner  Runner
	factory Factory[TBundle]
}

func NewBoundary[TBundle any](runner Runner, factory Factory[TBundle]) Boundary[TBundle] {
	return Boundary[TBundle]{
		runner:  runner,
		factory: factory,
	}
}

type Operation[TBundle any, TService any] struct {
	boundary      Boundary[TBundle]
	current       TService
	selectService func(TBundle) TService
}

func Bind[TBundle any, TService any](
	boundary Boundary[TBundle],
	current TService,
	selectService func(TBundle) TService,
) Operation[TBundle, TService] {
	return Operation[TBundle, TService]{
		boundary:      boundary,
		current:       current,
		selectService: selectService,
	}
}

func (op Operation[TBundle, TService]) Run(ctx context.Context, fn func(context.Context, TService) error) error {
	_, err := RunResult(ctx, op, func(runCtx context.Context, service TService) (struct{}, error) {
		return struct{}{}, fn(runCtx, service)
	})
	return err
}

func RunResult[TBundle any, TService any, TResult any](
	ctx context.Context,
	op Operation[TBundle, TService],
	fn func(context.Context, TService) (TResult, error),
) (TResult, error) {
	var zero TResult
	if op.boundary.runner == nil || op.boundary.factory == nil {
		return fn(ctx, op.current)
	}

	var result TResult
	err := op.boundary.runner(ctx, func(runCtx context.Context, unit UnitOfWork) error {
		if runCtx == nil {
			runCtx = ctx
		}
		txBundle, buildErr := op.boundary.factory(unit)
		if buildErr != nil {
			return buildErr
		}

		value, runErr := fn(runCtx, op.selectService(txBundle))
		if runErr != nil {
			return runErr
		}

		result = value
		return nil
	})
	if err != nil {
		return zero, err
	}

	return result, nil
}
