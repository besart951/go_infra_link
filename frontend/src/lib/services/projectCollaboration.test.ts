import type {
  RealtimeSocketConnection,
  ReconnectingWebSocketOptions
} from '$lib/infrastructure/realtime/reconnectingWebSocket.js';
import { afterEach, describe, expect, it, vi } from 'vitest';

vi.mock('$lib/i18n/index.js', () => ({
  t: (key: string) => key
}));

import {
  parseProjectSyncInboundMessage,
  ProjectSyncCoordinator,
  type ProjectChange,
  type ProjectChangesResponse,
  type ProjectSyncCoordinatorOptions
} from './projectCollaboration.svelte.js';

const projectId = '11111111-1111-4111-8111-111111111111';
const otherProjectId = '55555555-5555-4555-8555-555555555555';
const userId = '22222222-2222-4222-8222-222222222222';
const deviceId = '33333333-3333-4333-8333-333333333333';
const otherDeviceId = '66666666-6666-4666-8666-666666666666';
const eventId = '44444444-4444-4444-8444-444444444444';

class FakeConnection implements RealtimeSocketConnection {
  readonly sent: Array<Record<string, unknown>> = [];

  constructor(private readonly options: ReconnectingWebSocketOptions) {}

  connect(): void {
    this.options.onStatusChange?.('connected');
    this.options.onOpen?.({ wasReconnect: false });
  }

  disconnect(): void {
    this.options.onStatusChange?.('disconnected');
  }

  send(payload: Record<string, unknown>): void {
    this.sent.push(payload);
  }

  emit(payload: unknown): void {
    this.options.onMessage(JSON.stringify(payload));
  }
}

function change(revision: number): ProjectChange {
  return {
    revision,
    event_id: eventId,
    aggregate_type: 'field_device',
    aggregate_id: deviceId,
    action: 'updated',
    changed_fields: ['bmk'],
    occurred_at: '2026-01-01T10:00:00Z'
  };
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
}

function setup(options: Omit<ProjectSyncCoordinatorOptions, 'createConnection'> = {}) {
  let connection: FakeConnection | undefined;
  const coordinator = new ProjectSyncCoordinator({
    ...options,
    createConnection: (connectionOptions) => {
      connection = new FakeConnection(connectionOptions);
      return connection;
    }
  });
  coordinator.connect(projectId);
  if (!connection) throw new Error('connection was not created');
  return { coordinator, connection };
}

afterEach(() => {
  vi.useRealTimers();
});

describe('project sync message validation', () => {
  it('parses the v2 snapshot, change, revision, presence, and draft messages', () => {
    const snapshot = parseProjectSyncInboundMessage({
      type: 'snapshot',
      project_id: projectId,
      current_revision: 12,
      presence: [
        {
          user_id: userId,
          connected_at: '2026-01-01T10:00:00Z',
          last_seen_at: '2026-01-01T10:00:01Z'
        }
      ],
      draft_states: [],
      at: '2026-01-01T10:00:02Z'
    });
    expect(snapshot.type).toBe('snapshot');

    expect(
      parseProjectSyncInboundMessage({
        type: 'project_change',
        project_id: projectId,
        ...change(13)
      }).type
    ).toBe('project_change');
    expect(
      parseProjectSyncInboundMessage({
        type: 'revision',
        project_id: projectId,
        current_revision: 13,
        at: '2026-01-01T10:00:03Z'
      }).type
    ).toBe('revision');
    expect(
      parseProjectSyncInboundMessage({
        type: 'presence',
        project_id: projectId,
        presence: [],
        at: '2026-01-01T10:00:04Z'
      }).type
    ).toBe('presence');
    expect(
      parseProjectSyncInboundMessage({
        type: 'draft_states',
        project_id: projectId,
        draft_states: [
          {
            user_id: userId,
            entries: [
              {
                aggregate_type: 'field_device',
                aggregate_id: deviceId,
                action: 'update',
                base_version: 12,
                fields: [{ path: 'bmk', value: 'draft' }]
              }
            ],
            updated_at: '2026-01-01T10:00:05Z'
          }
        ],
        at: '2026-01-01T10:00:05Z'
      }).type
    ).toBe('draft_state');
  });

  it('rejects malformed identifiers and incomplete changes', () => {
    expect(() =>
      parseProjectSyncInboundMessage({
        type: 'revision',
        project_id: 'not-a-uuid',
        current_revision: -1,
        at: 'now'
      })
    ).toThrow();
    expect(() =>
      parseProjectSyncInboundMessage({
        type: 'project_change',
        project_id: projectId,
        revision: 1
      })
    ).toThrow();
  });
});

