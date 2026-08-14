import type { User } from '$lib/domain/user/index.js';
import { apiClient } from '$lib/api/generated/client.js';
import type {
  ReconnectingWebSocketOptions,
  RealtimeSocketConnection,
  RealtimeSocketStatus
} from '$lib/infrastructure/realtime/reconnectingWebSocket.js';
import {
  buildSameOriginWebSocketUrl,
  RealtimeJsonStream
} from '$lib/infrastructure/realtime/realtimeJsonStream.js';
import { createContext } from 'svelte';
import { z } from 'zod';

export interface ProjectCollaboratorPresence {
  user_id: string;
  connected_at: string;
  last_seen_at: string;
}

export interface ProjectDraftField {
  path: string;
  value: unknown;
}

export interface ProjectDraftEntry {
  aggregate_type: string;
  aggregate_id?: string;
  draft_id?: string;
  action: string;
  base_version: number;
  fields: ProjectDraftField[];
}

export interface ProjectCollaboratorDraftState {
  user_id: string;
  entries: ProjectDraftEntry[];
  updated_at: string;
}

export interface ProjectChange {
  revision: number;
  event_id: string;
  aggregate_type: string;
  aggregate_id: string;
  action: string;
  actor_id?: string;
  changed_fields: string[];
  parent_refs?: Record<string, unknown>;
  occurred_at: string;
}

export interface ProjectChangeMessage extends ProjectChange {
  type: 'project_change';
  project_id: string;
}

interface ProjectSnapshotMessage {
  type: 'snapshot';
  project_id: string;
  current_revision: number;
  presence: ProjectCollaboratorPresence[];
  drafts: ProjectCollaboratorDraftState[];
  at: string;
}

interface ProjectRevisionMessage {
  type: 'revision';
  project_id: string;
  current_revision: number;
  at: string;
}

interface ProjectPresenceMessage {
  type: 'presence';
  project_id: string;
  presence: ProjectCollaboratorPresence[];
  at: string;
}

interface ProjectDraftsMessage {
  type: 'draft_state';
  project_id: string;
  drafts: ProjectCollaboratorDraftState[];
  at: string;
}

interface ProjectDraftClearMessage {
  type: 'draft_clear';
  project_id: string;
  user_id: string;
  aggregate_type: string;
  aggregate_id?: string;
  draft_id?: string;
  at: string;
}

export type ProjectSyncInboundMessage =
  | ProjectSnapshotMessage
  | ProjectChangeMessage
  | ProjectRevisionMessage
  | ProjectPresenceMessage
  | ProjectDraftsMessage
  | ProjectDraftClearMessage;

export interface ProjectChangesResponse {
  project_id?: string;
  current_revision: number;
  events: ProjectChange[];
  has_more: boolean;
  reset_required: boolean;
}

export interface SharedFieldDeviceDraftState {
  devices: Array<{
    device_id: string;
    specification_id?: string;
    changed_fields: string[];
    field_values?: Record<string, unknown>;
  }>;
}

export interface SharedFieldDeviceEditor {
  userId: string;
  firstName: string;
  lastName: string;
  changedFields: string[];
  fieldValues?: Record<string, unknown>;
  updatedAt: string;
}

export type SharedFieldDeviceEditorsByDevice = Record<string, SharedFieldDeviceEditor[]>;

export interface DraftCollaborator {
  userId: string;
  value: unknown;
  updatedAt: string;
}

export type ProjectCollaboratorsByField = Record<string, Record<string, DraftCollaborator[]>>;

export interface ProjectSyncCoordinatorOptions {
  onProjectChange?: (change: ProjectChange) => void;
  onResetRequired?: () => void;
  onReconnect?: () => void;
  fetchChanges?: (projectId: string, afterRevision: number) => Promise<ProjectChangesResponse>;
  createConnection?: (options: ReconnectingWebSocketOptions) => RealtimeSocketConnection;
  draftDebounceMs?: number;
  catchUpRetryMs?: number;
}

const uuidSchema = z.string().uuid();
const dateTimeSchema = z.string().min(1).max(80);
const revisionSchema = z.number().int().nonnegative();
const boundedStringSchema = z.string().min(1).max(180);

const presenceSchema = z
  .object({
    user_id: uuidSchema,
    connected_at: dateTimeSchema,
    last_seen_at: dateTimeSchema
  })
  .strip();

