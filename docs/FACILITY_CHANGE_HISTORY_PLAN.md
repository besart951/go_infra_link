# Facility Change History Plan

Status: implemented; retained as the historical rollout plan

> Architekturhinweis (2026-07-20): Der aktuelle Umsetzungsstand und die
> verbindlichen Abgrenzungen stehen in
> `docs/FACILITY_MUTATION_ARCHITECTURE.md`. Der unten beschriebene
> `ChangeRecorder` entspricht heute dem transaktionalen
> `historycapture.ChangeStore`/`historysql.Store`-Pfad. Das entfernte
> `service/changecapture`-Sparse-Modell darf nicht als zweites Audit-System
> wieder eingefuehrt werden.

> Abschlussnotiz (2026-07-23): Die Phasen 1 bis 9 sind ueber
> `domain/history`, `historysql`, `historycapture`, die History-Application-
> Policy, HTTP-Routen und die Svelte-Timeline umgesetzt. Zwei geplante
> Persistenzformen wurden bewusst ersetzt: normalisierte
> `change_event_scopes` statt nullable Scope-Spalten auf jedem Event und
> append-only `entity_versions` statt eines monolithischen
> `delete_snapshots.graph_json`. Dadurch koennen auch sehr grosse Hierarchien
> gebatcht erfasst und zu einem Zeitpunkt rekonstruiert werden. Die
> verbindliche Mutation-/Delete-/Provenance-Architektur steht in
> `FACILITY_MUTATION_ARCHITECTURE.md`.

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

## Urspruenglich empfohlenes Datenmodell und umgesetzte Abweichung

Die folgenden Felder dokumentieren den urspruenglichen Entwurf. Umgesetzt sind:

- `change_events` fuer Actor, Aktion, Before/After/Diff und Batch-Korrelation;
- `change_event_scopes` als normalisierte, indexierte Zuordnung statt der unten
  skizzierten nullable Scope-Spalten;
- `entity_versions` als eine Snapshot-Version pro Entity und Event statt eines
  unbeschraenkt grossen Delete-Graph-JSON.

Restore sammelt Ziel-IDs aus aktuellen Beziehungen und historischen Scopes,
liest die jeweils letzte Version set-weise in 500er Chunks und spielt Tabellen
in expliziter Abhaengigkeitsreihenfolge zurueck.

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

## Phase 0 - Kontext festnageln - umgesetzt

Ziel: spaeteren Kontextverlust vermeiden.

Schritte:

- Dieses Dokument als Startpunkt verwenden.
- Bei jeder umgesetzten Phase den Abschnitt "Implementation Log" unten aktualisieren.
- Neue Architekturentscheidungen entweder hier unter "Entscheidungen" eintragen oder als ADR in `docs/adr/` ablegen, wenn sie dauerhaft relevant sind.

Entscheidungsstand:

- Aufbewahrungsdauer bleibt eine Produktions-/Compliance-Konfiguration; bis zu
  dieser Entscheidung werden Events nicht automatisch geloescht.
- Globale History und Restore sind zentral auf `SUPERADMIN`, `ADMIN_FZAG` und
  `FZAG` begrenzt. Projekt-History erfordert Mitgliedschaft/AccessPolicy;
  projektbezogener Cabinet-Restore zusaetzlich einen aktuellen oder
  historischen Root-Link.
- Restore mit Original-ID schlaegt bei ID-, FK- oder Unique-Konflikten
  transaktional fehl und laesst Daten sowie History unveraendert.
  "Restore as copy" bleibt ein separates, nicht implizit aktiviertes Feature.

## Phase 1 - History-Basis ohne Facility-Integration - umgesetzt

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

## Phase 2 - Actor Context und Change Recorder - umgesetzt

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

## Phase 3 - Scope Resolver fuer Facility - umgesetzt

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

## Phase 4 - Change Capture fuer sichere Einzel-Mutationen - umgesetzt

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

## Phase 5 - Bulk und Replace Flows - umgesetzt

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

## Phase 6 - Delete Graph Snapshots - umgesetzt als gebatchte Entity-Versionen

Ziel: vollstaendige Wiederherstellung geloeschter Facility-Graphen vorbereiten.

