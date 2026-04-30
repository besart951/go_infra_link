import { t as translate } from '$lib/i18n/index.js';
import {
  TableViewState,
  type TableGroupingDefinition
} from '$lib/state/table/TableViewState.svelte.js';
import type {
  Building,
  ControlCabinet,
  FieldDevice,
  SPSController
} from '$lib/domain/facility/index.js';

export type FieldDeviceGroupKey =
  | 'building'
  | 'controlCabinet'
  | 'spsController'
  | 'spsControllerSystemType';

export interface FieldDeviceGroupOption {
  key: FieldDeviceGroupKey;
  labelKey: string;
}

export interface FieldDeviceGroupingLookups {
  getSPSController(id: string): SPSController | undefined;
  getControlCabinet(id: string): ControlCabinet | undefined;
  getBuilding(id: string): Building | undefined;
}

export const FIELD_DEVICE_GROUP_OPTIONS: FieldDeviceGroupOption[] = [
  { key: 'building', labelKey: 'field_device.view.groups.building' },
  { key: 'controlCabinet', labelKey: 'field_device.view.groups.control_cabinet' },
  { key: 'spsController', labelKey: 'field_device.view.groups.sps_controller' },
  {
    key: 'spsControllerSystemType',
    labelKey: 'field_device.view.groups.sps_controller_system_type'
  }
];

export function formatFieldDeviceSPSControllerSystemType(fieldDevice: FieldDevice): string {
  const systemType = fieldDevice.sps_controller_system_type;
  if (!systemType) return '-';

  const deviceName = systemType.sps_controller_name ?? '';
  const number =
    systemType.number === null || systemType.number === undefined
      ? ''
      : String(systemType.number).padStart(4, '0');
  const documentName = systemType.document_name ?? '';

  let systemTypePart = '';
  if (number && documentName) {
    systemTypePart = `${number} - ${documentName}`;
  } else if (number) {
    systemTypePart = number;
  } else if (documentName) {
    systemTypePart = documentName;
  }

  if (deviceName && systemTypePart) return `${deviceName}_${systemTypePart}`;
  if (deviceName) return deviceName;
  if (systemTypePart) return systemTypePart;
  return '-';
}

export class FieldDeviceGroupingResolver {
  constructor(private readonly lookups: FieldDeviceGroupingLookups) {}

  createDefinitions(): TableGroupingDefinition<FieldDevice, FieldDeviceGroupKey>[] {
    return [
      {
        key: 'building',
        labelKey: 'field_device.view.groups.building',
        resolveId: (device) => this.getBuildingId(device),
        resolveLabel: (_device, groupId) =>
          this.lookups.getBuilding(groupId)?.iws_code ??
          this.fallbackGroupLabel(groupId, 'field_device.view.group_missing.building')
      },
      {
        key: 'controlCabinet',
        labelKey: 'field_device.view.groups.control_cabinet',
        resolveId: (device) => this.getControlCabinetId(device),
        resolveLabel: (_device, groupId) =>
          this.lookups.getControlCabinet(groupId)?.control_cabinet_nr ??
          this.fallbackGroupLabel(groupId, 'field_device.view.group_missing.control_cabinet')
      },
      {
        key: 'spsController',
        labelKey: 'field_device.view.groups.sps_controller',
        resolveId: (device) => device.sps_controller_system_type?.sps_controller_id,
        resolveLabel: (device, groupId) =>
          this.lookups.getSPSController(groupId)?.device_name ??
          device.sps_controller_system_type?.sps_controller_name ??
          this.fallbackGroupLabel(groupId, 'field_device.view.group_missing.sps_controller')
      },
      {
        key: 'spsControllerSystemType',
        labelKey: 'field_device.view.groups.sps_controller_system_type',
        resolveId: (device) => device.sps_controller_system_type_id,
        resolveLabel: (device, groupId) => {
          const label = formatFieldDeviceSPSControllerSystemType(device);
          return label === '-' ? groupId : label;
        }
      }
    ];
  }

  private getControlCabinetId(device: FieldDevice): string | undefined {
    const controllerId = device.sps_controller_system_type?.sps_controller_id;
    if (!controllerId) return undefined;
    return this.lookups.getSPSController(controllerId)?.control_cabinet_id;
  }

  private getBuildingId(device: FieldDevice): string | undefined {
    const cabinetId = this.getControlCabinetId(device);
    if (!cabinetId) return undefined;
    return this.lookups.getControlCabinet(cabinetId)?.building_id;
  }

  private fallbackGroupLabel(groupId: string, missingLabelKey: string): string {
    return groupId === 'unassigned' ? translate(missingLabelKey) : groupId;
  }
}

export class FieldDeviceTableViewState extends TableViewState<FieldDevice, FieldDeviceGroupKey> {
  readonly groupOptions = FIELD_DEVICE_GROUP_OPTIONS;

  constructor(lookups: FieldDeviceGroupingLookups) {
    const resolver = new FieldDeviceGroupingResolver(lookups);
    super({
      defaultDensity: 'medium',
      groupingDefinitions: resolver.createDefinitions(),
      storageKey: 'table-view:field-device'
    });
  }
}
