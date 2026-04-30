/// <reference types="vitest" />

import { describe, expect, it, vi } from 'vitest';
import { AlarmCatalogState } from './AlarmCatalogState.svelte.js';
import type { AlarmFieldRepository } from '$lib/domain/ports/facility/alarmFieldRepository.js';
import type { AlarmTypeRepository } from '$lib/domain/ports/facility/alarmTypeRepository.js';
import type { AlarmUnitRepository } from '$lib/domain/ports/facility/alarmUnitRepository.js';
import type { AlarmField, AlarmType, AlarmTypeField, Unit } from '$lib/domain/facility/index.js';

const unit: Unit = {
  id: 'unit-1',
  code: 'C',
  symbol: '°C',
  name: 'Celsius'
};

const field: AlarmField = {
  id: 'field-1',
  key: 'temperature',
  label: 'Temperature',
  data_type: 'number',
  default_unit_code: 'C'
};

function alarmType(id: string, code: string): AlarmType {
  return {
    id,
    code,
    name: code,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z'
  };
}

function mapping(id: string, alarmTypeId: string): AlarmTypeField {
  return {
    id,
    alarm_type_id: alarmTypeId,
    alarm_field_id: field.id,
    alarm_field: field,
    display_order: id === 'mapping-1' ? 1 : 2,
    is_required: id === 'mapping-1',
    is_user_editable: true,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z'
  };
}

function createRepositories() {
  const type1 = alarmType('type-1', 'A1');
  const type2 = alarmType('type-2', 'A2');

  const unitRepository: AlarmUnitRepository = {
    list: vi.fn().mockResolvedValue({
      items: [unit],
      metadata: { total: 1, page: 1, pageSize: 200, totalPages: 1 }
    }),
    get: vi.fn(),
    create: vi.fn().mockResolvedValue(unit),
    update: vi.fn(),
    delete: vi.fn().mockResolvedValue(undefined)
  };
  const fieldRepository: AlarmFieldRepository = {
    list: vi.fn().mockResolvedValue({
      items: [field],
      metadata: { total: 1, page: 1, pageSize: 200, totalPages: 1 }
    }),
    get: vi.fn(),
    create: vi.fn().mockResolvedValue(field),
    update: vi.fn(),
    delete: vi.fn().mockResolvedValue(undefined)
  };
  const typeRepository: AlarmTypeRepository = {
    list: vi.fn().mockResolvedValue({ items: [type1, type2], total: 2, page: 1, totalPages: 1 }),
    get: vi.fn(),
    getWithFields: vi.fn().mockImplementation(async (id: string) => ({
      ...(id === type2.id ? type2 : type1),
      fields: [id === type2.id ? mapping('mapping-2', type2.id) : mapping('mapping-1', type1.id)]
    })),
    create: vi.fn().mockResolvedValue(type1),
    update: vi.fn(),
    delete: vi.fn().mockResolvedValue(undefined),
    createField: vi.fn().mockResolvedValue(mapping('mapping-created', type1.id)),
    updateField: vi.fn(),
    deleteField: vi.fn().mockResolvedValue(undefined)
  };

  return { fieldRepository, type1, type2, typeRepository, unitRepository };
}

function createState() {
  const repositories = createRepositories();
  const toast = vi.fn();
  const state = new AlarmCatalogState({
    ...repositories,
    addToast: toast,
    getErrorMessage: (error) => (error instanceof Error ? error.message : String(error)),
    translate: (key) => `t:${key}`
  });

  return { ...repositories, state, toast };
}

describe('AlarmCatalogState', () => {
  it('loads catalog data and selects first alarm type', async () => {
    const { state, typeRepository } = createState();

    await state.loadAll();

    expect(state.units).toEqual([unit]);
    expect(state.fields).toEqual([field]);
    expect(state.selectedTypeId).toBe('type-1');
    expect(typeRepository.getWithFields).toHaveBeenCalledWith('type-1');
    expect(state.typeFields).toEqual([mapping('mapping-1', 'type-1')]);
  });

  it('loads mappings when selected type changes', async () => {
    const { state, typeRepository } = createState();

    await state.loadAll();
    await state.selectType('type-2');

    expect(state.selectedTypeId).toBe('type-2');
    expect(typeRepository.getWithFields).toHaveBeenLastCalledWith('type-2');
    expect(state.typeFields).toEqual([mapping('mapping-2', 'type-2')]);
  });

  it('creates mapping, normalizes optional empty fields, resets form, and reloads mappings', async () => {
    const { state, toast, typeRepository } = createState();
    state.selectedTypeId = 'type-1';
    state.mapForm = {
      alarm_field_id: 'field-1',
      display_order: 5,
      is_required: true,
      is_user_editable: false,
      ui_group: '',
      default_unit_id: ''
    };

    await state.createMapping();

    expect(typeRepository.createField).toHaveBeenCalledWith('type-1', {
      alarm_field_id: 'field-1',
      display_order: 5,
      is_required: true,
      is_user_editable: false,
      ui_group: undefined,
      default_unit_id: undefined
    });
    expect(state.mapForm).toEqual({
      alarm_field_id: '',
      display_order: 0,
      is_required: false,
      is_user_editable: true,
      ui_group: '',
      default_unit_id: ''
    });
    expect(typeRepository.getWithFields).toHaveBeenCalledWith('type-1');
    expect(toast).toHaveBeenCalledWith(
      't:facility.alarm_catalog_page.toasts.mapping_created',
      'success'
    );
  });

  it('deletes mapping and reloads mappings for current selected type', async () => {
    const { state, toast, typeRepository } = createState();
    state.selectedTypeId = 'type-2';

    await state.deleteMapping('mapping-2');

    expect(typeRepository.deleteField).toHaveBeenCalledWith('mapping-2');
    expect(typeRepository.getWithFields).toHaveBeenCalledWith('type-2');
    expect(toast).toHaveBeenCalledWith(
      't:facility.alarm_catalog_page.toasts.mapping_deleted',
      'success'
    );
  });

  it('deleting selected type clears selection before reload selects next available type', async () => {
    const { state, type2, typeRepository } = createState();
    state.selectedTypeId = 'type-1';
    vi.mocked(typeRepository.list).mockResolvedValueOnce({
      items: [type2],
      total: 1,
      page: 1,
      totalPages: 1
    });

    await state.deleteType('type-1');

    expect(typeRepository.delete).toHaveBeenCalledWith('type-1');
    expect(state.selectedTypeId).toBe('type-2');
    expect(typeRepository.getWithFields).toHaveBeenCalledWith('type-2');
  });
});
