import { t as translate } from '$lib/i18n/index.js';
import { canPerform } from '$lib/utils/permissions.js';
import { ManageSPSControllerUseCase } from '$lib/application/useCases/facility/manageSPSControllerUseCase.js';
import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
import type { ToastType } from '$lib/components/toast.svelte';

import type {
  Building,
  ControlCabinet,
  SPSController,
  SPSControllerSystemType
} from '$lib/domain/facility/index.js';

interface ConfirmOptions {
  title: string;
  message: string;
  confirmText: string;
  cancelText: string;
  variant?: 'destructive' | 'default';
}

interface OverviewItem {
  label: string;
  value: string;
  monospace?: boolean;
}

interface SystemTypeRow {
  id: string;
  number: string;
  documentName: string;
  systemTypeName: string;
  fieldDevicesCount: number;
}

export interface SPSControllerDetailData {
  controller: SPSController;
  cabinet?: ControlCabinet | null;
  building?: Building | null;
  systemTypes?: SPSControllerSystemType[];
}

export interface SPSControllerDetailStateOptions {
  data: () => SPSControllerDetailData;
  confirmAction: (options: ConfirmOptions) => Promise<boolean>;
  toastAction: (message: string, type?: ToastType) => void;
  gotoAction: (href: string) => Promise<void>;
  invalidateAllAction: () => Promise<void>;
}

export class SPSControllerDetailState {
  showEdit = $state(false);
  systemTypeSearchQuery = $state('');

  private readonly resolveData: () => SPSControllerDetailData;
  private readonly confirmAction: (options: ConfirmOptions) => Promise<boolean>;
  private readonly toastAction: (message: string, type?: ToastType) => void;
  private readonly gotoAction: (href: string) => Promise<void>;
  private readonly invalidateAllAction: () => Promise<void>;

  private readonly manageSpsController = new ManageSPSControllerUseCase(spsControllerRepository);

  constructor(options: SPSControllerDetailStateOptions) {
    this.resolveData = options.data;
    this.confirmAction = options.confirmAction;
    this.toastAction = options.toastAction;
    this.gotoAction = options.gotoAction;
    this.invalidateAllAction = options.invalidateAllAction;
  }

  get data(): SPSControllerDetailData {
    return this.resolveData();
  }

  get controller(): SPSController {
    return this.data.controller;
  }

  get cabinet(): ControlCabinet | null {
    return this.data.cabinet ?? null;
  }

  get building(): Building | null {
    return this.data.building ?? null;
  }

  get systemTypes(): SPSControllerSystemType[] {
    return this.data.systemTypes ?? [];
  }

  get canUpdateSps(): boolean {
    return canPerform('update', 'spscontroller');
  }

  get canDeleteSps(): boolean {
    return canPerform('delete', 'spscontroller');
  }

  get canUpdateSpsControllerSystemType(): boolean {
    // System-type edits are currently persisted through the parent controller form.
    return canPerform('update', 'spscontrollersystemtype') && canPerform('update', 'spscontroller');
  }

  get backHref(): string {
    if (this.cabinet?.id) {
      return `/facility/control-cabinets/${this.cabinet.id}`;
    }

    return '/facility/sps-controllers';
  }

  get title(): string {
    return this.controller.device_name;
  }

  get subtitle(): string {
    const description = this.controller.device_description?.trim();
    if (description) {
      return description;
    }

    if (this.cabinet && this.buildingLabel !== '-') {
      return `${translate('facility.control_cabinet')}: ${this.cabinet.control_cabinet_nr} • ${this.buildingLabel}`;
    }

    return translate('facility.sps_controller_detail.subtitle');
  }

  get buildingLabel(): string {
    if (!this.building) {
      return '-';
    }

    return `${this.building.iws_code}-${this.building.building_group}`;
  }

  get locationLabel(): string {
    const location = this.controller.device_location?.trim();
    if (location) {
      return location;
    }

    return this.buildingLabel;
  }

