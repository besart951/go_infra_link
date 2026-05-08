/// <reference types="vitest" />

import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { UserDirectoryUser } from '$lib/infrastructure/api/userRepository.js';

const mocks = vi.hoisted(() => ({
  goto: vi.fn(),
  addToast: vi.fn(),
  listDirectory: vi.fn(),
  setRole: vi.fn(),
  disable: vi.fn(),
  enable: vi.fn(),
  delete: vi.fn(),
  resendRegistration: vi.fn(),
  confirm: vi.fn()
}));

vi.mock('$app/navigation', () => ({
  goto: mocks.goto
}));

vi.mock('$lib/api/client.js', () => ({
  getErrorMessage: (error: unknown) => (error instanceof Error ? error.message : String(error))
}));

vi.mock('$lib/components/toast.svelte', () => ({
  addToast: mocks.addToast
}));

vi.mock('$lib/i18n/index.js', () => {
  const translations: Record<string, string> = {
    'user.invitation_resend_wait': 'Einladung kann in {duration} erneut gesendet werden.',
    'user.invitation_resend_unavailable': 'Einladung kann aktuell nicht erneut gesendet werden.',
    'user.duration_minute': '{count} Minute',
    'user.duration_minutes': '{count} Minuten',
    'user.duration_second': '{count} Sekunde',
    'user.duration_seconds': '{count} Sekunden'
  };

  return {
    t: (key: string, params?: Record<string, string | number>) => {
      let value = translations[key] ?? key;
      for (const [name, param] of Object.entries(params ?? {})) {
        value = value.replaceAll(`{${name}}`, String(param));
      }
      return value;
    }
  };
});

vi.mock('$lib/infrastructure/api/userRepository.js', () => ({
  userRepository: {
    listDirectory: mocks.listDirectory,
    setRole: mocks.setRole,
    disable: mocks.disable,
    enable: mocks.enable,
    delete: mocks.delete,
    resendRegistration: mocks.resendRegistration
  }
}));

vi.mock('$lib/stores/auth.svelte.js', () => ({
  auth: {
    canAccessUserDirectory: true
  },
  getAllowedRolesForCreation: () => []
}));

vi.mock('$lib/stores/confirm-dialog.js', () => ({
  confirm: mocks.confirm
}));

import { UserDirectoryPageState } from './UserDirectoryPageState.svelte.js';

function invitedUser(resendAvailableAt: string): UserDirectoryUser {
  return {
    id: '00000000-0000-0000-0000-000000000001',
    first_name: 'Ada',
    last_name: 'Lovelace',
    email: 'ada@example.com',
    is_active: false,
    role: 'planer',
    role_display_name: 'Planer',
    created_at: '2026-05-08T10:00:00.000Z',
    updated_at: '2026-05-08T10:00:00.000Z',
    last_login_at: null,
    disabled_at: null,
    locked_until: null,
    failed_login_attempts: 0,
    teams: [],
    capabilities: {
      can_update: false,
      can_delete: false,
      can_disable: false,
      can_enable: false,
      can_change_role: false
    },
    registration_process: {
      status: 'pending',
      email_status: 'sent',
      steps: [],
      can_resend: false,
      resend_available_at: resendAvailableAt,
      send_count: 1
    }
  };
}

describe('UserDirectoryPageState invitation resend action', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('keeps resend visible during cooldown and explains the wait time', () => {
    const now = Date.parse('2026-05-08T10:30:00.000Z');
    const user = invitedUser('2026-05-08T10:31:00.000Z');
    const state = new UserDirectoryPageState();
    state.pageCapabilities = { can_create_user: true };

    expect(state.hasInvitationResendAction(user)).toBe(true);
    expect(state.canResendInvitation(user, now)).toBe(false);
    expect(state.invitationResendDisabledReason(user, now)).toBe(
      'Einladung kann in 1 Minute erneut gesendet werden.'
    );
  });

  it('enables resend locally once the backend-provided cooldown has elapsed', () => {
    const now = Date.parse('2026-05-08T10:31:00.000Z');
    const user = invitedUser('2026-05-08T10:31:00.000Z');
    const state = new UserDirectoryPageState();
    state.pageCapabilities = { can_create_user: true };

    expect(state.canResendInvitation(user, now)).toBe(true);
    expect(state.invitationResendDisabledReason(user, now)).toBeNull();
  });
});
