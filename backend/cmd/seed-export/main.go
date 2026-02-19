package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/db"
	"gorm.io/gorm"
)

type infralinkBasicSnapshot struct {
	ExportedAt string                      `json:"exported_at"`
	Database   string                      `json:"database"`
	Schema     string                      `json:"schema"`
	Tables     map[string][]map[string]any `json:"tables"`
	Counts     map[string]int              `json:"counts"`
}

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func normalizeSQLValue(value any) any {
	switch typed := value.(type) {
	case []byte:
		return string(typed)
	case time.Time:
		return typed.UTC().Format(time.RFC3339Nano)
	default:
		return typed
	}
}

func readAllRows(database *gorm.DB, schema, table string) ([]map[string]any, error) {
	query := "SELECT * FROM " + quoteIdentifier(schema) + "." + quoteIdentifier(table)
	rows, err := database.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	items := make([]map[string]any, 0)
	for rows.Next() {
		values := make([]any, len(columns))
		valuePointers := make([]any, len(columns))
		for i := range values {
			valuePointers[i] = &values[i]
		}

		if err := rows.Scan(valuePointers...); err != nil {
			return nil, err
		}

		item := make(map[string]any, len(columns))
		for i, column := range columns {
			item[column] = normalizeSQLValue(values[i])
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func writeJSONFile(path string, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func exportInfralinkBasic(database *gorm.DB, seedDir string) error {
	schema := "public"

	tableRows, err := database.Raw(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name ASC
	`, schema).Rows()
	if err != nil {
		return err
	}
	defer tableRows.Close()

	tableNames := make([]string, 0)
	for tableRows.Next() {
		var tableName string
		if err := tableRows.Scan(&tableName); err != nil {
			return err
		}
		tableNames = append(tableNames, tableName)
	}
	if err := tableRows.Err(); err != nil {
		return err
	}

	allTables := make(map[string][]map[string]any, len(tableNames))
	counts := make(map[string]int, len(tableNames))
	for _, tableName := range tableNames {
		rows, err := readAllRows(database, schema, tableName)
		if err != nil {
			return err
		}
		allTables[tableName] = rows
		counts[tableName] = len(rows)
	}

	payload := infralinkBasicSnapshot{
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Database:   "postgres",
		Schema:     schema,
		Tables:     allTables,
		Counts:     counts,
	}

	outputPath := filepath.Join(seedDir, "infralink_basic.json")
	if err := writeJSONFile(outputPath, payload); err != nil {
		return err
	}

	log.Printf("wrote complete DB snapshot: %s (tables=%d)", outputPath, len(tableNames))
	return nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	seedDir := filepath.Join("data", "seed")
	if err := os.MkdirAll(seedDir, 0o755); err != nil {
		log.Fatalf("failed creating seed dir: %v", err)
	}

	if err := exportInfralinkBasic(database, seedDir); err != nil {
		log.Fatalf("failed exporting infralink_basic: %v", err)
	}
}
