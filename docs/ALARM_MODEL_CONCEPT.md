# Alarm-Konzept für BacnetObject + AlarmDefinition

## Zielbild

`AlarmDefinition` soll kein Dummy mehr sein, sondern:

1. beim Erstellen von `ObjectData`/`BacnetObject` einen **Alarmtyp** festlegen,
2. beim Kopieren auf `FieldDevice` die **passenden Felder dynamisch anzeigen**,
3. pro kopiertem `BacnetObject` die **Pflichtwerte stabil und validierbar speichern**.

Das Modell unten ist normalisiert, versionierbar und deckt die genannten Alarmfälle ab.

---

## Datenbankmodell (normalisiert)

### 1) Stammdaten

#### `alarm_types`

- Zweck: Technischer Typ (z. B. `limit_high_low`, `io_monitoring`, `priority_write`)
- Felder:
  - `id UUID PK`
  - `code VARCHAR(80) UNIQUE NOT NULL`
  - `name VARCHAR(120) NOT NULL`
  - `description TEXT NULL`
  - `version INT NOT NULL DEFAULT 1`
  - `is_active BOOLEAN NOT NULL DEFAULT TRUE`
  - `created_at`, `updated_at`, `deleted_at`

#### `alarm_fields`

- Zweck: Globaler Feldkatalog (ein Feldschlüssel nur einmal)
- Felder:
  - `id UUID PK`
  - `key VARCHAR(100) UNIQUE NOT NULL` (z. B. `high_limit`, `priority_for_writing`)
  - `label VARCHAR(150) NOT NULL`
  - `data_type VARCHAR(30) NOT NULL`
    - Werte: `number`, `integer`, `boolean`, `string`, `enum`, `duration`, `state_map`, `json`
  - `default_unit_code VARCHAR(30) NULL`
  - `description TEXT NULL`
  - `created_at`, `updated_at`, `deleted_at`

#### `units`

- Zweck: Einheitentabelle (statt Freitext)
- Felder:
  - `id UUID PK`
  - `code VARCHAR(30) UNIQUE NOT NULL` (z. B. `C`, `PERCENT`, `PA`, `K`, `LUX`, `RPM`, `M3_H`, `KWH`, `H`, `MIN`, `M_S`, `V`, `REF_PERCENT`)
  - `symbol VARCHAR(20) NOT NULL` (z. B. `°C`, `%`, `Pa`)
  - `name VARCHAR(100) NOT NULL`

#### `alarm_type_fields`

- Zweck: Welche Felder gehören zu welchem Alarmtyp (inkl. Pflicht/Default/Regeln)
- Felder:
  - `id UUID PK`
  - `alarm_type_id UUID FK -> alarm_types(id) NOT NULL`
  - `alarm_field_id UUID FK -> alarm_fields(id) NOT NULL`
  - `display_order INT NOT NULL DEFAULT 0`
  - `is_required BOOLEAN NOT NULL DEFAULT FALSE`
  - `is_user_editable BOOLEAN NOT NULL DEFAULT TRUE`
  - `default_value_json JSONB NULL`
  - `validation_json JSONB NULL` (min/max, regex, enum options, step, precision)
  - `default_unit_id UUID FK -> units(id) NULL`
  - `ui_group VARCHAR(80) NULL` (z. B. `limits`, `monitoring`, `pid`, `runtime`)
  - `created_at`, `updated_at`, `deleted_at`
- Constraints:
  - `UNIQUE(alarm_type_id, alarm_field_id)`

---

### 2) Definitionsebene (Template / ObjectData)

#### Bestehende Tabelle `alarm_definitions` erweitern

- Aktuell: `name`, `alarm_note`
- Neu:
  - `alarm_type_id UUID FK -> alarm_types(id) NOT NULL`
  - `is_active BOOLEAN NOT NULL DEFAULT TRUE`
  - `version INT NOT NULL DEFAULT 1`
  - `scope VARCHAR(30) NOT NULL DEFAULT 'template'`
  - `UNIQUE(name, version, deleted_at)`

