package facilitybenchmark

import (
	"strings"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type CursorQuery = domainFacility.FieldDeviceCursorQuery

var benchmarkSorts = []string{
	"created_at", "apparat_nr", "sps_system_type", "description",
	"spec_supplier", "spec_size", "spec_acdc", "spec_power",
}

func canonicalScenarios() []Scenario {
	scenarios := make([]Scenario, 0, 40)
	for _, sort := range benchmarkSorts {
		for _, order := range []string{"asc", "desc"} {
			query := CursorQuery{Limit: 300, OrderBy: sort, Order: order}
			scenarios = append(scenarios,
				Scenario{Name: sort + "_" + order + "_first", Query: query},
				Scenario{Name: sort + "_" + order + "_next", Query: query, PrimeNext: true},
			)
		}
	}
	scenarios = append(scenarios, scopeScenarios()...)
	scenarios = append(scenarios, searchScenarios()...)
	return scenarios
}

func scopeScenarios() []Scenario {
	building := deterministicID(1, 50)
	cabinet := deterministicID(5, 500)
	controller := deterministicID(6, 5_000)
	project := deterministicID(11, 50)
	return withNextPage([]Scenario{
		{Name: "building_scope", Query: CursorQuery{Limit: 300, Filters: domainFacility.FieldDeviceFilterParams{BuildingID: &building}}},
		{Name: "cabinet_scope", Query: CursorQuery{Limit: 300, Filters: domainFacility.FieldDeviceFilterParams{ControlCabinetID: &cabinet}}},
		{Name: "controller_scope", Query: CursorQuery{Limit: 300, Filters: domainFacility.FieldDeviceFilterParams{SPSControllerID: &controller}}},
		{Name: "project_scope", Query: CursorQuery{Limit: 300, Filters: domainFacility.FieldDeviceFilterParams{ProjectID: &project}}},
	})
}

func searchScenarios() []Scenario {
	return withNextPage([]Scenario{
		{Name: "search_0_1_percent", Query: CursorQuery{Limit: 300, Search: searchTokenPointOnePercent}},
		{Name: "search_1_percent", Query: CursorQuery{Limit: 300, Search: searchTokenOnePercent}},
		{Name: "search_10_percent", Query: CursorQuery{Limit: 300, Search: searchTokenTenPercent}},
		{Name: "combined_filter", Query: CursorQuery{Limit: 300, Search: searchTokenTenPercent, Filters: domainFacility.FieldDeviceFilterParams{SPSControllerIDs: []uuid.UUID{deterministicID(6, 10), deterministicID(6, 11)}}}},
	})
}

func withNextPage(first []Scenario) []Scenario {
	result := make([]Scenario, 0, len(first)*3)
	for _, base := range first {
		name := strings.TrimSuffix(base.Name, "_first")
		firstPage := base
		firstPage.Name = name + "_first"
		previousPage := base
		previousPage.Name, previousPage.PrimePrevious = name+"_previous", true
		nextPage := base
		nextPage.Name, nextPage.PrimeNext = name+"_next", true
		result = append(result, firstPage, previousPage, nextPage)
	}
	return result
}
