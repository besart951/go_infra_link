/// <reference types="vitest" />

import { describe, expect, it } from 'vitest';
import type { RegistrationProcess } from '$lib/infrastructure/api/userRepository.js';
import { RegistrationProcessStepperState } from './RegistrationProcessStepperState.svelte.js';

function processWithStatuses(
  statuses: Array<'completed' | 'current' | 'pending' | 'failed' | 'blocked' | 'skipped'>
): RegistrationProcess {
  return {
    status: statuses.includes('failed') ? 'email_failed' : 'pending',
    email_status: statuses.includes('failed') ? 'failed' : 'sent',
    can_resend: false,
    send_count: 1,
    steps: statuses.map((status, index) => ({
      key: `step_${index + 1}`,
      label: `Step ${index + 1}`,
      status
    }))
  };
}

describe('RegistrationProcessStepperState', () => {
  it('marks the explicit current step and exposes progress metadata', () => {
    const state = new RegistrationProcessStepperState(() =>
      processWithStatuses(['completed', 'current', 'pending', 'pending'])
    );

    expect(state.hasProcess).toBe(true);
    expect(state.totalSteps).toBe(4);
    expect(state.currentStepNumber).toBe(2);
    expect(state.currentLabel).toBe('Step 2');
    expect(state.stepViews.map((step) => step.visualStatus)).toEqual([
      'complete',
      'active',
      'pending',
      'pending'
    ]);
  });

  it('treats failed and blocked steps as the active problem state', () => {
    const state = new RegistrationProcessStepperState(() =>
      processWithStatuses(['completed', 'failed', 'pending', 'pending'])
    );

    expect(state.currentStepNumber).toBe(2);
    expect(state.currentStatus).toBe('failed');
    expect(state.isProblemState).toBe(true);
  });
});
