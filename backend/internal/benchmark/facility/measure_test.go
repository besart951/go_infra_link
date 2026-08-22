package facilitybenchmark

import (
	"strings"
	"testing"
	"time"
)

func TestInspectPlanRejectsFieldDeviceSeqScanAndDiskSort(t *testing.T) {
	plan := []byte(`[{"Plan":{"Node Type":"Sort","Sort Method":"external merge","Sort Space Type":"Disk","Plans":[{"Node Type":"Seq Scan","Relation Name":"field_devices"}]}}]`)
	inspection, err := inspectPlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	if !inspection.fieldDeviceSeq || !inspection.diskSort || inspection.fieldDeviceIndex {
		t.Fatalf("unexpected inspection: %+v", inspection)
	}
}

func TestInspectPlanAcceptsFieldDeviceIndexScan(t *testing.T) {
	plan := []byte(`[{"Plan":{"Node Type":"Index Scan","Relation Name":"field_devices","Index Name":"idx_field_devices_cursor"}}]`)
	inspection, err := inspectPlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.fieldDeviceSeq || inspection.diskSort || !inspection.fieldDeviceIndex {
		t.Fatalf("unexpected inspection: %+v", inspection)
	}
}

func TestWithNextPageCreatesIndependentDirections(t *testing.T) {
	scenarios := withNextPage([]Scenario{{Name: "project_scope"}})
	if len(scenarios) != 3 {
		t.Fatalf("expected three scenarios, got %d", len(scenarios))
	}
	if scenarios[0].Name != "project_scope_first" || scenarios[0].PrimeNext || scenarios[0].PrimePrevious {
		t.Fatalf("unexpected first-page scenario: %+v", scenarios[0])
	}
	if scenarios[1].Name != "project_scope_previous" || !scenarios[1].PrimePrevious || scenarios[1].PrimeNext {
		t.Fatalf("unexpected previous-page scenario: %+v", scenarios[1])
	}
	if scenarios[2].Name != "project_scope_next" || !scenarios[2].PrimeNext || scenarios[2].PrimePrevious {
		t.Fatalf("unexpected next-page scenario: %+v", scenarios[2])
	}
}

func TestAnchorOrderSkipsNullRankForRequiredColumn(t *testing.T) {
	order := anchorOrder(anchorSorts["created_at"], "asc")
	if strings.Contains(order, "IS NULL") || order != "field_devices.created_at asc,field_devices.id asc" {
		t.Fatalf("unexpected required-column order: %s", order)
	}
}

func TestAnchorOrderIncludesEveryCompositeCursorColumn(t *testing.T) {
	order := anchorOrder(anchorSorts["sps_system_type"], "desc")
	for _, expected := range []string{"fdcv.sps_number desc", "fdcv.sps_document_name desc", "fdcv.field_device_id desc"} {
		if !strings.Contains(order, expected) {
			t.Fatalf("anchor order %q does not include %q", order, expected)
		}
	}
}

func TestSummarizeDurationsReportsErrorRate(t *testing.T) {
	result := summarizeDurations("errors", []time.Duration{time.Millisecond, 2 * time.Millisecond}, 1)
	if result.ErrorRate != 0.5 {
		t.Fatalf("error rate = %f, want 0.5", result.ErrorRate)
	}
}

func TestSelectScenariosUsesExactName(t *testing.T) {
	scenarios := []Scenario{{Name: "created_at_asc_first"}, {Name: "created_at_asc_next"}}
	selected, err := selectScenarios(scenarios, "created_at_asc_next")
	if err != nil || len(selected) != 1 || selected[0].Name != "created_at_asc_next" {
		t.Fatalf("selected=%+v err=%v", selected, err)
	}
	if _, err := selectScenarios(scenarios, "created_at"); err == nil {
		t.Fatal("partial scenario name was accepted")
	}
}

func TestBenchmarkSearchTokensAreDisjoint(t *testing.T) {
	tokens := []string{searchTokenPointOnePercent, searchTokenOnePercent, searchTokenTenPercent}
	for left, token := range tokens {
		for right, candidate := range tokens {
			if left != right && strings.Contains(token, candidate) {
				t.Fatalf("search token %q contains %q", token, candidate)
			}
		}
	}
}

func TestPlanGateAllowsSelectiveCursorPlanForCommonSearch(t *testing.T) {
	checks := PlanChecks{
		NoOffset: true, NoExactCount: true, NoFieldDeviceSeq: true, NoDiskSort: true,
		StableIDTieBreak: true, UsesIndex: true, TrigramIndexRequired: false,
	}
	if !checks.passed() {
		t.Fatal("common search with an indexed cursor plan did not pass")
	}
	checks.TrigramIndexRequired = true
	if checks.passed() {
		t.Fatal("required trigram index was not enforced")
	}
}

func TestStableIDOrderAcceptsBoundedSearchCandidate(t *testing.T) {
	statement := "SELECT * FROM candidates ORDER BY sort_value DESC, candidates.id DESC LIMIT 301"
	if !hasStableIDOrder(strings.ToLower(statement)) {
		t.Fatal("bounded search candidate ID was not accepted as stable tie-breaker")
	}
}
