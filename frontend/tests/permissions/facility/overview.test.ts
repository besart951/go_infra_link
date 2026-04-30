import { render, screen } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const state = vi.hoisted(() => {
  const grantedPermissions = new Set<string>();

  return {
    setPermissions(permissions: string[]) {
      grantedPermissions.clear();
      for (const granted of permissions) {
        grantedPermissions.add(granted);
      }
    },
    resetPermissions() {
      grantedPermissions.clear();
    },
    canPerform(action: string, resource: string) {
      return grantedPermissions.has(`${resource}.${action}`);
    }
  };
});

vi.mock('$lib/i18n/translator', () => ({
  createTranslator: () => ({
    subscribe(run: (value: (key: string) => string) => void) {
      run((key: string) => key);
      return () => {};
    }
  })
}));

vi.mock('$lib/utils/permissions.js', () => ({
  canPerform: (action: string, resource: string) => state.canPerform(action, resource)
}));

import FacilityOverviewPage from '../../../src/routes/(app)/facility/+page.svelte';

describe('/facility overview', () => {
  beforeEach(() => {
    state.resetPermissions();
  });

  it('hides protected facility deep links without matching permissions', () => {
    render(FacilityOverviewPage);

    expect(screen.getByText('hub.no_access')).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: /facility.buildings/ })).not.toBeInTheDocument();
  });

  it('renders only facility deep links allowed by read permissions', () => {
    state.setPermissions(['building.read', 'fielddevice.read']);

    render(FacilityOverviewPage);

    expect(
      screen.getByRole('link', {
        name: 'facility.buildings facility.buildings_desc'
      })
    ).toHaveAttribute('href', '/facility/buildings');
    expect(
      screen.getByRole('link', { name: 'facility.field_devices facility.field_devices_desc' })
    ).toHaveAttribute('href', '/facility/field-devices');
    expect(
      screen.queryByRole('link', { name: /facility.control_cabinets/ })
    ).not.toBeInTheDocument();
  });
});
