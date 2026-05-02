import { describe, expect, it, vi } from 'vitest';
import { CrudPageActions, type CrudPageActionsOptions } from './crudPageActions.svelte.js';

function createActions(overrides: Partial<CrudPageActionsOptions<string>> = {}) {
  return new CrudPageActions<string>({
    reload: vi.fn(),
    deleteItem: vi.fn().mockResolvedValue(undefined),
    confirmDelete: vi.fn().mockResolvedValue(true),
    addToast: vi.fn(),
    getDeleteMessage: (item) => `Delete ${item}?`,
    getDeleteSuccessMessage: () => 'Deleted',
    getDeleteFailureMessage: () => 'Delete failed',
    getDeleteTitle: () => 'Delete',
    getDeleteConfirmText: () => 'Delete',
    getDeleteCancelText: () => 'Cancel',
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
});
