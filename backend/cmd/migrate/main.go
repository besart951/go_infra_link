package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func usage() {
	_, _ = fmt.Fprintln(os.Stderr, "Usage:")
	_, _ = fmt.Fprintln(os.Stderr, "  go run ./cmd/migrate [flags] <command> [args]")
	_, _ = fmt.Fprintln(os.Stderr, "")
	_, _ = fmt.Fprintln(os.Stderr, "Commands:")
	_, _ = fmt.Fprintln(os.Stderr, "  up                 Apply all up migrations")
	_, _ = fmt.Fprintln(os.Stderr, "  down <n>            Roll back n migrations")
	_, _ = fmt.Fprintln(os.Stderr, "  goto <version>      Migrate to a specific version")
	_, _ = fmt.Fprintln(os.Stderr, "  version             Print current migration version")
	_, _ = fmt.Fprintln(os.Stderr, "")
	_, _ = fmt.Fprintln(os.Stderr, "Flags:")
	flag.PrintDefaults()
}

func main() {
	var (
		migrationsPath = flag.String("path", "./migrations", "Path to migration files")
		dsn            = flag.String("database", os.Getenv("DATABASE_URL"), "Database DSN (or set DATABASE_URL)")
	)
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(2)
	}
	cmd := strings.ToLower(flag.Arg(0))

	if *dsn == "" {
		_, _ = fmt.Fprintln(os.Stderr, "DATABASE_URL is required (or pass -database)")
		os.Exit(2)
	}

	absPath, err := filepath.Abs(*migrationsPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "migrations path:", err)
		os.Exit(1)
	}

	m, err := migrate.New("file://"+filepath.ToSlash(absPath), *dsn)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "migrate init:", err)
		os.Exit(1)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			_, _ = fmt.Fprintln(os.Stderr, "migrate source close:", srcErr)
		}
		if dbErr != nil {
			_, _ = fmt.Fprintln(os.Stderr, "migrate db close:", dbErr)
		}
	}()

	switch cmd {
	case "up":
		err = m.Up()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			_, _ = fmt.Fprintln(os.Stderr, "migrate up:", err)
			os.Exit(1)
		}
		_, _ = fmt.Fprintln(os.Stdout, "ok")
	case "down":
		if flag.NArg() < 2 {
			_, _ = fmt.Fprintln(os.Stderr, "down requires <n>")
			os.Exit(2)
		}
		var n uint
		_, scanErr := fmt.Sscanf(flag.Arg(1), "%d", &n)
		if scanErr != nil {
			_, _ = fmt.Fprintln(os.Stderr, "invalid n:", scanErr)
			os.Exit(2)
		}
		err = m.Steps(-int(n))
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			_, _ = fmt.Fprintln(os.Stderr, "migrate down:", err)
			os.Exit(1)
		}
		_, _ = fmt.Fprintln(os.Stdout, "ok")
	case "goto":
		if flag.NArg() < 2 {
			_, _ = fmt.Fprintln(os.Stderr, "goto requires <version>")
			os.Exit(2)
		}
		var v uint
		_, scanErr := fmt.Sscanf(flag.Arg(1), "%d", &v)
		if scanErr != nil {
			_, _ = fmt.Fprintln(os.Stderr, "invalid version:", scanErr)
			os.Exit(2)
		}
		err = m.Migrate(v)
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			_, _ = fmt.Fprintln(os.Stderr, "migrate goto:", err)
			os.Exit(1)
		}
		_, _ = fmt.Fprintln(os.Stdout, "ok")
	case "version":
		v, dirty, verr := m.Version()
		if verr != nil {
			if errors.Is(verr, migrate.ErrNilVersion) {
				_, _ = fmt.Fprintln(os.Stdout, "version: none")
				return
			}
			_, _ = fmt.Fprintln(os.Stderr, "version:", verr)
			os.Exit(1)
		}
		_, _ = fmt.Fprintf(os.Stdout, "version: %d (dirty=%v)\n", v, dirty)
	default:
		usage()
		os.Exit(2)
	}
}