`bacnet_objects.alarm_definition_id` bleibt bestehen, aber **nullable** (ein `BacnetObject` kann ohne Alarm arbeiten).

Empfohlene FK-Regel:

- `bacnet_objects.alarm_definition_id -> alarm_definitions.id ON DELETE SET NULL`

So bleibt das `BacnetObject` gültig, auch wenn eine Definition entfernt wurde.

#### `alarm_definition_field_overrides` (optional, aber empfohlen)

- Zweck: Typ-Defaults pro Definition feinjustieren
- Felder:
  - `id UUID PK`
  - `alarm_definition_id UUID FK -> alarm_definitions(id) NOT NULL`
  - `alarm_type_field_id UUID FK -> alarm_type_fields(id) NOT NULL`
  - `is_required_override BOOLEAN NULL`
  - `default_value_override_json JSONB NULL`
  - `validation_override_json JSONB NULL`
  - `unit_override_id UUID FK -> units(id) NULL`
- Constraints:
  - `UNIQUE(alarm_definition_id, alarm_type_field_id)`

---

### 3) Werteebene (pro kopiertem FieldDevice-BacnetObject)

#### `bacnet_object_alarm_values`

- Zweck: Konkrete, vom Benutzer ausgefüllte Werte je geklontem `BacnetObject`
- Felder:
  - `id UUID PK`
  - `bacnet_object_id UUID FK -> bacnet_objects(id) NOT NULL`
  - `alarm_type_field_id UUID FK -> alarm_type_fields(id) NOT NULL`
  - `value_number NUMERIC(18,6) NULL`
  - `value_integer BIGINT NULL`
  - `value_boolean BOOLEAN NULL`
  - `value_string TEXT NULL`
  - `value_json JSONB NULL` (für State-Mapping/komplexe Strukturen)
  - `unit_id UUID FK -> units(id) NULL`
  - `source VARCHAR(20) NOT NULL DEFAULT 'user'` (`default`, `user`, `import`)
  - `created_at`, `updated_at`, `deleted_at`
- Constraints:
  - `UNIQUE(bacnet_object_id, alarm_type_field_id)`
  - Check: genau ein Value-Feld belegt (oder explizit `NULL` erlaubt für „noch offen“)
  - FK-Regel: `bacnet_object_id -> bacnet_objects.id ON DELETE CASCADE`

#### (Optional) `bacnet_object_alarm_status`

- Zweck: Schnell filtern „vollständig“ vs. „offen“
- Felder:
  - `bacnet_object_id UUID PK`
  - `required_total INT NOT NULL`
  - `required_filled INT NOT NULL`
  - `is_complete BOOLEAN NOT NULL`
  - `last_evaluated_at TIMESTAMP NOT NULL`

---

## Ablauf im bestehenden Prozess

1. User erstellt `ObjectData` + `BacnetObject` und setzt `alarm_definition_id`.
2. Beim Kopieren auf `FieldDevice` wird `BacnetObject` wie heute geklont.
3. UI lädt Feldschema über:
   - `bacnet_object -> alarm_definition -> alarm_type -> alarm_type_fields (+ overrides)`
4. UI zeigt Pflichtfelder; User füllt Werte.
5. Backend speichert Werte in `bacnet_object_alarm_values` und validiert gegen `validation_json`.
6. Vollständigkeit ergibt sich aus Required-Feldern.

Damit bleibt das Copy-Verhalten kompatibel, aber die Alarm-Eingabe wird typgesteuert.

---

## Abbildung deiner genannten Fälle auf Alarmtypen

Vorschlag Startkatalog:

1. `priority_write`
   - Felder: `priority_for_writing` (10/11/12/13/14), optional `alarm_state`
2. `active_inactive`
   - Felder: `alarm_state` (`active`/`inactive`), optional `alarm_delay`
3. `limit_high_low`
   - Felder: `high_limit`, `low_limit`, optional `prealarm_high`, `prealarm_low`, `alarm_delay`
4. `io_monitoring`
   - Felder: `nc_reference`, `feedback_value`, optional `monitor_mode`
