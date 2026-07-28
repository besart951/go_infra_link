package db

import (
	"fmt"

	"gorm.io/gorm"
)

const liveRevisionTriggerFunction = "bump_live_entity_revision"

// migrateLiveEntityRevisions is the expand phase of the blue-green revision
// rollout. It adds a defaulted column to every mutable Base-backed table and a
// database trigger so older application binaries also advance revisions.
func migrateLiveEntityRevisions(db *gorm.DB) error {
	if db == nil || db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}

	if err := db.Exec(fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION %s()
		RETURNS trigger
		LANGUAGE plpgsql
		AS $$
		BEGIN
			NEW.revision := OLD.revision + 1;
			RETURN NEW;
		END;
		$$
	`, liveRevisionTriggerFunction)).Error; err != nil {
		return fmt.Errorf("create live revision trigger function: %w", err)
	}

	if err := db.Exec(fmt.Sprintf(`
		DO $$
		DECLARE
			target record;
			trigger_name text;
		BEGIN
			FOR target IN
				SELECT table_schema, table_name
				FROM information_schema.columns
				WHERE table_schema = current_schema()
				GROUP BY table_schema, table_name
				HAVING BOOL_OR(column_name = 'id')
				   AND BOOL_OR(column_name = 'created_at')
				   AND BOOL_OR(column_name = 'updated_at')
			LOOP
				EXECUTE format(
					'ALTER TABLE %%I.%%I ADD COLUMN IF NOT EXISTS revision bigint NOT NULL DEFAULT 1',
					target.table_schema,
					target.table_name
				);
				trigger_name := 'trg_' || target.table_name || '_live_revision';
				EXECUTE format(
					'DROP TRIGGER IF EXISTS %%I ON %%I.%%I',
					trigger_name,
					target.table_schema,
					target.table_name
				);
				EXECUTE format(
					'CREATE TRIGGER %%I BEFORE UPDATE ON %%I.%%I FOR EACH ROW EXECUTE FUNCTION %s()',
					trigger_name,
					target.table_schema,
					target.table_name
				);
			END LOOP;
		END;
		$$
	`, liveRevisionTriggerFunction)).Error; err != nil {
		return fmt.Errorf("install live revision columns and triggers: %w", err)
	}
	return nil
}