const draftFieldSchema = z
  .object({
    path: z.string().min(1).max(180),
    value: z.unknown()
  })
  .strict();

const draftEntrySchema = z
  .object({
    aggregate_type: boundedStringSchema,
    aggregate_id: uuidSchema.optional(),
    draft_id: uuidSchema.optional(),
    action: boundedStringSchema,
    base_version: revisionSchema,
    fields: z.array(draftFieldSchema).max(128)
  })
  .strict()
  .refine((entry) => entry.aggregate_id || entry.draft_id, {
    message: 'aggregate_id or draft_id is required'
  });

const collaboratorDraftSchema = z
  .object({
    user_id: uuidSchema,
    entries: z.array(draftEntrySchema).max(250),
    updated_at: dateTimeSchema
  })
  .strip();

const parentRefsSchema = z.record(z.string().max(180), z.unknown());

const projectChangeFields = {
  revision: revisionSchema,
  event_id: uuidSchema,
  aggregate_type: boundedStringSchema,
  aggregate_id: uuidSchema,
  action: boundedStringSchema,
  actor_id: uuidSchema.optional(),
  changed_fields: z.array(z.string().min(1).max(180)).max(128),
  parent_refs: parentRefsSchema.optional(),
  occurred_at: dateTimeSchema
} as const;

const projectChangeSchema = z.object(projectChangeFields).strip();

const projectSyncInboundSchema = z.discriminatedUnion('type', [
  z
    .object({
      type: z.literal('snapshot'),
      project_id: uuidSchema,
      current_revision: revisionSchema,
      presence: z.array(presenceSchema).max(250),
      drafts: z.array(collaboratorDraftSchema).max(250),
      at: dateTimeSchema
    })
    .strip(),
  z
    .object({
      type: z.literal('project_change'),
      project_id: uuidSchema,
      ...projectChangeFields
    })
    .strip(),
  z
    .object({
      type: z.literal('revision'),
      project_id: uuidSchema,
      current_revision: revisionSchema,
      at: dateTimeSchema
    })
    .strip(),
  z
    .object({
      type: z.literal('presence'),
      project_id: uuidSchema,
      presence: z.array(presenceSchema).max(250),
      at: dateTimeSchema
    })
    .strip(),
  z
    .object({
      type: z.literal('draft_state'),
      project_id: uuidSchema,
      drafts: z.array(collaboratorDraftSchema).max(250),
      at: dateTimeSchema
    })
    .strip(),
  z
    .object({
      type: z.literal('draft_clear'),
      project_id: uuidSchema,
      user_id: uuidSchema,
      aggregate_type: boundedStringSchema,
      aggregate_id: uuidSchema.optional(),
      draft_id: uuidSchema.optional(),
      at: dateTimeSchema
    })
    .strip()
]);

const projectChangesResponseSchema = z
  .object({
    project_id: uuidSchema.optional(),
    current_revision: revisionSchema.optional(),
    events: z.array(projectChangeSchema).max(500).optional(),
    has_more: z.boolean().optional(),
    reset_required: z.boolean().optional(),
    items: z.array(projectChangeSchema).max(500).optional(),
    next_revision: revisionSchema.optional()
  })
  .strip();

export class ProjectSyncCoordinator {
  onlineUsers = $state.raw<ProjectCollaboratorPresence[]>([]);
  collaboratorDrafts = $state.raw<ProjectCollaboratorDraftState[]>([]);
  socketStatus = $state<RealtimeSocketStatus>('disconnected');
  currentRevision = $state(0);
  syncError = $state<string | null>(null);
  collaboratorsByField = $derived.by(() => buildCollaboratorsByField(this.collaboratorDrafts));

  private readonly onProjectChange?: (change: ProjectChange) => void;
  private readonly onResetRequired?: () => void;
  private readonly onReconnect?: () => void;
  private readonly fetchChanges: NonNullable<ProjectSyncCoordinatorOptions['fetchChanges']>;
  private readonly connection: RealtimeJsonStream<ProjectSyncInboundMessage>;
  private readonly draftDebounceMs: number;
  private readonly catchUpRetryMs: number;
  private readonly bufferedChanges = new Map<number, ProjectChange>();

