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
  - `created_at`, `updated_at`,

#### `alarm_fields`

- Zweck: Globaler Feldkatalog (ein Feldschlüssel nur einmal)
- Felder:
  - `id UUID PK`
  - `key VARCHAR(100) UNIQUE NOT NULL` (z. B. `high_limit`, `priority_for_writing`)
  - `label VARCHAR(150) NOT NULL`
  - `data_type VARCHAR(30) NOT NULL`
    - Werte: `number`, `integer`, `boolean`, `string`, `enum`, `duration`, `state_map`, `json`
  - `default_unit_code VARCHAR(30) NULL`
  - `created_at`, `updated_at`,

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
  - `created_at`, `updated_at`,
- Constraints:
  - `UNIQUE(alarm_type_id, alarm_field_id)`

---

### 2) Definitionsebene (Template / ObjectData)

#### Bestehende Tabelle `alarm_definitions` erweitern

- Aktuell: `name`, `alarm_note`
- Neu:
  - `alarm_type_id UUID FK -> alarm_types(id) NOT NULL`
  - `is_active BOOLEAN NOT NULL DEFAULT TRUE`
  - `scope VARCHAR(30) NOT NULL DEFAULT 'template'`

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
  - `created_at`, `updated_at`,
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

---

## Frontend-UI-Konzept: Alarm-Erstellung & Werteeingabe

Dieses Kapitel beschreibt, wie die drei zentralen Screens im Frontend (SvelteKit + Tailwind + bits-ui) aussehen und funktionieren sollen.  
Alle Komponenten folgen den bestehenden Mustern des Projekts (Svelte 5 Runes, `ManageEntityUseCase`, `AsyncCombobox`, `BacnetObjectRow`-Pattern).

---

### Übersicht der drei Screens

| # | Screen | Route / Kontext | Hauptziel |
|---|--------|-----------------|-----------|
| 1 | **AlarmDefinition anlegen / bearbeiten** | `/facility/alarm-definitions` | Typ wählen, Name vergeben, optionale Feld-Overrides |
| 2 | **BacnetObject – AlarmDefinition zuweisen** | `ObjectDataForm` (inline) | Pro BACnet-Zeile eine Definition auswählen |
| 3 | **FieldDevice – Alarmwerte erfassen** | FieldDevice-Tabelle (expand) | Konkrete Werte je kopiertem BACnet-Objekt eingeben |

---

### Screen 1 – AlarmDefinitionForm (erweitert)

**Datei:** `src/lib/components/facility/forms/AlarmDefinitionForm.svelte`

#### Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  Neue Alarm-Definition / Definition bearbeiten                  │
├────────────────────────┬────────────────────────────────────────┤
│  Name *                │  Alarmtyp *                            │
│  [________________]    │  [Combobox: z. B. Grenzwert (high/low)]│
├────────────────────────┴────────────────────────────────────────┤
│  ── Felder des Typs (Vorschau, nicht editierbar) ───────────────│
│  Gruppe: limits                                                  │
│   • high_limit     Zahl    °C    [Pflicht]                       │
│   • low_limit      Zahl    °C    [Pflicht]                       │
│   • prealarm_high  Zahl    °C    [Optional]                      │
│   • alarm_delay    Sekunde  s    [Optional]                      │
│                                                                  │
│  [+ Feld-Override hinzufügen]  ▸ aufklappbar pro Feld           │
├─────────────────────────────────────────────────────────────────┤
│                                    [Abbrechen]  [Erstellen]      │
└─────────────────────────────────────────────────────────────────┘
```

#### Verhalten

1. **Alarmtyp-Auswahl** (Pflichtfeld):
   - `AsyncCombobox` mit `fetcher` → `GET /alarm-types?search=...`
   - Nach Auswahl lädt die Komponente `GET /alarm-types/{id}/fields` und zeigt die Felder als Read-only-Liste (Gruppe, Label, Datentyp, Default-Einheit, Pflicht-Flag).
   - Solange kein Typ gewählt ist, bleibt der Override-Bereich ausgeblendet.

2. **Feld-Overrides** (optional, aufklappbar):
   - Jedes Feld hat einen ausklappbaren Bereich (Collapsible / Accordion).
   - Dort kann der User überschreiben: `is_required_override` (Checkbox), `default_value_override` (passendes Input-Element je Datentyp), `unit_override` (kleines Select aus `units`).
   - Nur geänderte Felder werden an den Server gesendet (sparse PATCH).

3. **Validierung:**
   - `name` erforderlich (maxlength 120).
   - `alarm_type_id` erforderlich.
   - Backend-Fehler werden wie im bestehenden `useFormState`-Hook angezeigt.

4. **State-Skizze (Svelte 5 Runes):**
```ts
let name = $state('');
let alarm_note = $state('');
let alarm_type_id = $state('');
let typeFields = $state<AlarmTypeField[]>([]);        // geladen nach Typ-Wahl
let overrides = $state<FieldOverrideDraft[]>([]);     // sparse, nur geänderte
```

---

### Screen 2 – BacnetObjectRow (Alarm-Abschnitt)

**Datei:** `src/lib/components/facility/bacnet/BacnetObjectRow.svelte`

Die bestehende BACnet-Zeile wird um einen **optionalen Alarm-Abschnitt** am unteren Rand erweitert.

#### Layout (unterhalb der vorhandenen Hardware/Software-Felder)

```
┌─────────────────────────────────────────────── BACnet-Objekt 2 ── [🗑]─┐
│  Text Fix *         │  Beschreibung                                     │
│  [_____________]    │  [_________________________]                      │
│  Software ──────────┤  Hardware ────────────────                        │
│  Typ [AI▼] Nr [001] │  Typ [AI▼] Anz [1]                               │
│  ☐ GMS-sichtbar  ☐ Optional  ☐ Individueller Text                       │
├──────────────────────────────────────────────────────────────────────────┤
│  🔔 Alarm-Definition                             [Badge: Grenzwert h/l] │
│  [AsyncCombobox: Definition suchen oder leer lassen ▼]         [✕]      │
└──────────────────────────────────────────────────────────────────────────┘
```

#### Verhalten

- Der Alarm-Abschnitt ist **immer sichtbar**, aber der Combobox-Wert ist leer (nullable).
- `AsyncCombobox` mit `fetcher` → `GET /alarm-definitions?search=...`, `fetchById` → `GET /alarm-definitions/{id}`.
- Wenn eine Definition ausgewählt ist, zeigt ein kleines **Badge** (z. B. `variant="outline"`) den Alarmtyp-Namen (z. B. `Grenzwert h/l`).
- Das `[✕]`-Icon entfernt die Auswahl (`alarm_definition_id = null`).
- Der Wert `alarm_definition_id` (string | null) wird wie andere Felder über `onUpdate('alarm_definition_id', value)` nach oben geleitet und in `ObjectDataForm` ins `BacnetObjectInput`-Array gemappt.

**Erweiterung von `BacnetObjectInput` (Domain-Typ):**
```ts
// src/lib/domain/facility/index.ts  – ergänzen:
export interface BacnetObjectInput {
  // ... bestehende Felder ...
  alarm_definition_id?: string | null;
}
```

---

### Screen 3 – FieldDevice: Alarmwerte erfassen

**Datei (neu):** `src/lib/components/facility/bacnet/BacnetAlarmValuesEditor.svelte`  
**Eingebunden in:** `FieldDeviceBacnetRow.svelte` (expandierte BACnet-Zeile unterhalb `FieldDeviceTableRow`)

#### Layout (erscheint nach Aufklappen einer BACnet-Zeile via `ChevronRight` / `ChevronDown`)

```
┌── BACnet-Objekt: RaumTemp_AI_001  [Typ: AI] ───────────────────────────────┐
│  Alarm-Definition: Grenzwert hoch/niedrig        ● 2 / 4 Pflichtfelder ✓  │
│                                                                             │
│  ┌─ limits ──────────────────────────────────────────────────────────────┐  │
│  │  Oberer Grenzwert *    [  22.0  ] [°C ▼]                             │  │
│  │  Unterer Grenzwert *   [  18.0  ] [°C ▼]                             │  │
│  │  Voralarm oben         [  21.5  ] [°C ▼]                             │  │
│  │  Voralarm unten        [  18.5  ] [°C ▼]                             │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│  ┌─ monitoring ──────────────────────────────────────────────────────────┐  │
│  │  Verzögerung           [    5   ] [s  ▼]                             │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                           [Werte speichern]  │
└─────────────────────────────────────────────────────────────────────────────┘
```

Falls kein Alarm zugewiesen: Placeholder „Kein Alarm konfiguriert".

#### Feldtypen → Input-Elemente

| `data_type`   | Input-Element                                         |
|---------------|-------------------------------------------------------|
| `number`      | `<Input type="number" step="any" />`                  |
| `integer`     | `<Input type="number" step="1" />`                    |
| `boolean`     | `<Checkbox />`                                        |
| `string`      | `<Input type="text" />`                               |
| `enum`        | `<select>` mit `validation_json.options`              |
| `duration`    | `<Input type="number" />` + Einheiten-`<select>`      |
| `state_map`   | Dynamische Liste: bis zu 16 Zeilen `Nr │ Text-Input`  |
| `json`        | `<Textarea />` (raw JSON, mit Format-Hinweis)         |

#### Verhalten

1. **Laden:** Beim Aufklappen der Zeile → `GET /bacnet-objects/{id}/alarm-schema` liefert das Feldschema (Typ + Validierungsregeln + Defaults). Parallel: `GET /bacnet-objects/{id}/alarm-values` liefert vorhandene Werte.
2. **Anzeige:** Felder werden nach `ui_group` in Accordion-Sektionen gegliedert. Pflichtfelder bekommen `*` im Label. Felder aus dem Schema ohne gespeicherten Wert zeigen den `default_value` als Platzhalter.
3. **Einheiten-Select:** Nur sichtbar, wenn das Feld `default_unit_code` hat oder `unit_override_id` gesetzt ist. Gefüllt aus einer kleinen statischen Liste (aus `units`-Tabelle gecacht).
4. **Vollständigkeits-Badge:**
   - Grün (✓): alle Pflichtfelder ausgefüllt.
   - Orange (!): Pflichtfelder fehlen noch.
   - Grau (–): kein Alarm zugewiesen.
5. **Speichern:** `PUT /bacnet-objects/{id}/alarm-values` (Batch-Update aller Felder auf einmal). Fehler werden inline neben dem jeweiligen Feld angezeigt.
6. **State-Skizze:**
```ts
let schema = $state<AlarmTypeField[]>([]);    // geladen on expand
let values = $state<AlarmValueDraft[]>([]);   // lokale Edits
let saving = $state(false);
let saveError = $state('');
```

---

### Komponenten-Baum & Dateiempfehlungen

```
src/lib/
├── components/facility/
│   ├── forms/
│   │   └── AlarmDefinitionForm.svelte          ← ERWEITERN (Typ-Combobox + Felder-Vorschau)
│   ├── bacnet/
│   │   ├── BacnetObjectRow.svelte              ← ERWEITERN (Alarm-Definition-Select unten)
│   │   ├── BacnetObjectsEditor.svelte          (unverändert)
│   │   └── BacnetAlarmValuesEditor.svelte      ← NEU (Screen 3)
│   └── selects/
│       └── AlarmDefinitionSelect.svelte        ← NEU (AsyncCombobox-Wrapper)
├── domain/facility/
│   └── index.ts                                ← ERWEITERN (AlarmTypeField, AlarmValueDraft, etc.)
├── infrastructure/api/
│   ├── alarmDefinitionRepository.ts            ← ERWEITERN (create/update mit alarm_type_id)
│   └── bacnetAlarmRepository.ts               ← NEU (getSchema, getValues, putValues)
└── i18n/
    └── de.ts (o. ä.)                           ← ERWEITERN (Keys unten)
