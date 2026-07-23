# AGENTS.md

## Stack

* Backend: Go
* Database: PostgreSQL
* Frontend: Svelte 5 SPA
* TypeScript strict mode
* Tailwind CSS
* shadcn-svelte

## Domain hierarchy

```text
Building
└── ControlCabinet
    └── SPSController
        └── SPSControllerSystemType
            └── FieldDevice
                └── BACnetObject
```

A parent can contain multiple children.

## Global entities and projects

`ControlCabinet`, `SPSController`, `SPSControllerSystemType`, `FieldDevice`, and `BACnetObject` are global facility entities.

They can exist without any project.
A project only links or assigns existing facility entities. A project does not own them.
Never assume that a hierarchy entity must belong to a project.
Never delete a global hierarchy entity when removing it from a project.
Keep these operations separate:

* global create, update, move, copy, and delete;
* assign entity to project;
* remove entity from project;
* project-scoped copy.

When copying:

* an entity assigned to a project is copied inside that project;
* a global entity without a project is copied globally without creating project links.

## Authorization

Authorization must always be enforced in the Go backend.
FZAG workers with the required permission may manage global facility entities.
Project users may access only projects where they are members.
Every project operation must validate:

* authenticated user;
* `ProjectID`;
* project membership;
* required permission;
* access to the affected entity.

Frontend visibility is not authorization.
Do not spread role string comparisons across handlers. Use central authorization policies.

## Project deletion

A project may be deleted only when:

* it is completed and all hierarchy links have been removed; or
* it is not completed but has no linked hierarchy entities.

Only `SUPERADMIN` and `ADMIN_FZAG` may delete projects.

Deleting a project must never delete global facility entities.

## Real-time collaboration

Users inside the same project must see:

* other active users;
* editing state;
* committed creates, updates, moves, copies, deletes, and assignments.

Project events must never be sent to users outside that project.
The database is the source of truth.
Publish collaboration events only after the transaction commits.
Use a transactional outbox with retries and idempotent consumers.
Presence may remain process-local because the application runs as one monolithic server.

## Concurrent editing

Every mutable entity must have a revision or version.
Update commands must contain the revision last seen by the client.
Reject stale updates with a typed conflict result.
Never silently overwrite another user’s committed changes.
Real-time editing state does not replace revision checks.

## Backend architecture

Use this dependency direction:

```text
Transport → Application → Domain
Infrastructure → Ports → Application
```

Handlers must remain thin.
Business rules must not exist in HTTP, RPC, WebSocket, PostgreSQL, or frontend adapters.
Each mutation must use one application command and one explicit transaction boundary.
Where applicable, the same transaction must include:

* entity mutation;
* project-link mutation;
* history;
* revision update;
* outbox event.

## Bulk operations

Bulk operations must return one structured result per item.
The frontend must receive:

* success or failure;
* affected entity ID;
* exact error code;
* exact invalid field;
* reason for the failure.

A `BatchID` is only for correlation. It does not mean that the complete request is atomic.
Process large workloads in sequential batches of at most 100 items. Adapt processing to available server resources.
Release a reserved `ApparatNr` when its item fails.

`ApparatNr` uniqueness must be enforced by PostgreSQL within its defined scope.

## Database integrity

Use database constraints for rules that must remain correct under concurrency.
Do not rely only on service-level uniqueness queries.
Normalize SPS names and GA values and use normalized unique indexes.
Use an atomic swap strategy when exchanging unique `ApparatNr` values.

Existing conflicting data must be reported by migrations. Do not silently modify or delete it.
Duplicate BACnet `TextFix` values are allowed.

## History and deletion

Every committed business mutation must create history.
Database cascades must not bypass required audit records.
For hierarchy deletion:

* capture bounded descendant snapshots;
* write history set-wise;
* clean up all related project links;
* avoid loading the complete facility hierarchy into memory.

## Frontend

Use three layers:

```text
API transport
State and application logic
Svelte UI
```

Use Svelte 5 runes consistently.
Do not place business rules, authorization, or complex synchronization logic inside components.
Remote changes must not silently replace unsaved form values.
Show:

* active project;
* connected users;
* editing state;
* save state;
* field-level validation errors;
* revision conflicts;
* connection state.

## Agent checklist

Before changing code, determine:

1. Is the operation global or project-scoped?
2. Does it change an entity or only a project link?
3. Which backend permission is required?
4. What is the transaction boundary?
5. Which history records are required?
6. Which collaboration event is required?
7. How are concurrent edits handled?
8. Which frontend state must be refreshed?

Never:

* treat project links as entity ownership;
* require global entities to belong to a project;
* delete global entities during project unlinking;
* publish events before commit;
* trust frontend authorization;
* silently overwrite stale revisions;
* load an unbounded hierarchy into memory;
* return only a generic bulk error.
