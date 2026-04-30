import { addToast } from '$lib/components/toast.svelte';
import { ApiException, localizeErrorText } from '$lib/api/client.js';
import { t as translate } from '$lib/i18n/index.js';
import {
  type FieldDeviceRowData,
  type FieldDeviceRowError,
  type MultiCreateSelection,
  createSelectionKey,
  hasRequiredSelections,
  canFetchAvailableNumbers,
  createNewRow,
  getUsedApparatNumbers,
  validateAllRows,
  loadPersistedState,
  savePersistedState,
  clearPersistedState
} from '$lib/domain/facility/fieldDeviceMultiCreate.js';
import { ManageFieldDeviceUseCase } from '$lib/application/useCases/facility/manageFieldDeviceUseCase.js';
import { ListSPSControllersUseCase } from '$lib/application/useCases/facility/listSPSControllersUseCase.js';
import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
import {
  applyPreselectionToSelection,
  createEmptyFieldDevicePreselection,
  createEmptyMultiCreateSelection,
  preselectionFromSelection,
  resetSelectionForSpsSystemType
} from './multiCreateSelection.js';
import {
  buildMultiCreatePayload,
  collectCreatedDevices,
  normalizeMultiCreateResponse,
  reconcileMultiCreateRows
} from './multiCreateSubmission.js';
import {
  MultiCreateAvailableNumbersService,
  autoAssignApparatNumbers
} from './multiCreateAvailableNumbersService.js';
import { MultiCreateObjectDataPreviewService } from './multiCreateObjectDataPreviewService.js';

import type {
  FieldDevice,
  ObjectData,
  SPSControllerSystemType
} from '$lib/domain/facility/index.js';
import type { FieldDevicePreselection as PreselectionType } from '$lib/domain/facility/preselectionFilter.js';

export interface FieldDeviceMultiCreateStateOptions {
  projectId?: () => string | undefined;
  onSuccess?: () => ((createdDevices: FieldDevice[]) => void) | undefined;
}

function createRowErrorMap(): Map<number, FieldDeviceRowError> {
  return new Map<number, FieldDeviceRowError>();
}

export class FieldDeviceMultiCreateState {
  selection = $state<MultiCreateSelection>(createEmptyMultiCreateSelection());

  preselectionValue = $state<PreselectionType>(createEmptyFieldDevicePreselection());

  rows = $state<FieldDeviceRowData[]>([]);
  rowErrors = $state<Map<number, FieldDeviceRowError>>(createRowErrorMap());
  availableNumbers = $state<number[]>([]);
  loadingAvailableNumbers = $state(false);
  selectedObjectData = $state<ObjectData | null>(null);
  loadingObjectDataPreview = $state(false);
  objectDataPreviewError = $state('');
  submitting = $state(false);
  globalError = $state('');
  projectOnlyOverride = $state<boolean | null>(null);
  selectionKey = $state('');

  private availableNumbersAbortController: AbortController | null = null;
  private objectDataPreviewAbortController: AbortController | null = null;

  private readonly resolveProjectId: () => string | undefined;
  private readonly resolveOnSuccess: () => ((createdDevices: FieldDevice[]) => void) | undefined;
  private readonly manageUseCase = new ManageFieldDeviceUseCase(fieldDeviceRepository);
  private readonly spsUseCase = new ListSPSControllersUseCase(spsControllerRepository);
  private readonly availableNumbersService = new MultiCreateAvailableNumbersService();
  private readonly objectDataPreviewService = new MultiCreateObjectDataPreviewService();

  constructor(options: FieldDeviceMultiCreateStateOptions = {}) {
    this.resolveProjectId = options.projectId ?? this.resolveUndefinedProjectId;
    this.resolveOnSuccess = options.onSuccess ?? this.resolveUndefinedOnSuccess;

    this.restorePersistedState();
    this.setupEffects();
  }

  get projectId(): string | undefined {
    return this.resolveProjectId();
  }

  get projectOnly(): boolean {
    if (!this.projectId) {
      return false;
    }

    return this.projectOnlyOverride ?? true;
  }

  get showConfiguration(): boolean {
    return hasRequiredSelections(this.selection);
  }

  get showRowsSection(): boolean {
    return this.showConfiguration;
  }

  get canAddRow(): boolean {
    return (
      this.showConfiguration &&
      this.availableNumbers.length > this.rows.length &&
      !this.loadingAvailableNumbers
    );
  }

  get hasValidationErrors(): boolean {
    return this.rowErrors.size > 0;
  }

  destroy(): void {
    this.availableNumbersAbortController?.abort();
    this.objectDataPreviewAbortController?.abort();

    if (this.rows.length > 0 || this.selection.spsControllerSystemTypeId) {
      savePersistedState(this.selection, this.rows);
    }
  }

  handleProjectOnlyChange(checked: boolean): void {
    this.projectOnlyOverride = checked;
    this.handleSpsSystemTypeChange('');
  }