  private projectId: string | null = null;
  private destroyed = true;
  private generation = 0;
  private highestObservedRevision = 0;
  private catchUpPromise: Promise<void> | null = null;
  private catchUpRetryTimer: ReturnType<typeof setTimeout> | null = null;
  private draftTimer: ReturnType<typeof setTimeout> | null = null;
  private desiredDraftEntries: ProjectDraftEntry[] = [];
  private publishedDraftEntries = new Map<string, ProjectDraftEntry>();

  constructor(options: ProjectSyncCoordinatorOptions = {}) {
    this.onProjectChange = options.onProjectChange;
    this.onResetRequired = options.onResetRequired;
    this.onReconnect = options.onReconnect;
    this.fetchChanges = options.fetchChanges ?? fetchProjectChanges;
    this.draftDebounceMs = options.draftDebounceMs ?? 150;
    this.catchUpRetryMs = options.catchUpRetryMs ?? 1000;
    this.connection = new RealtimeJsonStream<ProjectSyncInboundMessage>({
      url: () => buildProjectCollaborationUrl(this.projectId),
      parseMessage: parseProjectSyncInboundMessage,
      onMessage: (message) => this.handleMessage(message),
      onInvalidMessage: (raw, error) => logInvalidRealtimeMessage(raw, error),
      onOpen: ({ wasReconnect }) => {
        this.flushDraftState();
        if (wasReconnect) this.onReconnect?.();
      },
      onStatusChange: (status) => {
        this.socketStatus = status;
      },
      createConnection: options.createConnection
    });
  }

  connect(projectId: string): void {
    if (!projectId || (this.projectId === projectId && !this.destroyed)) return;

    this.clearLocalDrafts();
    this.connection.disconnect({ clearQueue: true });
    this.generation += 1;
    this.projectId = projectId;
    this.destroyed = false;
    this.currentRevision = 0;
    this.highestObservedRevision = 0;
    this.syncError = null;
    this.bufferedChanges.clear();
    this.catchUpPromise = null;
    this.clearCatchUpRetryTimer();
    this.clearDraftTimer();
    this.desiredDraftEntries = [];
    this.publishedDraftEntries.clear();
    this.connection.connect();
  }

  disconnect(): void {
    this.clearLocalDrafts();
    this.clearDraftTimer();
    this.clearCatchUpRetryTimer();
    this.generation += 1;
    this.destroyed = true;
    this.projectId = null;
    this.onlineUsers = [];
    this.collaboratorDrafts = [];
    this.currentRevision = 0;
    this.highestObservedRevision = 0;
    this.syncError = null;
    this.bufferedChanges.clear();
    this.catchUpPromise = null;
    this.connection.disconnect();
  }

  publishDraftState(entries: ProjectDraftEntry[]): void {
    this.desiredDraftEntries = entries.map(cloneDraftEntry);
    this.scheduleDraftFlush();
  }

  clearDraft(aggregateType: string, identity: { aggregateId?: string; draftId?: string }): void {
    const key = draftKey({
      aggregate_type: aggregateType,
      aggregate_id: identity.aggregateId,
      draft_id: identity.draftId
    });
    this.desiredDraftEntries = this.desiredDraftEntries.filter((entry) => draftKey(entry) !== key);
    this.sendDraftClear({
      aggregate_type: aggregateType,
      aggregate_id: identity.aggregateId,
      draft_id: identity.draftId
    });
    this.publishedDraftEntries.delete(key);
  }

  publishFieldDeviceDraftState(state: SharedFieldDeviceDraftState): void {
    this.publishDraftState(state.devices.flatMap(fieldDeviceDraftEntries));
  }

  /** @deprecated Project changes are authored by the backend after a successful transaction. */
  publishFieldDeviceDelta(): void {}

  buildFieldDeviceEditorsByDevice(
    usersById: Map<string, User>,
    currentUserId?: string
  ): SharedFieldDeviceEditorsByDevice {
    const editors: SharedFieldDeviceEditorsByDevice = {};

    for (const draft of this.collaboratorDrafts) {
      if (draft.user_id === currentUserId) continue;
      const user = usersById.get(draft.user_id);

      for (const entry of draft.entries) {
        if (entry.aggregate_type !== 'field_device' || !entry.aggregate_id) continue;
        const adaptedFields = entry.fields.map((field) => ({
          ...field,
          path: field.path === 'text_individuell' ? 'text_fix' : field.path
        }));
        const fieldValues = Object.fromEntries(
          adaptedFields.map((field) => [field.path, field.value])
        );
        const editor: SharedFieldDeviceEditor = {
          userId: draft.user_id,
          firstName: user?.first_name ?? 'User',
          lastName: user?.last_name ?? draft.user_id.slice(0, 6),
          changedFields: adaptedFields.map((field) => field.path),
          fieldValues,
          updatedAt: draft.updated_at
        };
        editors[entry.aggregate_id] = [...(editors[entry.aggregate_id] ?? []), editor];
      }
    }

    return editors;
  }

