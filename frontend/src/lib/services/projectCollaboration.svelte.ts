import type { ControlCabinet, FieldDevice, SPSController } from '$lib/domain/facility/index.js';
import type { User } from '$lib/domain/user/index.js';
import type { RealtimeSocketStatus } from '$lib/infrastructure/realtime/reconnectingWebSocket.js';
import {
  buildSameOriginWebSocketUrl,
  RealtimeJsonStream
} from '$lib/infrastructure/realtime/realtimeJsonStream.js';
import { z } from 'zod';

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

const PROJECT_COLLABORATION_FIELD_VALUE_MAX = 64;
const PROJECT_COLLABORATION_CHANGED_FIELDS_MAX = 64;
const PROJECT_COLLABORATION_DEVICES_MAX = 100;
const PROJECT_COLLABORATION_ENTITIES_MAX = 100;
const PROJECT_COLLABORATION_FIELD_NAME_MAX = 180;
const PROJECT_COLLABORATION_STRING_VALUE_MAX = 2048;

const uuidSchema = z.string().uuid();
const dateTimeSchema = z.string().min(1).max(80);
const nullableStringSchema = z.string().nullable().optional();
const scalarFieldValueSchema = z.union([
  z.string().max(PROJECT_COLLABORATION_STRING_VALUE_MAX),
  z.number(),
  z.boolean(),
  z.null()
]);

const allowedBaseFields = new Set([
  'bmk',
  'description',
  'text_fix',
  'apparat_nr',
  'apparat_id',
  'system_part_id',
  'sps_controller_system_type_id',
  'specification_id'
]);

const allowedSpecificationFields = new Set([
  'specification_supplier',
  'specification_brand',
  'specification_type',
  'additional_info_motor_valve',
  'additional_info_size',
  'additional_information_installation_location',
  'electrical_connection_ph',
  'electrical_connection_acdc',
  'electrical_connection_amperage',
  'electrical_connection_power',
  'electrical_connection_rotation'
]);

const allowedBacnetFields = new Set([
  'text_fix',
  'description',
  'gms_visible',
  'optional',
  'text_individual',
  'software_type',
  'software_number',
  'hardware_type',
  'hardware_quantity',
  'software_reference_id',
  'state_text_id',
  'notification_class_id',
  'alarm_type_id'
]);

const projectCollaborationFieldNameSchema = z
  .string()
  .transform((value) => value.trim())
  .refine((value) => isProjectCollaborationFieldName(value), 'invalid collaboration field');

const fieldValuesSchema = z
  .record(z.string(), scalarFieldValueSchema)
  .superRefine((values, ctx) => {
    const keys = Object.keys(values);
    if (keys.length > PROJECT_COLLABORATION_FIELD_VALUE_MAX) {
      ctx.addIssue({
        code: 'custom',
        message: 'too many field_values'
      });
      return;
    }

    for (const key of keys) {
      if (!isProjectCollaborationFieldName(key.trim())) {
        ctx.addIssue({
          code: 'custom',
          message: 'invalid field_values key',
          path: [key]
        });
      }
    }
  })
  .transform((values) =>
    Object.fromEntries(Object.entries(values).map(([key, value]) => [key.trim(), value]))
  );

const projectFieldDeviceEditStateSchema = z
  .object({
    device_id: uuidSchema,
    changed_fields: z
      .array(projectCollaborationFieldNameSchema)
      .min(1)
      .max(PROJECT_COLLABORATION_CHANGED_FIELDS_MAX),
    field_values: fieldValuesSchema.optional()
  })
  .strict()
  .superRefine((device, ctx) => {
    const changedFields = new Set(device.changed_fields);
    for (const key of Object.keys(device.field_values ?? {})) {
      if (!changedFields.has(key.trim())) {
        ctx.addIssue({
          code: 'custom',
          message: 'field_values key must be listed in changed_fields',
          path: ['field_values', key]
        });
      }
    }
  });

const presenceSchema = z
  .object({
    user_id: uuidSchema,
    connected_at: dateTimeSchema,
    last_seen_at: dateTimeSchema
  })
  .strict();

const editStateSchema = z
  .object({
    user_id: uuidSchema,
    devices: z.array(projectFieldDeviceEditStateSchema).max(PROJECT_COLLABORATION_DEVICES_MAX),
    updated_at: dateTimeSchema
  })
  .strict();

const controlCabinetSchema = z
  .object({
    id: uuidSchema,
    building_id: uuidSchema,
    control_cabinet_nr: nullableStringSchema,
    created_at: dateTimeSchema,
    updated_at: dateTimeSchema
  })
  .strict();

const spsControllerSchema = z
  .object({
    id: uuidSchema,
    control_cabinet_id: uuidSchema,
    ga_device: nullableStringSchema,
    device_name: z.string(),
    device_description: nullableStringSchema,
    device_location: nullableStringSchema,
    ip_address: nullableStringSchema,
    subnet: nullableStringSchema,
    gateway: nullableStringSchema,
    vlan: nullableStringSchema,
    created_at: dateTimeSchema,
    updated_at: dateTimeSchema
  })
  .strict();