  handleSpsSystemTypeChange(value: string): void {
    if (value === this.selection.spsControllerSystemTypeId) {
      return;
    }

    this.availableNumbersAbortController?.abort();
    this.objectDataPreviewAbortController?.abort();

    this.selection = resetSelectionForSpsSystemType(value);
    this.preselectionValue = createEmptyFieldDevicePreselection();
    this.rows = [];
    this.availableNumbers = [];
    this.selectedObjectData = null;
    this.loadingObjectDataPreview = false;
    this.objectDataPreviewError = '';
    this.rowErrors = createRowErrorMap();
  }

  handlePreselectionChange(next: PreselectionType): void {
    this.preselectionValue = next;
    this.selection = applyPreselectionToSelection(this.selection, next);
  }

  addRow(): void {
    if (!this.canAddRow) {
      return;
    }

    const usedNumbers = getUsedApparatNumbers(this.rows);
    const newRow = createNewRow(this.availableNumbers, usedNumbers);

    if (!newRow) {
      addToast(translate('field_device.multi_create.errors.no_more_numbers'), 'warning');
      return;
    }

    this.rows = [...this.rows, newRow];
  }

  removeRow(index: number): void {
    this.rows = this.rows.filter(function (_, currentIndex) {
      return currentIndex !== index;
    });
    this.revalidateAllRows();
  }

  handleRowBmkChange(index: number, value: string): void {
    const row = this.rows[index];
    if (!row) {
      return;
    }

    row.bmk = value;
    this.rows = [...this.rows];
  }

  handleRowDescriptionChange(index: number, value: string): void {
    const row = this.rows[index];
    if (!row) {
      return;
    }

    row.description = value;
    this.rows = [...this.rows];
  }

  handleRowTextFixChange(index: number, value: string): void {
    const row = this.rows[index];
    if (!row) {
      return;
    }

    row.textFix = value;
    this.rows = [...this.rows];
  }

  handleRowApparatNrChange(index: number, value: string): void {
    const row = this.rows[index];
    if (!row) {
      return;
    }

    const trimmed = value.trim();

    if (!trimmed) {
      row.apparatNr = null;
    } else {
      const parsed = Number.parseInt(trimmed, 10);
      row.apparatNr = Number.isNaN(parsed) ? null : parsed;
    }

    this.rows = [...this.rows];
    this.revalidateAllRows();
  }

  getPlaceholderForRow(index: number): string {
    const usedNumbers: number[] = [];

    for (let i = 0; i < this.rows.length; i += 1) {
      if (i === index) {
        continue;
      }

      const apparatNr = this.rows[i].apparatNr;
      if (apparatNr !== null) {
        usedNumbers.push(apparatNr);
      }
    }

    for (const number of this.availableNumbers) {
      if (!usedNumbers.includes(number)) {
        return translate('field_device.multi_create.next_available', { value: number });
      }
    }

    return '';
  }

  async fetchSpsControllerSystemTypes(search: string): Promise<SPSControllerSystemType[]> {
    const params: { search: string; limit: number; project_id?: string } = { search, limit: 50 };

    if (this.projectOnly && this.projectId) {
      params.project_id = this.projectId;
    }

    const result = await this.spsUseCase.listSystemTypes(params);
    return result.items || [];
  }

  async fetchSpsControllerSystemTypeById(id: string): Promise<SPSControllerSystemType | null> {
    try {
      return await this.spsUseCase.getSystemType(id);
    } catch (error) {
      if (error instanceof ApiException && error.status === 404) {
        if (this.selection.spsControllerSystemTypeId === id) {
          this.handleSpsSystemTypeChange('');
        }

        return null;
      }

      throw error;
    }
  }

  formatSpsControllerSystemTypeLabel(item: SPSControllerSystemType): string {
    const deviceName = item.sps_controller_name ?? '';
    const number = item.number ?? '';
    const documentName = item.document_name ?? '';
    const sysTypePart = number || documentName ? `${number}_${documentName}` : '';

    if (deviceName && sysTypePart) {
      return `${deviceName}_${sysTypePart}`;
    }

    return deviceName || sysTypePart || '';
  }

