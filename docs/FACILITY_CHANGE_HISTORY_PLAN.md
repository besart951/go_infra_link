# Facility Change History Plan

Status: draft

Dieser Plan beschreibt die inkrementelle Umsetzung fuer:

1. Change History Module - wer hat was wann erstellt, geaendert oder geloescht.
2. Deleted Graph / Restore Module - vollstaendig geloeschte Items wiederherstellen.
3. Timeline Projection Module - schnelle Zeitachse ueber Facility-Hierarchien.

Wichtigster Performance-Fall: `field_devices` kann 5 Mio+ Zeilen enthalten. Das Feature darf Hot-Queries auf `field_devices` nicht verlangsamen und soll so isoliert sein, dass spaeter weitere Tabellen angeschlossen werden koennen.

## Zielbild

Die Facility-Mutation bleibt fachlich in den bestehenden Services. Ein neues History-Modul sitzt daneben und wird innerhalb derselben DB-Transaktion aufgerufen.

Das History-Modul speichert append-only Events. Fuer Delete-Operationen speichert es zusaetzlich einen vollstaendigen Delete-Snapshot des betroffenen Graphs. Die Timeline liest aus optimierten Scope-Spalten statt die Facility-Hierarchie zur Laufzeit ueber Millionen FieldDevices zu rekonstruieren.

## Nicht-Ziele

- Keine globale Soft-Delete-Umstellung in der ersten Version.
- Keine Aenderung der normalen Facility-Listenqueries, ausser sie muessen explizit geloeschte Items anzeigen.
- Kein Restore mit neuer ID in Version 1. Restore versucht Original-IDs wieder einzusetzen und scheitert sauber bei Konflikten.
- Keine DB-Trigger in Version 1. Actor, Diff und Facility-Scope sind im Go-Code besser kontrollierbar.

## Kernbegriffe

- Change Event: append-only Eintrag fuer create, update, delete oder restore.
- Delete Snapshot: vollstaendiger JSONB-Snapshot eines geloeschten Entity-Graphs.
- Scope: vorberechnete IDs fuer schnelle Timeline-Filter, z.B. `project_id`, `control_cabinet_id`, `field_device_id`.
- Restore Run: ein transaktionaler Wiederherstellungsversuch aus einem Delete Snapshot.

## Empfohlenes Datenmodell

### `change_events`

Pflichtfelder:

- `id uuid primary key`
- `occurred_at timestamptz not null`
- `actor_id uuid null`
- `action varchar not null` - `create`, `update`, `delete`, `restore`
- `entity_table varchar not null`
- `entity_id uuid not null`
- `batch_id uuid null`
- `summary text null`
- `before_json jsonb null`
- `after_json jsonb null`
- `diff_json jsonb null`
- `metadata_json jsonb null`

Scope-Felder:

- `project_id uuid null`
- `building_id uuid null`
- `control_cabinet_id uuid null`
- `sps_controller_id uuid null`
- `sps_controller_system_type_id uuid null`
- `field_device_id uuid null`

Indexes:

- `(entity_table, entity_id, occurred_at desc)`
- partial `(field_device_id, occurred_at desc) where field_device_id is not null`
- partial `(control_cabinet_id, occurred_at desc) where control_cabinet_id is not null`
- partial `(sps_controller_id, occurred_at desc) where sps_controller_id is not null`
- partial `(project_id, occurred_at desc) where project_id is not null`
- `occurred_at desc` or BRIN on `occurred_at` if the table grows very large.

### `delete_snapshots`

Pflichtfelder:

- `id uuid primary key`
- `change_event_id uuid not null`
- `root_table varchar not null`
- `root_id uuid not null`
- `deleted_at timestamptz not null`
- `deleted_by_id uuid null`
- `graph_json jsonb not null`
- `counts_json jsonb not null`
- `restore_status varchar not null default 'restorable'`
- `restored_at timestamptz null`
- `restored_by_id uuid null`
- `restore_error text null`

Indexes:

- `(root_table, root_id, deleted_at desc)`
- `(restore_status, deleted_at desc)`
- partial `(deleted_by_id, deleted_at desc) where deleted_by_id is not null`

## Phase 0 - Kontext festnageln

Ziel: spaeteren Kontextverlust vermeiden.

Schritte:

