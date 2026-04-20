import { render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

vi.mock('$lib/i18n/translator', () => ({
  createTranslator: () => ({
    subscribe(run: (value: (key: string) => string) => void) {
      run((key: string) => key);
      return () => {};
    }
  })
}));

import FacilityOverviewPage from '../../../src/routes/(app)/facility/+page.svelte';

describe('/facility overview', () => {
  it('renders protected facility deep links without checking permissions', () => {
    render(FacilityOverviewPage);

    expect(
      screen.getByRole('link', {
        name: 'facility.buildings facility.manage_building_infrastructure'
      })
    ).toHaveAttribute('href', '/facility/buildings');
    expect(
      screen.getByRole('link', {
        name: 'facility.control_cabinets facility.manage_cabinet_configurations'
      })
    ).toHaveAttribute('href', '/facility/control-cabinets');
    expect(
      screen.getByRole('link', { name: 'facility.sps_controllers facility.manage_sps_devices' })
    ).toHaveAttribute('href', '/facility/sps-controllers');
    expect(
      screen.getByRole('link', { name: 'facility.field_devices facility.field_devices_desc' })
    ).toHaveAttribute('href', '/facility/field-devices');
  });
});
