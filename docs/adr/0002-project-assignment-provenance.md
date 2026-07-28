# ADR 0002: Project assignment provenance

- Status: Accepted
- Date: 2026-07-23

## Context

`ControlCabinet`, `SPSController`, and `FieldDevice` are global facility
entities. A project link grants access but never owns the linked entity.

Assigning a parent materializes links for its current descendants so project
queries remain simple. The same descendant can also be assigned directly or
through several parents. A single `project_sps_controllers` or
`project_field_devices` row therefore cannot explain why access exists.
Deleting or moving one parent without that explanation either revokes valid
explicit access or leaves stale inherited access.

## Decision

Project hierarchy links remain the materialized access projection. Normalized
source rows record every reason for that projection:

- `explicit`: a user assigned the exact SPSController or FieldDevice;
- `control_cabinet`: the link is inherited from a linked ControlCabinet;
- `sps_controller`: a FieldDevice link is inherited from a linked
  SPSController;
- `sps_controller_system_type`: a project-scoped system-type copy materialized
  the copied FieldDevices.

A materialized SPSController or FieldDevice link remains while at least one
source row claims it. Removing one source deletes the link only when no source
remains. ControlCabinet links are roots and need no separate source table.

The `sps_controller_system_type` source is a copy provenance claim, not a live
parent subscription: there is no ProjectSPSControllerSystemType root link to
unlink later. It survives a FieldDevice move, like an explicit assignment.

All new assignment paths write the materialized link and its source in the same
transaction. Parent assignment, exact unlink, reassignment, and global
FieldDevice/SPSController moves add new claims before pruning old claims. This
order preserves the same materialized row when old and new parents are linked
to the same project. A moved FieldDevice receives an `sps_controller` claim only
from projects whose SPS link has an explicit source; a cabinet-inherited SPS
projection must not become a new direct claim. Source pruning uses ordered
batches of at most 100 rows.
Materialized link create/delete history and the collaboration outbox share the
mutation BatchID and transaction.

The expand migration backfills every existing materialized link as
`explicit`. It never infers ancestry from current placement, because doing so
could revoke legacy access after a later unlink or move. Consequently,
post-migration source-aware operations are exact while legacy rows remain
conservatively accessible until a user explicitly changes them.

Source rows are internal projection metadata rather than separately exposed
business entities. Adding or removing a redundant source does not emit a
second history event when project access is unchanged. A materialized access
change always creates the existing project-link history event.

Project-link reassignment is allowed only when the selected materialized link
has the expected explicit source and no inherited source. A multiply claimed
link returns a typed conflict instead of silently moving access granted by a
parent. Collaboration reconciliation covers both old and new entity or
hierarchy scopes.

## Consequences

Projects never own or delete global facility entities. Exact unlink and global
delete remain distinct commands.

The migration is additive and blue-green compatible. Older binaries ignore the
source tables; newer binaries fall back to conservative link behavior when the
tables are unavailable. The source tables can be removed independently during
rollback without deleting materialized project links.

Operational audits should compare materialized links with source claims after
the cutover. A link with no source is invalid for newly written data but may be
preserved during mixed-version operation until the backfill has completed.
