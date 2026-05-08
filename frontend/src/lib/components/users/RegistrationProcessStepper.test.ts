/// <reference types="vitest" />

import { render, screen } from '@testing-library/svelte';
import type { RegistrationProcess } from '$lib/infrastructure/api/userRepository.js';
import RegistrationProcessStepper from './RegistrationProcessStepper.svelte';

vi.mock('$lib/i18n/translator', () => ({
  createTranslator: () => ({
    subscribe(fn: (value: (key: string, params?: Record<string, unknown>) => string) => void) {
      fn((key: string, params?: Record<string, unknown>) => {
        if (key === 'user.registration_step') return `Schritt ${params?.step}/${params?.total}`;
        if (key === 'user.registration_progress_aria') {
          return `Registrierung Schritt ${params?.step} von ${params?.total}: ${params?.label}`;
        }
        return key;
      });
      return () => {};
    }
  })
}));

function registrationProcess(): RegistrationProcess {
  return {
    status: 'pending',
    email_status: 'sent',
    can_resend: true,
    send_count: 1,
    steps: [
      {
        key: 'created',
        label: 'Angelegt',
        status: 'completed'
      },
      {
        key: 'email_sent',
        label: 'E-Mail versendet',
        status: 'completed'
      },
      {
        key: 'registered',
        label: 'Registriert',
        status: 'current'
      },
      {
        key: 'first_login',
        label: 'Erste Anmeldung',
        status: 'pending'
      }
    ]
  };
}

describe('RegistrationProcessStepper', () => {
  it('renders the active registration step as an accessible progressbar', () => {
    render(RegistrationProcessStepper, {
      process: registrationProcess()
    });

    const progressbar = screen.getByRole('progressbar', {
      name: 'Registrierung Schritt 3 von 4: Registriert'
    });

    expect(progressbar).toHaveAttribute('aria-valuenow', '3');
    expect(progressbar).toHaveAttribute('aria-valuemax', '4');
    expect(screen.getByText('Schritt 3/4')).toBeInTheDocument();
    expect(screen.getAllByText('Registriert')).toHaveLength(2);
  });
});
