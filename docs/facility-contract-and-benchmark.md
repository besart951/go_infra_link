# Facility contract cutover and performance benchmark

## Contract cutover

The contract migration is a maintenance operation. It must not run as part of
normal blue-green bootstrap.

Prerequisites:

- one compatible release has been deployed;
- no legacy `base_version`, copy-job route, or copy progress event has been
  observed for at least 14 days;
- API and worker instances are stopped;
- no Facility job is `queued` or `running`;
- the ownership preflight is clean;
- a restorable PostgreSQL backup has been verified.

Run from `backend`:

```powershell
$env:DATABASE_URL = '<postgres dsn>'
$env:APP_ENV = 'production'
go run ./cmd/db-maintenance `
  --backup-verified `
  --compatible-release-delivered `
  --applications-stopped `
  --legacy-idle-since 2026-08-01T00:00:00Z
```

The command acquires a transaction-scoped PostgreSQL advisory lock, repeats
all preflight checks, applies the destructive DDL transactionally, performs
postflight checks, and records schema version `202609050001`. There is no down
migration. Recovery after DDL starts means restoring the verified backup and
starting the previous application version.

## Five-million-FieldDevice benchmark

The benchmark database name must end in `_benchmark`; the seeder refuses any
other database. The Docker profile uses PostgreSQL 18.3, 4 CPUs, 8 GiB memory,
2 GiB shared memory, synchronous commits, `fsync=on`, `track_io_timing=on`, and
`jit=off`. Disabling JIT is intentional for these short OLTP cursor queries;
the generated-code setup cost dominates their execution under concurrency.

```powershell
docker compose -f docker-compose.benchmark.yml up -d --wait

Set-Location backend
$dsn = 'postgres://benchmark:benchmark@localhost:55433/go_infra_link_benchmark?sslmode=disable'
go run ./cmd/facility-benchmark -dsn $dsn seed
go run ./cmd/facility-benchmark -dsn $dsn analyze
go run ./cmd/facility-benchmark -dsn $dsn -output ../benchmark-results measure
```

For a fast diagnosis against the same production reader, select one exact
scenario, for example `-smoke -scenario search_0_1_percent_next`.

`measure` runs 10 warmups, 100 sequential samples, and 800 samples from eight
parallel readers for every canonical scenario. It writes JSON and Markdown,
including SQL, `EXPLAIN (ANALYZE, BUFFERS, WAL, FORMAT JSON)`, PostgreSQL
configuration, relation sizes, migration versions, latency percentiles, and
error rates. The command fails if any scenario has errors, p95 exceeds 750 ms,
or a query-plan invariant fails.

The production reader uses three PostgreSQL projections for stable keyset
access: `field_device_cursor_values`, project-linked creation timestamps, and
`field_device_building_cursor_values`. Search first probes at most 50,000 rows
in cursor order. If that cannot produce a complete page, the immutable query
falls back to the combined BMK/description trigram index. Lifecycle visibility
uses a statement-level empty-set guard; the full owner hierarchy is still
checked whenever a relevant aggregate is staging, deleting, or restoring.

The final local acceptance run on 2026-08-22 used exactly 5,000,000
FieldDevices and 184 scenarios. All 165,600 measured requests completed
without error and all plan checks passed. The worst p95 was 679.22 ms
(`sps_system_type_desc_p99_next`), below the 750 ms release gate. The complete
artifacts are in `benchmark-results-5m-contract-final-full`.

Cold-cache measurements are operational measurements and are recorded
separately: restart the dedicated database host (or clear the host page cache
using the benchmark operator's approved procedure), run one measurement pass,
and retain that report next to the warm-cache gate. Cold-cache values are not
the release gate.

Pull requests run the same query-shape checks with 50,000 FieldDevices. The
scheduled/manual workflow uses the full five-million-row dataset and publishes
both reports as CI artifacts.
