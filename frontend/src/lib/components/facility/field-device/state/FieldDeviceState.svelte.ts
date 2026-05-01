import { addToast } from '$lib/components/toast.svelte';
import { useFieldDeviceEditing } from '$lib/hooks/useFieldDeviceEditing.svelte.js';
import { t as translate } from '$lib/i18n/index.js';
import { ManageFieldDeviceUseCase } from '$lib/application/useCases/facility/manageFieldDeviceUseCase.js';
import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
import { canPerform } from '$lib/utils/permissions.js';
import { BaseDataTableState } from '$lib/state/table/BaseDataTableState.svelte.js';
import { sanitizeFilters } from '$lib/state/table/sanitizeFilters.js';
import { createFieldDevicePermissionPolicy } from './fieldDevicePermissionPolicy.js';
import { FieldDeviceGroupingLookupService } from './fieldDeviceGroupingLookupService.js';
import { FieldDeviceLookupService } from './fieldDeviceLookupService.js';
import { ProjectFieldDeviceAssociationService } from './projectFieldDeviceAssociationService.js';
import {
  applySPSControllerNameDelta,
  planSPSControllerDeviceRefresh,
  planVisibleDeviceDelta,
  planVisibleDeviceRefresh
} from './fieldDeviceVisibleRows.js';
import type {
  Apparat,
  Building,
  ControlCabinet,
  FieldDevice,
  SPSController,
  SystemPart
} from '$lib/domain/facility/index.js';
import type {
  FieldDeviceFilters,
  FieldDeviceStateProps,
  SharedFieldDeviceEditorsByDevice
} from './types.js';
import { resolveFieldDeviceFilters, resolvePageSize, toProjectIdResolver } from './types.js';
import { FieldDeviceFetchStrategyFactory } from './strategies/FieldDeviceFetchStrategyFactory.js';
import { FieldDevicePanelState } from './FieldDevicePanelState.svelte.js';
import {
  FieldDeviceTableViewState,
  type FieldDeviceGroupKey
} from './FieldDeviceTableView.svelte.js';

export class FieldDeviceState extends BaseDataTableState<FieldDevice, FieldDeviceFilters> {
  readonly editing: ReturnType<typeof useFieldDeviceEditing>;
  readonly view = new FieldDeviceTableViewState({
    getSPSController: (id) => this.groupingSPSControllers.get(id),
    getControlCabinet: (id) => this.groupingControlCabinets.get(id),
    getBuilding: (id) => this.groupingBuildings.get(id)
  });

  allApparats = $state<Apparat[]>([]);
  allSystemParts = $state<SystemPart[]>([]);
  groupingSPSControllers = $state<Map<string, SPSController>>(new Map());
  groupingControlCabinets = $state<Map<string, ControlCabinet>>(new Map());
  groupingBuildings = $state<Map<string, Building>>(new Map());
  readonly panels = new FieldDevicePanelState();
  loadingSpecifications = $state(false);
  loadingGroupingLookups = $state(false);

  readonly showBulkEditPanel = $derived.by(
    () => this.panels.bulkEditPanelOpen && this.selectedCount > 0
  );
  readonly tableGroups = $derived.by(() => this.view.groupItems(this.items));

  private readonly resolveProjectId: () => string | undefined;
  private readonly fixedFilters: FieldDeviceFilters;
  private readonly resolveSharedFieldDeviceEditors: () => SharedFieldDeviceEditorsByDevice;
  private readonly onFieldDevicesSaved?: (devices: FieldDevice[]) => void;
  private readonly manageFieldDeviceUseCase = new ManageFieldDeviceUseCase(fieldDeviceRepository);
  private readonly lookupService = new FieldDeviceLookupService();
  private readonly groupingLookupService = new FieldDeviceGroupingLookupService();
  private readonly projectAssociationService = new ProjectFieldDeviceAssociationService();
  private readonly permissionPolicy: ReturnType<typeof createFieldDevicePermissionPolicy>;

  constructor(props: FieldDeviceStateProps = {}) {
    const resolveProjectId = toProjectIdResolver(props.projectId);
    const strategyFactory = new FieldDeviceFetchStrategyFactory(resolveProjectId);

    const initialFilters = sanitizeFilters(resolveFieldDeviceFilters(props.initialFilters) ?? {});

    super(strategyFactory.create(), {
      pageSize: resolvePageSize(props.pageSize) ?? 300,
      initialFilters
    });

    this.resolveProjectId = resolveProjectId;
    this.fixedFilters = initialFilters;
    this.resolveSharedFieldDeviceEditors = props.sharedFieldDeviceEditors ?? (() => ({}));
    this.onFieldDevicesSaved = props.onFieldDevicesSaved;
    this.permissionPolicy = createFieldDevicePermissionPolicy({
      isProjectContext: () => this.isProjectContext,
      canPerform,
      canPerformAny: (actions, resource) => actions.some((action) => canPerform(action, resource))
    });
    this.editing = useFieldDeviceEditing({
      projectId: () => this.projectId,
      onSharedStateChange: props.onSharedFieldDeviceStateChange,
      onSaveSuccess: () => undefined
    });
  }

