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
import { ManageObjectDataUseCase } from '$lib/application/useCases/facility/manageObjectDataUseCase.js';
import { ListSPSControllersUseCase } from '$lib/application/useCases/facility/listSPSControllersUseCase.js';
import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
import { objectDataRepository } from '$lib/infrastructure/api/objectDataRepository.js';
import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';

import type {
  FieldDevice,
  ObjectData,
  SPSControllerSystemType,
  CreateFieldDeviceRequest,
  MultiCreateFieldDeviceResponse
} from '$lib/domain/facility/index.js';
import type { FieldDevicePreselection as PreselectionType } from '$lib/domain/facility/preselectionFilter.js';

export interface FieldDeviceMultiCreateStateOptions {
  projectId?: () => string | undefined;
  onSuccess?: () => ((createdDevices: FieldDevice[]) => void) | undefined;
}

export class FieldDeviceMultiCreateState {
  selection = $state<MultiCreateSelection>({
    spsControllerSystemTypeId: '',
    objectDataId: '',
    apparatId: '',
    systemPartId: ''
  });

  preselectionValue = $state<PreselectionType>({
    objectDataId: '',
    apparatId: '',
    systemPartId: ''
  });

  rows = $state<FieldDeviceRowData[]>([]);
  rowErrors = $state<Map<number, FieldDeviceRowError>>(new Map());
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
  private readonly manageObjectDataUseCase = new ManageObjectDataUseCase(objectDataRepository);
  private readonly spsUseCase = new ListSPSControllersUseCase(spsControllerRepository);

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

    this.selection = {
      spsControllerSystemTypeId: value,
      objectDataId: '',
      apparatId: '',
      systemPartId: ''
    };
    this.preselectionValue = {
      objectDataId: '',
      apparatId: '',
      systemPartId: ''
    };
    this.rows = [];
    this.availableNumbers = [];
    this.selectedObjectData = null;
    this.loadingObjectDataPreview = false;
    this.objectDataPreviewError = '';
    this.rowErrors = new Map();
  }

  handlePreselectionChange(next: PreselectionType): void {
    this.preselectionValue = next;
    this.selection = {
      ...this.selection,
      objectDataId: next.objectDataId,
      apparatId: next.apparatId,
      systemPartId: next.systemPartId
    };
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
      const fieldDevices = this.buildCreatePayload();
      const rawResponse = await this.manageUseCase.multiCreate({ field_devices: fieldDevices });
      const response = this.normalizeMultiCreateResponse(rawResponse);

      const successfulIndices = new Set<number>();
      const backendErrors = new Map<number, FieldDeviceRowError>();

      for (const result of response.results) {
        if (result.index < 0 || result.index >= this.rows.length) {
          continue;
        }

        if (result.success) {
          successfulIndices.add(result.index);
          continue;
        }

        backendErrors.set(result.index, {
          message: localizeErrorText(result.error, result.error_field),
          field: (result.error_field as FieldDeviceRowError['field']) || ''
        });
      }

      const indexMap = new Map<number, number>();
      const remainingRows: FieldDeviceRowData[] = [];

      for (let index = 0; index < this.rows.length; index += 1) {
        const row = this.rows[index];
        if (!successfulIndices.has(index)) {
          indexMap.set(index, remainingRows.length);
          remainingRows.push(row);
        }
      }

      const remappedErrors = new Map<number, FieldDeviceRowError>();
      for (const entry of backendErrors.entries()) {
        const originalIndex = entry[0];
        const error = entry[1];
        const nextIndex = indexMap.get(originalIndex);

        if (nextIndex !== undefined) {
          remappedErrors.set(nextIndex, error);
        }
      }

      this.rows = remainingRows;
      this.rowErrors = remappedErrors;

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
      this.rowErrors = new Map();

      const onSuccess = this.resolveOnSuccess();
      if (onSuccess) {
        onSuccess(this.collectCreatedDevices(response));
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
    this.preselectionValue = {
      objectDataId: persisted.selection.objectDataId,
      apparatId: persisted.selection.apparatId,
      systemPartId: persisted.selection.systemPartId
    };
  }

  private setupEffects(): void {
    $effect(() => {
      const newKey = createSelectionKey(this.selection);
      const keyChanged = this.selectionKey !== '' && newKey !== this.selectionKey;

      if (keyChanged) {
        this.availableNumbers = [];
        this.rows = [];
        this.rowErrors = new Map();
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
      const response = await this.manageUseCase.getAvailableApparatNumbers(
        this.selection.spsControllerSystemTypeId,
        this.selection.apparatId,
        this.selection.systemPartId,
        controller.signal
      );

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
      const objectData = await this.manageObjectDataUseCase.get(objectDataId, controller.signal);

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
    const usedNumbers = getUsedApparatNumbers(this.rows);
    let changed = false;

    for (const row of this.rows) {
      if (row.apparatNr !== null) {
        continue;
      }

      let nextAvailable: number | undefined;
      for (const candidate of this.availableNumbers) {
        if (!usedNumbers.has(candidate)) {
          nextAvailable = candidate;
          break;
        }
      }

      if (nextAvailable !== undefined) {
        row.apparatNr = nextAvailable;
        usedNumbers.add(nextAvailable);
        changed = true;
      }
    }

    if (changed) {
      this.rows = [...this.rows];
      this.revalidateAllRows();
    }
  }

  private revalidateAllRows(): void {
    this.rowErrors = validateAllRows(this.rows, this.availableNumbers, false);
  }

  private buildCreatePayload(): CreateFieldDeviceRequest[] {
    const fieldDevices: CreateFieldDeviceRequest[] = [];

    for (const row of this.rows) {
      fieldDevices.push({
        bmk: row.bmk || undefined,
        description: row.description || undefined,
        text_fix: row.textFix || undefined,
        apparat_nr: row.apparatNr!,
        sps_controller_system_type_id: this.selection.spsControllerSystemTypeId,
        system_part_id: this.selection.systemPartId,
        apparat_id: this.selection.apparatId,
        object_data_id: this.selection.objectDataId || undefined
      });
    }

    return fieldDevices;
  }

  private normalizeMultiCreateResponse(
    response: MultiCreateFieldDeviceResponse | { preview?: MultiCreateFieldDeviceResponse }
  ): MultiCreateFieldDeviceResponse {
    if (this.isMultiCreateFieldDeviceResponse(response)) {
      return response;
    }

    if (
      'preview' in response &&
      response.preview &&
      this.isMultiCreateFieldDeviceResponse(response.preview)
    ) {
      return response.preview;
    }

    throw new Error(translate('field_device.multi_create.errors.create'));
  }

  private isMultiCreateFieldDeviceResponse(
    value: unknown
  ): value is MultiCreateFieldDeviceResponse {
    if (!value || typeof value !== 'object') {
      return false;
    }

    const candidate = value as Record<string, unknown>;

    return (
      Array.isArray(candidate.results) &&
      typeof candidate.total_requests === 'number' &&
      typeof candidate.success_count === 'number' &&
      typeof candidate.failure_count === 'number'
    );
  }

  private collectCreatedDevices(response: MultiCreateFieldDeviceResponse): FieldDevice[] {
    const createdDevices: FieldDevice[] = [];

    for (const result of response.results) {
      if (result.success && result.field_device) {
        createdDevices.push(result.field_device);
      }
    }

    return createdDevices;
  }
}