- Dieses Dokument als Startpunkt verwenden.
- Bei jeder umgesetzten Phase den Abschnitt "Implementation Log" unten aktualisieren.
- Neue Architekturentscheidungen entweder hier unter "Entscheidungen" eintragen oder als ADR in `docs/adr/` ablegen, wenn sie dauerhaft relevant sind.

Offene Entscheidungen vor Coding:

- Wie lange sollen History-Events aufbewahrt werden?
- Duerfen Admins alle Delete Snapshots sehen oder nur projektbezogen?
- Was soll bei Restore-Konflikten passieren: hart fehlschlagen, Konfliktbericht liefern, oder optional "restore as copy" spaeter ergaenzen?

## Phase 1 - History-Basis ohne Facility-Integration

Ziel: append-only Infrastruktur steht, aber noch ohne Verhalten in bestehenden Flows.

Schritte:

- Domain-Module anlegen, z.B. `backend/internal/domain/history`.
- Typen definieren: `ChangeEvent`, `DeleteSnapshot`, `ChangeAction`, `EntityScope`, `ChangeSet`.
- Store-Interface definieren:
  - `RecordChange(ctx, change)`
  - `RecordChanges(ctx, changes)`
  - `ListEntityHistory(ctx, entityTable, entityID, params)`
  - `ListTimeline(ctx, filter, params)`
  - `GetDeleteSnapshot(ctx, id)`
- SQL-Adapter anlegen, z.B. `backend/internal/repository/historysql`.
- Migration fuer `change_events` und `delete_snapshots` erstellen und in `backend/internal/db/migrations.go` registrieren.
- Tests:
  - Migration erzeugt Tabellen und Indexe.
  - Append-only Store kann Events schreiben und paginiert lesen.
  - `RecordChanges` schreibt batchweise.

Akzeptanz:

- Backend baut.
- Store-Tests laufen.
- Keine Facility-Route hat ihr Verhalten geaendert.

## Phase 2 - Actor Context und Change Recorder

Ziel: jeder Write kann den Actor aus `context.Context` lesen, ohne Handler-Signaturen ueberall aufzublaehen.

Schritte:

- Kleines Context-Modul anlegen, z.B. `backend/internal/service/auditctx`.
- Middleware/Handler helper nutzt vorhandenes `middleware.GetUserID` und setzt Actor in den Request Context.
- Fuer Systemprozesse einen expliziten System-Actor oder `nil` mit `metadata_json.source = "system"` erlauben.
- `ChangeRecorder` als Service-Modul bauen:
  - berechnet JSON-Snapshots
  - berechnet Diffs fuer Updates
  - akzeptiert bereits vorberechneten Scope
  - schreibt in History-Store
- Tests:
  - Actor wird aus Context gelesen.
  - Diff ignoriert unveraenderte Felder.
  - `nil` Actor ist erlaubt, aber erkennbar.

Akzeptanz:

- Kein Facility-Service muss Actor als Parameter erhalten.
- Recorder kann isoliert getestet werden.

## Phase 3 - Scope Resolver fuer Facility

Ziel: Timeline-Abfragen sind schnelle Index-Scans.

Schritte:

- Facility-spezifischen Scope Resolver bauen, z.B. `backend/internal/service/facilityhistory`.
- Resolver-Funktionen:
  - `ControlCabinetScope(control_cabinet_id)`
  - `SPSControllerScope(sps_controller_id)`
  - `SPSControllerSystemTypeScope(sps_controller_system_type_id)`
  - `FieldDeviceScope(field_device_id)`
  - `BacnetObjectScope(bacnet_object_id or snapshot)`
  - `ProjectScope` ueber Projekt-Link-Tabellen, wenn vorhanden.
- Fuer Deletes muss der Scope aus dem Before-Snapshot kommen, nicht aus Tabellen nach dem Delete.
- Tests mit kleiner Hierarchie:
  - FieldDevice Event enthaelt `control_cabinet_id`, `sps_controller_id`, `field_device_id`.
  - BacnetObject Event erbt Scope vom FieldDevice.
  - ControlCabinet Event setzt `control_cabinet_id`.

Akzeptanz:

- Scope Resolver macht keine unnoetigen Vollscans.
- FieldDevice-Scopes laufen ueber ID-basierte Queries.

## Phase 4 - Change Capture fuer sichere Einzel-Mutationen

