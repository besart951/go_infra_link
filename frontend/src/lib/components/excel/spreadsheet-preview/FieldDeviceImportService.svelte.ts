import type { WorksheetPreview } from '$lib/domain/excel/index.js';
import type {
  CreateFieldDeviceRequest,
  SPSControllerSystemType
} from '$lib/domain/facility/index.js';
import { getErrorMessage, getFieldErrors } from '$lib/api/client.js';
import { t as translate } from '$lib/i18n/index.js';
import { alarmTypeRepository } from '$lib/infrastructure/api/alarmTypeRepository.js';
import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
import { notificationClassRepository } from '$lib/infrastructure/api/notificationClassRepository.js';
import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
import { stateTextRepository } from '$lib/infrastructure/api/stateTextRepository.js';
import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';
import { systemTypeRepository } from '$lib/infrastructure/api/systemTypeRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { ListRepository } from '$lib/domain/ports/listRepository.js';
import {
  buildImportCellMarkers,
  type FieldDeviceImportDevicePlan,
  type FieldDeviceImportLookups,
  type FieldDeviceImportPlan,
  type ImportDiagnostic,
  type ImportEntityKind,
  type ImportSourceCell,
  transformWorksheetToFieldDeviceImport
} from './fieldDeviceExportImporter.js';

export interface FieldDeviceImportReport {
  status: 'success' | 'partial' | 'failed';
  createdControlCabinetId?: string;
  createdSpsControllerId?: string;
  createdControlCabinetLabel?: string;
  createdSpsControllerLabel?: string;
  createdFieldDevices: number;
  failedFieldDevices: number;
  errorMessage?: string;
}

const PAGE_SIZE = 500;

export class FieldDeviceImportService {
  plan = $state<FieldDeviceImportPlan | null>(null);
  isTransforming = $state(false);
  isImporting = $state(false);
  transformError = $state<string | null>(null);
  importReport = $state<FieldDeviceImportReport | null>(null);
  backendDiagnostics = $state<ImportDiagnostic[]>([]);

  allDiagnostics = $derived.by(() => [
    ...(this.plan?.diagnostics ?? []),
    ...this.backendDiagnostics
  ]);

  cellMarkers = $derived.by(() => buildImportCellMarkers(this.allDiagnostics));
  canImport = $derived.by(
    () =>
      Boolean(this.plan?.canImport) &&
      !this.isImporting &&
      !this.importReport?.createdControlCabinetId &&
      !this.importReport?.createdSpsControllerId
  );

  private lookups: FieldDeviceImportLookups | null = null;
  private lookupLoad: Promise<FieldDeviceImportLookups> | null = null;

  clearTransform(): void {
    this.plan = null;
    this.transformError = null;
    this.importReport = null;
    this.backendDiagnostics = [];
  }

  async transform(worksheet: WorksheetPreview | null): Promise<void> {
    if (!worksheet || this.isTransforming) return;

    this.isTransforming = true;
    this.transformError = null;
    this.importReport = null;
    this.backendDiagnostics = [];

    try {
      this.lookups = null;
      this.lookupLoad = null;
      const lookups = await this.loadLookups();
      this.plan = transformWorksheetToFieldDeviceImport(worksheet, lookups);
    } catch (error) {
      this.plan = null;
      this.transformError =
        error instanceof Error
          ? error.message
          : translate('field_device.importer.errors.transform_prepare_failed');
    } finally {
      this.isTransforming = false;
    }
  }