  get overviewItems(): OverviewItem[] {
    return [
      {
        label: translate('facility.ga_device'),
        value: this.valueOrDash(this.controller.ga_device)
      },
      {
        label: translate('facility.ip_address'),
        value: this.valueOrDash(this.controller.ip_address),
        monospace: true
      },
      {
        label: 'VLAN',
        value: this.valueOrDash(this.controller.vlan),
        monospace: true
      },
      {
        label: translate('facility.forms.sps_controller.subnet_label'),
        value: this.valueOrDash(this.controller.subnet),
        monospace: true
      },
      {
        label: translate('facility.forms.sps_controller.gateway_label'),
        value: this.valueOrDash(this.controller.gateway),
        monospace: true
      },
      {
        label: translate('facility.control_cabinet'),
        value: this.cabinet?.control_cabinet_nr ?? this.controller.control_cabinet_id
      },
      { label: translate('facility.building'), value: this.buildingLabel },
      {
        label: translate('facility.sps_controller_detail.location'),
        value: this.valueOrDash(this.locationLabel)
      }
    ];
  }

  get allSystemTypeRows(): SystemTypeRow[] {
    return this.systemTypes.map((systemType) => ({
      id: systemType.id,
      number: systemType.number != null ? String(systemType.number) : '-',
      documentName: this.valueOrDash(systemType.document_name),
      systemTypeName: systemType.system_type_name ?? '',
      fieldDevicesCount: systemType.field_devices_count ?? 0
    }));
  }

  get systemTypeRows(): SystemTypeRow[] {
    const query = this.systemTypeSearchQuery.trim().toLowerCase();

    if (!query) {
      return this.allSystemTypeRows;
    }

    return this.allSystemTypeRows.filter((row) => this.matchesSystemTypeSearch(row, query));
  }

  get systemTypeCountLabel(): string {
    if (this.systemTypeSearchQuery.trim()) {
      return `${this.systemTypeRows.length} / ${this.systemTypes.length} ${translate('facility.system_types')}`;
    }

    return `${this.systemTypes.length} ${translate('facility.system_types')}`;
  }

  setSystemTypeSearchQuery(value: string): void {
    this.systemTypeSearchQuery = value;
  }

  getSystemTypeHref(id: string): string {
    return `/facility/sps-controller-system-type/${id}`;
  }

  getSystemTypeActionHref(id: string): string {
    if (this.canUpdateSpsControllerSystemType) {
      return `${this.getSystemTypeHref(id)}?edit=1`;
    }

    return this.getSystemTypeHref(id);
  }

  startEdit(): void {
    this.showEdit = true;
  }

  cancelEdit(): void {
    this.showEdit = false;
  }

  async refreshAfterChange(): Promise<void> {
    await this.invalidateAllAction();
    this.showEdit = false;
  }

  async deleteController(): Promise<void> {
    const confirmed = await this.confirmAction({
      title: translate('common.delete'),
      message: translate('facility.delete_sps_controller_confirm').replace(
        '{name}',
        this.controller.device_name
      ),
      confirmText: translate('common.delete'),
      cancelText: translate('common.cancel'),
      variant: 'destructive'
    });

    if (!confirmed) {
      return;
    }

    try {
      await this.manageSpsController.delete(this.controller.id);
      this.toastAction(translate('facility.sps_controller_deleted'), 'success');
      await this.gotoAction(this.backHref);
    } catch (error) {
      this.toastAction(
        error instanceof Error ? error.message : translate('facility.delete_sps_controller_failed'),
        'error'
      );
    }
  }

  private valueOrDash(value: string | null | undefined): string {
    const normalized = value?.trim();
    return normalized && normalized.length > 0 ? normalized : '-';
  }

  private matchesSystemTypeSearch(row: SystemTypeRow, query: string): boolean {
    return (
      row.documentName.toLowerCase().includes(query) || row.number.toLowerCase().includes(query)
    );
  }
}
