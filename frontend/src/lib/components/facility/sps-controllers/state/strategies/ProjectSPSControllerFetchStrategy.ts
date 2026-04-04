import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
import type { ControlCabinet, SPSController } from '$lib/domain/facility/index.js';
import type { DataTableFetchStrategy, DataTableQuery } from '$lib/state/table/contracts.js';
import type { SPSControllerFilters } from '../types.js';

export class ProjectSPSControllerFetchStrategy implements DataTableFetchStrategy<
  SPSController,
  SPSControllerFilters
> {
  private readonly cabinetLabels = new Map<string, string>();
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

  async fetch(query: DataTableQuery<SPSControllerFilters>, signal?: AbortSignal) {
    const linksResponse = await projectRepository.listSPSControllers(
      this.projectId,
      { page: 1, limit: 1000 },
      signal
    );

    const controllerIds = [...new Set(linksResponse.items.map((item) => item.sps_controller_id))];
    const controllers =
      controllerIds.length > 0 ? await spsControllerRepository.getBulk(controllerIds, signal) : [];

    await this.loadCabinetLabels(controllers, signal);

    const filteredItems = this.filterItems(controllers, query.searchText);
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

  private filterItems(items: SPSController[], searchText: string): SPSController[] {
    const query = searchText.trim().toLowerCase();
    if (!query) return items;

    return items.filter((item) =>
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
