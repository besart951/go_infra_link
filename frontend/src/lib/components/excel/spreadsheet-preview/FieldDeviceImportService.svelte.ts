import type { WorksheetPreview } from '$lib/domain/excel/index.js';
import type {
  AlarmType,
  Apparat,
  Building,
  ControlCabinet,
  CreateNotificationClassRequest,
  CreateControlCabinetRequest,
  CreateFieldDeviceRequest,
  CreateSPSControllerRequest,
  FieldDevice,
  MultiCreateFieldDeviceRequest,
  MultiCreateFieldDeviceResponse,
  NotificationClass,
  SPSController,
  SPSControllerSystemTypeInput,
  SPSControllerSystemType,
  StateText,
  SystemPart,
  UpdateSPSControllerRequest,
  UpdateFieldDeviceRequest
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
  buildImportRowMarkers,
  collectFieldDeviceImportLookupCriteria,
  type FieldDeviceImportDevicePlan,
  type FieldDeviceImportLookupCriteria,
  type FieldDeviceImportLookups,
  type FieldDeviceImportPlan,
  type ImportDiagnostic,
  type ImportEntityKind,
  type ImportNotificationClassResolution,
  type ImportSourceCell,
  normalizeLookupKey,
  transformWorksheetToFieldDeviceImport
} from './fieldDeviceExportImporter.js';

export const CONTROL_CABINET_IMPORT_NODE_KEY = 'control-cabinet';
export const SPS_CONTROLLER_IMPORT_NODE_KEY = 'sps-controller';

export type ImportNodeStatus = 'pending' | 'identical' | 'delta' | 'success' | 'failed';

export interface ImportNodeState {
  entity: ImportEntityKind;
  status: ImportNodeStatus;
  id?: string;
  message?: string;
}

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

export interface FieldDeviceImportBackend {
  loadLookups(criteria: FieldDeviceImportLookupCriteria): Promise<FieldDeviceImportLookups>;
  getControlCabinet(id: string): Promise<ControlCabinet>;
  createControlCabinet(data: CreateControlCabinetRequest): Promise<ControlCabinet>;
  getSpsController(id: string): Promise<SPSController>;
  createSpsController(data: CreateSPSControllerRequest): Promise<SPSController>;
  updateSpsController(id: string, data: UpdateSPSControllerRequest): Promise<SPSController>;
  createNotificationClass(data: CreateNotificationClassRequest): Promise<NotificationClass>;
  multiCreateFieldDevices(
    data: MultiCreateFieldDeviceRequest
  ): Promise<MultiCreateFieldDeviceResponse>;
  updateFieldDevice(id: string, data: UpdateFieldDeviceRequest): Promise<FieldDevice>;
}

export interface FieldDeviceImportServiceDeps {
  backend?: Partial<FieldDeviceImportBackend>;
}

const defaultFieldDeviceImportBackend: FieldDeviceImportBackend = {
  loadLookups: loadTargetedLookups,
  getControlCabinet: (id) => controlCabinetRepository.get(id),
  createControlCabinet: (data) => controlCabinetRepository.create(data),
  getSpsController: (id) => spsControllerRepository.get(id),
  createSpsController: (data) => spsControllerRepository.create(data),
  updateSpsController: (id, data) => spsControllerRepository.update(id, data),
  createNotificationClass: (data) => notificationClassRepository.create(data),
  multiCreateFieldDevices: (data) => fieldDeviceRepository.multiCreate(data),
  updateFieldDevice: (id, data) => fieldDeviceRepository.update(id, data)
};

const PAGE_SIZE = 500;
const TARGETED_PAGE_SIZE = 100;
const BACNET_OBJECT_LOOKUP_CONCURRENCY = 10;
export const CREATE_NOTIFICATION_CLASS_VALUE = '__create_notification_class__';
const SUCCESS_REMOVAL_DELAY_MS = 5000;

export class FieldDeviceImportService {
  plan = $state<FieldDeviceImportPlan | null>(null);
  isTransforming = $state(false);
  isImporting = $state(false);
  transformError = $state<string | null>(null);
  importReport = $state<FieldDeviceImportReport | null>(null);
  backendDiagnostics = $state<ImportDiagnostic[]>([]);
  nodeStates = $state<Record<string, ImportNodeState>>({});
  notificationClassOptions = $state<NotificationClass[]>([]);
  hiddenNodeKeys = $state<Set<string>>(new Set());

  allDiagnostics = $derived.by(() => [
    ...(this.plan?.diagnostics ?? []),
    ...this.backendDiagnostics
  ]);

  cellMarkers = $derived.by(() => buildImportCellMarkers(this.allDiagnostics));
  rowMarkers = $derived.by(() =>
    buildImportRowMarkers(this.allDiagnostics, this.plan, this.nodeStates, this.hiddenNodeKeys)
  );
  visibleFieldDevicesBySystemType = $derived.by(() => {
    const bySystemType = new Map<string, FieldDeviceImportDevicePlan[]>();
    for (const fieldDevice of this.plan?.controller.fieldDevices ?? []) {
      if (this.hiddenNodeKeys.has(fieldDevice.key)) continue;
      const current = bySystemType.get(fieldDevice.spsControllerSystemTypeKey) ?? [];
      current.push(fieldDevice);
      bySystemType.set(fieldDevice.spsControllerSystemTypeKey, current);
    }
    return bySystemType;
  });
  visibleSystemTypes = $derived.by(
    () =>
      this.plan?.controller.systemTypes.filter((systemType) => {
        const childCount = this.visibleFieldDevicesBySystemType.get(systemType.key)?.length ?? 0;
        return childCount > 0 || this.nodeNeedsAttention(systemType.key);
      }) ?? []
  );
  hasFieldDevicesToMigrate = $derived.by(
    () =>
      this.plan?.controller.fieldDevices.some((fieldDevice) =>
        this.isFieldDeviceImportCandidate(fieldDevice)
      ) ?? false
  );
  hasStructuralWork = $derived.by(() => {
    if (!this.plan) return false;
    if (this.nodeNeedsAttention(CONTROL_CABINET_IMPORT_NODE_KEY)) return true;
    if (this.nodeNeedsAttention(SPS_CONTROLLER_IMPORT_NODE_KEY)) return true;
    return this.plan.controller.systemTypes.some((systemType) =>
      this.nodeNeedsAttention(systemType.key)
    );
  });
  hasImportWork = $derived.by(() => this.hasFieldDevicesToMigrate || this.hasStructuralWork);
  hasUnresolvedNotificationClasses = $derived.by(
    () =>
      this.plan?.controller.fieldDevices.some((fieldDevice) =>
        this.isFieldDeviceImportCandidate(fieldDevice)
          ? fieldDevice.bacnetObjects.some((object) =>
              ['invalid', 'missing'].includes(object.notificationClass.status)
            )
          : false
      ) ?? false
  );
  canImport = $derived.by(
    () =>
      Boolean(this.plan?.canImport) &&
      !this.isImporting &&
      this.hasImportWork &&
      !this.hasUnresolvedNotificationClasses
  );