Umsetzungshinweis: Der bounded Hierarchy-Delete-Cleaner liest Nachfahren in
begrenzten Seiten, schreibt Events/Entity-Versionen set-weise, bereinigt alle
Project-Links und behandelt geteilte BACnet-Ownership explizit. Ein einzelnes
unbeschraenktes `graph_json` wird nicht erzeugt.

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

## Phase 7 - Restore Runner - umgesetzt

Ziel: Delete Snapshot kann transaktional wiederhergestellt werden.

Umsetzungshinweis: Entity-Restore und ControlCabinet-Zeitpunkt-Restore verwenden
`entity_versions`. PostgreSQL-Tests decken `to_jsonb`, erfolgreichen Restore,
Unique-Konflikte und vollstaendigen Rollback inklusive Restore-History ab.

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

## Phase 8 - Timeline Query Interface - umgesetzt

Ziel: Backend kann Zeitachsen effizient liefern.

Die produktiven Routen bieten globale sowie projektisolierte Timeline,
Event-Restore und ControlCabinet-Restore. Die UI findet geloeschte Items als
Delete-Events in der Timeline; ein zusaetzlicher `delete_snapshots`-Endpunkt ist
wegen des Entity-Version-Modells nicht erforderlich.

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

## Phase 9 - Frontend Integration - umgesetzt

Ziel: sichtbare Zeitachse und Restore-Aktion.

Die globale Timeline-Seite, wiederverwendbare Hierarchie-Dialoge, Filter,
Infinite Loading, Actor-/Entity-Darstellung, Berechtigungspruefung,
Konfliktfehler und globale/projektbezogene Restore-Aktionen sind implementiert.

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
- 2026-07-20: Das konkurrierende `service/changecapture`-Modell entfernt und
  `historycapture.ChangeStore` als Consumer-Interface eingefuehrt. Plurale
  Create/Delete-Aufzeichnung, Snapshot-Reads, Scope-Aufloesung und
  Event/Scope/Version-Inserts sind in 500er Batches umgesetzt. Query-Budget-
  Tests decken FieldDevice, Specification, BACnet/Alarm, SystemType sowie die
  501-Item-Grenze ab. PostgreSQL-Integrationstests fuer `to_jsonb`, Restore und
  echtes Transaktions-Rollback bleiben offen.
- 2026-07-20: Bei einem FieldDevice-Parent-Move ermittelt die gebatchte
  Scope-Aufloesung die Vereinigungsmenge der alten und neuen
  SPSControllerSystemType-Hierarchie aus Before/After-Snapshots. Damit bleiben
  Building-, Cabinet-, SPS-, SystemType- und Project-Scopes beider Pfade ohne
  N+1 Queries auffindbar. Direkte `ProjectFieldDevice` Links werden weiterhin
  unveraendert beibehalten; ihre Herkunft ist im aktuellen Schema nicht
  gespeichert.
- 2026-07-20: `PUT SPSController` laeuft jetzt ueber eine application-eigene
  Transaktion. Controller-Update, optionaler Cabinet-Move,
  SystemType-Replacement und `historycapture` teilen eine Operation-/Batch-ID;
  erst nach erfolgreichem Commit werden direkte `ProjectSPSController` Links
  gebatcht aufgeloest und ein typisiertes Update-/Move-Kommando dispatcht. Die
  Scope-Aufloesung nimmt bei Moves beide Cabinets, Buildings und direkten
  Cabinet-Projekte auf, ohne Projekte fremder SPS-Nachfahren im alten Cabinet
  zu leaken. Query-Budget-Tests belegen konstante Statement-Zahlen fuer einen
  und zwanzig SPS-Moves; Write- und Commit-Fehler erzeugen keinen Dispatch.
- 2026-07-20: `PUT ControlCabinet` laeuft ebenfalls ueber eine
  application-eigene Transaktion. Cabinet-Update bzw. Building-Move und die
  paginierte Regenerierung aller SPS-DeviceNames werden gemeinsam mit History
  committed oder zurueckgerollt. Die Root-History enthaelt alte und neue
  Building-Scopes; die SPS-Kindevents teilen dieselbe Batch-ID. Erst nach Commit
  werden direkte `ProjectControlCabinet` Empfaenger aufgeloest und das bestehende
  v1-`control_cabinet`-Delta ueber ein typisiertes Update-/Move-Kommando erzeugt.
  Ein-versus-zwanzig Move-Batches bleiben im Scope-Resolver query-konstant.
