package facilitybenchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	"github.com/besart951/go_infra_link/backend/internal/repository/facilitysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func (d *Database) Measure(ctx context.Context, options MeasureOptions) (Report, error) {
	report, err := d.reportMetadata(ctx)
	if err != nil {
		return Report{}, err
	}
	scenarios, err := d.measureScenarios(ctx, options.Scenario)
	if err != nil {
		return Report{}, err
	}
	for _, scenario := range scenarios {
		result, err := d.measureScenario(ctx, scenario)
		if err != nil {
			return Report{}, fmt.Errorf("scenario %s: %w", scenario.Name, err)
		}
		report.Scenarios = append(report.Scenarios, result)
	}
	return report, nil
}

func (d *Database) measureScenarios(ctx context.Context, name string) ([]Scenario, error) {
	canonical := canonicalScenarios()
	if name != "" {
		if selected, err := selectScenarios(canonical, name); err == nil {
			return selected, nil
		}
		if selected, ok, err := d.positionedScenarioByName(ctx, name); ok || err != nil {
			return selected, err
		}
	}
	scenarios, err := d.allScenarios(ctx)
	if err != nil {
		return nil, err
	}
	return selectScenarios(scenarios, name)
}

func selectScenarios(scenarios []Scenario, name string) ([]Scenario, error) {
	if name == "" {
		return scenarios, nil
	}
	for _, scenario := range scenarios {
		if scenario.Name == name {
			return []Scenario{scenario}, nil
		}
	}
	return nil, fmt.Errorf("unknown benchmark scenario %q", name)
}

func (d *Database) measureScenario(ctx context.Context, scenario Scenario) (ScenarioResult, error) {
	reader := facilitysql.NewFieldDeviceRepository(d.Gorm).(domainFieldDevice.CursorReader)
	query, err := scenarioWithCursor(ctx, reader, scenario)
	if err != nil {
		return ScenarioResult{}, err
	}
	for range d.Samples.Warmups {
		_, _ = reader.GetCursorPage(ctx, query)
	}
	durations, failures := sequentialSamplesN(ctx, reader, query, d.Samples.SequentialRuns)
	parallelDurations, parallelFailures := parallelSamples(ctx, reader, query, d.Samples)
	durations = append(durations, parallelDurations...)
	queries, plans, checks, err := d.capturePlans(ctx, query)
	if err != nil {
		return ScenarioResult{}, err
	}
	result := summarizeDurations(scenario.Name, durations, failures+parallelFailures)
	result.Queries, result.Plans, result.PlanChecks = queries, plans, checks
	result.Passed = result.Failures == 0 && result.P95MS <= float64(LatencyGate.Milliseconds()) && checks.passed()
	return result, nil
}

func scenarioWithCursor(ctx context.Context, reader domainFieldDevice.CursorReader, scenario Scenario) (CursorQuery, error) {
	query := scenario.Query
	if (!scenario.PrimeNext && !scenario.PrimePrevious) || query.Cursor != "" {
		return query, nil
	}
	page, err := reader.GetCursorPage(ctx, query)
	if err != nil {
		return query, err
	}
	if page.NextCursor == "" {
		return query, nil
	}
	query.Cursor = page.NextCursor
	if scenario.PrimePrevious {
		second, secondErr := reader.GetCursorPage(ctx, query)
		if secondErr != nil {
			return query, secondErr
		}
		query.Cursor = second.PreviousCursor
	}
	return query, nil
}

func parallelSamples(ctx context.Context, reader domainFieldDevice.CursorReader, query CursorQuery, profile SampleProfile) ([]time.Duration, int) {
	var mutex sync.Mutex
	durations := make([]time.Duration, 0, profile.ParallelReaders*profile.RunsPerReader)
	failures := 0
	var group sync.WaitGroup
	for range profile.ParallelReaders {
		group.Add(1)
		go func() {
			defer group.Done()
			local, failed := sequentialSamplesN(ctx, reader, query, profile.RunsPerReader)
			mutex.Lock()
			durations, failures = append(durations, local...), failures+failed
			mutex.Unlock()
		}()
	}
	group.Wait()
	return durations, failures
}

