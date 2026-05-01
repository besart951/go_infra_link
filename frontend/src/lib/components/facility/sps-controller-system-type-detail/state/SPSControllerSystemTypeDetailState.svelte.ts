import { canPerform } from '$lib/utils/permissions.js';
import { t as translate } from '$lib/i18n/index.js';

import type {
  Building,
  ControlCabinet,
  FieldDevice,
  SPSController,
  SPSControllerSystemType
} from '$lib/domain/facility/index.js';

interface OverviewItem {
  label: string;
  value: string;
  monospace?: boolean;
}

export interface SPSControllerSystemTypeDetailData {
  systemType: SPSControllerSystemType;
  controller: SPSController;
  cabinet?: ControlCabinet | null;
  building?: Building | null;
  fieldDevices?: FieldDevice[];
  fieldDevicesTotal?: number;
  editRequested?: boolean;
}

export interface SPSControllerSystemTypeDetailStateOptions {
  data: () => SPSControllerSystemTypeDetailData;
  invalidateAllAction: () => Promise<void>;
}

export class SPSControllerSystemTypeDetailState {
  showEdit = $state(false);

  private readonly resolveData: () => SPSControllerSystemTypeDetailData;
  private readonly invalidateAllAction: () => Promise<void>;

  constructor(options: SPSControllerSystemTypeDetailStateOptions) {
    this.resolveData = options.data;
    this.invalidateAllAction = options.invalidateAllAction;
    this.showEdit = (options.data().editRequested ?? false) && this.canEdit;
  }

  get data(): SPSControllerSystemTypeDetailData {
    return this.resolveData();
  }

  get systemType(): SPSControllerSystemType {
    return this.data.systemType;
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

  get fieldDevices(): FieldDevice[] {
    return this.data.fieldDevices ?? [];
  }

  get fieldDevicesTotal(): number {
    return this.data.fieldDevicesTotal ?? this.fieldDevices.length;
  }

  get canEdit(): boolean {
    // System-type edits are currently persisted through the parent controller form.
    return canPerform('update', 'spscontrollersystemtype') && canPerform('update', 'spscontroller');
  }

  get backHref(): string {
    return `/facility/sps-controllers/${this.controller.id}`;
  }

  get title(): string {
    const documentName = this.systemType.document_name?.trim();
    if (documentName) {
      return documentName;
    }

    if (this.systemType.number != null) {
      return `${translate('facility.control_cabinet_detail.number')} ${this.systemType.number}`;
    }

    return this.systemType.system_type_name ?? translate('facility.sps_controller_system_type');
  }

  get subtitle(): string {
    return (
      this.systemType.system_type_name ||
      translate('facility.sps_controller_system_type_detail.subtitle')
    );
  }

  get overviewItems(): OverviewItem[] {
    return [
      {
        label: translate('facility.sps_controller'),
        value: this.valueOrDash(this.controller.device_name)
      },
      {
        label: translate('facility.system_type'),
        value: this.valueOrDash(this.systemType.system_type_name)
      },
      {
        label: translate('facility.control_cabinet_detail.number'),
        value: this.systemType.number != null ? String(this.systemType.number) : '-',
        monospace: true
      },
      {
        label: translate('facility.sps_controller_detail.document_name'),
        value: this.valueOrDash(this.systemType.document_name)
      },
      {
        label: translate('facility.sps_controller_detail.field_devices'),
        value: String(this.systemType.field_devices_count ?? 0),
        monospace: true
      },
      {
        label: translate('facility.control_cabinet'),
        value: this.cabinet?.control_cabinet_nr ?? this.controller.control_cabinet_id
      },
      {
        label: translate('facility.building'),
        value: this.building ? `${this.building.iws_code}-${this.building.building_group}` : '-'
      }
    ];
  }

  startEdit(): void {
    if (!this.canEdit) {
      return;
    }

    this.showEdit = true;
  }

  cancelEdit(): void {
    this.showEdit = false;
  }

  async refreshAfterChange(): Promise<void> {
    await this.invalidateAllAction();
    this.showEdit = false;
  }

  private valueOrDash(value: string | null | undefined): string {
    const normalized = value?.trim();
    return normalized && normalized.length > 0 ? normalized : '-';
  }
}