describe('ProjectSyncCoordinator', () => {
  it('catches up from zero when the initial snapshot has a revision watermark', async () => {
    const applied: number[] = [];
    const fetchChanges = vi.fn().mockResolvedValue({
      project_id: projectId,
      current_revision: 2,
      events: [change(1), change(2)],
      has_more: false,
      reset_required: false
    });
    const { coordinator, connection } = setup({
      fetchChanges,
      onProjectChange: (event: ProjectChange) => applied.push(event.revision)
    });

    connection.emit({
      type: 'snapshot',
      project_id: projectId,
      current_revision: 2,
      presence: [],
      drafts: [],
      at: '2026-01-01T10:00:00Z'
    });

    await vi.waitFor(() => expect(applied).toEqual([1, 2]));
    expect(fetchChanges).toHaveBeenCalledWith(projectId, 0);
    expect(coordinator.currentRevision).toBe(2);
  });

  it('buffers an out-of-order change and catches up the missing revision', async () => {
    const applied: number[] = [];
    const fetchChanges = vi.fn().mockResolvedValue({
      project_id: projectId,
      current_revision: 2,
      events: [change(1)],
      has_more: false,
      reset_required: false
    });
    const { coordinator, connection } = setup({
      fetchChanges,
      onProjectChange: (event: ProjectChange) => applied.push(event.revision)
    });

    connection.emit({
      type: 'snapshot',
      project_id: projectId,
      current_revision: 0,
      presence: [],
      drafts: [],
      at: '2026-01-01T10:00:00Z'
    });
    connection.emit({ type: 'project_change', project_id: projectId, ...change(2) });

    await vi.waitFor(() => expect(applied).toEqual([1, 2]));
    expect(fetchChanges).toHaveBeenCalledWith(projectId, 0);
    expect(coordinator.currentRevision).toBe(2);
  });

  it('keeps catching up when a higher watermark arrives during an in-flight request', async () => {
    const firstResponse = deferred<ProjectChangesResponse>();
    const applied: number[] = [];
    const fetchChanges = vi
      .fn()
      .mockImplementationOnce(() => firstResponse.promise)
      .mockResolvedValueOnce({
        project_id: projectId,
        current_revision: 4,
        events: [change(4)],
        has_more: false,
        reset_required: false
      });
    const { coordinator, connection } = setup({
      fetchChanges,
      onProjectChange: (event: ProjectChange) => applied.push(event.revision)
    });

    connection.emit({ type: 'project_change', project_id: projectId, ...change(1) });
    connection.emit({
      type: 'revision',
      project_id: projectId,
      current_revision: 3,
      at: '2026-01-01T10:00:00Z'
    });
    await vi.waitFor(() => expect(fetchChanges).toHaveBeenCalledWith(projectId, 1));
    connection.emit({
      type: 'revision',
      project_id: projectId,
      current_revision: 4,
      at: '2026-01-01T10:00:01Z'
    });
    firstResponse.resolve({
      project_id: projectId,
      current_revision: 3,
      events: [change(2), change(3)],
      has_more: false,
      reset_required: false
    });

    await vi.waitFor(() => expect(applied).toEqual([1, 2, 3, 4]));
    expect(fetchChanges).toHaveBeenNthCalledWith(2, projectId, 3);
    expect(coordinator.currentRevision).toBe(4);
  });

  it('retries a failed catch-up while the observed watermark remains ahead', async () => {
    const fetchChanges = vi
      .fn()
      .mockRejectedValueOnce(new Error('temporary failure'))
      .mockResolvedValueOnce({
        project_id: projectId,
        current_revision: 2,
        events: [change(2)],
        has_more: false,
        reset_required: false
      });
    const { coordinator, connection } = setup({ fetchChanges, catchUpRetryMs: 1 });

    connection.emit({ type: 'project_change', project_id: projectId, ...change(1) });
    connection.emit({
      type: 'revision',
      project_id: projectId,
      current_revision: 2,
      at: '2026-01-01T10:00:00Z'
    });

    await vi.waitFor(() => expect(fetchChanges).toHaveBeenCalledTimes(2));
    await vi.waitFor(() => expect(coordinator.currentRevision).toBe(2));
    expect(coordinator.syncError).toBeNull();
  });

  it('ignores a rejected catch-up from a previous project generation', async () => {
    const firstResponse = deferred<never>();
    const fetchChanges = vi.fn().mockImplementation(() => firstResponse.promise);
    const { coordinator, connection } = setup({ fetchChanges });

    connection.emit({
      type: 'revision',
      project_id: projectId,
      current_revision: 1,
      at: '2026-01-01T10:00:00Z'
    });
    await vi.waitFor(() => expect(fetchChanges).toHaveBeenCalledWith(projectId, 0));

    coordinator.connect(otherProjectId);
    firstResponse.reject(new Error('old project failed'));
    await Promise.resolve();
    await Promise.resolve();

    expect(coordinator.syncError).toBeNull();
    expect(coordinator.currentRevision).toBe(0);
  });

  it('requests a full refresh when catch-up says the cursor is no longer available', async () => {
    const onResetRequired = vi.fn();
    const { coordinator, connection } = setup({
      onResetRequired,
      fetchChanges: vi.fn().mockResolvedValue({
        project_id: projectId,
        current_revision: 99,
        events: [],
        has_more: false,
        reset_required: true
      })
    });
    connection.emit({
      type: 'snapshot',
      project_id: projectId,
      current_revision: 2,
      presence: [],
      drafts: [],
      at: '2026-01-01T10:00:00Z'
    });
    connection.emit({ type: 'revision', project_id: projectId, current_revision: 99, at: 'now' });

    await vi.waitFor(() => expect(onResetRequired).toHaveBeenCalledOnce());
    expect(coordinator.currentRevision).toBe(99);
  });

  it('debounces drafts and clears drafts removed from the local overlay', () => {
    vi.useFakeTimers();
    const { coordinator, connection } = setup({ draftDebounceMs: 50 });

    coordinator.publishFieldDeviceDraftState({
      devices: [
        {
          device_id: deviceId,
          changed_fields: ['bmk'],
          field_values: { bmk: 'first' }
        }
      ]
    });
    coordinator.publishFieldDeviceDraftState({
      devices: [
        {
          device_id: deviceId,
          changed_fields: ['bmk'],
          field_values: { bmk: 'latest' }
        }
      ]
    });
    vi.advanceTimersByTime(50);

    expect(connection.sent).toEqual([
      {
        type: 'draft_state',
        entries: [
          {
            aggregate_type: 'field_device',
            aggregate_id: deviceId,
            action: 'update',
            base_version: 0,
            fields: [{ path: 'bmk', value: 'latest' }]
          }
        ]
      }
    ]);

    coordinator.publishFieldDeviceDraftState({ devices: [] });
    vi.advanceTimersByTime(50);
    expect(connection.sent.at(-1)).toEqual({
      type: 'draft_clear',
      aggregate_type: 'field_device',
      aggregate_id: deviceId
    });
  });

  it('clears one draft without cancelling a pending update for another draft', () => {
    vi.useFakeTimers();
    const { coordinator, connection } = setup({ draftDebounceMs: 50 });
    const first = {
      aggregate_type: 'field_device',
      aggregate_id: deviceId,
      action: 'update',
      base_version: 1,
      fields: [{ path: 'bmk', value: 'first' }]
    };
    const second = {
      aggregate_type: 'field_device',
      aggregate_id: otherDeviceId,
      action: 'update',
      base_version: 1,
      fields: [{ path: 'bmk', value: 'second' }]
    };

    coordinator.publishDraftState([first]);
    vi.advanceTimersByTime(50);
    coordinator.publishDraftState([
      { ...first, fields: [{ path: 'bmk', value: 'latest' }] },
      second
    ]);
    coordinator.clearDraft('field_device', { aggregateId: deviceId });
    vi.advanceTimersByTime(50);

    expect(connection.sent.at(-2)).toEqual({
      type: 'draft_clear',
      aggregate_type: 'field_device',
      aggregate_id: deviceId
    });
    expect(connection.sent.at(-1)).toEqual({ type: 'draft_state', entries: [second] });
  });

  it('clears known local drafts before disconnecting', () => {
    vi.useFakeTimers();
    const { coordinator, connection } = setup({ draftDebounceMs: 50 });

    coordinator.publishFieldDeviceDraftState({
      devices: [
        {
          device_id: deviceId,
          changed_fields: ['bmk'],
          field_values: { bmk: 'draft' }
        }
      ]
    });
    vi.advanceTimersByTime(50);
    coordinator.disconnect();

    expect(connection.sent.at(-1)).toEqual({
      type: 'draft_clear',
      aggregate_type: 'field_device',
      aggregate_id: deviceId
    });
  });

  it('derives collaborators by aggregate field and keeps the field-device adapter', () => {
    const { coordinator, connection } = setup();
    connection.emit({
      type: 'snapshot',
      project_id: projectId,
      current_revision: 0,
      presence: [],
      drafts: [
        {
          user_id: userId,
          entries: [
            {
              aggregate_type: 'field_device',
              aggregate_id: deviceId,
              action: 'update',
              base_version: 1,
              fields: [{ path: 'bmk', value: 'remote draft' }]
            }
          ],
          updated_at: '2026-01-01T10:00:00Z'
        }
      ],
      at: '2026-01-01T10:00:00Z'
    });

    expect(coordinator.collaboratorsByField[`field_device:${deviceId}`].bmk).toEqual([
      { userId, value: 'remote draft', updatedAt: '2026-01-01T10:00:00Z' }
    ]);
    expect(coordinator.buildFieldDeviceEditorsByDevice(new Map())[deviceId][0]).toMatchObject({
      userId,
      changedFields: ['bmk'],
      fieldValues: { bmk: 'remote draft' }
    });
  });
});
