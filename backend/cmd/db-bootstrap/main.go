package main

import (
	"fmt"
	"os"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "config load:", err)
		os.Exit(1)
	}

	if err := db.Bootstrap(cfg.DBConfig); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "db bootstrap:", err)
		os.Exit(1)
	}
}