  async handleSubmit(): Promise<void> {
    if (this.rows.length === 0) {
      addToast(translate('field_device.multi_create.errors.no_rows'), 'warning');
      return;
    }

    const errors = validateAllRows(this.rows, this.availableNumbers, true);
    this.rowErrors = errors;

    if (errors.size > 0) {
      addToast(translate('field_device.multi_create.errors.validation'), 'error');
      return;
    }

    this.submitting = true;
    this.globalError = '';

    try {
      const fieldDevices = buildMultiCreatePayload(this.rows, this.selection);
      const rawResponse = this.projectId
        ? await projectRepository.createFieldDevices(this.projectId, {
            field_devices: fieldDevices
          })
        : await this.manageUseCase.multiCreate({ field_devices: fieldDevices });
      const response = normalizeMultiCreateResponse(
        rawResponse,
        translate('field_device.multi_create.errors.create')
      );
      const reconciledRows = reconcileMultiCreateRows(this.rows, response, localizeErrorText);

      this.rows = reconciledRows.remainingRows;
      this.rowErrors = reconciledRows.rowErrors;

      if (response.failure_count > 0) {
        addToast(
          translate('field_device.multi_create.toasts.partial_created', {
            success: response.success_count,
            total: response.total_requests,
            failed: response.failure_count
          }),
          'warning'
        );
        return;
      }

      addToast(
        translate('field_device.multi_create.toasts.created', {
          count: response.success_count
        }),
        'success'
      );

      clearPersistedState();
      this.rows = [];
      this.rowErrors = createRowErrorMap();

      const onSuccess = this.resolveOnSuccess();
      if (onSuccess) {
        onSuccess(collectCreatedDevices(response));
      }
    } catch (error) {
      this.globalError =
        error instanceof Error
          ? error.message
          : translate('field_device.multi_create.errors.create');
      addToast(
        translate('field_device.multi_create.errors.create_with_message', {
          message: this.globalError
        }),
        'error'
      );
    } finally {
      this.submitting = false;
    }
  }

  private resolveUndefinedProjectId(): string | undefined {
    return undefined;
  }

  private resolveUndefinedOnSuccess(): ((createdDevices: FieldDevice[]) => void) | undefined {
    return undefined;
  }

  private restorePersistedState(): void {
    const persisted = loadPersistedState();
    if (!persisted) {
      return;
    }

    this.selection = persisted.selection;
    this.rows = persisted.rows;
    this.preselectionValue = preselectionFromSelection(persisted.selection);
  }

  private setupEffects(): void {
    $effect(() => {
      const newKey = createSelectionKey(this.selection);
      const keyChanged = this.selectionKey !== '' && newKey !== this.selectionKey;

      if (keyChanged) {
        this.availableNumbers = [];
        this.rows = [];
        this.rowErrors = createRowErrorMap();
      }

      this.selectionKey = newKey;

      if (canFetchAvailableNumbers(this.selection)) {
        void this.fetchAvailableNumbers();
      }
    });

    $effect(() => {
      if (this.availableNumbers.length > 0 && this.rows.length > 0) {
        this.revalidateAllRows();
      }
    });

    $effect(() => {
      const objectDataId = this.selection.objectDataId?.trim() ?? '';

      if (!objectDataId) {
        this.objectDataPreviewAbortController?.abort();
        this.selectedObjectData = null;
        this.loadingObjectDataPreview = false;
        this.objectDataPreviewError = '';
        return;
      }

      void this.fetchSelectedObjectDataPreview(objectDataId);
    });
  }

  private async fetchAvailableNumbers(): Promise<void> {
    if (!canFetchAvailableNumbers(this.selection)) {
      this.availableNumbers = [];
      return;
    }

    this.availableNumbersAbortController?.abort();

    this.loadingAvailableNumbers = true;
    const controller = new AbortController();
    this.availableNumbersAbortController = controller;

    try {
      const response = await this.availableNumbersService.fetch(this.selection, controller.signal);

      if (!controller.signal.aborted) {
        this.availableNumbers = response.available;
        this.autoAssignApparatNumbers();
      }
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') {
        return;
      }

      const message =
        error instanceof ApiException
          ? translate('field_device.multi_create.errors.fetch_numbers_status', {
              status: error.status,
              message: error.message
            })
          : translate('field_device.multi_create.errors.fetch_numbers_failed', {
              message: error instanceof Error ? error.message : String(error)
            });

      addToast(message, 'error');
      this.availableNumbers = [];
    } finally {
      if (this.availableNumbersAbortController === controller) {
        this.loadingAvailableNumbers = false;
      }
    }
  }

  private async fetchSelectedObjectDataPreview(objectDataId: string): Promise<void> {
    this.objectDataPreviewAbortController?.abort();

    const controller = new AbortController();
    this.objectDataPreviewAbortController = controller;

    this.loadingObjectDataPreview = true;
    this.objectDataPreviewError = '';

    try {
      const objectData = await this.objectDataPreviewService.fetch(objectDataId, controller.signal);

      if (!controller.signal.aborted) {
        this.selectedObjectData = objectData;
      }
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') {
        return;
      }

      if (!controller.signal.aborted) {
        this.selectedObjectData = null;
        this.objectDataPreviewError = translate(
          'field_device.multi_create.object_data_preview.load_failed'
        );
      }
    } finally {
      if (this.objectDataPreviewAbortController === controller) {
        this.loadingObjectDataPreview = false;
      }
    }
  }

  private autoAssignApparatNumbers(): void {
    const { rows, changed } = autoAssignApparatNumbers(this.rows, this.availableNumbers);
    if (changed) {
      this.rows = rows;
      this.revalidateAllRows();
    }
  }

  private revalidateAllRows(): void {
    this.rowErrors = validateAllRows(this.rows, this.availableNumbers, false);
  }
}
