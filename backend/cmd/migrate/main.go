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
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
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
	loadEnvFiles()

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

	driver := normalizeDriver(os.Getenv("DB_DRIVER"))
	if driver == "" {
		driver = "postgres"
	}

	resolvedPath := resolveMigrationsPath(*migrationsPath, driver)
	resolvedDSN, dsnErr := resolveDatabaseURL(*dsn, driver)
	if dsnErr != nil {
		_, _ = fmt.Fprintln(os.Stderr, dsnErr.Error())
		os.Exit(2)
	}

	absPath, err := filepath.Abs(resolvedPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "migrations path:", err)
		os.Exit(1)
	}

	m, err := migrate.New("file://"+filepath.ToSlash(absPath), resolvedDSN)
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

func loadEnvFiles() {
	// .env files are optional; ignore missing/unreadable files.
	_ = godotenv.Load()
	_ = godotenv.Load("configs/.env")
	_ = godotenv.Load("../.env")
}

func normalizeDriver(driver string) string {
	driver = strings.ToLower(strings.TrimSpace(driver))
	switch driver {
	case "sqlite3", "sqlite":
		return "sqlite"
	case "postgres", "pg", "postgresql", "pgx":
		return "postgres"
	default:
		return driver
	}
}

func resolveMigrationsPath(path string, driver string) string {
	// If user left the default migrations path and is using sqlite,
	// automatically switch to dialect-specific migrations if present.
	if normalizeSlashes(path) == "./migrations" && driver == "sqlite" {
		candidate := filepath.Join(path, "sqlite")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return path
}

func normalizeSlashes(p string) string {
	return filepath.ToSlash(strings.TrimSpace(p))
}

func resolveDatabaseURL(raw string, driver string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("DB_DSN"))
	}

	if raw == "" {
		return "", fmt.Errorf("DATABASE_URL is required (or pass -database)")
	}

	if driver != "sqlite" {
		// Postgres/MySQL: assume caller provided a working DSN.
		return raw, nil
	}

	// SQLite: allow a plain file path in DATABASE_URL/DB_DSN and convert to sqlite:///...
	if strings.HasPrefix(raw, "sqlite://") {
		return raw, nil
	}

	// Ensure directory exists.
	absFile, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("invalid sqlite database path: %w", err)
	}
	if dir := filepath.Dir(absFile); dir != "." {
		if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
			return "", fmt.Errorf("create sqlite db dir: %w", mkErr)
		}
	}

	// Use sqlite:///C:/... form on Windows.
	return "sqlite:///" + filepath.ToSlash(absFile), nil
}
