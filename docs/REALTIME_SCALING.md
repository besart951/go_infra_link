# Realtime Scaling

The WebSocket endpoints stay process-local: each backend instance owns only the clients connected to that instance. Cross-instance fanout is handled by the application `RealtimeBus` interface in `backend/internal/application/realtime`.

## Adapters

- `memory`: default for local development and tests. It keeps all events in-process and requires no external dependency.
- `postgres`: production adapter using PostgreSQL LISTEN/NOTIFY plus the `realtime_events` table. NOTIFY carries only the event id; the JSON payload is read from the table, so larger project deltas are not limited by the NOTIFY payload size.

Enable PostgreSQL fanout with:

```env
REALTIME_BUS=postgres
REALTIME_POSTGRES_CHANNEL=go_infra_link_realtime
REALTIME_SUBSCRIBER_BUFFER=64
REALTIME_EVENT_TTL=10m
```

`REALTIME_NODE_ID` is optional. If omitted, each process gets a generated id. The id is used to ignore events that the same process already delivered to its local clients.

## Distributed vs Local State

Distributed across backend instances:

- Project refresh requests.
- Project entity deltas for control cabinets, SPS controllers, and field devices.
- System notification created/updated/deleted/read-all events.
- Project field-device edit-state changes after they are observed by a connected instance.

Local to one backend instance for now:

- Presence snapshots and presence updates. They only include users connected to the same backend process.
- Initial edit-state completeness is best-effort across instances. An instance combines local edit states with remote edit-state events it has observed. A newly started instance will not reconstruct stale edit states from PostgreSQL; disconnects publish an empty edit state to clear remote views.

The hubs never hold client-map locks while writing to WebSocket send buffers. Slow local clients are disconnected when their bounded send channel is full. Slow bus subscribers drop events with `ErrBackpressure` rather than blocking publishers indefinitely.
