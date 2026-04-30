import type { FieldDevice, SPSController } from '$lib/domain/facility/index.js';

interface VisibleRowContext {
  items: FieldDevice[];
  searchText?: string;
  orderBy?: string;
  order?: string;
  hasActiveFilters: boolean;
}

export type DeviceRefreshPlan =
  | { action: 'reload' }
  | { action: 'refresh'; ids: string[] }
  | { action: 'none' };

export type DeviceDeltaPlan =
  | { action: 'reload' }
  | { action: 'replace'; devices: FieldDevice[] }
  | { action: 'none' };

function hasQueryState(context: VisibleRowContext): boolean {
  return Boolean(
    context.searchText || context.orderBy || context.order || context.hasActiveFilters
  );
}

export function planVisibleDeviceRefresh(
  context: VisibleRowContext,
  deviceIds: string[]
): DeviceRefreshPlan {
  const uniqueDeviceIds = [...new Set(deviceIds.filter(Boolean))];

  if (uniqueDeviceIds.length === 0) {
    return { action: 'reload' };
  }

  if (hasQueryState(context)) {
    return { action: 'reload' };
  }

  const visibleIds = new Set(context.items.map((item) => item.id));
  if (uniqueDeviceIds.some((id) => !visibleIds.has(id))) {
    return { action: 'reload' };
  }

  return { action: 'refresh', ids: uniqueDeviceIds };
}

export function planVisibleDeviceDelta(
  context: VisibleRowContext,
  fieldDevices: FieldDevice[]
): DeviceDeltaPlan {
  const updatedDevices = [...new Map(fieldDevices.map((item) => [item.id, item])).values()];

  if (updatedDevices.length === 0) {
    return { action: 'none' };
  }

  if (hasQueryState(context)) {
    return { action: 'reload' };
  }

  const visibleIds = new Set(context.items.map((item) => item.id));
  const visibleDevices = updatedDevices.filter((item) => visibleIds.has(item.id));
  const hasNewDevices = updatedDevices.some((item) => !visibleIds.has(item.id));

  if (hasNewDevices) {
    return { action: 'reload' };
  }

  return visibleDevices.length > 0
    ? { action: 'replace', devices: visibleDevices }
    : { action: 'none' };
}

export function planSPSControllerDeviceRefresh(
  context: VisibleRowContext,
  spsControllerIds: string[]
): DeviceRefreshPlan {
  const uniqueControllerIds = [...new Set(spsControllerIds.filter(Boolean))];

  if (uniqueControllerIds.length === 0) {
    return { action: 'reload' };
  }

  if (hasQueryState(context)) {
    return { action: 'reload' };
  }

  const controllerIds = new Set(uniqueControllerIds);
  const visibleDeviceIds = context.items
    .filter((item) => {
      const controllerId = item.sps_controller_system_type?.sps_controller_id;
      return controllerId ? controllerIds.has(controllerId) : false;
    })
    .map((item) => item.id);

  return visibleDeviceIds.length > 0
    ? { action: 'refresh', ids: visibleDeviceIds }
    : { action: 'none' };
}

export function applySPSControllerNameDelta(
  items: FieldDevice[],
  spsControllers: SPSController[]
): FieldDevice[] {
  const controllerNames = new Map(
    spsControllers
      .filter((item) => item.id && item.device_name)
      .map((item) => [item.id, item.device_name])
  );

  if (controllerNames.size === 0) {
    return items;
  }

  return items.map((item) => {
    const systemType = item.sps_controller_system_type;
    if (!systemType?.sps_controller_id) {
      return item;
    }

    const nextName = controllerNames.get(systemType.sps_controller_id);
    if (!nextName || systemType.sps_controller_name === nextName) {
      return item;
    }

    return {
      ...item,
      sps_controller_system_type: {
        ...systemType,
        sps_controller_name: nextName
      }
    };
  });
}
