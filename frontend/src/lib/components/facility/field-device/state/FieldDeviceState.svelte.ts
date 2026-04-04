import { addToast } from '$lib/components/toast.svelte';
import { useFieldDeviceEditing } from '$lib/hooks/useFieldDeviceEditing.svelte.js';
import { t as translate } from '$lib/i18n/index.js';
import { ManageFieldDeviceUseCase } from '$lib/application/useCases/facility/manageFieldDeviceUseCase.js';
import { ListEntityUseCase } from '$lib/application/useCases/listEntityUseCase.js';
import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';
import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
import { canPerform } from '$lib/utils/permissions.js';
import { BaseDataTableState } from '$lib/state/table/BaseDataTableState.svelte.js';
import type { Apparat, FieldDevice, SystemPart } from '$lib/domain/facility/index.js';
import type { FieldDeviceFilters, FieldDeviceStateProps } from './types.js';
import { toProjectIdResolver } from './types.js';
import { FieldDeviceFetchStrategyFactory } from './strategies/FieldDeviceFetchStrategyFactory.js';

export class FieldDeviceState extends BaseDataTableState<FieldDevice, FieldDeviceFilters> {
  readonly editing: ReturnType<typeof useFieldDeviceEditing>;

  allApparats = $state<Apparat[]>([]);
  allSystemParts = $state<SystemPart[]>([]);
  showMultiCreateForm = $state(false);
  bulkEditPanelOpen = $state(false);
  showExportPanel = $state(false);
  showFilterPanel = $state(false);
  showSpecifications = $state(false);
  expandedBacnetRows = $state<Set<string>>(new Set());

  readonly showBulkEditPanel = $derived.by(() => this.bulkEditPanelOpen && this.selectedCount > 0);

  private readonly resolveProjectId: () => string | undefined;
  private readonly manageFieldDeviceUseCase = new ManageFieldDeviceUseCase(fieldDeviceRepository);
  private readonly listApparatsUseCase = new ListEntityUseCase(apparatRepository);
  private readonly listSystemPartsUseCase = new ListEntityUseCase(systemPartRepository);

  constructor(props: FieldDeviceStateProps = {}) {
    const resolveProjectId = toProjectIdResolver(props.projectId);
    const strategyFactory = new FieldDeviceFetchStrategyFactory(resolveProjectId);

    super(strategyFactory.create(), { pageSize: props.pageSize ?? 300 });

    this.resolveProjectId = resolveProjectId;
    this.editing = useFieldDeviceEditing(() => this.projectId);
  }

  get projectId() {
    return this.resolveProjectId();
  }

  protected override onSelectionChanged() {
    if (this.selectedIds.size === 0) {
      this.bulkEditPanelOpen = false;
    }
  }

  async initialize(): Promise<void> {
    await Promise.all([this.load(), this.loadLookups()]);
  }

  private async loadLookups(): Promise<void> {
    const [apparatsResult, systemPartsResult] = await Promise.allSettled([
      this.listApparatsUseCase.execute({
        pagination: { page: 1, pageSize: 1000 },
        search: { text: '' }
      }),
      this.listSystemPartsUseCase.execute({
        pagination: { page: 1, pageSize: 1000 },
        search: { text: '' }
      })
    ]);

    if (apparatsResult.status === 'fulfilled') {
      this.allApparats = apparatsResult.value.items;
    } else {
      console.error('Failed to load apparats', apparatsResult.reason);
    }

    if (systemPartsResult.status === 'fulfilled') {
      this.allSystemParts = systemPartsResult.value.items;
    } else {
      console.error('Failed to load system parts', systemPartsResult.reason);
    }
  }

  async applyFilters(filters: FieldDeviceFilters): Promise<void> {
    await this.setFilters(filters);
  }

  async clearFilters(): Promise<void> {
    await this.clearAllFilters();
  }

  openMultiCreateForm(): void {
    this.showMultiCreateForm = true;
  }

  closeMultiCreateForm(): void {
    this.showMultiCreateForm = false;
  }

  toggleBulkEditPanel(): void {
    if (this.selectedCount === 0) return;
    this.bulkEditPanelOpen = !this.bulkEditPanelOpen;
  }

  toggleExportPanel(): void {
    this.showExportPanel = !this.showExportPanel;
  }

  toggleFilterPanel(): void {
    this.showFilterPanel = !this.showFilterPanel;
  }

  toggleSpecifications(): void {
    this.showSpecifications = !this.showSpecifications;
  }

  toggleBacnetExpansion(deviceId: string): void {
    const nextExpanded = new Set(this.expandedBacnetRows);
    if (nextExpanded.has(deviceId)) {
      nextExpanded.delete(deviceId);
    } else {
      nextExpanded.add(deviceId);
    }

    this.expandedBacnetRows = nextExpanded;
  }

  isBacnetExpanded(deviceId: string): boolean {
    return this.expandedBacnetRows.has(deviceId);
  }

  async copyToClipboard(value: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(value);
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
    }
  }

  async deleteDevice(device: FieldDevice): Promise<void> {
    if (!canPerform('delete', 'fielddevice')) return;
    if (
      !confirm(
        translate('field_device.confirm.delete', {
          label: device.bmk ?? device.id
        })
      )
    ) {
      return;
    }

    try {
      await this.manageFieldDeviceUseCase.delete(device.id);
      addToast(translate('field_device.toasts.deleted'), 'success');

      const nextSelectedIds = new Set(this.selectedIds);
      nextSelectedIds.delete(device.id);
      this.selectedIds = nextSelectedIds;
      this.onSelectionChanged();

      await this.reload();
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      addToast(translate('field_device.toasts.delete_failed', { message }), 'error');
    }
  }

  async bulkDeleteSelected(): Promise<void> {
    if (this.selectedIds.size === 0) return;
    if (!canPerform('delete', 'fielddevice')) return;

    const ids = [...this.selectedIds];
    if (!confirm(translate('field_device.confirm.bulk_delete', { count: ids.length }))) {
      return;
    }

    try {
      const result = await this.manageFieldDeviceUseCase.bulkDelete(ids);

      if (result.success_count > 0) {
        addToast(
          translate('field_device.toasts.bulk_deleted', { count: result.success_count }),
          'success'
        );
      }

      if (result.failure_count > 0) {
        addToast(
          translate('field_device.toasts.bulk_delete_failed', { count: result.failure_count }),
          'error'
        );
      }

      this.selectedIds = new Set();
      this.onSelectionChanged();
      await this.reload();
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      addToast(translate('field_device.toasts.bulk_delete_failed_message', { message }), 'error');
    }
  }

  async handleMultiCreateSuccess(createdDevices: FieldDevice[]): Promise<void> {
    this.showMultiCreateForm = false;

    if (this.projectId) {
      try {
        await Promise.all(
          createdDevices.map((device) =>
            projectRepository.addFieldDevice(this.projectId!, device.id)
          )
        );
      } catch (error) {
        const message =
          error instanceof Error
            ? error.message
            : translate('field_device.toasts.partial_link_failed');
        addToast(translate('field_device.toasts.link_failed', { message }), 'error');
      }
    }

    await this.reload();
  }

  savePendingEdits(): void {
    this.editing.saveAllPendingEdits(this.items, (updatedItems) => {
      this.replaceItems(updatedItems);
    });
  }

  discardPendingEdits(): void {
    this.editing.discardAllEdits();
  }
}
