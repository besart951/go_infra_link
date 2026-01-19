package main

import (
	"fmt"
	"os"

	"github.com/besart951/go_infra_link/backend/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
