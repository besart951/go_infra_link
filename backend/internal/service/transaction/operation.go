package transaction

import "gorm.io/gorm"

type Runner func(func(tx *gorm.DB) error) error

type Factory[TBundle any] func(tx *gorm.DB) (TBundle, error)

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

func (op Operation[TBundle, TService]) Run(fn func(TService) error) error {
	_, err := RunResult(op, func(service TService) (struct{}, error) {
		return struct{}{}, fn(service)
	})
	return err
}

func RunResult[TBundle any, TService any, TResult any](
	op Operation[TBundle, TService],
	fn func(TService) (TResult, error),
) (TResult, error) {
	var zero TResult
	if op.boundary.runner == nil || op.boundary.factory == nil {
		return fn(op.current)
	}

	var result TResult
	err := op.boundary.runner(func(gormTx *gorm.DB) error {
		txBundle, buildErr := op.boundary.factory(gormTx)
		if buildErr != nil {
			return buildErr
		}

		value, runErr := fn(op.selectService(txBundle))
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