Ziel: erste echte History fuer create/update/delete, aber noch ohne komplexe Bulk/Replace-Flows.

Start mit Tabellen:

- `control_cabinets`
- `sps_controllers`
- `field_devices`
- `bacnet_objects`

Schritte:

- History-Abhaengigkeit in Facility Services verdrahten.
- Beim Create: Event nach erfolgreichem Insert, mit `after_json`.
- Beim Update: Before laden, Update ausfuehren, Event mit `before_json`, `after_json`, `diff_json`.
- Beim Delete: Before laden, Delete ausfuehren, Event mit `before_json`.
- Alles innerhalb derselben Facility-Transaktion ausfuehren, wenn der Flow bereits transaktional ist.
- Fuer einfache Services, die `baseService` nutzen, erst spaeter einen generischen Adapter bauen. Keine grosse Umbauaktion in dieser Phase.
- Tests:
  - Create/Update/Delete erzeugt je ein Event.
  - Fehlerhafte Updates erzeugen kein Event.
  - Transaction Rollback entfernt auch History-Event.

Akzeptanz:

- Einzel-Mutationen haben History.
- Bestehende API-Responses bleiben gleich.

## Phase 5 - Bulk und Replace Flows

Ziel: die riskanten echten Facility-Flows korrekt erfassen.

Wichtige Flows:

- `FieldDeviceService.MultiCreate`
- `FieldDeviceService.BulkUpdate`
- `FieldDeviceService.BulkDelete`
- `UpdateWithBacnetObjects`
- `replaceFieldDeviceBacnetObjects`
- `replaceFieldDeviceBacnetObjectsFromObjectData`
- `objectDataTemplate.replaceBacnetObjects`

Schritte:

- `batch_id` fuer jeden Bulk-/Replace-Request erzeugen.
- Pro betroffener Entity ein Change Event schreiben, aber gesammelt mit `RecordChanges`.
- Bei Replace:
  - alte BACnetObjects als Delete Events erfassen
  - neue BACnetObjects als Create Events erfassen
  - optional ein Parent-Event mit Summary "BACnet objects replaced" schreiben.
- Bei BulkUpdate:
  - vorhandene `existingMap` und `proposedMap` wiederverwenden.
  - keine zusaetzlichen Per-Row Before-Queries einfuehren.
- Tests:
  - BulkUpdate schreibt nur fuer erfolgreiche Items Events.
  - BulkDelete schreibt pro erfolgreichem Delete Event.
  - Replace hat nachvollziehbare Delete/Create Events mit gleichem `batch_id`.

Akzeptanz:

- Keine N+1-Explosion bei BulkUpdate.
- Batch-History ist zusammenhaengend lesbar.

## Phase 6 - Delete Graph Snapshots

Ziel: vollstaendige Wiederherstellung geloeschter Facility-Graphen vorbereiten.

Schritte:

- Graph Collector bauen, z.B. `backend/internal/service/facilityrestore`.
- Snapshot-Form:
  - `tables`: Liste aus `{table, rows}`
  - `order`: Insert-Reihenfolge fuer Restore
  - `root`: `{table, id}`
  - `counts`: Anzahl pro Tabelle
  - `schema_version`
- Collector fuer:
  - ControlCabinet mit SPSController, SPSControllerSystemType, FieldDevice, Specification, BacnetObject, BacnetObjectAlarmValue, Project Links
  - SPSController mit Untergraph
  - FieldDevice mit Specification, BacnetObject, AlarmValues, Project Links
  - BacnetObject einzeln
- Vor dem Delete in derselben Transaktion Snapshot schreiben.
- Delete Event mit `delete_snapshot_id` in `metadata_json` verknuepfen.
- Tests:
  - Snapshot enthaelt alle erwarteten Kindtabellen.
  - Cascade Delete verliert keine Daten im Snapshot.
  - Snapshot wird nicht geschrieben, wenn Delete fehlschlaegt.

Akzeptanz:

- Geloeschte ControlCabinets und FieldDevices sind als restorable sichtbar.
- Noch kein Restore-Endpunkt noetig.

## Phase 7 - Restore Runner

Ziel: Delete Snapshot kann transaktional wiederhergestellt werden.

Schritte:

