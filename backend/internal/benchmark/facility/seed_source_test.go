package facilitybenchmark

import (
	"strings"
	"testing"
)

func TestSearchRowsKeepAQueryableDescription(t *testing.T) {
	for _, index := range []int64{1_000, 100, 10} {
		row, err := fieldDeviceRow(index)
		if err != nil {
			t.Fatal(err)
		}
		description, ok := row[5].(string)
		if !ok || !strings.Contains(description, "selectivity_") {
			t.Fatalf("index %d description = %#v", index, row[5])
		}
	}
}

func TestSearchTokenSelectivityIsDeterministic(t *testing.T) {
	counts := map[string]int{}
	for index := int64(0); index < 100_000; index++ {
		value := textValue(index)
		for _, token := range []string{searchTokenPointOnePercent, searchTokenOnePercent, searchTokenTenPercent} {
			if strings.Contains(value, token) {
				counts[token]++
			}
		}
	}
	if counts[searchTokenPointOnePercent] != 100 || counts[searchTokenOnePercent] != 1_000 || counts[searchTokenTenPercent] != 10_000 {
		t.Fatalf("unexpected exclusive token counts: %v", counts)
	}
}
