package transaction

import (
	"context"
	"fmt"

	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"gorm.io/gorm"
)

func NewGormRunner(db *gorm.DB) apptransaction.Runner {
	return gormRunner{db: db}.Run
}

func GormDB(unit apptransaction.UnitOfWork) (*gorm.DB, error) {
	tx, ok := unit.(*gorm.DB)
	if !ok || tx == nil {
		return nil, fmt.Errorf("unsupported transaction unit %T", unit)
	}
	return tx, nil
}

type gormRunner struct {
	db *gorm.DB
}

func (r gormRunner) Run(ctx context.Context, run func(context.Context, apptransaction.UnitOfWork) error) error {
	if r.db == nil {
		return fmt.Errorf("transaction database is nil")
	}
	if run == nil {
		return fmt.Errorf("transaction callback is nil")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return run(ctx, tx)
	})
}