  private handleMessage(message: ProjectSyncInboundMessage): void {
    if (!this.projectId || message.project_id !== this.projectId) return;

    switch (message.type) {
      case 'snapshot':
        this.onlineUsers = message.presence;
        this.collaboratorDrafts = message.drafts;
        this.observeWatermark(message.current_revision);
        break;
      case 'presence':
        this.onlineUsers = message.presence;
        break;
      case 'draft_state':
        this.collaboratorDrafts = message.drafts;
        break;
      case 'draft_clear':
        this.removeRemoteDraft(message);
        break;
      case 'project_change':
        this.bufferChange(message);
        break;
      case 'revision':
        this.observeWatermark(message.current_revision);
        break;
    }
  }

  private observeWatermark(revision: number): void {
    this.highestObservedRevision = Math.max(this.highestObservedRevision, revision);
    if (revision <= this.currentRevision) return;
    void this.catchUp();
  }

  private bufferChange(change: ProjectChange): void {
    this.highestObservedRevision = Math.max(this.highestObservedRevision, change.revision);
    if (change.revision <= this.currentRevision) return;
    this.bufferedChanges.set(change.revision, change);
    this.drainBufferedChanges();
    if (this.bufferedChanges.size > 0 && !this.bufferedChanges.has(this.currentRevision + 1)) {
      void this.catchUp();
    }
  }

  private drainBufferedChanges(): void {
    for (;;) {
      const next = this.bufferedChanges.get(this.currentRevision + 1);
      if (!next) return;
      this.bufferedChanges.delete(next.revision);
      this.applyChange(next);
    }
  }

  private applyChange(change: ProjectChange): void {
    if (change.revision <= this.currentRevision) return;
    this.currentRevision = change.revision;
    this.highestObservedRevision = Math.max(this.highestObservedRevision, change.revision);
    this.onProjectChange?.(change);
  }

  private catchUp(): Promise<void> {
    if (!this.projectId || this.destroyed) return Promise.resolve();
    if (this.catchUpPromise) return this.catchUpPromise;

    const projectId = this.projectId;
    const generation = this.generation;
    this.clearCatchUpRetryTimer();
    this.syncError = null;
    const catchUp = this.runCatchUp(projectId, generation)
      .catch((error: unknown) => {
        if (!this.isCurrentGeneration(projectId, generation)) return;
        this.syncError = error instanceof Error ? error.message : 'project sync catch-up failed';
        this.scheduleCatchUpRetry(projectId, generation);
      })
      .finally(() => {
        if (this.catchUpPromise === catchUp) {
          this.catchUpPromise = null;
          if (!this.isCurrentGeneration(projectId, generation)) return;
          this.drainBufferedChanges();
          if (this.currentRevision < this.highestObservedRevision && !this.catchUpRetryTimer) {
            void this.catchUp();
          }
        }
      });
    this.catchUpPromise = catchUp;
    return catchUp;
  }

  private async runCatchUp(projectId: string, generation: number): Promise<void> {
    while (this.currentRevision < this.highestObservedRevision) {
      const previousRevision = this.currentRevision;
      const response = await this.fetchChanges(projectId, this.currentRevision);
      if (!this.isCurrentGeneration(projectId, generation)) return;

      if (response.reset_required) {
        this.bufferedChanges.clear();
        this.currentRevision = response.current_revision;
        this.highestObservedRevision = Math.max(
          this.highestObservedRevision,
          response.current_revision
        );
        this.onResetRequired?.();
        return;
      }

      for (const event of [...response.events].sort((a, b) => a.revision - b.revision)) {
        if (event.revision > this.currentRevision) {
          this.bufferedChanges.set(event.revision, event);
        }
      }
      this.drainBufferedChanges();

      if (
        this.currentRevision === previousRevision &&
        this.currentRevision < this.highestObservedRevision
      ) {
        throw new Error('project sync catch-up made no revision progress');
      }

      if (response.has_more) continue;
    }
  }

