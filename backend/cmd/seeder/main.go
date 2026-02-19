package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

type fkDependency struct {
	ChildTable  string
	ParentTable string
}

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func loadSnapshot(path string) (*infralinkBasicSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var payload infralinkBasicSnapshot
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if len(payload.Tables) == 0 {
		return nil, fmt.Errorf("snapshot has no tables: %s", path)
	}
	if strings.TrimSpace(payload.Schema) == "" {
		payload.Schema = "public"
	}

	return &payload, nil
}

func loadExistingTables(database *gorm.DB, schema string) (map[string]struct{}, error) {
	rows, err := database.Raw(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_type = 'BASE TABLE'
	`, schema).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make(map[string]struct{})
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables[tableName] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func loadForeignKeyDependencies(database *gorm.DB, schema string) ([]fkDependency, error) {
	rows, err := database.Raw(`
		SELECT
			tc.table_name AS child_table,
			ccu.table_name AS parent_table
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_schema = ?
	`, schema).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deps := make([]fkDependency, 0)
	for rows.Next() {
		var dep fkDependency
		if err := rows.Scan(&dep.ChildTable, &dep.ParentTable); err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return deps, nil
}

func computeInsertOrder(tableSet map[string]struct{}, deps []fkDependency) []string {
	adjacency := make(map[string][]string, len(tableSet))
	inDegree := make(map[string]int, len(tableSet))

	for tableName := range tableSet {
		adjacency[tableName] = []string{}
		inDegree[tableName] = 0
	}

	for _, dep := range deps {
		_, childExists := tableSet[dep.ChildTable]
		_, parentExists := tableSet[dep.ParentTable]
		if !childExists || !parentExists || dep.ChildTable == dep.ParentTable {
			continue
		}
		adjacency[dep.ParentTable] = append(adjacency[dep.ParentTable], dep.ChildTable)
		inDegree[dep.ChildTable]++
	}

	ready := make([]string, 0)
	for tableName, degree := range inDegree {
		if degree == 0 {
			ready = append(ready, tableName)
		}
	}
	sort.Strings(ready)

	order := make([]string, 0, len(tableSet))
	for len(ready) > 0 {
		current := ready[0]
		ready = ready[1:]
		order = append(order, current)

		nextChildren := adjacency[current]
		sort.Strings(nextChildren)
		for _, child := range nextChildren {
			inDegree[child]--
			if inDegree[child] == 0 {
				ready = append(ready, child)
				sort.Strings(ready)
			}
		}
	}

	if len(order) < len(tableSet) {
		remaining := make([]string, 0)
		seen := make(map[string]struct{}, len(order))
		for _, tableName := range order {
			seen[tableName] = struct{}{}
		}
		for tableName := range tableSet {
			if _, ok := seen[tableName]; !ok {
				remaining = append(remaining, tableName)
			}
		}
		sort.Strings(remaining)
		order = append(order, remaining...)
	}

	return order
}

func seedFromSnapshot(database *gorm.DB, snapshotPath string) error {
	snapshot, err := loadSnapshot(snapshotPath)
	if err != nil {
		return err
	}

	existingTables, err := loadExistingTables(database, snapshot.Schema)
	if err != nil {
		return err
	}

	tableSet := make(map[string]struct{})
	for tableName := range snapshot.Tables {
		if _, ok := existingTables[tableName]; !ok {
			continue
		}
		tableSet[tableName] = struct{}{}
	}
	if len(tableSet) == 0 {
		return fmt.Errorf("no matching tables in DB for snapshot schema=%s", snapshot.Schema)
	}

	deps, err := loadForeignKeyDependencies(database, snapshot.Schema)
	if err != nil {
		return err
	}
	insertOrder := computeInsertOrder(tableSet, deps)

	return database.Transaction(func(tx *gorm.DB) error {
		truncateTargets := make([]string, 0, len(insertOrder))
		for _, tableName := range insertOrder {
			truncateTargets = append(truncateTargets, quoteIdentifier(snapshot.Schema)+"."+quoteIdentifier(tableName))
		}
		if err := tx.Exec("TRUNCATE TABLE " + strings.Join(truncateTargets, ", ") + " RESTART IDENTITY CASCADE").Error; err != nil {
			return err
		}

		for _, tableName := range insertOrder {
			rows := snapshot.Tables[tableName]
			if len(rows) == 0 {
				continue
			}

			rowsJSON, err := json.Marshal(rows)
			if err != nil {
				return err
			}

			query := "INSERT INTO " + quoteIdentifier(snapshot.Schema) + "." + quoteIdentifier(tableName) +
				" SELECT * FROM json_populate_recordset(NULL::" + quoteIdentifier(snapshot.Schema) + "." + quoteIdentifier(tableName) + ", ?::json)"
			if err := tx.Exec(query, string(rowsJSON)).Error; err != nil {
				return fmt.Errorf("table %s: %w", tableName, err)
			}
		}

		return nil
	})
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

	snapshotPath := filepath.Join("data", "seed", "infralink_basic.json")
	if err := seedFromSnapshot(database, snapshotPath); err != nil {
		log.Fatalf("failed seeding from %s: %v", snapshotPath, err)
	}

	log.Printf("seeded database from snapshot: %s", snapshotPath)
}
