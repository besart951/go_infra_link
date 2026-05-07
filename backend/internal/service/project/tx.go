package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/application/transaction"
)

type TxRunner = transaction.Runner

type Config struct {
	TxRunner       TxRunner
	TxDependencies func(tx transaction.UnitOfWork) (Dependencies, error)
}

type txCoordinator struct {
	boundary transaction.Boundary[*Services]
}

func newTxCoordinator(cfg Config) txCoordinator {
	var factory transaction.Factory[*Services]
	if cfg.TxDependencies != nil {
		factory = func(tx transaction.UnitOfWork) (*Services, error) {
			deps, err := cfg.TxDependencies(tx)
			if err != nil {
				return nil, err
			}
			return NewServices(deps), nil
		}
	}

	return txCoordinator{
		boundary: transaction.NewBoundary(cfg.TxRunner, factory),
	}
}

type projectTx[TService any] struct {
	operation transaction.Operation[*Services, TService]
}

func newProjectTx[TService any](
	tx txCoordinator,
	current TService,
	selectService func(*Services) TService,
) projectTx[TService] {
	return projectTx[TService]{
		operation: transaction.Bind(tx.boundary, current, selectService),
	}
}

func (tx projectTx[TService]) run(ctx context.Context, fn func(context.Context, TService) error) error {
	return tx.operation.Run(ctx, fn)
}

func runProjectTxResult[TResult any, TService any](
	ctx context.Context,
	tx projectTx[TService],
	fn func(context.Context, TService) (TResult, error),
) (TResult, error) {
	return transaction.RunResult(ctx, tx.operation, fn)
}