- 2026-07-20: `PUT BacnetObject` laeuft jetzt ueber eine application-eigene
  Transaktion mit autoritativem Load, typisiertem Patch und optionaler
  FieldDevice-Reassignment. Row-Write und `historycapture` teilen die
  Operation-/Batch-ID; erst nach Commit wird die Vereinigungsmenge alter und
  neuer direkter `ProjectFieldDevice` Links aufgeloest und pro Projekt als
  bestehender v1-FieldDevice-Refresh dispatcht. Before/After-Snapshots liefern
  beide FieldDevice-Hierarchie-/Project-Scopes ohne N+1 Queries. Eine neue
  ObjectData-Verknuepfung wird innerhalb der Transaktion vor dem dekorierten
  Row-Update angelegt, damit deren ObjectData-/Project-Scope im Event enthalten
  ist. Join-Restore und ein v1-Template-Realtime-Scope bleiben offen.
- 2026-07-23: Die Grundlage fuer den transaktionalen Collaboration-Outbox-Pfad
  ist additiv vorhanden: persistierte Events, Zustellversuche und
  Consumer-Idempotenz, projektweise Sequenzen, Delivery-Leases, begrenztes
  Claiming mit Retry sowie ein typisierter Command-Codec. Der Runtime-Worker
  verarbeitet maximal 100 Events je Lauf, fordert abgelaufene Claims zurueck
  und erzeugt versionierte `committed_event`-Nachrichten.
- 2026-07-23: Transaktionale Outbox-Producer sind fuer projektbezogene
  FieldDevice-, ControlCabinet- und SPSController-Zuweisung/Reassignment/Copy,
  die einzelnen Hierarchie-Updates und -Deletes, FieldDevice-Bulk-Delete,
  projektbezogenen ControlCabinet-Restore, ObjectData-Aktivierung sowie
  BACnet-Create/Update/Alarmwert-Replacement verdrahtet. Projektlinks werden
  innerhalb der Mutationstransaktion gelesen; Scope- oder Outbox-Fehler rollen
  Facility-Write und History gemeinsam zurueck. ObjectData hat einen typisierten
  v2-Scope. Der bisherige serverseitige v1-Dispatch bleibt nur als
  Kompatibilitaetsprojektion nach Commit aktiv.
- 2026-07-23: Der Hub akzeptiert keine browserseitig erzeugten committed
  `entity_delta`-Nachrichten mehr. Der Svelte-Client dedupliziert v2-Events per
  EventID, verfolgt die Projektsequenz und optionale Entity-Revisionswerte und
  faellt bei Luecken auf einen autoritativen HTTP-Refresh zurueck. Der
  nicht-transaktionale FieldDevice-Bulk-Update bleibt bewusst ausserhalb des
  Outbox-Pfads, bis seine phasenweisen Writes eine sichere Item-Transaktion
  erhalten; ein nachtraegliches autocommit Enqueue ist verboten.
- 2026-07-23: Abgelaufene Delivery-Leases sind gegen spaete Worker-Abschluesse
  abgesichert: Status und Attempt werden beim Abschluss atomar verglichen, damit
  ein alter Worker keinen neu beanspruchten Retry ueberschreibt. `go test ./...`,
  `go vet ./...`, `pnpm check`, alle 307 Frontend-Tests und der
  Produktions-Build sind erfolgreich.
- 2026-07-23: Die drei projektbezogenen Hierarchie-DELETE-Routen entfernen jetzt
  ausschliesslich den angeforderten Project-Link. Globale ControlCabinets,
  SPSController, FieldDevices sowie deren Kinder werden nie mehr durch ein
  Projekt-Unlink geloescht; fremde Projektlinks und mangels Provenance nicht
  eindeutig geerbte Kindlinks bleiben erhalten. Link-Delete, History mit
  gemeinsamer Operation-/Batch-ID und v2-Outbox-Scope committen in derselben
  Transaktion. Ein falscher ProjectID oder Outbox-Fehler rollt alles zurueck.
