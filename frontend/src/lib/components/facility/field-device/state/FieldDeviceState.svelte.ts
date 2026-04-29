import { addToast } from '$lib/components/toast.svelte';
import { useFieldDeviceEditing } from '$lib/hooks/useFieldDeviceEditing.svelte.js';
import { t as translate } from '$lib/i18n/index.js';
import { ManageFieldDeviceUseCase } from '$lib/application/useCases/facility/manageFieldDeviceUseCase.js';
import { ListEntityUseCase } from '$lib/application/useCases/listEntityUseCase.js';
import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';
import { canPerform, canPerformAny } from '$lib/utils/permissions.js';
import { BaseDataTableState } from '$lib/state/table/BaseDataTableState.svelte.js';
import type { Apparat, FieldDevice, SystemPart } from '$lib/domain/facility/index.js';
import type {
  FieldDeviceFilters,
  FieldDeviceStateProps,
  SharedFieldDeviceEditorsByDevice
} from './types.js';
import { resolvePageSize, toProjectIdResolver } from './types.js';
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
  loadingBacnetRows = $state<Set<string>>(new Set());
  loadingSpecifications = $state(false);

  readonly showBulkEditPanel = $derived.by(() => this.bulkEditPanelOpen && this.selectedCount > 0);

  private readonly resolveProjectId: () => string | undefined;
  private readonly resolveSharedFieldDeviceEditors: () => SharedFieldDeviceEditorsByDevice;
  private readonly onFieldDevicesSaved?: (devices: FieldDevice[]) => void;
  private readonly manageFieldDeviceUseCase = new ManageFieldDeviceUseCase(fieldDeviceRepository);
  private readonly listApparatsUseCase = new ListEntityUseCase(apparatRepository);
  private readonly listSystemPartsUseCase = new ListEntityUseCase(systemPartRepository);

  constructor(props: FieldDeviceStateProps = {}) {
    const resolveProjectId = toProjectIdResolver(props.projectId);
    const strategyFactory = new FieldDeviceFetchStrategyFactory(resolveProjectId);

    super(strategyFactory.create(), { pageSize: resolvePageSize(props.pageSize) ?? 300 });

    this.resolveProjectId = resolveProjectId;
    this.resolveSharedFieldDeviceEditors = props.sharedFieldDeviceEditors ?? (() => ({}));
    this.onFieldDevicesSaved = props.onFieldDevicesSaved;
    this.editing = useFieldDeviceEditing({
      projectId: () => this.projectId,
      onSharedStateChange: props.onSharedFieldDeviceStateChange,
      onSaveSuccess: () => undefined
    });
  }

  get projectId() {
    return this.resolveProjectId();
  }

  get isProjectContext(): boolean {
    return Boolean(this.projectId);
  }

  private canPerformProjectFieldDevice(...actions: string[]): boolean {
    return canPerformAny(actions, 'project.fielddevice');
  }

  private canPerformProjectFieldDeviceSpecification(...actions: string[]): boolean {
    return canPerformAny(actions, 'project.fielddevice_specification');
  }

  private canPerformProjectFieldDeviceBacnetObjects(...actions: string[]): boolean {
    return canPerformAny(actions, 'project.fielddevice.bacnetobjects');
  }

  canCreateFieldDevice(): boolean {
    return this.isProjectContext
      ? this.canPerformProjectFieldDevice('create', 'edit')
      : canPerform('create', 'fielddevice');
  }

  canUpdateFieldDevice(): boolean {
    return this.isProjectContext
      ? this.canPerformProjectFieldDevice('update', 'edit')
      : canPerform('update', 'fielddevice');
  }

  canDeleteFieldDevice(): boolean {
    return this.isProjectContext
      ? this.canPerformProjectFieldDevice('delete', 'edit')
      : canPerform('delete', 'fielddevice');
  }

  canUpdateFieldDeviceSpecification(): boolean {
    if (!this.isProjectContext) {
      return this.canUpdateFieldDevice();
    }

    return (
      this.canPerformProjectFieldDeviceSpecification('update', 'edit') ||
      this.canPerformProjectFieldDevice('edit')
    );
  }

  canUpdateFieldDeviceBacnetObjects(): boolean {
    if (!this.isProjectContext) {
      return this.canUpdateFieldDevice();
    }

    return (
      this.canPerformProjectFieldDeviceBacnetObjects('update', 'edit') ||
      this.canPerformProjectFieldDevice('edit')
    );
  }

  canOpenBulkEditPanel(): boolean {
    if (!this.isProjectContext) {
      return this.canUpdateFieldDevice();
    }

    return this.canUpdateFieldDevice() || this.canUpdateFieldDeviceSpecification();
  }

  canSavePendingEdits(): boolean {
    if (!this.editing.hasUnsavedChanges) {
      return false;
    }

    if (!this.isProjectContext) {
      return this.canUpdateFieldDevice();
    }

    if (this.editing.hasPendingBaseEdits && !this.canUpdateFieldDevice()) {
      return false;
    }

    if (this.editing.hasPendingSpecificationEdits && !this.canUpdateFieldDeviceSpecification()) {
      return false;
    }

    if (this.editing.hasPendingBacnetEdits && !this.canUpdateFieldDeviceBacnetObjects()) {
      return false;
    }

    return true;
  }

  getEditorsForDevice(deviceId: string) {
    return this.resolveSharedFieldDeviceEditors()[deviceId] ?? [];
  }

  protected override onSelectionChanged() {
    if (this.selectedIds.size === 0) {
      this.bulkEditPanelOpen = false;
    }
  }

  async initialize(): Promise<void> {
    await Promise.all([this.load(), this.loadLookups()]);
  }

  override async load(): Promise<void> {
    await super.load();

    if (!this.showSpecifications || this.error || this.loading) {
      return;
    }

    await this.loadSpecificationDetailsForVisibleDevices();
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
    if (!this.canOpenBulkEditPanel()) return;
    this.bulkEditPanelOpen = !this.bulkEditPanelOpen;
  }

  toggleExportPanel(): void {
    this.showExportPanel = !this.showExportPanel;
  }

  toggleFilterPanel(): void {
    this.showFilterPanel = !this.showFilterPanel;
  }

  async toggleSpecifications(): Promise<void> {
    const nextShowSpecifications = !this.showSpecifications;
    this.showSpecifications = nextShowSpecifications;

    if (nextShowSpecifications) {
      await this.loadSpecificationDetailsForVisibleDevices();
    }
  }

  async toggleBacnetExpansion(deviceId: string): Promise<void> {
    const nextExpanded = new Set(this.expandedBacnetRows);
    if (nextExpanded.has(deviceId)) {
      nextExpanded.delete(deviceId);
    } else {
      nextExpanded.add(deviceId);
    }

    this.expandedBacnetRows = nextExpanded;

    if (nextExpanded.has(deviceId)) {
      await this.loadBacnetObjectsForDevice(deviceId);
    }
  }

  isBacnetExpanded(deviceId: string): boolean {
    return this.expandedBacnetRows.has(deviceId);
  }

  isBacnetLoading(deviceId: string): boolean {
    return this.loadingBacnetRows.has(deviceId);
  }

  async copyToClipboard(value: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(value);
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
    }
  }

  async deleteDevice(device: FieldDevice): Promise<void> {
    if (!this.canDeleteFieldDevice()) return;
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
      if (this.projectId) {
        await this.removeProjectFieldDevice(device.id);
      } else {
        await this.manageFieldDeviceUseCase.delete(device.id);
      }
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
    if (!this.canDeleteFieldDevice()) return;

    const ids = [...this.selectedIds];
    if (!confirm(translate('field_device.confirm.bulk_delete', { count: ids.length }))) {
      return;
    }

    try {
      const result = this.projectId
        ? await this.removeProjectFieldDevices(ids)
        : await this.manageFieldDeviceUseCase.bulkDelete(ids);

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

  async handleMultiCreateSuccess(_createdDevices: FieldDevice[]): Promise<void> {
    this.showMultiCreateForm = false;

    await this.reload();
  }

  async refreshDevices(deviceIds: string[]): Promise<void> {
    const uniqueDeviceIds = [...new Set(deviceIds.filter(Boolean))];

    if (uniqueDeviceIds.length === 0) {
      await this.reload();
      return;
    }

    if (this.searchText || this.orderBy || this.order || this.hasActiveFilters) {
      await this.reload();
      return;
    }

    const visibleIds = new Set(this.items.map((item) => item.id));
    if (uniqueDeviceIds.some((id) => !visibleIds.has(id))) {
      await this.reload();
      return;
    }

    try {
      const updatedItems = await Promise.all(
        uniqueDeviceIds.map((id) => fieldDeviceRepository.get(id))
      );

      this.replaceItems(updatedItems);
    } catch (error) {
      console.error('Failed to refresh field devices:', error);
      await this.reload();
    }
  }

  async applyDeviceDelta(fieldDevices: FieldDevice[]): Promise<void> {
    const updatedDevices = [...new Map(fieldDevices.map((item) => [item.id, item])).values()];

    if (updatedDevices.length === 0) {
      return;
    }

    if (this.searchText || this.orderBy || this.order || this.hasActiveFilters) {
      await this.reload();
      return;
    }

    const visibleIDs = new Set(this.items.map((item) => item.id));
    const visibleDevices = updatedDevices.filter((item) => visibleIDs.has(item.id));
    const hasNewDevices = updatedDevices.some((item) => !visibleIDs.has(item.id));

    if (hasNewDevices) {
      await this.reload();
      return;
    }

    this.replaceItems(visibleDevices);
  }

  async refreshDevicesForSPSControllers(spsControllerIds: string[]): Promise<void> {
    const uniqueSPSControllerIDs = [...new Set(spsControllerIds.filter(Boolean))];

    if (uniqueSPSControllerIDs.length === 0) {
      await this.reload();
      return;
    }

    if (this.searchText || this.orderBy || this.order || this.hasActiveFilters) {
      await this.reload();
      return;
    }

    const controllerIDs = new Set(uniqueSPSControllerIDs);
    const visibleDeviceIDs = this.items
      .filter((item) => {
        const controllerID = item.sps_controller_system_type?.sps_controller_id;
        return controllerID ? controllerIDs.has(controllerID) : false;
      })
      .map((item) => item.id);

    if (visibleDeviceIDs.length === 0) {
      return;
    }

    await this.refreshDevices(visibleDeviceIDs);
  }

  applySPSControllerDelta(
    spsControllers: import('$lib/domain/facility/index.js').SPSController[]
  ): void {
    const controllerNames = new Map(
      spsControllers
        .filter((item) => item.id && item.device_name)
        .map((item) => [item.id, item.device_name])
    );

    if (controllerNames.size === 0) {
      return;
    }

    this.items = this.items.map((item) => {
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

  savePendingEdits(): void {
    if (!this.canSavePendingEdits()) return;
    this.editing.saveAllPendingEdits(this.items, (updatedItems) => {
      this.replaceItems(updatedItems);
      this.onFieldDevicesSaved?.(updatedItems);
    });
  }

  discardPendingEdits(): void {
    this.editing.discardAllEdits();
  }

  private async removeProjectFieldDevice(deviceId: string): Promise<void> {
    const linkId = await this.resolveProjectFieldDeviceLinkId(deviceId);

    if (!this.projectId || !linkId) {
      throw new Error(translate('projects.errors.load_failed'));
    }

    await projectRepository.removeFieldDevice(this.projectId, linkId);
  }

  private async removeProjectFieldDevices(ids: string[]): Promise<{
    results: Array<{ id: string; success: boolean }>;
    total_count: number;
    success_count: number;
    failure_count: number;
  }> {
    const linkIdsByDeviceId = await this.loadProjectFieldDeviceLinkIds();

    const results = await Promise.all(
      ids.map(async (id) => {
        const linkId = linkIdsByDeviceId.get(id);
        if (!this.projectId || !linkId) {
          return { id, success: false };
        }

        try {
          await projectRepository.removeFieldDevice(this.projectId, linkId);
          return { id, success: true };
        } catch {
          return { id, success: false };
        }
      })
    );

    const successCount = results.filter((item) => item.success).length;
    return {
      results,
      total_count: ids.length,
      success_count: successCount,
      failure_count: ids.length - successCount
    };
  }

  private async resolveProjectFieldDeviceLinkId(deviceId: string): Promise<string | undefined> {
    const linkIdsByDeviceId = await this.loadProjectFieldDeviceLinkIds();
    return linkIdsByDeviceId.get(deviceId);
  }

  private async loadProjectFieldDeviceLinkIds(): Promise<Map<string, string>> {
    if (!this.projectId) {
      return new Map();
    }

    const links = await projectRepository.listFieldDevices(this.projectId, {
      page: 1,
      limit: 1000
    });
    return new Map(links.items.map((item) => [item.field_device_id, item.id]));
  }

  private async loadSpecificationDetailsForVisibleDevices(): Promise<void> {
    if (this.loadingSpecifications) return;

    const deviceIds = this.items
      .filter((item) => item.specification_id && !item.specification)
      .map((item) => item.id);

    if (deviceIds.length === 0) {
      return;
    }

    this.loadingSpecifications = true;
    try {
      const updatedItems = await this.mapWithConcurrency(deviceIds, 6, (id) =>
        fieldDeviceRepository.get(id)
      );
      this.replaceItems(updatedItems.map((item) => this.mergeHydratedDevice(item)));
    } catch (error) {
      console.error('Failed to load field device specifications:', error);
    } finally {
      this.loadingSpecifications = false;
    }
  }

  private async loadBacnetObjectsForDevice(deviceId: string): Promise<void> {
    const device = this.items.find((item) => item.id === deviceId);
    if (!device || device.bacnet_objects || this.loadingBacnetRows.has(deviceId)) {
      return;
    }

    const nextLoading = new Set(this.loadingBacnetRows);
    nextLoading.add(deviceId);
    this.loadingBacnetRows = nextLoading;

    try {
      const bacnetObjects = await fieldDeviceRepository.listBacnetObjects(deviceId);
      const currentDevice = this.items.find((item) => item.id === deviceId) ?? device;
      this.replaceItem({ ...currentDevice, bacnet_objects: bacnetObjects });
    } catch (error) {
      console.error('Failed to load BACnet objects:', error);
      addToast('BACnet-Objekte konnten nicht geladen werden.', 'error');
    } finally {
      const nextLoading = new Set(this.loadingBacnetRows);
      nextLoading.delete(deviceId);
      this.loadingBacnetRows = nextLoading;
    }
  }

  private mergeHydratedDevice(updated: FieldDevice): FieldDevice {
    const current = this.items.find((item) => item.id === updated.id);
    if (!current) {
      return updated;
    }

    return {
      ...current,
      ...updated,
      sps_controller_system_type:
        updated.sps_controller_system_type ?? current.sps_controller_system_type,
      apparat: updated.apparat ?? current.apparat,
      system_part: updated.system_part ?? current.system_part,
      bacnet_objects: current.bacnet_objects ?? updated.bacnet_objects
    };
  }

  private async mapWithConcurrency<T, TResult>(
    items: T[],
    concurrency: number,
    worker: (item: T) => Promise<TResult>
  ): Promise<TResult[]> {
    const results = new Array<TResult>(items.length);
    let nextIndex = 0;
    const workerCount = Math.min(Math.max(concurrency, 1), items.length);

    await Promise.all(
      Array.from({ length: workerCount }, async () => {
        while (nextIndex < items.length) {
          const currentIndex = nextIndex;
          nextIndex += 1;
          results[currentIndex] = await worker(items[currentIndex]);
        }
      })
    );

    return results;
  }
}