  async importPlan(): Promise<void> {
    if (!this.plan || !this.plan.canImport || this.isImporting) return;

    this.isImporting = true;
    this.importReport = null;
    this.backendDiagnostics = [];

    const plan = this.plan;
    let currentStep: {
      entity: ImportEntityKind;
      cell: ImportSourceCell | undefined;
      entityKey?: string;
    } = {
      entity: 'control_cabinet',
      cell: plan.controller.controlCabinetSourceCell
    };

    let createdControlCabinetId: string | undefined;
    let createdSpsControllerId: string | undefined;

    try {
      if (!plan.controller.controlCabinetRequest || !plan.controller.spsControllerRequest) {
        throw new Error(translate('field_device.importer.errors.incomplete_plan'));
      }

      currentStep = {
        entity: 'control_cabinet',
        cell: plan.controller.controlCabinetSourceCell
      };
      const controlCabinet = await controlCabinetRepository.create(
        plan.controller.controlCabinetRequest
      );
      createdControlCabinetId = controlCabinet.id;

      currentStep = {
        entity: 'sps_controller',
        cell: plan.controller.spsControllerSourceCell
      };
      const spsController = await spsControllerRepository.create({
        ...plan.controller.spsControllerRequest,
        control_cabinet_id: controlCabinet.id
      });
      createdSpsControllerId = spsController.id;

      currentStep = {
        entity: 'sps_controller_system_type',
        cell: plan.controller.systemTypes[0]?.sourceCell
      };
      const systemTypeMap = await this.fetchCreatedSystemTypeMap(spsController.id, plan);
      const missingSystemTypes = plan.controller.systemTypes.filter(
        (item) => !systemTypeMap.has(item.key)
      );
      if (missingSystemTypes.length > 0) {
        for (const item of missingSystemTypes) {
          this.addBackendDiagnostic(
            translate('field_device.importer.errors.created_system_type_missing'),
            'sps_controller_system_type',
            item.sourceCell,
            item.key
          );
        }
        throw new Error(translate('field_device.importer.errors.system_types_unresolved'));
      }

      currentStep = {
        entity: 'field_device',
        cell: plan.controller.fieldDevices[0]?.sourceCell
      };
      const fieldDeviceRequests = plan.controller.fieldDevices.map((fieldDevice) =>
        toCreateFieldDeviceRequest(fieldDevice, systemTypeMap)
      );
      const response = await fieldDeviceRepository.multiCreate({
        field_devices: fieldDeviceRequests
      });

      for (const result of response.results) {
        if (result.success) continue;
        const fieldDevice = plan.controller.fieldDevices[result.index];
        this.addBackendDiagnostic(
          result.error || translate('field_device.importer.errors.field_device_create_failed'),
          'field_device',
          fieldDevice ? cellForFieldDeviceError(fieldDevice, result.error_field) : undefined,
          fieldDevice?.key
        );
      }

      const hasFailures = response.failure_count > 0;
      this.importReport = {
        status: hasFailures ? 'partial' : 'success',
        createdControlCabinetId,
        createdSpsControllerId,
        createdControlCabinetLabel:
          controlCabinet.control_cabinet_nr || plan.controller.controlCabinetNr,
        createdSpsControllerLabel:
          spsController.device_name ||
          plan.controller.spsControllerRequest.device_name ||
          plan.controller.spsControllerRequest.ga_device,
        createdFieldDevices: response.success_count,
        failedFieldDevices: response.failure_count
      };
    } catch (error) {
      if (this.backendDiagnostics.length === 0) {
        this.addApiDiagnostics(error, currentStep.entity, currentStep.cell, currentStep.entityKey);
      }

      this.importReport = {
        status: 'failed',
        createdControlCabinetId,
        createdSpsControllerId,
        createdControlCabinetLabel: plan.controller.controlCabinetNr,
        createdSpsControllerLabel:
          plan.controller.spsControllerRequest?.device_name ||
          plan.controller.spsControllerRequest?.ga_device,
        createdFieldDevices: 0,
        failedFieldDevices: plan.controller.fieldDevices.length,
        errorMessage: getErrorMessage(error)
      };
    } finally {
      this.lookups = null;
      this.isImporting = false;
    }
  }

  private async fetchCreatedSystemTypeMap(
    spsControllerId: string,
    plan: FieldDeviceImportPlan
  ): Promise<Map<string, string>> {
    const items: SPSControllerSystemType[] = [];
    let page = 1;

    while (true) {
      const response = await spsControllerRepository.listSystemTypes({
        page,
        limit: PAGE_SIZE,
        sps_controller_id: spsControllerId
      });
      items.push(...response.items);
      if (response.page >= response.total_pages) break;
      page += 1;
    }

    const wanted = new Map(
      plan.controller.systemTypes.map((item) => [`${item.systemTypeId}|${item.number}`, item.key])
    );
    const out = new Map<string, string>();

    for (const item of items) {
      if (item.number === undefined || item.number === null) continue;
      const wantedKey = wanted.get(`${item.system_type_id}|${item.number}`);
      if (wantedKey) {
        out.set(wantedKey, item.id);
      }
    }

    return out;
  }

