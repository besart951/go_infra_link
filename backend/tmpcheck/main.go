package main

import (
  "fmt"
  "strings"

  "github.com/besart951/go_infra_link/backend/internal/config"
  "github.com/besart951/go_infra_link/backend/internal/db"
)

func main() {
  cfg, err := config.Load(); if err != nil { panic(err) }
  dbConn, err := db.Connect(cfg.DBConfig); if err != nil { panic(err) }

  type row struct { Child string; ChildCols string; Parent string; ParentCols string }
  rows := []row{}
  if err := dbConn.Raw(`
SELECT
  cls.relname AS child,
  string_agg(att2.attname, ',') AS child_cols,
  refcls.relname AS parent,
  string_agg(att.attname, ',') AS parent_cols
FROM pg_constraint con
JOIN pg_class cls ON cls.oid = con.conrelid
JOIN pg_namespace nsp ON nsp.oid = cls.relnamespace
JOIN pg_class refcls ON refcls.oid = con.confrelid
JOIN unnest(con.conkey) WITH ORDINALITY AS ck(attnum, ord) ON true
JOIN pg_attribute att2 ON att2.attrelid = cls.oid AND att2.attnum = ck.attnum
JOIN unnest(con.confkey) WITH ORDINALITY AS fk(attnum, ord) ON fk.ord = ck.ord
JOIN pg_attribute att ON att.attrelid = refcls.oid AND att.attnum = fk.attnum
WHERE con.contype = 'f'
  AND nsp.nspname = 'public'
  AND cls.relname IN ('buildings','system_types','sps_controllers','control_cabinets','sps_controller_system_types','system_parts','system_part_apparats','apparats','specifications','field_devices','units','phases')
GROUP BY child,parent,con.oid
ORDER BY child,parent,child_cols;
`, ).Scan(&rows).Error; err != nil { panic(err) }

  for _, r := range rows {
    if !strings.Contains(r.ChildCols, ",") && r.ChildCols == "" {
      continue
    }
    fmt.Printf("%s (%s) -> %s (%s)\n", r.Child, r.ChildCols, r.Parent, r.ParentCols)
  }
}