- 2026-07-23: Projektloeschung ist jetzt ein eigener transaktionaler
  Application-Command. Nur aktive `SUPERADMIN`/`ADMIN_FZAG` werden nach
  erneutem DB-Rollen-Lookup zugelassen; jeder verbleibende Cabinet-, SPS- oder
  FieldDevice-Projektlink fuehrt fuer abgeschlossene und nicht abgeschlossene
  Projekte zu einem typisierten Konflikt. Projekteigenes ObjectData wird in
  100er Batches mit gemeinsamer History-BatchID geloescht, Memberships,
  Projektrow und v2-Outbox-Event committen gemeinsam. PostgreSQL-FKs verhindern
  konkurrierende Link-Inserts und melden vorhandene Orphan-Zeilen, ohne Daten
  automatisch zu veraendern. Globale Facility-Entities sind fuer den Command
  nicht als Delete-Capability erreichbar.
- 2026-07-23: Projektbezogene FieldDevice-MultiCreate-Fehler geben ihre
  request-lokale `ApparatNr`-Reservierung nach dem Savepoint-Rollback frei;
  ein spaeteres gueltiges Item kann die Nummer verwenden. Die bestehende
  Transaction-Seam-Regression prueft Root-/Kind-/History-Rollback und
  anschliessende Nummernwiederverwendung.
- 2026-07-23: BACnet-Alarmwert-Replacement validiert vor jedem Write die
  ausgewaehlte AlarmType-Zuordnung. Null-, doppelte, fehlende oder zu einem
  fremden AlarmType gehoerende Field-IDs liefern exakte indexierte
  ValidationErrors; alle Field-Zuordnungen werden gebatcht und innerhalb der
  Mutationstransaktion gesperrt gelesen.
- 2026-07-23: Projektbezogene ControlCabinet-, SPSController- und
  SPSControllerSystemType-Copies verlangen innerhalb der Clone-Transaktion
  einen gesperrten Source-Link im Zielprojekt. Nicht sichtbare Quellen werden
  vor Copy, History und Outbox als NotFound abgewiesen; Events und Kopien
  bleiben ausschliesslich auf das autorisierte Zielprojekt begrenzt.
- 2026-07-23: Alle globalen und projektbezogenen Hierarchie-Copy-Commands
  verwenden fuer ihre aeussere PostgreSQL-Transaktion `REPEATABLE READ`.
  Dadurch sehen alle Keyset-Seiten denselben Source-Snapshot. SPS- und
  FieldDevice-Seiten sowie alle Copy-BulkWrites sind auf maximal 100 Items
  begrenzt; Cabinet-Copy verarbeitet jede SPS-Seite inklusive ihrer Kinder,
  bevor die naechste Seite geladen wird.
- 2026-07-23: Global ausgeloeste Cabinet-Copies uebernehmen innerhalb der
  Copy-Transaktion alle direkten Source-Projekte. Der bestehende
  ProjectFacilityLink-Service materialisiert Cabinet-, SPS- und
  FieldDevice-Links fuer die Kopie; Link-History und je ein v2-Outbox-Event
  committen atomar. Eine globale, unzugeordnete Quelle bleibt global und
  unverlinkt.
- 2026-07-23: Global ausgeloeste SPSControllerSystemType-Copies leiten ihre
  Empfaenger aus gesperrt gelesenen direkten Projektlinks des owning
  SPSControllers ab. Deduplizierte Projekt-Events werden in derselben
  Transaktion persistiert und erst nach Commit als v1-Kompatibilitaetsrefresh
  dispatcht; ein globaler owning SPS ohne Projektlink bleibt still.
- 2026-07-23: Globale Timeline-, Event- und Restore-Zugriffe laufen ueber eine
  zentrale Application-Policy. Sie laedt den aktiven Actor erneut aus der
  Datenbank und erlaubt ausschliesslich `SUPERADMIN`, `ADMIN_FZAG` und `FZAG`;
  die Permission-Middleware bleibt zusaetzliche Routensicherung. Projektbezogene
  History behaelt ihre ProjectAccessPolicy.
