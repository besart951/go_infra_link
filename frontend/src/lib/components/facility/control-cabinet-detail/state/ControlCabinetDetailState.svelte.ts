import { t as translate } from '$lib/i18n/index.js';
import { canPerform } from '$lib/utils/permissions.js';
import { ManageControlCabinetUseCase } from '$lib/application/useCases/facility/manageControlCabinetUseCase.js';
import { ManageSPSControllerUseCase } from '$lib/application/useCases/facility/manageSPSControllerUseCase.js';
import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
import type { ToastType } from '$lib/components/toast.svelte';

import type {
  Building,
  ControlCabinet,
  SPSController,
  SPSControllerSystemType
} from '$lib/domain/facility/index.js';

export interface ControlCabinetDetailData {
  cabinet: ControlCabinet;
  building?: Building | null;
  spsControllers?: SPSController[];
  systemTypesByController?: Record<string, SPSControllerSystemType[]>;
}

interface ConfirmOptions {
  title: string;
  message: string;
  confirmText: string;
  cancelText: string;
  variant?: 'destructive' | 'default';
}

export interface ControlCabinetDetailStateOptions {
  data: () => ControlCabinetDetailData;
  confirmAction: (options: ConfirmOptions) => Promise<boolean>;
  toastAction: (message: string, type?: ToastType) => void;
  gotoAction: (href: string) => Promise<void>;
  invalidateAllAction: () => Promise<void>;
}

export class ControlCabinetDetailState {
  showCabinetEdit = $state(false);
  showSpsCreate = $state(false);
  editingSps = $state<SPSController | undefined>(undefined);

  private readonly resolveData: () => ControlCabinetDetailData;
  private readonly confirmAction: (options: ConfirmOptions) => Promise<boolean>;
  private readonly toastAction: (message: string, type?: ToastType) => void;
  private readonly gotoAction: (href: string) => Promise<void>;
  private readonly invalidateAllAction: () => Promise<void>;

  private readonly manageControlCabinet = new ManageControlCabinetUseCase(controlCabinetRepository);
  private readonly manageSpsController = new ManageSPSControllerUseCase(spsControllerRepository);

  constructor(options: ControlCabinetDetailStateOptions) {
    this.resolveData = options.data;
    this.confirmAction = options.confirmAction;
    this.toastAction = options.toastAction;
    this.gotoAction = options.gotoAction;
    this.invalidateAllAction = options.invalidateAllAction;
  }

  get data(): ControlCabinetDetailData {
    return this.resolveData();
  }

  get cabinet() {
    return this.data.cabinet;
  }

  get building() {
    return this.data.building ?? null;
  }

  get controllers(): SPSController[] {
    return this.data.spsControllers ?? [];
  }

  get systemTypesByController(): Record<string, SPSControllerSystemType[]> {
    return this.data.systemTypesByController ?? {};
  }

  get canReadSps(): boolean {
    return canPerform('read', 'spscontroller');
  }

  get canCreateSps(): boolean {
    return canPerform('create', 'spscontroller');
  }

  get canUpdateSps(): boolean {
    return canPerform('update', 'spscontroller');
  }

  get canDeleteSps(): boolean {
    return canPerform('delete', 'spscontroller');
  }

  get canUpdateCabinet(): boolean {
    return canPerform('update', 'controlcabinet');
  }

  get canDeleteCabinet(): boolean {
    return canPerform('delete', 'controlcabinet');
  }

  getSystemTypesForController(controllerId: string): SPSControllerSystemType[] {
    return this.systemTypesByController[controllerId] ?? [];
  }

  startCabinetEdit(): void {
    this.showCabinetEdit = true;
    this.showSpsCreate = false;
    this.editingSps = undefined;
  }

  cancelCabinetEdit(): void {
    this.showCabinetEdit = false;
  }

  startSpsCreate(): void {
    this.showSpsCreate = true;
    this.showCabinetEdit = false;
    this.editingSps = undefined;
  }

  cancelSpsCreate(): void {
    this.showSpsCreate = false;
  }

  startSpsEdit(controller: SPSController): void {
    this.editingSps = controller;
    this.showCabinetEdit = false;
    this.showSpsCreate = false;
  }

  cancelSpsEdit(): void {
    this.editingSps = undefined;
  }

  async refreshAfterChange(): Promise<void> {
    await this.invalidateAllAction();
    this.resetForms();
  }

  async deleteCabinet(): Promise<void> {
    try {
      const impact = await this.manageControlCabinet.getDeleteImpact(this.cabinet.id);

      if (impact.sps_controllers_count > 0) {
        const firstConfirm = await this.confirmAction({
          title: translate('facility.delete_control_cabinet_confirm'),
          message: translate('facility.delete_control_cabinet_message').replace(
            '{count}',
            impact.sps_controllers_count.toString()
          ),
          confirmText: translate('common.confirm'),
          cancelText: translate('common.cancel'),
          variant: 'destructive'
        });

        if (!firstConfirm) {
          return;
        }

        const secondConfirm = await this.confirmAction({
          title: translate('facility.confirm_cascading_delete'),
          message: translate('facility.cascading_delete_message')
            .replace('{systemTypes}', impact.sps_controller_system_types_count.toString())
            .replace('{fieldDevices}', impact.field_devices_count.toString())
            .replace('{bacnetObjects}', impact.bacnet_objects_count.toString()),
          confirmText: translate('facility.delete_everything'),
          cancelText: translate('common.cancel'),
          variant: 'destructive'
        });

        if (!secondConfirm) {
          return;
        }
      }

      await this.manageControlCabinet.delete(this.cabinet.id);
      this.toastAction(translate('facility.control_cabinet_deleted'), 'success');
      await this.gotoAction('/facility/control-cabinets');
    } catch (error) {
      this.toastAction(
        error instanceof Error
          ? error.message
          : translate('facility.delete_control_cabinet_failed'),
        'error'
      );
    }
  }

  async deleteSps(controller: SPSController): Promise<void> {
    const confirmed = await this.confirmAction({
      title: translate('common.delete'),
      message: translate('facility.delete_sps_controller_confirm').replace(
        '{name}',
        controller.device_name
      ),
      confirmText: translate('common.delete'),
      cancelText: translate('common.cancel'),
      variant: 'destructive'
    });

    if (!confirmed) {
      return;
    }

    try {
      await this.manageSpsController.delete(controller.id);
      this.toastAction(translate('facility.sps_controller_deleted'), 'success');
      await this.refreshAfterChange();
    } catch (error) {
      this.toastAction(
        error instanceof Error ? error.message : translate('facility.delete_sps_controller_failed'),
        'error'
      );
    }
  }

  async copySps(controller: SPSController): Promise<void> {
    try {
      await this.manageSpsController.copy(controller.id);
      this.toastAction(translate('facility.sps_controller_copied'), 'success');
      await this.refreshAfterChange();
    } catch (error) {
      this.toastAction(
        error instanceof Error ? error.message : translate('facility.copy_failed'),
        'error'
      );
    }
  }

  async goToSpsController(controllerId: string): Promise<void> {
    await this.gotoAction(`/facility/sps-controllers/${controllerId}`);
  }

  private resetForms(): void {
    this.showCabinetEdit = false;
    this.showSpsCreate = false;
    this.editingSps = undefined;
  }
}
