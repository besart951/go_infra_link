# Transaction Boundaries

## Intent

Multi-step application use-cases must either commit as one unit or roll back as one unit. Application services must not depend on GORM or on a concrete database transaction handle.

## Application Port

The application-level port lives in `backend/internal/application/transaction`.

- `Runner` starts a transaction and invokes a callback with an opaque `UnitOfWork`.
- `Factory` turns that opaque unit into a service bundle backed by transaction-scoped repositories.
- `Operation` binds the currently wired service to the transactional service bundle and keeps local development mode working when no transaction runner is configured.

`UnitOfWork` is intentionally opaque. Domain and application services must not type assert it.

## Infrastructure Adapter

The GORM adapter lives in `backend/internal/infrastructure/transaction`.

- `NewGormRunner(db)` wraps `db.WithContext(ctx).Transaction(...)`.
- `GormDB(unit)` is only used from `wire` to rebuild repository adapters against the transaction handle.

This keeps `*gorm.DB` in infrastructure/wiring. Facility and Project services only depend on the application port.

## Covered Use-Cases

Facility use-cases now route critical multi-step writes through the transaction boundary:

- field device create/update with BACnet object selection
- field device specification creation
- BACnet object create/update/replace-for-object-data
- SPS controller create/update with system types
- control cabinet update with SPS device-name regeneration
- hierarchy copy flows for control cabinets, SPS controllers, and SPS controller system types

Project use-cases now route critical link changes through the same boundary:

- project creation plus creator/template initialization
- project facility-link create/update/delete
- project copy flows that create facility hierarchy and project links
- multi-create-and-assign field devices

Duplicate project-link writes are translated to `domain.ErrConflict` inside the `projectsql` adapter. The project application service no longer checks GORM errors.

## Rollback Tests

Rollback and seam coverage exists in:

- `backend/internal/infrastructure/transaction/gorm_runner_test.go`
- `backend/internal/application/transaction/operation_test.go`
- `backend/internal/service/facility/transaction_seam_test.go`
- `backend/internal/service/project/transaction_seam_test.go`
- `backend/internal/repository/projectsql/project_link_cleanup_repo_test.go`

The GORM runner test verifies that a write performed before a later error is rolled back.
