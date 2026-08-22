package facilitybenchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

func (d *Database) Analyze(ctx context.Context) error {
	if _, err := d.PGX.Exec(ctx, "VACUUM (ANALYZE) field_devices, specifications, project_field_devices"); err != nil {
		return err
	}
	return nil
}

func (d *Database) reportMetadata(ctx context.Context) (Report, error) {
	report := newReportMetadata()
	if err := d.collectPostgresSettings(ctx, &report); err != nil {
		return Report{}, err
	}
	if err := d.collectDatabaseStats(ctx, &report); err != nil {
		return Report{}, err
	}
	report.Database["expected_field_devices"] = d.FieldDeviceCount
	return report, nil
}

func newReportMetadata() Report {
	return Report{
		GeneratedAt: time.Now().UTC(), GitSHA: gitSHA(), Postgres: map[string]string{}, Database: map[string]int64{},
		Hardware: map[string]string{
			"go_os": runtime.GOOS, "go_arch": runtime.GOARCH,
			"runner_logical_cpus": fmt.Sprint(runtime.NumCPU()), "database_profile_vcpus": "4",
			"database_profile_memory": "8 GiB", "database_storage": "persistent SSD volume",
		},
		ColdCache: "record separately; warm-cache p95 is the release gate",
	}
}

func (d *Database) collectPostgresSettings(ctx context.Context, report *Report) error {
	for _, setting := range benchmarkPostgresSettings() {
		var value string
		if err := d.PGX.QueryRow(ctx, "SELECT current_setting($1)", setting).Scan(&value); err != nil {
			return err
		}
		report.Postgres[setting] = value
	}
	return nil
}

func benchmarkPostgresSettings() []string {
	return []string{
		"server_version", "shared_buffers", "effective_cache_size", "work_mem",
		"random_page_cost", "track_io_timing", "jit", "fsync", "synchronous_commit",
	}
}

func (d *Database) collectDatabaseStats(ctx context.Context, report *Report) error {
	for _, table := range benchmarkRelations() {
		if err := d.collectRelationStats(ctx, report, table); err != nil {
			return err
		}
	}
	var databaseBytes int64
	if err := d.PGX.QueryRow(ctx, "SELECT pg_database_size(current_database())").Scan(&databaseBytes); err != nil {
		return err
	}
	report.Database["database_bytes"] = databaseBytes
	if err := d.collectMigrations(ctx, report); err != nil {
		return err
	}
	return d.collectDistributionStats(ctx, report)
}

func benchmarkRelations() []string {
	return []string{
		"field_devices", "field_device_cursor_values", "field_device_building_cursor_values",
		"specifications", "project_field_devices",
	}
}

func (d *Database) collectRelationStats(ctx context.Context, report *Report, table string) error {
	query := "SELECT (SELECT count(*) FROM " + table + "),pg_relation_size('" + table + "'),pg_indexes_size('" + table + "')"
	var count, tableBytes, indexBytes int64
	if err := d.PGX.QueryRow(ctx, query).Scan(&count, &tableBytes, &indexBytes); err != nil {
		return err
	}
	report.Database[table+"_rows"] = count
	report.Database[table+"_table_bytes"] = tableBytes
	report.Database[table+"_index_bytes"] = indexBytes
	return nil
}

func (d *Database) collectMigrations(ctx context.Context, report *Report) error {
	rows, err := d.PGX.Query(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return err
		}
		report.Migrations = append(report.Migrations, version)
	}
	report.Database["schema_migrations"] = int64(len(report.Migrations))
	return rows.Err()
}

func (d *Database) collectDistributionStats(ctx context.Context, report *Report) error {
	queries := map[string]string{
		"field_devices_bmk_null_rows":      "SELECT count(*) FROM field_devices WHERE bmk IS NULL",
		"search_0_1_percent_rows":          searchCountSQL(searchTokenPointOnePercent),
		"search_1_percent_rows":            searchCountSQL(searchTokenOnePercent),
		"search_10_percent_rows":           searchCountSQL(searchTokenTenPercent),
		"specification_supplier_null_rows": "SELECT count(*) FROM specifications WHERE specification_supplier IS NULL",
	}
	for key, query := range queries {
		var count int64
		if err := d.PGX.QueryRow(ctx, query).Scan(&count); err != nil {
			return err
		}
		report.Database[key] = count
	}
	return nil
}

func searchCountSQL(token string) string {
	return "SELECT count(*) FROM field_devices WHERE description ILIKE '%" + token + "%'"
}

func WriteReport(directory string, report Report) error {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "facility-benchmark.json"), encoded, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(directory, "facility-benchmark.md"), []byte(markdownReport(report)), 0o644)
}

func ReadReport(path string) (Report, error) {
	file, err := os.Open(path)
	if err != nil {
		return Report{}, err
	}
	defer file.Close()
	var report Report
	err = json.NewDecoder(file).Decode(&report)
	return report, err
}

func markdownReport(report Report) string {
	var output strings.Builder
	fmt.Fprintf(&output, "# Facility benchmark (%d FieldDevices)\n\nGit: `%s`  \nGenerated: %s\n\n", report.Database["field_devices_rows"], report.GitSHA, report.GeneratedAt.Format(time.RFC3339))
	writeStringMetadata(&output, "PostgreSQL", report.Postgres)
	writeStringMetadata(&output, "Hardware", report.Hardware)
	writeIntegerMetadata(&output, "Database", report.Database)
	fmt.Fprintf(&output, "## Cold cache\n\n%s\n\n", report.ColdCache)
	fmt.Fprintf(&output, "## Migrations\n\n`%s`\n\n", strings.Join(report.Migrations, "`, `"))
	output.WriteString("| Scenario | Samples | Failures | Error rate | p50 ms | p95 ms | p99 ms | Gate |\n|---|---:|---:|---:|---:|---:|---:|---|\n")
	for _, result := range report.Scenarios {
		fmt.Fprintf(&output, "| %s | %d | %d | %.4f%% | %.2f | %.2f | %.2f | %t |\n", result.Name, result.Samples, result.Failures, result.ErrorRate*100, result.P50MS, result.P95MS, result.P99MS, result.Passed)
	}
	return output.String()
}

func writeStringMetadata(output *strings.Builder, title string, values map[string]string) {
	fmt.Fprintf(output, "## %s\n\n", title)
	for _, key := range sortedKeys(values) {
		fmt.Fprintf(output, "- `%s`: `%s`\n", key, values[key])
	}
	output.WriteString("\n")
}

func writeIntegerMetadata(output *strings.Builder, title string, values map[string]int64) {
	fmt.Fprintf(output, "## %s\n\n", title)
	for _, key := range sortedKeys(values) {
		fmt.Fprintf(output, "- `%s`: `%d`\n", key, values[key])
	}
	output.WriteString("\n")
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func gitSHA() string {
	output, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func Passed(report Report) bool {
	if report.Database["field_devices_rows"] != report.Database["expected_field_devices"] {
		return false
	}
	if report.Database["field_device_cursor_values_rows"] != report.Database["expected_field_devices"] {
		return false
	}
	if !strings.HasPrefix(report.Postgres["server_version"], "18.3") ||
		report.Postgres["fsync"] != "on" || report.Postgres["synchronous_commit"] != "on" ||
		report.Postgres["track_io_timing"] != "on" || report.Postgres["jit"] != "off" {
		return false
	}
	for _, scenario := range report.Scenarios {
		if !scenario.Passed || scenario.Failures > 0 {
			return false
		}
	}
	return true
}