  private lookups: FieldDeviceImportLookups | null = null;
  private lookupLoad: Promise<FieldDeviceImportLookups> | null = null;
  private readonly backend: FieldDeviceImportBackend;
  private readonly removalTimers = new Map<string, ReturnType<typeof setTimeout>>();

  constructor(deps: FieldDeviceImportServiceDeps = {}) {
    this.backend = {
      ...defaultFieldDeviceImportBackend,
      ...deps.backend
    };
  }

  clearTransform(): void {
    this.clearRemovalTimers();
    this.plan = null;
    this.transformError = null;
    this.importReport = null;
    this.backendDiagnostics = [];
    this.nodeStates = {};
    this.notificationClassOptions = [];
    this.hiddenNodeKeys = new Set();
  }

  async transform(worksheet: WorksheetPreview | null): Promise<void> {
    if (!worksheet || this.isTransforming) return;

    this.isTransforming = true;
    this.transformError = null;
    this.importReport = null;
    this.backendDiagnostics = [];
    this.nodeStates = {};

    try {
      this.lookups = null;
      this.lookupLoad = null;
      const criteria = collectFieldDeviceImportLookupCriteria(worksheet);
      const lookups = await this.loadLookups(criteria);
      this.notificationClassOptions = lookups.notificationClasses;
      this.plan = transformWorksheetToFieldDeviceImport(worksheet, lookups);
      this.nodeStates = buildInitialNodeStates(this.plan);
    } catch (error) {
      this.plan = null;
      this.nodeStates = {};
      this.notificationClassOptions = [];
      this.transformError =
        error instanceof Error
          ? error.message
          : translate('field_device.importer.errors.transform_prepare_failed');
    } finally {
      this.isTransforming = false;
    }
  }

  async importPlan(): Promise<void> {
    if (!this.plan || !this.plan.canImport || !this.hasImportWork || this.isImporting) {
      return;
    }

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
      cell: plan.controller.controlCabinetSourceCell,
      entityKey: CONTROL_CABINET_IMPORT_NODE_KEY
    };

    let controlCabinet: ControlCabinet | undefined;
    let spsController: SPSController | undefined;
    let migratedFieldDevices = 0;
    let failedFieldDevices = 0;