```

---

### API-Aufrufe je Screen

#### Screen 1 – AlarmDefinition-Formular

| Aktion | Methode | Endpoint |
|--------|---------|----------|
| Typen suchen | `GET` | `/alarm-types?search=&page=1&pageSize=20` |
| Felder eines Typs laden | `GET` | `/alarm-types/{id}/fields` |
| Definition anlegen | `POST` | `/alarm-definitions` |
| Definition bearbeiten | `PATCH` | `/alarm-definitions/{id}` |

#### Screen 2 – BacnetObjectRow

| Aktion | Methode | Endpoint |
|--------|---------|----------|
| Definitionen suchen | `GET` | `/alarm-definitions?search=&page=1&pageSize=20` |
| Definition per ID laden | `GET` | `/alarm-definitions/{id}` |
| ObjectData speichern | `POST/PATCH` | `/object-data` (bestehend, `alarm_definition_id` ergänzen) |

#### Screen 3 – Alarmwerte-Editor

| Aktion | Methode | Endpoint |
|--------|---------|----------|
| Feldschema laden | `GET` | `/bacnet-objects/{id}/alarm-schema` |
| Gespeicherte Werte laden | `GET` | `/bacnet-objects/{id}/alarm-values` |
| Werte speichern | `PUT` | `/bacnet-objects/{id}/alarm-values` |

---

### i18n-Keys (Ergänzungen)

```ts
// Vorschlag – zu ergänzen in der Übersetzungsdatei
'facility.forms.alarm_definition.alarm_type_label': 'Alarmtyp',
'facility.forms.alarm_definition.alarm_type_placeholder': 'Typ suchen...',
'facility.forms.alarm_definition.fields_preview_title': 'Felder des Typs (Vorschau)',
'facility.forms.alarm_definition.override_section': 'Feld-Override',
'facility.forms.alarm_definition.override_required': 'Pflichtfeld erzwingen',
'facility.forms.alarm_definition.override_default': 'Standard-Wert überschreiben',
'facility.forms.alarm_definition.override_unit': 'Einheit überschreiben',