  private isCurrentGeneration(projectId: string, generation: number): boolean {
    return !this.destroyed && this.projectId === projectId && this.generation === generation;
  }

  private scheduleCatchUpRetry(projectId: string, generation: number): void {
    if (this.catchUpRetryTimer || !this.isCurrentGeneration(projectId, generation)) return;
    this.catchUpRetryTimer = setTimeout(() => {
      this.catchUpRetryTimer = null;
      if (this.isCurrentGeneration(projectId, generation)) void this.catchUp();
    }, this.catchUpRetryMs);
  }

  private clearCatchUpRetryTimer(): void {
    if (!this.catchUpRetryTimer) return;
    clearTimeout(this.catchUpRetryTimer);
    this.catchUpRetryTimer = null;
  }

  private removeRemoteDraft(message: ProjectDraftClearMessage): void {
    const key = draftKey(message);
    this.collaboratorDrafts = this.collaboratorDrafts
      .map((draft) =>
        draft.user_id === message.user_id
          ? { ...draft, entries: draft.entries.filter((entry) => draftKey(entry) !== key) }
          : draft
      )
      .filter((draft) => draft.entries.length > 0);
  }

  private scheduleDraftFlush(): void {
    this.clearDraftTimer();
    this.draftTimer = setTimeout(() => {
      this.draftTimer = null;
      this.flushDraftState();
    }, this.draftDebounceMs);
  }

  private flushDraftState(): void {
    if (this.destroyed) return;

    const next = new Map(this.desiredDraftEntries.map((entry) => [draftKey(entry), entry]));
    for (const [key, entry] of this.publishedDraftEntries) {
      if (!next.has(key)) this.sendDraftClear(entry);
    }

    if (this.desiredDraftEntries.length > 0) {
      this.connection.send(
        { type: 'draft_state', entries: this.desiredDraftEntries.map(cloneDraftEntry) },
        { queueWhenClosed: true }
      );
    }
    this.publishedDraftEntries = next;
  }

  private sendDraftClear(entry: {
    aggregate_type: string;
    aggregate_id?: string;
    draft_id?: string;
  }): void {
    this.connection.send(
      {
        type: 'draft_clear',
        aggregate_type: entry.aggregate_type,
        ...(entry.aggregate_id ? { aggregate_id: entry.aggregate_id } : {}),
        ...(entry.draft_id ? { draft_id: entry.draft_id } : {})
      },
      { queueWhenClosed: true }
    );
  }

  private clearLocalDrafts(): void {
    const entries = new Map<string, ProjectDraftEntry>();
    for (const entry of this.publishedDraftEntries.values()) {
      entries.set(draftKey(entry), entry);
    }
    for (const entry of this.desiredDraftEntries) {
      entries.set(draftKey(entry), entry);
    }
    for (const entry of entries.values()) {
      this.sendDraftClear(entry);
    }
    this.desiredDraftEntries = [];
    this.publishedDraftEntries.clear();
  }

  private clearDraftTimer(): void {
    if (!this.draftTimer) return;
    clearTimeout(this.draftTimer);
    this.draftTimer = null;
  }
}

const [getProjectSyncCoordinator, setProjectSyncCoordinator] =
  createContext<ProjectSyncCoordinator>();

export function provideProjectSyncCoordinator(
  options: ProjectSyncCoordinatorOptions = {}
): ProjectSyncCoordinator {
  const coordinator = new ProjectSyncCoordinator(options);
  setProjectSyncCoordinator(coordinator);
  return coordinator;
}

export function useProjectSyncCoordinator(): ProjectSyncCoordinator {
  return getProjectSyncCoordinator();
}

/** @deprecated Use ProjectSyncCoordinator. */
export const ProjectCollaborationState = ProjectSyncCoordinator;

export function parseProjectSyncInboundMessage(message: unknown): ProjectSyncInboundMessage {
  return projectSyncInboundSchema.parse(
    normalizeProjectSyncInboundMessage(message)
  ) as ProjectSyncInboundMessage;
}

/** @deprecated Use parseProjectSyncInboundMessage. */
export const parseProjectCollaborationInboundMessage = parseProjectSyncInboundMessage;