const bacnetObjectSchema = z
  .object({
    id: uuidSchema,
    text_fix: z.string(),
    description: nullableStringSchema,
    gms_visible: z.boolean(),
    optional: z.boolean(),
    text_individual: nullableStringSchema,
    software_type: z.string(),
    software_number: z.number(),
    hardware_type: z.string(),
    hardware_quantity: z.number(),
    field_device_id: uuidSchema.optional(),
    software_reference_id: uuidSchema.nullable().optional(),
    state_text_id: uuidSchema.nullable().optional(),
    notification_class_id: uuidSchema.nullable().optional(),
    alarm_type_id: uuidSchema.nullable().optional(),
    created_at: dateTimeSchema.optional(),
    updated_at: dateTimeSchema.optional()
  })
  .strict();

const fieldDeviceSchema = z
  .object({
    id: uuidSchema,
    bmk: nullableStringSchema,
    description: nullableStringSchema,
    text_fix: nullableStringSchema,
    apparat_nr: z.union([z.string(), z.number()]).transform(String),
    sps_controller_system_type_id: uuidSchema,
    system_part_id: uuidSchema.nullable().optional(),
    specification_id: uuidSchema.nullable().optional(),
    apparat_id: uuidSchema,
    created_at: dateTimeSchema,
    updated_at: dateTimeSchema,
    sps_controller_system_type: z.unknown().optional(),
    apparat: z.unknown().optional(),
    system_part: z.unknown().optional(),
    specification: z.unknown().optional(),
    bacnet_objects: z.array(bacnetObjectSchema).max(250).optional()
  })
  .strict();

const projectCollaborationInboundMessageSchema = z.discriminatedUnion('type', [
  z
    .object({
      type: z.literal('snapshot'),
      project_id: uuidSchema.optional(),
      presence: z.array(presenceSchema).max(250),
      edit_states: z.array(editStateSchema).max(250),
      at: dateTimeSchema.optional()
    })
    .strict(),
  z
    .object({
      type: z.literal('presence'),
      project_id: uuidSchema.optional(),
      presence: z.array(presenceSchema).max(250),
      at: dateTimeSchema.optional()
    })
    .strict(),
  z
    .object({
      type: z.literal('edit_states'),
      project_id: uuidSchema.optional(),
      edit_states: z.array(editStateSchema).max(250),
      at: dateTimeSchema.optional()
    })
    .strict(),
  z
    .object({
      type: z.literal('entity_delta'),
      project_id: uuidSchema,
      scope: z.enum(['control_cabinet', 'sps_controller', 'field_device']),
      actor_id: uuidSchema.optional(),
      control_cabinets: z.array(controlCabinetSchema).max(100).optional(),
      sps_controllers: z.array(spsControllerSchema).max(100).optional(),
      field_devices: z.array(fieldDeviceSchema).max(100).optional(),
      at: dateTimeSchema
    })
    .strict(),
  z
    .object({
      type: z.literal('refresh_request'),
      project_id: uuidSchema,
      scope: z.enum([
        'project',
        'project_users',
        'control_cabinet',
        'sps_controller',
        'field_device'
      ]),
      actor_id: uuidSchema.optional(),
      entity_ids: z.array(uuidSchema).max(PROJECT_COLLABORATION_ENTITIES_MAX).optional(),
      device_ids: z.array(uuidSchema).max(PROJECT_COLLABORATION_ENTITIES_MAX).optional(),
      at: dateTimeSchema
    })
    .strict()
]);

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
  private readonly connection: RealtimeJsonStream<ProjectCollaborationInboundMessage>;

  private projectId: string | null = null;
  private destroyed = false;
  private desiredEditState: SharedFieldDeviceDraftState = {
    devices: []
  };

  constructor(options: ProjectCollaborationStateOptions = {}) {
    this.onRefreshRequest = options.onRefreshRequest;
    this.onEntityDelta = options.onEntityDelta;
    this.onReconnect = options.onReconnect;
    this.connection = new RealtimeJsonStream<ProjectCollaborationInboundMessage>({
      url: () => buildProjectCollaborationUrl(this.projectId),
      parseMessage: parseProjectCollaborationInboundMessage,
      onMessage: (message) => this.handleMessage(message),
      onInvalidMessage: (raw, error) =>
        logInvalidRealtimeMessage('project collaboration', raw, error),
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

  private handleMessage(message: ProjectCollaborationInboundMessage): void {
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
  if (!projectId) return null;
  return buildSameOriginWebSocketUrl(`/api/v1/projects/${projectId}/collaboration`);
}

export function parseProjectCollaborationInboundMessage(
  message: unknown
): ProjectCollaborationInboundMessage {
  return projectCollaborationInboundMessageSchema.parse(
    message
  ) as ProjectCollaborationInboundMessage;
}

function isProjectCollaborationFieldName(value: string): boolean {
  if (value.length === 0 || value.length > PROJECT_COLLABORATION_FIELD_NAME_MAX) return false;
  if (allowedBaseFields.has(value)) return true;

  if (value.startsWith('specification.')) {
    const [, field] = value.split('.');
    return !!field && allowedSpecificationFields.has(field);
  }

  const parts = value.split('.');
  return (
    parts.length === 3 &&
    parts[0] === 'bacnet_objects' &&
    uuidSchema.safeParse(parts[1]).success &&
    allowedBacnetFields.has(parts[2])
  );
}

function logInvalidRealtimeMessage(streamName: string, raw: string, error: unknown): void {
  console.warn(`Ignored invalid ${streamName} WebSocket message`, {
    reason: error instanceof Error ? error.message : 'invalid realtime message',
    bytes: raw.length
  });
}
