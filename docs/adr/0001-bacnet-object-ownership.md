# ADR 0001: BACnet object ownership and lifecycle

- Status: Accepted
- Date: 2026-07-23

## Context

A `BacnetObject` is a global facility entity with two independent attachment
paths:

- `field_device_id` is a nullable direct attachment to one `FieldDevice`;
- `object_data_bacnet_objects` attaches the same object to one or more
  `ObjectData` templates.

The old pointer-only update payload could not distinguish an omitted
`field_device_id` from an explicit JSON `null`. Database cascades could also
delete a directly attached object without considering template attachments or
recording alarm-value and join history.

## Decision

`BacnetObject` identity is global. Attachments describe use and lifecycle
claims; neither a project nor an `ObjectData` row owns that global identity
exclusively.

1. A direct `FieldDevice` attachment is optional and at most one.
2. An object may be attached to any number of `ObjectData` templates.
3. Attaching an existing object never copies or transfers its identity.
4. Detaching from a FieldDevice sets `field_device_id` to null. Detaching from
   ObjectData deletes only the selected join row.
5. An automatic parent cleanup deletes the BACnet object only after its direct
   attachment and all ObjectData attachments are absent. Otherwise it preserves
   the object and removes only the attachment owned by the deleted parent.
6. An explicit global BACnet delete removes every attachment and dependent
   alarm value in one authorized transaction, records each removal in history,
   and publishes project reconciliation only after commit.
7. Multiple ObjectData attachments are supported. Template software
   `(normalized software_type, software_number)` remains unique independently
   within each ObjectData. Duplicate `TextFix` values are allowed.
8. API patches use presence-aware nullable fields for detach. Omitted means
   unchanged; JSON `null` means detach; a UUID means attach/reassign.
9. Restore recreates alarm values and attachment rows before exposing the
   restored object. Association rows therefore need stable audit identity and
   snapshots; database cascades are not an audit mechanism.
10. Collaboration recipients are the union of projects reached through the
    direct FieldDevice and every attached ObjectData. Mixed ownership uses a
    project-level reconciliation event so neither view remains stale.

## Consequences

Hierarchy deletion must page BACnet descendants. For each object it determines
whether deleting the parent removes the last attachment, then either detaches
or deletes explicitly. It must not rely on `ON DELETE CASCADE`.

The ObjectData join requires an auditable association model and migration.
Existing rows are retained and backfilled; conflicting or orphaned rows are
reported rather than silently deleted. Until that migration and restore support
are deployed, global descendant deletion must fail safely when it encounters a
mixed-ownership object it cannot audit correctly.
