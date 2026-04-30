import { addToast } from '$lib/components/toast.svelte';
import { getErrorMessage } from '$lib/api/client.js';
import { alarmUnitRepository } from '$lib/infrastructure/api/alarmUnitRepository.js';
import { alarmFieldRepository } from '$lib/infrastructure/api/alarmFieldRepository.js';
import { alarmTypeRepository } from '$lib/infrastructure/api/alarmTypeRepository.js';
import type { AlarmFieldRepository } from '$lib/domain/ports/facility/alarmFieldRepository.js';
import type { AlarmTypeRepository } from '$lib/domain/ports/facility/alarmTypeRepository.js';
import type { AlarmUnitRepository } from '$lib/domain/ports/facility/alarmUnitRepository.js';
import type {
  AlarmField,
  AlarmType,
  AlarmTypeField,
  CreateAlarmTypeFieldRequest,
  Unit
} from '$lib/domain/facility/index.js';

export type AlarmCatalogToastType = 'success' | 'error';
export type AlarmCatalogToast = (message: string, type: AlarmCatalogToastType) => void;

export interface AlarmCatalogStateOptions {
  unitRepository?: AlarmUnitRepository;
  fieldRepository?: AlarmFieldRepository;
  typeRepository?: AlarmTypeRepository;
  addToast?: AlarmCatalogToast;
  getErrorMessage?: (error: unknown) => string;
  translate?: (key: string) => string;
}

const emptyUnitForm = () => ({ code: '', symbol: '', name: '' });
const emptyFieldForm = () => ({
  key: '',
  label: '',
  data_type: 'string' as AlarmField['data_type'],
  default_unit_code: ''
});
const emptyTypeForm = () => ({ code: '', name: '' });
const emptyMapForm = (): CreateAlarmTypeFieldRequest => ({
  alarm_field_id: '',
  display_order: 0,
  is_required: false,
  is_user_editable: true,
  ui_group: '',
  default_unit_id: ''
});

export class AlarmCatalogState {
  units = $state<Unit[]>([]);
  fields = $state<AlarmField[]>([]);
  types = $state<AlarmType[]>([]);
  selectedTypeId = $state('');
  typeFields = $state<AlarmTypeField[]>([]);
  loading = $state(false);

  unitForm = $state(emptyUnitForm());
  fieldForm = $state(emptyFieldForm());
  typeForm = $state(emptyTypeForm());
  mapForm = $state(emptyMapForm());

  readonly dataTypeOptions: AlarmField['data_type'][] = [
    'number',
    'integer',
    'boolean',
    'string',
    'enum',
    'duration',
    'state_map',
    'json'
  ];

  readonly selectClass =
    'h-9 w-full rounded-md border border-input bg-background px-3 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50';

  private readonly unitRepository: AlarmUnitRepository;
  private readonly fieldRepository: AlarmFieldRepository;
  private readonly typeRepository: AlarmTypeRepository;
  private readonly notify: AlarmCatalogToast;
  private readonly formatError: (error: unknown) => string;
  private readonly translate: (key: string) => string;

  constructor(options: AlarmCatalogStateOptions = {}) {
    this.unitRepository = options.unitRepository ?? alarmUnitRepository;
    this.fieldRepository = options.fieldRepository ?? alarmFieldRepository;
    this.typeRepository = options.typeRepository ?? alarmTypeRepository;
    this.notify = options.addToast ?? addToast;
    this.formatError = options.getErrorMessage ?? getErrorMessage;
    this.translate = options.translate ?? ((key) => key);
  }

  async loadAll(): Promise<void> {
    this.loading = true;
    try {
      const [unitsRes, fieldsRes, typesRes] = await Promise.all([
        this.unitRepository.list({ pagination: { page: 1, pageSize: 200 }, search: { text: '' } }),
        this.fieldRepository.list({ pagination: { page: 1, pageSize: 200 }, search: { text: '' } }),
        this.typeRepository.list({ page: 1, pageSize: 200 })
      ]);
      this.units = unitsRes.items;
      this.fields = fieldsRes.items;
      this.types = typesRes.items;
      if (!this.selectedTypeId && this.types.length > 0) {
        this.selectedTypeId = this.types[0].id;
      }
      await this.loadTypeFields(this.selectedTypeId);
    } catch (error) {
      this.notify(this.formatError(error), 'error');
    } finally {
      this.loading = false;
    }
  }

