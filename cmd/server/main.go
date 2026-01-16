package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/besart951/go_infra_link/internal/database"
	"github.com/besart951/go_infra_link/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env if present (development convenience)
	_ = godotenv.Load(".env")

	// create logger
	l := logger.New(os.Getenv("APP_LOG_LEVEL"))

	// build db config
	cfg := database.NewConfigFromEnv(l)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	dbConn, err := database.Open(ctx, cfg)
	if err != nil {
		l.Fatal().Err(err).Msg("cannot open db")
	}
	defer dbConn.Close()

	// TODO: wire repositories -> usecases -> http handlers
	l.Info().Msg("App started")
	// wait for shutdown signal
	<-ctx.Done()
	l.Info().Msg("Shutting down")
	time.Sleep(200 * time.Millisecond)
}
