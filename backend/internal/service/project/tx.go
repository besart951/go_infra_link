package project

import (
	"github.com/besart951/go_infra_link/backend/internal/service/transaction"
	"gorm.io/gorm"
)

type TxRunner = transaction.Runner

type Config struct {
	TxRunner       TxRunner
	TxDependencies func(tx *gorm.DB) (Dependencies, error)
}

type txCoordinator struct {
	boundary transaction.Boundary[*Services]
}

func newTxCoordinator(cfg Config) txCoordinator {
	var factory transaction.Factory[*Services]
	if cfg.TxDependencies != nil {
		factory = func(tx *gorm.DB) (*Services, error) {
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

func (tx projectTx[TService]) run(fn func(TService) error) error {
	return tx.operation.Run(fn)
}

func runProjectTxResult[TResult any, TService any](
	tx projectTx[TService],
	fn func(TService) (TResult, error),
) (TResult, error) {
	return transaction.RunResult(tx.operation, fn)
}