'field_device.bacnet.row.alarm_definition': 'Alarm-Definition',
'field_device.bacnet.row.alarm_definition_placeholder': 'Definition suchen oder leer lassen',
'field_device.bacnet.row.alarm_definition_remove': 'Alarm-Definition entfernen',

'field_device.bacnet.alarm_editor.title': 'Alarm-Werte',
'field_device.bacnet.alarm_editor.no_alarm': 'Kein Alarm konfiguriert',
'field_device.bacnet.alarm_editor.save': 'Werte speichern',
'field_device.bacnet.alarm_editor.completeness': '{filled} / {total} Pflichtfelder ausgefüllt',
'field_device.bacnet.alarm_editor.complete': 'Vollständig',
'field_device.bacnet.alarm_editor.incomplete': 'Unvollständig',
```

---

### Zusammenfassung des Auftrags (Vorlage)

Wenn du mir die Implementierung beauftragen möchtest, kannst du folgende Vorlage verwenden:

> **Auftrag: Alarm-Frontend implementieren (gemäß ALARM_MODEL_CONCEPT.md)**
>
> Bitte implementiere die drei Screens laut Konzept-Abschnitt „Frontend-UI-Konzept":
>
> 1. **AlarmDefinitionForm erweitern** (`AlarmDefinitionForm.svelte`):
>    - `alarm_type_id`-Feld mit `AsyncCombobox` (Endpoint: `GET /alarm-types`).
>    - Felder-Vorschau-Sektion nach Typ-Auswahl (Endpoint: `GET /alarm-types/{id}/fields`).
>    - Optionale Feld-Override-Accordion-Sektion.
>
> 2. **BacnetObjectRow erweitern** (`BacnetObjectRow.svelte`):
>    - Neuen Alarm-Abschnitt am unteren Rand mit `AlarmDefinitionSelect.svelte`.
>    - Badge für Alarmtyp-Namen, X-Button zum Entfernen.
>    - `alarm_definition_id` in `BacnetObjectInput`-Typ und `ObjectDataForm` ergänzen.
>
> 3. **BacnetAlarmValuesEditor neu erstellen** (`BacnetAlarmValuesEditor.svelte`):
>    - Feldschema und gespeicherte Werte laden.
>    - Dynamische Felder je `data_type` rendern (Accordion nach `ui_group`).
>    - Einheiten-Select, Vollständigkeits-Badge, Speichern-Button.
>    - In `FieldDeviceBacnetRow` (oder bestehende BACnet-Expand-Zeile) einbinden.
>
> 4. Neue Dateien: `AlarmDefinitionSelect.svelte`, `bacnetAlarmRepository.ts`, Domain-Typen ergänzen.
>
> 5. i18n-Keys wie im Konzept ergänzen.
