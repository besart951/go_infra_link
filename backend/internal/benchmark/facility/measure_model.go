package facilitybenchmark

import (
	"encoding/json"
	"time"
)

const (
	WarmupCount           = 10
	SequentialRuns        = 100
	ParallelReaders       = 8
	ParallelRunsPerReader = 100
	LatencyGate           = 750 * time.Millisecond
)

type SampleProfile struct {
	Warmups         int
	SequentialRuns  int
	ParallelReaders int
	RunsPerReader   int
}

type MeasureOptions struct {
	Scenario string
}

func FullSampleProfile() SampleProfile {
	return SampleProfile{Warmups: WarmupCount, SequentialRuns: SequentialRuns, ParallelReaders: ParallelReaders, RunsPerReader: ParallelRunsPerReader}
}

func SmokeSampleProfile() SampleProfile {
	return SampleProfile{Warmups: 2, SequentialRuns: 5, ParallelReaders: 2, RunsPerReader: 5}
}

type Report struct {
	GeneratedAt time.Time         `json:"generated_at"`
	GitSHA      string            `json:"git_sha"`
	Postgres    map[string]string `json:"postgres"`
	Database    map[string]int64  `json:"database"`
	Hardware    map[string]string `json:"hardware"`
	Migrations  []string          `json:"migrations"`
	ColdCache   string            `json:"cold_cache"`
	Scenarios   []ScenarioResult  `json:"scenarios"`
}

type ScenarioResult struct {
	Name       string            `json:"name"`
	Samples    int               `json:"samples"`
	Failures   int               `json:"failures"`
	ErrorRate  float64           `json:"error_rate"`
	MinMS      float64           `json:"min_ms"`
	P50MS      float64           `json:"p50_ms"`
	P95MS      float64           `json:"p95_ms"`
	P99MS      float64           `json:"p99_ms"`
	MaxMS      float64           `json:"max_ms"`
	Passed     bool              `json:"passed"`
	PlanChecks PlanChecks        `json:"plan_checks"`
	Queries    []string          `json:"queries"`
	Plans      []json.RawMessage `json:"plans"`
}

type PlanChecks struct {
	NoOffset             bool `json:"no_offset"`
	NoExactCount         bool `json:"no_exact_count"`
	NoFieldDeviceSeq     bool `json:"no_field_device_seq_scan"`
	NoDiskSort           bool `json:"no_disk_sort"`
	StableIDTieBreak     bool `json:"stable_id_tie_breaker"`
	UsesIndex            bool `json:"uses_index"`
	ScopeIndexRequired   bool `json:"scope_index_required"`
	UsesScopeIndex       bool `json:"uses_scope_index"`
	TrigramIndexRequired bool `json:"trigram_index_required"`
	UsesTrigramIndex     bool `json:"uses_trigram_index"`
}

type Scenario struct {
	Name          string
	Query         CursorQuery
	PrimeNext     bool
	PrimePrevious bool
}
