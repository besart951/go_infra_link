package facility

import "gorm.io/gorm"

type TxRunner func(func(tx *gorm.DB) error) error

type Config struct {
	TxRunner  TxRunner
	TxFactory func(tx *gorm.DB) (*Services, error)
}

type txCoordinator struct {
	runner  TxRunner
	factory func(tx *gorm.DB) (*Services, error)
}

func newTxCoordinator(cfg Config) txCoordinator {
	return txCoordinator{
		runner:  cfg.TxRunner,
		factory: cfg.TxFactory,
	}
}

func runWithFacilityTx[TService any](
	tx txCoordinator,
	current TService,
	selectService func(*Services) TService,
	fn func(TService) error,
) error {
	_, err := runWithFacilityTxResult(tx, current, selectService, func(service TService) (struct{}, error) {
		return struct{}{}, fn(service)
	})
	return err
}

func runWithFacilityTxResult[TResult any, TService any](
	tx txCoordinator,
	current TService,
	selectService func(*Services) TService,
	fn func(TService) (TResult, error),
) (TResult, error) {
	var zero TResult
	if tx.runner == nil || tx.factory == nil {
		return fn(current)
	}

	var result TResult
	err := tx.runner(func(gormTx *gorm.DB) error {
		txServices, buildErr := tx.factory(gormTx)
		if buildErr != nil {
			return buildErr
		}

		value, runErr := fn(selectService(txServices))
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
