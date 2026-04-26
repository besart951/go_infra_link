import { addToast } from '$lib/components/toast.svelte';
import { confirm } from '$lib/stores/confirm-dialog.js';
import { t as translate } from '$lib/i18n/index.js';
import { ManageSPSControllerUseCase } from '$lib/application/useCases/facility/manageSPSControllerUseCase.js';
import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
import { spsControllerSystemTypeRepository } from '$lib/infrastructure/api/spsControllerSystemTypeRepository.js';
import { canPerform } from '$lib/utils/permissions.js';
import { BaseDataTableState } from '$lib/state/table/BaseDataTableState.svelte.js';
import type {
  ControlCabinet,
  SPSController,
  SPSControllerSystemType
} from '$lib/domain/facility/index.js';
import type { SPSControllerFilters, SPSControllerStateProps } from './types.js';
import { toProjectIdResolver, toRefreshKeyResolver } from './types.js';
import { ContextualSPSControllerFetchStrategy } from './strategies/ContextualSPSControllerFetchStrategy.js';
import { SPSControllerFetchStrategyFactory } from './SPSControllerFetchStrategyFactory.js';
import { groupSystemTypesByController } from './groupSystemTypesByController.js';

export class SPSControllerState extends BaseDataTableState<SPSController, SPSControllerFilters> {
  showForm = $state(false);
  editingItem: SPSController | undefined = $state(undefined);
  cabinetMap = $state(new Map<string, string>());
  systemTypesByController = $state<Record<string, SPSControllerSystemType[]>>({});

  private readonly resolveProjectId: () => string | undefined;
  private readonly resolveControlCabinetRefreshKey: () => string | number | undefined;
  private readonly onChanged?: (
    event?: import('../../shared/entityRefresh.js').EntityChangeEvent<SPSController>
  ) => void;
  private readonly manageSPSControllerUseCase = new ManageSPSControllerUseCase(
    spsControllerRepository
  );
  private readonly fetchStrategy: ContextualSPSControllerFetchStrategy;

  constructor(props: SPSControllerStateProps = {}) {
    const resolveProjectId = toProjectIdResolver(props.projectId);
    const strategyFactory = new SPSControllerFetchStrategyFactory(resolveProjectId);
    const fetchStrategy = strategyFactory.create();

    super(fetchStrategy, { pageSize: props.pageSize ?? 10 });

    this.resolveProjectId = resolveProjectId;
    this.resolveControlCabinetRefreshKey = toRefreshKeyResolver(props.controlCabinetRefreshKey);
    this.onChanged = props.onChanged;
    this.fetchStrategy = fetchStrategy;
  }

  get projectId(): string | undefined {
    return this.resolveProjectId();
  }

  get controlCabinetRefreshKey(): string | number | undefined {
    return this.resolveControlCabinetRefreshKey();
  }

  get isProjectContext(): boolean {
    return Boolean(this.projectId);
  }

  async initialize(): Promise<void> {
    await this.load();
  }

  override async load(): Promise<void> {
    await super.load();

    if (this.error) return;

    this.mergeCabinetLabels(this.fetchStrategy.getCabinetLabels());
    await Promise.all([this.ensureCabinetLabels(this.items), this.loadSystemTypes(this.items)]);
  }

  async refreshControllers(controllerIds: string[]): Promise<void> {
    const uniqueControllerIDs = [...new Set(controllerIds.filter(Boolean))];

    if (uniqueControllerIDs.length === 0) {
      await this.reload();
      return;
    }

    if (this.searchText || this.orderBy || this.order || this.hasActiveFilters) {
      await this.reload();
      return;
    }

    const visibleIDs = new Set(this.items.map((item) => item.id));
    if (uniqueControllerIDs.some((id) => !visibleIDs.has(id))) {
      await this.reload();
      return;
    }

    try {
      const updatedControllers = await Promise.all(
        uniqueControllerIDs.map((id) => spsControllerRepository.get(id))
      );

      this.replaceItems(updatedControllers);
      await Promise.all([
        this.ensureCabinetLabels(updatedControllers),
        this.loadSystemTypesForControllerIDs(uniqueControllerIDs)
      ]);
    } catch (error) {
      console.error('Failed to refresh SPS controllers:', error);
      await this.reload();
    }
  }

