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
  async function createEditing() {
    let editingApi: ReturnType<typeof useFieldDeviceEditing> | undefined;
    render(UseFieldDeviceEditingHarness, {
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
});