  get projectId() {
    return this.resolveProjectId();
  }

  get effectiveProjectId(): string | undefined {
    return this.projectId ?? this.fixedFilters.projectId ?? this.filters.projectId;
  }

  fixedFilterValue(key: keyof FieldDeviceFilters): string | undefined {
    if (key === 'projectId') {
      return this.projectId ?? this.fixedFilters.projectId;
    }

    return this.fixedFilters[key];
  }

  isFilterFixed(key: keyof FieldDeviceFilters): boolean {
    return Boolean(this.fixedFilterValue(key));
  }

  get isProjectContext(): boolean {
    return Boolean(this.projectId);
  }

  get showMultiCreateForm(): boolean {
    return this.panels.showMultiCreateForm;
  }

  set showMultiCreateForm(value: boolean) {
    this.panels.showMultiCreateForm = value;
  }

  get bulkEditPanelOpen(): boolean {
    return this.panels.bulkEditPanelOpen;
  }

  set bulkEditPanelOpen(value: boolean) {
    this.panels.bulkEditPanelOpen = value;
  }

  get showExportPanel(): boolean {
    return this.panels.showExportPanel;
  }

  set showExportPanel(value: boolean) {
    this.panels.showExportPanel = value;
  }

  get showFilterPanel(): boolean {
    return this.panels.showFilterPanel;
  }

  set showFilterPanel(value: boolean) {
    this.panels.showFilterPanel = value;
  }

  get showSpecifications(): boolean {
    return this.panels.showSpecifications;
  }

  set showSpecifications(value: boolean) {
    this.panels.showSpecifications = value;
  }

  canCreateFieldDevice(): boolean {
    return this.permissionPolicy.canCreateFieldDevice();
  }

  canUpdateFieldDevice(): boolean {
    return this.permissionPolicy.canUpdateFieldDevice();
  }

  canDeleteFieldDevice(): boolean {
    return this.permissionPolicy.canDeleteFieldDevice();
  }

  canUpdateFieldDeviceSpecification(): boolean {
    return this.permissionPolicy.canUpdateFieldDeviceSpecification();
  }

  canUpdateFieldDeviceBacnetObjects(): boolean {
    return this.permissionPolicy.canUpdateFieldDeviceBacnetObjects();
  }

  canOpenBulkEditPanel(): boolean {
    return this.permissionPolicy.canOpenBulkEditPanel();
  }

  canSavePendingEdits(): boolean {
    return this.permissionPolicy.canSavePendingEdits({
      hasUnsavedChanges: this.editing.hasUnsavedChanges,
      hasPendingBaseEdits: this.editing.hasPendingBaseEdits,
      hasPendingSpecificationEdits: this.editing.hasPendingSpecificationEdits,
      hasPendingBacnetEdits: this.editing.hasPendingBacnetEdits
    });
  }

  getEditorsForDevice(deviceId: string) {
    return this.resolveSharedFieldDeviceEditors()[deviceId] ?? [];
  }

  protected override onSelectionChanged() {
    if (this.selectedIds.size === 0) {
      this.panels.closeBulkEditPanel();
    }
  }

  async initialize(): Promise<void> {
    await Promise.all([this.load(), this.loadLookups()]);
  }

  override async load(): Promise<void> {
    await super.load();

    if (this.view.grouping.isGrouped && !this.error && !this.loading) {
      await this.loadGroupingLookupsForVisibleDevices();
    }

    if (!this.showSpecifications || this.error || this.loading) {
      return;
    }

    await this.loadSpecificationDetailsForVisibleDevices();
  }

  private async loadLookups(): Promise<void> {
    const result = await this.lookupService.loadStaticLookups();
    if (result.apparats) this.allApparats = result.apparats;
    if (result.systemParts) this.allSystemParts = result.systemParts;
    if (result.apparatsError) console.error('Failed to load apparats', result.apparatsError);
    if (result.systemPartsError) {
      console.error('Failed to load system parts', result.systemPartsError);
    }
  }

  async applyFilters(filters: FieldDeviceFilters): Promise<void> {
    await this.setFilters({ ...sanitizeFilters(filters), ...this.fixedFilters });
  }

  async clearFilters(): Promise<void> {
    await this.clearAllFilters();
  }

  openMultiCreateForm(): void {
    this.panels.openMultiCreateForm();
  }

  closeMultiCreateForm(): void {
    this.panels.closeMultiCreateForm();
  }

  toggleBulkEditPanel(): void {
    if (this.selectedCount === 0) return;
    if (!this.canOpenBulkEditPanel()) return;
    this.panels.toggleBulkEditPanel();
  }

  toggleExportPanel(): void {
    this.panels.toggleExportPanel();
  }

  toggleFilterPanel(): void {
    this.panels.toggleFilterPanel();
  }

  async toggleGrouping(key: FieldDeviceGroupKey): Promise<void> {
    this.view.grouping.toggle(key);

    if (this.view.grouping.isGrouped) {
      await this.loadGroupingLookupsForVisibleDevices();
    }
  }