    try {
      if (!plan.controller.controlCabinetRequest || !plan.controller.spsControllerRequest) {
        throw new Error(translate('field_device.importer.errors.incomplete_plan'));
      }

      currentStep = {
        entity: 'control_cabinet',
        cell: plan.controller.controlCabinetSourceCell,
        entityKey: CONTROL_CABINET_IMPORT_NODE_KEY
      };
      controlCabinet = await this.resolveControlCabinet(plan);

      currentStep = {
        entity: 'sps_controller',
        cell: plan.controller.spsControllerSourceCell,
        entityKey: SPS_CONTROLLER_IMPORT_NODE_KEY
      };
      spsController = await this.resolveSpsController(plan, controlCabinet.id);

      currentStep = {
        entity: 'sps_controller_system_type',
        cell: plan.controller.systemTypes[0]?.sourceCell,
        entityKey: plan.controller.systemTypes[0]?.key
      };
      const systemTypeMap = await this.ensureSystemTypeMap(spsController, plan);

      currentStep = {
        entity: 'bacnet_object',
        cell: plan.controller.fieldDevices[0]?.bacnetObjects[0]?.sourceCells.notification,
        entityKey: plan.controller.fieldDevices[0]?.bacnetObjects[0]?.key
      };
      await this.ensureNotificationClassAssignments(plan);

      currentStep = {
        entity: 'field_device',
        cell: plan.controller.fieldDevices[0]?.sourceCell,
        entityKey: plan.controller.fieldDevices[0]?.key
      };
      const fieldDeviceResult = await this.importFieldDevices(plan, systemTypeMap);
      migratedFieldDevices = fieldDeviceResult.successCount;
      failedFieldDevices = fieldDeviceResult.failureCount;

      const hasFailures = failedFieldDevices > 0;
      this.importReport = {
        status: hasFailures ? 'partial' : 'success',
        createdControlCabinetId: controlCabinet.id,
        createdSpsControllerId: spsController.id,
        createdControlCabinetLabel:
          controlCabinet.control_cabinet_nr || plan.controller.controlCabinetNr,
        createdSpsControllerLabel:
          spsController.device_name ||
          plan.controller.spsControllerRequest.device_name ||
          plan.controller.spsControllerRequest.ga_device,
        createdFieldDevices: migratedFieldDevices,
        failedFieldDevices
      };
    } catch (error) {
      if (this.backendDiagnostics.length === 0) {
        this.addApiDiagnostics(error, currentStep.entity, currentStep.cell, currentStep.entityKey);
      }

      this.importReport = {
        status: 'failed',
        createdControlCabinetId: controlCabinet?.id,
        createdSpsControllerId: spsController?.id,
        createdControlCabinetLabel:
          controlCabinet?.control_cabinet_nr || plan.controller.controlCabinetNr,
        createdSpsControllerLabel:
          spsController?.device_name ||
          plan.controller.spsControllerRequest?.device_name ||
          plan.controller.spsControllerRequest?.ga_device,
        createdFieldDevices: migratedFieldDevices,
        failedFieldDevices: failedFieldDevices || plan.controller.fieldDevices.length,
        errorMessage: getErrorMessage(error)
      };
    } finally {
      this.lookups = null;
      this.isImporting = false;
    }
  }

  private async resolveControlCabinet(plan: FieldDeviceImportPlan): Promise<ControlCabinet> {
    const state = this.nodeStates[CONTROL_CABINET_IMPORT_NODE_KEY];
    if (state?.id) {
      const status = state.status === 'success' ? 'success' : 'identical';
      const controlCabinet =
        plan.controller.existingControlCabinet?.id === state.id
          ? plan.controller.existingControlCabinet
          : await this.backend.getControlCabinet(state.id);
      this.setNodeState(CONTROL_CABINET_IMPORT_NODE_KEY, {
        entity: 'control_cabinet',
        status,
        id: controlCabinet.id
      });
      return controlCabinet;
    }

    if (plan.controller.existingControlCabinet) {
      this.setNodeState(CONTROL_CABINET_IMPORT_NODE_KEY, {
        entity: 'control_cabinet',
        status: 'identical',
        id: plan.controller.existingControlCabinet.id
      });
      return plan.controller.existingControlCabinet;
    }

    if (!plan.controller.controlCabinetRequest) {
      throw new Error(translate('field_device.importer.errors.incomplete_plan'));
    }

    const controlCabinet = await this.backend.createControlCabinet(
      plan.controller.controlCabinetRequest
    );
    this.setNodeState(CONTROL_CABINET_IMPORT_NODE_KEY, {
      entity: 'control_cabinet',
      status: 'success',
      id: controlCabinet.id
    });
    return controlCabinet;
  }

  private async resolveSpsController(
    plan: FieldDeviceImportPlan,
    controlCabinetId: string
  ): Promise<SPSController> {
    const state = this.nodeStates[SPS_CONTROLLER_IMPORT_NODE_KEY];
    if (state?.id) {
      const status =
        state.status === 'success' || state.status === 'delta' ? state.status : 'identical';
      const spsController =
        plan.controller.existingSpsController?.id === state.id
          ? plan.controller.existingSpsController
          : await this.backend.getSpsController(state.id);
      this.setNodeState(SPS_CONTROLLER_IMPORT_NODE_KEY, {
        entity: 'sps_controller',
        status,
        id: spsController.id
      });
      return spsController;
    }

    if (plan.controller.existingSpsController) {
      this.setNodeState(SPS_CONTROLLER_IMPORT_NODE_KEY, {
        entity: 'sps_controller',
        status: plan.controller.existingSpsControllerComparison?.kind ?? 'identical',
        id: plan.controller.existingSpsController.id
      });
      return plan.controller.existingSpsController;
    }

    if (!plan.controller.spsControllerRequest) {
      throw new Error(translate('field_device.importer.errors.incomplete_plan'));
    }

    const spsController = await this.backend.createSpsController({
      ...plan.controller.spsControllerRequest,
      control_cabinet_id: controlCabinetId
    });
    this.setNodeState(SPS_CONTROLLER_IMPORT_NODE_KEY, {
      entity: 'sps_controller',
      status: 'success',
      id: spsController.id
    });
    return spsController;
  }

  private async ensureSystemTypeMap(
    spsController: SPSController,
    plan: FieldDeviceImportPlan
  ): Promise<Map<string, string>> {
    let current = await fetchSpsControllerSystemTypes(spsController.id);
    const existingBefore = new Set(current.map(systemTypeSignature));
    const missing = plan.controller.systemTypes.filter(
      (item) => !existingBefore.has(systemTypePlanSignature(item.systemTypeId, item.number))
    );
    const controllerHasDelta = this.nodeStates[SPS_CONTROLLER_IMPORT_NODE_KEY]?.status === 'delta';

    if (missing.length > 0 || controllerHasDelta) {
      try {
        const mergedInputs = mergeSystemTypeInputs(current, missing);
        await this.backend.updateSpsController(spsController.id, {
          id: spsController.id,
          control_cabinet_id: spsController.control_cabinet_id,
          ga_device: plan.controller.spsControllerRequest?.ga_device ?? spsController.ga_device,
          device_name:
            plan.controller.spsControllerRequest?.device_name ?? spsController.device_name,
          device_description:
            plan.controller.spsControllerRequest?.device_description ??
            spsController.device_description,
          device_location:
            plan.controller.spsControllerRequest?.device_location ?? spsController.device_location,
          ip_address: plan.controller.spsControllerRequest?.ip_address ?? spsController.ip_address,
          subnet: plan.controller.spsControllerRequest?.subnet ?? spsController.subnet,
          gateway: plan.controller.spsControllerRequest?.gateway ?? spsController.gateway,
          vlan: plan.controller.spsControllerRequest?.vlan ?? spsController.vlan,
          system_types: mergedInputs
        });
        if (controllerHasDelta) {
          this.setNodeState(SPS_CONTROLLER_IMPORT_NODE_KEY, {
            entity: 'sps_controller',
            status: 'success',
            id: spsController.id
          });
        }
        current = await fetchSpsControllerSystemTypes(spsController.id);
      } catch (error) {
        if (controllerHasDelta) {
          this.addApiDiagnostics(
            error,
            'sps_controller',
            plan.controller.spsControllerSourceCell,
            SPS_CONTROLLER_IMPORT_NODE_KEY
          );
        }
        for (const item of missing) {
          this.addApiDiagnostics(error, 'sps_controller_system_type', item.sourceCell, item.key);
        }
        throw error;
      }
    }

    const currentBySignature = new Map(current.map((item) => [systemTypeSignature(item), item]));
    const out = new Map<string, string>();
    const controllerWasCreated =
      this.nodeStates[SPS_CONTROLLER_IMPORT_NODE_KEY]?.status === 'success';

    for (const item of plan.controller.systemTypes) {
      const signature = systemTypePlanSignature(item.systemTypeId, item.number);
      const existing = currentBySignature.get(signature);
      if (!existing) {
        this.addBackendDiagnostic(
          translate('field_device.importer.errors.created_system_type_missing'),
          'sps_controller_system_type',
          item.sourceCell,
          item.key
        );
        continue;
      }

      out.set(item.key, existing.id);
      this.setNodeState(item.key, {
        entity: 'sps_controller_system_type',
        status:
          !controllerWasCreated &&
          (item.existingSpsControllerSystemTypeId || existingBefore.has(signature))
            ? 'identical'
            : 'success',
        id: existing.id
      });
    }

    if (out.size !== plan.controller.systemTypes.length) {
      throw new Error(translate('field_device.importer.errors.system_types_unresolved'));
    }

    return out;
  }

  private async ensureNotificationClassAssignments(plan: FieldDeviceImportPlan): Promise<void> {
    const byNumber = new Map(this.notificationClassOptions.map((item) => [item.nc, item]));

    for (const fieldDevice of plan.controller.fieldDevices) {
      if (!this.isFieldDeviceImportCandidate(fieldDevice)) continue;

      for (const bacnetObject of fieldDevice.bacnetObjects) {
        if (bacnetObject.notificationClass.status !== 'create') continue;

        const number = bacnetObject.notificationClass.number;
        if (!number) {
          this.addBackendDiagnostic(
            translate('field_device.importer.validation.notification_no_match', {
              value: bacnetObject.notificationClass.raw
            }),
            'bacnet_object',
            bacnetObject.sourceCells.notification,
            bacnetObject.key
          );
          continue;
        }

        const cached = byNumber.get(number);
        if (cached) {
          this.applyNotificationClass(bacnetObject, cached, 'manual');
          continue;
        }

        try {
          const created = await this.backend.createNotificationClass(
            createNotificationClassRequest(number, bacnetObject.notificationClass.raw)
          );
          byNumber.set(created.nc, created);
          this.notificationClassOptions = uniqueById([...this.notificationClassOptions, created]);
          this.applyNotificationClass(bacnetObject, created, 'manual');
        } catch (error) {
          this.addApiDiagnostics(
            error,
            'bacnet_object',
            bacnetObject.sourceCells.notification,
            bacnetObject.key
          );
        }
      }
    }

    const unresolved = plan.controller.fieldDevices.some((fieldDevice) =>
      this.isFieldDeviceImportCandidate(fieldDevice)
        ? fieldDevice.bacnetObjects.some(
            (object) =>
              object.notificationClass.status === 'create' ||
              object.notificationClass.status === 'missing' ||
              object.notificationClass.status === 'invalid'
          )
        : false
    );
    if (unresolved) {
      throw new Error(translate('field_device.importer.errors.notification_classes_unresolved'));
    }
  }

  private async importFieldDevices(
    plan: FieldDeviceImportPlan,
    systemTypeMap: Map<string, string>
  ): Promise<{ successCount: number; failureCount: number }> {
    const candidates = plan.controller.fieldDevices.filter((fieldDevice) =>
      this.isFieldDeviceImportCandidate(fieldDevice)
    );
    let successCount = 0;
    let failureCount = 0;

    const existingDevices = candidates.filter((fieldDevice) =>
      this.fieldDeviceTargetId(fieldDevice)
    );
    for (const fieldDevice of existingDevices) {
      const ok = await this.updateExistingFieldDevice(fieldDevice, systemTypeMap);
      if (ok) {
        successCount += 1;
      } else {
        failureCount += 1;
      }
    }

    const newDevices = candidates.filter((fieldDevice) => !this.fieldDeviceTargetId(fieldDevice));
    if (newDevices.length === 0) {
      return { successCount, failureCount };
    }

    const response = await this.backend.multiCreateFieldDevices({
      field_devices: newDevices.map((fieldDevice) =>
        toCreateFieldDeviceRequest(fieldDevice, systemTypeMap)
      )
    });

    for (const result of response.results) {
      const fieldDevice = newDevices[result.index];
      if (!fieldDevice) continue;

      if (result.success) {
        this.markFieldDeviceSuccess(fieldDevice, result.field_device?.id);
        successCount += 1;
        continue;
      }

      this.addFieldDeviceDiagnostic(
        fieldDevice,
        result.error || translate('field_device.importer.errors.field_device_create_failed'),
        result.error_field
      );
      failureCount += 1;
    }

    return { successCount, failureCount };
  }

  private async updateExistingFieldDevice(
    fieldDevice: FieldDeviceImportDevicePlan,
    systemTypeMap: Map<string, string>
  ): Promise<boolean> {
    const targetId = this.fieldDeviceTargetId(fieldDevice);
    if (!targetId) return false;

    try {
      const updated = await this.backend.updateFieldDevice(
        targetId,
        toUpdateFieldDeviceRequest(fieldDevice, systemTypeMap)
      );
      this.markFieldDeviceSuccess(fieldDevice, updated.id);
      return true;
    } catch (error) {
      this.addFieldDeviceApiDiagnostics(error, fieldDevice);
      return false;
    }
  }

  nodeState(key: string): ImportNodeState | undefined {
    return this.nodeStates[key];
  }

  diagnosticsForNode(key: string): ImportDiagnostic[] {
    return this.allDiagnostics.filter((diagnostic) => diagnostic.entityKey === key);
  }

  visibleDevicesForSystemType(key: string): FieldDeviceImportDevicePlan[] {
    return this.visibleFieldDevicesBySystemType.get(key) ?? [];
  }

  isNodeHidden(key: string): boolean {
    return this.hiddenNodeKeys.has(key);
  }

  notificationClassSelectionValue(
    bacnetObject: FieldDeviceImportDevicePlan['bacnetObjects'][number]
  ): string {
    if (bacnetObject.notificationClass.status === 'create') return CREATE_NOTIFICATION_CLASS_VALUE;
    return bacnetObject.request.notification_class_id ?? '';
  }

  updateBacnetNotificationClass(
    fieldDeviceKey: string,
    bacnetObjectKey: string,
    value: string
  ): void {
    const bacnetObject = this.findBacnetObject(fieldDeviceKey, bacnetObjectKey);
    if (!bacnetObject) return;

    if (value === CREATE_NOTIFICATION_CLASS_VALUE && bacnetObject.notificationClass.number) {
      bacnetObject.notificationClass = {
        ...bacnetObject.notificationClass,
        id: undefined,
        status: 'create'
      };
      bacnetObject.request.notification_class_id = undefined;
      this.resetBacnetNode(fieldDeviceKey, bacnetObjectKey);
      return;
    }

    const notificationClass = this.notificationClassOptions.find((item) => item.id === value);
    bacnetObject.notificationClass = {
      raw: bacnetObject.notificationClass.raw,
      number: notificationClass?.nc ?? bacnetObject.notificationClass.number,
      id: notificationClass?.id,
      status: notificationClass ? 'manual' : 'missing'
    };
    bacnetObject.request.notification_class_id = notificationClass?.id;
    this.resetBacnetNode(fieldDeviceKey, bacnetObjectKey);
  }

  dispose(): void {
    this.clearRemovalTimers();
  }

  updateFieldDeviceBmk(fieldDeviceKey: string, value: string): void {
    const fieldDevice = this.findFieldDevice(fieldDeviceKey);
    if (!fieldDevice) return;
    fieldDevice.request.bmk = emptyToUndefined(value);
    this.resetFieldDeviceNode(fieldDevice);
  }

  updateFieldDeviceDescription(fieldDeviceKey: string, value: string): void {
    const fieldDevice = this.findFieldDevice(fieldDeviceKey);
    if (!fieldDevice) return;
    fieldDevice.request.description = emptyToUndefined(value);
    this.resetFieldDeviceNode(fieldDevice);
  }

  updateFieldDeviceApparatNr(fieldDeviceKey: string, value: string): void {
    const fieldDevice = this.findFieldDevice(fieldDeviceKey);
    if (!fieldDevice) return;
    const parsed = Number.parseInt(value, 10);
    fieldDevice.apparatNr = Number.isFinite(parsed) ? parsed : undefined;
    fieldDevice.request.apparat_nr = Number.isFinite(parsed) ? parsed : 0;
    this.resetFieldDeviceNode(fieldDevice);
  }

  updateBacnetTextFix(fieldDeviceKey: string, bacnetObjectKey: string, value: string): void {
    const bacnetObject = this.findBacnetObject(fieldDeviceKey, bacnetObjectKey);
    if (!bacnetObject) return;
    bacnetObject.textFix = value;
    bacnetObject.request.text_fix = value;
    this.resetBacnetNode(fieldDeviceKey, bacnetObjectKey);
  }

  updateBacnetAddress(fieldDeviceKey: string, bacnetObjectKey: string, value: string): void {
    const bacnetObject = this.findBacnetObject(fieldDeviceKey, bacnetObjectKey);
    if (!bacnetObject) return;

    bacnetObject.address = value;
    const parsed = parseBacnetAddress(value);
    bacnetObject.request.software_type = parsed?.type ?? '';
    bacnetObject.request.software_number = parsed?.number ?? 0;
    this.resetBacnetNode(fieldDeviceKey, bacnetObjectKey);
  }

  private findFieldDevice(fieldDeviceKey: string): FieldDeviceImportDevicePlan | undefined {
    return this.plan?.controller.fieldDevices.find((item) => item.key === fieldDeviceKey);
  }

  private findBacnetObject(
    fieldDeviceKey: string,
    bacnetObjectKey: string
  ): FieldDeviceImportDevicePlan['bacnetObjects'][number] | undefined {
    return this.findFieldDevice(fieldDeviceKey)?.bacnetObjects.find(
      (item) => item.key === bacnetObjectKey
    );
  }

  private applyNotificationClass(
    bacnetObject: FieldDeviceImportDevicePlan['bacnetObjects'][number],
    notificationClass: NotificationClass,
    status: Extract<ImportNotificationClassResolution['status'], 'matched' | 'manual'>
  ): void {
    bacnetObject.notificationClass = {
      raw: bacnetObject.notificationClass.raw,
      number: notificationClass.nc,
      id: notificationClass.id,
      status
    };
    bacnetObject.request.notification_class_id = notificationClass.id;
  }

  private fieldDeviceTargetId(fieldDevice: FieldDeviceImportDevicePlan): string | undefined {
    return this.nodeStates[fieldDevice.key]?.id ?? fieldDevice.existingFieldDeviceId;
  }

  private markFieldDeviceSuccess(fieldDevice: FieldDeviceImportDevicePlan, id?: string): void {
    this.setNodeState(fieldDevice.key, {
      entity: 'field_device',
      status: 'success',
      id: id ?? this.fieldDeviceTargetId(fieldDevice)
    });

    for (const object of fieldDevice.bacnetObjects) {
      this.setNodeState(object.key, {
        entity: 'bacnet_object',
        status: 'success'
      });
    }
    this.scheduleSuccessfulNodeRemoval([
      fieldDevice.key,
      ...fieldDevice.bacnetObjects.map((object) => object.key)
    ]);
  }

  private addFieldDeviceApiDiagnostics(
    error: unknown,
    fieldDevice: FieldDeviceImportDevicePlan
  ): void {
    const fieldErrors = getFieldErrors(error);
    const entries = Object.entries(fieldErrors);
    if (entries.length === 0) {
      this.addFieldDeviceDiagnostic(fieldDevice, getErrorMessage(error), '');
      return;
    }

    for (const [field, message] of entries) {
      this.addFieldDeviceDiagnostic(fieldDevice, message, field);
    }
  }

  private addFieldDeviceDiagnostic(
    fieldDevice: FieldDeviceImportDevicePlan,
    message: string,
    errorField: string
  ): void {
    const target = diagnosticTargetForFieldDeviceError(fieldDevice, errorField);
    this.addBackendDiagnostic(message, target.entity, target.cell, target.entityKey);
    if (target.entityKey !== fieldDevice.key) {
      this.setNodeState(fieldDevice.key, {
        entity: 'field_device',
        status: 'failed'
      });
    }
  }

  private resetFieldDeviceNode(fieldDevice: FieldDeviceImportDevicePlan): void {
    const keys = [fieldDevice.key, ...fieldDevice.bacnetObjects.map((object) => object.key)];
    this.cancelNodeRemoval(keys);
    this.clearDiagnosticsForNodes(keys);
    this.setNodeState(fieldDevice.key, {
      entity: 'field_device',
      status: fieldDevice.existingFieldDeviceId ? 'delta' : 'pending',
      id: fieldDevice.existingFieldDeviceId
    });
    for (const object of fieldDevice.bacnetObjects) {
      this.setNodeState(object.key, {
        entity: 'bacnet_object',
        status: 'pending'
      });
    }
    this.touchPlan();
  }

  private resetBacnetNode(fieldDeviceKey: string, bacnetObjectKey: string): void {
    this.cancelNodeRemoval([fieldDeviceKey, bacnetObjectKey]);
    this.clearDiagnosticsForNodes([fieldDeviceKey, bacnetObjectKey]);
    this.setNodeState(fieldDeviceKey, {
      entity: 'field_device',
      status: this.findFieldDevice(fieldDeviceKey)?.existingFieldDeviceId ? 'delta' : 'pending',
      id: this.findFieldDevice(fieldDeviceKey)?.existingFieldDeviceId
    });
    this.setNodeState(bacnetObjectKey, {
      entity: 'bacnet_object',
      status: 'pending'
    });
    this.touchPlan();
  }

  private isFieldDeviceImportCandidate(fieldDevice: FieldDeviceImportDevicePlan): boolean {
    if (this.hiddenNodeKeys.has(fieldDevice.key)) return false;
    const status = this.nodeStates[fieldDevice.key]?.status ?? 'pending';
    return status !== 'success' && status !== 'identical';
  }

  private nodeNeedsAttention(key: string): boolean {
    if (this.hiddenNodeKeys.has(key)) return false;
    const status = this.nodeStates[key]?.status;
    return status === 'failed' || status === 'pending' || status === 'delta';
  }

  private scheduleSuccessfulNodeRemoval(keys: string[]): void {
    for (const key of keys) {
      const existing = this.removalTimers.get(key);
      if (existing) {
        clearTimeout(existing);
      }

      const timer = setTimeout(() => {
        this.removalTimers.delete(key);
        this.hiddenNodeKeys = new Set([...this.hiddenNodeKeys, key]);
      }, SUCCESS_REMOVAL_DELAY_MS);

      this.removalTimers.set(key, timer);
    }
  }

  private cancelNodeRemoval(keys: string[]): void {
    let changed = false;
    const nextHidden = new Set(this.hiddenNodeKeys);

    for (const key of keys) {
      const timer = this.removalTimers.get(key);
      if (timer) {
        clearTimeout(timer);
        this.removalTimers.delete(key);
      }
      if (nextHidden.delete(key)) {
        changed = true;
      }
    }

    if (changed) {
      this.hiddenNodeKeys = nextHidden;
    }
  }

  private clearRemovalTimers(): void {
    for (const timer of this.removalTimers.values()) {
      clearTimeout(timer);
    }
    this.removalTimers.clear();
  }

  private clearDiagnosticsForNodes(keys: string[]): void {
    const keySet = new Set(keys);
    this.backendDiagnostics = this.backendDiagnostics.filter(
      (diagnostic) => !diagnostic.entityKey || !keySet.has(diagnostic.entityKey)
    );

    if (!this.plan) return;
    this.plan.diagnostics = this.plan.diagnostics.filter(
      (diagnostic) => !diagnostic.entityKey || !keySet.has(diagnostic.entityKey)
    );
    refreshPlanValidation(this.plan);
  }

  private touchPlan(): void {
    if (!this.plan) return;
    this.plan = {
      ...this.plan,
      controller: {
        ...this.plan.controller,
        systemTypes: [...this.plan.controller.systemTypes],
        fieldDevices: [...this.plan.controller.fieldDevices]
      }
    };
  }

  private setNodeState(key: string, next: ImportNodeState): void {
    const merged = {
      ...this.nodeStates[key],
      ...next
    };
    if (next.status !== 'failed' && next.message === undefined) {
      delete merged.message;
    }

    this.nodeStates = {
      ...this.nodeStates,
      [key]: merged
    };
  }

  private async loadLookups(
    criteria: FieldDeviceImportLookupCriteria
  ): Promise<FieldDeviceImportLookups> {
    if (this.lookups) return this.lookups;
    if (this.lookupLoad) return this.lookupLoad;

    this.lookupLoad = this.backend.loadLookups(criteria).then((lookups) => {
      this.lookups = lookups;
      return lookups;
    });

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
    if (entityKey) {
      this.setNodeState(entityKey, {
        entity,
        status: 'failed',
        message
      });
    }

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

function toUpdateFieldDeviceRequest(
  fieldDevice: FieldDeviceImportDevicePlan,
  systemTypeMap: Map<string, string>
): UpdateFieldDeviceRequest {
  const createRequest = toCreateFieldDeviceRequest(fieldDevice, systemTypeMap);
  return {
    bmk: createRequest.bmk,
    description: createRequest.description,
    text_fix: createRequest.text_fix,
    apparat_nr: createRequest.apparat_nr,
    sps_controller_system_type_id: createRequest.sps_controller_system_type_id,
    system_part_id: createRequest.system_part_id,
    apparat_id: createRequest.apparat_id,
    object_data_id: createRequest.object_data_id,
    bacnet_objects: createRequest.bacnet_objects
  };
}

function buildInitialNodeStates(plan: FieldDeviceImportPlan): Record<string, ImportNodeState> {
  const states: Record<string, ImportNodeState> = {
    [CONTROL_CABINET_IMPORT_NODE_KEY]: {
      entity: 'control_cabinet',
      status: plan.controller.existingControlCabinet ? 'identical' : 'pending',
      id: plan.controller.existingControlCabinet?.id
    },
    [SPS_CONTROLLER_IMPORT_NODE_KEY]: {
      entity: 'sps_controller',
      status: plan.controller.existingSpsController
        ? (plan.controller.existingSpsControllerComparison?.kind ?? 'identical')
        : 'pending',
      id: plan.controller.existingSpsController?.id
    }
  };

  for (const systemType of plan.controller.systemTypes) {
    states[systemType.key] = {
      entity: 'sps_controller_system_type',
      status: systemType.existingSpsControllerSystemTypeId ? 'identical' : 'pending',
      id: systemType.existingSpsControllerSystemTypeId
    };
  }

  for (const fieldDevice of plan.controller.fieldDevices) {
    states[fieldDevice.key] = {
      entity: 'field_device',
      status: fieldDevice.existingFieldDeviceId
        ? (fieldDevice.existingComparison?.kind ?? 'delta')
        : 'pending',
      id: fieldDevice.existingFieldDeviceId
    };
    for (const object of fieldDevice.bacnetObjects) {
      states[object.key] = {
        entity: 'bacnet_object',
        status: 'pending'
      };
    }
  }

  return states;
}

function diagnosticTargetForFieldDeviceError(
  fieldDevice: FieldDeviceImportDevicePlan,
  errorField: string
): { entity: ImportEntityKind; cell: ImportSourceCell; entityKey: string } {
  const field = (errorField || '').toLowerCase();
  const bacnetMatch = field.match(/bacnet_objects\.(\d+)\.([a-z_]+)/);
  if (bacnetMatch) {
    const bacnet = fieldDevice.bacnetObjects[Number.parseInt(bacnetMatch[1], 10)];
    const bacnetField = bacnetMatch[2] ?? '';
    if (bacnet) {
      const cell = bacnetField.includes('text_fix')
        ? bacnet.sourceCells.textFix
        : bacnetField.includes('notification_class')
          ? bacnet.sourceCells.notification
          : bacnetField.includes('software')
            ? bacnet.sourceCells.address
            : bacnet.sourceCell;
      return {
        entity: 'bacnet_object',
        cell,
        entityKey: bacnet.key
      };
    }
  }

  let cell = fieldDevice.sourceCell;
  if (field.includes('bmk')) cell = fieldDevice.sourceCells.bmk;
  else if (field.includes('description')) cell = fieldDevice.sourceCells.description;
  else if (field.includes('apparat_nr')) cell = fieldDevice.sourceCells.apparatNr;
  else if (field.includes('system_part')) cell = fieldDevice.sourceCells.systemPart;
  else if (field.includes('apparat')) cell = fieldDevice.sourceCells.apparat;
  if (field.includes('sps_controller_system_type')) {
    cell = fieldDevice.sourceCells.spsControllerSystemType;
  }
  return {
    entity: 'field_device',
    cell,
    entityKey: fieldDevice.key
  };
}

function mergeSystemTypeInputs(
  current: SPSControllerSystemType[],
  missing: FieldDeviceImportPlan['controller']['systemTypes']
): SPSControllerSystemTypeInput[] {
  return [
    ...current.map((item) => ({
      id: item.id,
      system_type_id: item.system_type_id,
      number: item.number,
      document_name: item.document_name
    })),
    ...missing.map((item) => ({
      system_type_id: item.systemTypeId,
      number: item.number
    }))
  ];
}

function systemTypeSignature(item: SPSControllerSystemType): string {
  return systemTypePlanSignature(item.system_type_id, item.number);
}

function systemTypePlanSignature(systemTypeId: string, number?: number): string {
  return `${systemTypeId}|${number ?? ''}`;
}

function refreshPlanValidation(plan: FieldDeviceImportPlan): void {
  plan.errorCount = plan.diagnostics.filter((item) => item.severity === 'error').length;
  plan.warningCount = plan.diagnostics.filter((item) => item.severity === 'warning').length;
  plan.canImport =
    plan.errorCount === 0 &&
    Boolean(plan.controller.controlCabinetRequest) &&
    Boolean(plan.controller.spsControllerRequest) &&
    plan.controller.fieldDevices.length > 0;
}

function parseBacnetAddress(value: string): { type: string; number: number } | null {
  const match = value.trim().match(/^([a-z]{2})(\d+)$/i);
  if (!match) return null;

  const number = Number.parseInt(match[2], 10);
  if (!Number.isFinite(number) || number < 0 || number > 65535) return null;

  return {
    type: match[1].toLowerCase(),
    number
  };
}

function emptyToUndefined(value: string): string | undefined {
  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : undefined;
}

function createNotificationClassRequest(
  number: number,
  raw: string
): CreateNotificationClassRequest {
  const label = raw.trim() || `NC ${number}`;
  return {
    event_category: `NC ${number}`,
    nc: number,
    object_description: label,
    internal_description: label,
    meaning: label,
    ack_required_not_normal: false,
    ack_required_error: false,
    ack_required_normal: false,
    norm_not_normal: 0,
    norm_error: 0,
    norm_normal: 0
  };
}

async function loadTargetedLookups(
  criteria: FieldDeviceImportLookupCriteria
): Promise<FieldDeviceImportLookups> {
  const [
    buildings,
    systemTypes,
    systemParts,
    apparats,
    stateTexts,
    notificationClasses,
    alarmTypes
  ] = await Promise.all([
    fetchBuildingsForCriteria(criteria),
    fetchAllListPages(systemTypeRepository),
    fetchMatchingSystemParts(criteria.systemPartLabels),
    fetchMatchingApparats(criteria.apparatLabels),
    fetchMatchingStateTexts(criteria.stateTextLabels),
    fetchNotificationClassesForImport(),
    fetchMatchingAlarmTypes(criteria.alarmTypeLabels)
  ]);

  const controlCabinets = await fetchControlCabinetsForCriteria(criteria, buildings);
  const spsControllers = await fetchSpsControllersForCriteria(criteria, controlCabinets);
  const spsControllerSystemTypes = await fetchSpsControllerSystemTypesForCriteria(
    criteria,
    spsControllers
  );
  const fieldDevices = await fetchFieldDevicesForCriteria(criteria, spsControllerSystemTypes);
  const fieldDevicesWithBacnetObjects = await attachBacnetObjectsToFieldDevices(fieldDevices);

  return {
    buildings,
    controlCabinets,
    spsControllers,
    spsControllerSystemTypes,
    fieldDevices: fieldDevicesWithBacnetObjects,
    systemTypes,
    systemParts,
    apparats,
    stateTexts,
    notificationClasses,
    alarmTypes
  };
}

async function fetchBuildingsForCriteria(
  criteria: FieldDeviceImportLookupCriteria
): Promise<Building[]> {
  if (!criteria.iwsCode) return [];

  const normalizedIws = normalizeLookupKey(criteria.iwsCode);
  const buildings = await fetchAllListPages(buildingRepository, {
    search: criteria.iwsCode,
    pageSize: TARGETED_PAGE_SIZE
  });

  return uniqueById(
    buildings.filter(
      (building) =>
        normalizeLookupKey(building.iws_code) === normalizedIws &&
        (criteria.buildingGroup === undefined || building.building_group === criteria.buildingGroup)
    )
  );
}

async function fetchControlCabinetsForCriteria(
  criteria: FieldDeviceImportLookupCriteria,
  buildings: Building[]
): Promise<ControlCabinet[]> {
  if (!criteria.controlCabinetNr || buildings.length === 0) return [];

  const normalizedCabinetNr = normalizeLookupKey(criteria.controlCabinetNr);
  const batches = await Promise.all(
    buildings.map((building) =>
      fetchAllListPages(controlCabinetRepository, {
        search: criteria.controlCabinetNr,
        filters: { building_id: building.id },
        pageSize: TARGETED_PAGE_SIZE
      })
    )
  );

  const buildingIds = new Set(buildings.map((building) => building.id));
  return uniqueById(
    batches
      .flat()
      .filter(
        (cabinet) =>
          buildingIds.has(cabinet.building_id) &&
          normalizeLookupKey(cabinet.control_cabinet_nr) === normalizedCabinetNr
      )
  );
}

async function fetchSpsControllersForCriteria(
  criteria: FieldDeviceImportLookupCriteria,
  controlCabinets: ControlCabinet[]
): Promise<SPSController[]> {
  const terms = uniqueSearchTerms([criteria.deviceName, criteria.gaDevice]);
  if (terms.length === 0 || controlCabinets.length === 0) return [];

  const batches = await Promise.all(
    controlCabinets.flatMap((controlCabinet) =>
      terms.map((term) =>
        fetchAllListPages(spsControllerRepository, {
          search: term.value,
          filters: { control_cabinet_id: controlCabinet.id },
          pageSize: TARGETED_PAGE_SIZE
        })
      )
    )
  );

  const cabinetIds = new Set(controlCabinets.map((controlCabinet) => controlCabinet.id));
  const normalizedDeviceName = normalizeLookupKey(criteria.deviceName);
  const normalizedGaDevice = normalizeLookupKey(criteria.gaDevice);

  return uniqueById(
    batches.flat().filter((controller) => {
      const matchesDeviceName =
        normalizedDeviceName.length > 0 &&
        normalizeLookupKey(controller.device_name) === normalizedDeviceName;
      const matchesGaDevice =
        normalizedGaDevice.length > 0 &&
        normalizeLookupKey(controller.ga_device ?? '') === normalizedGaDevice;
      return (
        cabinetIds.has(controller.control_cabinet_id) && (matchesDeviceName || matchesGaDevice)
      );
    })
  );
}

async function fetchSpsControllerSystemTypesForCriteria(
  criteria: FieldDeviceImportLookupCriteria,
  spsControllers: SPSController[]
): Promise<SPSControllerSystemType[]> {
  if (spsControllers.length === 0) return [];

  const wantedNumbers = new Set(criteria.systemTypeNumbers);
  const batches = await Promise.all(
    spsControllers.map((controller) => fetchSpsControllerSystemTypes(controller.id))
  );

  return uniqueById(
    batches
      .flat()
      .filter(
        (systemType) =>
          wantedNumbers.size === 0 ||
          (systemType.number !== undefined && wantedNumbers.has(systemType.number))
      )
  );
}

async function fetchFieldDevicesForCriteria(
  criteria: FieldDeviceImportLookupCriteria,
  spsControllerSystemTypes: SPSControllerSystemType[]
): Promise<FieldDevice[]> {
  if (spsControllerSystemTypes.length === 0) return [];

  const wantedNumbers = new Set(criteria.systemTypeNumbers);
  const relevantSystemTypes = spsControllerSystemTypes.filter(
    (systemType) =>
      wantedNumbers.size === 0 ||
      (systemType.number !== undefined && wantedNumbers.has(systemType.number))
  );
  const batches = await Promise.all(
    relevantSystemTypes.map((systemType) =>
      fetchAllListPages(fieldDeviceRepository, {
        filters: { sps_controller_system_type_id: systemType.id }
      })
    )
  );

  const relevantIds = new Set(relevantSystemTypes.map((systemType) => systemType.id));
  return uniqueById(
    batches
      .flat()
      .filter((fieldDevice) => relevantIds.has(fieldDevice.sps_controller_system_type_id))
  );
}

async function fetchMatchingSystemParts(labels: string[]): Promise<SystemPart[]> {
  return fetchMatchingListItems(systemPartRepository, labels, (systemPart, normalizedTerm) =>
    matchesAnyLookupValue(normalizedTerm, systemPart.short_name, systemPart.name)
  );
}

async function fetchMatchingApparats(labels: string[]): Promise<Apparat[]> {
  return fetchMatchingListItems(apparatRepository, labels, (apparat, normalizedTerm) =>
    matchesAnyLookupValue(normalizedTerm, apparat.short_name, apparat.name)
  );
}

async function fetchMatchingStateTexts(labels: string[]): Promise<StateText[]> {
  return fetchMatchingListItems(stateTextRepository, labels, (stateText, normalizedTerm, term) => {
    const refNumber = parseNumberTerm(term);
    return (
      (refNumber !== undefined && stateText.ref_number === refNumber) ||
      stateTextValuesForLookup(stateText).some(
        (value) => normalizeLookupKey(value) === normalizedTerm
      )
    );
  });
}

async function fetchNotificationClassesForImport(): Promise<NotificationClass[]> {
  return fetchAllListPages(notificationClassRepository, {
    pageSize: PAGE_SIZE
  });
}

async function fetchMatchingAlarmTypes(labels: string[]): Promise<AlarmType[]> {
  const terms = uniqueSearchTerms(labels);
  if (terms.length === 0) return [];

  const batches = await Promise.all(
    terms.map((term) =>
      fetchAllAlarmTypes(term.value, TARGETED_PAGE_SIZE).then((items) =>
        items.filter((alarmType) =>
          matchesAnyLookupValue(term.normalized, alarmType.code, alarmType.name)
        )
      )
    )
  );

  return uniqueById(batches.flat());
}

async function attachBacnetObjectsToFieldDevices(
  fieldDevices: FieldDevice[]
): Promise<FieldDevice[]> {
  if (fieldDevices.length === 0) return [];

  const rows = await mapWithConcurrency(
    fieldDevices,
    BACNET_OBJECT_LOOKUP_CONCURRENCY,
    async (fieldDevice) => ({
      ...fieldDevice,
      bacnet_objects: await fieldDeviceRepository.listBacnetObjects(fieldDevice.id)
    })
  );

  return rows;
}

async function mapWithConcurrency<T, R>(
  items: readonly T[],
  limit: number,
  mapper: (item: T) => Promise<R>
): Promise<R[]> {
  const out = new Array<R>(items.length);
  let nextIndex = 0;

  async function worker(): Promise<void> {
    while (nextIndex < items.length) {
      const index = nextIndex;
      nextIndex += 1;
      out[index] = await mapper(items[index]);
    }
  }

  await Promise.all(Array.from({ length: Math.min(limit, items.length) }, () => worker()));
  return out;
}

async function fetchMatchingListItems<T extends { id: string }>(
  repository: ListRepository<T>,
  labels: string[],
  matches: (item: T, normalizedTerm: string, term: string) => boolean
): Promise<T[]> {
  const terms = uniqueSearchTerms(labels);
  if (terms.length === 0) return [];

  const batches = await Promise.all(
    terms.map((term) =>
      fetchAllListPages(repository, {
        search: term.value,
        pageSize: TARGETED_PAGE_SIZE
      }).then((items) => items.filter((item) => matches(item, term.normalized, term.value)))
    )
  );

  return uniqueById(batches.flat());
}

interface ListPageOptions {
  search?: string;
  filters?: Record<string, string>;
  pageSize?: number;
}

async function fetchAllListPages<T>(
  repository: ListRepository<T>,
  options: ListPageOptions = {}
): Promise<T[]> {
  return fetchAllPages(async (page, pageSize) => {
    const response: PaginatedResponse<T> = await repository.list(
      listParams(page, pageSize, options.search ?? '', options.filters)
    );
    return {
      items: response.items,
      page: response.metadata.page,
      totalPages: response.metadata.totalPages
    };
  }, options.pageSize ?? PAGE_SIZE);
}

async function fetchSpsControllerSystemTypes(
  spsControllerId: string
): Promise<SPSControllerSystemType[]> {
  return fetchAllPages(async (page, pageSize) => {
    const response = await spsControllerRepository.listSystemTypes({
      page,
      limit: pageSize,
      sps_controller_id: spsControllerId
    });
    return {
      items: response.items,
      page: response.page,
      totalPages: response.total_pages
    };
  });
}

async function fetchAllAlarmTypes(search = '', pageSize = PAGE_SIZE): Promise<AlarmType[]> {
  return fetchAllPages(async (page, pageSize) => {
    const response = await alarmTypeRepository.list({ page, pageSize, search });
    return {
      items: response.items,
      page: response.page,
      totalPages: response.totalPages
    };
  }, pageSize);
}

async function fetchAllPages<T>(
  fetchPage: (
    page: number,
    pageSize: number
  ) => Promise<{ items: T[]; page: number; totalPages: number }>,
  pageSize = PAGE_SIZE
): Promise<T[]> {
  const items: T[] = [];
  let page = 1;

  while (true) {
    const response = await fetchPage(page, pageSize);
    items.push(...response.items);
    if (response.page >= response.totalPages) break;
    page += 1;
  }

  return items;
}

function listParams(
  page: number,
  pageSize: number,
  search = '',
  filters?: Record<string, string>
): ListParams {
  return {
    pagination: { page, pageSize },
    search: { text: search },
    filters
  };
}

function uniqueSearchTerms(labels: string[]): { value: string; normalized: string }[] {
  const out: { value: string; normalized: string }[] = [];
  const seen = new Set<string>();

  for (const label of labels) {
    const value = label.trim();
    const normalized = normalizeLookupKey(value);
    if (!normalized || seen.has(normalized)) continue;
    seen.add(normalized);
    out.push({ value, normalized });
  }

  return out;
}

function matchesAnyLookupValue(normalizedTerm: string, ...values: (string | undefined)[]): boolean {
  return values.some((value) => normalizeLookupKey(value ?? '') === normalizedTerm);
}

function stateTextValuesForLookup(item: StateText): string[] {
  return Array.from({ length: 16 }, (_, index) => {
    const key = `state_text${index + 1}` as keyof StateText;
    return typeof item[key] === 'string' ? (item[key] as string) : '';
  }).filter(Boolean);
}

function parseNumberTerm(value: string): number | undefined {
  const parsed = Number.parseInt(value.trim(), 10);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function uniqueById<T extends { id: string }>(items: T[]): T[] {
  const byId = new Map<string, T>();
  for (const item of items) {
    if (!byId.has(item.id)) {
      byId.set(item.id, item);
    }
  }
  return Array.from(byId.values());
}
