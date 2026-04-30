import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
import type {
  Building,
  ControlCabinet,
  FieldDevice,
  SPSController
} from '$lib/domain/facility/index.js';
import type { FieldDeviceGroupKey } from './FieldDeviceTableView.svelte.js';

export interface FieldDeviceGroupingLookups {
  spsControllers: Map<string, SPSController>;
  controlCabinets: Map<string, ControlCabinet>;
  buildings: Map<string, Building>;
}

interface LoadGroupingLookupsInput extends FieldDeviceGroupingLookups {
  items: FieldDevice[];
  activeGroups: Set<FieldDeviceGroupKey>;
}

export class FieldDeviceGroupingLookupService {
  async loadForVisibleDevices(
    input: LoadGroupingLookupsInput
  ): Promise<FieldDeviceGroupingLookups> {
    const shouldLoadControllers =
      input.activeGroups.has('building') ||
      input.activeGroups.has('controlCabinet') ||
      input.activeGroups.has('spsController');
    const shouldLoadCabinets =
      input.activeGroups.has('building') || input.activeGroups.has('controlCabinet');
    const shouldLoadBuildings = input.activeGroups.has('building');

    let spsControllers = input.spsControllers;
    let controlCabinets = input.controlCabinets;
    let buildings = input.buildings;

    if (!shouldLoadControllers && !shouldLoadCabinets && !shouldLoadBuildings) {
      return { spsControllers, controlCabinets, buildings };
    }

    const visibleControllerIds = [
      ...new Set(
        input.items
          .map((item) => item.sps_controller_system_type?.sps_controller_id)
          .filter((id): id is string => Boolean(id))
      )
    ];

    if (shouldLoadControllers) {
      const missingControllerIds = visibleControllerIds.filter((id) => !spsControllers.has(id));
      if (missingControllerIds.length > 0) {
        spsControllers = mergeLookupItems(
          spsControllers,
          await spsControllerRepository.getBulk(missingControllerIds)
        );
      }
    }

    if (shouldLoadCabinets) {
      const visibleCabinetIds = [
        ...new Set(
          visibleControllerIds
            .map((id) => spsControllers.get(id)?.control_cabinet_id)
            .filter((id): id is string => Boolean(id))
        )
      ];
      const missingCabinetIds = visibleCabinetIds.filter((id) => !controlCabinets.has(id));
      if (missingCabinetIds.length > 0) {
        controlCabinets = mergeLookupItems(
          controlCabinets,
          await controlCabinetRepository.getBulk(missingCabinetIds)
        );
      }
    }

    if (shouldLoadBuildings) {
      const visibleBuildingIds = [
        ...new Set(
          [...controlCabinets.values()]
            .map((cabinet) => cabinet.building_id)
            .filter((id): id is string => Boolean(id))
        )
      ];
      const missingBuildingIds = visibleBuildingIds.filter((id) => !buildings.has(id));
      if (missingBuildingIds.length > 0) {
        buildings = mergeLookupItems(
          buildings,
          await buildingRepository.getBulk(missingBuildingIds)
        );
      }
    }

    return { spsControllers, controlCabinets, buildings };
  }
}

function mergeLookupItems<TItem extends { id: string }>(
  current: Map<string, TItem>,
  items: TItem[]
): Map<string, TItem> {
  const next = new Map(current);
  for (const item of items) {
    next.set(item.id, item);
  }
  return next;
}