  async applyControllerDelta(controllers: SPSController[]): Promise<void> {
    const updatedControllers = [...new Map(controllers.map((item) => [item.id, item])).values()];

    if (updatedControllers.length === 0) {
      return;
    }

    if (this.searchText || this.orderBy || this.order || this.hasActiveFilters) {
      await this.reload();
      return;
    }

    const visibleIDs = new Set(this.items.map((item) => item.id));
    const visibleControllers = updatedControllers.filter((item) => visibleIDs.has(item.id));
    const hasNewControllers = updatedControllers.some((item) => !visibleIDs.has(item.id));

    if (hasNewControllers) {
      await this.reload();
      return;
    }

    this.replaceItems(visibleControllers);
    await this.ensureCabinetLabels(visibleControllers);
  }

  async refreshCabinetLabels(cabinetIds: string[]): Promise<void> {
    const uniqueCabinetIDs = [...new Set(cabinetIds.filter(Boolean))];

    if (uniqueCabinetIDs.length === 0) {
      return;
    }

    try {
      const cabinets = await controlCabinetRepository.getBulk(uniqueCabinetIDs);
      this.updateCabinetMap(cabinets);
    } catch (error) {
      console.error('Failed to refresh control cabinet labels:', error);
      await this.reload();
    }
  }

  applyCabinetLabelDelta(cabinets: ControlCabinet[]): void {
    if (cabinets.length === 0) {
      return;
    }

    this.updateCabinetMap(cabinets);
  }

  openCreateForm(): void {
    this.editingItem = undefined;
    this.showForm = true;
  }

  editSPSController(controller: SPSController): void {
    this.editingItem = controller;
    this.showForm = true;
  }

  cancelForm(): void {
    this.showForm = false;
    this.editingItem = undefined;
  }

  async handleFormSuccess(controller: SPSController): Promise<void> {
    const isUpdate = Boolean(this.editingItem);

    if (this.projectId && !this.editingItem) {
      try {
        await projectRepository.addSPSController(this.projectId, controller.id);
        addToast(translate('projects.sps_controllers.created'), 'success');
      } catch (error) {
        const message =
          error instanceof Error
            ? error.message
            : translate('projects.sps_controllers.save_failed');
        addToast(message, 'error');
        return;
      }
    } else if (this.isProjectContext) {
      addToast(translate('projects.sps_controllers.updated'), 'success');
    }

    this.cancelForm();

    if (isUpdate) {
      await this.applyControllerDelta([controller]);
      this.notifyChanged({ entityIds: [controller.id], items: [controller] });
      return;
    }

    await this.reload();
    this.notifyChanged({ entityIds: [controller.id] });
  }

  async deleteSPSController(controller: SPSController): Promise<void> {
    if (!canPerform('delete', 'spscontroller')) return;

    const confirmed = await confirm({
      title: this.isProjectContext
        ? translate('projects.sps_controllers.delete_title')
        : translate('common.delete'),
      message: this.isProjectContext
        ? translate('projects.sps_controllers.delete_message', { name: controller.device_name })
        : translate('facility.delete_sps_controller_confirm').replace(
            '{name}',
            controller.device_name
          ),
      confirmText: translate('common.delete'),
      cancelText: translate('common.cancel'),
      variant: 'destructive'
    });
    if (!confirmed) return;

    try {
      await this.manageSPSControllerUseCase.delete(controller.id);
      addToast(
        this.isProjectContext
          ? translate('projects.sps_controllers.deleted')
          : translate('facility.sps_controller_deleted'),
        'success'
      );
      await this.reload();
      this.notifyChanged({ entityIds: [controller.id] });
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : this.isProjectContext
            ? translate('projects.sps_controllers.delete_failed')
            : translate('facility.delete_sps_controller_failed');
      addToast(message, 'error');
    }
  }

  async duplicateSPSController(controller: SPSController): Promise<void> {
    if (!canPerform('create', 'spscontroller')) return;

    try {
      if (this.projectId) {
        await projectRepository.copySPSController(this.projectId, controller.id);
        addToast(translate('projects.sps_controllers.duplicated'), 'success');
      } else {
        await this.manageSPSControllerUseCase.copy(controller.id);
        addToast(translate('facility.sps_controller_copied'), 'success');
      }

      await this.reload();
      this.notifyChanged({ entityIds: [controller.id] });
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : this.isProjectContext
            ? translate('projects.sps_controllers.duplicate_failed')
            : translate('facility.copy_failed');
      addToast(message, 'error');
    }
  }