  async toggleSpecifications(): Promise<void> {
    const nextShowSpecifications = this.panels.toggleSpecifications();

    if (nextShowSpecifications) {
      await this.loadSpecificationDetailsForVisibleDevices();
    }
  }

  async toggleBacnetExpansion(deviceId: string): Promise<void> {
    if (this.panels.toggleBacnetExpansion(deviceId)) {
      await this.loadBacnetObjectsForDevice(deviceId);
    }
  }

  isBacnetExpanded(deviceId: string): boolean {
    return this.panels.isBacnetExpanded(deviceId);
  }

  isBacnetLoading(deviceId: string): boolean {
    return this.panels.isBacnetLoading(deviceId);
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
    this.panels.closeMultiCreateForm();

    await this.reload();
  }

  async refreshDevices(deviceIds: string[]): Promise<void> {
    const plan = planVisibleDeviceRefresh(this.visibleRowContext(), deviceIds);
    if (plan.action === 'reload') {
      await this.reload();
      return;
    }

    if (plan.action === 'none') {
      return;
    }

    try {
      const updatedItems = await Promise.all(plan.ids.map((id) => fieldDeviceRepository.get(id)));

      this.replaceItems(updatedItems);
      if (this.view.grouping.isGrouped) {
        await this.loadGroupingLookupsForVisibleDevices();
      }
    } catch (error) {
      console.error('Failed to refresh field devices:', error);
      await this.reload();
    }
  }

  async applyDeviceDelta(fieldDevices: FieldDevice[]): Promise<void> {
    const plan = planVisibleDeviceDelta(this.visibleRowContext(), fieldDevices);
    if (plan.action === 'reload') {
      await this.reload();
      return;
    }

    if (plan.action === 'none') {
      return;
    }

    this.replaceItems(plan.devices);
    if (this.view.grouping.isGrouped) {
      await this.loadGroupingLookupsForVisibleDevices();
    }
  }

  async refreshDevicesForSPSControllers(spsControllerIds: string[]): Promise<void> {
    const plan = planSPSControllerDeviceRefresh(this.visibleRowContext(), spsControllerIds);
    if (plan.action === 'reload') {
      await this.reload();
      return;
    }

    if (plan.action === 'none') {
      return;
    }

    await this.refreshDevices(plan.ids);
  }

  applySPSControllerDelta(
    spsControllers: import('$lib/domain/facility/index.js').SPSController[]
  ): void {
    this.items = applySPSControllerNameDelta(this.items, spsControllers);
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

  private visibleRowContext() {
    return {
      items: this.items,
      searchText: this.searchText,
      orderBy: this.orderBy,
      order: this.order,
      hasActiveFilters: this.hasActiveFilters
    };
  }

  private async removeProjectFieldDevice(deviceId: string): Promise<void> {
    if (!this.projectId) {
      throw new Error(translate('projects.errors.load_failed'));
    }

    await this.projectAssociationService.removeFieldDevice(
      this.projectId,
      deviceId,
      translate('projects.errors.load_failed')
    );
  }

  private async removeProjectFieldDevices(ids: string[]): Promise<{
    results: Array<{ id: string; success: boolean }>;
    total_count: number;
    success_count: number;
    failure_count: number;
  }> {
    if (!this.projectId) {
      return {
        results: ids.map((id) => ({ id, success: false })),
        total_count: ids.length,
        success_count: 0,
        failure_count: ids.length
      };
    }

    return this.projectAssociationService.removeFieldDevices(this.projectId, ids);
  }

  private async loadGroupingLookupsForVisibleDevices(): Promise<void> {
    if (this.loadingGroupingLookups) return;

    this.loadingGroupingLookups = true;

    try {
      const lookups = await this.groupingLookupService.loadForVisibleDevices({
        items: this.items,
        activeGroups: new Set(this.view.grouping.activeKeys),
        spsControllers: this.groupingSPSControllers,
        controlCabinets: this.groupingControlCabinets,
        buildings: this.groupingBuildings
      });
      this.groupingSPSControllers = lookups.spsControllers;
      this.groupingControlCabinets = lookups.controlCabinets;
      this.groupingBuildings = lookups.buildings;
    } catch (error) {
      console.error('Failed to load field device grouping lookups:', error);
    } finally {
      this.loadingGroupingLookups = false;
    }
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
    if (!device || device.bacnet_objects || this.panels.isBacnetLoading(deviceId)) {
      return;
    }

    this.panels.markBacnetLoading(deviceId);

    try {
      const bacnetObjects = await fieldDeviceRepository.listBacnetObjects(deviceId);
      const currentDevice = this.items.find((item) => item.id === deviceId) ?? device;
      this.replaceItem({ ...currentDevice, bacnet_objects: bacnetObjects });
    } catch (error) {
      console.error('Failed to load BACnet objects:', error);
      addToast('BACnet-Objekte konnten nicht geladen werden.', 'error');
    } finally {
      this.panels.clearBacnetLoading(deviceId);
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
