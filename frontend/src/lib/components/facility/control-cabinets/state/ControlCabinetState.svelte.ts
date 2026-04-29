import { addToast } from '$lib/components/toast.svelte';
import { confirm } from '$lib/stores/confirm-dialog.js';
import { t as translate } from '$lib/i18n/index.js';
import { ManageControlCabinetUseCase } from '$lib/application/useCases/facility/manageControlCabinetUseCase.js';
import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
import { canPerform, canPerformAny } from '$lib/utils/permissions.js';
import { BaseDataTableState } from '$lib/state/table/BaseDataTableState.svelte.js';
import type { Building, ControlCabinet } from '$lib/domain/facility/index.js';
import type { ControlCabinetFilters, ControlCabinetStateProps } from './types.js';
import { toProjectIdResolver } from './types.js';
import { ControlCabinetFetchStrategyFactory } from './ControlCabinetFetchStrategyFactory.js';
import { ContextualControlCabinetFetchStrategy } from './strategies/ContextualControlCabinetFetchStrategy.js';

export class ControlCabinetState extends BaseDataTableState<ControlCabinet, ControlCabinetFilters> {
  showForm = $state(false);
  editingItem: ControlCabinet | undefined = $state(undefined);
  buildingMap = $state(new Map<string, string>());

  private readonly buildingRequests = new Set<string>();
  private readonly resolveProjectId: () => string | undefined;
  private readonly onChanged?: (
    event?: import('../../shared/entityRefresh.js').EntityChangeEvent<ControlCabinet>
  ) => void;
  private readonly manageControlCabinetUseCase = new ManageControlCabinetUseCase(
    controlCabinetRepository
  );
  private readonly fetchStrategy: ContextualControlCabinetFetchStrategy;

  constructor(props: ControlCabinetStateProps = {}) {
    const resolveProjectId = toProjectIdResolver(props.projectId);
    const strategyFactory = new ControlCabinetFetchStrategyFactory(resolveProjectId);
    const fetchStrategy = strategyFactory.create();

    super(fetchStrategy, { pageSize: props.pageSize ?? 10 });

    this.resolveProjectId = resolveProjectId;
    this.onChanged = props.onChanged;
    this.fetchStrategy = fetchStrategy;
  }

  get projectId(): string | undefined {
    return this.resolveProjectId();
  }

  get isProjectContext(): boolean {
    return Boolean(this.projectId);
  }

  canCreateControlCabinet(): boolean {
    return this.isProjectContext
      ? canPerformAny(['create', 'edit'], 'project.controlcabinet')
      : canPerform('create', 'controlcabinet');
  }

  canUpdateControlCabinet(): boolean {
    return this.isProjectContext
      ? canPerformAny(['update', 'edit'], 'project.controlcabinet')
      : canPerform('update', 'controlcabinet');
  }

  canDeleteControlCabinet(): boolean {
    return this.isProjectContext
      ? canPerformAny(['delete', 'edit'], 'project.controlcabinet')
      : canPerform('delete', 'controlcabinet');
  }

  async initialize(): Promise<void> {
    await this.load();
  }

  override async load(): Promise<void> {
    await super.load();

    if (this.error) return;

    this.mergeBuildingLabels(this.fetchStrategy.getBuildingLabels());
    await this.ensureBuildingLabels(this.items);
  }

  async refreshCabinets(controlCabinetIds: string[]): Promise<void> {
    const uniqueControlCabinetIDs = [...new Set(controlCabinetIds.filter(Boolean))];

    if (uniqueControlCabinetIDs.length === 0) {
      await this.reload();
      return;
    }

    if (this.searchText || this.orderBy || this.order || this.hasActiveFilters) {
      await this.reload();
      return;
    }

    const visibleIDs = new Set(this.items.map((item) => item.id));
    if (uniqueControlCabinetIDs.some((id) => !visibleIDs.has(id))) {
      await this.reload();
      return;
    }

    try {
      const updatedCabinets = await Promise.all(
        uniqueControlCabinetIDs.map((id) => controlCabinetRepository.get(id))
      );

      this.replaceItems(updatedCabinets);
      await this.ensureBuildingLabels(updatedCabinets);
    } catch (error) {
      console.error('Failed to refresh control cabinets:', error);
      await this.reload();
    }
  }

  async applyCabinetDelta(controlCabinets: ControlCabinet[]): Promise<void> {
    const updatedCabinets = [...new Map(controlCabinets.map((item) => [item.id, item])).values()];

    if (updatedCabinets.length === 0) {
      return;
    }

    if (this.searchText || this.orderBy || this.order || this.hasActiveFilters) {
      await this.reload();
      return;
    }

    const visibleIDs = new Set(this.items.map((item) => item.id));
    const visibleCabinets = updatedCabinets.filter((item) => visibleIDs.has(item.id));
    const hasNewCabinets = updatedCabinets.some((item) => !visibleIDs.has(item.id));

    if (hasNewCabinets) {
      await this.reload();
      return;
    }

    this.replaceItems(visibleCabinets);
    await this.ensureBuildingLabels(visibleCabinets);
  }

  openCreateForm(): void {
    this.editingItem = undefined;
    this.showForm = true;
  }

  editControlCabinet(controlCabinet: ControlCabinet): void {
    this.editingItem = controlCabinet;
    this.showForm = true;
  }

  cancelForm(): void {
    this.showForm = false;
    this.editingItem = undefined;
  }

