package sqlutil

import (
	"fmt"
	"strings"
)

type Dialect string

const (
	DialectPostgres Dialect = "postgres"
	DialectSQLite   Dialect = "sqlite"
	DialectMySQL    Dialect = "mysql"
)

func DialectFromDriver(driver string) Dialect {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "postgres", "pg", "postgresql", "pgx":
		return DialectPostgres
	case "sqlite", "sqlite3":
		return DialectSQLite
	case "mysql", "mariadb":
		return DialectMySQL
	default:
		return Dialect(driver)
	}
}

// Rebind converts a query written with '?' placeholders into the
// dialect-specific placeholder format.
func Rebind(d Dialect, query string) string {
	if d != DialectPostgres {
		return query
	}

	var b strings.Builder
	b.Grow(len(query) + 8)
	arg := 1
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			b.WriteString(fmt.Sprintf("$%d", arg))
			arg++
			continue
		}
		b.WriteByte(query[i])
	}
	return b.String()
}

func LikeOperator(d Dialect) string {
	if d == DialectPostgres {
		return "ILIKE"
	}
	return "LIKE"
}

func Placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	if n == 1 {
		return "?"
	}
	return strings.TrimRight(strings.Repeat("?,", n), ",")
}
