package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/db"
	"github.com/besart951/go_infra_link/backend/internal/repository/historysql"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "history v2 backfill:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	database, err := db.Connect(cfg.DBConfig)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return backfill(ctx, historysql.NewStore(database))
}

func backfill(ctx context.Context, store *historysql.Store) error {
	request := historysql.V2BackfillRequest{Limit: 500}
	for {
		result, err := store.BackfillV2(ctx, request)
		if err != nil {
			return err
		}
		if result.Processed > 0 {
			_, _ = fmt.Fprintf(os.Stdout, "backfilled=%d cursor=%s/%s\n", result.Processed, result.NextOccurredAt.Format("2006-01-02T15:04:05.999999999Z07:00"), result.NextID)
		}
		if result.Done {
			if err := store.VerifyAndEnableV2Reads(ctx); err != nil {
				return fmt.Errorf("verify and enable V2 reader: %w", err)
			}
			_, _ = fmt.Fprintln(os.Stdout, "history V2 reader enabled")
			return nil
		}
		request.AfterOccurredAt = result.NextOccurredAt
		request.AfterID = result.NextID
	}
}