func sequentialSamplesN(ctx context.Context, reader domainFieldDevice.CursorReader, query CursorQuery, count int) ([]time.Duration, int) {
	durations := make([]time.Duration, 0, count)
	failures := 0
	for range count {
		started := time.Now()
		_, err := reader.GetCursorPage(ctx, query)
		durations = append(durations, time.Since(started))
		if err != nil {
			failures++
		}
	}
	return durations, failures
}

func summarizeDurations(name string, durations []time.Duration, failures int) ScenarioResult {
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	errorRate := float64(failures) / float64(len(durations))
	return ScenarioResult{Name: name, Samples: len(durations), Failures: failures,
		ErrorRate: errorRate,
		MinMS:     milliseconds(durations[0]), P50MS: percentile(durations, .50), P95MS: percentile(durations, .95),
		P99MS: percentile(durations, .99), MaxMS: milliseconds(durations[len(durations)-1])}
}

func percentile(values []time.Duration, quantile float64) float64 {
	index := int(float64(len(values)-1) * quantile)
	return milliseconds(values[index])
}

func milliseconds(value time.Duration) float64 { return float64(value.Microseconds()) / 1000 }

func (c PlanChecks) passed() bool {
	return c.NoOffset && c.NoExactCount && c.NoFieldDeviceSeq && c.NoDiskSort &&
		c.StableIDTieBreak && c.UsesIndex &&
		(!c.ScopeIndexRequired || c.UsesScopeIndex) &&
		(!c.TrigramIndexRequired || c.UsesTrigramIndex)
}

func (d *Database) capturePlans(ctx context.Context, query CursorQuery) ([]string, []json.RawMessage, PlanChecks, error) {
	capture := &sqlCapture{}
	database := d.Gorm.Session(&gorm.Session{Logger: capture.LogMode(logger.Info)})
	reader := facilitysql.NewFieldDeviceRepository(database).(domainFieldDevice.CursorReader)
	if _, err := reader.GetCursorPage(ctx, query); err != nil {
		return nil, nil, PlanChecks{}, err
	}
	plans, checks, err := d.explainStatements(ctx, query, capture.statements)
	return capture.statements, plans, checks, err
}

func (d *Database) explainStatements(ctx context.Context, query CursorQuery, statements []string) ([]json.RawMessage, PlanChecks, error) {
	requireScope, requireTrigram := hasBenchmarkScope(query), requiresBenchmarkTrigram(query)
	checks := PlanChecks{
		NoOffset: true, NoExactCount: true, NoFieldDeviceSeq: true, NoDiskSort: true,
		StableIDTieBreak: true, UsesIndex: true,
		ScopeIndexRequired: requireScope, TrigramIndexRequired: requireTrigram,
	}
	plans := make([]json.RawMessage, 0, len(statements))
	for _, statement := range statements {
		lower := strings.ToLower(statement)
		checks.NoOffset = checks.NoOffset && !strings.Contains(lower, " offset ")
		checks.NoExactCount = checks.NoExactCount && !strings.Contains(lower, "count(")
		if strings.Contains(lower, "cursor_value") {
			checks.StableIDTieBreak = checks.StableIDTieBreak && hasStableIDOrder(lower)
		}
		var plan []byte
		if err := d.PGX.QueryRow(ctx, "EXPLAIN (ANALYZE, BUFFERS, WAL, FORMAT JSON) "+statement).Scan(&plan); err != nil {
			return nil, checks, err
		}
		plans = append(plans, json.RawMessage(plan))
		inspection, inspectErr := inspectPlan(plan)
		if inspectErr != nil {
			return nil, checks, inspectErr
		}
		checks.NoFieldDeviceSeq = checks.NoFieldDeviceSeq && !inspection.fieldDeviceSeq
		checks.NoDiskSort = checks.NoDiskSort && !inspection.diskSort
		if inspection.sawFieldDevice {
			checks.UsesIndex = checks.UsesIndex && inspection.fieldDeviceIndex
		}
		checks.UsesScopeIndex = checks.UsesScopeIndex || inspection.scopeIndex
		checks.UsesTrigramIndex = checks.UsesTrigramIndex || inspection.trigramIndex
	}
	return plans, checks, nil
}