  async handleFormSuccess(controlCabinet: ControlCabinet): Promise<void> {
    const isUpdate = Boolean(this.editingItem);

    if (this.projectId && !this.editingItem) {
      try {
        await projectRepository.addControlCabinet(this.projectId, controlCabinet.id);
        addToast(translate('projects.control_cabinets.created'), 'success');
      } catch (error) {
        const message =
          error instanceof Error
            ? error.message
            : translate('projects.control_cabinets.link_failed');
        addToast(message, 'error');
        return;
      }
    } else if (this.isProjectContext) {
      addToast(translate('projects.control_cabinets.updated'), 'success');
    }

    this.cancelForm();

    if (isUpdate) {
      await this.applyCabinetDelta([controlCabinet]);
      this.notifyChanged({ entityIds: [controlCabinet.id], items: [controlCabinet] });
      return;
    }

    await this.reload();
    this.notifyChanged({ entityIds: [controlCabinet.id] });
  }

  async deleteControlCabinet(controlCabinet: ControlCabinet): Promise<void> {
    if (!this.canDeleteControlCabinet()) return;

    if (this.isProjectContext) {
      await this.removeProjectControlCabinet(controlCabinet);
      return;
    }

    await this.deleteFacilityControlCabinet(controlCabinet);
  }

  async duplicateControlCabinet(controlCabinet: ControlCabinet): Promise<void> {
    if (!this.canCreateControlCabinet()) return;

    try {
      if (this.projectId) {
        await projectRepository.copyControlCabinet(this.projectId, controlCabinet.id);
        addToast(translate('projects.control_cabinets.duplicated'), 'success');
      } else {
        await this.manageControlCabinetUseCase.copy(controlCabinet.id);
        addToast(translate('facility.control_cabinet_copied'), 'success');
      }

      await this.reload();
      this.notifyChanged();
    } catch (error) {
      const message = error instanceof Error ? error.message : translate('facility.copy_failed');
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

  getBuildingLabel(buildingId: string | undefined | null): string {
    if (!buildingId) return '-';
    return this.buildingMap.get(buildingId) ?? buildingId;
  }

  private mergeBuildingLabels(labels: Map<string, string>): void {
    if (labels.size === 0) return;

    const next = new Map(this.buildingMap);
    for (const [buildingId, label] of labels.entries()) {
      next.set(buildingId, label);
    }
    this.buildingMap = next;
  }

  private async ensureBuildingLabels(controlCabinets: ControlCabinet[]): Promise<void> {
    const buildingIds = [
      ...new Set(controlCabinets.map((item) => item.building_id).filter(Boolean))
    ];
    const missingIds = buildingIds.filter(
      (id) => !this.buildingMap.has(id) && !this.buildingRequests.has(id)
    );

    if (missingIds.length === 0) return;

    for (const id of missingIds) {
      this.buildingRequests.add(id);
    }

    try {
      const buildings = await buildingRepository.getBulk(missingIds);
      this.updateBuildingMap(buildings);
    } catch (error) {
      console.error('Failed to load buildings:', error);
    } finally {
      for (const id of missingIds) {
        this.buildingRequests.delete(id);
      }
    }
  }

  private updateBuildingMap(buildings: Building[]): void {
    const next = new Map(this.buildingMap);
    for (const building of buildings) {
      next.set(building.id, `${building.iws_code}-${building.building_group}`);
    }
    this.buildingMap = next;
  }

  private async deleteFacilityControlCabinet(controlCabinet: ControlCabinet): Promise<void> {
    try {
      const impact = await this.manageControlCabinetUseCase.getDeleteImpact(controlCabinet.id);

      if (impact.sps_controllers_count > 0) {
        const firstConfirmation = await confirm({
          title: translate('facility.delete_control_cabinet_confirm'),
          message: translate('facility.delete_control_cabinet_message').replace(
            '{count}',
            impact.sps_controllers_count.toString()
          ),
          confirmText: translate('common.confirm'),
          cancelText: translate('common.cancel'),
          variant: 'destructive'
        });
        if (!firstConfirmation) return;

        const secondConfirmation = await confirm({
          title: translate('facility.confirm_cascading_delete'),
          message: translate('facility.cascading_delete_message')
            .replace('{systemTypes}', impact.sps_controller_system_types_count.toString())
            .replace('{fieldDevices}', impact.field_devices_count.toString())
            .replace('{bacnetObjects}', impact.bacnet_objects_count.toString()),
          confirmText: translate('facility.delete_everything'),
          cancelText: translate('common.cancel'),
          variant: 'destructive'
        });
        if (!secondConfirmation) return;
      }

      await this.manageControlCabinetUseCase.delete(controlCabinet.id);
      addToast(translate('facility.control_cabinet_deleted'), 'success');
      await this.reload();
      this.notifyChanged();
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : translate('facility.delete_control_cabinet_failed');
      addToast(message, 'error');
    }
  }

  private async removeProjectControlCabinet(controlCabinet: ControlCabinet): Promise<void> {
    const linkId = this.fetchStrategy.getLinkId(controlCabinet.id);
    if (!this.projectId || !linkId) return;

    const confirmed = await confirm({
      title: translate('projects.control_cabinets.remove_title'),
      message: translate('projects.control_cabinets.remove_message'),
      confirmText: translate('projects.control_cabinets.remove_confirm'),
      cancelText: translate('common.cancel'),
      variant: 'destructive'
    });
    if (!confirmed) return;

    try {
      await projectRepository.removeControlCabinet(this.projectId, linkId);
      addToast(translate('projects.control_cabinets.removed'), 'success');
      await this.reload();
      this.notifyChanged();
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : translate('projects.control_cabinets.remove_failed');
      addToast(message, 'error');
    }
  }

  private notifyChanged(
    event?: import('../../shared/entityRefresh.js').EntityChangeEvent<ControlCabinet>
  ): void {
    this.onChanged?.(event);
  }
}
