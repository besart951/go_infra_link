/// <reference types="vitest" />

import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => {
  class MockApiException extends Error {
    constructor(
      public status: number,
      public error: string,
      message: string
    ) {
      super(message);
      this.name = 'ApiException';
    }
  }

  return {
    ApiException: MockApiException,
    addToast: vi.fn(),
    createPhasePermission: vi.fn(),
    deletePhasePermission: vi.fn(),
    updatePhasePermission: vi.fn()
  };
});

vi.mock('$lib/api/client.js', () => ({
  ApiException: mocks.ApiException,
  getErrorMessage: (error: unknown) => (error instanceof Error ? error.message : String(error))
}));

vi.mock('$lib/components/toast.svelte', () => ({
  addToast: mocks.addToast
}));

vi.mock('$lib/i18n/index.js', () => ({
  t: (key: string) => key
}));

vi.mock('$lib/infrastructure/api/phasePermission.adapter.js', () => ({
  createPhasePermission: mocks.createPhasePermission,
  deletePhasePermission: mocks.deletePhasePermission,
  updatePhasePermission: mocks.updatePhasePermission
}));

import type { Phase, PhasePermission } from '$lib/domain/phase/index.js';
import type { Permission, Role } from '$lib/domain/role/index.js';
import { PhasePermissionRulesState } from './PhasePermissionRulesState.svelte.js';

function buildState(options: {
  rules?: () => PhasePermission[];
  onRulesChange?: () => Promise<void> | void;
} = {}): PhasePermissionRulesState {
  const roles: Role[] = [];
  const phases: Phase[] = [{ id: 'phase-1', name: 'Planung', created_at: '', updated_at: '' }];
  const permissions: Permission[] = [];

  return new PhasePermissionRulesState({
    roles: () => roles,
    phases: () => phases,
    permissions: () => permissions,
    rules: () => options.rules?.() ?? [],
    canManage: () => true,
    onRulesChange: options.onRulesChange
  });
}

describe('PhasePermissionRulesState', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('creates an empty rule with an empty permission list', async () => {
    mocks.createPhasePermission.mockResolvedValue({
      id: 'rule-1',
      phase_id: 'phase-1',
      role: 'admin_fzag',
      permissions: []
    });
    const onRulesChange = vi.fn();
    const state = buildState({ onRulesChange });

    await state.createEmptyRule('phase-1', 'admin_fzag');

    expect(mocks.createPhasePermission).toHaveBeenCalledWith({
      phase_id: 'phase-1',
      role: 'admin_fzag',
      permissions: []
    });
    expect(onRulesChange).toHaveBeenCalledTimes(1);
    expect(mocks.addToast).not.toHaveBeenCalled();
  });

  it('updates an existing stale rule when empty rule creation returns a conflict', async () => {
    let rules: PhasePermission[] = [];
    const existingRule: PhasePermission = {
      id: 'rule-1',
      phase_id: 'phase-1',
      role: 'admin_fzag',
      permissions: ['project.controlcabinet.read'],
      created_at: '',
      updated_at: ''
    };
    const onRulesChange = vi.fn(() => {
      rules = [existingRule];
    });
    mocks.createPhasePermission.mockRejectedValue(
      new mocks.ApiException(409, 'conflict', 'Berechtigung konnte nicht erstellt werden.')
    );
    mocks.updatePhasePermission.mockResolvedValue({
      ...existingRule,
      permissions: []
    });
    const state = buildState({ rules: () => rules, onRulesChange });

    await state.createEmptyRule('phase-1', 'admin_fzag');

    expect(onRulesChange).toHaveBeenCalledTimes(2);
    expect(mocks.updatePhasePermission).toHaveBeenCalledWith('rule-1', { permissions: [] });
    expect(mocks.addToast).not.toHaveBeenCalled();
  });
});
