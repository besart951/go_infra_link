import { describe, expect, it, vi } from 'vitest';
import { CrudPageActions, type CrudPageActionsOptions } from './crudPageActions.svelte.js';

function createActions(overrides: Partial<CrudPageActionsOptions<string>> = {}) {
  return new CrudPageActions<string>({
    reload: vi.fn(),
    deleteItem: vi.fn().mockResolvedValue(undefined),
    confirmDelete: vi.fn().mockResolvedValue(true),
    addToast: vi.fn(),
    confirmImpactUpdate: vi.fn().mockResolvedValue(true),
    getItemId: (item) => item,
    getDeleteMessage: (item) => `Delete ${item}?`,
    getDeleteSuccessMessage: () => 'Deleted',
    getDeleteFailureMessage: () => 'Delete failed',
    getDeleteTitle: () => 'Delete',
    getDeleteConfirmText: () => 'Delete',
    getDeleteCancelText: () => 'Cancel',
    getBacnetUsageMessage: (count) => `${count} uses`,
    getBacnetDeleteBlockedMessage: (count) => `Used ${count} times`,
    getBacnetUpdateConfirmTitle: () => 'Used',
    getBacnetUpdateConfirmMessage: (count) => `Used ${count}`,
    getBacnetUpdateConfirmAgainTitle: () => 'Confirm again',
    getBacnetUpdateConfirmAgainMessage: (count) => `Still used ${count}`,
    getBacnetUpdateConfirmText: () => 'Update',
    ...overrides
  });
}

describe('CrudPageActions', () => {
  it('opens create and edit forms through one interface', () => {
    const actions = createActions();

    actions.create();
    expect(actions.showForm).toBe(true);
    expect(actions.editingItem).toBeUndefined();

    actions.edit('item-1');
    expect(actions.showForm).toBe(true);
    expect(actions.editingItem).toBe('item-1');

    actions.cancel();
    expect(actions.showForm).toBe(false);
    expect(actions.editingItem).toBeUndefined();
  });

  it('confirms, deletes, reloads, and reports success', async () => {
    const reload = vi.fn();
    const deleteItem = vi.fn().mockResolvedValue(undefined);
    const addToast = vi.fn();
    const actions = createActions({ reload, deleteItem, addToast });

    await actions.delete('item-1');

    expect(deleteItem).toHaveBeenCalledWith('item-1');
    expect(addToast).toHaveBeenCalledWith('Deleted', 'success');
    expect(reload).toHaveBeenCalled();
  });

  it('does not delete when confirmation is rejected', async () => {
    const deleteItem = vi.fn();
    const actions = createActions({
      confirmDelete: vi.fn().mockResolvedValue(false),
      deleteItem
    });

    await actions.delete('item-1');

    expect(deleteItem).not.toHaveBeenCalled();
  });

  it('reports delete failures without reloading', async () => {
    const reload = vi.fn();
    const addToast = vi.fn();
    const actions = createActions({
      reload,
      addToast,
      deleteItem: vi.fn().mockRejectedValue(new Error('Backend said no'))
    });

    await actions.delete('item-1');

    expect(addToast).toHaveBeenCalledWith('Backend said no', 'error');
    expect(reload).not.toHaveBeenCalled();
  });

  it('blocks delete when the item is used by bacnet objects', async () => {
    const deleteItem = vi.fn();
    const addToast = vi.fn();
    const actions = createActions({ deleteItem, addToast });
    actions.setBacnetUsageCounts({ 'item-1': 3 });

    await actions.delete('item-1');

    expect(actions.isDeleteDisabled('item-1')).toBe(true);
    expect(actions.getDeleteLabel('item-1', 'Delete')).toBe('Used 3 times');
    expect(deleteItem).not.toHaveBeenCalled();
    expect(addToast).toHaveBeenCalledWith('Used 3 times', 'error');
  });

  it('asks twice before updating a bacnet-linked item', async () => {
    const confirmImpactUpdate = vi.fn().mockResolvedValue(true);
    const actions = createActions({ confirmImpactUpdate });
    actions.setBacnetUsageCounts({ 'item-1': 2 });

    await expect(actions.confirmBacnetImpactUpdate('item-1')).resolves.toBe(true);

    expect(confirmImpactUpdate).toHaveBeenCalledTimes(2);
    expect(confirmImpactUpdate).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({ title: 'Used', message: 'Used 2', confirmText: 'Update' })
    );
    expect(confirmImpactUpdate).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({
        title: 'Confirm again',
        message: 'Still used 2',
        confirmText: 'Update'
      })
    );
  });

  it('cancels the second update confirmation', async () => {
    const confirmImpactUpdate = vi.fn().mockResolvedValueOnce(true).mockResolvedValueOnce(false);
    const actions = createActions({ confirmImpactUpdate });
    actions.setBacnetUsageCounts({ 'item-1': 2 });

    await expect(actions.confirmBacnetImpactUpdate('item-1')).resolves.toBe(false);
  });
});
