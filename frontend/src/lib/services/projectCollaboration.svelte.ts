import type { ControlCabinet, FieldDevice, SPSController } from '$lib/domain/facility/index.js';
import type { User } from '$lib/domain/user/index.js';

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

type SocketStatus = 'disconnected' | 'connecting' | 'connected';

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
  socketStatus = $state<SocketStatus>('disconnected');

  private readonly onRefreshRequest?: (message: ProjectCollaborationRefreshRequest) => void;
  private readonly onEntityDelta?: (message: ProjectCollaborationEntityDeltaMessage) => void;
  private readonly onReconnect?: () => void;

  private projectId: string | null = null;
  private socket: WebSocket | null = null;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectDelayMs = 2000;
  private hasConnectedOnce = false;
  private pendingMessages: Array<Record<string, unknown>> = [];
  private destroyed = false;
  private desiredEditState: SharedFieldDeviceDraftState = {
    devices: []
  };

  constructor(options: ProjectCollaborationStateOptions = {}) {
    this.onRefreshRequest = options.onRefreshRequest;
    this.onEntityDelta = options.onEntityDelta;
    this.onReconnect = options.onReconnect;
  }

  connect(projectId: string): void {
    if (!projectId) return;

    if (this.projectId === projectId && this.socket) {
      return;
    }

    this.projectId = projectId;
    this.destroyed = false;
    this.clearReconnectTimer();
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
    this.openSocket();
  }

  disconnect(): void {
    this.destroyed = true;
    this.projectId = null;
    this.clearReconnectTimer();
    this.socketStatus = 'disconnected';
    this.onlineUsers = [];
    this.fieldDeviceEditStates = [];
    this.hasConnectedOnce = false;
    this.pendingMessages = [];

    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
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

  private openSocket(): void {
    if (!this.projectId || this.destroyed) return;

    this.socketStatus = 'connecting';
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const socket = new WebSocket(
      `${protocol}//${window.location.host}/api/v1/projects/${this.projectId}/collaboration`
    );
    this.socket = socket;

    socket.addEventListener('open', () => {
      if (this.socket !== socket) return;
      const wasReconnect = this.hasConnectedOnce;
      this.hasConnectedOnce = true;
      this.socketStatus = 'connected';
      this.reconnectDelayMs = 2000;
      this.publishFieldDeviceDraftState(this.desiredEditState);
      this.flushPendingMessages();
      if (wasReconnect) {
        this.onReconnect?.();
      }
    });

    socket.addEventListener('message', (event) => {
      if (this.socket !== socket) return;
      this.handleMessage(event.data);
    });

    socket.addEventListener('close', () => {
      if (this.socket === socket) {
        this.socket = null;
      }
      if (this.destroyed) return;

      this.socketStatus = 'disconnected';
      this.scheduleReconnect();
    });

    socket.addEventListener('error', () => {
      socket.close();
    });
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
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      if (options.queueWhenClosed) {
        this.queuePendingMessage(payload);
      }
      return;
    }

    this.socket.send(JSON.stringify(payload));
  }

  private queuePendingMessage(payload: Record<string, unknown>): void {
    this.pendingMessages = [...this.pendingMessages, payload].slice(-50);
  }

  private flushPendingMessages(): void {
    if (this.pendingMessages.length === 0) return;

    const messages = this.pendingMessages;
    this.pendingMessages = [];
    for (const message of messages) {
      this.send(message);
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer || !this.projectId) return;

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.openSocket();
    }, this.reconnectDelayMs);
    this.reconnectDelayMs = Math.min(this.reconnectDelayMs * 2, 10000);
  }

  private clearReconnectTimer(): void {
    if (!this.reconnectTimer) return;

    clearTimeout(this.reconnectTimer);
    this.reconnectTimer = null;
  }
}