  private async loadLookups(): Promise<FieldDeviceImportLookups> {
    if (this.lookups) return this.lookups;
    if (this.lookupLoad) return this.lookupLoad;

    this.lookupLoad = Promise.all([
      fetchAllListPages(buildingRepository),
      fetchAllListPages(controlCabinetRepository),
      fetchAllListPages(fieldDeviceRepository),
      fetchAllListPages(systemTypeRepository),
      fetchAllListPages(systemPartRepository),
      fetchAllListPages(apparatRepository),
      fetchAllListPages(stateTextRepository),
      fetchAllListPages(notificationClassRepository),
      fetchAllAlarmTypes()
    ]).then(
      ([
        buildings,
        controlCabinets,
        fieldDevices,
        systemTypes,
        systemParts,
        apparats,
        stateTexts,
        notificationClasses,
        alarmTypes
      ]) => {
        this.lookups = {
          buildings,
          controlCabinets,
          fieldDevices,
          systemTypes,
          systemParts,
          apparats,
          stateTexts,
          notificationClasses,
          alarmTypes
        };
        return this.lookups;
      }
    );

    try {
      return await this.lookupLoad;
    } finally {
      this.lookupLoad = null;
    }
  }

  private addApiDiagnostics(
    error: unknown,
    entity: ImportEntityKind,
    fallbackCell?: ImportSourceCell,
    entityKey?: string
  ): void {
    const fieldErrors = getFieldErrors(error);
    const entries = Object.entries(fieldErrors);

    if (entries.length === 0) {
      this.addBackendDiagnostic(getErrorMessage(error), entity, fallbackCell, entityKey);
      return;
    }

    for (const [, message] of entries) {
      this.addBackendDiagnostic(message, entity, fallbackCell, entityKey);
    }
  }

  private addBackendDiagnostic(
    message: string,
    entity: ImportEntityKind,
    cell?: ImportSourceCell,
    entityKey?: string
  ): void {
    this.backendDiagnostics = [
      ...this.backendDiagnostics,
      {
        id: `backend:${this.backendDiagnostics.length + 1}`,
        severity: 'error',
        message,
        entity,
        cell,
        entityKey
      }
    ];
  }
}

function toCreateFieldDeviceRequest(
  fieldDevice: FieldDeviceImportDevicePlan,
  systemTypeMap: Map<string, string>
): CreateFieldDeviceRequest {
  const { spsControllerSystemTypeKey, ...request } = fieldDevice.request;
  return {
    ...request,
    sps_controller_system_type_id: systemTypeMap.get(spsControllerSystemTypeKey) ?? ''
  };
}

function cellForFieldDeviceError(
  fieldDevice: FieldDeviceImportDevicePlan,
  errorField: string
): ImportSourceCell {
  const field = (errorField || '').toLowerCase();
  const bacnetMatch = field.match(/bacnet_objects\.(\d+)\.([a-z_]+)/);
  if (bacnetMatch) {
    const bacnet = fieldDevice.bacnetObjects[Number.parseInt(bacnetMatch[1], 10)];
    const bacnetField = bacnetMatch[2] ?? '';
    if (bacnet) {
      if (bacnetField.includes('text_fix')) return bacnet.sourceCells.textFix;
      if (bacnetField.includes('software')) return bacnet.sourceCells.address;
      return bacnet.sourceCell;
    }
  }

  if (field.includes('bmk')) return fieldDevice.sourceCells.bmk;
  if (field.includes('description')) return fieldDevice.sourceCells.description;
  if (field.includes('apparat_nr')) return fieldDevice.sourceCells.apparatNr;
  if (field.includes('system_part')) return fieldDevice.sourceCells.systemPart;
  if (field.includes('apparat')) return fieldDevice.sourceCells.apparat;
  if (field.includes('sps_controller_system_type')) {
    return fieldDevice.sourceCells.spsControllerSystemType;
  }
  return fieldDevice.sourceCell;
}

async function fetchAllListPages<T>(repository: ListRepository<T>): Promise<T[]> {
  return fetchAllPages(async (page, pageSize) => {
    const response: PaginatedResponse<T> = await repository.list(listParams(page, pageSize));
    return {
      items: response.items,
      page: response.metadata.page,
      totalPages: response.metadata.totalPages
    };
  });
}

async function fetchAllAlarmTypes() {
  return fetchAllPages(async (page, pageSize) => {
    const response = await alarmTypeRepository.list({ page, pageSize, search: '' });
    return {
      items: response.items,
      page: response.page,
      totalPages: response.totalPages
    };
  });
}

async function fetchAllPages<T>(
  fetchPage: (
    page: number,
    pageSize: number
  ) => Promise<{ items: T[]; page: number; totalPages: number }>
): Promise<T[]> {
  const items: T[] = [];
  let page = 1;

  while (true) {
    const response = await fetchPage(page, PAGE_SIZE);
    items.push(...response.items);
    if (response.page >= response.totalPages) break;
    page += 1;
  }

  return items;
}

function listParams(page: number, pageSize: number): ListParams {
  return {
    pagination: { page, pageSize },
    search: { text: '' }
  };
}
