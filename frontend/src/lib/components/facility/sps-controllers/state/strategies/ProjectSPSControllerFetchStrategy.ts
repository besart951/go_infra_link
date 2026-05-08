import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
import type { ControlCabinet, SPSController } from '$lib/domain/facility/index.js';
import type { DataTableFetchStrategy, DataTableQuery } from '$lib/state/table/contracts.js';
import {
  decodeMultiFilter,
  sortMultiFilterOptions,
  type MultiFilterOption
} from '$lib/components/facility/shared/projectFacilityListFilters.js';
import { fetchAllPages } from '$lib/components/facility/shared/paginatedListFetcher.js';
import type { SPSControllerFilters } from '../types.js';

export class ProjectSPSControllerFetchStrategy implements DataTableFetchStrategy<
  SPSController,
  SPSControllerFilters
> {
  private readonly cabinetLabels = new Map<string, string>();
  private cabinetFilterOptions: MultiFilterOption[] = [];
  private readonly projectId: string;

  constructor(projectId: string) {
    this.projectId = projectId;
  }

  getProjectId(): string {
    return this.projectId;
  }

  getCabinetLabels(): Map<string, string> {
    return new Map(this.cabinetLabels);
  }

  getCabinetFilterOptions(): MultiFilterOption[] {
    return [...this.cabinetFilterOptions];
  }

  async fetch(query: DataTableQuery<SPSControllerFilters>, signal?: AbortSignal) {
    const links = await fetchAllPages(
      (page, pageSize, requestSignal) =>
        projectRepository.listSPSControllers(
          this.projectId,
          { page, limit: pageSize },
          requestSignal
        ),
      signal
    );

    const controllerIds = [...new Set(links.map((item) => item.sps_controller_id))];
    const controllers =
      controllerIds.length > 0 ? await spsControllerRepository.getBulk(controllerIds, signal) : [];

    await this.loadCabinetLabels(controllers, signal);
    this.updateCabinetFilterOptions(controllers);

    const filteredItems = this.filterItems(controllers, query.searchText, query.filters);
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

  private async loadCabinetLabels(
    controllers: SPSController[],
    signal?: AbortSignal
  ): Promise<void> {
    const cabinetIds = [
      ...new Set(controllers.map((item) => item.control_cabinet_id).filter(Boolean))
    ];
    if (cabinetIds.length === 0) return;

    const cabinets = await controlCabinetRepository.getBulk(cabinetIds, signal);
    this.updateCabinetLabels(cabinets);
  }

  private updateCabinetLabels(cabinets: ControlCabinet[]): void {
    for (const cabinet of cabinets) {
      this.cabinetLabels.set(cabinet.id, cabinet.control_cabinet_nr ?? cabinet.id);
    }
  }

  private updateCabinetFilterOptions(controllers: SPSController[]): void {
    const counts = new Map<string, number>();
    for (const controller of controllers) {
      if (!controller.control_cabinet_id) continue;
      counts.set(
        controller.control_cabinet_id,
        (counts.get(controller.control_cabinet_id) ?? 0) + 1
      );
    }

    this.cabinetFilterOptions = sortMultiFilterOptions(
      [...counts.entries()].map(([id, count]) => ({
        id,
        label: this.cabinetLabels.get(id) ?? id,
        count
      }))
    );
  }

  private filterItems(
    items: SPSController[],
    searchText: string,
    filters: SPSControllerFilters
  ): SPSController[] {
    const availableCabinetIds = new Set(
      items.map((item) => item.control_cabinet_id).filter(Boolean)
    );
    const selectedCabinetIds = decodeMultiFilter(filters.controlCabinetIds).filter((id) =>
      availableCabinetIds.has(id)
    );
    const scopedItems =
      selectedCabinetIds.length > 0
        ? items.filter((item) => selectedCabinetIds.includes(item.control_cabinet_id))
        : items;

    const query = searchText.trim().toLowerCase();
    if (!query) return scopedItems;

    return scopedItems.filter((item) =>
      [
        item.device_name,
        item.ga_device,
        item.ip_address,
        this.cabinetLabels.get(item.control_cabinet_id)
      ]
        .filter(Boolean)
        .some((value) => value!.toLowerCase().includes(query))
    );
  }
}
