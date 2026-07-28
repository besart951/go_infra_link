package db

import "testing"

func TestProjectAssociationForeignKeysCoverHierarchyAndMembershipRows(t *testing.T) {
	want := map[string]string{
		"project_control_cabinets": "RESTRICT",
		"project_sps_controllers":  "RESTRICT",
		"project_field_devices":    "RESTRICT",
		"project_users":            "CASCADE",
	}

	for _, foreignKey := range projectAssociationForeignKeys {
		onDelete, exists := want[foreignKey.table]
		if !exists {
			t.Fatalf("unexpected project association table %q", foreignKey.table)
		}
		if foreignKey.name == "" {
			t.Fatalf("constraint name missing for %s", foreignKey.table)
		}
		if foreignKey.onDelete != onDelete {
			t.Fatalf(
				"%s ON DELETE: got %s, want %s",
				foreignKey.table,
				foreignKey.onDelete,
				onDelete,
			)
		}
		delete(want, foreignKey.table)
	}
	if len(want) != 0 {
		t.Fatalf("missing project association constraints: %v", want)
	}
}
