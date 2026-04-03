package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/config"
	applogger "github.com/besart951/go_infra_link/backend/pkg/logger"
)

func serveHTTP(cfg config.Config, log applogger.Logger, handler http.Handler) error {
	log.Info("Starting server", "addr", cfg.HTTPAddr, "env", cfg.AppEnv)

	server := newHTTPServer(cfg.HTTPAddr, handler)
	serverErr := startHTTPServer(server)

	return awaitShutdown(log, server, serverErr)
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func startHTTPServer(server *http.Server) <-chan error {
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()
	return serverErr
}

func awaitShutdown(log applogger.Logger, server *http.Server, serverErr <-chan error) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Server stopped unexpectedly", "err", err)
			return fmt.Errorf("server listen: %w", err)
		}
		return nil
	case <-ctx.Done():
		log.Info("Shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Graceful shutdown failed", "err", err)
		_ = server.Close()
		return fmt.Errorf("server shutdown: %w", err)
	}

	if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("Server stopped with error", "err", err)
		return fmt.Errorf("server stop: %w", err)
	}

	log.Info("Server stopped")
	return nil
}