  async selectType(typeId: string): Promise<void> {
    this.selectedTypeId = typeId;
    await this.loadTypeFields(typeId);
  }

  async loadTypeFields(typeId = this.selectedTypeId): Promise<void> {
    if (!typeId) {
      this.typeFields = [];
      return;
    }
    try {
      const detail = await this.typeRepository.getWithFields(typeId);
      this.typeFields = detail.fields ?? [];
    } catch (error) {
      this.notify(this.formatError(error), 'error');
    }
  }

  async createUnit(): Promise<void> {
    try {
      await this.unitRepository.create(this.unitForm);
      this.unitForm = emptyUnitForm();
      await this.loadAll();
      this.success('facility.alarm_catalog_page.toasts.unit_created');
    } catch (error) {
      this.notify(this.formatError(error), 'error');
    }
  }

  async deleteUnit(id: string): Promise<void> {
    try {
      await this.unitRepository.delete(id);
      await this.loadAll();
      this.success('facility.alarm_catalog_page.toasts.unit_deleted');
    } catch (error) {
      this.notify(this.formatError(error), 'error');
    }
  }

  async createField(): Promise<void> {
    try {
      await this.fieldRepository.create({
        key: this.fieldForm.key,
        label: this.fieldForm.label,
        data_type: this.fieldForm.data_type,
        default_unit_code: this.fieldForm.default_unit_code || undefined
      });
      this.fieldForm = emptyFieldForm();
      await this.loadAll();
      this.success('facility.alarm_catalog_page.toasts.field_created');
    } catch (error) {
      this.notify(this.formatError(error), 'error');
    }
  }

  async deleteField(id: string): Promise<void> {
    try {
      await this.fieldRepository.delete(id);
      await this.loadAll();
      this.success('facility.alarm_catalog_page.toasts.field_deleted');
    } catch (error) {
      this.notify(this.formatError(error), 'error');
    }
  }

  async createType(): Promise<void> {
    try {
      const created = await this.typeRepository.create(this.typeForm);
      this.typeForm = emptyTypeForm();
      this.selectedTypeId = created.id;
      await this.loadAll();
      this.success('facility.alarm_catalog_page.toasts.type_created');
    } catch (error) {
      this.notify(this.formatError(error), 'error');
    }
  }

  async deleteType(id: string): Promise<void> {
    try {
      await this.typeRepository.delete(id);
      if (this.selectedTypeId === id) {
        this.selectedTypeId = '';
      }
      await this.loadAll();
      this.success('facility.alarm_catalog_page.toasts.type_deleted');
    } catch (error) {
      this.notify(this.formatError(error), 'error');
    }
  }

  async createMapping(): Promise<void> {
    if (!this.selectedTypeId) return;

    try {
      await this.typeRepository.createField(this.selectedTypeId, {
        ...this.mapForm,
        default_unit_id: this.mapForm.default_unit_id || undefined,
        ui_group: this.mapForm.ui_group || undefined
      });
      this.mapForm = emptyMapForm();
      await this.loadTypeFields(this.selectedTypeId);
      this.success('facility.alarm_catalog_page.toasts.mapping_created');
    } catch (error) {
      this.notify(this.formatError(error), 'error');
    }
  }

  async deleteMapping(id: string): Promise<void> {
    try {
      await this.typeRepository.deleteField(id);
      await this.loadTypeFields(this.selectedTypeId);
      this.success('facility.alarm_catalog_page.toasts.mapping_deleted');
    } catch (error) {
      this.notify(this.formatError(error), 'error');
    }
  }

  private success(key: string): void {
    this.notify(this.translate(key), 'success');
  }
}