5. `cov_logging`
   - Felder: `logging_type` (z. B. `COV`), `buffer_size` (z. B. 576)
6. `elapsed_active_time`
   - Felder: `elapsed_active_time`, optional Einheit `h`/`min`/`s`
7. `pid_control`
   - Felder: `controlled_variable_value`, `present_value`, `setpoint`, `proportional_constant`, `integral_constant`, `maximum_output`, `minimum_output`, `error_limit`, `time_delay`
8. `state_mapping`
   - Felder: `state_map_json` (State [1..16] Texte/Default)
9. `position_control`
   - Felder: `position_mode`, `state_map_json`
10. `custom_value`

- Felder: `custom_label`, `custom_value`, `unit`

Einheiten aus Liste (`°C`, `%`, `Pa`, `K`, `Lux`, `U/min`, `m`, `kW`, `m3/h`, `kWh`, `g/kg`, `hPa`, `J`, `m/s`, `V`, `h`, `min`, `% ref`) kommen in `units` und werden pro Feld zugewiesen.

---

## Stabilitätsregeln (wichtig)

1. **Keine Freitext-Schlüssel** im Value-Store, nur FK auf `alarm_type_fields`.
2. **Versionierung** auf `alarm_types` und `alarm_definitions` (breaking changes ohne Datenverlust).
3. **Hard Constraints**:
   - Unique auf `(bacnet_object_id, alarm_type_field_id)`
   - Unique auf `(alarm_type_id, alarm_field_id)`
4. **Soft Delete** wie bestehendes Domain-Pattern (`deleted_at`).
5. **Validation zentral im Backend** (nicht nur Frontend).
6. **Template/Instance-Trennung**: Definitionen enthalten Struktur, Values enthalten Instanzdaten.
7. **Delete-Semantik klar trennen**:
   - Löschen eines `BacnetObject` löscht nur dessen Instanzwerte (`bacnet_object_alarm_values`) per `CASCADE`.
   - `AlarmDefinition` wird dabei **nicht automatisch gelöscht**, weil sie meist von mehreren `BacnetObjects` verwendet wird.
   - Nur verwaiste Definitionen (0 Referenzen) dürfen optional per Cleanup-Job entfernt werden.

### Kaskaden-Regeln (Kurzfassung)

- `bacnet_objects` gelöscht -> `bacnet_object_alarm_values` gelöscht (`CASCADE`)
- `alarm_definitions` gelöscht -> `bacnet_objects.alarm_definition_id = NULL` (`SET NULL`)
- `alarm_types` gelöscht -> in der Praxis blockieren (`RESTRICT`), solange `alarm_definitions` darauf zeigen

---

## Migrationsplan (ohne Big Bang)

1. Neue Tabellen anlegen: `alarm_types`, `alarm_fields`, `alarm_type_fields`, `units`, `alarm_definition_field_overrides`, `bacnet_object_alarm_values`.
2. `alarm_definitions` um `alarm_type_id`, `version`, `scope`, `is_active` erweitern.
3. Legacy-Daten backfill:
   - Bestehende Dummy-`alarm_definitions` auf `custom_value` oder `active_inactive` mappen.
4. Backend lesen/schreiben erweitern:
   - Schema-Endpunkt für UI (pro `bacnet_object_id`)
   - Value-CRUD mit Validierung.
5. Nach Stabilisierung: alte freie `alarm_note` nur noch optional als Kommentar nutzen.

---

## Minimaler MVP-Schnitt

Wenn du klein starten willst:

1. Nur 5 Typen aktivieren: `priority_write`, `active_inactive`, `limit_high_low`, `io_monitoring`, `custom_value`.
2. Nur 1 Value-Tabelle (`bacnet_object_alarm_values`) mit `value_string`, `value_number`, `value_json`.
3. State-Listen zunächst nur als `value_json` speichern.
4. `PID`/`Position` in Phase 2 ergänzen.

So bekommst du schnell ein stabiles Fundament ohne Refactoring der ganzen Facility-Domain.