async function fetchProjectChanges(
  projectId: string,
  afterRevision: number
): Promise<ProjectChangesResponse> {
  const { data } = await apiClient.GET('/api/v1/projects/{id}/changes', {
    params: {
      path: { id: projectId },
      query: { after_revision: afterRevision, limit: 500 }
    }
  });
  const parsed = projectChangesResponseSchema.parse(data);
  const events = parsed.events ?? parsed.items ?? [];
  return {
    project_id: parsed.project_id,
    current_revision:
      parsed.current_revision ?? parsed.next_revision ?? events.at(-1)?.revision ?? afterRevision,
    events,
    has_more: parsed.has_more ?? false,
    reset_required: parsed.reset_required ?? false
  };
}

function buildProjectCollaborationUrl(projectId: string | null): string | null {
  if (!projectId) return null;
  return buildSameOriginWebSocketUrl(`/api/v1/projects/${projectId}/collaboration`);
}

function cloneDraftEntry(entry: ProjectDraftEntry): ProjectDraftEntry {
  return { ...entry, fields: entry.fields.map((field) => ({ ...field })) };
}

function normalizeProjectSyncInboundMessage(message: unknown): unknown {
  if (!message || typeof message !== 'object' || Array.isArray(message)) return message;
  const value = message as Record<string, unknown>;
  if (value.type === 'snapshot' && value.drafts === undefined && value.draft_states !== undefined) {
    return { ...value, drafts: value.draft_states };
  }
  if (value.type === 'draft_states') {
    return { ...value, type: 'draft_state', drafts: value.draft_states };
  }
  return message;
}

function fieldDeviceDraftEntries(
  device: SharedFieldDeviceDraftState['devices'][number]
): ProjectDraftEntry[] {
  const baseFields: ProjectDraftField[] = [];
  const specificationFields: ProjectDraftField[] = [];
  const bacnetFields = new Map<string, ProjectDraftField[]>();

  for (const originalPath of device.changed_fields) {
    const value = device.field_values?.[originalPath];
    if (originalPath.startsWith('specification.')) {
      if (device.specification_id) {
        specificationFields.push({ path: originalPath.slice('specification.'.length), value });
      }
      continue;
    }

    const bacnetMatch = /^bacnet_objects\.([0-9a-f-]{36})\.(.+)$/i.exec(originalPath);
    if (bacnetMatch) {
      const [, objectId, path] = bacnetMatch;
      bacnetFields.set(objectId, [...(bacnetFields.get(objectId) ?? []), { path, value }]);
      continue;
    }

    baseFields.push({
      path: originalPath === 'text_fix' ? 'text_individuell' : originalPath,
      value
    });
  }

  const entries: ProjectDraftEntry[] = [];
  if (baseFields.length > 0) {
    entries.push(updateDraftEntry('field_device', device.device_id, baseFields));
  }
  if (device.specification_id && specificationFields.length > 0) {
    entries.push(updateDraftEntry('specification', device.specification_id, specificationFields));
  }
  for (const [objectId, fields] of bacnetFields) {
    entries.push(updateDraftEntry('bacnet_object', objectId, fields));
  }
  return entries;
}

function updateDraftEntry(
  aggregateType: string,
  aggregateId: string,
  fields: ProjectDraftField[]
): ProjectDraftEntry {
  return {
    aggregate_type: aggregateType,
    aggregate_id: aggregateId,
    action: 'update',
    base_version: 0,
    fields
  };
}

function draftKey(entry: {
  aggregate_type: string;
  aggregate_id?: string;
  draft_id?: string;
}): string {
  return `${entry.aggregate_type}:${entry.aggregate_id ?? `draft:${entry.draft_id ?? ''}`}`;
}

function buildCollaboratorsByField(
  drafts: ProjectCollaboratorDraftState[]
): ProjectCollaboratorsByField {
  const result: ProjectCollaboratorsByField = {};
  for (const draft of drafts) {
    for (const entry of draft.entries) {
      const aggregateKey = draftKey(entry);
      const fields = (result[aggregateKey] ??= {});
      for (const field of entry.fields) {
        fields[field.path] = [
          ...(fields[field.path] ?? []),
          { userId: draft.user_id, value: field.value, updatedAt: draft.updated_at }
        ];
      }
    }
  }
  return result;
}

function logInvalidRealtimeMessage(raw: string, error: unknown): void {
  console.warn('Ignored invalid project sync WebSocket message', {
    reason: error instanceof Error ? error.message : 'invalid realtime message',
    bytes: raw.length
  });
}
