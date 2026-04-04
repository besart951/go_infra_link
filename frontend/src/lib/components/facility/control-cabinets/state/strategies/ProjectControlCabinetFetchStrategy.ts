import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
import type { Building, ControlCabinet } from '$lib/domain/facility/index.js';
import type { DataTableFetchStrategy, DataTableQuery } from '$lib/state/table/contracts.js';
import type { ControlCabinetFilters } from '../types.js';

export class ProjectControlCabinetFetchStrategy implements DataTableFetchStrategy<
  ControlCabinet,
  ControlCabinetFilters
> {
  private readonly buildingLabels = new Map<string, string>();
  private readonly linkIdsByCabinetId = new Map<string, string>();
  private readonly projectId: string;

  constructor(projectId: string) {
    this.projectId = projectId;
  }

  getProjectId(): string {
    return this.projectId;
  }

  getBuildingLabels(): Map<string, string> {
    return new Map(this.buildingLabels);
  }

  getLinkId(controlCabinetId: string): string | undefined {
    return this.linkIdsByCabinetId.get(controlCabinetId);
  }

  async fetch(query: DataTableQuery<ControlCabinetFilters>, signal?: AbortSignal) {
    const linksResponse = await projectRepository.listControlCabinets(
      this.projectId,
      { page: 1, limit: 1000 },
      signal
    );

    this.linkIdsByCabinetId.clear();
    for (const link of linksResponse.items) {
      this.linkIdsByCabinetId.set(link.control_cabinet_id, link.id);
    }

    const controlCabinetIds = linksResponse.items.map((item) => item.control_cabinet_id);
    const allCabinets =
      controlCabinetIds.length > 0
        ? await controlCabinetRepository.getBulk([...new Set(controlCabinetIds)], signal)
        : [];

    await this.loadBuildingLabels(allCabinets, signal);

    const filteredItems = this.filterItems(allCabinets, query.searchText);
    const total = filteredItems.length;
    const totalPages = total === 0 ? 0 : Math.ceil(total / query.pageSize);
    const page = totalPages === 0 ? 1 : Math.min(query.page, totalPages);
    const start = (page - 1) * query.pageSize;

    return {
      items: filteredItems.slice(start, start + query.pageSize),
      total,
      page,
      totalPages
    };
  }

  private async loadBuildingLabels(
    controlCabinets: ControlCabinet[],
    signal?: AbortSignal
  ): Promise<void> {
    const buildingIds = [
      ...new Set(controlCabinets.map((item) => item.building_id).filter(Boolean))
    ];
    if (buildingIds.length === 0) return;

    const buildings = await buildingRepository.getBulk(buildingIds, signal);
    this.updateBuildingLabels(buildings);
  }

  private updateBuildingLabels(buildings: Building[]): void {
    for (const building of buildings) {
      this.buildingLabels.set(building.id, `${building.iws_code}-${building.building_group}`);
    }
  }

  private filterItems(items: ControlCabinet[], searchText: string): ControlCabinet[] {
    const query = searchText.trim().toLowerCase();
    if (!query) return items;

    return items.filter((item) =>
      [item.control_cabinet_nr, this.buildingLabels.get(item.building_id)]
        .filter(Boolean)
        .some((value) => value!.toLowerCase().includes(query))
    );
  }
}
