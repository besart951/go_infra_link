/// <reference types="vitest" />

import { useFieldDeviceEditing } from './useFieldDeviceEditing.svelte.js';
import { buildFieldDevice } from '$lib/test/fieldDevice.fixtures.js';
import { render, waitFor } from '@testing-library/svelte';
import UseFieldDeviceEditingHarness from './__tests__/UseFieldDeviceEditingHarness.svelte';

const mockBulkUpdate = vi.fn();

vi.mock('$lib/infrastructure/api/fieldDeviceRepository.js', () => ({
  fieldDeviceRepository: {
    bulkUpdate: (...args: unknown[]) => mockBulkUpdate(...args)
  }
}));

vi.mock('$lib/components/toast.svelte', () => ({
  addToast: vi.fn()
}));

vi.mock('$lib/i18n/index.js', () => ({
  t: (key: string) => key
}));

describe('useFieldDeviceEditing', () => {
  const storageKey = 'fielddevice-editing';

  async function createEditing(
    props: Omit<Parameters<typeof render<typeof UseFieldDeviceEditingHarness>>[1], 'onReady'> = {}
  ) {
    let editingApi: ReturnType<typeof useFieldDeviceEditing> | undefined;
    render(UseFieldDeviceEditingHarness, {
      ...props,
      onReady: (editing) => {
        editingApi = editing;
      }
    });

    await waitFor(() => {
      expect(editingApi).toBeDefined();
    });

    return editingApi!;
  }

  beforeEach(() => {
    vi.clearAllMocks();
    sessionStorage.clear();
    mockBulkUpdate.mockResolvedValue({
      results: [{ id: 'fd-1', success: true }],
      total_count: 1,
      success_count: 1,
      failure_count: 0
    });
  });

  it('loads persisted draft edits from session storage', async () => {
    const device = buildFieldDevice();
    window.sessionStorage.setItem(
      storageKey,
      JSON.stringify({
        edits: [
          [
            device.id,
            {
              bmk: 'FD-PERSISTED',
              specification: {
                specification_brand: 'Persisted Brand'
              }
            }
          ]
        ],
        bacnetEdits: [
          [
            device.id,
            [
              [
                'bo-1',
                {
                  text_fix: 'TF-PERSISTED'
                }
              ]
            ]
          ]
        ],
        timestamp: Date.now()
      })
    );

    const editing = await createEditing();

    expect(editing.hasUnsavedChanges).toBe(true);
    expect(editing.pendingDeviceIds).toEqual([device.id]);
    expect(editing.getPendingValue(device.id, 'bmk')).toBe('FD-PERSISTED');
    expect(editing.getPendingSpecValue(device.id, 'specification_brand')).toBe('Persisted Brand');
    expect(editing.getBacnetPendingEdits(device.id).get('bo-1')?.text_fix).toBe('TF-PERSISTED');
  });

  it('keeps invalid BACnet edits pending and exposes field-level validation errors', async () => {
    const device = buildFieldDevice();
    const editing = await createEditing();

    editing.queueBacnetEdit(device.id, 'bo-1', 'software_number', -1);

    await editing.saveAllPendingEdits([device]);

    expect(mockBulkUpdate).not.toHaveBeenCalled();
    expect(editing.getBacnetPendingEdits(device.id).get('bo-1')?.software_number).toBe(-1);
    expect(editing.getBacnetClientErrors(device.id).get('bo-1')?.software_number).toBe(
      'field_device.bacnet.validation.software_number_range'
    );
  });

  it('clears successful edit phases while retaining failed phases after partial save success', async () => {
    const device = buildFieldDevice();
    const editing = await createEditing();
    const onSuccess = vi.fn();

    mockBulkUpdate.mockResolvedValueOnce({
      results: [
        {
          id: device.id,
          success: false,
          error: 'validation failed',
          fields: {
            'specification.specification_brand': 'brand rejected'
          }
        }
      ],
      total_count: 1,
      success_count: 0,
      failure_count: 1
    });

    editing.queueEdit(device.id, 'bmk', 'FD-PARTIAL');
    editing.queueSpecEdit(device.id, 'specification_brand', 'Rejected Brand');
    editing.queueBacnetEdit(device.id, 'bo-1', 'text_fix', 'TF-PARTIAL');

    await editing.saveAllPendingEdits([device], onSuccess);

    expect(editing.isFieldDirty(device.id, 'bmk')).toBe(false);
    expect(editing.isSpecFieldDirty(device.id, 'specification_brand')).toBe(true);
    expect(editing.getBacnetPendingEdits(device.id).size).toBe(0);
    expect(editing.getFieldError(device.id, 'specification_brand')).toBe('brand rejected');

    const updated = onSuccess.mock.calls[0][0][0];
    expect(updated.bmk).toBe('FD-PARTIAL');
    expect(updated.specification.specification_brand).toBe(
      device.specification?.specification_brand
    );
    expect(updated.bacnet_objects[0].text_fix).toBe('TF-PARTIAL');
  });

  it('notifies shared draft state with changed fields and pending values', async () => {
    const device = buildFieldDevice();
    const onSharedStateChange = vi.fn();
    const editing = await createEditing({ onSharedStateChange });

    editing.queueEdit(device.id, 'bmk', 'FD-SHARED');
    editing.queueSpecEdit(device.id, 'specification_brand', 'Shared Brand');
    editing.queueBacnetEdit(device.id, 'bo-1', 'text_fix', 'TF-SHARED');

    await waitFor(() => {
      expect(onSharedStateChange).toHaveBeenLastCalledWith({
        devices: [
          {
            device_id: device.id,
            changed_fields: [
              'bacnet_objects.bo-1.text_fix',
              'bmk',
              'specification.specification_brand'
            ],
            field_values: {
              'bacnet_objects.bo-1.text_fix': 'TF-SHARED',
              bmk: 'FD-SHARED',
              'specification.specification_brand': 'Shared Brand'
            }
          }
        ]
      });
    });
  });

  it('expires stale persisted draft edits', async () => {
    const device = buildFieldDevice();
    window.sessionStorage.setItem(
      storageKey,
      JSON.stringify({
        edits: [[device.id, { bmk: 'FD-STALE' }]],
        bacnetEdits: [],
        timestamp: Date.now() - 25 * 60 * 60 * 1000
      })
    );

    const editing = await createEditing();

    expect(editing.hasUnsavedChanges).toBe(false);
    expect(editing.pendingDeviceIds).toEqual([]);
    expect(window.sessionStorage.getItem(storageKey)).toBeNull();
  });

  it('clears persisted draft storage when saved edits leave no pending changes', async () => {
    const device = buildFieldDevice();
    const editing = await createEditing();

    editing.queueEdit(device.id, 'bmk', 'FD-UPDATED');

    await waitFor(() => {
      expect(window.sessionStorage.getItem(storageKey)).not.toBeNull();
    });

    await editing.saveAllPendingEdits([device]);

    await waitFor(() => {
      expect(window.sessionStorage.getItem(storageKey)).toBeNull();
    });
  });

  it('submits update API with complete payload for base, specification, and bacnet edits', async () => {
    const device = buildFieldDevice();
    const editing = await createEditing();

    editing.queueEdit(device.id, 'bmk', 'FD-UPDATED');
    editing.queueEdit(device.id, 'description', 'Updated description');
    editing.queueSpecEdit(device.id, 'specification_brand', 'Brand X');
    editing.queueSpecEdit(device.id, 'electrical_connection_power', 7.5);
    editing.queueBacnetEdit(device.id, 'bo-1', 'text_fix', 'TF-UPDATED');
    editing.queueBacnetEdit(device.id, 'bo-1', 'alarm_type_id', 'alarm-9');

    await editing.saveAllPendingEdits([device]);

    expect(mockBulkUpdate).toHaveBeenCalledWith({
      updates: [
        {
          id: 'fd-1',
          bmk: 'FD-UPDATED',
          description: 'Updated description',
          specification: {
            specification_brand: 'Brand X',
            electrical_connection_power: 7.5
          },
          bacnet_objects: [
            {
              id: 'bo-1',
              text_fix: 'TF-UPDATED',
              alarm_type_id: 'alarm-9'
            }
          ]
        }
      ]
    });
  });

  it('updates specification independently without including bacnet patches', async () => {
    const device = buildFieldDevice();
    const editing = await createEditing();
    const onSuccess = vi.fn();

    editing.queueSpecEdit(device.id, 'specification_supplier', 'Supplier B');

    await editing.saveAllPendingEdits([device], onSuccess);

    expect(mockBulkUpdate).toHaveBeenCalledWith({
      updates: [
        {
          id: 'fd-1',
          specification: {
            specification_supplier: 'Supplier B'
          }
        }
      ]
    });

    const updated = onSuccess.mock.calls[0][0][0];
    expect(updated.specification.specification_supplier).toBe('Supplier B');
    expect(updated.bacnet_objects[0].text_fix).toBe(device.bacnet_objects?.[0].text_fix);
  });

  it('updates bacnet objects independently without including specification payload', async () => {
    const device = buildFieldDevice();
    const editing = await createEditing();

    editing.queueBacnetEdit(device.id, 'bo-1', 'software_number', 42);

    await editing.saveAllPendingEdits([device]);

    expect(mockBulkUpdate).toHaveBeenCalledWith({
      updates: [
        {
          id: 'fd-1',
          bacnet_objects: [
            {
              id: 'bo-1',
              software_number: 42
            }
          ]
        }
      ]
    });
  });

  it('updates alarm mapping independently and preserves sibling bacnet data', async () => {
    const device = buildFieldDevice();
    const editing = await createEditing();
    const onSuccess = vi.fn();

    editing.queueBacnetEdit(device.id, 'bo-1', 'alarm_type_id', 'alarm-2');

    await editing.saveAllPendingEdits([device], onSuccess);

    expect(mockBulkUpdate).toHaveBeenCalledWith({
      updates: [
        {
          id: 'fd-1',
          bacnet_objects: [
            {
              id: 'bo-1',
              alarm_type_id: 'alarm-2'
            }
          ]
        }
      ]
    });

    const updated = onSuccess.mock.calls[0][0][0];
    expect(updated.bacnet_objects[0].alarm_type_id).toBe('alarm-2');
    expect(updated.bacnet_objects[0].software_type).toBe(device.bacnet_objects?.[0].software_type);
    expect(updated.bacnet_objects[0].text_fix).toBe(device.bacnet_objects?.[0].text_fix);
  });

  it('tracks pending edit categories for project-specific permission gating', async () => {
    const device = buildFieldDevice();
    const editing = await createEditing();

    expect(editing.hasPendingBaseEdits).toBe(false);
    expect(editing.hasPendingSpecificationEdits).toBe(false);
    expect(editing.hasPendingBacnetEdits).toBe(false);

    editing.queueEdit(device.id, 'bmk', 'FD-UPDATED');
    expect(editing.hasPendingBaseEdits).toBe(true);
    expect(editing.hasPendingSpecificationEdits).toBe(false);
    expect(editing.hasPendingBacnetEdits).toBe(false);

    editing.queueSpecEdit(device.id, 'specification_brand', 'Brand X');
    expect(editing.hasPendingSpecificationEdits).toBe(true);

    editing.queueBacnetEdit(device.id, 'bo-1', 'text_fix', 'TF-UPDATED');
    expect(editing.hasPendingBacnetEdits).toBe(true);
  });

  it('keeps all base field edits pending when apparat number validation rejects the base update', async () => {
    const device = buildFieldDevice({
      apparat_id: 'app-krg',
      apparat_nr: '2',
      system_part_id: 'sp-1'
    });
    const editing = await createEditing();
    const onSuccess = vi.fn();

    mockBulkUpdate.mockResolvedValueOnce({
      results: [
        {
          id: device.id,
          success: false,
          error: 'apparatnummer ist bereits vergeben',
          fields: {
            'fielddevice.apparat_nr': 'apparatnummer ist bereits vergeben'
          }
        }
      ],
      total_count: 1,
      success_count: 0,
      failure_count: 1
    });

    editing.queueEdit(device.id, 'apparat_id', 'app-abk');
    editing.queueEdit(device.id, 'apparat_nr', 2);

    await editing.saveAllPendingEdits([device], onSuccess);

    expect(onSuccess).not.toHaveBeenCalled();
    expect(editing.isFieldDirty(device.id, 'apparat_id')).toBe(true);
    expect(editing.isFieldDirty(device.id, 'apparat_nr')).toBe(true);

    mockBulkUpdate.mockResolvedValueOnce({
      results: [{ id: device.id, success: true }],
      total_count: 1,
      success_count: 1,
      failure_count: 0
    });

    editing.queueEdit(device.id, 'apparat_nr', 1);

    await editing.saveAllPendingEdits([device], onSuccess);

    expect(mockBulkUpdate).toHaveBeenLastCalledWith({
      updates: [
        {
          id: device.id,
          apparat_nr: 1,
          apparat_id: 'app-abk'
        }
      ]
    });
    expect(editing.isFieldDirty(device.id, 'apparat_id')).toBe(false);
    expect(editing.isFieldDirty(device.id, 'apparat_nr')).toBe(false);
  });
});