  async copyToClipboard(value: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(value);
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
    }
  }

  getCabinetLabel(cabinetId: string | undefined | null): string {
    if (!cabinetId) return '-';
    return this.cabinetMap.get(cabinetId) ?? cabinetId;
  }

  hasLoadedSystemTypes(controllerId: string): boolean {
    return Object.prototype.hasOwnProperty.call(this.systemTypesByController, controllerId);
  }

  getSystemTypes(controllerId: string): SPSControllerSystemType[] {
    return this.systemTypesByController[controllerId] ?? [];
  }

  formatSystemTypeTitle(systemType: SPSControllerSystemType): string {
    return systemType.system_type_name ?? systemType.system_type_id;
  }

  formatSystemTypeMeta(systemType: SPSControllerSystemType): string {
    const parts: string[] = [];
    if (systemType.number != null) {
      parts.push(`${translate('facility.control_cabinet_detail.number')}: ${systemType.number}`);
    }
    if (systemType.document_name) {
      parts.push(systemType.document_name);
    }
    return parts.join(' | ');
  }

  private mergeCabinetLabels(labels: Map<string, string>): void {
    if (labels.size === 0) return;

    const next = new Map(this.cabinetMap);
    for (const [cabinetId, label] of labels.entries()) {
      next.set(cabinetId, label);
    }
    this.cabinetMap = next;
  }

  private async ensureCabinetLabels(controllers: SPSController[]): Promise<void> {
    const cabinetIds = [
      ...new Set(controllers.map((item) => item.control_cabinet_id).filter(Boolean))
    ];
    const missingIds = cabinetIds.filter((id) => !this.cabinetMap.has(id));

    if (missingIds.length === 0) return;

    try {
      const cabinets = await controlCabinetRepository.getBulk(missingIds);
      this.updateCabinetMap(cabinets);
    } catch (error) {
      console.error('Failed to load control cabinets:', error);
    }
  }

  private updateCabinetMap(cabinets: ControlCabinet[]): void {
    const next = new Map(this.cabinetMap);
    for (const cabinet of cabinets) {
      next.set(cabinet.id, cabinet.control_cabinet_nr ?? cabinet.id);
    }
    this.cabinetMap = next;
  }

  private async loadSystemTypes(controllers: SPSController[]): Promise<void> {
    const controllerIds = [...new Set(controllers.map((item) => item.id).filter(Boolean))];

    if (controllerIds.length === 0) {
      this.systemTypesByController = {};
      return;
    }

    if (this.projectId) {
      try {
        const response = await spsControllerSystemTypeRepository.list({
          pagination: { page: 1, pageSize: 1000 },
          search: { text: '' },
          filters: { project_id: this.projectId }
        });

        this.systemTypesByController = groupSystemTypesByController(response.items);
      } catch (error) {
        console.error('Failed to load project SPS controller system types:', error);
        this.systemTypesByController = {};
      }

      return;
    }

    await this.loadSystemTypesForControllerIDs(controllerIds, {});
  }

  private async loadSystemTypesForControllerIDs(
    controllerIds: string[],
    current: Record<string, SPSControllerSystemType[]> = this.systemTypesByController
  ): Promise<void> {
    if (controllerIds.length === 0) {
      return;
    }

    const results = await Promise.allSettled(
      controllerIds.map(async (controllerId) => {
        const response = await spsControllerSystemTypeRepository.list({
          pagination: { page: 1, pageSize: 1000 },
          search: { text: '' },
          filters: this.buildSystemTypeFilters(controllerId)
        });

        return { controllerId, items: response.items };
      })
    );

    const next: Record<string, SPSControllerSystemType[]> = { ...current };

    for (const result of results) {
      if (result.status === 'fulfilled') {
        next[result.value.controllerId] = result.value.items;
        continue;
      }

      console.error('Failed to load SPS controller system types:', result.reason);
    }

    this.systemTypesByController = next;
  }

  private buildSystemTypeFilters(controllerId: string): Record<string, string> {
    const filters: Record<string, string> = { sps_controller_id: controllerId };

    if (this.projectId) {
      filters.project_id = this.projectId;
    }

    return filters;
  }

  private notifyChanged(
    event?: import('../../shared/entityRefresh.js').EntityChangeEvent<SPSController>
  ): void {
    this.onChanged?.(event);
  }
}
