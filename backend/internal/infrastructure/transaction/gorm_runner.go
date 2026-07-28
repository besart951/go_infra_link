package transaction

import (
	"context"
	"database/sql"
	"fmt"

	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/repository/collaborationoutbox"
	"gorm.io/gorm"
)

func NewGormRunner(db *gorm.DB) apptransaction.Runner {
	return gormRunner{db: db}.Run
}

// NewGormRunnerWithIsolation creates a transaction runner with an explicit
// database isolation level. It is used by paged read/write workflows that need
// one source snapshot across every statement.
func NewGormRunnerWithIsolation(
	db *gorm.DB,
	isolation sql.IsolationLevel,
) apptransaction.Runner {
	return gormRunner{db: db, isolation: isolation}.Run
}

func GormDB(unit apptransaction.UnitOfWork) (*gorm.DB, error) {
	tx, ok := unit.(*gorm.DB)
	if !ok || tx == nil {
		return nil, fmt.Errorf("unsupported transaction unit %T", unit)
	}
	return tx, nil
}

type gormRunner struct {
	db        *gorm.DB
	isolation sql.IsolationLevel
}

func (r gormRunner) Run(ctx context.Context, run func(context.Context, apptransaction.UnitOfWork) error) error {
	if r.db == nil {
		return fmt.Errorf("transaction database is nil")
	}
	if run == nil {
		return fmt.Errorf("transaction callback is nil")
	}
	runTransaction := func(tx *gorm.DB) error {
		runCtx := domainCollaboration.WithOutboxStore(
			ctx,
			collaborationoutbox.NewStore(tx),
		)
		return run(runCtx, tx)
	}
	if r.isolation == sql.LevelDefault {
		return r.db.WithContext(ctx).Transaction(runTransaction)
	}
	return r.db.WithContext(ctx).Transaction(
		runTransaction,
		&sql.TxOptions{Isolation: r.isolation},
	)
}
