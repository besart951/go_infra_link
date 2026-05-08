package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/service/changecapture"
)

type TxRunner = transaction.Runner

type Config struct {
	TxRunner       TxRunner
	TxRepositories func(tx transaction.UnitOfWork) (Repositories, error)
	ChangeRecorder changecapture.Recorder
}

type txCoordinator struct {
	boundary transaction.Boundary[*Services]
}

func newTxCoordinator(cfg Config) txCoordinator {
	var factory transaction.Factory[*Services]
	if cfg.TxRepositories != nil {
		factory = func(tx transaction.UnitOfWork) (*Services, error) {
			repos, err := cfg.TxRepositories(tx)
			if err != nil {
				return nil, err
			}
			return NewServices(repos, Config{ChangeRecorder: cfg.ChangeRecorder}), nil
		}
	}

	return txCoordinator{
		boundary: transaction.NewBoundary(cfg.TxRunner, factory),
	}
}

type facilityTx[TService any] struct {
	operation transaction.Operation[*Services, TService]
}

func newFacilityTx[TService any](
	tx txCoordinator,
	current TService,
	selectService func(*Services) TService,
) facilityTx[TService] {
	return facilityTx[TService]{
		operation: transaction.Bind(tx.boundary, current, selectService),
	}
}

func (tx facilityTx[TService]) run(ctx context.Context, fn func(context.Context, TService) error) error {
	return tx.operation.Run(ctx, fn)
}

func runWithFacilityTxResult[TResult any, TService any](
	ctx context.Context,
	tx facilityTx[TService],
	fn func(context.Context, TService) (TResult, error),
) (TResult, error) {
	return transaction.RunResult(ctx, tx.operation, fn)
}