type planInspection struct {
	sawFieldDevice   bool
	fieldDeviceSeq   bool
	fieldDeviceIndex bool
	diskSort         bool
	scopeIndex       bool
	trigramIndex     bool
}

func inspectPlan(encoded []byte) (planInspection, error) {
	var document any
	if err := json.Unmarshal(encoded, &document); err != nil {
		return planInspection{}, fmt.Errorf("decode explain plan: %w", err)
	}
	inspection := planInspection{}
	walkPlan(document, &inspection)
	return inspection, nil
}

func walkPlan(value any, inspection *planInspection) {
	switch typed := value.(type) {
	case []any:
		for _, child := range typed {
			walkPlan(child, inspection)
		}
	case map[string]any:
		inspectPlanNode(typed, inspection)
		for _, child := range typed {
			walkPlan(child, inspection)
		}
	}
}

func inspectPlanNode(node map[string]any, inspection *planInspection) {
	nodeType := strings.ToLower(fmt.Sprint(node["Node Type"]))
	relation := strings.ToLower(fmt.Sprint(node["Relation Name"]))
	if relation == "field_devices" {
		inspection.sawFieldDevice = true
		inspection.fieldDeviceSeq = inspection.fieldDeviceSeq || nodeType == "seq scan"
		inspection.fieldDeviceIndex = inspection.fieldDeviceIndex || strings.Contains(nodeType, "index") || strings.Contains(nodeType, "bitmap")
	}
	sortMethod := strings.ToLower(fmt.Sprint(node["Sort Method"]))
	sortSpace := strings.ToLower(fmt.Sprint(node["Sort Space Type"]))
	inspection.diskSort = inspection.diskSort || strings.Contains(sortMethod, "external") || sortSpace == "disk"
	indexName := strings.ToLower(fmt.Sprint(node["Index Name"]))
	inspection.scopeIndex = inspection.scopeIndex || isScopeIndex(indexName)
	inspection.trigramIndex = inspection.trigramIndex || strings.Contains(indexName, "trgm")
}

func hasBenchmarkSearch(query CursorQuery) bool {
	return strings.TrimSpace(query.Search) != "" || strings.TrimSpace(query.Filters.Search) != ""
}

func requiresBenchmarkTrigram(query CursorQuery) bool {
	term := strings.TrimSpace(query.Search)
	if term == "" {
		term = strings.TrimSpace(query.Filters.Search)
	}
	return term == searchTokenPointOnePercent
}

func hasBenchmarkScope(query CursorQuery) bool {
	filters := query.Filters
	return filters.ProjectID != nil || len(filters.ProjectIDs) > 0 || filters.BuildingID != nil || len(filters.BuildingIDs) > 0 ||
		filters.ControlCabinetID != nil || len(filters.ControlCabinetIDs) > 0 || filters.SPSControllerID != nil || len(filters.SPSControllerIDs) > 0 ||
		filters.SPSControllerSystemTypeID != nil || len(filters.SPSControllerSystemTypeIDs) > 0
}

func isScopeIndex(name string) bool {
	for _, marker := range []string{"project_field_device", "building", "control_cabinet", "sps_controller", "system_type"} {
		if strings.Contains(name, marker) {
			return true
		}
	}
	return false
}

func hasStableIDOrder(statement string) bool {
	orderIndex := strings.LastIndex(statement, " order by ")
	if orderIndex < 0 {
		return false
	}
	order := statement[orderIndex:]
	return strings.Contains(order, "field_devices.id") || strings.Contains(order, "fd_cursor.field_device_id") ||
		strings.Contains(order, "pfd_cursor.field_device_id") || strings.Contains(order, "fd_building_scope.field_device_id") ||
		strings.Contains(order, "candidates.id")
}