- Restore Runner liest Snapshot, validiert Konflikte und spielt Rows in Insert-Reihenfolge ein.
- Konfliktvalidierung:
  - Original-ID existiert bereits -> fail
  - Unique Constraints wuerden verletzt -> fail mit Konfliktbericht
  - referenzierte Projekt/Lookup-Entities fehlen -> fail oder skip nur bei explizit optionalen Links
- Nach erfolgreichem Restore:
  - `delete_snapshots.restore_status = restored`
  - `restore` Change Events fuer Root und wichtige Kind-Entities schreiben
  - optional Collaboration Refresh fuer betroffene Projekte/Scopes ausloesen
- Tests:
  - Restore FieldDevice inklusive BACnetObjects und Specification.
  - Restore ControlCabinet inklusive Untergraph.
  - Restore scheitert sauber bei Unique-Konflikt.
  - Restore laeuft komplett in einer Transaktion.

Akzeptanz:

- Vollstaendig geloeschte Items koennen wiederhergestellt werden.
- Konflikte sind erklaerbar und lassen die DB unveraendert.

## Phase 8 - Timeline Query Interface

Ziel: Backend kann Zeitachsen effizient liefern.

Endpoints, minimal:

- `GET /api/v1/history/entities/:entity_table/:entity_id`
- `GET /api/v1/facility/control-cabinets/:id/timeline`
- `GET /api/v1/facility/field-devices/:id/timeline`
- `GET /api/v1/history/deleted`
- `POST /api/v1/history/deleted/:id/restore`

Query-Regeln:

- Default Limit klein halten, z.B. 50.
- Cursor- oder Page-Pagination verwenden.
- Sortierung immer `occurred_at desc`.
- Keine Live-Joins auf `field_devices` fuer Timeline. Nur Scope-Indexe verwenden.
- User-Anzeige ueber separate User-Lookup-Batch laden oder spaeter im Frontend cachen.

Tests:

- Timeline nach Entity.
- Timeline nach ControlCabinet Scope.
- Deleted list filtert nach Root und Restore-Status.
- Berechtigungen pruefen.

Akzeptanz:

- Timeline fuer ControlCabinet liest ohne Hierarchie-Scan.
- FieldDevice-History bleibt schnell und indexfreundlich.

## Phase 9 - Frontend Integration

Ziel: sichtbare Zeitachse und Restore-Aktion.

Schritte:

- API-Adapter im Frontend anlegen.
- Timeline-Panel als wiederverwendbares Modul bauen.
- Tabellen/Detailseiten anbinden:
  - ControlCabinet
  - SPSController
  - FieldDevice
  - BACnetObject
- Deleted Items View oder Drawer fuer Restore.
- Restore-Dialog mit Konfliktanzeige.

Akzeptanz:

- User sieht wer, was, wann geaendert hat.
- Geloeschte Items sind auffindbar und wiederherstellbar.

## Performance-Regeln

- Keine `deleted_at`-Spalte auf `field_devices` in Version 1.
- Keine globalen GORM Preloads fuer History.
- Bulk-Events in Batches schreiben.
- Scope-Spalten beim Schreiben berechnen, nicht beim Lesen.
- History-Queries muessen immer durch `(scope_id, occurred_at)` oder `(entity_table, entity_id, occurred_at)` Indexe laufen.
- Fuer grosse alte History optional Partitionierung nach Monat oder BRIN auf `occurred_at` einplanen.

## Reihenfolge fuer kleine Commits

1. Add history domain types and migration.
2. Add history SQL store and store tests.
3. Add actor context and recorder tests.
4. Add facility scope resolver.
5. Capture ControlCabinet create/update/delete.
6. Capture SPSController create/update/delete.
7. Capture FieldDevice simple create/update/delete.
8. Capture BACnetObject create/update/delete.
9. Capture FieldDevice bulk/replace flows.
10. Add delete graph collector and snapshots.
11. Add restore runner.
12. Add history endpoints.
13. Add frontend timeline panel.
14. Add deleted-items restore UI.

## Implementation Log

- 2026-04-30: Initial plan created.

## Resume Notes For Future Codex Sessions

When resuming this work:

1. Read this file first.
2. Check `git status --short`.
3. Search for existing history code with `rg -n "ChangeEvent|DeleteSnapshot|ChangeRecorder|facilityhistory|facilityrestore" backend`.
4. Continue from the first unchecked or missing phase.
5. After each completed phase, update "Implementation Log" with what changed and which tests passed.