- 2026-07-23: Die facility-globale Semantik externer BACnet-
  SoftwareReferences ist durch einen Copy-Fidelity-Test fixiert: Referenzen
  innerhalb des Copy-Sets werden remapped, ausserhalb bleiben sie auf der
  originalen globalen Identitaet erhalten.
- 2026-07-23: Alle mutierenden FieldDevice-Batch-Routen sind auf 100 Items
  begrenzt und liefern pro Eingabe ein indexstabiles Ergebnis mit betroffener
  ID, Erfolg, stabilem Fehlercode, exaktem Feld und Grund. Das gilt fuer globale
  und projektbezogene MultiCreate-, BulkUpdate-, BulkDelete- sowie Existing-ID-
  Assignment-Pfade; alte Fehlerfelder bleiben kompatibel. Die Svelte-
  Reconciliation zeigt bevorzugt den exakten `reason` am gemeldeten Feld.
- 2026-07-23: Globale SPSController-, SPSControllerSystemType- und
  ControlCabinet-Deletes verwenden einen bounded Hierarchy-Cleaner. Er erfasst
  Nachfahren-Snapshots und History set-weise, entfernt alle zugehoerigen
  Project-Links, loest externe BACnet-SoftwareReferences und unterscheidet
  exklusive von ueber ObjectData geteilten BACnetObjects. Der PostgreSQL-Test
  beweist Root-/Kind-/Link-History und dass geteilte globale Objekte erhalten
  bleiben.
- 2026-07-23: Live-Revisions sind additiv fuer mutable Base-Entities
  ausgerollt. Alte Binaries werden durch PostgreSQL-Bump-Trigger abgesichert;
  Facility- und Project-Link-Updates nutzen Compare-and-Swap, liefern typisierte
  Konflikte und geben Revisionen in HTTP/v2-Realtime zurueck. Child-only
  FieldDevice-Bulkphasen erhoehen die Root-Aggregatrevision vor Kindwrites.
- 2026-07-23: Die Datenbank erzwingt FieldDevice-Placement inklusive
  `ApparatNr` mit einer deferred Unique-Constraint, sodass Swaps innerhalb
  derselben Transaktion atomar sind. SPS-DeviceName und GA verwenden
  normalisierte Unique-Indexes; BACnet-Template-Zuordnungen eine normalisierte
  DB-Constraint. Alle Migrations-Audits melden bestehende Konflikt-IDs und
  veraendern keine Bestandsdaten.
- 2026-07-23: ProjectSPSController- und ProjectFieldDevice-Links speichern
  normalisierte Assignment-Sources. Bestehende Links werden konservativ als
  explizit backfilled. Neue Parent-Zuweisung, exaktes Unlink, Reassignment sowie
  globale FieldDevice-/SPS-Moves addieren zuerst neue Claims und entfernen alte
  Claims danach in 100er Batches. Explizite, Copy- und Legacy-Claims bleiben
  erhalten; alte und neue Projekte erhalten dieselbe transaktionale
  Reconciliation. Die Entscheidung ist in ADR 0002 dokumentiert.
- 2026-07-23: Der opt-in PostgreSQL-Restore-Test deckt echte `to_jsonb`-
  Snapshots, erfolgreichen Before-Restore und einen Unique-Konflikt ab. Beim
  Konflikt bleiben die aktuellen Fachdaten erhalten und es wird kein
  Restore-Event committed. Damit ist die letzte in diesem Plan markierte
  PostgreSQL-Integrationsluecke geschlossen.

## Resume Notes For Future Codex Sessions

The implementation plan is complete. Future sessions should treat this section
as maintenance guidance rather than an unchecked task queue:

1. Read this file first.
2. Check `git status --short`.
3. Search for existing history code with
   `rg -n "ChangeEvent|EntityVersion|historycapture|historysql" backend`.
4. Validate new history behavior against the transaction, bounded-query,
   authorization, revision, and outbox rules in
   `FACILITY_MUTATION_ARCHITECTURE.md`.
5. Record future extensions in "Implementation Log" with their PostgreSQL and
   frontend verification evidence.
