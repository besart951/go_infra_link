/// <reference types="vitest" />

import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  addToast: vi.fn(),
  getCurrent: vi.fn(),
  updateCurrent: vi.fn(),
  updateCurrentPassword: vi.fn(),
  listTeams: vi.fn(),
  listMembers: vi.fn(),
  getUserPreference: vi.fn(),
  saveUserPreference: vi.fn(),
  sendEmailVerificationCode: vi.fn(),
  verifyEmail: vi.fn()
}));

vi.mock('$lib/api/client.js', () => ({
  getErrorMessage: (error: unknown) => (error instanceof Error ? error.message : String(error))
}));

vi.mock('$lib/components/toast.svelte', () => ({
  addToast: mocks.addToast
}));

vi.mock('$lib/i18n/index.js', () => ({
  t: (key: string) => key
}));

vi.mock('$lib/infrastructure/api/userRepository.js', () => ({
  userRepository: {
    getCurrent: mocks.getCurrent,
    updateCurrent: mocks.updateCurrent,
    updateCurrentPassword: mocks.updateCurrentPassword
  }
}));

vi.mock('$lib/infrastructure/api/teamRepository.js', () => ({
  teamRepository: {
    list: mocks.listTeams,
    listMembers: mocks.listMembers
  }
}));

vi.mock('$lib/infrastructure/api/notificationPreferenceRepository.js', () => ({
  notificationPreferenceRepository: {
    getUserPreference: mocks.getUserPreference,
    saveUserPreference: mocks.saveUserPreference,
    sendEmailVerificationCode: mocks.sendEmailVerificationCode,
    verifyEmail: mocks.verifyEmail
  }
}));

import { AccountPageState } from './AccountPageState.svelte.js';

function userWithPermissions(
  permissions: string[] = [],
  role:
    | 'superadmin'
    | 'admin_fzag'
    | 'fzag'
    | 'admin_planer'
    | 'planer'
    | 'admin_entrepreneur'
    | 'entrepreneur' = 'planer'
) {
  return {
    id: 'user-1',
    first_name: 'Marc',
    last_name: 'Dahinden',
    email: 'marc.dahinden@gmail.com',
    is_active: true,
    role,
    role_display_name: 'Planer',
    permissions,
    can_access_user_directory: false,
    created_at: '2026-05-04T00:00:00Z',
    updated_at: '2026-05-04T00:00:00Z',
    failed_login_attempts: 0
  };
}

function notificationPreference() {
  return {
    id: 'pref-1',
    user_id: 'user-1',
    notification_email: 'marc.dahinden@gmail.com',
    notification_email_verified_at: null,
    channel: 'both',
    frequency: 'immediate'
  };
}

describe('AccountPageState', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.getUserPreference.mockResolvedValue(notificationPreference());
  });

  it('loads the account for users without team.read without calling protected team APIs', async () => {
    const user = userWithPermissions();
    mocks.getCurrent.mockResolvedValue(user);

    const state = new AccountPageState();

    await state.loadAccount();

    expect(state.currentUser).toEqual(user);
    expect(state.email).toBe('marc.dahinden@gmail.com');
    expect(state.userTeams).toEqual([]);
    expect(state.teamsError).toBeNull();
    expect(mocks.listTeams).not.toHaveBeenCalled();
    expect(mocks.listMembers).not.toHaveBeenCalled();
    expect(mocks.addToast).not.toHaveBeenCalled();
  });

  it('loads teams for superadmin even when team.read is not listed', async () => {
    const user = userWithPermissions([], 'superadmin');
    mocks.getCurrent.mockResolvedValue(user);
    mocks.listTeams.mockResolvedValue({
      items: [{ id: 'team-1', name: 'Core Team' }]
    });
    mocks.listMembers.mockResolvedValue({
      items: [{ user_id: 'user-1' }]
    });

    const state = new AccountPageState();

    await state.loadAccount();

    expect(state.userTeams).toEqual(['Core Team']);
    expect(mocks.listTeams).toHaveBeenCalledWith({ page: 1, limit: 100, search: '' });
    expect(mocks.listMembers).toHaveBeenCalledWith('team-1', { page: 1, limit: 1000 });
  });

  it('requires current password before changing password', async () => {
    const state = new AccountPageState();
    state.currentUser = userWithPermissions();
    state.newPassword = 'new-password';
    state.confirmPassword = 'new-password';

    await state.handlePasswordSubmit({ preventDefault: vi.fn() } as unknown as SubmitEvent);

    expect(mocks.updateCurrentPassword).not.toHaveBeenCalled();
    expect(mocks.addToast).toHaveBeenCalledWith('validation.required', 'error');
  });

  it('passes current and new password to password update endpoint', async () => {
    const state = new AccountPageState();
    state.currentUser = userWithPermissions();
    state.currentPassword = 'old-password';
    state.newPassword = 'new-password';
    state.confirmPassword = 'new-password';
    mocks.updateCurrentPassword.mockResolvedValue(state.currentUser);

    await state.handlePasswordSubmit({ preventDefault: vi.fn() } as unknown as SubmitEvent);

    expect(mocks.updateCurrentPassword).toHaveBeenCalledWith(
      'user-1',
      'old-password',
      'new-password'
    );
    expect(state.currentPassword).toBe('');
    expect(state.newPassword).toBe('');
    expect(state.confirmPassword).toBe('');
  });
});
