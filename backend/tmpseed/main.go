package main

import (
	"encoding/json"
	"flag"
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
	ExportedAt string                     `json:"exported_at"`
	Database   string                     `json:"database"`
	Schema     string                     `json:"schema"`
	Tables     map[string][]map[string]any `json:"tables"`
	Counts     map[string]int             `json:"counts"`
}

type fkDependency struct { ChildTable string; ParentTable string }

func quoteIdentifier(value string) string { return `"` + strings.ReplaceAll(value, `"`, `""`) + `"` }

func loadSnapshot(path string) (*infralinkBasicSnapshot, error) {
	data, err := os.ReadFile(path); if err != nil { return nil, err }
	var payload infralinkBasicSnapshot
	if err := json.Unmarshal(data, &payload); err != nil { return nil, err }
	if len(payload.Tables) == 0 { return nil, fmt.Errorf("snapshot has no tables: %s", path) }
	if strings.TrimSpace(payload.Schema) == "" { payload.Schema = "public" }
	return &payload, nil
}

func loadExistingTables(database *gorm.DB, schema string) (map[string]struct{}, error) {
	rows, err := database.Raw(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_type = 'BASE TABLE'
	`, schema).Rows(); if err != nil { return nil, err }
	defer rows.Close()
	tables := make(map[string]struct{})
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil { return nil, err }
		tables[tableName] = struct{}{}
	}
	if err := rows.Err(); err != nil { return nil, err }
	return tables, nil
}

func loadForeignKeyDependencies(database *gorm.DB, schema string) ([]fkDependency, error) {
	rows, err := database.Raw(`
		SELECT tc.table_name AS child_table, ccu.table_name AS parent_table
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON tc.constraint_name = kcu.constraint_name
		 AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage ccu
		  ON ccu.constraint_name = tc.constraint_name
		 AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_schema = ?
	`, schema).Rows(); if err != nil { return nil, err }
	defer rows.Close()
	deps := make([]fkDependency,0)
	for rows.Next() {
		var dep fkDependency
		if err := rows.Scan(&dep.ChildTable, &dep.ParentTable); err != nil { return nil, err }
		deps = append(deps, dep)
	}
	if err := rows.Err(); err != nil { return nil, err }
	return deps,nil
}

func computeInsertOrder(tableSet map[string]struct{}, deps []fkDependency) []string {
	adj:=make(map[string][]string,len(tableSet)); deg:=make(map[string]int,len(tableSet))
	for t:= range tableSet { adj[t]=[]string{}; deg[t]=0 }
	for _,d:= range deps {
		if _,ok:=tableSet[d.ChildTable]; !ok { continue }
		if _,ok:=tableSet[d.ParentTable]; !ok { continue }
		if d.ChildTable==d.ParentTable { continue }
		adj[d.ParentTable]=append(adj[d.ParentTable], d.ChildTable)
		deg[d.ChildTable]++
	}
	ready:=make([]string,0)
	for t,v:= range deg { if v==0 { ready=append(ready,t) } }
	sort.Strings(ready)
	order:=make([]string,0,len(tableSet))
	for len(ready)>0 {
		cur:=ready[0]; ready=ready[1:]; order=append(order,cur)
		children:=adj[cur]; sort.Strings(children)
		for _,ch:= range children {
			deg[ch]--
			if deg[ch]==0 { ready=append(ready,ch); sort.Strings(ready) }
		}
	}
	if len(order)<len(tableSet) {
		seen:=make(map[string]struct{},len(order)); rem:=make([]string,0)
		for _,t:= range order { seen[t]=struct{}{} }
		for t:= range tableSet { if _,ok:=seen[t]; !ok { rem=append(rem,t) } }
		sort.Strings(rem); order=append(order, rem...)
	}
	return order
}

func seedFromSnapshot(database *gorm.DB, snapshotPath string) error {
	snap, err := loadSnapshot(snapshotPath); if err != nil { return err }
	existing, err := loadExistingTables(database,snap.Schema); if err != nil { return err }
	tableSet:=make(map[string]struct{})
	for t:= range snap.Tables { if _,ok:=existing[t]; ok { tableSet[t]=struct{}{} } }
	if len(tableSet)==0 { return fmt.Errorf("no matching tables in DB for schema=%s", snap.Schema) }
	deps, err := loadForeignKeyDependencies(database,snap.Schema); if err != nil { return err }
	order := computeInsertOrder(tableSet,deps)

	return database.Transaction(func(tx *gorm.DB) error {
		trs := make([]string,0,len(order))
		for _,t := range order { trs=append(trs, quoteIdentifier(snap.Schema)+"."+quoteIdentifier(t)) }
		if len(trs)>0 {
			if err := tx.Exec("TRUNCATE TABLE "+strings.Join(trs, ", ")+" RESTART IDENTITY CASCADE").Error; err != nil { return err }
		}
		for _, tableName := range order {
			rows := snap.Tables[tableName]
			if len(rows)==0 { continue }
			b, err := json.Marshal(rows); if err != nil { return err }
			q := "INSERT INTO "+quoteIdentifier(snap.Schema)+"."+quoteIdentifier(tableName)+" SELECT * FROM json_populate_recordset(NULL::"+quoteIdentifier(snap.Schema)+"."+quoteIdentifier(tableName)+", ?::json)"
			if err := tx.Exec(q, string(b)).Error; err != nil { return fmt.Errorf("table %s: %w", tableName, err) }
		}
		return nil
	})
}

func main() {
	snapshot := flag.String("snapshot", filepath.Join("data","seed","infralink_basic.json"), "snapshot path")
	flag.Parse()
	cfg, err := config.Load(); if err != nil { log.Fatalf("failed to load config: %v", err) }
	dbconn, err := db.Connect(cfg.DBConfig); if err != nil { log.Fatalf("failed to connect db: %v", err) }
	if err := seedFromSnapshot(dbconn, *snapshot); err != nil { log.Fatalf("failed seeding from %s: %v", *snapshot, err) }
	log.Printf("seeded database from snapshot: %s", *snapshot)
}
