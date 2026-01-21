package sqlutil

import "database/sql"

// ColumnExists checks whether the given table contains the given column.
// It supports Postgres, SQLite and MySQL.
func ColumnExists(db *sql.DB, dialect Dialect, tableName, columnName string) (bool, error) {
	switch dialect {
	case DialectSQLite:
		// PRAGMA table_info(<table>) returns: cid, name, type, notnull, dflt_value, pk
		rows, err := db.Query("PRAGMA table_info(" + tableName + ")")
		if err != nil {
			return false, err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var cid int
			var name string
			var typ sql.NullString
			var notnull int
			var dflt sql.NullString
			var pk int
			if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
				return false, err
			}
			if name == columnName {
				return true, nil
			}
		}
		if err := rows.Err(); err != nil {
			return false, err
		}
		return false, nil

	case DialectPostgres:
		q := "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = ? AND column_name = ? LIMIT 1"
		q = Rebind(dialect, q)
		var one int
		err := db.QueryRow(q, tableName, columnName).Scan(&one)
		if err == nil {
			return true, nil
		}
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err

	case DialectMySQL:
		q := "SELECT 1 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ? LIMIT 1"
		q = Rebind(dialect, q)
		var one int
		err := db.QueryRow(q, tableName, columnName).Scan(&one)
		if err == nil {
			return true, nil
		}
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	default:
		// Unknown dialect: best-effort failure.
		return false, nil
	}
}
