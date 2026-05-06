package db

import "testing"

func TestProductionMigrationsAreBlueGreenCompatible(t *testing.T) {
	for _, migration := range migrations {
		if !migration.blueGreenCompatible {
			t.Fatalf(
				"migration %s (%s) is not blue-green compatible; use a maintenance migration path instead",
				migration.version,
				migration.description,
			)
		}
	}
}
