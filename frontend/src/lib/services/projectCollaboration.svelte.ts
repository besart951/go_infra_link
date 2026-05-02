import type { ControlCabinet, FieldDevice, SPSController } from '$lib/domain/facility/index.js';
import type { User } from '$lib/domain/user/index.js';
import {
  ReconnectingWebSocket,
  type RealtimeSocketStatus
} from '$lib/infrastructure/realtime/reconnectingWebSocket.js';

export interface ProjectCollaboratorPresence {
  user_id: string;
  connected_at: string;
  last_seen_at: string;
}

export interface ProjectFieldDeviceEditState {
  user_id: string;
  devices: Array<{
    device_id: string;
    changed_fields: string[];
    field_values?: Record<string, unknown>;
  }>;
  updated_at: string;
}

export interface ProjectCollaborationRefreshRequest {
  type: 'refresh_request';
  project_id: string;
  scope: string;
  actor_id?: string;
  entity_ids?: string[];
  device_ids?: string[];
  at: string;
}

export interface ProjectCollaborationEntityDeltaMessage {
  type: 'entity_delta';
  project_id: string;
  scope: string;
  actor_id?: string;
  control_cabinets?: ControlCabinet[];
  sps_controllers?: SPSController[];
  field_devices?: FieldDevice[];
  at: string;
}

interface ProjectCollaborationSnapshotMessage {
  type: 'snapshot';
  presence: ProjectCollaboratorPresence[];
  edit_states: ProjectFieldDeviceEditState[];
}

interface ProjectCollaborationPresenceMessage {
  type: 'presence';
  presence: ProjectCollaboratorPresence[];
}

interface ProjectCollaborationEditStatesMessage {
  type: 'edit_states';
  edit_states: ProjectFieldDeviceEditState[];
}

type ProjectCollaborationInboundMessage =
  | ProjectCollaborationSnapshotMessage
  | ProjectCollaborationPresenceMessage
  | ProjectCollaborationEditStatesMessage
  | ProjectCollaborationEntityDeltaMessage
  | ProjectCollaborationRefreshRequest;

export interface SharedFieldDeviceDraftState {
  devices: Array<{
    device_id: string;
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

interface ProjectCollaborationStateOptions {
  onRefreshRequest?: (message: ProjectCollaborationRefreshRequest) => void;
  onEntityDelta?: (message: ProjectCollaborationEntityDeltaMessage) => void;
  onReconnect?: () => void;
}

export class ProjectCollaborationState {
  onlineUsers = $state<ProjectCollaboratorPresence[]>([]);
  fieldDeviceEditStates = $state<ProjectFieldDeviceEditState[]>([]);
  socketStatus = $state<RealtimeSocketStatus>('disconnected');

  private readonly onRefreshRequest?: (message: ProjectCollaborationRefreshRequest) => void;
  private readonly onEntityDelta?: (message: ProjectCollaborationEntityDeltaMessage) => void;
  private readonly onReconnect?: () => void;
  private readonly connection: ReconnectingWebSocket;

  private projectId: string | null = null;
  private destroyed = false;
  private desiredEditState: SharedFieldDeviceDraftState = {
    devices: []
  };

  constructor(options: ProjectCollaborationStateOptions = {}) {
    this.onRefreshRequest = options.onRefreshRequest;
    this.onEntityDelta = options.onEntityDelta;
    this.onReconnect = options.onReconnect;
    this.connection = new ReconnectingWebSocket({
      url: () => buildProjectCollaborationUrl(this.projectId),
      onMessage: (raw) => this.handleMessage(raw),
      onOpen: ({ wasReconnect }) => {
        this.publishFieldDeviceDraftState(this.desiredEditState);
        if (wasReconnect) {
          this.onReconnect?.();
        }
      },
      onStatusChange: (status) => {
        this.socketStatus = status;
      }
    });
  }

  connect(projectId: string): void {
    if (!projectId) return;

    if (this.projectId === projectId && !this.destroyed) {
      return;
    }

    this.projectId = projectId;
    this.destroyed = false;
    this.connection.disconnect({ clearQueue: true });
    this.connection.connect();
  }

  disconnect(): void {
    this.destroyed = true;
    this.projectId = null;
    this.onlineUsers = [];
    this.fieldDeviceEditStates = [];
    this.connection.disconnect();
  }

  publishFieldDeviceDraftState(state: SharedFieldDeviceDraftState): void {
    this.desiredEditState = {
      devices: state.devices.map((device) => ({
        device_id: device.device_id,
        changed_fields: [...device.changed_fields],
        field_values: device.field_values ? { ...device.field_values } : undefined
      }))
    };

    this.send({
      type: 'edit_state',
      devices: state.devices
    });
  }

  requestFieldDeviceRefresh(deviceIds: string[]): void {
    this.send(
      {
        type: 'refresh_request',
        scope: 'field_device',
        device_ids: deviceIds
      },
      { queueWhenClosed: true }
    );
  }

  publishFieldDeviceDelta(fieldDevices: FieldDevice[]): void {
    if (fieldDevices.length === 0) {
      return;
    }

    this.send(
      {
        type: 'entity_delta',
        scope: 'field_device',
        field_devices: fieldDevices
      },
      { queueWhenClosed: true }
    );
  }

  buildFieldDeviceEditorsByDevice(
    usersById: Map<string, User>,
    currentUserId?: string
  ): SharedFieldDeviceEditorsByDevice {
    const editors: SharedFieldDeviceEditorsByDevice = {};

    for (const state of this.fieldDeviceEditStates) {
      if (!state.devices?.length) continue;
      if (currentUserId && state.user_id === currentUserId) continue;

      const user = usersById.get(state.user_id);

      for (const device of state.devices) {
        const editor: SharedFieldDeviceEditor = {
          userId: state.user_id,
          firstName: user?.first_name ?? 'User',
          lastName: user?.last_name ?? state.user_id.slice(0, 6),
          changedFields: device.changed_fields || [],
          fieldValues: device.field_values,
          updatedAt: state.updated_at
        };

        const deviceId = device.device_id;
        editors[deviceId] = [...(editors[deviceId] ?? []), editor];
      }
    }

    return editors;
  }

  private handleMessage(raw: string): void {
    let message: ProjectCollaborationInboundMessage;

    try {
      message = JSON.parse(raw) as ProjectCollaborationInboundMessage;
    } catch {
      return;
    }

    switch (message.type) {
      case 'snapshot':
        this.onlineUsers = message.presence ?? [];
        this.fieldDeviceEditStates = message.edit_states ?? [];
        break;
      case 'presence':
        this.onlineUsers = message.presence ?? [];
        break;
      case 'edit_states':
        this.fieldDeviceEditStates = message.edit_states ?? [];
        break;
      case 'entity_delta':
        this.onEntityDelta?.(message);
        break;
      case 'refresh_request':
        this.onRefreshRequest?.(message);
        break;
    }
  }

  private send(
    payload: Record<string, unknown>,
    options: { queueWhenClosed?: boolean } = {}
  ): void {
    if (this.destroyed) return;
    this.connection.send(payload, options);
  }
}

function buildProjectCollaborationUrl(projectId: string | null): string | null {
  if (!projectId || typeof window === 'undefined') return null;

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocol}//${window.location.host}/api/v1/projects/${projectId}/collaboration`;
}
